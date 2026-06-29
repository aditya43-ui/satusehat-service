package errors

import (
	"context"
	"net/http"

	"service/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCErrorInterceptor creates unary interceptor for error handling
func GRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, HandleGRPCError(err)
		}
		return resp, nil
	}
}

// GRPCStreamInterceptor creates stream interceptor for error handling
func GRPCStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if err != nil {
			return HandleGRPCError(err)
		}
		return nil
	}
}

// HandleGRPCError converts error to gRPC error
func HandleGRPCError(err error) error {
	appErr := FromError(err)
	if appErr == nil {
		return status.Error(codes.Internal, "Internal server error")
	}

	// Log gRPC error
	LogGRPCError(appErr)

	// Convert to gRPC status
	return status.Error(appErr.GRPCCode(), appErr.GetLocalizedMessage("en"))
}

// LogGRPCError logs gRPC error
func LogGRPCError(err Error) {
	log := logger.Default()

	log.Error("gRPC error occurred",
		logger.String("error_code", err.Code()),
		logger.String("category", err.Category()),
		logger.String("grpc_code", err.GRPCCode().String()),
		logger.Any("metadata", err.Metadata()),
	)
}

// FromGRPCError converts gRPC error to AppError
func FromGRPCError(err error) Error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return NewWithCode(ErrCodeInternalError, err.Error())
	}

	// Map gRPC codes to error codes
	code := mapGRPCToErrorCode(st.Code())
	appErr := NewWithCode(code, st.Message())

	// Add gRPC metadata
	appErr = appErr.WithMetadata("grpc_code", st.Code().String())

	return appErr
}

// mapGRPCToErrorCode maps gRPC codes to error codes
func mapGRPCToErrorCode(grpcCode codes.Code) string {
	switch grpcCode {
	case codes.OK:
		return ""
	case codes.Canceled:
		return ErrCodeTimeout
	case codes.Unknown:
		return ErrCodeInternalError
	case codes.InvalidArgument:
		return ErrCodeValidationFailed
	case codes.DeadlineExceeded:
		return ErrCodeTimeout
	case codes.NotFound:
		return ErrCodeNotFound
	case codes.AlreadyExists:
		return ErrCodeDuplicateEntry
	case codes.PermissionDenied:
		return ErrCodeForbidden
	case codes.ResourceExhausted:
		return ErrCodeRateLimitExceeded
	case codes.FailedPrecondition:
		return ErrCodeBusinessRule
	case codes.Aborted:
		return ErrCodeConflict
	case codes.OutOfRange:
		return ErrCodeInvalidRange
	case codes.Unimplemented:
		return ErrCodeServiceUnavailable
	case codes.Internal:
		return ErrCodeInternalError
	case codes.Unavailable:
		return ErrCodeExternalError
	case codes.DataLoss:
		return ErrCodeInternalError
	case codes.Unauthenticated:
		return ErrCodeUnauthorized
	default:
		return ErrCodeInternalError
	}
}

// GRPCMiddlewareConfig configures gRPC middleware
type GRPCMiddlewareConfig struct {
	SkipMethods       []string
	LogAllErrors      bool
	SanitizeErrors    bool
	DefaultLanguage   string
	EnableMetrics     bool
	CustomCodeMapping map[codes.Code]string
}

// DefaultGRPCMiddlewareConfig returns default gRPC configuration
func DefaultGRPCMiddlewareConfig() *GRPCMiddlewareConfig {
	return &GRPCMiddlewareConfig{
		SkipMethods:     []string{"/grpc.health.v1.Health/Check"},
		LogAllErrors:    true,
		SanitizeErrors:  true,
		DefaultLanguage: "en",
		EnableMetrics:   true,
	}
}

// GRPCUnaryInterceptorWithConfig creates interceptor with custom configuration
func GRPCUnaryInterceptorWithConfig(config *GRPCMiddlewareConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip specified methods
		for _, method := range config.SkipMethods {
			if info.FullMethod == method {
				return handler(ctx, req)
			}
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return nil, handleGRPCErrorWithConfig(err, config)
		}
		return resp, nil
	}
}

// handleGRPCErrorWithConfig handles gRPC error with custom configuration
func handleGRPCErrorWithConfig(err error, config *GRPCMiddlewareConfig) error {
	appErr := FromError(err)
	if appErr == nil {
		return status.Error(codes.Internal, "Internal server error")
	}

	// Apply custom code mapping
	if customCode, exists := config.CustomCodeMapping[appErr.GRPCCode()]; exists {
		appErr = NewWithCode(customCode, appErr.Error())
	}

	// Sanitize error if needed
	if config.SanitizeErrors {
		appErr = sanitizeError(appErr).(Error)
	}

	// Log error
	if config.LogAllErrors {
		LogGRPCError(appErr)
	}

	return status.Error(appErr.GRPCCode(), appErr.GetLocalizedMessage(config.DefaultLanguage))
}

// ConvertToHTTPStatus converts gRPC code to HTTP status
func ConvertToHTTPStatus(grpcCode codes.Code) int {
	switch grpcCode {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// ConvertToGRPCCode converts HTTP status to gRPC code
func ConvertToGRPCCode(httpStatus int) codes.Code {
	switch httpStatus {
	case http.StatusOK:
		return codes.OK
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusRequestTimeout:
		return codes.DeadlineExceeded
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusInternalServerError:
		return codes.Internal
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	case http.StatusBadGateway:
		return codes.Unavailable
	default:
		return codes.Unknown
	}
}
