package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// SQLQueryBuilder implements QueryBuilder interface for SQL databases
type SQLQueryBuilder struct {
	dbType         DBType
	dialect        DatabaseDialect
	sqlBuilder     squirrel.StatementBuilderType
	allowedColumns map[string]bool // Security: only allow specified columns
	allowedTables  map[string]bool // Security: only allow specified tables
	// Security settings
	enableSecurityChecks bool
	maxAllowedRows       int
	// Query logging
	enableQueryLogging bool
	logger             Logger
	// Connection timeout settings
	queryTimeout time.Duration
}

// Logger interface for query logging
type Logger interface {
	Debugf(format string, args ...interface{})
}

// NewSQLQueryBuilder creates a new SQL query builder instance for a specific database type
func NewSQLQueryBuilder(dbType DBType) *SQLQueryBuilder {
	var placeholderFormat squirrel.PlaceholderFormat

	switch dbType {
	case DBTypePostgreSQL:
		placeholderFormat = squirrel.Dollar
	case DBTypeMySQL, DBTypeSQLite:
		placeholderFormat = squirrel.Question
	case DBTypeSQLServer:
		placeholderFormat = squirrel.AtP
	default:
		placeholderFormat = squirrel.Question
	}

	dialect := GetDialect(dbType)

	return &SQLQueryBuilder{
		dbType:               dbType,
		dialect:              dialect,
		sqlBuilder:           squirrel.StatementBuilder.PlaceholderFormat(placeholderFormat),
		allowedColumns:       make(map[string]bool),
		allowedTables:        make(map[string]bool),
		enableSecurityChecks: true,
		maxAllowedRows:       10000,
		enableQueryLogging:   true,
		queryTimeout:         30 * time.Second,
	}
}

// SetLogger sets the logger for query logging
func (qb *SQLQueryBuilder) SetLogger(logger Logger) QueryBuilder {
	qb.logger = logger
	return qb
}

// SetSecurityOptions configures security settings
func (qb *SQLQueryBuilder) SetSecurityOptions(enableChecks bool, maxRows int) QueryBuilder {
	qb.enableSecurityChecks = enableChecks
	qb.maxAllowedRows = maxRows
	return qb
}

// SetAllowedColumns sets the list of allowed columns for security
func (qb *SQLQueryBuilder) SetAllowedColumns(columns []string) QueryBuilder {
	qb.allowedColumns = make(map[string]bool)
	for _, col := range columns {
		qb.allowedColumns[col] = true
	}
	return qb
}

// SetAllowedTables sets the list of allowed tables for security
func (qb *SQLQueryBuilder) SetAllowedTables(tables []string) QueryBuilder {
	qb.allowedTables = make(map[string]bool)
	for _, table := range tables {
		qb.allowedTables[table] = true
	}
	return qb
}

// SetQueryLogging enables or disables query logging
func (qb *SQLQueryBuilder) SetQueryLogging(enable bool) QueryBuilder {
	qb.enableQueryLogging = enable
	return qb
}

// SetQueryTimeout sets the default query timeout
func (qb *SQLQueryBuilder) SetQueryTimeout(timeout time.Duration) QueryBuilder {
	qb.queryTimeout = timeout
	return qb
}

// logDebug logs debug messages if logging is enabled
func (qb *SQLQueryBuilder) logDebug(format string, args ...interface{}) {
	if qb.enableQueryLogging {
		if qb.logger != nil {
			qb.logger.Debugf(format, args...)
		} else {
			// Fallback to fmt.Printf if no logger is set
			fmt.Printf("[DEBUG] "+format+"\n", args...)
		}
	}
}

// BuildQuery builds the complete SQL SELECT query with support for CTEs, JOINs, and UNIONs
func (qb *SQLQueryBuilder) BuildQuery(query DynamicQuery) (string, []interface{}, error) {
	var allArgs []interface{}
	var queryParts []string

	// --- Langkah 1: Validasi ---
	if qb.enableSecurityChecks && query.Limit > qb.maxAllowedRows {
		return "", nil, fmt.Errorf("requested limit %d exceeds maximum allowed %d", query.Limit, qb.maxAllowedRows)
	}
	if qb.enableSecurityChecks && len(qb.allowedTables) > 0 && !qb.allowedTables[query.From] {
		return "", nil, fmt.Errorf("disallowed table: %s", query.From)
	}

	// --- Langkah 2: Bangun CTEs (WITH clause) ---
	if len(query.CTEs) > 0 {
		cteClause, cteArgs, err := qb.buildCTEClause(query.CTEs)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build CTE clause: %w", err)
		}
		queryParts = append(queryParts, cteClause)
		allArgs = append(allArgs, cteArgs...)
	}

	// --- Langkah 3: Bangun Query Utama ---
	// Ini adalah satu-satunya tempat kita memanggil buildMainQuery
	mainQuery, err := qb.buildMainQuery(query)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build main query: %w", err)
	}

	// Konversi mainQuery (squirrel.SelectBuilder) ke string dan args
	sql, args, err := mainQuery.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert main query to SQL: %w", err)
	}
	queryParts = append(queryParts, sql)
	allArgs = append(allArgs, args...)

	// --- Langkah 4: Terapkan UNIONs ---
	if len(query.Unions) > 0 {
		unionClause, unionArgs, err := qb.buildUnionClause(query.Unions)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build UNION clause: %w", err)
		}
		queryParts = append(queryParts, unionClause)
		allArgs = append(allArgs, unionArgs...)
	}

	// --- Langkah 5: Gabungkan dan Validasi ---
	finalSQL := strings.Join(queryParts, " ")

	// Validasi keamanan akhir menggunakan SQL parser
	if qb.enableSecurityChecks {
		if err := qb.validateParsedSQL(finalSQL); err != nil {
			return "", nil, fmt.Errorf("query validation failed: %w", err)
		}
	}

	qb.logDebug("Final SQL query: %s", finalSQL)
	qb.logDebug("Final query args: %v", allArgs)

	return finalSQL, allArgs, nil
}

// buildMainQuery builds the main SELECT query
func (qb *SQLQueryBuilder) buildMainQuery(query DynamicQuery) (squirrel.SelectBuilder, error) {
	fromClause := qb.buildFromClause(query.From, query.Aliases)
	selectFields := qb.buildSelectFields(query.Fields)

	// Start building the main query
	mainQuery := qb.sqlBuilder.Select(selectFields...).From(fromClause)

	// Add JOINs
	if len(query.Joins) > 0 {
		mainQuery = qb.applyJoins(mainQuery, query.Joins)
	}

	// Apply WHERE conditions - HANYA SEKALI
	if len(query.Filters) > 0 {
		// Panggil BuildWhereClause hanya satu kali
		whereClause, whereArgs, err := qb.BuildWhereClause(query.Filters)
		if err != nil {
			return squirrel.SelectBuilder{}, fmt.Errorf("failed to build WHERE clause: %w", err)
		}
		// Terapkan hasilnya ke builder Squirrel hanya satu kali
		mainQuery = mainQuery.Where(whereClause, whereArgs...)
	}

	// Apply GROUP BY
	if len(query.GroupBy) > 0 {
		mainQuery = mainQuery.GroupBy(qb.buildGroupByColumns(query.GroupBy)...)
	}

	// Apply HAVING conditions
	if len(query.Having) > 0 {
		havingClause, havingArgs, err := qb.BuildWhereClause(query.Having)
		if err != nil {
			return squirrel.SelectBuilder{}, fmt.Errorf("failed to build HAVING clause: %w", err)
		}
		mainQuery = mainQuery.Having(havingClause, havingArgs...)
	}

	// Apply ORDER BY
	if len(query.Sort) > 0 {
		for _, sort := range query.Sort {
			column := qb.validateAndEscapeColumn(sort.Column)
			if column == "" {
				continue
			}
			order := "ASC"
			if strings.ToUpper(sort.Order) == "DESC" {
				order = "DESC"
			}
			mainQuery = mainQuery.OrderBy(fmt.Sprintf("%s %s", column, order))
		}
	}

	// Handle window functions and JSON operations if present
	if len(query.WindowFunctions) > 0 || len(query.JsonOperations) > 0 {
		mainQuery = qb.applyAdvancedFeatures(mainQuery, query, selectFields, fromClause)
	}

	// Apply pagination with dialect-specific syntax
	if query.Limit > 0 {
		if qb.dbType == DBTypeSQLServer {
			if len(query.Sort) == 0 {
				mainQuery = mainQuery.OrderBy("(SELECT 1)")
			}
			mainQuery = mainQuery.Suffix(fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", query.Offset, query.Limit))
		} else {
			mainQuery = mainQuery.Limit(uint64(query.Limit))
			if query.Offset > 0 {
				mainQuery = mainQuery.Offset(uint64(query.Offset))
			}
		}
	} else if query.Offset > 0 && qb.dbType != DBTypeSQLServer {
		mainQuery = mainQuery.Offset(uint64(query.Offset))
	}

	return mainQuery, nil
}

// applyJoins applies JOIN clauses to the query
func (qb *SQLQueryBuilder) applyJoins(mainQuery squirrel.SelectBuilder, joins []Join) squirrel.SelectBuilder {
	for _, join := range joins {
		// Security check for joined table
		if qb.enableSecurityChecks && len(qb.allowedTables) > 0 && !qb.allowedTables[join.Table] {
			// This should be handled at a higher level, but we'll log it
			qb.logDebug("Warning: disallowed table in join: %s", join.Table)
			continue
		}

		joinType, tableWithAlias, onClause, joinArgs, err := qb.buildSingleJoinClause(join)
		if err != nil {
			qb.logDebug("Warning: failed to build join clause for table %s: %v", join.Table, err)
			continue
		}

		joinStr := tableWithAlias + " ON " + onClause
		switch strings.ToUpper(joinType) {
		case "LEFT":
			if join.Lateral {
				mainQuery = mainQuery.LeftJoin("LATERAL "+joinStr, joinArgs...)
			} else {
				mainQuery = mainQuery.LeftJoin(joinStr, joinArgs...)
			}
		case "RIGHT":
			mainQuery = mainQuery.RightJoin(joinStr, joinArgs...)
		case "FULL":
			mainQuery = mainQuery.Join("FULL JOIN "+joinStr, joinArgs...)
		default:
			if join.Lateral {
				mainQuery = mainQuery.Join("LATERAL "+joinStr, joinArgs...)
			} else {
				mainQuery = mainQuery.Join(joinStr, joinArgs...)
			}
		}
	}
	return mainQuery
}

// applyAdvancedFeatures applies window functions and JSON operations
func (qb *SQLQueryBuilder) applyAdvancedFeatures(mainQuery squirrel.SelectBuilder, query DynamicQuery, selectFields []string, fromClause string) squirrel.SelectBuilder {
	// Rebuild the SELECT clause with window functions and JSON operations
	var finalSelectFields []string
	finalSelectFields = append(finalSelectFields, selectFields...)

	// Add window functions
	for _, wf := range query.WindowFunctions {
		windowFunc, err := qb.buildWindowFunction(wf)
		if err != nil {
			qb.logDebug("Warning: failed to build window function: %v", err)
			continue
		}
		finalSelectFields = append(finalSelectFields, windowFunc)
	}

	// Add JSON operations
	for _, jo := range query.JsonOperations {
		jsonExpr, jsonArgs, err := qb.buildJsonOperation(jo)
		if err != nil {
			qb.logDebug("Warning: failed to build JSON operation: %v", err)
			continue
		}
		if jo.Alias != "" {
			jsonExpr += " AS " + qb.dialect.EscapeIdentifier(jo.Alias)
		}
		finalSelectFields = append(finalSelectFields, jsonExpr)
		// Note: JSON args would need to be handled differently in this context
		_ = jsonArgs // Suppress unused warning
	}

	// Rebuild the query with the complete SELECT clause
	return qb.sqlBuilder.Select(finalSelectFields...).From(fromClause)
}

// buildCTEClause builds the WITH clause for CTEs
func (qb *SQLQueryBuilder) buildCTEClause(ctes []CTE) (string, []interface{}, error) {
	var cteParts []string
	var allArgs []interface{}

	for i, cte := range ctes {
		// Security check for CTE table
		if qb.enableSecurityChecks && len(qb.allowedTables) > 0 && !qb.allowedTables[cte.Query.From] {
			return "", nil, fmt.Errorf("disallowed table in CTE: %s", cte.Query.From)
		}

		cteSQL, cteArgs, err := qb.BuildQuery(cte.Query)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build CTE '%s': %w", cte.Name, err)
		}

		ctePart := fmt.Sprintf("%s AS (%s)", qb.dialect.EscapeIdentifier(cte.Name), cteSQL)
		cteParts = append(cteParts, ctePart)
		allArgs = append(allArgs, cteArgs...)

		if i < len(ctes)-1 {
			cteParts = append(cteParts, ",")
		}
	}

	return "WITH " + strings.Join(cteParts, " "), allArgs, nil
}

// buildUnionClause builds UNION clauses
func (qb *SQLQueryBuilder) buildUnionClause(unions []Union) (string, []interface{}, error) {
	var unionParts []string
	var allArgs []interface{}

	for _, union := range unions {
		// Security check for union table
		if qb.enableSecurityChecks && len(qb.allowedTables) > 0 && !qb.allowedTables[union.Query.From] {
			return "", nil, fmt.Errorf("disallowed table in union: %s", union.Query.From)
		}

		unionSQL, unionArgs, err := qb.BuildQuery(union.Query)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build union query: %w", err)
		}

		unionType := "UNION"
		if union.Type != "" {
			unionType = strings.ToUpper(union.Type)
		}

		unionParts = append(unionParts, unionType, unionSQL)
		allArgs = append(allArgs, unionArgs...)
	}

	return strings.Join(unionParts, " "), allArgs, nil
}

// buildFromClause builds the FROM clause
func (qb *SQLQueryBuilder) buildFromClause(table, alias string) string {
	fromClause := qb.dialect.EscapeIdentifier(table)
	if alias != "" {
		fromClause += " AS " + qb.dialect.EscapeIdentifier(alias)
	}
	return fromClause
}

// buildSelectFields builds the SELECT fields
func (qb *SQLQueryBuilder) buildSelectFields(fields []SelectField) []string {
	if len(fields) == 0 {
		return []string{"*"}
	}

	var selectFields []string
	for _, field := range fields {
		if field.Expression == "" {
			continue
		}

		selectExpr := field.Expression
		if field.Alias != "" {
			selectExpr += " AS " + qb.dialect.EscapeIdentifier(field.Alias)
		}
		selectFields = append(selectFields, selectExpr)
	}

	if len(selectFields) == 0 {
		return []string{"*"}
	}

	return selectFields
}

// buildGroupByColumns builds GROUP BY columns
func (qb *SQLQueryBuilder) buildGroupByColumns(columns []string) []string {
	var groupByColumns []string
	for _, col := range columns {
		escapedCol := qb.validateAndEscapeColumn(col)
		if escapedCol != "" {
			groupByColumns = append(groupByColumns, escapedCol)
		}
	}
	return groupByColumns
}

// buildSingleJoinClause builds a single JOIN clause
func (qb *SQLQueryBuilder) buildSingleJoinClause(join Join) (string, string, string, []interface{}, error) {
	joinType := "INNER"
	if join.Type != "" {
		joinType = strings.ToUpper(join.Type)
	}

	tableWithAlias := qb.dialect.EscapeIdentifier(join.Table)
	if join.Alias != "" {
		tableWithAlias += " AS " + qb.dialect.EscapeIdentifier(join.Alias)
	}

	onClause, onArgs, err := qb.BuildWhereClause([]FilterGroup{join.OnConditions})
	if err != nil {
		return "", "", "", nil, err
	}

	return joinType, tableWithAlias, onClause, onArgs, nil
}

// buildWindowFunction builds a window function expression
func (qb *SQLQueryBuilder) buildWindowFunction(wf WindowFunction) (string, error) {
	var funcExpr string

	switch strings.ToUpper(wf.Function) {
	case "ROW_NUMBER":
		funcExpr = "ROW_NUMBER()"
	case "RANK":
		funcExpr = "RANK()"
	case "DENSE_RANK":
		funcExpr = "DENSE_RANK()"
	case "LEAD":
		funcExpr = "LEAD()"
	case "LAG":
		funcExpr = "LAG()"
	default:
		return "", fmt.Errorf("unsupported window function: %s", wf.Function)
	}

	overClause := "OVER ("
	if wf.Over != "" {
		overClause += wf.Over
	}
	if wf.OrderBy != "" {
		if wf.Over != "" {
			overClause += " "
		}
		overClause += "ORDER BY " + wf.OrderBy
	}
	if wf.Frame != "" {
		if wf.Over != "" || wf.OrderBy != "" {
			overClause += " "
		}
		overClause += wf.Frame
	}
	overClause += ")"

	windowExpr := funcExpr + " " + overClause
	if wf.Alias != "" {
		windowExpr += " AS " + qb.dialect.EscapeIdentifier(wf.Alias)
	}

	return windowExpr, nil
}

// buildJsonOperation builds a JSON operation expression
func (qb *SQLQueryBuilder) buildJsonOperation(jo JsonOperation) (string, []interface{}, error) {
	var expr string
	var args []interface{}

	switch jo.Type {
	case "extract":
		switch qb.dbType {
		case DBTypePostgreSQL:
			expr = fmt.Sprintf("%s->'%s'", qb.dialect.EscapeIdentifier(jo.Column), jo.Path)
		case DBTypeMySQL:
			expr = fmt.Sprintf("JSON_EXTRACT(%s, '$.%s')", qb.dialect.EscapeIdentifier(jo.Column), jo.Path)
		case DBTypeSQLServer:
			expr = fmt.Sprintf("JSON_VALUE(%s, '$.%s')", qb.dialect.EscapeIdentifier(jo.Column), jo.Path)
		case DBTypeSQLite:
			expr = fmt.Sprintf("json_extract(%s, '$.%s')", qb.dialect.EscapeIdentifier(jo.Column), jo.Path)
		default:
			return "", nil, fmt.Errorf("JSON operations not supported for database type: %s", qb.dbType)
		}
	case "exists":
		switch qb.dbType {
		case DBTypePostgreSQL:
			expr = fmt.Sprintf("jsonb_path_exists(%s, '$.%s')", qb.dialect.EscapeIdentifier(jo.Column), jo.Path)
		case DBTypeMySQL:
			expr = fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '$.%s')", qb.dialect.EscapeIdentifier(jo.Column), jo.Path)
		case DBTypeSQLServer:
			expr = fmt.Sprintf("JSON_VALUE(%s, '$.%s') IS NOT NULL", qb.dialect.EscapeIdentifier(jo.Column), jo.Path)
		case DBTypeSQLite:
			expr = fmt.Sprintf("json_extract(%s, '$.%s') IS NOT NULL", qb.dialect.EscapeIdentifier(jo.Column), jo.Path)
		default:
			return "", nil, fmt.Errorf("JSON operations not supported for database type: %s", qb.dbType)
		}
	default:
		return "", nil, fmt.Errorf("unsupported JSON operation: %s", jo.Type)
	}
	if jo.Alias != "" {
		expr += " AS " + qb.dialect.EscapeIdentifier(jo.Alias)
	}
	return expr, args, nil
}

// BuildCountQuery builds a count query
func (qb *SQLQueryBuilder) BuildCountQuery(query DynamicQuery) (string, []interface{}, error) {
	// For a count query, we don't need fields, joins, or unions.
	// We only need FROM, WHERE, GROUP BY, HAVING.
	countQuery := DynamicQuery{
		From:    query.From,
		Aliases: query.Aliases,
		Filters: query.Filters,
		GroupBy: query.GroupBy,
		Having:  query.Having,
		// Joins are important for count with filters on joined tables
		Joins: query.Joins,
	}

	// Build the base query for the count using Squirrel's From and Join methods
	fromClause := qb.buildFromClause(countQuery.From, countQuery.Aliases)
	baseQuery := qb.sqlBuilder.Select("COUNT(*)").From(fromClause)

	// Add JOINs using Squirrel's Join method
	if len(countQuery.Joins) > 0 {
		for _, join := range countQuery.Joins {
			// Security check for joined table
			if qb.enableSecurityChecks && len(qb.allowedTables) > 0 && !qb.allowedTables[join.Table] {
				return "", nil, fmt.Errorf("disallowed table in join: %s", join.Table)
			}

			joinType, tableWithAlias, onClause, joinArgs, err := qb.buildSingleJoinClause(join)
			if err != nil {
				return "", nil, err
			}
			joinStr := tableWithAlias + " ON " + onClause
			switch strings.ToUpper(joinType) {
			case "LEFT":
				baseQuery = baseQuery.LeftJoin(joinStr, joinArgs...)
			case "RIGHT":
				baseQuery = baseQuery.RightJoin(joinStr, joinArgs...)
			case "FULL":
				baseQuery = baseQuery.Join("FULL JOIN "+joinStr, joinArgs...)
			default:
				baseQuery = baseQuery.Join(joinStr, joinArgs...)
			}
		}
	}

	if len(countQuery.Filters) > 0 {
		whereClause, whereArgs, err := qb.BuildWhereClause(countQuery.Filters)
		if err != nil {
			return "", nil, err
		}
		baseQuery = baseQuery.Where(whereClause, whereArgs...)
	}

	if len(countQuery.GroupBy) > 0 {
		baseQuery = baseQuery.GroupBy(qb.buildGroupByColumns(countQuery.GroupBy)...)
	}

	if len(countQuery.Having) > 0 {
		havingClause, havingArgs, err := qb.BuildWhereClause(countQuery.Having)
		if err != nil {
			return "", nil, err
		}
		baseQuery = baseQuery.Having(havingClause, havingArgs...)
	}

	sql, args, err := baseQuery.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build COUNT query: %w", err)
	}

	if qb.enableQueryLogging {
		fmt.Printf("[DEBUG] COUNT SQL query: %s\n", sql)
		fmt.Printf("[DEBUG] COUNT query args: %v\n", args)
	}
	return sql, args, nil
}

// BuildInsertQuery builds an INSERT query
func (qb *SQLQueryBuilder) BuildInsertQuery(table string, data InsertData, returningColumns ...string) (string, []interface{}, error) {
	// Validate columns
	for _, col := range data.Columns {
		if qb.allowedColumns != nil && !qb.allowedColumns[col] {
			return "", nil, fmt.Errorf("disallowed column: %s", col)
		}
	}

	// Escape table name for case sensitivity
	escapedTable := qb.dialect.EscapeIdentifier(table)

	// Escape column names for case sensitivity
	escapedColumns := make([]string, len(data.Columns))
	for i, col := range data.Columns {
		escapedColumns[i] = qb.dialect.EscapeIdentifier(col)
	}

	// Start with basic insert
	insert := qb.sqlBuilder.Insert(escapedTable).Columns(escapedColumns...).Values(data.Values...)

	// Handle JSON values - we need to modify the insert statement
	if len(data.JsonValues) > 0 {
		// Create a new insert builder with all columns including JSON columns
		allColumns := make([]string, len(escapedColumns))
		copy(allColumns, escapedColumns)

		allValues := make([]interface{}, len(data.Values))
		copy(allValues, data.Values)

		for col, val := range data.JsonValues {
			escapedJsonCol := qb.dialect.EscapeIdentifier(col)
			allColumns = append(allColumns, escapedJsonCol)
			jsonVal, err := json.Marshal(val)
			if err != nil {
				return "", nil, fmt.Errorf("failed to marshal JSON value for column %s: %w", col, err)
			}
			allValues = append(allValues, jsonVal)
		}

		insert = qb.sqlBuilder.Insert(escapedTable).Columns(allColumns...).Values(allValues...)
	}

	if len(returningColumns) > 0 {
		if qb.dbType == DBTypePostgreSQL {
			// Escape returning columns
			escapedReturningColumns := make([]string, len(returningColumns))
			for i, col := range returningColumns {
				escapedReturningColumns[i] = qb.dialect.EscapeIdentifier(col)
			}
			insert = insert.Suffix("RETURNING " + strings.Join(escapedReturningColumns, ", "))
		} else {
			return "", nil, fmt.Errorf("RETURNING not supported for database type: %s", qb.dbType)
		}
	}

	sql, args, err := insert.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build INSERT query: %w", err)
	}

	return sql, args, nil
}

// BuildUpdateQuery builds an UPDATE query
func (qb *SQLQueryBuilder) BuildUpdateQuery(table string, updateData UpdateData, filters []FilterGroup, returningColumns ...string) (string, []interface{}, error) {
	// Validate columns
	for _, col := range updateData.Columns {
		if qb.allowedColumns != nil && !qb.allowedColumns[col] {
			return "", nil, fmt.Errorf("disallowed column: %s", col)
		}
	}

	// Escape table name for case sensitivity
	escapedTable := qb.dialect.EscapeIdentifier(table)

	// Create set map with proper escaping
	setMap := make(map[string]interface{})
	for i, col := range updateData.Columns {
		escapedCol := qb.dialect.EscapeIdentifier(col)
		setMap[escapedCol] = updateData.Values[i]
	}

	// Start with basic update
	update := qb.sqlBuilder.Update(escapedTable).SetMap(setMap)

	// Handle JSON updates if any
	if len(updateData.JsonUpdates) > 0 {
		for col, jsonUpdate := range updateData.JsonUpdates {
			escapedJsonCol := qb.dialect.EscapeIdentifier(col)
			switch qb.dbType {
			case DBTypePostgreSQL:
				jsonVal, err := json.Marshal(jsonUpdate.Value)
				if err != nil {
					return "", nil, fmt.Errorf("failed to marshal JSON value for column %s: %w", col, err)
				}
				setMap[escapedJsonCol] = squirrel.Expr(fmt.Sprintf("jsonb_set(%s, '%s', ?)", escapedJsonCol, jsonUpdate.Path), jsonVal)
			case DBTypeMySQL:
				jsonVal, err := json.Marshal(jsonUpdate.Value)
				if err != nil {
					return "", nil, fmt.Errorf("failed to marshal JSON value for column %s: %w", col, err)
				}
				setMap[escapedJsonCol] = squirrel.Expr(fmt.Sprintf("JSON_SET(%s, '%s', ?)", escapedJsonCol, jsonUpdate.Path), jsonVal)
			case DBTypeSQLServer:
				jsonVal, err := json.Marshal(jsonUpdate.Value)
				if err != nil {
					return "", nil, fmt.Errorf("failed to marshal JSON value for column %s: %w", col, err)
				}
				setMap[escapedJsonCol] = squirrel.Expr(fmt.Sprintf("JSON_MODIFY(%s, '%s', ?)", escapedJsonCol, jsonUpdate.Path), jsonVal)
			case DBTypeSQLite:
				jsonVal, err := json.Marshal(jsonUpdate.Value)
				if err != nil {
					return "", nil, fmt.Errorf("failed to marshal JSON value for column %s: %w", col, err)
				}
				setMap[escapedJsonCol] = squirrel.Expr(fmt.Sprintf("json_patch(%s, ?)", escapedJsonCol), jsonVal)
			}
		}
		update = qb.sqlBuilder.Update(escapedTable).SetMap(setMap)
	}

	if len(filters) > 0 {
		whereClause, whereArgs, err := qb.BuildWhereClause(filters)
		if err != nil {
			return "", nil, err
		}
		update = update.Where(whereClause, whereArgs...)
	}

	if len(returningColumns) > 0 {
		if qb.dbType == DBTypePostgreSQL {
			escapedReturningColumns := make([]string, len(returningColumns))
			for i, col := range returningColumns {
				escapedReturningColumns[i] = qb.dialect.EscapeIdentifier(col)
			}
			update = update.Suffix("RETURNING " + strings.Join(escapedReturningColumns, ", "))
		} else {
			return "", nil, fmt.Errorf("RETURNING not supported for database type: %s", qb.dbType)
		}
	}

	sql, args, err := update.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build UPDATE query: %w", err)
	}

	return sql, args, nil
}

// BuildDeleteQuery builds a DELETE query
func (qb *SQLQueryBuilder) BuildDeleteQuery(table string, filters []FilterGroup, returningColumns ...string) (string, []interface{}, error) {
	// Escape table name for case sensitivity
	escapedTable := qb.dialect.EscapeIdentifier(table)

	delete := qb.sqlBuilder.Delete(escapedTable)

	if len(filters) > 0 {
		whereClause, whereArgs, err := qb.BuildWhereClause(filters)
		if err != nil {
			return "", nil, err
		}
		delete = delete.Where(whereClause, whereArgs...)
	}

	if len(returningColumns) > 0 {
		if qb.dbType == DBTypePostgreSQL {
			// Escape returning columns
			escapedReturningColumns := make([]string, len(returningColumns))
			for i, col := range returningColumns {
				escapedReturningColumns[i] = qb.dialect.EscapeIdentifier(col)
			}
			delete = delete.Suffix("RETURNING " + strings.Join(escapedReturningColumns, ", "))
		} else {
			return "", nil, fmt.Errorf("RETURNING not supported for database type: %s", qb.dbType)
		}
	}

	sql, args, err := delete.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build DELETE query: %w", err)
	}

	return sql, args, nil
}

// BuildUpsertQuery builds an UPSERT query
func (qb *SQLQueryBuilder) BuildUpsertQuery(table string, insertData InsertData, conflictColumns []string, updateColumns []string, returningColumns ...string) (string, []interface{}, error) {
	// Validate columns
	for _, col := range insertData.Columns {
		if qb.allowedColumns != nil && !qb.allowedColumns[col] {
			return "", nil, fmt.Errorf("disallowed column: %s", col)
		}
	}
	for _, col := range updateColumns {
		if qb.allowedColumns != nil && !qb.allowedColumns[col] {
			return "", nil, fmt.Errorf("disallowed column: %s", col)
		}
	}

	// Escape table name for case sensitivity
	escapedTable := qb.dialect.EscapeIdentifier(table)

	switch qb.dbType {
	case DBTypePostgreSQL:
		// Handle JSON values for PostgreSQL
		allColumns := make([]string, len(insertData.Columns))
		copy(allColumns, insertData.Columns)

		allValues := make([]interface{}, len(insertData.Values))
		copy(allValues, insertData.Values)

		for col, val := range insertData.JsonValues {
			allColumns = append(allColumns, col)
			jsonVal, err := json.Marshal(val)
			if err != nil {
				return "", nil, fmt.Errorf("failed to marshal JSON value for column %s: %w", col, err)
			}
			allValues = append(allValues, jsonVal)
		}

		insert := qb.sqlBuilder.Insert(escapedTable).Columns(allColumns...).Values(allValues...)
		if len(conflictColumns) > 0 {
			conflictTarget := strings.Join(conflictColumns, ", ")
			setClause := ""
			for _, col := range updateColumns {
				if setClause != "" {
					setClause += ", "
				}
				setClause += fmt.Sprintf("%s = EXCLUDED.%s", qb.dialect.EscapeIdentifier(col), qb.dialect.EscapeIdentifier(col))
			}
			insert = insert.Suffix(fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", conflictTarget, setClause))
		}
		if len(returningColumns) > 0 {
			insert = insert.Suffix("RETURNING " + strings.Join(returningColumns, ", "))
		}
		sql, args, err := insert.ToSql()
		if err != nil {
			return "", nil, fmt.Errorf("failed to build UPSERT query: %w", err)
		}
		return sql, args, nil
	case DBTypeMySQL:
		// Handle JSON values for MySQL
		allColumns := make([]string, len(insertData.Columns))
		copy(allColumns, insertData.Columns)

		allValues := make([]interface{}, len(insertData.Values))
		copy(allValues, insertData.Values)

		for col, val := range insertData.JsonValues {
			allColumns = append(allColumns, col)
			jsonVal, err := json.Marshal(val)
			if err != nil {
				return "", nil, fmt.Errorf("failed to marshal JSON value for column %s: %w", col, err)
			}
			allValues = append(allValues, jsonVal)
		}

		insert := qb.sqlBuilder.Insert(table).Columns(allColumns...).Values(allValues...)
		if len(updateColumns) > 0 {
			setClause := ""
			for _, col := range updateColumns {
				if setClause != "" {
					setClause += ", "
				}
				setClause += fmt.Sprintf("%s = VALUES(%s)", qb.dialect.EscapeIdentifier(col), qb.dialect.EscapeIdentifier(col))
			}
			insert = insert.Suffix(fmt.Sprintf("ON DUPLICATE KEY UPDATE %s", setClause))
		}
		sql, args, err := insert.ToSql()
		if err != nil {
			return "", nil, fmt.Errorf("failed to build UPSERT query: %w", err)
		}
		return sql, args, nil
	default:
		return "", nil, fmt.Errorf("UPSERT not supported for database type: %s", qb.dbType)
	}
}

// BuildWhereClause builds WHERE/HAVING conditions from FilterGroups
func (qb *SQLQueryBuilder) BuildWhereClause(filterGroups []FilterGroup) (string, []interface{}, error) {
	if len(filterGroups) == 0 {
		// Jika tidak ada filter, kembalikan string kosong dan args kosong.
		return "", nil, nil
	}

	// Slice untuk menampung semua string kondisi dari setiap group.
	var groupConditions []string
	// Slice untuk menampung semua argumen dari semua filter.
	var allArgs []interface{}

	// Iterasi setiap FilterGroup yang diberikan.
	// Contoh filterGroups: [ {Filters: [filter1, filter2], LogicOp: "AND"}, {Filters: [filter3], LogicOp: "OR"} ]
	for i, group := range filterGroups {
		if len(group.Filters) == 0 {
			// Lewati group yang tidak memiliki filter sama sekali.
			qb.logDebug("Skipping empty filter group at index %d", i)
			continue
		}

		// Bangun string kondisi untuk group saat ini (misal: "col1 = ? AND col2 LIKE ?")
		// dan kumpulkan argumennya.
		groupCondition, groupArgs, err := qb.buildFilterGroup(group)
		if err != nil {
			// Jika terjadi error saat membangun group, hentikan proses dan kembalikan error.
			return "", nil, fmt.Errorf("failed to build filter group at index %d: %w", i, err)
		}

		// Tambahkan kondisi group yang sudah dibuat ke daftar.
		// Kita bungkus dengan parentheses untuk memastikan urutan operasi yang benar.
		groupConditions = append(groupConditions, fmt.Sprintf("(%s)", groupCondition))
		// Tambahkan argumen dari group ini ke daftar argumen utama.
		allArgs = append(allArgs, groupArgs...)
	}

	// Gabungkan semua kondisi group menjadi satu string WHERE.
	// Kita gunakan "AND" sebagai operator default antar group utama.
	// Contoh hasil: ("id" = $1 AND "name" LIKE $2) OR ("status" = $3)
	whereClause := strings.Join(groupConditions, " AND ")

	qb.logDebug("Final WHERE clause built: %s", whereClause)
	qb.logDebug("Final WHERE args: %v", allArgs)

	return whereClause, allArgs, nil
}

// buildFilterGroup membangun satu string kondisi dari sebuah FilterGroup.
// Ini adalah pembantu untuk BuildWhereClause.
func (qb *SQLQueryBuilder) buildFilterGroup(group FilterGroup) (string, []interface{}, error) {
	var conditions []string
	var args []interface{}

	// Tentukan operator logika untuk group ini. Defaultnya adalah "AND".
	logicOp := "AND"
	if group.LogicOp != "" {
		logicOp = strings.ToUpper(group.LogicOp)
	}

	// Iterasi setiap DynamicFilter dalam group.
	// Contoh group.Filters: [ {Column: "id", Operator: "_eq", Value: 5}, {Column: "name", Operator: "_like", Value: "test"} ]
	for i, filter := range group.Filters {
		// Bangun kondisi individual untuk setiap filter (misal: "id" = $1).
		condition, filterArgs, err := qb.buildFilterCondition(filter)
		if err != nil {
			// Jika gagal membangun kondisi filter, hentikan proses.
			return "", nil, fmt.Errorf("failed to build filter condition at index %d: %w", i, err)
		}

		// Jika ini bukan filter pertama, tambahkan operator logika sebelum kondisi.
		if len(conditions) > 0 {
			conditions = append(conditions, logicOp)
		}
		// Tambahkan string kondisi.
		conditions = append(conditions, condition)
		// Tambahkan argumen.
		args = append(args, filterArgs...)
	}

	// Gabungkan semua kondisi dalam group dengan spasi.
	// Contoh hasil: "id" = $1 AND "name" LIKE $2
	groupClause := strings.Join(conditions, " ")
	return groupClause, args, nil
}

// ExecuteQuery executes a query with parameters and returns rows
func (qb *SQLQueryBuilder) ExecuteQuery(ctx context.Context, exec sqlx.ExtContext, query DynamicQuery, dest interface{}) error {
	sql, args, err := qb.BuildQuery(query)
	if err != nil {
		return err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && qb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qb.queryTimeout)
		defer cancel()
	}

	start := time.Now()

	// Check if dest is a pointer to a slice of maps
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.IsNil() {
		return fmt.Errorf("dest must be a non-nil pointer")
	}

	destElem := destValue.Elem()
	if destElem.Kind() == reflect.Slice {
		sliceType := destElem.Type().Elem()
		if sliceType.Kind() == reflect.Map &&
			sliceType.Key().Kind() == reflect.String &&
			sliceType.Elem().Kind() == reflect.Interface {

			// Handle slice of map[string]interface{}
			rows, err := exec.QueryxContext(ctx, sql, args...)
			if err != nil {
				return err
			}
			defer rows.Close()

			for rows.Next() {
				row := make(map[string]interface{})
				if err := rows.MapScan(row); err != nil {
					return err
				}
				destElem.Set(reflect.Append(destElem, reflect.ValueOf(row)))
			}

			if qb.enableQueryLogging {
				fmt.Printf("[DEBUG] Query executed in %v\n", time.Since(start))
			}
			return nil
		}
	}

	// Default case: use SelectContext
	err = sqlx.SelectContext(ctx, exec, dest, sql, args...)
	if qb.enableQueryLogging {
		fmt.Printf("[DEBUG] Query executed in %v\n", time.Since(start))
	}
	return err
}

// ExecuteQueryRow executes a query with parameters and returns a single row
func (qb *SQLQueryBuilder) ExecuteQueryRow(ctx context.Context, exec sqlx.ExtContext, query DynamicQuery, dest interface{}) error {
	sql, args, err := qb.BuildQuery(query)
	if err != nil {
		return err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && qb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qb.queryTimeout)
		defer cancel()
	}

	start := time.Now()
	err = sqlx.GetContext(ctx, exec, dest, sql, args...)
	if qb.enableQueryLogging {
		fmt.Printf("[DEBUG] QueryRow executed in %v\n", time.Since(start))
	}
	return err
}

// ExecuteCount executes a count query
func (qb *SQLQueryBuilder) ExecuteCount(ctx context.Context, exec sqlx.ExtContext, query DynamicQuery) (int64, error) {
	sql, args, err := qb.BuildCountQuery(query)
	if err != nil {
		return 0, err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && qb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qb.queryTimeout)
		defer cancel()
	}

	start := time.Now()
	var count int64
	err = sqlx.GetContext(ctx, exec, &count, sql, args...)
	if qb.enableQueryLogging {
		fmt.Printf("[DEBUG] Count query executed in %v\n", time.Since(start))
	}
	return count, err
}

// ExecuteInsert executes an insert operation
func (qb *SQLQueryBuilder) ExecuteInsert(ctx context.Context, exec sqlx.ExtContext, table string, data InsertData, returningColumns ...string) (sql.Result, error) {
	sql, args, err := qb.BuildInsertQuery(table, data, returningColumns...)
	if err != nil {
		return nil, err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && qb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qb.queryTimeout)
		defer cancel()
	}

	start := time.Now()
	result, err := exec.ExecContext(ctx, sql, args...)
	if qb.enableQueryLogging {
		fmt.Printf("[DEBUG] Insert executed in %v\n", time.Since(start))
	}
	return result, err
}

// ExecuteUpdate executes an update operation
func (qb *SQLQueryBuilder) ExecuteUpdate(ctx context.Context, exec sqlx.ExtContext, table string, updateData UpdateData, filters []FilterGroup, returningColumns ...string) (sql.Result, error) {
	sql, args, err := qb.BuildUpdateQuery(table, updateData, filters, returningColumns...)
	if err != nil {
		return nil, err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && qb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qb.queryTimeout)
		defer cancel()
	}

	start := time.Now()
	result, err := exec.ExecContext(ctx, sql, args...)
	if qb.enableQueryLogging {
		fmt.Printf("[DEBUG] Update executed in %v\n", time.Since(start))
	}
	return result, err
}

// ExecuteDelete executes a delete operation
func (qb *SQLQueryBuilder) ExecuteDelete(ctx context.Context, exec sqlx.ExtContext, table string, filters []FilterGroup, returningColumns ...string) (sql.Result, error) {
	sql, args, err := qb.BuildDeleteQuery(table, filters, returningColumns...)
	if err != nil {
		return nil, err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && qb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qb.queryTimeout)
		defer cancel()
	}

	start := time.Now()
	result, err := exec.ExecContext(ctx, sql, args...)
	if qb.enableQueryLogging {
		fmt.Printf("[DEBUG] Delete executed in %v\n", time.Since(start))
	}
	return result, err
}

// ExecuteUpsert executes an upsert operation
func (qb *SQLQueryBuilder) ExecuteUpsert(ctx context.Context, exec sqlx.ExtContext, table string, insertData InsertData, conflictColumns []string, updateColumns []string, returningColumns ...string) (sql.Result, error) {
	sql, args, err := qb.BuildUpsertQuery(table, insertData, conflictColumns, updateColumns, returningColumns...)
	if err != nil {
		return nil, err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && qb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qb.queryTimeout)
		defer cancel()
	}

	start := time.Now()
	result, err := exec.ExecContext(ctx, sql, args...)
	if qb.enableQueryLogging {
		fmt.Printf("[DEBUG] Upsert executed in %v\n", time.Since(start))
	}
	return result, err
}

// buildSetMap builds a map for SetMap from UpdateData
func (qb *SQLQueryBuilder) buildSetMap(updateData UpdateData) map[string]interface{} {
	setMap := make(map[string]interface{})
	for i, col := range updateData.Columns {
		// PERBAIKAN: Pastikan kolom yang benar di-escape
		escapedCol := qb.validateAndEscapeColumn(col)
		if escapedCol != "" {
			setMap[escapedCol] = updateData.Values[i]
		}
	}
	return setMap
}
