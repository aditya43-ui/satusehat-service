package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// QueryParserImpl implements QueryParser interface
type QueryParserImpl struct {
	defaultLimit int
	maxLimit     int
}

// NewQueryParser creates a new query parser instance
func NewQueryParser() *QueryParserImpl {
	return &QueryParserImpl{defaultLimit: 10, maxLimit: 100}
}

// SetLimits sets default and maximum limits for pagination
func (qp *QueryParserImpl) SetLimits(defaultLimit, maxLimit int) QueryParser {
	qp.defaultLimit = defaultLimit
	qp.maxLimit = maxLimit
	return qp
}

// ParseQuery parses URL query parameters into a DynamicQuery struct
func (qp *QueryParserImpl) ParseQuery(values interface{}, defaultTable string) (DynamicQuery, error) {
	query := DynamicQuery{
		From:   defaultTable,
		Limit:  qp.defaultLimit,
		Offset: 0,
	}

	// Convert values to url.Values
	var urlValues url.Values
	switch v := values.(type) {
	case url.Values:
		urlValues = v
	case map[string][]string:
		urlValues = make(url.Values)
		for key, vals := range v {
			urlValues[key] = vals
		}
	default:
		return query, fmt.Errorf("unsupported values type: %T", values)
	}

	// Parse fields
	if fields := urlValues.Get("fields"); fields != "" {
		if fields == "*" {
			query.Fields = []SelectField{{Expression: "*"}}
		} else {
			fieldList := strings.Split(fields, ",")
			for _, field := range fieldList {
				query.Fields = append(query.Fields, SelectField{Expression: strings.TrimSpace(field)})
			}
		}
	} else {
		query.Fields = []SelectField{{Expression: "*"}}
	}

	// Parse pagination
	if limit := urlValues.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= qp.maxLimit {
			query.Limit = l
		}
	}
	if offset := urlValues.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			query.Offset = o
		}
	}

	// Parse filters
	filters, err := qp.parseFilters(urlValues)
	if err != nil {
		return query, err
	}
	query.Filters = filters

	// Parse sorting
	sorts, err := qp.parseSorting(urlValues)
	if err != nil {
		return query, err
	}
	query.Sort = sorts

	return query, nil
}

// ParseQueryWithDefaultFields parses URL query parameters into a DynamicQuery struct with default fields
func (qp *QueryParserImpl) ParseQueryWithDefaultFields(values interface{}, defaultTable string, defaultFields []string) (DynamicQuery, error) {
	query, err := qp.ParseQuery(values, defaultTable)
	if err != nil {
		return query, err
	}

	// If no fields specified, use default fields
	if len(query.Fields) == 0 || (len(query.Fields) == 1 && query.Fields[0].Expression == "*") {
		query.Fields = make([]SelectField, len(defaultFields))
		for i, field := range defaultFields {
			query.Fields[i] = SelectField{Expression: field}
		}
	}

	return query, nil
}

// parseFilters parses filter parameters from URL values
func (qp *QueryParserImpl) parseFilters(values url.Values) ([]FilterGroup, error) {
	filterMap := make(map[string]map[string]string)
	for key, vals := range values {
		if strings.HasPrefix(key, "filter[") && strings.HasSuffix(key, "]") {
			parts := strings.Split(key[7:len(key)-1], "][")
			if len(parts) == 2 {
				column, operator := parts[0], parts[1]
				if filterMap[column] == nil {
					filterMap[column] = make(map[string]string)
				}
				if len(vals) > 0 {
					filterMap[column][operator] = vals[0]
				}
			}
		}
	}
	if len(filterMap) == 0 {
		return nil, nil
	}
	var filters []DynamicFilter
	for column, operators := range filterMap {
		for opStr, value := range operators {
			operator := FilterOperator(opStr)
			var parsedValue interface{}
			switch operator {
			case OpIn, OpNotIn:
				if value != "" {
					parsedValue = strings.Split(value, ",")
				}
			case OpBetween, OpNotBetween:
				if value != "" {
					parts := strings.Split(value, ",")
					if len(parts) == 2 {
						parsedValue = []interface{}{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])}
					}
				}
			case OpNull, OpNotNull:
				parsedValue = nil
			default:
				parsedValue = value
			}
			filters = append(filters, DynamicFilter{Column: column, Operator: operator, Value: parsedValue})
		}
	}
	if len(filters) == 0 {
		return nil, nil
	}
	return []FilterGroup{{Filters: filters, LogicOp: "AND"}}, nil
}

// parseSorting parses sorting parameters from URL values
func (qp *QueryParserImpl) parseSorting(values url.Values) ([]SortField, error) {
	sortParam := values.Get("sort")
	if sortParam == "" {
		return nil, nil
	}
	var sorts []SortField
	fields := strings.Split(sortParam, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		order, column := "ASC", field
		if strings.HasPrefix(field, "-") {
			order = "DESC"
			column = field[1:]
		} else if strings.HasPrefix(field, "+") {
			column = field[1:]
		}
		sorts = append(sorts, SortField{Column: column, Order: order})
	}
	return sorts, nil
}
