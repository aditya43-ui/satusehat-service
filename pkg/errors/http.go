package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"service/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPErrorResponse represents standardized HTTP error response
type HTTPErrorResponse struct {
	Status string `json:"status"`
	Error  struct {
		Code       string                 `json:"code"`
		Details    map[string]interface{} `json:"details"`
		Message    string                 `json:"message"`
		RequestID  string                 `json:"request_id"`
		Retryable  bool                   `json:"retryable"`
		StackTrace interface{}            `json:"stack_trace"`
		Timestamp  string                 `json:"timestamp"`
	} `json:"error"`
	Meta struct {
		Category   string `json:"category"`
		HTTPStatus int    `json:"http_status"`
	} `json:"meta,omitempty"`
}

// HTTPMiddleware creates error handling middleware for Gin
func HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last().Err
			HandleHTTPError(c, err)
		}
	}
}

// HandleHTTPError handles HTTP error response
func HandleHTTPError(c *gin.Context, err error) {
	appErr := FromError(err)
	if appErr == nil {
		appErr = NewWithCode(ErrCodeInternalError, "Unknown error")
	}

	// Create error response
	response := &HTTPErrorResponse{}
	response.Status = "error"
	response.Error.Code = appErr.Code()
	response.Error.Details = appErr.Metadata()
	if response.Error.Details == nil {
		response.Error.Details = make(map[string]interface{})
	}
	response.Error.Message = appErr.GetLocalizedMessage(getLanguageFromContext(c))
	response.Error.RequestID = c.GetString("request_id")
	response.Error.Retryable = IsRetryable(appErr.Code())
	response.Error.StackTrace = nil
	if ts, ok := appErr.Metadata()["timestamp"].(string); ok {
		response.Error.Timestamp = ts
	}
	response.Meta.Category = appErr.Category()
	response.Meta.HTTPStatus = appErr.HTTPStatus()

	// Log error
	LogHTTPError(c, appErr)

	// Cegah print ganda jika controller sudah print
	if c.Writer.Written() {
		return
	}

	// Send response
	c.JSON(appErr.HTTPStatus(), response)
}

// getLanguageFromContext extracts language from context
func getLanguageFromContext(c *gin.Context) string {
	// Try Accept-Language header
	lang := c.GetHeader("Accept-Language")
	if lang != "" {
		return parseAcceptLanguage(lang)
	}

	// Try query parameter
	lang = c.Query("lang")
	if lang != "" {
		return lang
	}

	// Default to English
	return "en"
}

// parseAcceptLanguage parses Accept-Language header
func parseAcceptLanguage(header string) string {
	// Simple parsing - take first language
	for idx := 0; idx < len(header); idx++ {
		if header[idx] == ',' || header[idx] == ';' {
			return header[:idx]
		}
	}
	return header
}

// LogHTTPError logs HTTP error with context
func LogHTTPError(c *gin.Context, err Error) {
	log := logger.Default()

	log.Error("HTTP error occurred",
		logger.String("error_code", err.Code()),
		logger.String("category", err.Category()),
		logger.Int("http_status", err.HTTPStatus()),
		logger.Any("metadata", err.Metadata()),
		logger.String("path", c.Request.URL.Path),
		logger.String("method", c.Request.Method),
	)
}

// ValidationErrorHandler handles validation errors specifically
func ValidationErrorHandler(c *gin.Context, err error) {
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

// PanicHandler handles panics
func PanicHandler(c *gin.Context, recovered interface{}) {
	var err error

	switch x := recovered.(type) {
	case string:
		err = NewWithCode(ErrCodeInternalError, x)
	case error:
		err = FromError(x)
	default:
		err = NewWithCode(ErrCodeUnexpectedError, fmt.Sprintf("Unknown panic: %v", x))
	}

	HandleHTTPError(c, err)
}

// WriteErrorResponse writes error response directly
func WriteErrorResponse(w http.ResponseWriter, err error) {
	appErr := FromError(err)
	if appErr == nil {
		appErr = NewWithCode(ErrCodeInternalError, "Unknown error")
	}

	response := &HTTPErrorResponse{}
	response.Status = "error"
	response.Error.Code = appErr.Code()
	response.Error.Details = appErr.Metadata()
	if response.Error.Details == nil {
		response.Error.Details = make(map[string]interface{})
	}
	response.Error.Message = appErr.GetLocalizedMessage("en")
	response.Error.Retryable = IsRetryable(appErr.Code())
	response.Error.StackTrace = nil
	if ts, ok := appErr.Metadata()["timestamp"].(string); ok {
		response.Error.Timestamp = ts
	}
	response.Meta.Category = appErr.Category()
	response.Meta.HTTPStatus = appErr.HTTPStatus()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus())
	json.NewEncoder(w).Encode(response)
}

// sanitizeError removes sensitive information from errors
func sanitizeError(err error) error {
	appErr := FromError(err)
	if appErr == nil {
		return err
	}

	// Remove sensitive metadata
	sanitized := make(Metadata)
	for k, v := range appErr.Metadata() {
		if !isSensitiveField(k) {
			sanitized[k] = v
		}
	}

	return NewWithMetadata(appErr.Code(), appErr.Error(), sanitized)
}

// isSensitiveField checks if field contains sensitive information
func isSensitiveField(field string) bool {
	sensitiveFields := []string{"password", "token", "secret", "key", "auth"}
	for _, sensitive := range sensitiveFields {
		if strings.Contains(strings.ToLower(field), sensitive) {
			return true
		}
	}
	return false
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
