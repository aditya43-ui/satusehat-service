package query

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
)

// Main entry point for backward compatibility
// This file provides the main interface and factory functions

// QueryBuilderFactory creates query builders for different database types
type QueryBuilderFactory interface {
	CreateSQLQueryBuilder(dbType DBType) QueryBuilder
	CreateMongoQueryBuilder() MongoQueryBuilder
	CreateQueryParser() QueryParser
}

// DefaultQueryBuilderFactory implements QueryBuilderFactory
type DefaultQueryBuilderFactory struct{}

// NewQueryBuilderFactory creates a new query builder factory
func NewQueryBuilderFactory() QueryBuilderFactory {
	return &DefaultQueryBuilderFactory{}
}

// CreateSQLQueryBuilder creates a new SQL query builder
func (f *DefaultQueryBuilderFactory) CreateSQLQueryBuilder(dbType DBType) QueryBuilder {
	return NewSQLQueryBuilder(dbType)
}

// CreateMongoQueryBuilder creates a new MongoDB query builder
func (f *DefaultQueryBuilderFactory) CreateMongoQueryBuilder() MongoQueryBuilder {
	return NewMongoQueryBuilder()
}

// CreateQueryParser creates a new query parser
func (f *DefaultQueryBuilderFactory) CreateQueryParser() QueryParser {
	return NewQueryParser()
}

// QueryManager provides a unified interface for all query operations
type QueryManager struct {
	SQLBuilder   QueryBuilder
	MongoBuilder MongoQueryBuilder
	QueryParser  QueryParser
}

// NewQueryManager creates a new query manager with all builders
func NewQueryManager(dbType DBType) *QueryManager {
	return &QueryManager{
		SQLBuilder:   NewSQLQueryBuilder(dbType),
		MongoBuilder: NewMongoQueryBuilder(),
		QueryParser:  NewQueryParser(),
	}
}

// NewQueryManagerWithFactory creates a new query manager using factory
func NewQueryManagerWithFactory(factory QueryBuilderFactory, dbType DBType) *QueryManager {
	return &QueryManager{
		SQLBuilder:   factory.CreateSQLQueryBuilder(dbType),
		MongoBuilder: factory.CreateMongoQueryBuilder(),
		QueryParser:  factory.CreateQueryParser(),
	}
}

// GetSQLBuilder returns the SQL query builder
func (qm *QueryManager) GetSQLBuilder() QueryBuilder {
	return qm.SQLBuilder
}

// GetMongoBuilder returns the MongoDB query builder
func (qm *QueryManager) GetMongoBuilder() MongoQueryBuilder {
	return qm.MongoBuilder
}

// GetQueryParser returns the query parser
func (qm *QueryManager) GetQueryParser() QueryParser {
	return qm.QueryParser
}

// ExecuteSQLQuery executes a SQL query with the SQL builder
func (qm *QueryManager) ExecuteSQLQuery(ctx context.Context, db *sqlx.DB, query DynamicQuery, dest interface{}) error {
	return qm.SQLBuilder.ExecuteQuery(ctx, db, query, dest)
}

// ExecuteMongoQuery executes a MongoDB query with the MongoDB builder
func (qm *QueryManager) ExecuteMongoQuery(ctx context.Context, collection *mongo.Collection, query DynamicQuery, dest interface{}) error {
	return qm.MongoBuilder.ExecuteFind(ctx, collection, query, dest)
}

// ParseQuery parses URL query parameters into a DynamicQuery
func (qm *QueryManager) ParseQuery(values interface{}, defaultTable string) (DynamicQuery, error) {
	return qm.QueryParser.ParseQuery(values, defaultTable)
}

// ParseQueryWithDefaultFields parses URL query parameters with default fields
func (qm *QueryManager) ParseQueryWithDefaultFields(values interface{}, defaultTable string, defaultFields []string) (DynamicQuery, error) {
	return qm.QueryParser.ParseQueryWithDefaultFields(values, defaultTable, defaultFields)
}

// Convenience functions for common operations

// CreateQueryBuilder creates a query builder with default settings
func CreateQueryBuilder(dbType DBType) QueryBuilder {
	return NewSQLQueryBuilder(dbType).
		SetSecurityOptions(true, 1000).
		SetQueryLogging(true).
		SetQueryTimeout(30)
}

// CreateMongoQueryBuilder creates a MongoDB query builder with default settings
func CreateMongoQueryBuilder() MongoQueryBuilder {
	return NewMongoQueryBuilder().
		SetSecurityOptions(true, 1000).
		SetQueryLogging(true).
		SetQueryTimeout(30)
}

// CreateQueryParser creates a query parser with default settings
func CreateQueryParser() QueryParser {
	return NewQueryParser().SetLimits(10, 100)
}

// ExecuteQuery is a convenience function for executing SQL queries
func ExecuteQuery(ctx context.Context, db *sqlx.DB, dbType DBType, query DynamicQuery, dest interface{}) error {
	builder := CreateQueryBuilder(dbType)
	return builder.ExecuteQuery(ctx, db, query, dest)
}

// ExecuteMongoQuery is a convenience function for executing MongoDB queries
func ExecuteMongoQuery(ctx context.Context, collection *mongo.Collection, query DynamicQuery, dest interface{}) error {
	builder := CreateMongoQueryBuilder()
	return builder.ExecuteFind(ctx, collection, query, dest)
}

// ParseURLQuery is a convenience function for parsing URL queries
func ParseURLQuery(urlValues interface{}, defaultTable string) (DynamicQuery, error) {
	parser := CreateQueryParser()
	return parser.ParseQuery(urlValues, defaultTable)
}

// ParseURLQueryWithFields is a convenience function for parsing URL queries with default fields
func ParseURLQueryWithFields(urlValues interface{}, defaultTable string, defaultFields []string) (DynamicQuery, error) {
	parser := CreateQueryParser()
	return parser.ParseQueryWithDefaultFields(urlValues, defaultTable, defaultFields)
}

// Database-specific helpers

// ForPostgreSQL creates a PostgreSQL-specific query builder
func ForPostgreSQL() QueryBuilder {
	return NewSQLQueryBuilder(DBTypePostgreSQL)
}

// ForMySQL creates a MySQL-specific query builder
func ForMySQL() QueryBuilder {
	return NewSQLQueryBuilder(DBTypeMySQL)
}

// ForSQLite creates a SQLite-specific query builder
func ForSQLite() QueryBuilder {
	return NewSQLQueryBuilder(DBTypeSQLite)
}

// ForSQLServer creates a SQL Server-specific query builder
func ForSQLServer() QueryBuilder {
	return NewSQLQueryBuilder(DBTypeSQLServer)
}

// ForMongoDB creates a MongoDB-specific query builder
func ForMongoDB() MongoQueryBuilder {
	return NewMongoQueryBuilder()
}

// Query execution helpers

// ExecuteCount executes a count query
func ExecuteCount(ctx context.Context, db *sqlx.DB, dbType DBType, query DynamicQuery) (int64, error) {
	builder := CreateQueryBuilder(dbType)
	return builder.ExecuteCount(ctx, db, query)
}

// ExecuteInsert executes an insert query
func ExecuteInsert(ctx context.Context, db *sqlx.DB, dbType DBType, table string, data InsertData, returningColumns ...string) (sql.Result, error) {
	builder := CreateQueryBuilder(dbType)
	return builder.ExecuteInsert(ctx, db, table, data, returningColumns...)
}

// ExecuteUpdate executes an update query
func ExecuteUpdate(ctx context.Context, db *sqlx.DB, dbType DBType, table string, updateData UpdateData, filters []FilterGroup, returningColumns ...string) (sql.Result, error) {
	builder := CreateQueryBuilder(dbType)
	return builder.ExecuteUpdate(ctx, db, table, updateData, filters, returningColumns...)
}

// ExecuteDelete executes a delete query
func ExecuteDelete(ctx context.Context, db *sqlx.DB, dbType DBType, table string, filters []FilterGroup, returningColumns ...string) (sql.Result, error) {
	builder := CreateQueryBuilder(dbType)
	return builder.ExecuteDelete(ctx, db, table, filters, returningColumns...)
}

// ExecuteUpsert executes an upsert query
func ExecuteUpsert(ctx context.Context, db *sqlx.DB, dbType DBType, table string, insertData InsertData, conflictColumns []string, updateColumns []string, returningColumns ...string) (sql.Result, error) {
	builder := CreateQueryBuilder(dbType)
	return builder.ExecuteUpsert(ctx, db, table, insertData, conflictColumns, updateColumns, returningColumns...)
}

// MongoDB execution helpers

// ExecuteMongoCount executes a MongoDB count query
func ExecuteMongoCount(ctx context.Context, collection *mongo.Collection, query DynamicQuery) (int64, error) {
	builder := CreateMongoQueryBuilder()
	return builder.ExecuteCount(ctx, collection, query)
}

// ExecuteMongoInsert executes a MongoDB insert query
func ExecuteMongoInsert(ctx context.Context, collection *mongo.Collection, data InsertData) (*mongo.InsertOneResult, error) {
	builder := CreateMongoQueryBuilder()
	return builder.ExecuteInsert(ctx, collection, data)
}

// ExecuteMongoUpdate executes a MongoDB update query
func ExecuteMongoUpdate(ctx context.Context, collection *mongo.Collection, updateData UpdateData, filters []FilterGroup) (*mongo.UpdateResult, error) {
	builder := CreateMongoQueryBuilder()
	return builder.ExecuteUpdate(ctx, collection, updateData, filters)
}

// ExecuteMongoDelete executes a MongoDB delete query
func ExecuteMongoDelete(ctx context.Context, collection *mongo.Collection, filters []FilterGroup) (*mongo.DeleteResult, error) {
	builder := CreateMongoQueryBuilder()
	return builder.ExecuteDelete(ctx, collection, filters)
}

// ExecuteMongoAggregate executes a MongoDB aggregation query
func ExecuteMongoAggregate(ctx context.Context, collection *mongo.Collection, query DynamicQuery, dest interface{}) error {
	builder := CreateMongoQueryBuilder()
	return builder.ExecuteAggregate(ctx, collection, query, dest)
}

// Utility functions for creating common queries

// CreateSimpleQuery creates a simple query with basic fields
func CreateSimpleQuery(table string, fields []string, limit int) DynamicQuery {
	selectFields := make([]SelectField, len(fields))
	for i, field := range fields {
		selectFields[i] = SelectField{Expression: field}
	}

	return DynamicQuery{
		From:   table,
		Fields: selectFields,
		Limit:  limit,
		Offset: 0,
	}
}

// CreateFilterQuery creates a query with filters
func CreateFilterQuery(table string, fields []string, filters []DynamicFilter, limit int) DynamicQuery {
	selectFields := make([]SelectField, len(fields))
	for i, field := range fields {
		selectFields[i] = SelectField{Expression: field}
	}

	return DynamicQuery{
		From:    table,
		Fields:  selectFields,
		Filters: []FilterGroup{{Filters: filters, LogicOp: "AND"}},
		Limit:   limit,
		Offset:  0,
	}
}

// CreateSortedQuery creates a query with sorting
func CreateSortedQuery(table string, fields []string, sorts []SortField, limit int) DynamicQuery {
	selectFields := make([]SelectField, len(fields))
	for i, field := range fields {
		selectFields[i] = SelectField{Expression: field}
	}

	return DynamicQuery{
		From:   table,
		Fields: selectFields,
		Sort:   sorts,
		Limit:  limit,
		Offset: 0,
	}
}

// CreatePaginatedQuery creates a paginated query
func CreatePaginatedQuery(table string, fields []string, filters []DynamicFilter, sorts []SortField, limit, offset int) DynamicQuery {
	selectFields := make([]SelectField, len(fields))
	for i, field := range fields {
		selectFields[i] = SelectField{Expression: field}
	}

	return DynamicQuery{
		From:    table,
		Fields:  selectFields,
		Filters: []FilterGroup{{Filters: filters, LogicOp: "AND"}},
		Sort:    sorts,
		Limit:   limit,
		Offset:  offset,
	}
}

// CreateInsertData creates insert data from a map
func CreateInsertData(data map[string]interface{}) InsertData {
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)
	}

	return InsertData{
		Columns: columns,
		Values:  values,
	}
}

// CreateUpdateData creates update data from a map
func CreateUpdateData(data map[string]interface{}) UpdateData {
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)
	}

	return UpdateData{
		Columns: columns,
		Values:  values,
	}
}

// CreateFilter creates a simple filter
func CreateFilter(column string, operator FilterOperator, value interface{}) DynamicFilter {
	return DynamicFilter{
		Column:   column,
		Operator: operator,
		Value:    value,
	}
}

// CreateEqualFilter creates an equality filter
func CreateEqualFilter(column string, value interface{}) DynamicFilter {
	return CreateFilter(column, OpEqual, value)
}

// CreateLikeFilter creates a LIKE filter
func CreateLikeFilter(column string, value string) DynamicFilter {
	return CreateFilter(column, OpLike, value)
}

// CreateInFilter creates an IN filter
func CreateInFilter(column string, values []interface{}) DynamicFilter {
	return CreateFilter(column, OpIn, values)
}

// CreateBetweenFilter creates a BETWEEN filter
func CreateBetweenFilter(column string, min, max interface{}) DynamicFilter {
	return CreateFilter(column, OpBetween, []interface{}{min, max})
}

// CreateSort creates a sort field
func CreateSort(column string, order string) SortField {
	return SortField{
		Column: column,
		Order:  order,
	}
}

// CreateAscSort creates an ascending sort
func CreateAscSort(column string) SortField {
	return CreateSort(column, "ASC")
}

// CreateDescSort creates a descending sort
func CreateDescSort(column string) SortField {
	return CreateSort(column, "DESC")
}

// CreateFilterGroup creates a filter group
func CreateFilterGroup(filters []DynamicFilter, logicOp string) FilterGroup {
	return FilterGroup{
		Filters: filters,
		LogicOp: logicOp,
	}
}

// CreateAndFilterGroup creates an AND filter group
func CreateAndFilterGroup(filters []DynamicFilter) FilterGroup {
	return CreateFilterGroup(filters, "AND")
}

// CreateOrFilterGroup creates an OR filter group
func CreateOrFilterGroup(filters []DynamicFilter) FilterGroup {
	return CreateFilterGroup(filters, "OR")
}
