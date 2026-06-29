package role

import (
	"context"
	"fmt"

	"service/pkg/logger"
	"service/pkg/utils/query"

	rolepPermission "service/internal/master/role/permission"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

// RoleAccessRepository defines the contract for role access data operations
type RoleAccessRepository interface {
	// GetRoleAccess retrieves role access information for a specific user
	GetRoleAccess(ctx context.Context, userID string) (*RoleAccessData, error)
	// GetRolePermissions retrieves permissions for a specific role
	GetRolePermissions(ctx context.Context, roleID int64) ([]*rolepPermission.RolPermission, error)
	// GetRolePages retrieves pages accessible by role
	GetRolePages(ctx context.Context, roleIDs []int64) ([]*RolPages, error)
	// GetUserRoles retrieves roles for a specific user
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

// RoleAccessData represents the complete role access information
type RoleAccessData struct {
	Roles       []string
	Permissions []*rolepPermission.RolPermission
	Pages       []*RolPages
}

type roleAccessRepository struct {
	db     *gorm.DB
	qb     query.QueryBuilder
	dbType query.DBType
}

// NewRoleAccessRepository creates a new instance of RoleAccessRepository
func NewRoleAccessRepository(db *gorm.DB) RoleAccessRepository {
	dbType := query.DBTypePostgreSQL
	qb := query.NewSQLQueryBuilder(dbType).
		SetSecurityOptions(true, 1000).
		SetQueryLogging(true).
		SetQueryTimeout(30)

	return &roleAccessRepository{
		db:     db,
		qb:     qb,
		dbType: dbType,
	}
}

// getSQLXDB extracts *sqlx.DB from *gorm.DB
func (r *roleAccessRepository) getSQLXDB() (*sqlx.DB, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}
	return sqlx.NewDb(sqlDB, "pgx"), nil
}

// GetRoleAccess retrieves complete role access information for a user
func (r *roleAccessRepository) GetRoleAccess(ctx context.Context, userID string) (*RoleAccessData, error) {
	// Get user roles
	roles, err := r.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(roles) == 0 {
		return &RoleAccessData{
			Roles:       []string{},
			Permissions: []*rolepPermission.RolPermission{},
			Pages:       []*RolPages{},
		}, nil
	}

	// For now, we'll use mock role IDs - in real implementation, this should come from database
	// Mock role IDs based on role names
	var roleIDs []int64
	for range roles {
		roleIDs = append(roleIDs, 1) // Mock role ID
	}

	// Get permissions and pages for the roles
	permissions, err := r.GetRolePermissions(ctx, 1) // Mock role ID
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	pages, err := r.GetRolePages(ctx, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get role pages: %w", err)
	}

	return &RoleAccessData{
		Roles:       roles,
		Permissions: permissions,
		Pages:       pages,
	}, nil
}

// GetRolePermissions retrieves permissions for a specific role
func (r *roleAccessRepository) GetRolePermissions(ctx context.Context, roleID int64) ([]*rolepPermission.RolPermission, error) {
	sqlxDB, err := r.getSQLXDB()
	if err != nil {
		return nil, err
	}

	// For now, return all permissions - in real implementation, this should filter by role
	q := query.DynamicQuery{
		From: "rol_permission",
		Filters: []query.FilterGroup{{
			Filters: []query.DynamicFilter{
				query.CreateFilter("DeletedAt", query.OpNull, nil),
			},
		}},
		Limit: 100,
	}

	var results []*rolepPermission.RolPermission
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, q, &results); err != nil {
		return nil, fmt.Errorf("failed to execute role permissions query: %w", err)
	}

	return results, nil
}

// GetRolePages retrieves pages accessible by role
func (r *roleAccessRepository) GetRolePages(ctx context.Context, roleIDs []int64) ([]*RolPages, error) {
	sqlxDB, err := r.getSQLXDB()
	if err != nil {
		return nil, err
	}

	// Get all active pages - in real implementation, this should join with role-page permissions
	q := query.DynamicQuery{
		From: "role_access.rol_pages",
		Filters: []query.FilterGroup{{
			Filters: []query.DynamicFilter{
				query.CreateEqualFilter("Active", true),
			},
		}},
		Sort: []query.SortField{
			query.CreateAscSort("Level"),
			query.CreateAscSort("Sort"),
		},
		Limit: 100,
	}

	var results []*RolPages
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, q, &results); err != nil {
		return nil, fmt.Errorf("failed to execute role pages query: %w", err)
	}

	return results, nil
}

// GetUserRoles retrieves roles for a specific user
func (r *roleAccessRepository) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	// Mock implementation - in real implementation, this should query user-role mapping table
	// For now, return mock roles based on user ID
	logger.Default().Info("Getting user roles", logger.String("userID", userID))

	// Mock roles - in real implementation, get from database
	mockRoles := []string{"admin1", "admin2"}

	return mockRoles, nil
}
