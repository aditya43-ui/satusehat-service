package permission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"service/internal/infrastructure/database"
	"service/pkg/logger"
	"service/pkg/utils/query"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

type CommandRepository interface {
	Create(ctx context.Context, entity *RolPermission) error
	Update(ctx context.Context, entity *RolPermission) error
	Delete(ctx context.Context, id int64) error
}

type QueryRepository interface {
	FindAll(ctx context.Context, limit, offset int) ([]RolPermission, int64, error)
	FindByID(ctx context.Context, id int64) (*RolPermission, error)
	Search(ctx context.Context, filters map[string]interface{}, sorts []query.SortField, limit, offset int) ([]RolPermission, int64, error)
	FindByRoleAndGroup(ctx context.Context, roleKeycloak string, groupKeycloak []string) ([]*RolPermission, error)
	FindByRoleAndPages(ctx context.Context, roleKeycloak string, pageIDs []int64) ([]*RolPermission, error)
	FindPermissionsWithPages(ctx context.Context, roleKeycloak string, groupKeycloak []string, activeOnly *bool) ([]*RolPermissionWithPage, error)
	FindPermissionsWithRole(ctx context.Context, roleKeycloak string, activeOnly *bool) ([]*RolPermissionWithPage, error)
	FindActivePagesWithPermission(ctx context.Context, active bool) ([]*RolPermissionWithPage, error)
}

type repository struct {
	dbManager         database.Service
	dbName            string
	qb                query.QueryBuilder
	dbType            query.DBType
	allowedColumnsMap map[string]bool // Cache untuk validasi kolom yang diizinkan
}

func NewCommandRepository(dbManager database.Service, dbName string) CommandRepository {
	return NewRepository(dbManager, dbName)
}

func NewQueryRepository(dbManager database.Service, dbName string) QueryRepository {
	return NewRepository(dbManager, dbName)
}

func NewRepository(dbManager database.Service, dbName string) *repository {
	dbType := query.DBTypePostgreSQL
	allowedColumns := []string{
		"id",
		"create",
		"read",
		"update",
		"disable",
		"delete",
		"active",
		"fk_rol_pages_id",
		"role_keycloak",
		"group_keycloak",
		"created_at",
		"updated_at",
		"role_master_name",
	}
	qb := query.NewSQLQueryBuilder(dbType).
		SetSecurityOptions(true, 1000).
		SetQueryLogging(true).
		SetQueryTimeout(30).
		SetAllowedColumns(allowedColumns)
	return &repository{dbManager: dbManager, dbName: dbName, qb: qb, dbType: dbType, allowedColumnsMap: createAllowedColumnsMap(allowedColumns)}
}

// Helper function untuk membuat map dari allowed columns
func createAllowedColumnsMap(columns []string) map[string]bool {
	result := make(map[string]bool)
	for _, col := range columns {
		result[col] = true
	}
	return result
}

// isColumnAllowed memvalidasi apakah kolom diizinkan untuk digunakan
func (r *repository) isColumnAllowed(column string) bool {
	return r.allowedColumnsMap[column]
}

// Helper mengambil koneksi Master dari DB Manager
func (r *repository) getWriteGormDB() (*gorm.DB, error) {
	return r.dbManager.GetGormDB(r.dbName)
}

// Helper mengambil koneksi Read-Replica dari DB Manager
func (r *repository) getReadSQLXDB() (*sqlx.DB, error) {
	sqlDB, err := r.dbManager.GetReadDB(r.dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get read db: %w", err)
	}
	return sqlx.NewDb(sqlDB, "pgx").Unsafe(), nil
}

// FindAll fetches all RolPermission with pagination
func (r *repository) FindAll(ctx context.Context, limit, offset int) ([]RolPermission, int64, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, 0, err
	}

	// Build query dengan filter active = true
	dq := query.DynamicQuery{
		From:    "role_access.rol_permission",
		Aliases: "rp",
		Fields: []query.SelectField{
			{Expression: "rp.*"},
			{Expression: "rm.name", Alias: "role_master_name"},
		},
		Joins: []query.Join{
			{
				Type:  "INNER",
				Table: "role_access.rol_master",
				Alias: "rm",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("rm.id", "rp.role_keycloak"),
					},
				},
			},
		},
		Limit:  limit,
		Offset: offset,
		Sort:   []query.SortField{query.CreateDescSort("rp.created_at")},
	}

	var results []RolPermission
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, dq, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to execute find all query: %w", err)
	}

	count, err := r.qb.ExecuteCount(ctx, sqlxDB, dq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	return results, count, nil
}

// FindByID fetches a single RolPermission by ID
func (r *repository) FindByID(ctx context.Context, id int64) (*RolPermission, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	// Build query dengan filter ID dan active = true
	dq := query.DynamicQuery{
		From:    "role_access.rol_permission",
		Aliases: "rp",
		Fields: []query.SelectField{
			{Expression: "rp.*"},
			{Expression: "rm.name", Alias: "role_master_name"},
		},
		Joins: []query.Join{
			{
				Type:  "INNER",
				Table: "role_access.rol_master",
				Alias: "rm",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("rm.id", "rp.role_keycloak"),
					},
				},
			},
		},
		Filters: []query.FilterGroup{{
			Filters: []query.DynamicFilter{
				query.CreateEqualFilter("rp.id", id),
			},
		}},
		Limit: 1,
	}

	var result RolPermission
	if err := r.qb.ExecuteQueryRow(ctx, sqlxDB, dq, &result); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch RolPermission: %w", err)
	}
	return &result, nil
}
func (r *repository) Search(ctx context.Context, filters map[string]interface{}, sorts []query.SortField, limit, offset int) ([]RolPermission, int64, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, 0, err
	}

	// Build dynamic filters dengan validasi kolom yang aman
	var dynamicFilters []query.DynamicFilter

	for k, v := range filters {
		// Normalisasi nama kolom menjadi lowercase agar sesuai dengan allowedColumns
		colName := strings.ToLower(k)
		if !r.isColumnAllowed(colName) {
			continue
		}

		dbCol := "rp." + colName
		if colName == "role_master_name" {
			dbCol = "rm.name"
		}

		switch val := v.(type) {
		case string:
			if val != "" {
				if colName == "role_keycloak" || colName == "group_keycloak" || colName == "role_master_name" {
					dynamicFilters = append(dynamicFilters, query.CreateFilter(dbCol, query.OpILike, "%"+val+"%"))
				} else {
					if boolVal, err := strconv.ParseBool(val); err == nil && (val == "true" || val == "false" || val == "1" || val == "0") {
						dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(dbCol, boolVal))
					} else if intVal, err := strconv.Atoi(val); err == nil {
						dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(dbCol, intVal))
					} else {
						dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(dbCol, val))
					}
				}
			}
		default:
			dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(dbCol, val))
		}
	}

	// Konversi sort fields dengan validasi keamanan
	var sortFields []query.SortField
	for _, sort := range sorts {
		if r.isColumnAllowed(sort.Column) {
			dbCol := "rp." + sort.Column
			if strings.ToLower(sort.Column) == "role_master_name" {
				dbCol = "rm.name"
			}
			sortFields = append(sortFields, query.SortField{
				Column: dbCol,
				Order:  sort.Order,
			})
		}
	}

	// Jika tidak ada sort yang valid, gunakan default
	if len(sortFields) == 0 {
		sortFields = []query.SortField{query.CreateDescSort("rp.created_at")}
	}

	// Build query dengan query builder
	q := query.DynamicQuery{
		From:    "role_access.rol_permission",
		Aliases: "rp",
		Fields: []query.SelectField{
			{Expression: "rp.*"},
			{Expression: "rm.name", Alias: "role_master_name"},
		},
		Joins: []query.Join{
			{
				Type:  "INNER",
				Table: "role_access.rol_master",
				Alias: "rm",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("rm.id", "rp.role_keycloak"),
					},
				},
			},
		},
		Filters: []query.FilterGroup{{Filters: dynamicFilters}},
		Limit:   limit,
		Offset:  offset,
		Sort:    sortFields,
	}

	logger.Default().Info("Built search query", logger.String("request", fmt.Sprintf("%+v", q)))

	var results []RolPermission
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, q, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}

	count, err := r.qb.ExecuteCount(ctx, sqlxDB, q)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search count query: %w", err)
	}

	return results, count, nil
}

func (r *repository) FindByRoleAndGroup(ctx context.Context, roleKeycloak string, groupKeycloak []string) ([]*RolPermission, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	// Build dynamic filters
	var dynamicFilters []query.DynamicFilter

	// Filter berdasarkan role
	if roleKeycloak != "" {
		dynamicFilters = append(dynamicFilters, query.CreateEqualFilter("rp.role_keycloak", roleKeycloak))
	}

	// Filter berdasarkan group jika disediakan
	if len(groupKeycloak) > 0 {
		dynamicFilters = append(dynamicFilters, query.CreateFilter("rp.group_keycloak", query.OpIn, groupKeycloak))
	}

	// Jika tidak ada role maupun group, return empty result
	if roleKeycloak == "" && len(groupKeycloak) == 0 {
		return []*RolPermission{}, nil
	}

	// Build query dengan query builder
	dq := query.DynamicQuery{
		From:    "role_access.rol_permission",
		Aliases: "rp",
		Fields: []query.SelectField{
			{Expression: "rp.*"},
			{Expression: "rm.name", Alias: "role_master_name"},
		},
		Joins: []query.Join{
			{
				Type:  "INNER",
				Table: "role_access.rol_master",
				Alias: "rm",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("rm.id", "rp.role_keycloak"),
					},
				},
			},
		},
		Filters: []query.FilterGroup{{Filters: dynamicFilters}},
	}

	var results []*RolPermission
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, dq, &results); err != nil {
		return nil, fmt.Errorf("failed to execute find by role and group query: %w", err)
	}

	return results, nil
}

// Method baru untuk query dengan JOIN - lebih efisien
func (r *repository) FindPermissionsWithPages(ctx context.Context, roleKeycloak string, groupKeycloak []string, activeOnly *bool) ([]*RolPermissionWithPage, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	// Build dynamic filters
	var dynamicFilters []query.DynamicFilter

	// Filter berdasarkan role
	if roleKeycloak != "" {
		dynamicFilters = append(dynamicFilters, query.CreateEqualFilter("rp.role_keycloak", roleKeycloak))
	}

	// Filter berdasarkan group jika disediakan
	if len(groupKeycloak) > 0 {
		dynamicFilters = append(dynamicFilters, query.CreateFilter("rp.group_keycloak", query.OpIn, groupKeycloak))
	}

	// Filter active jika diminta
	if activeOnly != nil {
		dynamicFilters = append(dynamicFilters, query.CreateEqualFilter("rp.active", *activeOnly))
	}

	// Build query dengan JOIN menggunakan query builder
	dq := query.DynamicQuery{
		From:    "role_access.rol_permission",
		Aliases: "rp",
		Fields: []query.SelectField{
			{Expression: "rp.*"},
			{Expression: "rm.name", Alias: "role_master_name"},
			{Expression: "p.name", Alias: "page_name"},
			{Expression: "p.url", Alias: "page_url"},
			{Expression: "p.level", Alias: "page_level"},
			{Expression: "p.sort", Alias: "page_sort"},
			{Expression: "p.active", Alias: "page_active"},
			{Expression: "p.icon", Alias: "page_icon"},
			{Expression: "p.parent", Alias: "page_parent"},
		},
		Joins: []query.Join{
			{
				Type:  "INNER",
				Table: "role_access.rol_master",
				Alias: "rm",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("rm.id", "rp.role_keycloak"),
					},
				},
			},
			{
				Type:  "LEFT",
				Table: "role_access.rol_pages",
				Alias: "p",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("p.id", "rp.fk_rol_pages_id"),
					},
				},
			},
		},
		Filters: []query.FilterGroup{{Filters: dynamicFilters}},
		Sort:    []query.SortField{query.CreateAscSort("p.sort"), query.CreateAscSort("p.id")},
	}

	var results []*RolPermissionWithPage
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, dq, &results); err != nil {
		return nil, fmt.Errorf("failed to execute permissions with pages query: %w", err)
	}

	return results, nil
}

// Method baru untuk query dengan JOIN - lebih efisien
func (r *repository) FindPermissionsWithRole(ctx context.Context, roleKeycloak string, activeOnly *bool) ([]*RolPermissionWithPage, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	// Build dynamic filters
	var dynamicFilters []query.DynamicFilter

	// Filter berdasarkan role
	if roleKeycloak != "" {
		dynamicFilters = append(dynamicFilters, query.CreateEqualFilter("rp.role_keycloak", roleKeycloak))
	}

	// Filter active jika diminta
	if activeOnly != nil {
		dynamicFilters = append(dynamicFilters, query.CreateEqualFilter("rp.active", *activeOnly))
	}

	// Build query dengan JOIN menggunakan query builder
	dq := query.DynamicQuery{
		From:    "role_access.rol_permission",
		Aliases: "rp",
		Fields: []query.SelectField{
			{Expression: "rp.*"},
			{Expression: "rm.name", Alias: "role_master_name"},
			{Expression: "p.name", Alias: "page_name"},
			{Expression: "p.url", Alias: "page_url"},
			{Expression: "p.level", Alias: "page_level"},
			{Expression: "p.sort", Alias: "page_sort"},
			{Expression: "p.active", Alias: "page_active"},
			{Expression: "p.icon", Alias: "page_icon"},
			{Expression: "p.parent", Alias: "page_parent"},
		},
		Joins: []query.Join{
			{
				Type:  "INNER",
				Table: "role_access.rol_master",
				Alias: "rm",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("rm.id", "rp.role_keycloak"),
					},
				},
			},
			{
				Type:  "LEFT",
				Table: "role_access.rol_pages",
				Alias: "p",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("p.id", "rp.fk_rol_pages_id"),
					},
				},
			},
		},
		Filters: []query.FilterGroup{{Filters: dynamicFilters}},
		Sort:    []query.SortField{query.CreateAscSort("p.sort"), query.CreateAscSort("p.id")},
	}

	var results []*RolPermissionWithPage
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, dq, &results); err != nil {
		return nil, fmt.Errorf("failed to execute permissions with pages query: %w", err)
	}

	return results, nil
}

// FindActivePagesWithPermission menjalankan Query "dataAll"
func (r *repository) FindActivePagesWithPermission(ctx context.Context, active bool) ([]*RolPermissionWithPage, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	dq := query.DynamicQuery{
		From:    "role_access.rol_pages",
		Aliases: "p",
		Fields: []query.SelectField{
			{Expression: "p.id", Alias: "fk_rol_pages_id"},
			{Expression: "p.name", Alias: "page_name"},
			{Expression: "p.icon", Alias: "page_icon"},
			{Expression: "p.url", Alias: "page_url"},
			{Expression: "p.level", Alias: "page_level"},
			{Expression: "p.sort", Alias: "page_sort"},
			{Expression: "p.parent", Alias: "page_parent"},
			{Expression: "p.active", Alias: "page_active"},
			{Expression: "p.created_at", Alias: "created_at"},
			{Expression: "p.updated_at", Alias: "updated_at"},
			{Expression: "COALESCE(BOOL_OR(rp.create), false)", Alias: "create"},
			{Expression: "COALESCE(BOOL_OR(rp.read), false)", Alias: "read"},
			{Expression: "COALESCE(BOOL_OR(rp.update), false)", Alias: "update"},
			{Expression: "COALESCE(BOOL_OR(rp.disable), false)", Alias: "disable"},
			{Expression: "COALESCE(BOOL_OR(rp.delete), false)", Alias: "delete"},
			{Expression: "COALESCE(BOOL_OR(rp.active), false)", Alias: "active"},
		},
		Joins: []query.Join{
			{
				Type:  "LEFT",
				Table: "role_access.rol_permission",
				Alias: "rp",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("p.id", "rp.fk_rol_pages_id"),
					},
				},
			},
		},
		GroupBy: []string{
			"p.id",
		},
		Filters: []query.FilterGroup{
			{
				Filters: []query.DynamicFilter{
					query.CreateEqualFilter("p.active", active),
				},
			},
		},
		Sort: []query.SortField{
			query.CreateAscSort("p.id"),
		},
	}

	var results []*RolPermissionWithPage
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, dq, &results); err != nil {
		return nil, fmt.Errorf("failed to execute active pages query: %w", err)
	}
	return results, nil
}

func (r *repository) FindByRoleAndPages(ctx context.Context, roleOrGroupKeycloak string, pageIDs []int64) ([]*RolPermission, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	// Build dynamic filters
	var filterGroups []query.FilterGroup

	// Filter berdasarkan role atau group (OR logic)
	if roleOrGroupKeycloak != "" {
		roleFilter := query.CreateEqualFilter("rp.role_keycloak", roleOrGroupKeycloak)
		groupFilter := query.CreateEqualFilter("rp.group_keycloak", roleOrGroupKeycloak)
		// Buat filter group dengan OR logic
		filterGroups = append(filterGroups, query.FilterGroup{
			Filters: []query.DynamicFilter{roleFilter, groupFilter},
			LogicOp: "OR",
		})
	}

	// Filter berdasarkan page IDs jika disediakan
	if len(pageIDs) > 0 {
		pageFilter := query.CreateFilter("rp.fk_rol_pages_id", query.OpIn, pageIDs)
		filterGroups = append(filterGroups, query.FilterGroup{
			Filters: []query.DynamicFilter{pageFilter},
		})
	}

	// Build query dengan query builder
	dq := query.DynamicQuery{
		From:    "role_access.rol_permission",
		Aliases: "rp",
		Fields: []query.SelectField{
			{Expression: "rp.*"},
			{Expression: "rm.name", Alias: "role_master_name"},
		},
		Joins: []query.Join{
			{
				Type:  "INNER",
				Table: "role_access.rol_master",
				Alias: "rm",
				OnConditions: query.FilterGroup{
					Filters: []query.DynamicFilter{
						query.CreateEqualFilter("rm.id", "rp.role_keycloak"),
					},
				},
			},
		},
		Filters: filterGroups,
	}

	var results []*RolPermission
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, dq, &results); err != nil {
		return nil, fmt.Errorf("failed to execute find by role and pages query: %w", err)
	}

	return results, nil
}

// Create creates a new RolPermission
func (r *repository) Create(ctx context.Context, entity *RolPermission) error {
	if entity == nil {
		return errors.New("entity cannot be nil")
	}

	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}

	result := db.WithContext(ctx).Table(entity.TableName()).Create(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to create RolPermission: %w", result.Error)
	}

	return nil
}

func (r *repository) Update(ctx context.Context, entity *RolPermission) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Table(entity.TableName()).Save(entity).Error
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Delete(&RolPermission{}, id).Error
}
