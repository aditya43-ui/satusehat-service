package handlers

import (
	"service/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// / handleServiceError mengkonversi error dari service layer ke gRPC status error
func handleServiceError(err error) error {
	appErr := errors.FromError(err)
	if appErr == nil {
		return status.Error(codes.Internal, "internal server error")
	}

	// Mapping kategori error ke gRPC codes
	switch appErr.Category() {
	case errors.CategoryValidation:
		return status.Error(codes.InvalidArgument, appErr.Error())
	case errors.CategoryNotFound:
		return status.Error(codes.NotFound, appErr.Error())
	case errors.CategoryConflict:
		return status.Error(codes.AlreadyExists, appErr.Error())
	case errors.CategoryUnauthorized:
		return status.Error(codes.Unauthenticated, appErr.Error())
	case errors.CategoryForbidden:
		return status.Error(codes.PermissionDenied, appErr.Error())
	default:
		return status.Error(codes.Internal, appErr.Error())
	}
}
