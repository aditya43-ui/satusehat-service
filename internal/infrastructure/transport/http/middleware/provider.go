package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	pkgErrors "service/pkg/errors"
	"strings"
	"sync"
	"time"

	"service/internal/infrastructure/cache"
	"service/internal/infrastructure/config"
	"service/pkg/logger"
	"service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/sync/singleflight"
)

// Definisi Error kustom untuk autentikasi
var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrMissingClaims     = errors.New("missing claims")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrInvalidIssuer     = errors.New("invalid issuer")
	ErrInvalidAudience   = errors.New("invalid audience")
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthHeader = errors.New("invalid authorization header format")
)

// JWTClaims menyimpan struktur payload token yang terekstrak
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Name     string `json:"name"`
}

// AuthProvider interface for different authentication methods
type AuthProvider interface {
	ValidateToken(tokenString string) (*JWTClaims, error)
	Name() string
}

// ProviderFactory creates authentication providers based on configuration
type ProviderFactory struct {
	config *config.Config
}

func NewProviderFactory(config *config.Config) *ProviderFactory {
	return &ProviderFactory{
		config: config,
	}
}

func (f *ProviderFactory) CreateProviders() []AuthProvider {
	var providers []AuthProvider

	reqLogger := logger.Default()
	reqLogger.Info("Creating authentication providers",
		logger.String("auth_type", f.config.Auth.Type),
		logger.Bool("keycloak_enabled", f.config.Keycloak.Enabled),
		logger.String("keycloak_issuer", f.config.Keycloak.Issuer),
		logger.Int("static_tokens_len", len(f.config.Auth.StaticTokens)),
		logger.String("fallback_to", f.config.Auth.FallbackTo),
	)

	switch f.config.Auth.Type {
	case "static":
		if len(f.config.Auth.StaticTokens) > 0 {
			providers = append(providers, NewStaticTokenProvider(f.config.Auth.StaticTokens))
		} else {
			reqLogger.Warn("No static tokens configured for static auth type", logger.String("type", "static"))
		}
	case "jwt":
		providers = append(providers, NewJWTAuthProvider())
		reqLogger.Info("JWT provider added")
	case "keycloak":
		if f.config.Keycloak.Issuer != "" {
			providers = append(providers, NewKeycloakAuthProvider(f.config))
			reqLogger.Info("Keycloak provider added")
		} else {
			reqLogger.Warn("Keycloak issuer not configured for keycloak auth type", logger.String("type", "keycloak"))
		}
	case "hybrid":
		if f.config.Keycloak.Issuer != "" {
			providers = append(providers, NewKeycloakAuthProvider(f.config))
			reqLogger.Info("Keycloak provider added for hybrid")
		} else {
			reqLogger.Warn("Keycloak issuer not configured for hybrid auth type", logger.String("type", "keycloak"))
		}
		switch f.config.Auth.FallbackTo {
		case "static":
			if len(f.config.Auth.StaticTokens) > 0 {
				providers = append(providers, NewStaticTokenProvider(f.config.Auth.StaticTokens))
			} else {
				reqLogger.Warn("No static tokens configured for hybrid fallback", logger.String("type", "static"))
			}
		case "jwt":
			providers = append(providers, NewJWTAuthProvider())
		default:
			providers = append(providers, NewJWTAuthProvider())
			reqLogger.Info("JWT fallback provider added as default")
		}
	default:
		providers = append(providers, NewJWTAuthProvider())
	}

	return providers
}

// StaticTokenProvider handles static token authentication
type StaticTokenProvider struct {
	tokens map[string]bool
}

func NewStaticTokenProvider(tokens []string) *StaticTokenProvider {
	tokenMap := make(map[string]bool)
	for _, token := range tokens {
		if token != "" {
			tokenMap[token] = true
		}
	}
	return &StaticTokenProvider{tokens: tokenMap}
}

func (s *StaticTokenProvider) ValidateToken(tokenString string) (*JWTClaims, error) {
	if !s.tokens[tokenString] {
		return nil, ErrInvalidToken
	}

	return &JWTClaims{
		UserID:   "static-user",
		Username: "static-user",
		Email:    "static@example.com",
		Role:     "user",
		Name:     "Static User",
	}, nil
}

func (s *StaticTokenProvider) Name() string {
	return "static"
}

// JWTAuthProvider handles JWT authentication
type JWTAuthProvider struct {
	secret string
}

func NewJWTAuthProvider() *JWTAuthProvider {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "fallback_secret_key_change_in_production"
	}
	return &JWTAuthProvider{secret: secret}
}

func (j *JWTAuthProvider) ValidateToken(tokenString string) (*JWTClaims, error) {
	parsedToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.secret), nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrMissingClaims
	}

	return &JWTClaims{
		UserID: fmt.Sprintf("%v", claims["user_id"]),
		Email:  fmt.Sprintf("%v", claims["email"]),
		Role:   fmt.Sprintf("%v", claims["role_id"]),
		Name:   fmt.Sprintf("%v", claims["name"]),
	}, nil
}

func (j *JWTAuthProvider) Name() string {
	return "jwt"
}

// KeycloakAuthProvider handles Keycloak JWT authentication
type KeycloakAuthProvider struct {
	jwksCache *JwksCache
	config    *config.Config
}

func NewKeycloakAuthProvider(cfg *config.Config) *KeycloakAuthProvider {
	return &KeycloakAuthProvider{
		jwksCache: NewJwksCache(cfg),
		config:    cfg,
	}
}

func (k *KeycloakAuthProvider) ValidateToken(tokenString string) (*JWTClaims, error) {
	parsedToken, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Extract claims for logging
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrMissingClaims
	}

	// Check if token is expired
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, ErrTokenExpired
		}
	}

	// Pastikan token yang diterima adalah Access Token ("Bearer"), bukan ID Token
	if typ, ok := claims["typ"].(string); ok {
		if typ != "Bearer" {
			return nil, fmt.Errorf("invalid token type: expected Bearer, got %s", typ)
		}
	}

	// Now parse with verification
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			// Dihilangkan logger Warn agar tidak menyebabkan log spam ketika mekanisme fallback JWT aktif
			return nil, ErrInvalidSignature
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid header not found")
		}

		key, err := k.jwksCache.GetKey(kid)
		if err != nil {
			return nil, err
		}
		return key, nil
	}, jwt.WithIssuer(k.config.Keycloak.Issuer))

	if err != nil {
		// Return specific error based on the error type
		if strings.Contains(err.Error(), "expired") {
			return nil, ErrTokenExpired
		} else if strings.Contains(err.Error(), "signature") {
			return nil, ErrInvalidSignature
		} else if strings.Contains(err.Error(), "issuer") {
			return nil, ErrInvalidIssuer
		}

		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok = token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrMissingClaims
	}

	// Validasi custom untuk Audience (aud) atau Authorized Party (azp)
	// Keycloak menempatkan Client ID di 'azp' untuk Access Token
	expectedAudience := k.config.Keycloak.Audience
	if expectedAudience != "" {
		validAudience := false

		// 1. Cek klaim azp (Authorized Party)
		if azp := getClaimString(claims, "azp"); azp == expectedAudience {
			validAudience = true
		}

		// 2. Cek klaim aud (Audience) jika azp tidak cocok
		if !validAudience {
			if audValue, ok := claims["aud"]; ok {
				if audList, err := extractAudience(audValue); err == nil {
					for _, a := range audList {
						if a == expectedAudience {
							validAudience = true
							break
						}
					}
				}
			}
		}

		if !validAudience {
			return nil, ErrInvalidAudience
		}
	}

	// Validate required claims
	userID := getClaimString(claims, "sub")
	if userID == "" {
		return nil, ErrMissingClaims
	}

	// Ekstraksi nested roles dari Keycloak (realm_access.roles)
	var roleStr string
	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			var roleList []string
			for _, r := range roles {
				roleList = append(roleList, fmt.Sprintf("%v", r))
			}
			roleStr = strings.Join(roleList, ",") // Menggabungkan array role menjadi string: "admin,user"
		}
	}

	return &JWTClaims{
		UserID:   userID,
		Username: getClaimString(claims, "preferred_username"),
		Email:    getClaimString(claims, "email"),
		Role:     roleStr,
		Name:     getClaimString(claims, "name"),
	}, nil
}

func (k *KeycloakAuthProvider) Name() string {
	return "keycloak"
}

// AuthMiddleware provides flexible authentication based on configuration and implements redis blacklist check
func AuthMiddleware(cfg *config.Config, cacheManager *cache.Manager) gin.HandlerFunc {
	factory := NewProviderFactory(cfg)
	providers := factory.CreateProviders()

	// Validate that we have at least one provider
	if len(providers) == 0 {
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authentication service not configured"})
		}
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrMissingAuthHeader.Error()})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidAuthHeader.Error()})
			return
		}

		tokenString := parts[1]

		// Cek Blacklist Token (User yang sudah logout dilarang menggunakan token yang sama)
		if cacheManager != nil {
			var isBlacklisted bool
			if err := cacheManager.Get(c.Request.Context(), "blacklist_token:"+tokenString, &isBlacklisted); err == nil && isBlacklisted {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked or logged out"})
				return
			}
		}

		// Coba setiap provider sampai salah satu berhasil
		var claims *JWTClaims
		var err error
		var providerName string
		providerErrorDetails := make(map[string]string)

		for _, provider := range providers {
			claims, err = provider.ValidateToken(tokenString)
			if err == nil {
				providerName = provider.Name()
				break // Berhenti jika ada yang berhasil
			}
			providerErrorDetails[provider.Name()] = err.Error()
		}

		if err != nil {
			var finalErr error

			if errors.Is(err, ErrTokenExpired) {
				finalErr = pkgErrors.UnauthorizedError().
					Code(pkgErrors.ErrCodeTokenExpired).
					Message(pkgErrors.GetLocalizedMessage(pkgErrors.ErrCodeTokenExpired, "id", "Token telah kadaluarsa")).
					Metadata("provider_errors", providerErrorDetails).Build()
			} else {
				finalErr = pkgErrors.UnauthorizedError().
					Code(pkgErrors.ErrCodeInvalidToken).
					Message(pkgErrors.GetLocalizedMessage(pkgErrors.ErrCodeInvalidToken, "id", "Token tidak valid")).
					Metadata("provider_errors", providerErrorDetails).Build()
			}

			appErr := pkgErrors.FromError(finalErr)
			response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
			c.Abort()
			return
		}

		// Set informasi pengguna di konteks
		if claims != nil {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("name", claims.Name)
			c.Set("role_id", claims.Role) // Kompatibilitas untuk handler lama
			c.Set("token", tokenString)   // Kompatibilitas untuk handler lama
			c.Set("auth_provider", providerName)
		}

		c.Next()
	}
}

// InitializeAuth initializes authentication configuration
func InitializeAuth(cfg *config.Config) {
	// This function can be used to initialize global auth settings if needed
	logger.Default().Info("Authentication initialized", logger.String("auth_type", cfg.Auth.Type))
}

// Helper functions
func getClaimString(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key]; ok && value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// extractAudience parses audience claim which can be a string or an array of strings
func extractAudience(audValue interface{}) ([]string, error) {
	switch v := audValue.(type) {
	case string:
		return []string{v}, nil
	case []interface{}:
		var auds []string
		for _, a := range v {
			if s, ok := a.(string); ok {
				auds = append(auds, s)
			}
		}
		return auds, nil
	default:
		return nil, errors.New("invalid audience format")
	}
}

// JwksCache and related functions
type JwksCache struct {
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	sfGroup   singleflight.Group
	config    *config.Config
}

func NewJwksCache(cfg *config.Config) *JwksCache {
	return &JwksCache{
		keys:   make(map[string]*rsa.PublicKey),
		config: cfg,
	}
}

func (c *JwksCache) GetKey(kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	if key, ok := c.keys[kid]; ok && time.Now().Before(c.expiresAt) {
		c.mu.RUnlock()
		return key, nil
	}
	c.mu.RUnlock()

	// Fetch keys with singleflight to avoid concurrent fetches
	v, err, _ := c.sfGroup.Do("fetch_jwks", func() (interface{}, error) {
		return c.fetchKeys()
	})
	if err != nil {
		return nil, err
	}

	keys := v.(map[string]*rsa.PublicKey)

	c.mu.Lock()
	c.keys = keys
	c.expiresAt = time.Now().Add(1 * time.Hour) // cache for 1 hour
	c.mu.Unlock()

	key, ok := keys[kid]
	if !ok {
		return nil, fmt.Errorf("key with kid %s not found", kid)
	}
	return key, nil
}

func (c *JwksCache) fetchKeys() (map[string]*rsa.PublicKey, error) {
	if c.config.Keycloak.Issuer == "" {
		return nil, fmt.Errorf("keycloak issuer is not configured")
	}

	jwksURL := c.config.Keycloak.JwksURL
	if jwksURL == "" {
		// Construct JWKS URL from issuer if not explicitly provided
		jwksURL = c.config.Keycloak.Issuer + "/protocol/openid-connect/certs"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(jwksURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: HTTP %d", resp.StatusCode)
	}

	var jwksData struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwksData); err != nil {
		return nil, err
	}

	keys := make(map[string]*rsa.PublicKey)
	for _, key := range jwksData.Keys {
		if key.Kty != "RSA" {
			continue
		}
		pubKey, err := parseRSAPublicKey(key.N, key.E)
		if err != nil {
			continue
		}
		keys[key.Kid] = pubKey
	}
	return keys, nil
}

// parseRSAPublicKey parses RSA public key components from base64url strings
func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}
