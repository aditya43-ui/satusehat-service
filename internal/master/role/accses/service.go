package role

import (
	"context"
	"service/pkg/errors"
)

// RoleAccessService defines the contract for role access operations
type RoleAccessService interface {
	GetRoleAccess(ctx context.Context, userID string) (*RoleAccessResponse, error)
}

// roleAccessService implements RoleAccessService
type roleAccessService struct {
	repo RoleAccessRepository
}

// NewRoleAccessService creates a new instance of RoleAccessService
func NewRoleAccessService(repo RoleAccessRepository) RoleAccessService {
	return &roleAccessService{
		repo: repo,
	}
}

// GetRoleAccess retrieves role access information for a user
func (s *roleAccessService) GetRoleAccess(ctx context.Context, userID string) (*RoleAccessResponse, error) {
	if userID == "" {
		return nil, errors.NewValidationError().Message("User ID is required").Metadata("userID", userID).Build()
	}

	// Get role access data from repository
	roleAccessData, err := s.repo.GetRoleAccess(ctx, userID)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to retrieve role access data").Cause(err).Build()
	}

	// Map the data to response format
	response := MapPermissionToRoleAccessResponse(
		roleAccessData.Roles,
		roleAccessData.Permissions,
		roleAccessData.Pages,
	)

	return response, nil
}

// GetUserRoles retrieves roles for a specific user
func (s *roleAccessService) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, errors.NewValidationError().Message("User ID is required").Metadata("userID", userID).Build()
	}

	return s.repo.GetUserRoles(ctx, userID)
}
