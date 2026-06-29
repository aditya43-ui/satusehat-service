package errors

import (
	"context"
	"fmt"
	"net/http"
	"service/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// ErrorMiddlewareConfig configures error middleware
type ErrorMiddlewareConfig struct {
	// SkipRoutes defines routes to skip error handling
	SkipRoutes []string

	// CustomHandlers defines custom error handlers
	CustomHandlers map[string]ErrorHandlerFunc

	// LogAllErrors enables logging all errors
	LogAllErrors bool

	// SanitizeErrors enables error sanitization
	SanitizeErrors bool

	// DefaultLanguage sets default language for error messages
	DefaultLanguage string

	// RequestIDHeader sets header name for request ID
	RequestIDHeader string

	// EnableMetrics enables error metrics collection
	EnableMetrics bool

	// Timeout sets timeout for error handling
	Timeout time.Duration

	// RecoveryEnabled enables panic recovery
	RecoveryEnabled bool

	// CircuitBreakerEnabled enables circuit breaker
	CircuitBreakerEnabled bool
}

// ErrorHandlerFunc handles errors
type ErrorHandlerFunc func(*gin.Context, error)

// DefaultMiddlewareConfig returns default configuration
func DefaultMiddlewareConfig() *ErrorMiddlewareConfig {
	return &ErrorMiddlewareConfig{
		SkipRoutes:            []string{"/health", "/metrics", "/ping"},
		LogAllErrors:          true,
		SanitizeErrors:        true,
		DefaultLanguage:       "en",
		RequestIDHeader:       "X-Request-ID",
		EnableMetrics:         true,
		Timeout:               30 * time.Second,
		RecoveryEnabled:       true,
		CircuitBreakerEnabled: false,
	}
}

// ErrorMiddleware creates Gin middleware for error handling
func ErrorMiddleware(config ...*ErrorMiddlewareConfig) gin.HandlerFunc {
	var cfg *ErrorMiddlewareConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultMiddlewareConfig()
	}

	return func(c *gin.Context) {
		// Skip specified routes
		for _, route := range cfg.SkipRoutes {
			if c.Request.URL.Path == route {
				c.Next()
				return
			}
		}

		// Generate request ID if not present
		requestID := c.GetHeader(cfg.RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)

		// Set language
		lang := getLanguageFromContext(c)
		if lang == "" {
			lang = cfg.DefaultLanguage
		}
		c.Set("language", lang)

		// Create context with timeout
		if cfg.Timeout > 0 {
			ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.Timeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
		}

		// Process request
		c.Next()

		// Handle errors
		if len(c.Errors) > 0 {
			handleErrors(c, c.Errors.Last().Err, cfg)
		}
	}
}

// RecoveryMiddleware creates recovery middleware
func RecoveryMiddleware(config ...*ErrorMiddlewareConfig) gin.HandlerFunc {
	var cfg *ErrorMiddlewareConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultMiddlewareConfig()
	}

	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var err error

		switch x := recovered.(type) {
		case string:
			err = NewWithCode(ErrCodeInternalError, x)
		case error:
			err = FromError(x)
		default:
			err = NewWithCode(ErrCodeUnexpectedError,
				fmt.Sprintf("Unknown panic: %v", x))
		}

		handleErrors(c, err, cfg)
	})
}

// CombinedMiddleware creates combined error and recovery middleware
func CombinedMiddleware(config ...*ErrorMiddlewareConfig) gin.HandlerFunc {
	cfg := DefaultMiddlewareConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// Skip specified routes
		for _, route := range cfg.SkipRoutes {
			if c.Request.URL.Path == route {
				c.Next()
				return
			}
		}

		// Generate request ID if not present
		requestID := c.GetHeader(cfg.RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)

		// Set language
		lang := getLanguageFromContext(c)
		if lang == "" {
			lang = cfg.DefaultLanguage
		}
		c.Set("language", lang)

		// Create context with timeout
		if cfg.Timeout > 0 {
			ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.Timeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
		}

		// Process request with recovery
		defer func() {
			if r := recover(); r != nil {
				var err error

				switch x := r.(type) {
				case string:
					err = NewWithCode(ErrCodeInternalError, x)
				case error:
					err = FromError(x)
				default:
					err = NewWithCode(ErrCodeUnexpectedError,
						fmt.Sprintf("Unknown panic: %v", x))
				}

				handleErrors(c, err, cfg)
			}
		}()

		c.Next()

		// Handle errors
		if len(c.Errors) > 0 {
			handleErrors(c, c.Errors.Last().Err, cfg)
		}
	}
}

// handleErrors handles errors with configuration
func handleErrors(c *gin.Context, err error, cfg *ErrorMiddlewareConfig) {
	appErr := FromError(err)
	if appErr == nil {
		appErr = NewWithCode(ErrCodeInternalError, "Unknown error")
	}

	// Apply custom handlers
	if handler, exists := cfg.CustomHandlers[appErr.Code()]; exists {
		handler(c, appErr)
		return
	}

	// Apply registered handlers
	appErr = ApplyHandlers(appErr)

	// Sanitize error if needed
	if cfg.SanitizeErrors {
		appErr = sanitizeError(appErr).(Error)
	}

	// Record metrics
	if cfg.EnableMetrics {
		RecordError(appErr)
	}

	// Log error
	if cfg.LogAllErrors {
		log := logger.Default()
		log.Error("HTTP error occurred",
			logger.String("error_code", appErr.Code()),
			logger.String("category", appErr.Category()),
			logger.Int("http_status", appErr.HTTPStatus()),
			logger.Any("metadata", appErr.Metadata()),
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("request_id", c.GetString("request_id")),
		)
	}

	// Create error response
	response := createErrorResponse(c, appErr, cfg)

	// Send response
	if !c.Writer.Written() {
		c.JSON(appErr.HTTPStatus(), response)
	}
}

// createErrorResponse creates error response
func createErrorResponse(c *gin.Context, err Error, cfg *ErrorMiddlewareConfig) gin.H {
	lang := c.GetString("language")
	requestID := c.GetString("request_id")

	details := err.Metadata()
	if details == nil {
		details = make(map[string]interface{})
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	if ts, ok := details["timestamp"].(string); ok {
		timestamp = ts
	}

	response := gin.H{
		"status": "error",
		"error": gin.H{
			"code":        err.Code(),
			"details":     details,
			"message":     err.GetLocalizedMessage(lang),
			"request_id":  requestID,
			"retryable":   IsRetryable(err.Code()),
			"stack_trace": nil,
			"timestamp":   timestamp,
		},
		"meta": gin.H{
			"category":    err.Category(),
			"http_status": err.HTTPStatus(),
		},
	}

	// Add stack trace in debug mode
	if gin.Mode() == gin.DebugMode {
		response["error"].(gin.H)["stack_trace"] = err.Metadata()["stack_trace"]
	}

	return response
}

// CustomErrorHandler creates custom error handler
func CustomErrorHandler(code string, handler ErrorHandlerFunc) ErrorHandlerFunc {
	return func(c *gin.Context, err error) {
		appErr := FromError(err)
		if appErr != nil && appErr.Code() == code {
			handler(c, appErr)
			return
		}
		// Fallback to default handling
		HandleHTTPError(c, err)
	}
}

// ValidationHandler creates validation error handler
func ValidationHandler(c *gin.Context, err error) {
	appErr := FromError(err)
	if appErr == nil {
		appErr = NewWithCode(ErrCodeValidationFailed, err.Error())
	}

	// Enhance validation errors with field details
	if appErr.Category() == CategoryValidation {
		if details, ok := appErr.Metadata()["validation_errors"]; ok {
			response := gin.H{
				"status": "error",
				"error": gin.H{
					"code":    appErr.Code(),
					"message": appErr.GetLocalizedMessage(getLanguageFromContext(c)),
					"fields":  details,
				},
			}
			if !c.Writer.Written() {
				c.JSON(http.StatusBadRequest, response)
			}
			return
		}
	}

	HandleHTTPError(c, appErr)
}

// TimeoutHandler creates timeout error handler
func TimeoutHandler(c *gin.Context, err error) {
	appErr := NewWithCode(ErrCodeTimeout, "Request timeout").
		WithMetadata("timeout", c.GetHeader("X-Timeout")).
		WithMetadata("endpoint", c.Request.URL.Path)

	HandleHTTPError(c, appErr)
}

// RateLimitHandler creates rate limit error handler
func RateLimitHandler(c *gin.Context, err error) {
	appErr := NewWithCode(ErrCodeRateLimitExceeded, "Rate limit exceeded").
		WithMetadata("limit", c.GetHeader("X-Rate-Limit-Limit")).
		WithMetadata("remaining", c.GetHeader("X-Rate-Limit-Remaining")).
		WithMetadata("reset", c.GetHeader("X-Rate-Limit-Reset"))

	HandleHTTPError(c, appErr)
}

// NotFoundHandler creates 404 error handler
func NotFoundHandler(c *gin.Context) {
	err := NewWithCode(ErrCodeNotFound,
		fmt.Sprintf("Route %s %s not found", c.Request.Method, c.Request.URL.Path)).
		WithMetadata("method", c.Request.Method).
		WithMetadata("path", c.Request.URL.Path).
		WithMetadata("query", c.Request.URL.RawQuery)

	HandleHTTPError(c, err)
}

// MethodNotAllowedHandler creates 405 error handler
func MethodNotAllowedHandler(c *gin.Context) {
	err := NewWithCode(ErrCodeInvalidInput,
		fmt.Sprintf("Method %s not allowed for route %s", c.Request.Method, c.Request.URL.Path)).
		WithMetadata("method", c.Request.Method).
		WithMetadata("path", c.Request.URL.Path).
		WithMetadata("allowed_methods", c.GetHeader("Allow"))

	HandleHTTPError(c, err)
}

// GRPCErrorInterceptor creates gRPC error interceptor
func GRPCErrorInterceptor(config ...*ErrorMiddlewareConfig) grpc.UnaryServerInterceptor {
	cfg := DefaultMiddlewareConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, handleGRPCError(err, cfg)
		}
		return resp, nil
	}
}

// handleGRPCError handles gRPC error with configuration
func handleGRPCError(err error, cfg *ErrorMiddlewareConfig) error {
	appErr := FromError(err)
	if appErr == nil {
		appErr = NewWithCode(ErrCodeInternalError, "Internal server error")
	}

	// Apply registered handlers
	appErr = ApplyHandlers(appErr)

	// Sanitize error if needed
	if cfg.SanitizeErrors {
		appErr = sanitizeError(appErr).(Error)
	}

	// Record metrics
	if cfg.EnableMetrics {
		RecordError(appErr)
	}

	// Log error
	if cfg.LogAllErrors {
		log := logger.Default()
		log.Error("gRPC error occurred",
			logger.String("error_code", appErr.Code()),
			logger.String("category", appErr.Category()),
			logger.String("grpc_code", appErr.GRPCCode().String()),
			logger.Any("metadata", appErr.Metadata()),
		)
	}

	// Convert to gRPC status
	return status.Error(appErr.GRPCCode(), appErr.GetLocalizedMessage(cfg.DefaultLanguage))
}

// GetErrorMetrics returns current error metrics
func GetErrorMetrics() ErrorMetrics {
	m := GetMetrics()
	return ErrorMetrics{
		TotalErrors:      m.TotalErrors,
		ErrorsByCode:     m.ErrorsByCode,
		ErrorsByCategory: m.ErrorsByCategory,
	}
}

// MetricsHandler creates metrics endpoint handler
func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := GetErrorMetrics()
		c.JSON(http.StatusOK, gin.H{
			"error_metrics": metrics,
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// HealthHandler creates health endpoint handler
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "error-handler",
		})
	}
}
