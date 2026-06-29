package query

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
)

// buildFilterCondition builds a single filter condition with dialect-specific logic
func isColumnReference(val string) bool {
	if val == "" {
		return false
	}
	// Remove backslash-escaped quotes
	unquoted := strings.ReplaceAll(val, "\\\"", "")
	// Split by dot
	parts := strings.Split(unquoted, ".")
	if len(parts) != 2 {
		return false
	}
	// Both parts should be alphanumeric or quoted identifiers
	validPart := func(s string) bool {
		if s == "" {
			return false
		}
		// Allow quoted parts: "Quoted"
		if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
			return true
		}
		for _, r := range s {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
				return false
			}
		}
		return true
	}
	if !validPart(parts[0]) || !validPart(parts[1]) {
		return false
	}
	return true
}

func (qb *SQLQueryBuilder) buildFilterCondition(filter DynamicFilter) (string, []interface{}, error) {
	column := qb.validateAndEscapeColumn(filter.Column)
	if column == "" {
		return "", nil, fmt.Errorf("invalid or disallowed column: %s", filter.Column)
	}

	// Handle column-to-column comparison if filter.Value is a column reference string
	if valStr, ok := filter.Value.(string); ok && isColumnReference(valStr) {
		escapedVal := qb.escapeColumnReference(valStr)
		switch filter.Operator {
		case OpEqual:
			return fmt.Sprintf("%s = %s", column, escapedVal), nil, nil
		case OpNotEqual:
			return fmt.Sprintf("%s <> %s", column, escapedVal), nil, nil
		case OpGreaterThan:
			return fmt.Sprintf("%s > %s", column, escapedVal), nil, nil
		case OpLessThan:
			return fmt.Sprintf("%s < %s", column, escapedVal), nil, nil
		}
	}

	// Handle JSON operations
	switch filter.Operator {
	case OpJsonContains, OpJsonNotContains, OpJsonExists, OpJsonNotExists, OpJsonEqual, OpJsonNotEqual:
		return qb.dialect.BuildJsonFilter(column, filter)
	case OpArrayContains, OpArrayNotContains, OpArrayLength:
		return qb.dialect.BuildArrayFilter(column, filter)
	}

	// Handle standard operators
	switch filter.Operator {
	case OpEqual:
		if filter.Value == nil {
			return fmt.Sprintf("%s IS NULL", column), nil, nil
		}
		return fmt.Sprintf("%s = ?", column), []interface{}{filter.Value}, nil
	case OpNotEqual:
		if filter.Value == nil {
			return fmt.Sprintf("%s IS NOT NULL", column), nil, nil
		}
		return fmt.Sprintf("%s <> ?", column), []interface{}{filter.Value}, nil
	case OpLike:
		if filter.Value == nil {
			return "", nil, nil
		}
		return fmt.Sprintf("%s LIKE ?", column), []interface{}{filter.Value}, nil
	case OpILike:
		if filter.Value == nil {
			return "", nil, nil
		}
		return qb.dialect.BuildCaseInsensitiveLike(column), []interface{}{filter.Value}, nil
	case OpIn, OpNotIn:
		values := qb.parseArrayValue(filter.Value)
		if len(values) == 0 {
			return "1=0", nil, nil
		}
		op := "IN"
		if filter.Operator == OpNotIn {
			op = "NOT IN"
		}
		placeholders := squirrel.Placeholders(len(values))
		return fmt.Sprintf("%s %s (%s)", column, op, placeholders), values, nil
	case OpGreaterThan, OpGreaterThanEqual, OpLessThan, OpLessThanEqual:
		if filter.Value == nil {
			return "", nil, nil
		}
		opMap := map[FilterOperator]string{
			OpGreaterThan:      ">",
			OpGreaterThanEqual: ">=",
			OpLessThan:         "<",
			OpLessThanEqual:    "<=",
		}
		op, ok := opMap[filter.Operator]
		if !ok {
			return "", nil, fmt.Errorf("unhandled comparison operator: %s", filter.Operator)
		}
		return fmt.Sprintf("%s %s ?", column, op), []interface{}{filter.Value}, nil
	case OpBetween, OpNotBetween:
		values := qb.parseArrayValue(filter.Value)
		if len(values) != 2 {
			return "", nil, fmt.Errorf("between operator requires exactly 2 values")
		}
		op := "BETWEEN"
		if filter.Operator == OpNotBetween {
			op = "NOT BETWEEN"
		}
		return fmt.Sprintf("%s %s ? AND ?", column, op), []interface{}{values[0], values[1]}, nil
	case OpNull:
		return fmt.Sprintf("%s IS NULL", column), nil, nil
	case OpNotNull:
		return fmt.Sprintf("%s IS NOT NULL", column), nil, nil
	case OpContains, OpNotContains, OpStartsWith, OpEndsWith:
		if filter.Value == nil {
			return "", nil, nil
		}
		var value string
		switch filter.Operator {
		case OpContains, OpNotContains:
			value = fmt.Sprintf("%%%v%%", filter.Value)
		case OpStartsWith:
			value = fmt.Sprintf("%v%%", filter.Value)
		case OpEndsWith:
			value = fmt.Sprintf("%%%v", filter.Value)
		}

		switch qb.dbType {
		case DBTypePostgreSQL, DBTypeSQLite:
			op := "ILIKE"
			if strings.Contains(string(filter.Operator), "Not") {
				op = "NOT ILIKE"
			}
			return fmt.Sprintf("%s %s ?", column, op), []interface{}{value}, nil
		case DBTypeMySQL, DBTypeSQLServer:
			op := "LIKE"
			if strings.Contains(string(filter.Operator), "Not") {
				op = "NOT LIKE"
			}
			return fmt.Sprintf("LOWER(%s) %s LOWER(?)", column, op), []interface{}{value}, nil
		default:
			op := "LIKE"
			if strings.Contains(string(filter.Operator), "Not") {
				op = "NOT LIKE"
			}
			return fmt.Sprintf("%s %s ?", column, op), []interface{}{value}, nil
		}
	default:
		return "", nil, fmt.Errorf("unsupported operator: %s", filter.Operator)
	}
}

// parseArrayValue parses an array value for SQL queries
func (qb *SQLQueryBuilder) parseArrayValue(value interface{}) []interface{} {
	if value == nil {
		return nil
	}
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		v := reflect.ValueOf(value)
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
		}
		return result
	}
	if str, ok := value.(string); ok {
		if strings.Contains(str, ",") {
			parts := strings.Split(str, ",")
			result := make([]interface{}, len(parts))
			for i, part := range parts {
				result[i] = strings.TrimSpace(part)
			}
			return result
		}
		return []interface{}{str}
	}
	return []interface{}{value}
}
