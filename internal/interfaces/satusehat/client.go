package satusehat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"service/internal/infrastructure/config"
	"service/pkg/logger"
)

type client struct {
	cfg         config.SatuSehatConfig
	httpClient  *http.Client
	tokenData   map[string]interface{}
	tokenExpiry time.Time
	mu          sync.RWMutex
}

// NewSatuSehatClient membuat instance baru dari SatuSehatClient.
func NewSatuSehatClient(cfg config.SatuSehatConfig) SatuSehatClient {
	return &client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// GetAccessToken mengelola request token OAuth2 (Client Credentials) dan melakukan caching di memori.
func (c *client) GetAccessToken(ctx context.Context) (map[string]interface{}, error) {
	c.mu.RLock()
	// Cek apakah token masih valid (dengan buffer 1 menit untuk mencegah kegagalan)
	if c.tokenData != nil && time.Now().Before(c.tokenExpiry) {
		data := c.tokenData
		c.mu.RUnlock()
		return data, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check setelah mendapat lock
	if c.tokenData != nil && time.Now().Before(c.tokenExpiry) {
		return c.tokenData, nil
	}

	data := url.Values{}
	data.Set("client_id", c.cfg.ClientID)
	data.Set("client_secret", c.cfg.ClientSecret)
	// Sesuai standar OAuth2, grant_type disisipkan dalam body request
	data.Set("grant_type", "client_credentials")

	// Deteksi otomatis jika Auth URL di environment hanya berupa Base (tanpa /accesstoken)
	authURL := strings.TrimRight(c.cfg.AuthURL, "/")
	if !strings.HasSuffix(authURL, "accesstoken") {
		authURL += "/accesstoken"
	}
	// Kemenkes juga sering mensyaratkan query parameter grant_type di URL
	if !strings.Contains(authURL, "grant_type") {
		authURL += "?grant_type=client_credentials"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	// Baca terlebih dahulu seluruh response body untuk memudahkan debugging jika gagal parse JSON
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Default().Error("SatuSehat Auth Error", logger.Int("status", resp.StatusCode), logger.String("response", string(respBody)))

		// Coba parse response body sebagai FHIR OperationOutcome untuk meneruskan status code asli ke client
		var outcome map[string]interface{}
		if err := json.Unmarshal(respBody, &outcome); err == nil && outcome["resourceType"] == "OperationOutcome" {
			return nil, &ErrorOperationOutcome{
				StatusCode: resp.StatusCode,
				Outcome:    outcome,
			}
		}

		return nil, fmt.Errorf("SatuSehat Auth failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}

	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Default().Error("SatuSehat Auth Decode Error", logger.String("response", string(respBody)), logger.ErrorField(err))
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	if _, ok := result["access_token"].(string); !ok {
		return nil, fmt.Errorf("access_token missing or invalid in response: %s", string(respBody))
	}

	c.tokenData = result

	var expiresIn int
	if expRaw, ok := result["expires_in"]; ok {
		switch v := expRaw.(type) {
		case string:
			expiresIn, _ = strconv.Atoi(v)
		case float64: // Pada interface{}, JSON number dibaca sebagai float64 oleh Go
			expiresIn = int(v)
		default:
			expiresIn = 3599 // Fallback default 1 jam jika atribut tidak ada/tidak sesuai
		}
	} else {
		expiresIn = 3599
	}

	// Simpan masa berlaku token dikurangi 60 detik sebagai safety margin
	c.tokenExpiry = time.Now().Add(time.Duration(expiresIn-60) * time.Second)

	logger.Default().Info("🔑 SatuSehat Access Token refreshed successfully")

	return c.tokenData, nil
}

// RefreshToken memaksa pengambilan token baru (melewati cache).
func (c *client) RefreshToken(ctx context.Context) (map[string]interface{}, error) {
	c.mu.Lock()
	c.tokenData = nil // Invalidasi cache token
	c.mu.Unlock()
	return c.GetAccessToken(ctx)
}

// do merupakan helper HTTP request secara general ke berbagai Base URL SatuSehat
func (c *client) do(ctx context.Context, baseURL, method, endpoint string, body interface{}) ([]byte, error) {
	if !c.cfg.Enabled {
		return nil, fmt.Errorf("SatuSehat integration is disabled in configuration")
	}

	authData, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SatuSehat access token: %w", err)
	}
	token, ok := authData["access_token"].(string)
	if !ok {
		return nil, fmt.Errorf("access token is missing in auth data")
	}

	baseURL = strings.TrimRight(baseURL, "/")
	endpointPath := strings.TrimLeft(endpoint, "/")
	targetURL := fmt.Sprintf("%s/%s", baseURL, endpointPath)

	// Identifikasi dan format body berdasarkan tipenya (string/plaintext atau struct/JSON)
	var reqBody []byte
	var contentType string

	if body != nil {
		switch v := body.(type) {
		case string:
			// Jika body berupa string murni (seperti encrypted payload KYC), jangan di-marshal JSON
			reqBody = []byte(v)
			contentType = "text/plain"
		default:
			var err error
			reqBody, err = json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			if method == "PATCH" {
				contentType = "application/json-patch+json"
			} else {
				contentType = "application/json"
			}
		}
	}

	// Helper function untuk merakit ulang request (sangat berguna untuk skenario Retry)
	buildReq := func(accessToken string) (*http.Request, error) {
		var bodyReader io.Reader
		if reqBody != nil {
			bodyReader = bytes.NewBuffer(reqBody)
		}
		req, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")
		if reqBody != nil {
			req.Header.Set("Content-Type", contentType)
		}
		return req, nil
	}

	req, err := buildReq(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// [RADAR] Logging Outgoing Request
	logger.Default().Info("🛫 SatuSehat API Request Sent",
		logger.String("target_url", targetURL),
		logger.String("method", method),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	// [AUTO-RETRY LOGIC] Jika Kemenkes menjawab 401 Unauthorized, token mungkin expired/revoked
	if resp.StatusCode == http.StatusUnauthorized {
		// Tutup body request pertama yang gagal agar tidak memory leak
		resp.Body.Close()

		logger.Default().Warn("🔄 Token Satu Sehat ditolak (401). Melakukan Auto-Refresh dan mengulang request...")

		// 1. Refresh Token secara paksa
		authData, err = c.RefreshToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-refresh token for retry: %w", err)
		}
		token, ok = authData["access_token"].(string)
		if !ok {
			return nil, fmt.Errorf("access token is missing after auto-refresh")
		}

		// 2. Bangun ulang Request dengan token terbaru
		req, err = buildReq(token)
		if err != nil {
			return nil, fmt.Errorf("failed to rebuild retry request: %w", err)
		}

		// 3. Tembak ulang API Kemenkes
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http retry request failed: %w", err)
		}
	}

	// Pastikan body response dari eksekusi terakhir (berhasil atau retry) selalu ditutup
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Tangani response status code error (seperti FHIROperationOutcome)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Default().Error("SatuSehat API Error",
			logger.Int("status_code", resp.StatusCode),
			logger.String("response", string(respBody)),
		)

		// Coba parse response body sebagai FHIR OperationOutcome untuk error yang lebih terbaca
		var outcome map[string]interface{}
		if err := json.Unmarshal(respBody, &outcome); err == nil && outcome["resourceType"] == "OperationOutcome" {
			return nil, &ErrorOperationOutcome{
				StatusCode: resp.StatusCode,
				Outcome:    outcome,
			}
		}

		return nil, fmt.Errorf("SatuSehat API Error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *client) DoRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	return c.do(ctx, c.cfg.BaseURL, method, endpoint, body)
}

func (c *client) DoKFA(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	return c.do(ctx, c.cfg.KFAURL, method, endpoint, body)
}

func (c *client) DoConsent(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	return c.do(ctx, c.cfg.ConsentURL, method, endpoint, body)
}

func (c *client) DoKYC(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	return c.do(ctx, c.cfg.KYCURL, method, endpoint, body)
}

// UploadDICOM mengunggah file biner .dcm langsung ke SatuSehat STOW-RS endpoint
func (c *client) UploadDICOM(ctx context.Context, dicomBytes []byte) ([]byte, error) {
	if !c.cfg.Enabled {
		return nil, fmt.Errorf("SatuSehat integration is disabled")
	}

	authData, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SatuSehat access token: %w", err)
	}
	token, _ := authData["access_token"].(string)

	// 1. Buat body multipart
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 2. Buat header part khusus untuk DICOM (Kemenkes mewajibkan application/dicom)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "application/dicom")

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat multipart frame: %w", err)
	}

	// 3. Tulis biner DICOM ke dalam part
	if _, err := part.Write(dicomBytes); err != nil {
		return nil, fmt.Errorf("gagal menulis biner DICOM: %w", err)
	}
	writer.Close()

	// 4. Siapkan HTTP Request
	// NOTE: Content-Type untuk STOW-RS BUKAN multipart/form-data, TAPI multipart/related
	contentType := fmt.Sprintf(`multipart/related; type="application/dicom"; boundary=%s`, writer.Boundary())

	req, err := http.NewRequestWithContext(ctx, "POST", c.cfg.DicomURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create DICOM request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/dicom+json") // Format balasan Kemenkes

	logger.Default().Info("🛫 Mengirim DICOM ke SatuSehat", logger.String("target_url", c.cfg.DicomURL))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Default().Error("SatuSehat DICOM API Error",
			logger.Int("status_code", resp.StatusCode),
			logger.String("response", string(respBody)),
		)

		var outcome map[string]interface{}
		if err := json.Unmarshal(respBody, &outcome); err == nil && outcome["resourceType"] == "OperationOutcome" {
			return nil, &ErrorOperationOutcome{
				StatusCode: resp.StatusCode,
				Outcome:    outcome,
			}
		}

		return nil, fmt.Errorf("DICOM Upload Error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
