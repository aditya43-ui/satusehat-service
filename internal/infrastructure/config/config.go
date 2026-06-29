package config

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	Server       ServerConfig                `mapstructure:"server"`
	Databases    map[string]DatabaseConfig   `mapstructure:"databases"`
	ReadReplicas map[string][]DatabaseConfig `mapstructure:"read_replicas"`
	Auth         AuthConfig                  `mapstructure:"auth"`
	Keycloak     KeycloakConfig              `mapstructure:"keycloak"`
	Bpjs         BpjsConfig                  `mapstructure:"bpjs"`
	SatuSehat    SatuSehatConfig             `mapstructure:"satu_sehat"`
	Swagger      SwaggerConfig               `mapstructure:"swagger"`
	Security     SecurityConfig              `mapstructure:"security"`
	Cache        CacheConfig                 `mapstructure:"cache"`  // Tambahkan ini
	Logger       LoggerConfig                `mapstructure:"logger"` // Tambahkan ini
	Minio        MinioConfig                 `mapstructure:"minio"`
	Validator    *validator.Validate         `mapstructure:"-"`
}

type ServerConfig struct {
	REST         ServerRESTConfig `mapstructure:"rest"`
	GRPC         ServerGRPCConfig `mapstructure:"grpc"`
	Port         int              `mapstructure:"port"`
	Mode         string           `mapstructure:"mode"`
	ReadTimeout  int              `mapstructure:"read_timeout"`
	WriteTimeout int              `mapstructure:"write_timeout"`
}

type ServerRESTConfig struct {
	Enabled bool `mapstructure:"enabled" validate:"required"`
	Port    int  `mapstructure:"port" validate:"required,min=1,max=65535"`
}

type ServerGRPCConfig struct {
	Enabled bool `mapstructure:"enabled" validate:"required"`
	Port    int  `mapstructure:"port" validate:"required,min=1,max=65535"`
}

// GRPCConfig untuk pengaturan spesifik server gRPC.
type GRPCConfig struct {
	ReflectionEnabled bool          `mapstructure:"reflection_enabled"`
	Timeout           time.Duration `mapstructure:"timeout" validate:"required"`
}

type DatabaseConfig struct {
	Name            string        `mapstructure:"name"`
	Type            string        `mapstructure:"type"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Schema          string        `mapstructure:"schema"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	Path            string        `mapstructure:"path"`
	Options         string        `mapstructure:"options"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	// Security settings
	RequireSSL       bool          `mapstructure:"require_ssl"`
	SSLRootCert      string        `mapstructure:"ssl_root_cert"`
	SSLCert          string        `mapstructure:"ssl_cert"`
	SSLKey           string        `mapstructure:"ssl_key"`
	Timeout          time.Duration `mapstructure:"timeout"`
	ConnectTimeout   time.Duration `mapstructure:"connect_timeout"`
	ReadTimeout      time.Duration `mapstructure:"read_timeout"`
	WriteTimeout     time.Duration `mapstructure:"write_timeout"`
	StatementTimeout time.Duration `mapstructure:"statement_timeout"`
	// Connection pool settings
	MaxLifetime       time.Duration `mapstructure:"max_lifetime"`
	MaxIdleTime       time.Duration `mapstructure:"max_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
	IsActive          bool          `mapstructure:"is_active"`
}

type AuthConfig struct {
	Type         string   `mapstructure:"type"`
	StaticTokens []string `mapstructure:"static_tokens"`
	FallbackTo   string   `mapstructure:"fallback_to"`
}

type KeycloakConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	URL          string `mapstructure:"url"`           // URL base Keycloak, e.g., http://localhost:8080
	Realm        string `mapstructure:"realm"`         // Realm yang digunakan
	ClientID     string `mapstructure:"client_id"`     // Client ID dari service-account
	ClientSecret string `mapstructure:"client_secret"` // Client Secret
	// Fields untuk validasi token (jika service ini juga memvalidasi)
	Issuer   string `mapstructure:"issuer"`   // e.g., http://localhost:8080/realms/my-realm
	Audience string `mapstructure:"audience"` // Biasanya sama dengan ClientID
	JwksURL  string `mapstructure:"jwks_url"` // URL untuk mengambil public keys
}

type BpjsConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	KdPpk           string        `mapstructure:"kd_ppk"`
	BaseURL         string        `mapstructure:"base_url"`
	ServiceName     string        `mapstructure:"service_name"`
	ConsID          string        `mapstructure:"cons_id"`
	SecretKey       string        `mapstructure:"secret_key"`
	ApotekConsID    string        `mapstructure:"apotek_cons_id"`
	ApotekSecretKey string        `mapstructure:"apotek_secret_key"`
	UserKey         string        `mapstructure:"user_key"` // Internal active UserKey
	VclaimUserKey   string        `mapstructure:"vclaim_user_key"`
	AntrolUserKey   string        `mapstructure:"antrol_user_key"`   // Tambahan untuk Antrean RS
	ApotekUserKey   string        `mapstructure:"apotek_user_key"`   // Tambahan untuk Apotek
	AplicareUserKey string        `mapstructure:"aplicare_user_key"` // Tambahan untuk Aplicares
	IhsUserKey      string        `mapstructure:"ihs_user_key"`      // Tambahan untuk IHS / e-Rekam Medis
	Timeout         time.Duration `mapstructure:"timeout"`
}

type SatuSehatConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	OrgID            string        `mapstructure:"org_id"`
	FasyakesID       string        `mapstructure:"fasyakes_id"`
	ClientID         string        `mapstructure:"client_id"`
	ClientSecret     string        `mapstructure:"client_secret"`
	AuthURL          string        `mapstructure:"auth_url"`
	BaseURL          string        `mapstructure:"base_url"`
	ConsentURL       string        `mapstructure:"consent_url"`
	KFAURL           string        `mapstructure:"kfa_url"`
	KYCURL           string        `mapstructure:"kyc_url"`
	DicomURL         string        `mapstructure:"dicom_url"`
	KYCPublicKeyB64  string        `mapstructure:"kyc_public_key_b64"`
	KYCPrivateKeyB64 string        `mapstructure:"kyc_private_key_b64"`
	WebhookSecret    string        `mapstructure:"webhook_secret"`
	Timeout          time.Duration `mapstructure:"timeout"`
}

type SwaggerConfig struct {
	Title          string   `mapstructure:"title"`
	Description    string   `mapstructure:"description"`
	Version        string   `mapstructure:"version"`
	TermsOfService string   `mapstructure:"terms_of_service"`
	ContactName    string   `mapstructure:"contact_name"`
	ContactURL     string   `mapstructure:"contact_url"`
	ContactEmail   string   `mapstructure:"contact_email"`
	LicenseName    string   `mapstructure:"license_name"`
	LicenseURL     string   `mapstructure:"license_url"`
	Host           string   `mapstructure:"host"`
	BasePath       string   `mapstructure:"base_path"`
	Schemes        []string `mapstructure:"schemes"`
}

type SecurityConfig struct {
	TrustedOrigins []string        `mapstructure:"trusted_origins"`
	RateLimit      RateLimitConfig `mapstructure:"rate_limit"`
	MaxInputLength int             `mapstructure:"max_input_length"`
}

type RateLimitConfig struct {
	RequestsPerMinute int         `mapstructure:"requests_per_minute"`
	Redis             RedisConfig `mapstructure:"redis"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" env:"CACHE_REDIS_HOST"`
	Port     int    `mapstructure:"port" env:"CACHE_REDIS_PORT"`
	Password string `mapstructure:"password" env:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"db" env:"REDIS_DB"`
}

type CacheConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	DefaultTTL      time.Duration `mapstructure:"default_ttl"`
	SessionTTL      time.Duration `mapstructure:"session_ttl"`
	RateLimitTTL    time.Duration `mapstructure:"rate_limit_ttl"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
	MaxRetries      int           `mapstructure:"max_retries"`
	RetryDelay      time.Duration `mapstructure:"retry_delay"`
	Redis           RedisConfig   `mapstructure:"redis"`
}

type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type MinioConfig struct {
	Endpoint   string   `mapstructure:"endpoint"`
	Region     string   `mapstructure:"region"`
	AccessKey  string   `mapstructure:"access_key"`
	SecretKey  string   `mapstructure:"secret_key"`
	UseSSL     bool     `mapstructure:"use_ssl"`
	BucketName []string `mapstructure:"bucket_name"`
}

func (cfg BpjsConfig) SetHeader() (string, string, string, string, string) {
	timenow := time.Now().UTC()
	t, err := time.Parse(time.RFC3339, "1970-01-01T00:00:00Z")
	if err != nil {
		log.Fatal(err)
	}

	tstamp := timenow.Unix() - t.Unix()
	secret := []byte(cfg.SecretKey)
	message := []byte(cfg.ConsID + "&" + fmt.Sprint(tstamp))
	hash := hmac.New(sha256.New, secret)
	hash.Write(message)

	// to base64
	xSignature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// Gunakan Active UserKey alih-alih selalu menembak VClaim
	return cfg.ConsID, cfg.SecretKey, cfg.UserKey, fmt.Sprint(tstamp), xSignature
}

type ConfigBpjs struct {
	Cons_id    string
	Secret_key string
	User_key   string
}

// SetHeader for backward compatibility
func (cfg ConfigBpjs) SetHeader() (string, string, string, string, string) {
	bpjsConfig := BpjsConfig{
		ConsID:    cfg.Cons_id,
		SecretKey: cfg.Secret_key,
		UserKey:   cfg.User_key,
	}
	return bpjsConfig.SetHeader()
}

func LoadConfig() *Config {
	config := &Config{
		Server: ServerConfig{
			// Baca dari env untuk konfigurasi umum
			Port:         getEnvAsInt("SERVER_PORT", 8080), // Fallback jika REST port tidak set
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 10),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 10),

			REST: ServerRESTConfig{
				Enabled: getEnvAsBool("SERVER_REST_ENABLED", true),
				Port:    getEnvAsInt("SERVER_REST_PORT", 8080),
			},
			GRPC: ServerGRPCConfig{
				Enabled: getEnvAsBool("SERVER_GRPC_ENABLED", false),
				Port:    getEnvAsInt("SERVER_GRPC_PORT", 50051),
			},
		},
		Databases:    make(map[string]DatabaseConfig),
		ReadReplicas: make(map[string][]DatabaseConfig),
		Auth:         loadAuthConfig(),
		Keycloak:     loadKeycloakConfig(),

		// ... (Config BPJS, SatuSehat, dll tetap sama seperti kode asli Anda)
		Bpjs: BpjsConfig{
			Enabled:         getEnvAsBool("BPJS_ENABLED", false),
			KdPpk:           strings.TrimSpace(getEnv("BPJS_KDPPK", "")),
			BaseURL:         strings.TrimSpace(getEnv("BPJS_BASEURL", "https://apijkn.bpjs-kesehatan.go.id")),
			ServiceName:     strings.TrimSpace(getEnv("BPJS_SERVICE_NAME", "")),
			ConsID:          strings.TrimSpace(getEnv("BPJS_CONSID", "")),
			SecretKey:       strings.TrimSpace(getEnv("BPJS_SECRETKEY", "")),
			ApotekConsID:    strings.TrimSpace(getEnv("BPJS_APOTEK_CONSID", "")),
			ApotekSecretKey: strings.TrimSpace(getEnv("BPJS_APOTEK_SECRETKEY", "")),
			VclaimUserKey:   strings.TrimSpace(getEnv("BPJS_VCLAIM_USERKEY", getEnv("BPJS_USERKEY", ""))),
			AntrolUserKey:   strings.TrimSpace(getEnv("BPJS_ANTROL_USERKEY", "")),
			ApotekUserKey:   strings.TrimSpace(getEnv("BPJS_APOTEK_USERKEY", "")),
			AplicareUserKey: strings.TrimSpace(getEnv("BPJS_APLICARE_USERKEY", "")),
			IhsUserKey:      strings.TrimSpace(getEnv("BPJS_IHS_USERKEY", "")),
			Timeout:         parseDuration(getEnv("BPJS_TIMEOUT", "30s")),
		},
		SatuSehat: SatuSehatConfig{
			Enabled:          getEnvAsBool("SATUSEHAT_ENABLED", getEnvAsBool("SATU_SEHAT_ENABLED", false)),
			OrgID:            getEnv("SATUSEHAT_ORG_ID", getEnv("BRIDGING_SATUSEHAT_ORG_ID", "")),
			FasyakesID:       getEnv("SATUSEHAT_FASYAKES_ID", getEnv("BRIDGING_SATUSEHAT_FASYAKES_ID", "")),
			ClientID:         getEnv("SATUSEHAT_CLIENT_ID", getEnv("BRIDGING_SATUSEHAT_CLIENT_ID", "")),
			ClientSecret:     getEnv("SATUSEHAT_CLIENT_SECRET", getEnv("BRIDGING_SATUSEHAT_CLIENT_SECRET", "")),
			AuthURL:          getEnv("SATUSEHAT_AUTH_URL", getEnv("BRIDGING_SATUSEHAT_AUTH_URL", "https://api-satusehat.kemkes.go.id/oauth2/v1")),
			BaseURL:          getEnv("SATUSEHAT_BASE_URL", getEnv("BRIDGING_SATUSEHAT_BASE_URL", "https://api-satusehat.kemkes.go.id/fhir-r4/v1")),
			ConsentURL:       getEnv("SATUSEHAT_CONSENT_URL", getEnv("BRIDGING_SATUSEHAT_CONSENT_URL", "https://api-satusehat.dto.kemkes.go.id/consent/v1")),
			KFAURL:           getEnv("SATUSEHAT_KFA_URL", getEnv("BRIDGING_SATUSEHAT_KFA_URL", "https://api-satusehat.kemkes.go.id/kfa-v2")),
			KYCURL:           getEnv("SATUSEHAT_KYC_URL", getEnv("BRIDGING_SATUSEHAT_KYC_URL", "https://api-satusehat.dto.kemkes.go.id/kyc/v1")),
			DicomURL:         getEnv("SATUSEHAT_DICOM_URL", getEnv("BRIDGING_SATUSEHAT_DICOM_URL", "https://api-satusehat.kemkes.go.id/dicom/v1/dicomWeb/studies")),
			KYCPublicKeyB64:  getEnv("KYC_PUBLIC_KEY_B64", ""),
			KYCPrivateKeyB64: getEnv("KYC_PRIVATE_KEY_B64", ""),
			WebhookSecret:    getEnv("SATUSEHAT_WEBHOOK_SECRET", getEnv("BRIDGING_SATUSEHAT_WEBHOOK_SECRET", "")),
			Timeout:          parseDuration(getEnv("SATUSEHAT_TIMEOUT", getEnv("BRIDGING_SATUSEHAT_TIMEOUT", "30s"))),
		},
		Swagger: SwaggerConfig{
			Title:          getEnv("SWAGGER_TITLE", "SERVICE API"),
			Description:    getEnv("SWAGGER_DESCRIPTION", "CUSTUM SERVICE API"),
			Version:        getEnv("SWAGGER_VERSION", "1.0.0"),
			TermsOfService: getEnv("SWAGGER_TERMS_OF_SERVICE", "http://swagger.io/terms/"),
			ContactName:    getEnv("SWAGGER_CONTACT_NAME", "API Support"),
			ContactURL:     getEnv("SWAGGER_CONTACT_URL", "http://rssa.example.com/support"),
			ContactEmail:   getEnv("SWAGGER_CONTACT_EMAIL", "support@swagger.io"),
			LicenseName:    getEnv("SWAGGER_LICENSE_NAME", "Apache 2.0"),
			LicenseURL:     getEnv("SWAGGER_LICENSE_URL", "http://www.apache.org/licenses/LICENSE-2.0.html"),
			Host:           getEnv("SWAGGER_HOST", "localhost:8080"),
			BasePath:       getEnv("SWAGGER_BASE_PATH", "/api/v1"),
			Schemes:        parseSchemes(getEnv("SWAGGER_SCHEMES", "http,https")),
		},
		Security: SecurityConfig{
			TrustedOrigins: parseOrigins(getEnv("SECURITY_TRUSTED_ORIGINS", "http://localhost:3000,http://localhost:8080")),
			MaxInputLength: getEnvAsInt("SECURITY_MAX_INPUT_LENGTH", 500),
			RateLimit: RateLimitConfig{
				RequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
				Redis: RedisConfig{
					Host:     getEnv("CACHE_REDIS_HOST", "localhost"),
					Port:     getEnvAsInt("CACHE_REDIS_PORT", 6379),
					Password: getEnv("REDIS_PASSWORD", ""),
					DB:       getEnvAsInt("CACHE_REDIS_DB", 0),
				},
			},
		},
		Cache: CacheConfig{
			Enabled:         getEnvAsBool("CACHE_ENABLED", true),
			DefaultTTL:      parseDuration(getEnv("CACHE_DEFAULT_TTL", "1h")),
			SessionTTL:      parseDuration(getEnv("CACHE_SESSION_TTL", "24h")),
			RateLimitTTL:    parseDuration(getEnv("CACHE_RATE_LIMIT_TTL", "1m")),
			CleanupInterval: parseDuration(getEnv("CACHE_CLEANUP_INTERVAL", "5m")),
			MaxRetries:      getEnvAsInt("CACHE_MAX_RETRIES", 3),
			RetryDelay:      parseDuration(getEnv("CACHE_RETRY_DELAY", "100ms")),
			Redis: RedisConfig{
				Host:     getEnv("CACHE_REDIS_HOST", "localhost"),
				Port:     getEnvAsInt("CACHE_REDIS_PORT", 6379),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       getEnvAsInt("CACHE_REDIS_DB", 1),
			},
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOGGER_LEVEL", "info"),
			Format: getEnv("LOGGER_FORMAT", "text"),
		},
		Minio: MinioConfig{
			Endpoint:   getEnv("MINIO_ENDPOINT", ""),
			Region:     getEnv("MINIO_REGION", ""),
			AccessKey:  getEnv("MINIO_ACCESSKEY", ""),
			SecretKey:  getEnv("MINIO_SECRETKEY", ""),
			UseSSL:     getEnvAsBool("MINIO_USESSL", false),
			BucketName: parseStaticTokens(getEnv("MINIO_BUCKETNAME", "")), // Pisahkan berdasarkan koma
		},
	}
	// Initialize validator
	config.Validator = validator.New()

	// Load database configurations
	config.loadDatabaseConfigs()

	// Load read replica configurations
	config.loadReadReplicaConfigs()

	return config
}

func (c *Config) Validate() error {
	var errs []string

	if len(c.Databases) == 0 {
		errs = append(errs, "at least one database configuration is required")
	}

	for name, db := range c.Databases {
		if db.Type != "sqlite" && db.Host == "" {
			errs = append(errs, fmt.Sprintf("database host is required for %s", name))
		}
		if db.Type != "sqlite" && db.Username == "" {
			errs = append(errs, fmt.Sprintf("database username is required for %s", name))
		}
		// Opsional: Matikan validasi password mutlak jika Anda menggunakan database lokal tanpa password
		// if db.Type != "sqlite" && db.Password == "" {
		// 	errs = append(errs, fmt.Sprintf("database password is required for %s", name))
		// }
		if db.Type == "sqlite" && db.Path == "" {
			errs = append(errs, fmt.Sprintf("database path is required for SQLite database %s", name))
		}
		if db.Type != "sqlite" && db.Database == "" {
			errs = append(errs, fmt.Sprintf("database name is required for %s", name))
		}
	}

	if c.Bpjs.Enabled {
		if c.Bpjs.BaseURL == "" {
			errs = append(errs, "BPJS Base URL is required when BPJS is enabled")
		}
		if c.Bpjs.ConsID == "" && c.Bpjs.ApotekConsID == "" {
			errs = append(errs, "At least one BPJS Consumer ID (Default or Apotek) is required when BPJS is enabled")
		}
		if c.Bpjs.SecretKey == "" && c.Bpjs.ApotekSecretKey == "" {
			errs = append(errs, "At least one BPJS Secret Key (Default or Apotek) is required when BPJS is enabled")
		}
		if c.Bpjs.ServiceName == "" {
			errs = append(errs, "BPJS Service Name is required when BPJS is enabled")
		}
	}

	// Validate authentication configuration (auth.type is already loaded from env)
	switch c.Auth.Type {
	case "keycloak":
		if !c.Keycloak.Enabled {
			errs = append(errs, "keycloak.enabled must be true when auth.type is 'keycloak'")
		}
		if c.Keycloak.Issuer == "" {
			errs = append(errs, "keycloak.issuer is required when auth.type is 'keycloak'")
		}
		if c.Keycloak.Audience == "" {
			errs = append(errs, "keycloak.audience is required when auth.type is 'keycloak'")
		}
		if c.Keycloak.JwksURL == "" {
			errs = append(errs, "keycloak.jwks_url is required when auth.type is 'keycloak'")
		}
	case "static":
		if len(c.Auth.StaticTokens) == 0 {
			errs = append(errs, "auth.static_tokens is required when auth.type is 'static'")
		}
	case "hybrid":
		if c.Auth.FallbackTo == "" {
			errs = append(errs, "auth.fallback_to is required when auth.type is 'hybrid'")
		}
		// Validate fallback configuration
		switch c.Auth.FallbackTo {
		case "keycloak":
			if !c.Keycloak.Enabled {
				errs = append(errs, "keycloak.enabled must be true when auth.fallback_to is 'keycloak'")
			}
		case "static":
			if len(c.Auth.StaticTokens) == 0 {
				errs = append(errs, "auth.static_tokens is required when auth.fallback_to is 'static'")
			}
		}
	}

	// Legacy validation for backward compatibility
	if c.Auth.Type != "keycloak" && c.Keycloak.Enabled {
		if c.Keycloak.Issuer == "" {
			errs = append(errs, "Keycloak issuer is required when Keycloak is enabled")
		}
		if c.Keycloak.Audience == "" {
			errs = append(errs, "Keycloak audience is required when Keycloak is enabled")
		}
		if c.Keycloak.JwksURL == "" {
			errs = append(errs, "Keycloak JWKS URL is required when Keycloak is enabled")
		}
	}

	// Validate SatuSehat configuration
	if c.SatuSehat.Enabled {
		if c.SatuSehat.OrgID == "" {
			errs = append(errs, "SatuSehat Organization ID is required when SatuSehat is enabled")
		}
		if c.SatuSehat.FasyakesID == "" {
			errs = append(errs, "SatuSehat Fasyankes ID is required when SatuSehat is enabled")
		}
		if c.SatuSehat.ClientID == "" {
			errs = append(errs, "SatuSehat Client ID is required when SatuSehat is enabled")
		}
		if c.SatuSehat.ClientSecret == "" {
			errs = append(errs, "SatuSehat Client Secret is required when SatuSehat is enabled")
		}
		if c.SatuSehat.AuthURL == "" {
			errs = append(errs, "SatuSehat Auth URL is required when SatuSehat is enabled")
		}
		if c.SatuSehat.BaseURL == "" {
			errs = append(errs, "SatuSehat Base URL is required when SatuSehat is enabled")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Additional helper functions for configuration validation
func (c *Config) IsJWTAuthEnabled() bool {
	return c.Auth.Type == "jwt"
}

func (c *Config) IsKeycloakAuthEnabled() bool {
	return c.Auth.Type == "keycloak" && c.Keycloak.Enabled
}

func (c *Config) IsStaticAuthEnabled() bool {
	return c.Auth.Type == "static"
}

func (c *Config) IsHybridAuthEnabled() bool {
	return c.Auth.Type == "hybrid"
}

func (c *Config) IsBPJSEnabled() bool {
	hasCons := c.Bpjs.ConsID != "" || c.Bpjs.ApotekConsID != ""
	return c.Bpjs.Enabled && c.Bpjs.BaseURL != "" && c.Bpjs.ServiceName != "" && hasCons
}

func (c *Config) IsSatuSehatEnabled() bool {
	return c.SatuSehat.Enabled && c.SatuSehat.BaseURL != "" && c.SatuSehat.ClientID != "" && c.SatuSehat.ClientSecret != ""
}

func (c *Config) IsProductionMode() bool {
	return strings.ToLower(c.Server.Mode) == "release"
}

func (c *Config) IsDevelopmentMode() bool {
	return strings.ToLower(c.Server.Mode) == "debug" || strings.ToLower(c.Server.Mode) == "test"
}

func (c *Config) GetLogLevel() string {
	if c.IsProductionMode() {
		return "info"
	}
	return "debug"
}

func (c *Config) ShouldUseHTTPS() bool {
	return c.IsProductionMode() || strings.Contains(strings.ToLower(c.Server.Mode), "prod")
}

func (c *Config) GetServerPorts() (int, int) {
	return c.Server.REST.Port, c.Server.GRPC.Port
}

func (c *Config) GetMaxUploadSize() int64 {
	sizeStr := os.Getenv("MAX_UPLOAD_SIZE")
	if sizeStr == "" {
		sizeStr = "10485760" // 10MB default
	}
	if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return size
	}
	return 10485760
}

func (c *Config) GetRateLimitConfig() (int, int) {
	requestsPerMinute := c.Security.RateLimit.RequestsPerMinute
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60 // Default 60 requests per minute
	}
	burstSize := 10 // Default burst size
	return requestsPerMinute, burstSize
}

func (c *Config) GetTrustedOrigins() []string {
	origins := make([]string, len(c.Security.TrustedOrigins))
	for i, origin := range c.Security.TrustedOrigins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

func (c *Config) IsOriginTrusted(origin string) bool {
	for _, trusted := range c.GetTrustedOrigins() {
		if origin == trusted {
			return true
		}
	}
	return false
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func parseDuration(durationStr string) time.Duration {
	if duration, err := time.ParseDuration(durationStr); err == nil {
		return duration
	}
	return 5 * time.Minute
}

func parseSchemes(schemesStr string) []string {
	if schemesStr == "" {
		return []string{"http"}
	}

	schemes := strings.Split(schemesStr, ",")
	for i, scheme := range schemes {
		schemes[i] = strings.TrimSpace(scheme)
	}
	return schemes
}

func parseStaticTokens(tokensStr string) []string {
	if tokensStr == "" {
		return []string{}
	}

	tokens := strings.Split(tokensStr, ",")
	for i, token := range tokens {
		tokens[i] = strings.TrimSpace(token)
		// Remove empty tokens
		if tokens[i] == "" {
			tokens = append(tokens[:i], tokens[i+1:]...)
			i--
		}
	}
	return tokens
}

func parseOrigins(originsStr string) []string {
	if originsStr == "" {
		return []string{"http://localhost:8080"} // Default untuk pengembangan
	}
	origins := strings.Split(originsStr, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

// Additional database loading functions would go here...
func (c *Config) loadDatabaseConfigs() {
	// Load PostgreSQL configurations
	c.addPostgreSQLConfigs()

	// Load MySQL configurations
	c.addMySQLConfigs()

	// Load MongoDB configurations
	c.addMongoDBConfigs()

	// Load SQLite configurations
	c.addSQLiteConfigs()
	// Load SQL Server configurations
	c.addSQLServerConfigs()
	// Load custom database configurations from environment variables
	c.loadCustomDatabaseConfigs()

	// Remove duplicate database configurations
	c.removeDuplicateDatabases()
}

func (c *Config) addPostgreSQLConfigs() {
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]

		// Format yang diharapkan: POSTGRES_[NAMA_DB]_[PROPERTY]
		// Contoh: POSTGRES_DEFAULT_HOST
		if strings.HasPrefix(key, "POSTGRES_") && strings.Count(key, "_") >= 2 {
			segments := strings.Split(key, "_")
			// segments[0] = "POSTGRES"
			// segments[1...n-1] = Nama Database
			// segments[n] = Property (HOST, PORT, dll)

			dbName := strings.ToLower(strings.Join(segments[1:len(segments)-1], "_"))
			property := strings.ToLower(segments[len(segments)-1])
			value := parts[1]

			// Ambil config yang sudah ada, atau buat baru
			dbCfg, exists := c.Databases[dbName]
			if !exists {
				dbCfg = DatabaseConfig{
					Name:            dbName,
					Type:            "postgres",
					Host:            getEnv("POSTGRES_"+strings.ToUpper(dbName)+"_HOST", "localhost"),
					Port:            getEnvAsInt("POSTGRES_"+strings.ToUpper(dbName)+"_PORT", 5432),
					Username:        getEnv("POSTGRES_"+strings.ToUpper(dbName)+"_USERNAME", ""),
					Password:        getEnv("POSTGRES_"+strings.ToUpper(dbName)+"_PASSWORD", ""),
					Database:        getEnv("POSTGRES_"+strings.ToUpper(dbName)+"_DATABASE", dbName),
					Schema:          getEnv("POSTGRES_"+strings.ToUpper(dbName)+"_SCHEMA", "public"),
					SSLMode:         getEnv("POSTGRES_"+strings.ToUpper(dbName)+"_SSLMODE", "disable"),
					MaxOpenConns:    25,
					MaxIdleConns:    25,
					ConnMaxLifetime: 5 * time.Minute,
					IsActive:        true,
				}
			}

			// Logic update dinamis berdasarkan property yang sedang dibaca
			// Ini penting agar Host, Port, dan Password terisi dengan benar
			switch property {
			case "host":
				dbCfg.Host = value
			case "port":
				if port, err := strconv.Atoi(value); err == nil {
					dbCfg.Port = port
				}
			case "username":
				dbCfg.Username = value
			case "password":
				dbCfg.Password = value
			case "database":
				dbCfg.Database = value
			case "schema":
				dbCfg.Schema = value
			case "sslmode":
				dbCfg.SSLMode = value
			}

			c.Databases[dbName] = dbCfg
		}
	}
}

// Additional helper functions would be implemented here...
func (c *Config) addMySQLConfigs() {
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		if strings.HasPrefix(key, "MYSQL_") && strings.Count(key, "_") >= 2 {
			segments := strings.Split(key, "_")
			dbName := strings.ToLower(strings.Join(segments[1:len(segments)-1], "_"))
			property := strings.ToLower(segments[len(segments)-1])
			value := parts[1]

			dbCfg, exists := c.Databases[dbName]
			if !exists {
				dbCfg = DatabaseConfig{
					Name:            dbName,
					Type:            "mysql",
					Host:            getEnv("MYSQL_"+strings.ToUpper(dbName)+"_HOST", "localhost"),
					Port:            getEnvAsInt("MYSQL_"+strings.ToUpper(dbName)+"_PORT", 3306),
					Username:        getEnv("MYSQL_"+strings.ToUpper(dbName)+"_USERNAME", ""),
					Password:        getEnv("MYSQL_"+strings.ToUpper(dbName)+"_PASSWORD", ""),
					Database:        getEnv("MYSQL_"+strings.ToUpper(dbName)+"_DATABASE", dbName),
					MaxOpenConns:    25,
					MaxIdleConns:    25,
					ConnMaxLifetime: 5 * time.Minute,
					IsActive:        true,
				}
			}

			switch property {
			case "host":
				dbCfg.Host = value
			case "port":
				if port, err := strconv.Atoi(value); err == nil {
					dbCfg.Port = port
				}
			case "username":
				dbCfg.Username = value
			case "password":
				dbCfg.Password = value
			case "database":
				dbCfg.Database = value
			}
			c.Databases[dbName] = dbCfg
		}
	}
}

func (c *Config) addMongoDBConfigs() {
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		if strings.HasPrefix(key, "MONGODB_") && strings.Count(key, "_") >= 2 {
			segments := strings.Split(key, "_")
			dbName := strings.ToLower(strings.Join(segments[1:len(segments)-1], "_"))
			property := strings.ToLower(segments[len(segments)-1])
			value := parts[1]

			dbCfg, exists := c.Databases[dbName]
			if !exists {
				dbCfg = DatabaseConfig{
					Name:     dbName,
					Type:     "mongodb",
					Host:     getEnv("MONGODB_"+strings.ToUpper(dbName)+"_HOST", "localhost"),
					Port:     getEnvAsInt("MONGODB_"+strings.ToUpper(dbName)+"_PORT", 27017),
					Username: getEnv("MONGODB_"+strings.ToUpper(dbName)+"_USERNAME", ""),
					Password: getEnv("MONGODB_"+strings.ToUpper(dbName)+"_PASSWORD", ""),
					Database: getEnv("MONGODB_"+strings.ToUpper(dbName)+"_DATABASE", dbName),
					IsActive: true,
				}
			}

			switch property {
			case "host":
				dbCfg.Host = value
			case "port":
				if port, err := strconv.Atoi(value); err == nil {
					dbCfg.Port = port
				}
			case "username":
				dbCfg.Username = value
			case "password":
				dbCfg.Password = value
			case "database":
				dbCfg.Database = value
			}
			c.Databases[dbName] = dbCfg
		}
	}
}

func (c *Config) addSQLiteConfigs() {
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		if strings.HasPrefix(key, "SQLITE_") && strings.Count(key, "_") >= 2 {
			segments := strings.Split(key, "_")
			dbName := strings.ToLower(strings.Join(segments[1:len(segments)-1], "_"))
			property := strings.ToLower(segments[len(segments)-1])
			value := parts[1]

			dbCfg, exists := c.Databases[dbName]
			if !exists {
				dbCfg = DatabaseConfig{
					Name:     dbName,
					Type:     "sqlite",
					Path:     getEnv("SQLITE_"+strings.ToUpper(dbName)+"_PATH", "./"+dbName+".db"),
					IsActive: true,
				}
			}

			if property == "path" {
				dbCfg.Path = value
			}
			c.Databases[dbName] = dbCfg
		}
	}
}

func (c *Config) addSQLServerConfigs() {
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]

		// Format yang diharapkan: SQLSERVER_[NAMA_DB]_[PROPERTY]
		// Contoh: SQLSERVER_FARMASI_HOST
		if strings.HasPrefix(key, "SQLSERVER_") && strings.Count(key, "_") >= 2 {
			segments := strings.Split(key, "_")
			// segments[0] = "SQLSERVER"
			// segments[1...n-1] = Nama Database
			// segments[n] = Property (HOST, PORT, dll)

			dbName := strings.ToLower(strings.Join(segments[1:len(segments)-1], "_"))
			property := strings.ToLower(segments[len(segments)-1])
			value := parts[1]

			// Ambil config yang sudah ada, atau buat baru
			dbCfg, exists := c.Databases[dbName]
			if !exists {
				dbCfg = DatabaseConfig{
					Name:            dbName,
					Type:            "sqlserver",
					Host:            getEnv("SQLSERVER_"+strings.ToUpper(dbName)+"_HOST", "localhost"),
					Port:            getEnvAsInt("SQLSERVER_"+strings.ToUpper(dbName)+"_PORT", 1433),
					Username:        getEnv("SQLSERVER_"+strings.ToUpper(dbName)+"_USERNAME", ""),
					Password:        getEnv("SQLSERVER_"+strings.ToUpper(dbName)+"_PASSWORD", ""),
					Database:        getEnv("SQLSERVER_"+strings.ToUpper(dbName)+"_DATABASE", dbName),
					SSLMode:         getEnv("SQLSERVER_"+strings.ToUpper(dbName)+"_SSLMODE", "disable"),
					MaxOpenConns:    25,
					MaxIdleConns:    25,
					ConnMaxLifetime: 5 * time.Minute,
				}
			}

			// Update field berdasarkan property
			switch property {
			case "host":
				dbCfg.Host = value
			case "port":
				if port, err := strconv.Atoi(value); err == nil {
					dbCfg.Port = port
				}
			case "username":
				dbCfg.Username = value
			case "password":
				dbCfg.Password = value
			case "database":
				dbCfg.Database = value
			case "sslmode":
				dbCfg.SSLMode = value
			case "timeout":
				dbCfg.Timeout = parseDuration(value)
			case "max_open_conns":
				if conns, err := strconv.Atoi(value); err == nil {
					dbCfg.MaxOpenConns = conns
				}
			case "max_idle_conns":
				if conns, err := strconv.Atoi(value); err == nil {
					dbCfg.MaxIdleConns = conns
				}
			case "conn_max_lifetime":
				dbCfg.ConnMaxLifetime = parseDuration(value)
			}

			c.Databases[dbName] = dbCfg
		}
	}
}

func (c *Config) loadCustomDatabaseConfigs() {
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		// Format yang diharapkan: DB_[NAMA_DB]_[PROPERTY]
		if strings.HasPrefix(key, "DB_") && strings.Count(key, "_") >= 2 {
			segments := strings.Split(key, "_")
			dbName := strings.ToLower(strings.Join(segments[1:len(segments)-1], "_"))
			property := strings.ToLower(segments[len(segments)-1])
			value := parts[1]

			dbCfg, exists := c.Databases[dbName]
			if !exists {
				dbCfg = DatabaseConfig{
					Name:            dbName,
					MaxOpenConns:    25,
					MaxIdleConns:    25,
					ConnMaxLifetime: 5 * time.Minute,
					IsActive:        true,
				}
			}

			switch property {
			case "type":
				dbCfg.Type = value
			case "host":
				dbCfg.Host = value
			case "port":
				if port, err := strconv.Atoi(value); err == nil {
					dbCfg.Port = port
				}
			case "username":
				dbCfg.Username = value
			case "password":
				dbCfg.Password = value
			case "database":
				dbCfg.Database = value
			case "schema":
				dbCfg.Schema = value
			case "path":
				dbCfg.Path = value
			case "sslmode":
				dbCfg.SSLMode = value
			}
			c.Databases[dbName] = dbCfg
		}
	}
}

func (c *Config) removeDuplicateDatabases() {
	// Menghapus konfigurasi database yang tidak valid / kosong
	for name, db := range c.Databases {
		if db.Type == "" {
			delete(c.Databases, name)
			continue
		}
		// SQLite tidak butuh Host, sisanya butuh
		if db.Type != "sqlite" && db.Host == "" {
			delete(c.Databases, name)
			continue
		}
	}
}

func (c *Config) loadReadReplicaConfigs() {
	envVars := os.Environ()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		// Format: REPLICA_[DBNAME]_[INDEX]_[PROPERTY], cth: REPLICA_DEFAULT_0_HOST
		if strings.HasPrefix(key, "REPLICA_") && strings.Count(key, "_") >= 3 {
			segments := strings.Split(key, "_")
			property := strings.ToLower(segments[len(segments)-1])
			indexStr := segments[len(segments)-2]
			dbName := strings.ToLower(strings.Join(segments[1:len(segments)-2], "_"))
			value := parts[1]

			index, err := strconv.Atoi(indexStr)
			if err != nil {
				continue
			}

			replicas := c.ReadReplicas[dbName]
			for len(replicas) <= index {
				// Turunkan pengaturan dasar dari database master (seperti port, user)
				baseCfg := c.Databases[dbName]
				baseCfg.Name = fmt.Sprintf("%s-replica-%d", dbName, len(replicas))
				replicas = append(replicas, baseCfg)
			}

			replicaCfg := replicas[index]
			switch property {
			case "host":
				replicaCfg.Host = value
			case "port":
				if p, err := strconv.Atoi(value); err == nil {
					replicaCfg.Port = p
				}
			case "username":
				replicaCfg.Username = value
			case "password":
				replicaCfg.Password = value
			case "database":
				replicaCfg.Database = value
			}

			replicas[index] = replicaCfg
			c.ReadReplicas[dbName] = replicas
		}
	}
}

func loadAuthConfig() AuthConfig {
	return AuthConfig{
		Type:         getEnv("AUTH_TYPE", "jwt"),
		FallbackTo:   getEnv("AUTH_FALLBACK_TO", "jwt"),
		StaticTokens: parseStaticTokens(getEnv("AUTH_STATIC_TOKENS", "")),
	}
}

func loadKeycloakConfig() KeycloakConfig {
	return KeycloakConfig{
		Enabled:      getEnvAsBool("KEYCLOAK_ENABLED", false),
		URL:          getEnv("KEYCLOAK_URL", ""),
		Realm:        getEnv("KEYCLOAK_REALM", ""),
		ClientID:     getEnv("KEYCLOAK_CLIENT_ID", ""),
		ClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", ""),
		Issuer:       getEnv("KEYCLOAK_ISSUER", ""),
		Audience:     getEnv("KEYCLOAK_AUDIENCE", ""),
		JwksURL:      getEnv("KEYCLOAK_JWKS_URL", ""),
	}
}
