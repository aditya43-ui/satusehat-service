package query

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xwb1989/sqlparser"
)

// validateAndEscapeColumn validates and escapes a column name
func (qb *SQLQueryBuilder) validateAndEscapeColumn(field string) string {
	if field == "" {
		return ""
	}
	// Allow complex expressions like functions
	if strings.Contains(field, "(") {
		if qb.isValidExpression(field) {
			return field // Don't escape complex expressions, assume they are safe
		}
		return ""
	}

	// For simple column names (not containing a dot), check against the allow list.
	// For dotted names (table.column), we assume the check is not needed or handled elsewhere.
	// The dialect's EscapeIdentifier will handle splitting and escaping each part of a dotted name.
	if !strings.Contains(field, ".") {
		if qb.allowedColumns != nil && !qb.allowedColumns[field] {
			// Log or handle disallowed column
			return ""
		}
	}

	// Delegate the actual escaping to the dialect, which handles simple and dotted names correctly.
	return qb.dialect.EscapeIdentifier(field)
}

// PERBAIKAN AKHIR: Sanitasi SQL penuh untuk kompatibilitas parser MySQL
func (qb *SQLQueryBuilder) validateParsedSQL(sql string) error {
	// --- Langkah 1: Sanitasi untuk membuat SQL kompatibel dengan parser MySQL ---

	// PERBAIKAN 1: Ganti quote PostgreSQL (") dengan quote MySQL (`)
	sanitizedSQL := strings.ReplaceAll(sql, `"`, "`")

	// PERBAIKAN 2: Ganti placeholder PostgreSQL ($1, $2, ...) dengan placeholder MySQL (?)
	// Regex ini mencocokan '$' diikuti oleh satu atau lebih digit.
	re := regexp.MustCompile(`\$\d+`)
	sanitizedSQL = re.ReplaceAllString(sanitizedSQL, "?")

	// PERBAIKAN 3: Ganti ILIKE dengan LIKE karena parser MySQL tidak mengenali ILIKE
	reILike := regexp.MustCompile(`(?i)\bilike\b`)
	sanitizedSQL = reILike.ReplaceAllString(sanitizedSQL, "LIKE")

	// --- Langkah 2: Parse SQL yang sudah disanitasi ---
	stmt, err := sqlparser.Parse(sanitizedSQL)
	if err != nil {
		// Jika SQL tidak valid (bahkan untuk parser MySQL), kita anggap ini tidak aman.
		return fmt.Errorf("invalid SQL syntax detected during validation: %w", err)
	}

	// --- Langkah 3: Validasi struktur query (logikanya tetap sama) ---
	switch stmt.(type) {
	case *sqlparser.Select:
		// Izinkan statement SELECT sederhana.
		return nil // Aman

	case *sqlparser.Union:
		// Tolak statement UNION.
		return fmt.Errorf("UNION statements are not allowed in this context")

	case *sqlparser.Insert, *sqlparser.Update, *sqlparser.Delete:
		// Tolak statement DML.
		return fmt.Errorf("DML statements (INSERT/UPDATE/DELETE) are not allowed from user input")

	case *sqlparser.DDL:
		// Tolak statement DDL.
		return fmt.Errorf("DDL statements are not allowed from user input")

	default:
		// Blokir statement lain.
		return fmt.Errorf("unsupported or disallowed SQL statement type: %T", stmt)
	}
}
func (qb *SQLQueryBuilder) isValidExpression(expr string) bool {
	// This is a simplified check. A more robust solution might use a proper SQL parser library.
	// For now, we allow alphanumeric, underscore, dots, parentheses, and common operators.
	// For SQL Server, allow brackets [] and spaces for column names.
	allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.,() *-/[]"
	for _, r := range expr {
		if !strings.ContainsRune(allowedChars, r) {
			return false
		}
	}

	// PERBAIKAN: Gunakan Regex dengan word boundary untuk menghindari false positive
	// Ini akan mencegah "DeletedAt" dianggap sebagai "delete"
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(--|/\*|\*/)`),                                   // SQL comments
		regexp.MustCompile(`(?i)\bunion\b\s+\bselect\b`),                         // UNION followed by SELECT
		regexp.MustCompile(`(?i)\bselect\b`),                                     // Standalone SELECT (bisa disesuaikan)
		regexp.MustCompile(`(?i)\binsert\b`),                                     // Standalone INSERT
		regexp.MustCompile(`(?i)\bupdate\b`),                                     // Standalone UPDATE
		regexp.MustCompile(`(?i)\bdelete\b`),                                     // Standalone DELETE
		regexp.MustCompile(`(?i)\bdrop\b`),                                       // DROP
		regexp.MustCompile(`(?i)\balter\b`),                                      // ALTER
		regexp.MustCompile(`(?i)\bcreate\b`),                                     // CREATE
		regexp.MustCompile(`(?i)\b(exec|execute)\s*\(`),                          // EXEC/EXECUTE functions
		regexp.MustCompile(`(?i)\b(waitfor\s+delay|benchmark|sleep)\s*\(`),       // Time-based attacks
		regexp.MustCompile(`(?i)\b(information_schema|sysobjects|syscolumns)\b`), // DB enumeration
	}

	lowerExpr := strings.ToLower(expr)
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(lowerExpr) {
			// Log untuk debugging, jadi kita tahu pola mana yang cocok
			// fmt.Printf("[DEBUG] Potentially dangerous expression detected: %s matched by pattern %s\n", expr, pattern.String())
			return false
		}
	}
	return true
}

// isValidFunctionName validates if a function name is valid
func (qb *SQLQueryBuilder) isValidFunctionName(name string) bool {
	// Check if the function name is a valid SQL function
	validFunctions := map[string]bool{
		// Aggregate functions
		"count": true, "sum": true, "avg": true, "min": true, "max": true,
		// Window functions
		"row_number": true, "rank": true, "dense_rank": true, "ntile": true,
		"lag": true, "lead": true, "first_value": true, "last_value": true,
		// JSON functions
		"json_extract": true, "json_contains": true, "json_search": true,
		"json_array": true, "json_object": true, "json_merge": true,
		// Other functions
		"concat": true, "substring": true, "upper": true, "lower": true,
		"trim": true, "coalesce": true, "nullif": true, "isnull": true,
	}

	return validFunctions[strings.ToLower(name)]
}

// escapeColumnReference escapes a column reference (table.column)
func (qb *SQLQueryBuilder) escapeColumnReference(col string) string {
	return qb.dialect.EscapeIdentifier(col)
}

// escapeJsonPath escapes a JSON path for PostgreSQL
func (qb *SQLQueryBuilder) escapeJsonPath(path string) string {
	// Simple implementation - in a real scenario, you'd need more sophisticated escaping
	return "'" + strings.ReplaceAll(path, "'", "''") + "'"
}

// escapeSqlServerJsonPath escapes a JSON path for SQL Server
func (qb *SQLQueryBuilder) escapeSqlServerJsonPath(path string) string {
	// Convert JSONPath to SQL Server format
	// $.path.to.property -> '$.path.to.property'
	if !strings.HasPrefix(path, "$") {
		path = "$." + path
	}
	return strings.ReplaceAll(path, ".", ".")
}

// checkForSqlInjectionInArgs checks for potential SQL injection patterns in query arguments
func (qb *SQLQueryBuilder) checkForSqlInjectionInArgs(args []interface{}) error {
	if !qb.enableSecurityChecks {
		return nil
	}

	for _, arg := range args {
		if str, ok := arg.(string); ok {
			lowerStr := strings.ToLower(str)
			// Check for dangerous patterns specifically in user input values
			dangerousPatterns := []*regexp.Regexp{
				regexp.MustCompile(`(?i)(union\s+select)`),
				regexp.MustCompile(`(?i)(or\s+1\s*=\s*1)`),
				regexp.MustCompile(`(?i)(and\s+true)`),
				regexp.MustCompile(`(?i)(waitfor\s+delay)`),
				regexp.MustCompile(`(?i)(benchmark|sleep)\s*\(`),
				regexp.MustCompile(`(?i)(pg_sleep)\s*\(`),
				regexp.MustCompile(`(?i)(load_file|into\s+outfile)`),
				regexp.MustCompile(`(?i)(information_schema|sysobjects|syscolumns)`),
				regexp.MustCompile(`(?i)(--|\/\*|\*\/)`),
			}

			for _, pattern := range dangerousPatterns {
				if pattern.MatchString(lowerStr) {
					return fmt.Errorf("potential SQL injection detected in query argument: pattern %s matched", pattern.String())
				}
			}
		}
	}
	return nil
}

// checkForSqlInjectionInSQL checks for potential SQL injection patterns in the final SQL
func (qb *SQLQueryBuilder) checkForSqlInjectionInSQL(sql string) error {
	if !qb.enableSecurityChecks {
		return nil
	}

	// Check for dangerous patterns in the final SQL
	// But allow valid SQL keywords in their proper context
	lowerSQL := strings.ToLower(sql)

	// More specific patterns that actually indicate injection attempts
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select)`),                                        // UNION followed by SELECT
		regexp.MustCompile(`(?i)(select\s+.*\s+from\s+.*\s+where\s+.*\s+or\s+1\s*=\s*1)`), // Classic SQL injection
		regexp.MustCompile(`(?i)(drop\s+table)`),                                          // DROP TABLE
		regexp.MustCompile(`(?i)(delete\s+from)`),                                         // DELETE FROM
		regexp.MustCompile(`(?i)(insert\s+into)`),                                         // INSERT INTO
		regexp.MustCompile(`(?i)(update\s+.*\s+set)`),                                     // UPDATE SET
		regexp.MustCompile(`(?i)(alter\s+table)`),                                         // ALTER TABLE
		regexp.MustCompile(`(?i)(create\s+table)`),                                        // CREATE TABLE
		regexp.MustCompile(`(?i)(exec\s*\(|execute\s*\()`),                                // EXEC/EXECUTE functions
		regexp.MustCompile(`(?i)(--|\/\*|\*\/)`),                                          // SQL comments
	}

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(lowerSQL) {
			return fmt.Errorf("potential SQL injection detected in SQL: pattern %s matched", pattern.String())
		}
	}

	return nil
}
