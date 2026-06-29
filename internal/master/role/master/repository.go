package master

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
	"gorm.io/gorm/clause"
)

type CommandRepository interface {
	Create(ctx context.Context, entity *RoleMaster) error
	Update(ctx context.Context, entity *RoleMaster) error
	Upsert(ctx context.Context, entity *RoleMaster) error
	Delete(ctx context.Context, id int64) error
}

type QueryRepository interface {
	FindAll(ctx context.Context, limit, offset int) ([]RoleMaster, int64, error)
	FindByID(ctx context.Context, id int64) (*RoleMaster, error)
	Search(ctx context.Context, filters map[string]interface{}, sorts []query.SortField, limit, offset int) ([]RoleMaster, int64, error)
}

type repository struct {
	dbManager         database.Service
	dbName            string
	qb                query.QueryBuilder
	dbType            query.DBType
	allowedColumnsMap map[string]bool
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
		"name",
		"active",
		"created_at",
		"updated_at",
	}
	qb := query.NewSQLQueryBuilder(dbType).
		SetSecurityOptions(true, 1000).
		SetQueryLogging(true).
		SetQueryTimeout(30).
		SetAllowedColumns(allowedColumns)

	allowedMap := make(map[string]bool)
	for _, col := range allowedColumns {
		allowedMap[col] = true
	}

	return &repository{dbManager: dbManager, dbName: dbName, qb: qb, dbType: dbType, allowedColumnsMap: allowedMap}
}

// isColumnAllowed memvalidasi apakah kolom diizinkan untuk digunakan
func (r *repository) isColumnAllowed(column string) bool {
	return r.allowedColumnsMap[column]
}

// getWriteGormDB extracts *gorm.DB for Write/Command operations
func (r *repository) getWriteGormDB() (*gorm.DB, error) {
	return r.dbManager.GetGormDB(r.dbName)
}

// getReadSQLXDB extracts *sqlx.DB from read replicas for Read/Query operations
func (r *repository) getReadSQLXDB() (*sqlx.DB, error) {
	db, err := r.dbManager.GetReadDB(r.dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get read db: %w", err)
	}
	// Gunakan Unsafe() agar sqlx mengabaikan kolom hasil query yang tidak terdapat di dalam struct
	return sqlx.NewDb(db, "pgx").Unsafe(), nil
}

// FindAll fetches all RoleMaster with pagination
func (r *repository) FindAll(ctx context.Context, limit, offset int) ([]RoleMaster, int64, error) {
	db, err := r.getReadSQLXDB()
	if err != nil {
		return nil, 0, err
	}

	dq := query.DynamicQuery{
		From:  "role_access.rol_master",
		Limit: limit, Offset: offset,
		Sort: []query.SortField{query.CreateAscSort("id")},
	}

	var results []RoleMaster
	if err := r.qb.ExecuteQuery(ctx, db, dq, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to execute find all query: %w", err)
	}

	count, err := r.qb.ExecuteCount(ctx, db, dq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	return results, count, nil
}

// FindByID fetches a single RoleMaster by ID
func (r *repository) FindByID(ctx context.Context, id int64) (*RoleMaster, error) {
	db, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	var result RoleMaster
	q := query.DynamicQuery{
		From: "role_access.rol_master",
		Filters: []query.FilterGroup{{
			Filters: []query.DynamicFilter{
				query.CreateEqualFilter("id", id),
			},
		}},
		Limit: 1,
	}

	if err := r.qb.ExecuteQueryRow(ctx, db, q, &result); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil, nil jika data tidak ada, jangan lempar error query
		}
		// Wrapping error origin
		return nil, fmt.Errorf("failed to fetch RoleAccessRolMaster: %w", err)
	}
	return &result, nil
}

// Search fetches RoleMaster based on dynamic filters and sorting
func (r *repository) Search(ctx context.Context, filters map[string]interface{}, sorts []query.SortField, limit, offset int) ([]RoleMaster, int64, error) {
	db, err := r.getReadSQLXDB()
	if err != nil {
		return nil, 0, err
	}

	var dynamicFilters []query.DynamicFilter
	for k, v := range filters {
		colName := strings.ToLower(k)
		if !r.isColumnAllowed(colName) {
			continue
		}

		switch val := v.(type) {
		case string:
			if val != "" {
				if colName == "name" {
					dynamicFilters = append(dynamicFilters, query.CreateFilter(colName, query.OpILike, "%"+val+"%"))
				} else if colName == "active" {
					if boolVal, err := strconv.ParseBool(val); err == nil && (val == "true" || val == "false" || val == "1" || val == "0") {
						dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, boolVal))
					} else {
						dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, val))
					}
				} else {
					if intVal, err := strconv.Atoi(val); err == nil {
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

	var sortFields []query.SortField
	for _, sort := range sorts {
		colName := strings.ToLower(sort.Column)
		if r.isColumnAllowed(colName) {
			sortFields = append(sortFields, query.SortField{
				Column: colName,
				Order:  sort.Order,
			})
		}
	}

	// Jika tidak ada sort yang valid, gunakan default
	if len(sortFields) == 0 {
		sortFields = []query.SortField{query.CreateAscSort("id")}
	}

	q := query.DynamicQuery{
		From: "role_access.rol_master",
		Fields: []query.SelectField{
			{Expression: "*"},
		},
		Filters: []query.FilterGroup{{Filters: dynamicFilters}},
		Limit:   limit, Offset: offset,
		Sort: sortFields,
	}

	logger.Default().Info("Built search query", logger.String("request", fmt.Sprintf("%+v", q)))
	var results []RoleMaster
	if err := r.qb.ExecuteQuery(ctx, db, q, &results); err != nil {
		return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
	}

	count, err := r.qb.ExecuteCount(ctx, db, q)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search count query: %w", err)
	}

	return results, count, nil
}

// Create inserts a new RoleMaster record
func (r *repository) Create(ctx context.Context, entity *RoleMaster) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Table(entity.TableName()).Create(entity).Error
}

func (r *repository) Update(ctx context.Context, entity *RoleMaster) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Table(entity.TableName()).Save(entity).Error
}

func (r *repository) Upsert(ctx context.Context, entity *RoleMaster) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Table(entity.TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}}, // Kolom acuan untuk konflik
		UpdateAll: true,                          // Update seluruh kolom jika ada konflik
	}).Create(entity).Error
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Delete(&RoleMaster{}, id).Error
}
