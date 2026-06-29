package query

import (
	"fmt"
	"strings"
)

// DatabaseDialect defines the interface for database-specific operations
type DatabaseDialect interface {
	// SQL generation methods
	Placeholder() string
	EscapeIdentifier(identifier string) string
	BuildLimitOffset(limit, offset int) string
	BuildJsonExtract(column, path string) string
	BuildJsonContains(column, path string) (string, []interface{})
	BuildJsonExists(column, path string) string
	BuildArrayContains(column string) string
	BuildJsonFilter(column string, filter DynamicFilter) (string, []interface{}, error)
	BuildArrayFilter(column string, filter DynamicFilter) (string, []interface{}, error)
	BuildArrayLength(column string) string
	BuildCaseInsensitiveLike(column string) string
	BuildUpsert(table string, conflictColumns, updateColumns []string) string

	// Type information
	GetType() DBType
}

// BaseDialect provides common functionality for all dialects
type BaseDialect struct {
	dbType DBType
}

func (d *BaseDialect) GetType() DBType {
	return d.dbType
}

// PostgreSQLDialect implements DatabaseDialect for PostgreSQL
type PostgreSQLDialect struct {
	BaseDialect
}

func NewPostgreSQLDialect() *PostgreSQLDialect {
	return &PostgreSQLDialect{
		BaseDialect: BaseDialect{dbType: DBTypePostgreSQL},
	}
}

func (d *PostgreSQLDialect) Placeholder() string {
	return "$%d"
}

func (d *PostgreSQLDialect) EscapeIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	escapedParts := make([]string, len(parts))
	for i, part := range parts {
		// Hindari meng-quote karakter '*' yang digunakan di COUNT(*)
		if part == "*" {
			escapedParts[i] = "*"
		} else if len(part) > 1 && strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			// Already quoted
			escapedParts[i] = part
		} else {
			escapedParts[i] = `"` + strings.ReplaceAll(part, `"`, `""`) + `"`
		}
	}
	return strings.Join(escapedParts, ".")
}

func (d *PostgreSQLDialect) BuildLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf("LIMIT %d", limit)
	} else if offset > 0 {
		return fmt.Sprintf("OFFSET %d", offset)
	}
	return ""
}

func (d *PostgreSQLDialect) BuildJsonExtract(column, path string) string {
	return fmt.Sprintf("%s->'%s'", column, path)
}

func (d *PostgreSQLDialect) BuildJsonContains(column, path string) (string, []interface{}) {
	return fmt.Sprintf("%s @> ?", column), []interface{}{}
}

func (d *PostgreSQLDialect) BuildJsonExists(column, path string) string {
	return fmt.Sprintf("jsonb_path_exists(%s, '%s')", column, path)
}

func (d *PostgreSQLDialect) BuildArrayContains(column string) string {
	return fmt.Sprintf("? = ANY(%s)", column)
}

func (d *PostgreSQLDialect) BuildArrayLength(column string) string {
	return fmt.Sprintf("array_length(%s, 1)", column)
}

func (d *PostgreSQLDialect) BuildCaseInsensitiveLike(column string) string {
	return fmt.Sprintf("%s ILIKE ?", column)
}

func (d *PostgreSQLDialect) BuildUpsert(table string, conflictColumns, updateColumns []string) string {
	if len(conflictColumns) == 0 {
		return ""
	}

	conflictTarget := strings.Join(conflictColumns, ", ")
	setClause := ""
	for _, col := range updateColumns {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = EXCLUDED.%s", d.EscapeIdentifier(col), d.EscapeIdentifier(col))
	}

	return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", conflictTarget, setClause)
}

func (d *PostgreSQLDialect) BuildJsonFilter(column string, filter DynamicFilter) (string, []interface{}, error) {
	path := "$"
	if pathOption, ok := filter.Options["path"].(string); ok && pathOption != "" {
		path = pathOption
	}

	var expr string
	var args []interface{}

	switch filter.Operator {
	case OpJsonContains:
		expr = fmt.Sprintf("%s @> ?", column)
		args = append(args, filter.Value)
	case OpJsonNotContains:
		expr = fmt.Sprintf("NOT (%s @> ?)", column)
		args = append(args, filter.Value)
	case OpJsonExists:
		expr = fmt.Sprintf("jsonb_path_exists(%s, '%s')", column, path)
	case OpJsonNotExists:
		expr = fmt.Sprintf("NOT jsonb_path_exists(%s, '%s')", column, path)
	case OpJsonEqual:
		expr = fmt.Sprintf("%s->>'%s' = ?", column, path)
		args = append(args, filter.Value)
	case OpJsonNotEqual:
		expr = fmt.Sprintf("%s->>'%s' <> ?", column, path)
		args = append(args, filter.Value)
	default:
		return "", nil, fmt.Errorf("unsupported JSON operator for PostgreSQL: %s", filter.Operator)
	}
	return expr, args, nil
}

func (d *PostgreSQLDialect) BuildArrayFilter(column string, filter DynamicFilter) (string, []interface{}, error) {
	var expr string
	var args []interface{}

	switch filter.Operator {
	case OpArrayContains:
		expr = fmt.Sprintf("? = ANY(%s)", column)
		args = append(args, filter.Value)
	case OpArrayNotContains:
		expr = fmt.Sprintf("? <> ALL(%s)", column)
		args = append(args, filter.Value)
	case OpArrayLength:
		if lengthOption, ok := filter.Options["length"].(int); ok {
			expr = fmt.Sprintf("array_length(%s, 1) = ?", column)
			args = append(args, lengthOption)
		} else {
			return "", nil, fmt.Errorf("array_length operator requires 'length' option")
		}
	default:
		return "", nil, fmt.Errorf("unsupported array operator for PostgreSQL: %s", filter.Operator)
	}
	return expr, args, nil
}

// MySQLDialect implements DatabaseDialect for MySQL
type MySQLDialect struct {
	BaseDialect
}

func NewMySQLDialect() *MySQLDialect {
	return &MySQLDialect{
		BaseDialect: BaseDialect{dbType: DBTypeMySQL},
	}
}

func (d *MySQLDialect) Placeholder() string {
	return "?"
}

func (d *MySQLDialect) EscapeIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	escapedParts := make([]string, len(parts))
	for i, part := range parts {
		if part == "*" {
			escapedParts[i] = "*"
		} else if len(part) > 1 && strings.HasPrefix(part, "`") && strings.HasSuffix(part, "`") {
			// Already quoted
			escapedParts[i] = part
		} else {
			escapedParts[i] = "`" + strings.ReplaceAll(part, "`", "``") + "`"
		}
	}
	return strings.Join(escapedParts, ".")
}

func (d *MySQLDialect) BuildLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf("LIMIT %d, %d", offset, limit)
	} else if limit > 0 {
		return fmt.Sprintf("LIMIT %d", limit)
	} else if offset > 0 {
		return fmt.Sprintf("LIMIT %d, 18446744073709551615", offset)
	}
	return ""
}

func (d *MySQLDialect) BuildJsonExtract(column, path string) string {
	return fmt.Sprintf("JSON_EXTRACT(%s, '$.%s')", column, path)
}

func (d *MySQLDialect) BuildJsonContains(column, path string) (string, []interface{}) {
	return fmt.Sprintf("JSON_CONTAINS(%s, ?, '$.%s')", column, path), []interface{}{}
}

func (d *MySQLDialect) BuildJsonExists(column, path string) string {
	return fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '$.%s')", column, path)
}

func (d *MySQLDialect) BuildArrayContains(column string) string {
	return fmt.Sprintf("JSON_CONTAINS(%s, JSON_QUOTE(?))", column)
}

func (d *MySQLDialect) BuildArrayLength(column string) string {
	return fmt.Sprintf("JSON_LENGTH(%s)", column)
}

func (d *MySQLDialect) BuildCaseInsensitiveLike(column string) string {
	return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", column)
}

func (d *MySQLDialect) BuildUpsert(table string, conflictColumns, updateColumns []string) string {
	if len(updateColumns) == 0 {
		return ""
	}

	setClause := ""
	for _, col := range updateColumns {
		if setClause != "" {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = VALUES(%s)", d.EscapeIdentifier(col), d.EscapeIdentifier(col))
	}

	return fmt.Sprintf("ON DUPLICATE KEY UPDATE %s", setClause)
}

func (d *MySQLDialect) BuildJsonFilter(column string, filter DynamicFilter) (string, []interface{}, error) {
	path := "$"
	if pathOption, ok := filter.Options["path"].(string); ok && pathOption != "" {
		path = pathOption
	}

	var expr string
	var args []interface{}

	switch filter.Operator {
	case OpJsonContains:
		expr = fmt.Sprintf("JSON_CONTAINS(%s, ?, '%s')", column, path)
		args = append(args, filter.Value)
	case OpJsonNotContains:
		expr = fmt.Sprintf("NOT JSON_CONTAINS(%s, ?, '%s')", column, path)
		args = append(args, filter.Value)
	case OpJsonExists:
		expr = fmt.Sprintf("JSON_CONTAINS_PATH(%s, 'one', '%s')", column, path)
	case OpJsonNotExists:
		expr = fmt.Sprintf("NOT JSON_CONTAINS_PATH(%s, 'one', '%s')", column, path)
	case OpJsonEqual:
		expr = fmt.Sprintf("JSON_EXTRACT(%s, '%s') = ?", column, path)
		args = append(args, filter.Value)
	case OpJsonNotEqual:
		expr = fmt.Sprintf("JSON_EXTRACT(%s, '%s') <> ?", column, path)
		args = append(args, filter.Value)
	default:
		return "", nil, fmt.Errorf("unsupported JSON operator for MySQL: %s", filter.Operator)
	}
	return expr, args, nil
}

func (d *MySQLDialect) BuildArrayFilter(column string, filter DynamicFilter) (string, []interface{}, error) {
	var expr string
	var args []interface{}

	switch filter.Operator {
	case OpArrayContains:
		expr = fmt.Sprintf("JSON_CONTAINS(%s, JSON_QUOTE(?))", column)
		args = append(args, filter.Value)
	case OpArrayNotContains:
		expr = fmt.Sprintf("NOT JSON_CONTAINS(%s, JSON_QUOTE(?))", column)
		args = append(args, filter.Value)
	case OpArrayLength:
		if lengthOption, ok := filter.Options["length"].(int); ok {
			expr = fmt.Sprintf("JSON_LENGTH(%s) = ?", column)
			args = append(args, lengthOption)
		} else {
			return "", nil, fmt.Errorf("array_length operator requires 'length' option")
		}
	default:
		return "", nil, fmt.Errorf("unsupported array operator for MySQL: %s", filter.Operator)
	}
	return expr, args, nil
}

// SQLiteDialect implements DatabaseDialect for SQLite
type SQLiteDialect struct {
	BaseDialect
}

func NewSQLiteDialect() *SQLiteDialect {
	return &SQLiteDialect{
		BaseDialect: BaseDialect{dbType: DBTypeSQLite},
	}
}

func (d *SQLiteDialect) Placeholder() string {
	return "?"
}

func (d *SQLiteDialect) EscapeIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	escapedParts := make([]string, len(parts))
	for i, part := range parts {
		if part == "*" {
			escapedParts[i] = "*"
		} else if len(part) > 1 && strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			// Already quoted
			escapedParts[i] = part
		} else {
			escapedParts[i] = `"` + strings.ReplaceAll(part, `"`, `""`) + `"`
		}
	}
	return strings.Join(escapedParts, ".")
}

func (d *SQLiteDialect) BuildLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf("LIMIT %d", limit)
	} else if offset > 0 {
		return fmt.Sprintf("OFFSET %d", offset)
	}
	return ""
}

func (d *SQLiteDialect) BuildJsonExtract(column, path string) string {
	return fmt.Sprintf("json_extract(%s, '$.%s')", column, path)
}

func (d *SQLiteDialect) BuildJsonContains(column, path string) (string, []interface{}) {
	return fmt.Sprintf("json_extract(%s, '$.%s') = ?", column, path), []interface{}{}
}

func (d *SQLiteDialect) BuildJsonExists(column, path string) string {
	return fmt.Sprintf("json_extract(%s, '$.%s') IS NOT NULL", column, path)
}

func (d *SQLiteDialect) BuildArrayContains(column string) string {
	return fmt.Sprintf("EXISTS (SELECT 1 FROM json_each(%s) WHERE json_each.value = ?)", column)
}

func (d *SQLiteDialect) BuildArrayLength(column string) string {
	return fmt.Sprintf("json_array_length(%s)", column)
}

func (d *SQLiteDialect) BuildCaseInsensitiveLike(column string) string {
	return fmt.Sprintf("%s LIKE ?", column)
}

func (d *SQLiteDialect) BuildUpsert(table string, conflictColumns, updateColumns []string) string {
	// SQLite doesn't have ON CONFLICT for INSERT, use INSERT OR REPLACE or separate logic
	return ""
}

func (d *SQLiteDialect) BuildJsonFilter(column string, filter DynamicFilter) (string, []interface{}, error) {
	path := "$"
	if pathOption, ok := filter.Options["path"].(string); ok && pathOption != "" {
		path = pathOption
	}

	var expr string
	var args []interface{}

	switch filter.Operator {
	case OpJsonContains:
		expr = fmt.Sprintf("json_extract(%s, '%s') = ?", column, path)
		args = append(args, filter.Value)
	case OpJsonNotContains:
		expr = fmt.Sprintf("json_extract(%s, '%s') <> ?", column, path)
		args = append(args, filter.Value)
	case OpJsonExists:
		expr = fmt.Sprintf("json_extract(%s, '%s') IS NOT NULL", column, path)
	case OpJsonNotExists:
		expr = fmt.Sprintf("json_extract(%s, '%s') IS NULL", column, path)
	case OpJsonEqual:
		expr = fmt.Sprintf("json_extract(%s, '%s') = ?", column, path)
		args = append(args, filter.Value)
	case OpJsonNotEqual:
		expr = fmt.Sprintf("json_extract(%s, '%s') <> ?", column, path)
		args = append(args, filter.Value)
	default:
		return "", nil, fmt.Errorf("unsupported JSON operator for SQLite: %s", filter.Operator)
	}
	return expr, args, nil
}

func (d *SQLiteDialect) BuildArrayFilter(column string, filter DynamicFilter) (string, []interface{}, error) {
	var expr string
	var args []interface{}

	switch filter.Operator {
	case OpArrayContains:
		expr = fmt.Sprintf("EXISTS (SELECT 1 FROM json_each(%s) WHERE json_each.value = ?)", column)
		args = append(args, filter.Value)
	case OpArrayNotContains:
		expr = fmt.Sprintf("NOT EXISTS (SELECT 1 FROM json_each(%s) WHERE json_each.value = ?)", column)
		args = append(args, filter.Value)
	case OpArrayLength:
		if lengthOption, ok := filter.Options["length"].(int); ok {
			expr = fmt.Sprintf("json_array_length(%s) = ?", column)
			args = append(args, lengthOption)
		} else {
			return "", nil, fmt.Errorf("array_length operator requires 'length' option")
		}
	default:
		return "", nil, fmt.Errorf("unsupported array operator for SQLite: %s", filter.Operator)
	}
	return expr, args, nil
}

// SQLServerDialect implements DatabaseDialect for SQL Server
type SQLServerDialect struct {
	BaseDialect
}

func NewSQLServerDialect() *SQLServerDialect {
	return &SQLServerDialect{
		BaseDialect: BaseDialect{dbType: DBTypeSQLServer},
	}
}

func (d *SQLServerDialect) Placeholder() string {
	return "@p%d"
}

func (d *SQLServerDialect) EscapeIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	escapedParts := make([]string, len(parts))
	for i, part := range parts {
		if part == "*" {
			escapedParts[i] = "*"
		} else if len(part) > 1 && strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Already quoted
			escapedParts[i] = part
		} else {
			escapedParts[i] = "[" + strings.ReplaceAll(part, "]", "]]") + "]"
		}
	}
	return strings.Join(escapedParts, ".")
}

func (d *SQLServerDialect) BuildLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
	} else if limit > 0 {
		return fmt.Sprintf("FETCH NEXT %d ROWS ONLY", limit)
	} else if offset > 0 {
		return fmt.Sprintf("OFFSET %d ROWS", offset)
	}
	return ""
}

func (d *SQLServerDialect) BuildJsonExtract(column, path string) string {
	return fmt.Sprintf("JSON_VALUE(%s, '%s')", column, d.escapeSqlServerJsonPath(path))
}

func (d *SQLServerDialect) BuildJsonContains(column, path string) (string, []interface{}) {
	return fmt.Sprintf("JSON_VALUE(%s, '%s') = ?", column, d.escapeSqlServerJsonPath(path)), []interface{}{}
}

func (d *SQLServerDialect) BuildJsonExists(column, path string) string {
	return fmt.Sprintf("JSON_VALUE(%s, '%s') IS NOT NULL", column, d.escapeSqlServerJsonPath(path))
}

func (d *SQLServerDialect) BuildArrayContains(column string) string {
	return fmt.Sprintf("? IN (SELECT value FROM OPENJSON(%s))", column)
}

func (d *SQLServerDialect) BuildArrayLength(column string) string {
	return fmt.Sprintf("(SELECT COUNT(*) FROM OPENJSON(%s))", column)
}

func (d *SQLServerDialect) BuildCaseInsensitiveLike(column string) string {
	return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", column)
}

func (d *SQLServerDialect) BuildUpsert(table string, conflictColumns, updateColumns []string) string {
	// SQL Server uses MERGE for upsert operations
	return ""
}

func (d *SQLServerDialect) escapeSqlServerJsonPath(path string) string {
	// Convert $.path to proper SQL Server JSON path format
	if strings.HasPrefix(path, "$.") {
		return path
	}
	return "$." + path
}

func (d *SQLServerDialect) BuildJsonFilter(column string, filter DynamicFilter) (string, []interface{}, error) {
	path := "$"
	if pathOption, ok := filter.Options["path"].(string); ok && pathOption != "" {
		path = pathOption
	}

	var expr string
	var args []interface{}

	switch filter.Operator {
	case OpJsonContains:
		expr = fmt.Sprintf("JSON_VALUE(%s, '%s') = ?", column, d.escapeSqlServerJsonPath(path))
		args = append(args, filter.Value)
	case OpJsonNotContains:
		expr = fmt.Sprintf("JSON_VALUE(%s, '%s') <> ?", column, d.escapeSqlServerJsonPath(path))
		args = append(args, filter.Value)
	case OpJsonExists:
		expr = fmt.Sprintf("JSON_VALUE(%s, '%s') IS NOT NULL", column, d.escapeSqlServerJsonPath(path))
	case OpJsonNotExists:
		expr = fmt.Sprintf("JSON_VALUE(%s, '%s') IS NULL", column, d.escapeSqlServerJsonPath(path))
	case OpJsonEqual:
		expr = fmt.Sprintf("JSON_VALUE(%s, '%s') = ?", column, d.escapeSqlServerJsonPath(path))
		args = append(args, filter.Value)
	case OpJsonNotEqual:
		expr = fmt.Sprintf("JSON_VALUE(%s, '%s') <> ?", column, d.escapeSqlServerJsonPath(path))
		args = append(args, filter.Value)
	default:
		return "", nil, fmt.Errorf("unsupported JSON operator for SQL Server: %s", filter.Operator)
	}
	return expr, args, nil
}

func (d *SQLServerDialect) BuildArrayFilter(column string, filter DynamicFilter) (string, []interface{}, error) {
	var expr string
	var args []interface{}

	switch filter.Operator {
	case OpArrayContains:
		expr = fmt.Sprintf("? IN (SELECT value FROM OPENJSON(%s))", column)
		args = append(args, filter.Value)
	case OpArrayNotContains:
		expr = fmt.Sprintf("? NOT IN (SELECT value FROM OPENJSON(%s))", column)
		args = append(args, filter.Value)
	case OpArrayLength:
		if lengthOption, ok := filter.Options["length"].(int); ok {
			expr = fmt.Sprintf("(SELECT COUNT(*) FROM OPENJSON(%s)) = ?", column)
			args = append(args, lengthOption)
		} else {
			return "", nil, fmt.Errorf("array_length operator requires 'length' option")
		}
	default:
		return "", nil, fmt.Errorf("unsupported array operator for SQL Server: %s", filter.Operator)
	}
	return expr, args, nil
}

// GetDialect returns the appropriate dialect for the given database type
func GetDialect(dbType DBType) DatabaseDialect {
	switch dbType {
	case DBTypePostgreSQL:
		return NewPostgreSQLDialect()
	case DBTypeMySQL:
		return NewMySQLDialect()
	case DBTypeSQLite:
		return NewSQLiteDialect()
	case DBTypeSQLServer:
		return NewSQLServerDialect()
	default:
		return NewPostgreSQLDialect() // Default fallback
	}
}
