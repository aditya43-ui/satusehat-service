package errors

import (
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
)

// Builder provides fluent API for building errors
type Builder struct {
	code       string
	message    string
	category   string
	httpStatus int
	grpcCode   codes.Code
	metadata   Metadata
	cause      error
}

// NewBuilder creates a new error builder
func NewBuilder() *Builder {
	return &Builder{
		metadata: make(Metadata),
	}
}

// Code sets error code
func (b *Builder) Code(code string) *Builder {
	b.code = code
	info := GetErrorInfo(code)
	b.category = info.Category
	b.httpStatus = info.HTTPStatus
	b.grpcCode = info.GRPCCode
	return b
}

// Message sets error message
func (b *Builder) Message(message string) *Builder {
	b.message = message
	return b
}

// Category sets error category
func (b *Builder) Category(category string) *Builder {
	b.category = category
	return b
}

// HTTPStatus sets HTTP status code
func (b *Builder) HTTPStatus(status int) *Builder {
	b.httpStatus = status
	return b
}

// GRPCCode sets gRPC code
func (b *Builder) GRPCCode(code codes.Code) *Builder {
	b.grpcCode = code
	return b
}

// Metadata adds metadata
func (b *Builder) Metadata(key string, value interface{}) *Builder {
	if b.metadata == nil {
		b.metadata = make(Metadata)
	}
	b.metadata[key] = value
	return b
}

// Metadatas adds multiple metadata
func (b *Builder) Metadatas(metadata Metadata) *Builder {
	if b.metadata == nil {
		b.metadata = make(Metadata)
	}
	for k, v := range metadata {
		b.metadata[k] = v
	}
	return b
}

// Cause sets underlying error
func (b *Builder) Cause(err error) *Builder {
	b.cause = err
	return b
}

// Timestamp adds timestamp
func (b *Builder) Timestamp(t time.Time) *Builder {
	b.Metadata("timestamp", t.Format(time.RFC3339))
	return b
}

// RequestID adds request ID
func (b *Builder) RequestID(id string) *Builder {
	b.Metadata("request_id", id)
	return b
}

// UserID adds user ID
func (b *Builder) UserID(id string) *Builder {
	b.Metadata("user_id", id)
	return b
}

// Service adds service name
func (b *Builder) Service(name string) *Builder {
	b.Metadata("service", name)
	return b
}

// Operation adds operation name
func (b *Builder) Operation(name string) *Builder {
	b.Metadata("operation", name)
	return b
}

// Retryable sets retryable flag
func (b *Builder) Retryable(retryable bool) *Builder {
	b.Metadata("retryable", retryable)
	return b
}

// Severity sets error severity
func (b *Builder) Severity(severity string) *Builder {
	b.Metadata("severity", severity)
	return b
}

// Component adds component name
func (b *Builder) Component(name string) *Builder {
	b.Metadata("component", name)
	return b
}

// Build creates the final error
func (b *Builder) Build() Error {
	// Use defaults if not set
	if b.code == "" {
		b.code = ErrCodeInternalError
	}
	if b.message == "" {
		b.message = "An error occurred"
	}
	if b.category == "" {
		b.category = CategoryInternal
	}
	if b.httpStatus == 0 {
		b.httpStatus = http.StatusInternalServerError
	}
	if b.grpcCode == codes.OK {
		b.grpcCode = codes.Internal
	}
	if b.metadata == nil {
		b.metadata = make(Metadata)
	}
	if _, exists := b.metadata["timestamp"]; !exists {
		b.Metadata("timestamp", time.Now().Format(time.RFC3339))
	}

	return &AppError{
		code:       b.code,
		message:    b.message,
		category:   b.category,
		httpStatus: b.httpStatus,
		grpcCode:   b.grpcCode,
		metadata:   b.metadata,
		cause:      b.cause,
		timestamp:  time.Now(),
		stackTrace: captureStackTrace(),
	}
}

// NewValidationError creates validation error builder
func NewValidationError() *Builder {
	return NewBuilder().
		Code(ErrCodeValidationFailed).
		Category(CategoryValidation).
		HTTPStatus(http.StatusBadRequest).
		GRPCCode(codes.InvalidArgument).
		Severity("warning")
}

// NotFoundError creates not found error builder
func NotFoundError() *Builder {
	return NewBuilder().
		Code(ErrCodeNotFound).
		Category(CategoryNotFound).
		HTTPStatus(http.StatusNotFound).
		GRPCCode(codes.NotFound).
		Severity("info")
}

// UnauthorizedError creates unauthorized error builder
func UnauthorizedError() *Builder {
	return NewBuilder().
		Code(ErrCodeUnauthorized).
		Category(CategoryUnauthorized).
		HTTPStatus(http.StatusUnauthorized).
		GRPCCode(codes.Unauthenticated).
		Severity("warning")
}

// ForbiddenError creates forbidden error builder
func ForbiddenError() *Builder {
	return NewBuilder().
		Code(ErrCodeForbidden).
		Category(CategoryForbidden).
		HTTPStatus(http.StatusForbidden).
		GRPCCode(codes.PermissionDenied).
		Severity("warning")
}

// AlreadyExistsError creates already exists error builder
func AlreadyExistsError() *Builder {
	return NewBuilder().
		Code(ErrCodeConflict).
		Category(CategoryConflict).
		HTTPStatus(http.StatusConflict).
		GRPCCode(codes.AlreadyExists).
		Severity("warning")
}

// DuplicateError creates duplicate entry error builder
func DuplicateError() *Builder {
	return NewBuilder().
		Code(ErrCodeDuplicateEntry).
		Category(CategoryConflict).
		HTTPStatus(http.StatusConflict).
		GRPCCode(codes.AlreadyExists).
		Severity("warning")
}

// ConflictError creates conflict error builder
func ConflictError() *Builder {
	return NewBuilder().
		Code(ErrCodeConflict).
		Category(CategoryConflict).
		HTTPStatus(http.StatusConflict).
		GRPCCode(codes.AlreadyExists).
		Severity("warning")
}

// InternalError creates internal error builder
func InternalError() *Builder {
	return NewBuilder().
		Code(ErrCodeInternalError).
		Category(CategoryInternal).
		HTTPStatus(http.StatusInternalServerError).
		GRPCCode(codes.Internal).
		Severity("error")
}

// ExternalError creates external error builder
func ExternalError() *Builder {
	return NewBuilder().
		Code(ErrCodeExternalError).
		Category(CategoryExternal).
		HTTPStatus(http.StatusBadGateway).
		GRPCCode(codes.Unavailable).
		Severity("error").
		Retryable(true)
}

// TimeoutError creates timeout error builder
func TimeoutError() *Builder {
	return NewBuilder().
		Code(ErrCodeTimeout).
		Category(CategoryTimeout).
		HTTPStatus(http.StatusRequestTimeout).
		GRPCCode(codes.DeadlineExceeded).
		Severity("warning").
		Retryable(true)
}

// BusinessError creates business error builder
func BusinessError() *Builder {
	return NewBuilder().
		Code(ErrCodeBusinessRule).
		Category(CategoryBusiness).
		HTTPStatus(http.StatusBadRequest).
		GRPCCode(codes.FailedPrecondition).
		Severity("warning")
}

// DatabaseError creates database error builder
func DatabaseError() *Builder {
	return NewBuilder().
		Code(ErrCodeDatabaseError).
		Category(CategoryDatabase).
		HTTPStatus(http.StatusInternalServerError).
		GRPCCode(codes.Internal).
		Severity("error")
}

// NetworkError creates network error builder
func NetworkError() *Builder {
	return NewBuilder().
		Code(ErrCodeNetworkError).
		Category(CategoryNetwork).
		HTTPStatus(http.StatusBadGateway).
		GRPCCode(codes.Unavailable).
		Severity("error").
		Retryable(true)
}

// CustomError creates custom error builder
func CustomError(code, message string) *Builder {
	return NewBuilder().
		Code(code).
		Message(message)
}

// FromError creates builder from existing error
func FromErrorBuilder(err error) *Builder {
	appErr := FromError(err)
	if appErr == nil {
		return NewBuilder()
	}

	return NewBuilder().
		Code(appErr.Code()).
		Message(appErr.Error()).
		Category(appErr.Category()).
		HTTPStatus(appErr.HTTPStatus()).
		GRPCCode(appErr.GRPCCode()).
		Metadatas(appErr.Metadata()).
		Cause(appErr.Cause())
}
