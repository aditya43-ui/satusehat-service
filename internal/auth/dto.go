package auth

// RefreshTokenRequest mendefinisikan struktur untuk request refresh token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	// Provider membantu membedakan antara "jwt" dan "keycloak".
	// Ini membuat endpoint menjadi fleksibel.
	Provider string `json:"provider" binding:"required,oneof=jwt keycloak"`
}

// LoginResponse adalah respons standar untuk login dan refresh token.
type LoginResponse struct {
	Provider     string `json:"provider,omitempty"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	IdToken      string `json:"id_token,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	RoleID   int64  `json:"role_id" binding:"required"`
}
