package bpjs

import (
	"context"
	"net/http"
	"time"

	"service/internal/infrastructure/config"
)

// VClaimService interface for VClaim operations
type VClaimService interface {
	// Basic HTTP methods
	Get(ctx context.Context, endpoint string, result interface{}) error
	Post(ctx context.Context, endpoint string, payload interface{}, result interface{}) error
	Put(ctx context.Context, endpoint string, payload interface{}, result interface{}) error
	Patch(ctx context.Context, endpoint string, payload interface{}, result interface{}) error
	Delete(ctx context.Context, endpoint string, result interface{}) error

	// Raw response methods
	GetRawResponse(ctx context.Context, endpoint string) (*ResponseDTO, error)
	PostRawResponse(ctx context.Context, endpoint string, payload interface{}) (*ResponseDTO, error)
	PutRawResponse(ctx context.Context, endpoint string, payload interface{}) (*ResponseDTO, error)
	PatchRawResponse(ctx context.Context, endpoint string, payload interface{}) (*ResponseDTO, error)
	DeleteRawResponse(ctx context.Context, endpoint string) (*ResponseDTO, error)

	// Configuration methods
	SetHTTPClient(client *http.Client)
	GetConfig() config.BpjsConfig
}

// Service struct for VClaim service
type Service struct {
	config     config.BpjsConfig
	httpClient *http.Client
}

// Response structures
type RawResponseDTO struct {
	MetaData struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"metaData"`
	Response string `json:"response"`
}

type ResponseDTO struct {
	MetaData struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"metaData"`
	Response interface{} `json:"response"`
}

// Request headers structure
type RequestHeaders struct {
	ConsID     string
	SecretKey  string
	UserKey    string
	Timestamp  string
	Signature  string
	AuthHeader string
}

// Decryption result structure
type DecryptionResult struct {
	Success bool
	Data    string
	Error   error
	Method  string
}

// Decompression result structure
type DecompressionResult struct {
	Success bool
	Data    string
	Error   error
	Method  string
}

// Crypto configuration
type CryptoConfig struct {
	KeyLength     int
	IVLength      int
	PaddingScheme string
	CipherMode    string
}

// Decompression configuration
type DecompressionConfig struct {
	MaxRetries      int
	Timeout         time.Duration
	EnabledMethods  []string
	FallbackMethods []string
}

// HTTP configuration
type HTTPConfig struct {
	Timeout         time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	UserAgent       string
	CompressionType string
	EnableKeepAlive bool
	MaxIdleConns    int
	IdleConnTimeout time.Duration
}

// Error types
type VClaimError struct {
	Code       string
	Message    string
	Endpoint   string
	HTTPStatus int
	Timestamp  time.Time
	Cause      error
}

func (e *VClaimError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *VClaimError) Unwrap() error {
	return e.Cause
}

// Constants
const (
	// Cipher modes
	CipherModeCBC = "CBC"
	CipherModeECB = "ECB"
	CipherModeGCM = "GCM"

	// Padding schemes
	PaddingPKCS7 = "PKCS7"
	PaddingANSI  = "ANSI"
	PaddingISO   = "ISO"

	// Decompression methods
	MethodLZString = "lzstring"
	MethodGzip     = "gzip"
	MethodBase64   = "base64"
	MethodPlain    = "plain"

	// HTTP methods
	MethodGET    = "GET"
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodPATCH  = "PATCH"
	MethodDELETE = "DELETE"

	// Default configurations
	DefaultTimeout         = 30 * time.Second
	DefaultMaxRetries      = 3
	DefaultRetryDelay      = 1 * time.Second
	DefaultKeyLength       = 32
	DefaultIVLength        = 16
	DefaultPaddingScheme   = PaddingPKCS7
	DefaultCipherMode      = CipherModeCBC
	DefaultUserAgent       = "BPJS-VClaim-Service/1.0"
	DefaultCompressionType = "gzip"
)

// Default configurations
var (
	DefaultCryptoConfig = CryptoConfig{
		KeyLength:     DefaultKeyLength,
		IVLength:      DefaultIVLength,
		PaddingScheme: DefaultPaddingScheme,
		CipherMode:    DefaultCipherMode,
	}

	DefaultDecompressionConfig = DecompressionConfig{
		MaxRetries:      3,
		Timeout:         10 * time.Second,
		EnabledMethods:  []string{MethodLZString, MethodGzip, MethodBase64, MethodPlain},
		FallbackMethods: []string{MethodPlain},
	}

	DefaultHTTPConfig = HTTPConfig{
		Timeout:         DefaultTimeout,
		MaxRetries:      DefaultMaxRetries,
		RetryDelay:      DefaultRetryDelay,
		UserAgent:       DefaultUserAgent,
		CompressionType: DefaultCompressionType,
	}
)

// ContextKey digunakan untuk mengirimkan token secara dinamis via context
type ContextKey string

const TokenContextKey ContextKey = "x-token"
