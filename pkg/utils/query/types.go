package query

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
)

// DBType represents the type of database
type DBType string

const (
	DBTypePostgreSQL DBType = "postgres"
	DBTypeMySQL      DBType = "mysql"
	DBTypeSQLite     DBType = "sqlite"
	DBTypeSQLServer  DBType = "sqlserver"
	DBTypeMongoDB    DBType = "mongodb"
)

// FilterOperator represents supported filter operators
type FilterOperator string

const (
	OpEqual            FilterOperator = "_eq"
	OpNotEqual         FilterOperator = "_neq"
	OpLike             FilterOperator = "_like"
	OpILike            FilterOperator = "_ilike"
	OpNotLike          FilterOperator = "_nlike"
	OpNotILike         FilterOperator = "_nilike"
	OpIn               FilterOperator = "_in"
	OpNotIn            FilterOperator = "_nin"
	OpGreaterThan      FilterOperator = "_gt"
	OpGreaterThanEqual FilterOperator = "_gte"
	OpLessThan         FilterOperator = "_lt"
	OpLessThanEqual    FilterOperator = "_lte"
	OpBetween          FilterOperator = "_between"
	OpNotBetween       FilterOperator = "_nbetween"
	OpNull             FilterOperator = "_null"
	OpNotNull          FilterOperator = "_nnull"
	OpContains         FilterOperator = "_contains"
	OpNotContains      FilterOperator = "_ncontains"
	OpStartsWith       FilterOperator = "_starts_with"
	OpEndsWith         FilterOperator = "_ends_with"
	OpJsonContains     FilterOperator = "_json_contains"
	OpJsonNotContains  FilterOperator = "_json_ncontains"
	OpJsonExists       FilterOperator = "_json_exists"
	OpJsonNotExists    FilterOperator = "_json_nexists"
	OpJsonEqual        FilterOperator = "_json_eq"
	OpJsonNotEqual     FilterOperator = "_json_neq"
	OpArrayContains    FilterOperator = "_array_contains"
	OpArrayNotContains FilterOperator = "_array_ncontains"
	OpArrayLength      FilterOperator = "_array_length"
)

// DynamicFilter represents a single filter condition
type DynamicFilter struct {
	Column   string         `json:"column"`
	Operator FilterOperator `json:"operator"`
	Value    interface{}    `json:"value"`
	// Additional options for complex filters
	Options map[string]interface{} `json:"options,omitempty"`
}

// FilterGroup represents a group of filters with a logical operator (AND/OR)
type FilterGroup struct {
	Filters []DynamicFilter `json:"filters"`
	LogicOp string          `json:"logic_op"` // AND, OR
}

// SelectField represents a field in the SELECT clause, supporting expressions and aliases
type SelectField struct {
	Expression string `json:"expression"` // e.g., "TMLogBarang.Nama", "COUNT(*)"
	Alias      string `json:"alias"`      // e.g., "obat_nama", "total_count"
	// Window function support
	WindowFunction *WindowFunction `json:"window_function,omitempty"`
}

// WindowFunction represents a window function with its configuration
type WindowFunction struct {
	Function string `json:"function"` // e.g., "ROW_NUMBER", "RANK", "DENSE_RANK", "LEAD", "LAG"
	Over     string `json:"over"`     // PARTITION BY expression
	OrderBy  string `json:"order_by"` // ORDER BY expression
	Frame    string `json:"frame"`    // ROWS/RANGE clause
	Alias    string `json:"alias"`    // Alias for the window function
}

// Join represents a JOIN clause
type Join struct {
	Type         string      `json:"type"`          // "INNER", "LEFT", "RIGHT", "FULL"
	Table        string      `json:"table"`         // Table name to join
	Alias        string      `json:"alias"`         // Table alias
	OnConditions FilterGroup `json:"on_conditions"` // Conditions for the ON clause
	// LATERAL JOIN support
	Lateral bool `json:"lateral,omitempty"`
}

// Union represents a UNION clause
type Union struct {
	Type  string       `json:"type"`  // "UNION", "UNION ALL"
	Query DynamicQuery `json:"query"` // The subquery to union with
}

// CTE (Common Table Expression) represents a WITH clause
type CTE struct {
	Name  string       `json:"name"`  // CTE alias name
	Query DynamicQuery `json:"query"` // The query defining the CTE
	// Recursive CTE support
	Recursive bool `json:"recursive,omitempty"`
}

// DynamicQuery represents the complete query structure
type DynamicQuery struct {
	Fields  []SelectField `json:"fields,omitempty"`
	From    string        `json:"from"`    // Main table name
	Aliases string        `json:"aliases"` // Main table alias
	Joins   []Join        `json:"joins,omitempty"`
	Filters []FilterGroup `json:"filters,omitempty"`
	GroupBy []string      `json:"group_by,omitempty"`
	Having  []FilterGroup `json:"having,omitempty"`
	Unions  []Union       `json:"unions,omitempty"`
	CTEs    []CTE         `json:"ctes,omitempty"`
	Sort    []SortField   `json:"sort,omitempty"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
	// Window function support
	WindowFunctions []WindowFunction `json:"window_functions,omitempty"`
	// JSON operations
	JsonOperations []JsonOperation `json:"json_operations,omitempty"`
}

// JsonOperation represents a JSON operation
type JsonOperation struct {
	Type   string      `json:"type"`            // "extract", "exists", "contains", etc.
	Column string      `json:"column"`          // JSON column
	Path   string      `json:"path"`            // JSON path
	Value  interface{} `json:"value,omitempty"` // Value for comparison
	Alias  string      `json:"alias,omitempty"` // Alias for the result
}

// SortField represents sorting configuration
type SortField struct {
	Column string `json:"column"`
	Order  string `json:"order"` // ASC, DESC
}

// UpdateData represents data for UPDATE operations
type UpdateData struct {
	Columns []string      `json:"columns"`
	Values  []interface{} `json:"values"`
	// JSON update support
	JsonUpdates map[string]JsonUpdate `json:"json_updates,omitempty"`
}

// JsonUpdate represents a JSON update operation
type JsonUpdate struct {
	Path  string      `json:"path"`  // JSON path
	Value interface{} `json:"value"` // New value
}

// InsertData represents data for INSERT operations
type InsertData struct {
	Columns []string      `json:"columns"`
	Values  []interface{} `json:"values"`
	// JSON insert support
	JsonValues map[string]interface{} `json:"json_values,omitempty"`
}

// QueryBuilder interface defines the contract for query builders
type QueryBuilder interface {
	// Configuration methods
	SetSecurityOptions(enableChecks bool, maxRows int) QueryBuilder
	SetAllowedColumns(columns []string) QueryBuilder
	SetAllowedTables(tables []string) QueryBuilder
	SetQueryLogging(enable bool) QueryBuilder
	SetQueryTimeout(timeout time.Duration) QueryBuilder

	// SQL building methods
	BuildQuery(query DynamicQuery) (string, []interface{}, error)
	BuildCountQuery(query DynamicQuery) (string, []interface{}, error)
	BuildInsertQuery(table string, data InsertData, returningColumns ...string) (string, []interface{}, error)
	BuildUpdateQuery(table string, updateData UpdateData, filters []FilterGroup, returningColumns ...string) (string, []interface{}, error)
	BuildDeleteQuery(table string, filters []FilterGroup, returningColumns ...string) (string, []interface{}, error)
	BuildUpsertQuery(table string, insertData InsertData, conflictColumns []string, updateColumns []string, returningColumns ...string) (string, []interface{}, error)
	BuildWhereClause(filterGroups []FilterGroup) (string, []interface{}, error)

	// SQL execution methods
	ExecuteQuery(ctx context.Context, exec sqlx.ExtContext, query DynamicQuery, dest interface{}) error
	ExecuteQueryRow(ctx context.Context, exec sqlx.ExtContext, query DynamicQuery, dest interface{}) error
	ExecuteCount(ctx context.Context, exec sqlx.ExtContext, query DynamicQuery) (int64, error)
	ExecuteInsert(ctx context.Context, exec sqlx.ExtContext, table string, data InsertData, returningColumns ...string) (sql.Result, error)
	ExecuteUpdate(ctx context.Context, exec sqlx.ExtContext, table string, updateData UpdateData, filters []FilterGroup, returningColumns ...string) (sql.Result, error)
	ExecuteDelete(ctx context.Context, exec sqlx.ExtContext, table string, filters []FilterGroup, returningColumns ...string) (sql.Result, error)
	ExecuteUpsert(ctx context.Context, exec sqlx.ExtContext, table string, insertData InsertData, conflictColumns []string, updateColumns []string, returningColumns ...string) (sql.Result, error)
}

// MongoQueryBuilder interface defines the contract for MongoDB query builders
type MongoQueryBuilder interface {
	// Configuration methods
	SetSecurityOptions(enableChecks bool, maxDocs int) MongoQueryBuilder
	SetAllowedFields(fields []string) MongoQueryBuilder
	SetAllowedCollections(collections []string) MongoQueryBuilder
	SetQueryLogging(enable bool) MongoQueryBuilder
	SetQueryTimeout(timeout time.Duration) MongoQueryBuilder

	// MongoDB building methods
	BuildFindQuery(query DynamicQuery) (interface{}, interface{}, error)
	BuildAggregateQuery(query DynamicQuery) (interface{}, error)

	// MongoDB execution methods
	ExecuteFind(ctx context.Context, collection *mongo.Collection, query DynamicQuery, dest interface{}) error
	ExecuteAggregate(ctx context.Context, collection *mongo.Collection, query DynamicQuery, dest interface{}) error
	ExecuteCount(ctx context.Context, collection *mongo.Collection, query DynamicQuery) (int64, error)
	ExecuteInsert(ctx context.Context, collection *mongo.Collection, data InsertData) (*mongo.InsertOneResult, error)
	ExecuteUpdate(ctx context.Context, collection *mongo.Collection, updateData UpdateData, filters []FilterGroup) (*mongo.UpdateResult, error)
	ExecuteDelete(ctx context.Context, collection *mongo.Collection, filters []FilterGroup) (*mongo.DeleteResult, error)
}

// QueryParser interface defines the contract for query parsers
type QueryParser interface {
	SetLimits(defaultLimit, maxLimit int) QueryParser
	ParseQuery(values interface{}, defaultTable string) (DynamicQuery, error)
	ParseQueryWithDefaultFields(values interface{}, defaultTable string, defaultFields []string) (DynamicQuery, error)
}
