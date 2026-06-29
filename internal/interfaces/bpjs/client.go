package bpjs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"service/internal/infrastructure/config"
	"service/pkg/logger"
	"strings"
)

// BpjsClient adalah core HTTP client untuk API BPJS.
type BpjsClient interface {
	DoRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error)
}

type client struct {
	cfg        config.BpjsConfig
	httpClient *http.Client
}

// NewBpjsClient membuat instance baru dari BpjsClient.
func NewBpjsClient(cfg config.BpjsConfig) BpjsClient {
	return &client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// DoRequest secara otomatis menangani header signature, request, pengecekan error dari BPJS, dan dekripsi respons.
func (c *client) DoRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	if !c.cfg.Enabled {
		return nil, fmt.Errorf("BPJS VClaim integration is disabled in configuration")
	}

	baseURL := strings.TrimRight(c.cfg.BaseURL, "/")
	serviceName := strings.Trim(c.cfg.ServiceName, "/")
	endpointPath := strings.TrimLeft(endpoint, "/")

	var url string
	if serviceName != "" {
		url = fmt.Sprintf("%s/%s/%s", baseURL, serviceName, endpointPath)
	} else {
		url = fmt.Sprintf("%s/%s", baseURL, endpointPath)
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Gunakan method dari config Anda untuk generate header secara otomatis
	consID, secretKey, userKey, timestamp, signature := c.cfg.SetHeader()

	req.Header.Set("X-cons-id", consID)
	req.Header.Set("X-timestamp", timestamp)
	req.Header.Set("X-signature", signature)
	req.Header.Set("Accept", "application/json")

	// Aplicares tidak membutuhkan user_key, kirimkan hanya jika tidak kosong (VClaim/Antrol)
	if userKey != "" {
		req.Header.Set("user_key", userKey)
	}

	// Aplicares menggunakan application/json, sementara layanan lain menggunakan x-www-form-urlencoded
	if strings.Contains(strings.ToLower(c.cfg.ServiceName), "aplicaresws") {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "Application/x-www-form-urlencoded")
	}

	// Injeksi Token Khusus Antrean RS (x-token) via Context
	if token, ok := ctx.Value(TokenContextKey).(string); ok && token != "" {
		req.Header.Set("x-token", token)
	}

	// [RADAR] LOG OUTGOING REQUEST
	logger.Default().Info("🛫 BPJS API Request Sent",
		logger.String("target_url", url),
		logger.String("method", method),
		logger.String("used_cons_id", consID),
		logger.String("used_user_key", userKey),
		logger.String("used_x_token", req.Header.Get("x-token")),
		logger.String("timestamp", timestamp),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// [RADAR] LOG INCOMING RESPONSE
	// logger.Default().Info("🛬 BPJS API Response Received",
	// 	logger.Int("status_code", resp.StatusCode),
	// 	logger.String("raw_body", strings.TrimSpace(string(respBody))),
	// )

	// 1. Tangani penolakan API Gateway secara langsung (misal: 401/403 dengan balikan plaintext "Authentication failed")
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("BPJS Gateway Error (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	// 2. Tangani kasus BPJS mengembalikan HTTP 200 OK tetapi body berupa Plain Text (misal "Authentication failed")
	bodyStr := strings.TrimSpace(string(respBody))
	if !strings.HasPrefix(bodyStr, "{") && !strings.HasPrefix(bodyStr, "[") {
		logger.Default().Error("BPJS Non-JSON Response", logger.String("body", bodyStr), logger.String("endpoint", endpointPath))
		return nil, fmt.Errorf("BPJS API Error (Plain Text): %s", bodyStr)
	}

	// 3. Parsing struktur dasar BPJS Response secara fleksibel
	var baseResp struct {
		MetaData struct {
			Code    interface{} `json:"code"` // Gunakan interface{} krn VClaim mereturn string "200", Antrol mereturn int 200
			Message string      `json:"message"`
		} `json:"metaData"`
		Metadata struct { // Antrean RS (Antrol) menggunakan key dengan huruf kecil
			Code    interface{} `json:"code"`
			Message string      `json:"message"`
		} `json:"metadata"`
		Response json.RawMessage `json:"response"`
	}

	if err := json.Unmarshal(respBody, &baseResp); err != nil {
		return nil, fmt.Errorf("failed to parse base BPJS response: %w. Body: %s", err, string(respBody))
	}

	// 4. Ekstrak code dan message secara dinamis untuk format VClaim dan Antrean RS
	var codeStr, message string
	if baseResp.MetaData.Code != nil {
		codeStr = fmt.Sprintf("%v", baseResp.MetaData.Code)
		message = baseResp.MetaData.Message
	} else if baseResp.Metadata.Code != nil {
		codeStr = fmt.Sprintf("%v", baseResp.Metadata.Code)
		message = baseResp.Metadata.Message
	}

	// Kode sukses BPJS biasanya 200, 201, atau terkadang 1
	if codeStr != "200" && codeStr != "1" && codeStr != "201" {
		logger.Default().Warn("BPJS API Error", logger.String("code", codeStr), logger.String("message", message))
		return nil, fmt.Errorf("BPJS error %s: %s", codeStr, message)
	}

	// Jika "response" null / kosong, berarti sukses tanpa payload
	if len(baseResp.Response) == 0 || string(baseResp.Response) == "null" {
		return []byte("{}"), nil
	}

	// Dekripsi response (VClaim V2)
	var cipherText string
	if err := json.Unmarshal(baseResp.Response, &cipherText); err != nil {
		// Jika response bukan string terenkripsi (misal: response sukses tanpa data)
		return baseResp.Response, nil
	}

	// Lakukan dekripsi
	decryptedStr, err := DecryptPayload(cipherText, consID, secretKey, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt BPJS payload: %w", err)
	}

	return []byte(decryptedStr), nil
}
