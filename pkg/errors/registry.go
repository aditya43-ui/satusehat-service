package errors

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
)

// ErrorType represents a custom error type
type ErrorType interface {
	Name() string
	New(message string, metadata Metadata) Error
	FromError(err error) Error
	Validate(metadata Metadata) error
}

// ErrorRegistry manages custom error types
type ErrorRegistry struct {
	types    map[string]ErrorType
	handlers map[string]func(Error) Error
	mu       sync.RWMutex
}

var (
	registry = &ErrorRegistry{
		types:    make(map[string]ErrorType),
		handlers: make(map[string]func(Error) Error),
	}
)

// RegisterErrorType registers a custom error type
func RegisterErrorType(name string, errorType ErrorType) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.types[name] = errorType
}

// GetErrorType returns error type by name
func GetErrorType(name string) (ErrorType, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	errorType, exists := registry.types[name]
	return errorType, exists
}

// RegisterErrorHandler registers custom error handler
func RegisterErrorHandler(code string, handler func(Error) Error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.handlers[code] = handler
}

// GetErrorHandler returns error handler for code
func GetErrorHandler(code string) (func(Error) Error, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	handler, exists := registry.handlers[code]
	return handler, exists
}

// ApplyHandlers applies registered handlers to error
func ApplyHandlers(err Error) Error {
	if err == nil {
		return nil
	}

	// Apply specific handler for error code
	if handler, exists := GetErrorHandler(err.Code()); exists {
		err = handler(err)
	}

	// Apply category handlers
	categoryHandlers := GetCategoryHandlers(err.Category())
	for _, handler := range categoryHandlers {
		err = handler(err)
	}

	return err
}

// CategoryHandler handles errors by category
type CategoryHandler struct {
	Category string
	Handler  func(Error) Error
}

var (
	categoryHandlers = make(map[string][]func(Error) Error)
	categoryMu       sync.RWMutex
)

// RegisterCategoryHandler registers handler for error category
func RegisterCategoryHandler(category string, handler func(Error) Error) {
	categoryMu.Lock()
	defer categoryMu.Unlock()

	if categoryHandlers[category] == nil {
		categoryHandlers[category] = make([]func(Error) Error, 0)
	}
	categoryHandlers[category] = append(categoryHandlers[category], handler)
}

// GetCategoryHandlers returns handlers for category
func GetCategoryHandlers(category string) []func(Error) Error {
	categoryMu.RLock()
	defer categoryMu.RUnlock()

	handlers := make([]func(Error) Error, len(categoryHandlers[category]))
	copy(handlers, categoryHandlers[category])
	return handlers
}

// BaseErrorType provides base implementation for custom error types
type BaseErrorType struct {
	name        string
	defaultCode string
	category    string
	httpStatus  int
	grpcCode    int
	validator   func(Metadata) error
}

// NewBaseErrorType creates new base error type
func NewBaseErrorType(name, defaultCode, category string, httpStatus int, grpcCode int) *BaseErrorType {
	return &BaseErrorType{
		name:        name,
		defaultCode: defaultCode,
		category:    category,
		httpStatus:  httpStatus,
		grpcCode:    grpcCode,
	}
}

// Name returns error type name
func (bet *BaseErrorType) Name() string {
	return bet.name
}

// New creates new error of this type
func (bet *BaseErrorType) New(message string, metadata Metadata) Error {
	code := bet.defaultCode
	if c, exists := metadata["code"].(string); exists {
		code = c
	}

	return &AppError{
		code:       code,
		message:    message,
		category:   bet.category,
		httpStatus: bet.httpStatus,
		grpcCode:   codes.Code(bet.grpcCode),
		metadata:   metadata,
		timestamp:  time.Now(),
		stackTrace: captureStackTrace(),
	}
}

// FromError converts existing error to this type
func (bet *BaseErrorType) FromError(err error) Error {
	appErr := FromError(err)
	if appErr == nil {
		return bet.New(err.Error(), make(Metadata))
	}

	return bet.New(appErr.Error(), appErr.Metadata())
}

// Validate validates metadata
func (bet *BaseErrorType) Validate(metadata Metadata) error {
	if bet.validator != nil {
		return bet.validator(metadata)
	}
	return nil
}

// SetValidator sets validator function
func (bet *BaseErrorType) SetValidator(validator func(Metadata) error) {
	bet.validator = validator
}

// BusinessErrorType represents business logic errors
type BusinessErrorType struct {
	*BaseErrorType
	ruleName string
}

// NewBusinessErrorType creates new business error type
func NewBusinessErrorType(name, ruleName string) *BusinessErrorType {
	return &BusinessErrorType{
		BaseErrorType: NewBaseErrorType(name, ErrCodeBusinessRule, CategoryBusiness, 400, int(codes.FailedPrecondition)),
		ruleName:      ruleName,
	}
}

// RuleName returns business rule name
func (bet *BusinessErrorType) RuleName() string {
	return bet.ruleName
}

// ValidationErrorType represents validation errors
type ValidationErrorType struct {
	*BaseErrorType
	field string
	rule  string
}

// NewValidationErrorType creates new validation error type
func NewValidationErrorType(name, field, rule string) *ValidationErrorType {
	return &ValidationErrorType{
		BaseErrorType: NewBaseErrorType(name, ErrCodeValidationFailed, CategoryValidation, 400, int(codes.InvalidArgument)),
		field:         field,
		rule:          rule,
	}
}

// Field returns field name
func (vet *ValidationErrorType) Field() string {
	return vet.field
}

// Rule returns validation rule
func (vet *ValidationErrorType) Rule() string {
	return vet.rule
}

// ExternalErrorType represents external service errors
type ExternalErrorType struct {
	*BaseErrorType
	serviceName string
	retryable   bool
}

// NewExternalErrorType creates new external error type
func NewExternalErrorType(name, serviceName string, retryable bool) *ExternalErrorType {
	return &ExternalErrorType{
		BaseErrorType: NewBaseErrorType(name, ErrCodeExternalError, CategoryExternal, 502, int(codes.Unavailable)),
		serviceName:   serviceName,
		retryable:     retryable,
	}
}

// ServiceName returns external service name
func (eet *ExternalErrorType) ServiceName() string {
	return eet.serviceName
}

// Retryable returns if error is retryable
func (eet *ExternalErrorType) Retryable() bool {
	return eet.retryable
}

// ErrorFactory creates errors from configuration
type ErrorFactory struct {
	config map[string]interface{}
	mu     sync.RWMutex
}

// NewErrorFactory creates new error factory
func NewErrorFactory() *ErrorFactory {
	return &ErrorFactory{
		config: make(map[string]interface{}),
	}
}

// Configure configures error factory
func (ef *ErrorFactory) Configure(config map[string]interface{}) {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	ef.config = config
}

// CreateError creates error from configuration
func (ef *ErrorFactory) CreateError(errorName string, message string, metadata Metadata) Error {
	ef.mu.RLock()
	defer ef.mu.RUnlock()

	// Try to get error type from configuration
	if errorTypeConfig, exists := ef.config[errorName]; exists {
		if configMap, ok := errorTypeConfig.(map[string]interface{}); ok {
			name, _ := configMap["name"].(string)
			if errorType, exists := GetErrorType(name); exists {
				return errorType.New(message, metadata)
			}
		}
	}

	// Fallback to creating standard error
	return NewWithMetadata(ErrCodeInternalError, message, metadata)
}

// GetRegisteredTypes returns all registered error types
func GetRegisteredTypes() map[string]ErrorType {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	types := make(map[string]ErrorType)
	for k, v := range registry.types {
		types[k] = v
	}
	return types
}

// GetRegisteredHandlers returns all registered handlers
func GetRegisteredHandlers() map[string]func(Error) Error {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	handlers := make(map[string]func(Error) Error)
	for k, v := range registry.handlers {
		handlers[k] = v
	}
	return handlers
}

// ValidateErrorType validates error type configuration
func ValidateErrorType(errorType ErrorType) error {
	if errorType.Name() == "" {
		return fmt.Errorf("error type name cannot be empty")
	}

	// Test error creation
	testErr := errorType.New("test", make(Metadata))
	if testErr == nil {
		return fmt.Errorf("error type failed to create error")
	}

	if testErr.Code() == "" {
		return fmt.Errorf("error type must produce error with code")
	}

	return nil
}
