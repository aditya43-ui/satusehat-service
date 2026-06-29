package errors

import (
	"net/http"
	"strings"
	"sync"

	"google.golang.org/grpc/codes"
)

// Error categories
const (
	CategoryValidation   = "validation"
	CategoryNotFound     = "not_found"
	CategoryUnauthorized = "unauthorized"
	CategoryForbidden    = "forbidden"
	CategoryConflict     = "conflict"
	CategoryRateLimit    = "rate_limit"
	CategoryInternal     = "internal"
	CategoryExternal     = "external"
	CategoryBusiness     = "business"
	CategoryTimeout      = "timeout"
	CategoryNetwork      = "network"
	CategoryDatabase     = "database"
)

// Standard error codes
const (
	// Validation errors
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeMissingField     = "MISSING_FIELD"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
	ErrCodeInvalidLength    = "INVALID_LENGTH"
	ErrCodeInvalidRange     = "INVALID_RANGE"

	// Not found errors
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeUserNotFound     = "USER_NOT_FOUND"
	ErrCodeResourceNotFound = "RESOURCE_NOT_FOUND"
	ErrCodeDataNotFound     = "DATA_NOT_FOUND"

	// Authorization errors
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"

	// Forbidden errors
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeInsufficientRights = "INSUFFICIENT_RIGHTS"
	ErrCodeAccessDenied       = "ACCESS_DENIED"

	// Conflict errors
	ErrCodeConflict         = "CONFLICT"
	ErrCodeDuplicateEntry   = "DUPLICATE_ENTRY"
	ErrCodeResourceLocked   = "RESOURCE_LOCKED"
	ErrCodeConcurrentUpdate = "CONCURRENT_UPDATE"

	// Rate limit errors
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeTooManyRequests   = "TOO_MANY_REQUESTS"
	ErrCodeQuotaExceeded     = "QUOTA_EXCEEDED"

	// Internal errors
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeUnexpectedError    = "UNEXPECTED_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeConfigurationError = "CONFIGURATION_ERROR"

	// External errors
	ErrCodeExternalError   = "EXTERNAL_ERROR"
	ErrCodeThirdPartyError = "THIRD_PARTY_ERROR"
	ErrCodeAPIError        = "API_ERROR"

	// Business errors
	ErrCodeBusinessRule        = "BUSINESS_RULE"
	ErrCodeInsufficientBalance = "INSUFFICIENT_BALANCE"
	ErrCodeAccountSuspended    = "ACCOUNT_SUSPENDED"

	// Timeout errors
	ErrCodeTimeout           = "TIMEOUT"
	ErrCodeRequestTimeout    = "REQUEST_TIMEOUT"
	ErrCodeConnectionTimeout = "CONNECTION_TIMEOUT"

	// Network errors
	ErrCodeNetworkError     = "NETWORK_ERROR"
	ErrCodeConnectionFailed = "CONNECTION_FAILED"
	ErrCodeConnectionLost   = "CONNECTION_LOST"

	// Database errors
	ErrCodeDatabaseError     = "DATABASE_ERROR"
	ErrCodeQueryFailed       = "QUERY_FAILED"
	ErrCodeTransactionFailed = "TRANSACTION_FAILED"
)

// ErrorInfo contains error metadata
type ErrorInfo struct {
	Code       string
	Category   string
	HTTPStatus int
	GRPCCode   codes.Code
}

// ErrorTemplate for custom error types
type ErrorTemplate struct {
	Code       string
	Message    string
	Category   string
	HTTPStatus int
	GRPCCode   codes.Code
	Retryable  bool
}

var (
	errorCodes       = make(map[string]ErrorInfo)
	errorCodesMu     sync.RWMutex
	errorTemplates   = make(map[string]*ErrorTemplate)
	errorTemplatesMu sync.RWMutex
)

func init() {
	// Initialize standard error codes
	initializeErrorCodes()
}

// initializeErrorCodes sets up standard error codes
func initializeErrorCodes() {
	// Validation errors
	registerErrorCode(ErrCodeValidationFailed, CategoryValidation, http.StatusBadRequest, codes.InvalidArgument)
	registerErrorCode(ErrCodeInvalidInput, CategoryValidation, http.StatusBadRequest, codes.InvalidArgument)
	registerErrorCode(ErrCodeMissingField, CategoryValidation, http.StatusBadRequest, codes.InvalidArgument)
	registerErrorCode(ErrCodeInvalidFormat, CategoryValidation, http.StatusBadRequest, codes.InvalidArgument)
	registerErrorCode(ErrCodeInvalidLength, CategoryValidation, http.StatusBadRequest, codes.InvalidArgument)
	registerErrorCode(ErrCodeInvalidRange, CategoryValidation, http.StatusBadRequest, codes.InvalidArgument)

	// Not found errors
	registerErrorCode(ErrCodeNotFound, CategoryNotFound, http.StatusNotFound, codes.NotFound)
	registerErrorCode(ErrCodeUserNotFound, CategoryNotFound, http.StatusNotFound, codes.NotFound)
	registerErrorCode(ErrCodeResourceNotFound, CategoryNotFound, http.StatusNotFound, codes.NotFound)
	registerErrorCode(ErrCodeDataNotFound, CategoryNotFound, http.StatusNotFound, codes.NotFound)

	// Authorization errors
	registerErrorCode(ErrCodeUnauthorized, CategoryUnauthorized, http.StatusUnauthorized, codes.Unauthenticated)
	registerErrorCode(ErrCodeInvalidToken, CategoryUnauthorized, http.StatusUnauthorized, codes.Unauthenticated)
	registerErrorCode(ErrCodeTokenExpired, CategoryUnauthorized, http.StatusUnauthorized, codes.Unauthenticated)
	registerErrorCode(ErrCodeInvalidCredentials, CategoryUnauthorized, http.StatusUnauthorized, codes.Unauthenticated)

	// Forbidden errors
	registerErrorCode(ErrCodeForbidden, CategoryForbidden, http.StatusForbidden, codes.PermissionDenied)
	registerErrorCode(ErrCodeInsufficientRights, CategoryForbidden, http.StatusForbidden, codes.PermissionDenied)
	registerErrorCode(ErrCodeAccessDenied, CategoryForbidden, http.StatusForbidden, codes.PermissionDenied)

	// Conflict errors
	registerErrorCode(ErrCodeConflict, CategoryConflict, http.StatusConflict, codes.AlreadyExists)
	registerErrorCode(ErrCodeDuplicateEntry, CategoryConflict, http.StatusConflict, codes.AlreadyExists)
	registerErrorCode(ErrCodeResourceLocked, CategoryConflict, http.StatusLocked, codes.Aborted)
	registerErrorCode(ErrCodeConcurrentUpdate, CategoryConflict, http.StatusConflict, codes.Aborted)

	// Rate limit errors
	registerErrorCode(ErrCodeRateLimitExceeded, CategoryRateLimit, http.StatusTooManyRequests, codes.ResourceExhausted)
	registerErrorCode(ErrCodeTooManyRequests, CategoryRateLimit, http.StatusTooManyRequests, codes.ResourceExhausted)
	registerErrorCode(ErrCodeQuotaExceeded, CategoryRateLimit, http.StatusTooManyRequests, codes.ResourceExhausted)

	// Internal errors
	registerErrorCode(ErrCodeInternalError, CategoryInternal, http.StatusInternalServerError, codes.Internal)
	registerErrorCode(ErrCodeUnexpectedError, CategoryInternal, http.StatusInternalServerError, codes.Internal)
	registerErrorCode(ErrCodeServiceUnavailable, CategoryInternal, http.StatusServiceUnavailable, codes.Unavailable)
	registerErrorCode(ErrCodeConfigurationError, CategoryInternal, http.StatusInternalServerError, codes.Internal)

	// External errors
	registerErrorCode(ErrCodeExternalError, CategoryExternal, http.StatusBadGateway, codes.Unavailable)
	registerErrorCode(ErrCodeThirdPartyError, CategoryExternal, http.StatusBadGateway, codes.Unavailable)
	registerErrorCode(ErrCodeAPIError, CategoryExternal, http.StatusBadGateway, codes.Unavailable)

	// Business errors
	registerErrorCode(ErrCodeBusinessRule, CategoryBusiness, http.StatusBadRequest, codes.FailedPrecondition)
	registerErrorCode(ErrCodeInsufficientBalance, CategoryBusiness, http.StatusBadRequest, codes.FailedPrecondition)
	registerErrorCode(ErrCodeAccountSuspended, CategoryBusiness, http.StatusForbidden, codes.PermissionDenied)

	// Timeout errors
	registerErrorCode(ErrCodeTimeout, CategoryTimeout, http.StatusRequestTimeout, codes.DeadlineExceeded)
	registerErrorCode(ErrCodeRequestTimeout, CategoryTimeout, http.StatusRequestTimeout, codes.DeadlineExceeded)
	registerErrorCode(ErrCodeConnectionTimeout, CategoryTimeout, http.StatusRequestTimeout, codes.DeadlineExceeded)

	// Network errors
	registerErrorCode(ErrCodeNetworkError, CategoryNetwork, http.StatusBadGateway, codes.Unavailable)
	registerErrorCode(ErrCodeConnectionFailed, CategoryNetwork, http.StatusBadGateway, codes.Unavailable)
	registerErrorCode(ErrCodeConnectionLost, CategoryNetwork, http.StatusBadGateway, codes.Unavailable)

	// Database errors
	registerErrorCode(ErrCodeDatabaseError, CategoryDatabase, http.StatusInternalServerError, codes.Internal)
	registerErrorCode(ErrCodeQueryFailed, CategoryDatabase, http.StatusInternalServerError, codes.Internal)
	registerErrorCode(ErrCodeTransactionFailed, CategoryDatabase, http.StatusInternalServerError, codes.Aborted)
}

// registerErrorCode registers an error code
func registerErrorCode(code, category string, httpStatus int, grpcCode codes.Code) {
	errorCodesMu.Lock()
	defer errorCodesMu.Unlock()
	errorCodes[code] = ErrorInfo{
		Code:       code,
		Category:   category,
		HTTPStatus: httpStatus,
		GRPCCode:   grpcCode,
	}
}

// RegisterErrorCode registers a custom error code
func RegisterErrorCode(code, category string, httpStatus int, grpcCode codes.Code) {
	registerErrorCode(code, category, httpStatus, grpcCode)
}

// GetErrorInfo returns error information for a code
func GetErrorInfo(code string) ErrorInfo {
	errorCodesMu.RLock()
	defer errorCodesMu.RUnlock()

	if info, exists := errorCodes[code]; exists {
		return info
	}

	return ErrorInfo{
		Code:       code,
		Category:   CategoryInternal,
		HTTPStatus: http.StatusInternalServerError,
		GRPCCode:   codes.Internal,
	}
}

// RegisterErrorTemplate registers a custom error template
func RegisterErrorTemplate(name string, template *ErrorTemplate) {
	errorTemplatesMu.Lock()
	defer errorTemplatesMu.Unlock()
	errorTemplates[name] = template
}

// GetErrorTemplate returns error template by name
func GetErrorTemplate(name string) *ErrorTemplate {
	errorTemplatesMu.RLock()
	defer errorTemplatesMu.RUnlock()
	return errorTemplates[name]
}

// GetAllErrorCodes returns all registered error codes
func GetAllErrorCodes() map[string]ErrorInfo {
	errorCodesMu.RLock()
	defer errorCodesMu.RUnlock()

	result := make(map[string]ErrorInfo)
	for k, v := range errorCodes {
		result[k] = v
	}
	return result
}

// GetErrorCodesByCategory returns error codes by category
func GetErrorCodesByCategory(category string) map[string]ErrorInfo {
	errorCodesMu.RLock()
	defer errorCodesMu.RUnlock()

	result := make(map[string]ErrorInfo)
	for k, v := range errorCodes {
		if v.Category == category {
			result[k] = v
		}
	}
	return result
}

// IsRetryable checks if error is retryable
func IsRetryable(code string) bool {
	info := GetErrorInfo(code)
	switch info.Category {
	case CategoryNetwork, CategoryTimeout, CategoryExternal:
		return true
	case CategoryDatabase:
		return strings.Contains(code, "TIMEOUT") || strings.Contains(code, "CONNECTION")
	default:
		return false
	}
}

// IsClientError checks if error is client-side (4xx)
func IsClientError(code string) bool {
	info := GetErrorInfo(code)
	return info.HTTPStatus >= 400 && info.HTTPStatus < 500
}

// IsServerError checks if error is server-side (5xx)
func IsServerError(code string) bool {
	info := GetErrorInfo(code)
	return info.HTTPStatus >= 500
}
