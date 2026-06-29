package auth

import (
	"net/http"

	authService "service/internal/auth"
	"service/pkg/errors"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service authService.Service
}

func NewAuthHandler(service authService.Service) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.POST("/refresh", h.RefreshToken)
	}
}

// RegisterProtectedRoutes mendaftarkan endpoint auth yang membutuhkan token
func (h *AuthHandler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/logout", h.Logout)
		auth.GET("/info", h.TokenInfo)
	}
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Authenticate user and return JWT & Refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		authService.LoginRequest	true	"Login credentials"
//	@Success		200		{object}	response.Response
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req authService.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Pemanggilan service secara by-value
	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusOK, "Login successful", resp)
}

// Register godoc
//
//	@Summary		Register user
//	@Description	Register a new user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		authService.RegisterRequest	true	"Registration details"
//	@Success		201		{object}	response.Response
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req authService.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	resp, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", resp)
}

// RefreshToken godoc
//
//	@Summary		Refresh Token
//	@Description	Refresh expired access token using refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		authService.RefreshTokenRequest	true	"Refresh token payload"
//	@Success		200		{object}	response.Response
//	@Router			/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req authService.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	resp, err := h.service.RefreshToken(c.Request.Context(), req)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", resp)
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Logout current user and blacklist the token
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	response.Response
//	@Security		BearerAuth
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		response.Error(c, http.StatusBadRequest, "Token required", nil)
		return
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.service.Logout(c.Request.Context(), token)
	if err != nil {
		appErr := errors.FromError(err)
		response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
		return
	}

	response.Success(c, http.StatusOK, "Logged out successfully", nil)
}

// TokenInfo godoc
//
//	@Summary		Get Token Info
//	@Description	Retrieve information about the currently active token/user
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	response.Response
//	@Security		BearerAuth
//	@Router			/auth/info [get]
func (h *AuthHandler) TokenInfo(c *gin.Context) {
	// Data ini otomatis di-set oleh provider.go (UnifiedAuthMiddleware)
	authProvider := c.GetString("auth_provider")
	userID := c.GetString("user_id")
	username := c.GetString("username")
	email := c.GetString("email")
	role := c.GetString("role")
	name := c.GetString("name")

	data := gin.H{
		"auth_provider": authProvider, // Akan bernilai: "jwt", "keycloak", atau "static"
		"user_id":       userID,
		"username":      username,
		"email":         email,
		"name":          name,
		"role":          role,
	}

	response.Success(c, http.StatusOK, "Current active token information", data)
}
