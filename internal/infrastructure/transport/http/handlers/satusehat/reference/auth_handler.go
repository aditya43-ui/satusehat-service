package reference

import (
	stdErrors "errors"
	"net/http"
	"service/internal/interfaces/satusehat"
	"service/internal/satusehat/reference/auth"
	"service/pkg/errors"
	"service/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthHandler menangani endpoint HTTP untuk Auth Satu Sehat.
type AuthHandler struct {
	service auth.Service
}

// RegisterRoutes mendaftarkan endpoint handler ini ke router Gin
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/reference")
	{
		group.GET("/auth/token", h.GetToken)
		group.POST("/auth/token/refresh", h.RefreshToken)
	}
}

// NewAuthHandler membuat instance baru dari AuthHandler.
func NewAuthHandler(service auth.Service) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

// handleSatuSehatError mengekstrak OperationOutcome agar JSON bisa tampil terstruktur di response
func handleSatuSehatError(c *gin.Context, err error) {
	var ssErr *satusehat.ErrorOperationOutcome
	if stdErrors.As(err, &ssErr) {
		c.JSON(ssErr.StatusCode, gin.H{
			"status":  "error",
			"message": ssErr.Outcome,
			"error": gin.H{
				"severity":  "error",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}
	appErr := errors.FromError(err)
	response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
}

// GetToken godoc
//
//	@Summary		Get SatuSehat Access Token
//	@Description	Mendapatkan token aktif untuk API Satu Sehat Kemenkes (menggunakan cache internal)
//	@Tags			Satu Sehat - Auth
//	@Produce		json
//	@Success		200	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/satusehat/reference/auth/token [get]
//	@Security		BearerAuth
func (h *AuthHandler) GetToken(c *gin.Context) {
	data, err := h.service.GetToken(c.Request.Context())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil mendapatkan token Satu Sehat", data)
}

// RefreshToken godoc
//
//	@Summary		Refresh SatuSehat Access Token
//	@Description	Memaksa request token baru dari API Satu Sehat Kemenkes (bypass cache)
//	@Tags			Satu Sehat - Auth
//	@Produce		json
//	@Success		200	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/satusehat/reference/auth/token/refresh [post]
//	@Security		BearerAuth
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	data, err := h.service.RefreshToken(c.Request.Context())
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Berhasil me-refresh token Satu Sehat", data)
}
