package reference

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"service/internal/infrastructure/transport/http/middleware"
	"service/internal/satusehat/reference/kyc"
	"service/pkg/crypto"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type KYCHandler struct {
	service kyc.Service
}

func NewKYCHandler(service kyc.Service) *KYCHandler {
	return &KYCHandler{
		service: service,
	}
}

// GenerateKeys godoc
//
//	@Summary		Generate Static RSA Keys for KYC
//	@Description	Menghasilkan pasangan kunci RSA 2048-bit dalam format Base64 untuk disimpan di .env (KYC_PRIVATE_KEY_B64 & KYC_PUBLIC_KEY_B64)
//	@Tags			Satu Sehat - KYC
//	@Produce		json
//	@Success		200		{object}	response.Response
//	@Router			/satusehat/reference/kyc/generate-keys [get]
func (h *KYCHandler) GenerateKeys(c *gin.Context) {
	privPEM, pubPEM, err := crypto.GenerateRSAKeyPair(2048)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat RSA Keypair", err.Error())
		return
	}

	responseData := map[string]interface{}{
		"instructions":        "Salin nilai di bawah ini dan masukkan ke dalam file .env Anda",
		"KYC_PRIVATE_KEY_B64": base64.StdEncoding.EncodeToString([]byte(privPEM)),
		"KYC_PUBLIC_KEY_B64":  base64.StdEncoding.EncodeToString([]byte(pubPEM)),
	}

	response.Success(c, http.StatusOK, "Berhasil membuat kunci RSA Statis", responseData)
}

// GenerateURL godoc
//
//	@Summary		Generate KYC URL
//	@Description	Membuat link validasi KYC Satu Sehat
//	@Tags			Satu Sehat - KYC
//	@Accept			json
//	@Produce		json
//	@Param			request	body		kyc.GenerateURLRequest	true	"Payload KYC"
//	@Success		200		{object}	response.Response
//	@Router			/satusehat/reference/kyc/generate-url [post]
//	@Security		BearerAuth
func (h *KYCHandler) GenerateURL(c *gin.Context) {
	var req kyc.GenerateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Format request tidak valid", err.Error())
		return
	}

	// NOTE: Public Key statis otomatis disisipkan oleh layer Service yang membaca dari konfigurasi .env

	data, err := h.service.GenerateURL(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}

	responseData := map[string]interface{}{
		"kyc_data": data,
	}

	response.Success(c, http.StatusOK, "Berhasil membuat link KYC", responseData)
}

// Callback godoc
//
//	@Summary		Webhook/Callback KYC
//	@Description	Menerima notifikasi dari Satu Sehat setelah proses KYC selesai
//	@Tags			Satu Sehat - KYC
//	@Accept			text/plain
//	@Produce		json
//	@Param			request	body		string	true	"Payload Webhook Terenkripsi (-----BEGIN ENCRYPTED MESSAGE-----...)"
//	@Success		200		{object}	response.Response
//	@Router			/satusehat/reference/kyc/callback [post]
func (h *KYCHandler) Callback(c *gin.Context) {
	// Ekstrak token dari header Authorization (Format: "Bearer <token>")
	authHeader := c.GetHeader("Authorization")
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

	// Payload dari webhook Kemenkes berupa string terenkripsi (bukan JSON langsung)
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Gagal membaca body request", err.Error())
		return
	}
	encryptedPayload := string(bodyBytes)

	if err := h.service.HandleCallback(c.Request.Context(), encryptedPayload, token); err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Callback berhasil diproses", nil)
}

func (h *KYCHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/reference/kyc")
	{
		group.GET("/generate-keys", h.GenerateKeys)
		group.POST("/generate-url", h.GenerateURL)
		// Rate limit khusus untuk endpoint callback: max 2 Request/Detik, Burst Limit 5
		group.POST("/callback", middleware.MemoryRateLimitMiddleware(2.0, 5), h.Callback)
	}
}
