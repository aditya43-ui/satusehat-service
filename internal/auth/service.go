package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"service/internal/infrastructure/cache"
	"service/internal/infrastructure/config"
	"service/pkg/errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type StoredToken struct {
	Token     string
	IsRevoked bool
	UserID    int64
	ExpiresAt int64
}

func (t *StoredToken) IsExpired() bool { return false /* implementasi */ }

// Service mendefinisikan interface untuk operasi otentikasi.
type Service interface {
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*LoginResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req RegisterRequest) (interface{}, error)
	Logout(ctx context.Context, token string) error
}

type service struct {
	cmdRepo   CommandRepository
	queryRepo QueryRepository
	cache     *cache.Manager
	config    *config.Config // Ganti keycloakClient dengan config
}

// NewService membuat instance service auth baru.
// Signature diperbarui untuk menerima config secara keseluruhan.
func NewService(cmdRepo CommandRepository, queryRepo QueryRepository, cache *cache.Manager, cfg *config.Config) Service {
	return &service{
		cmdRepo:   cmdRepo,
		queryRepo: queryRepo,
		cache:     cache,
		config:    cfg,
	}
}

// Login mengimplementasikan otentikasi menggunakan data dummy untuk keperluan testing.
func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Simulasi pengecekan kredensial dengan beberapa akun dummy
	type dummyUser struct {
		Password string
		UserID   string
		RoleID   string
		Name     string
	}

	dummyUsers := map[string]dummyUser{
		"admin@example.com":    {"password123", "1001", "admin", "Admin Dummy"},
		"user@example.com":     {"password123", "1002", "user", "User Dummy"},
		"manager@example.com":  {"password123", "1003", "manager", "Manager Dummy"},
		"worker@satusehat.com": {"password123", "1004", "worker", "Worker Dummy"},
	}

	user, exists := dummyUsers[req.Email]
	if !exists || user.Password != req.Password {
		return nil, errors.UnauthorizedError().Message("Email atau password salah. Coba: admin@example.com, user@example.com, manager@example.com (password123)").Build()
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "fallback_secret_key_change_in_production"
	}

	// Buat Access Token baru (Umur pendek, misal 15 Menit)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserID,
		"email":   req.Email,
		"role_id": user.RoleID,
		"name":    user.Name,
		"exp":     time.Now().Add(60 * time.Minute).Unix(),
	})
	newAccessTokenStr, _ := accessToken.SignedString([]byte(secret))

	// Buat Refresh Token baru (Umur panjang, misal 7 Hari)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.UserID,
		"email":   req.Email,
		"role_id": user.RoleID,
		"name":    user.Name,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	newRefreshTokenStr, _ := refreshToken.SignedString([]byte(secret))

	return &LoginResponse{
		Provider:     "jwt",
		AccessToken:  newAccessTokenStr,
		RefreshToken: newRefreshTokenStr,
		ExpiresIn:    900, // 15 menit dalam detik
		TokenType:    "Bearer",
	}, nil
}

// Register mengimplementasikan pendaftaran user dengan respons data json dummy.
func (s *service) Register(ctx context.Context, req RegisterRequest) (interface{}, error) {
	// Mengembalikan response dummy seolah-olah user berhasil dibuat di DB
	return map[string]interface{}{
		"id":         1002, // Generated Dummy ID
		"name":       req.Name,
		"email":      req.Email,
		"role_id":    req.RoleID,
		"created_at": time.Now().Format(time.RFC3339),
		"status":     "active",
	}, nil
}

// Logout mengimplementasikan proses logout dengan memasukkan token ke cache blacklist.
func (s *service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return errors.NewValidationError().Message("Token tidak valid").Build()
	}

	// Invalidate token di cache (Blacklist) jika cache manager aktif
	if s.cache != nil {
		// Set token ke dalam blacklist dengan durasi 15 menit (Sesuai umur access token)
		_ = s.cache.Set(ctx, "blacklist_token:"+token, true, 15*time.Minute)
	}

	// TODO: Segera hapus/tandai Refresh Token sebagai revoked di database agar tidak bisa
	// digunakan lagi untuk generate access token baru setelah pengguna berhasil logout.
	// _ = s.cmdRepo.RevokeAllTokensByToken(ctx, token)

	return nil
}

// RefreshToken menangani refresh token untuk berbagai provider.
func (s *service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*LoginResponse, error) {
	switch req.Provider {
	case "jwt":
		return s.refreshInternalJWT(ctx, req.RefreshToken)
	case "keycloak":
		return s.refreshKeycloakToken(ctx, req.RefreshToken)

	default:
		return nil, errors.NewValidationError().Message("Provider tidak didukung").Metadata("provider", req.Provider).Build()
	}
}

// refreshInternalJWT menangani validasi dan pembuatan JWT internal baru.
func (s *service) refreshInternalJWT(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "fallback_secret_key_change_in_production"
	}

	// Parse dan validasi refresh token
	parsedToken, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, errors.UnauthorizedError().Message("Refresh token tidak valid atau telah kedaluwarsa").Build()
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.UnauthorizedError().Message("Klaim token tidak valid").Build()
	}

	// Cek apakah ini dummy user dari fungsi Login. Jika ya, lewati pengecekan DB.
	isDummyUser := false
	if userID, ok := claims["user_id"].(string); ok && (userID == "1001" || userID == "1002" || userID == "1003") {
		isDummyUser = true
	}

	// Hanya lakukan pengecekan ke database jika bukan dummy user
	if !isDummyUser {
		existingToken, err := s.queryRepo.FindRefreshToken(ctx, refreshToken)
		if err != nil {
			// Kesalahan saat query ke database
			return nil, errors.InternalError().Message("Gagal memvalidasi refresh token").Cause(err).Build()
		}
		if existingToken == nil {
			return nil, errors.UnauthorizedError().Message("Refresh token tidak valid atau tidak ditemukan").Build()
		}
		if existingToken.IsRevoked {
			return nil, errors.UnauthorizedError().Message("Refresh token telah dicabut (dibatalkan)").Metadata("reason", "revoked").Build()
		}
	}

	// Buat Access Token baru (Umur pendek, misal 15 Menit)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": claims["user_id"],
		"email":   claims["email"],
		"role_id": claims["role_id"],
		"name":    claims["name"],
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	newAccessTokenStr, _ := accessToken.SignedString([]byte(secret))

	// Buat Refresh Token baru (Umur panjang, misal 7 Hari)
	newRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": claims["user_id"],
		"email":   claims["email"],
		"role_id": claims["role_id"],
		"name":    claims["name"],
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	newRefreshTokenStr, _ := newRefreshToken.SignedString([]byte(secret))

	// Revoke token lama dan simpan yang baru (hanya untuk user non-dummy)
	if !isDummyUser {
		_ = s.cmdRepo.RevokeAndSaveTokens(ctx, refreshToken, newRefreshTokenStr)
	}

	return &LoginResponse{
		Provider:     "jwt",
		AccessToken:  newAccessTokenStr,
		RefreshToken: newRefreshTokenStr,
		ExpiresIn:    900, // 15 menit dalam detik
		TokenType:    "Bearer",
	}, nil
}

// refreshKeycloakToken menangani refresh token melalui Keycloak.
func (s *service) refreshKeycloakToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	if !s.config.Keycloak.Enabled || s.config.Keycloak.URL == "" {
		return nil, errors.InternalError().Message("Provider Keycloak tidak dikonfigurasi dengan benar").Build()
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", s.config.Keycloak.URL, s.config.Keycloak.Realm)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", s.config.Keycloak.ClientID)
	data.Set("client_secret", s.config.Keycloak.ClientSecret)
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, errors.InternalError().Cause(err).Message("Gagal membuat request ke Keycloak").Build()
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.InternalError().Cause(err).Message("Gagal berkomunikasi dengan Keycloak").Build()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var keycloakErr map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&keycloakErr)
		return nil, errors.UnauthorizedError().
			Message("Gagal me-refresh token dengan Keycloak").
			Metadata("upstream_status", resp.StatusCode).
			Metadata("upstream_error", keycloakErr).Build()
	}

	var tokenResponse LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, errors.InternalError().Cause(err).Message("Gagal mem-parsing respons Keycloak").Build()
	}

	// Tambahkan provider ke response
	tokenResponse.Provider = "keycloak"

	return &tokenResponse, nil
}
