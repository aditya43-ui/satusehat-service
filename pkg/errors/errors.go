package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"google.golang.org/grpc/codes"
)

// Core error interface
type Error interface {
	Error() string
	Code() string
	Category() string
	HTTPStatus() int
	GRPCCode() codes.Code
	Metadata() map[string]interface{}
	Cause() error
	GetLocalizedMessage(lang string) string
	WithMetadata(key string, value interface{}) Error
	WithCause(err error) Error
}

// Metadata type for error context
type Metadata map[string]interface{}

// Core error implementation
type AppError struct {
	code       string
	message    string
	category   string
	httpStatus int
	grpcCode   codes.Code
	metadata   Metadata
	cause      error
	timestamp  time.Time
	stackTrace []string
}

// New creates a new error
func New(message string) Error {
	return &AppError{
		code:       ErrCodeInternalError,
		message:    message,
		category:   CategoryInternal,
		httpStatus: http.StatusInternalServerError,
		grpcCode:   codes.Internal,
		metadata:   make(Metadata),
		timestamp:  time.Now(),
		stackTrace: captureStackTrace(),
	}
}

// NewWithCode creates a new error with code
func NewWithCode(code, message string) Error {
	info := GetErrorInfo(code)
	if info.Code == "" {
		info = ErrorInfo{
			Code:       code,
			Category:   CategoryInternal,
			HTTPStatus: http.StatusInternalServerError,
			GRPCCode:   codes.Internal,
		}
	}

	return &AppError{
		code:       code,
		message:    message,
		category:   info.Category,
		httpStatus: info.HTTPStatus,
		grpcCode:   info.GRPCCode,
		metadata:   make(Metadata),
		timestamp:  time.Now(),
		stackTrace: captureStackTrace(),
	}
}

// NewWithMetadata creates a new error with metadata
func NewWithMetadata(code, message string, metadata Metadata) Error {
	err := NewWithCode(code, message)
	err.(*AppError).metadata = metadata
	return err
}

// NewWithType creates a new error with custom type
func NewWithType(errorType string, metadata Metadata) Error {
	template := GetErrorTemplate(errorType)
	if template == nil {
		return NewWithMetadata(ErrCodeInternalError, "Unknown error type", metadata)
	}

	err := &AppError{
		code:       template.Code,
		message:    template.Message,
		category:   template.Category,
		httpStatus: template.HTTPStatus,
		grpcCode:   template.GRPCCode,
		metadata:   make(Metadata),
		timestamp:  time.Now(),
		stackTrace: captureStackTrace(),
	}

	for k, v := range metadata {
		err.metadata[k] = v
	}

	return err
}

// Wrap wraps an existing error
func Wrap(err error, code, message string) Error {
	if err == nil {
		return nil
	}

	if IsAppError(err) {
		return err.(Error)
	}

	appErr := NewWithCode(code, message)
	appErr.(*AppError).cause = err
	return appErr
}

// FromError converts any error to AppError
func FromError(err error) Error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(Error); ok {
		return appErr
	}

	return Wrap(err, ErrCodeInternalError, err.Error())
}

// IsAppError checks if error is AppError
func IsAppError(err error) bool {
	_, ok := err.(Error)
	return ok
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Code returns error code
func (e *AppError) Code() string {
	return e.code
}

// Category returns error category
func (e *AppError) Category() string {
	return e.category
}

// HTTPStatus returns HTTP status code
func (e *AppError) HTTPStatus() int {
	return e.httpStatus
}

// GRPCCode returns gRPC code
func (e *AppError) GRPCCode() codes.Code {
	return e.grpcCode
}

// Metadata returns error metadata
func (e *AppError) Metadata() map[string]interface{} {
	return e.metadata
}

// Cause returns underlying error
func (e *AppError) Cause() error {
	return e.cause
}

// Unwrap implements the standard library unwrap interface for Go 1.13+ error chains
func (e *AppError) Unwrap() error {
	return e.cause
}

// GetLocalizedMessage returns localized message
func (e *AppError) GetLocalizedMessage(lang string) string {
	return GetLocalizedMessage(e.code, lang, e.message)
}

// WithMetadata adds metadata to error
func (e *AppError) WithMetadata(key string, value interface{}) Error {
	if e.metadata == nil {
		e.metadata = make(Metadata)
	}
	e.metadata[key] = value
	return e
}

// WithCause adds cause to error
func (e *AppError) WithCause(err error) Error {
	e.cause = err
	return e
}

// ToJSON converts error to JSON
func (e *AppError) ToJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"code":        e.code,
		"message":     e.message,
		"category":    e.category,
		"http_status": e.httpStatus,
		"grpc_code":   e.grpcCode.String(),
		"metadata":    e.metadata,
		"timestamp":   e.timestamp,
		"stack_trace": e.stackTrace,
	})
}

// captureStackTrace captures current stack trace
func captureStackTrace() []string {
	var stack []string
	for i := 2; i < 15; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	return stack
}

// Is checks if error matches target
func Is(err, target error) bool {
	if err == target {
		return true
	}

	if appErr, ok := err.(Error); ok {
		if targetErr, ok := target.(Error); ok {
			return appErr.Code() == targetErr.Code()
		}
	}

	return false
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	if appErr, ok := err.(Error); ok {
		switch t := target.(type) {
		case **AppError:
			*t = appErr.(*AppError)
			return true
		case *Error:
			*t = appErr
			return true
		}
	}
	return false
}
