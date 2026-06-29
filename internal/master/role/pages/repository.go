package pages

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"service/internal/infrastructure/database"
	"service/pkg/utils/query"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

// CommandRepository menangani operasi Write (Create, Update, Delete) ke Master DB
type CommandRepository interface {
	Create(ctx context.Context, entity *RolPages) error
	Update(ctx context.Context, entity *RolPages) error
	Delete(ctx context.Context, id int64) error
}

// QueryRepository menangani operasi Read (Find, Search) ke Read-Replica DB
type QueryRepository interface {
	FindByID(ctx context.Context, id int64) (*RolPages, error)
	GetChildren(ctx context.Context, parentID int64) ([]RolPages, error)
	Search(ctx context.Context, filters map[string]interface{}, sorts []query.SortField, limit, offset int) ([]RolPages, int64, error)
}

type repository struct {
	dbManager         database.Service
	dbName            string
	qb                query.QueryBuilder
	dbType            query.DBType
	allowedColumnsMap map[string]bool
}

// Konstruktor Baru CQRS
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
		"name",
		"icon",
		"url",
		"level",
		"sort",
		"parent",
		"active",
		"created_at",
		"updated_at",
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
	return sqlx.NewDb(sqlDB, "pgx"), nil
}

// --- Implementasi Write Model ---
func (r *repository) Create(ctx context.Context, entity *RolPages) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Table(entity.TableName()).Create(entity).Error
}

func (r *repository) Update(ctx context.Context, entity *RolPages) error {
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
	return db.WithContext(ctx).Delete(&RolPages{}, id).Error
}

// --- Implementasi Read Model ---
func (r *repository) FindByID(ctx context.Context, id int64) (*RolPages, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	q := query.DynamicQuery{
		From: "role_access.rol_pages",
		Filters: []query.FilterGroup{{
			Filters: []query.DynamicFilter{query.CreateEqualFilter("id", id)},
		}},
		Limit: 1,
	}

	var entity RolPages
	if err := r.qb.ExecuteQueryRow(ctx, sqlxDB, q, &entity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch RolPages by ID: %w", err)
	}
	return &entity, nil
}

func (r *repository) GetChildren(ctx context.Context, parentID int64) ([]RolPages, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	q := query.DynamicQuery{
		From: "role_access.rol_pages",
		Filters: []query.FilterGroup{{
			Filters: []query.DynamicFilter{query.CreateEqualFilter("parent", parentID)},
		}},
		Sort: []query.SortField{query.CreateAscSort("sort")},
	}

	var entities []RolPages
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, q, &entities); err != nil {
		return nil, fmt.Errorf("failed to fetch RolPages children: %w", err)
	}
	return entities, nil
}

func (r *repository) Search(ctx context.Context, filters map[string]interface{}, sorts []query.SortField, limit, offset int) ([]RolPages, int64, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, 0, err
	}

	var dynamicFilters []query.DynamicFilter
	for k, v := range filters {
		// Normalisasi nama kolom menjadi lowercase agar sesuai dengan allowedColumns
		colName := strings.ToLower(k)
		if !r.isColumnAllowed(colName) {
			continue
		}
		switch val := v.(type) {
		case string:
			if val != "" {
				// Gunakan OpILike hanya untuk kolom yang bertipe teks (string)
				if colName == "name" || colName == "icon" || colName == "url" {
					dynamicFilters = append(dynamicFilters, query.CreateFilter(colName, query.OpILike, "%"+val+"%"))
				} else {
					if boolVal, err := strconv.ParseBool(val); err == nil && (val == "true" || val == "false" || val == "1" || val == "0") {
						dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, boolVal))
					} else if intVal, err := strconv.Atoi(val); err == nil {
						dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, intVal))
					} else {
						dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, val))
					}
				}
			}
		default:
			dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, val))
		}
	}

	if len(sorts) == 0 {
		sorts = []query.SortField{query.CreateAscSort("id")}
	}

	q := query.DynamicQuery{
		From:    "role_access.rol_pages",
		Fields:  []query.SelectField{{Expression: "*"}},
		Filters: []query.FilterGroup{{Filters: dynamicFilters}},
		Limit:   limit,
		Offset:  offset,
		Sort:    sorts,
	}

	var entities []RolPages
	if err := r.qb.ExecuteQuery(ctx, sqlxDB, q, &entities); err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}

	count, err := r.qb.ExecuteCount(ctx, sqlxDB, q)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search count query: %w", err)
	}

	return entities, count, nil
}
