package query

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoQueryBuilderImpl implements MongoQueryBuilder interface
type MongoQueryBuilderImpl struct {
	allowedFields      map[string]bool // Security: only allow specified fields
	allowedCollections map[string]bool // Security: only allow specified collections
	// Security settings
	enableSecurityChecks bool
	maxAllowedDocs       int
	// Query logging
	enableQueryLogging bool
	// Connection timeout settings
	queryTimeout int
}

// NewMongoQueryBuilder creates a new MongoDB query builder instance
func NewMongoQueryBuilder() *MongoQueryBuilderImpl {
	return &MongoQueryBuilderImpl{
		allowedFields:        make(map[string]bool),
		allowedCollections:   make(map[string]bool),
		enableSecurityChecks: true,
		maxAllowedDocs:       10000,
		enableQueryLogging:   true,
		queryTimeout:         30,
	}
}

// SetSecurityOptions configures security settings
func (mqb *MongoQueryBuilderImpl) SetSecurityOptions(enableChecks bool, maxDocs int) MongoQueryBuilder {
	mqb.enableSecurityChecks = enableChecks
	mqb.maxAllowedDocs = maxDocs
	return mqb
}

// SetAllowedFields sets the list of allowed fields for security
func (mqb *MongoQueryBuilderImpl) SetAllowedFields(fields []string) MongoQueryBuilder {
	mqb.allowedFields = make(map[string]bool)
	for _, field := range fields {
		mqb.allowedFields[field] = true
	}
	return mqb
}

// SetAllowedCollections sets the list of allowed collections for security
func (mqb *MongoQueryBuilderImpl) SetAllowedCollections(collections []string) MongoQueryBuilder {
	mqb.allowedCollections = make(map[string]bool)
	for _, collection := range collections {
		mqb.allowedCollections[collection] = true
	}
	return mqb
}

// SetQueryLogging enables or disables query logging
func (mqb *MongoQueryBuilderImpl) SetQueryLogging(enable bool) MongoQueryBuilder {
	mqb.enableQueryLogging = enable
	return mqb
}

// SetQueryTimeout sets the default query timeout
func (mqb *MongoQueryBuilderImpl) SetQueryTimeout(timeout time.Duration) MongoQueryBuilder {
	mqb.queryTimeout = int(timeout.Seconds())
	return mqb
}

// BuildFindQuery builds a MongoDB find query from DynamicQuery
func (mqb *MongoQueryBuilderImpl) BuildFindQuery(query DynamicQuery) (interface{}, interface{}, error) {
	filter := bson.M{}
	findOptions := options.Find()

	// Security check for limit
	if mqb.enableSecurityChecks && query.Limit > mqb.maxAllowedDocs {
		return nil, nil, fmt.Errorf("requested limit %d exceeds maximum allowed %d", query.Limit, mqb.maxAllowedDocs)
	}

	// Security check for collection name
	if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[query.From] {
		return nil, nil, fmt.Errorf("disallowed collection: %s", query.From)
	}

	// Build filter from DynamicQuery filters
	if len(query.Filters) > 0 {
		mongoFilter, err := mqb.buildFilter(query.Filters)
		if err != nil {
			return nil, nil, err
		}
		filter = mongoFilter
	}

	// Set projection from fields
	if len(query.Fields) > 0 {
		projection := bson.M{}
		for _, field := range query.Fields {
			if field.Expression == "*" {
				// Include all fields
				continue
			}
			fieldName := field.Expression
			if field.Alias != "" {
				fieldName = field.Alias
			}
			if mqb.allowedFields != nil && !mqb.allowedFields[fieldName] {
				return nil, nil, fmt.Errorf("disallowed field: %s", fieldName)
			}
			projection[fieldName] = 1
		}
		if len(projection) > 0 {
			findOptions.SetProjection(projection)
		}
	}

	// Set sort
	if len(query.Sort) > 0 {
		sort := bson.D{}
		for _, sortField := range query.Sort {
			fieldName := sortField.Column
			if mqb.allowedFields != nil && !mqb.allowedFields[fieldName] {
				return nil, nil, fmt.Errorf("disallowed field: %s", fieldName)
			}
			order := 1 // ASC
			if strings.ToUpper(sortField.Order) == "DESC" {
				order = -1 // DESC
			}
			sort = append(sort, bson.E{Key: fieldName, Value: order})
		}
		findOptions.SetSort(sort)
	}

	// Set limit and offset
	if query.Limit > 0 {
		findOptions.SetLimit(int64(query.Limit))
	}
	if query.Offset > 0 {
		findOptions.SetSkip(int64(query.Offset))
	}

	return filter, findOptions, nil
}

// BuildAggregateQuery builds a MongoDB aggregation pipeline from DynamicQuery
func (mqb *MongoQueryBuilderImpl) BuildAggregateQuery(query DynamicQuery) (interface{}, error) {
	pipeline := []bson.D{}

	// Security check for collection name
	if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[query.From] {
		return nil, fmt.Errorf("disallowed collection: %s", query.From)
	}

	// Handle CTEs as stages in the pipeline
	if len(query.CTEs) > 0 {
		for _, cte := range query.CTEs {
			// Security check for CTE collection
			if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[cte.Query.From] {
				return nil, fmt.Errorf("disallowed collection in CTE: %s", cte.Query.From)
			}

			subPipeline, err := mqb.BuildAggregateQuery(cte.Query)
			if err != nil {
				return nil, fmt.Errorf("failed to build CTE '%s': %w", cte.Name, err)
			}
			// Add $lookup stage for joins
			if len(cte.Query.Joins) > 0 {
				for _, join := range cte.Query.Joins {
					// Security check for joined collection
					if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[join.Table] {
						return nil, fmt.Errorf("disallowed collection in join: %s", join.Table)
					}

					lookupStage := bson.D{
						{Key: "$lookup", Value: bson.D{
							{Key: "from", Value: join.Table},
							{Key: "localField", Value: join.Alias},
							{Key: "foreignField", Value: "_id"},
							{Key: "as", Value: join.Alias},
						}},
					}
					pipeline = append(pipeline, lookupStage)
				}
			}
			// Add the sub-pipeline
			if subPipelineSlice, ok := subPipeline.([]bson.D); ok {
				pipeline = append(pipeline, subPipelineSlice...)
			}
		}
	}

	// Match stage for filters
	if len(query.Filters) > 0 {
		filter, err := mqb.buildFilter(query.Filters)
		if err != nil {
			return nil, err
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: filter}})
	}

	// Group stage for GROUP BY
	if len(query.GroupBy) > 0 {
		groupID := bson.D{}
		for _, field := range query.GroupBy {
			if mqb.allowedFields != nil && !mqb.allowedFields[field] {
				return nil, fmt.Errorf("disallowed field: %s", field)
			}
			groupID = append(groupID, bson.E{Key: field, Value: "$" + field})
		}

		groupStage := bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: groupID},
			}},
		}

		// Add any aggregations from fields
		for _, field := range query.Fields {
			if strings.Contains(field.Expression, "(") && strings.Contains(field.Expression, ")") {
				// This is an aggregation function
				funcName := strings.Split(field.Expression, "(")[0]
				funcField := strings.TrimSuffix(strings.Split(field.Expression, "(")[1], ")")

				if mqb.allowedFields != nil && !mqb.allowedFields[funcField] {
					return nil, fmt.Errorf("disallowed field: %s", funcField)
				}

				switch strings.ToLower(funcName) {
				case "count":
					groupStage = append(groupStage, bson.E{
						Key: field.Alias, Value: bson.D{{Key: "$sum", Value: 1}},
					})
				case "sum":
					groupStage = append(groupStage, bson.E{
						Key: field.Alias, Value: bson.D{{Key: "$sum", Value: "$" + funcField}},
					})
				case "avg":
					groupStage = append(groupStage, bson.E{
						Key: field.Alias, Value: bson.D{{Key: "$avg", Value: "$" + funcField}},
					})
				case "min":
					groupStage = append(groupStage, bson.E{
						Key: field.Alias, Value: bson.D{{Key: "$min", Value: "$" + funcField}},
					})
				case "max":
					groupStage = append(groupStage, bson.E{
						Key: field.Alias, Value: bson.D{{Key: "$max", Value: "$" + funcField}},
					})
				}
			}
		}

		pipeline = append(pipeline, groupStage)
	}

	// Sort stage
	if len(query.Sort) > 0 {
		sort := bson.D{}
		for _, sortField := range query.Sort {
			fieldName := sortField.Column
			if mqb.allowedFields != nil && !mqb.allowedFields[fieldName] {
				return nil, fmt.Errorf("disallowed field: %s", fieldName)
			}
			order := 1 // ASC
			if strings.ToUpper(sortField.Order) == "DESC" {
				order = -1 // DESC
			}
			sort = append(sort, bson.E{Key: fieldName, Value: order})
		}
		pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sort}})
	}

	// Skip and limit stages
	if query.Offset > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: query.Offset}})
	}
	if query.Limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: query.Limit}})
	}

	return pipeline, nil
}

// buildFilter builds a MongoDB filter from FilterGroups
func (mqb *MongoQueryBuilderImpl) buildFilter(filterGroups []FilterGroup) (bson.M, error) {
	if len(filterGroups) == 0 {
		return bson.M{}, nil
	}

	var result bson.M
	var err error

	for i, group := range filterGroups {
		if len(group.Filters) == 0 {
			continue
		}

		groupFilter, err := mqb.buildFilterGroup(group)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			result = groupFilter
		} else {
			logicOp := "$and"
			if group.LogicOp != "" {
				switch strings.ToUpper(group.LogicOp) {
				case "OR":
					logicOp = "$or"
				}
			}
			result = bson.M{logicOp: []bson.M{result, groupFilter}}
		}
	}

	return result, err
}

// buildFilterGroup builds a filter for a single filter group
func (mqb *MongoQueryBuilderImpl) buildFilterGroup(group FilterGroup) (bson.M, error) {
	var filters []bson.M
	logicOp := "$and"
	if group.LogicOp != "" {
		switch strings.ToUpper(group.LogicOp) {
		case "OR":
			logicOp = "$or"
		}
	}

	for _, filter := range group.Filters {
		fieldFilter, err := mqb.buildFilterCondition(filter)
		if err != nil {
			return nil, err
		}
		filters = append(filters, fieldFilter)
	}

	if len(filters) == 1 {
		return filters[0], nil
	}
	return bson.M{logicOp: filters}, nil
}

// buildFilterCondition builds a single filter condition for MongoDB
func (mqb *MongoQueryBuilderImpl) buildFilterCondition(filter DynamicFilter) (bson.M, error) {
	field := filter.Column
	if mqb.allowedFields != nil && !mqb.allowedFields[field] {
		return nil, fmt.Errorf("disallowed field: %s", field)
	}

	switch filter.Operator {
	case OpEqual:
		return bson.M{field: filter.Value}, nil
	case OpNotEqual:
		return bson.M{field: bson.M{"$ne": filter.Value}}, nil
	case OpIn:
		values := mqb.parseArrayValue(filter.Value)
		return bson.M{field: bson.M{"$in": values}}, nil
	case OpNotIn:
		values := mqb.parseArrayValue(filter.Value)
		return bson.M{field: bson.M{"$nin": values}}, nil
	case OpGreaterThan:
		return bson.M{field: bson.M{"$gt": filter.Value}}, nil
	case OpGreaterThanEqual:
		return bson.M{field: bson.M{"$gte": filter.Value}}, nil
	case OpLessThan:
		return bson.M{field: bson.M{"$lt": filter.Value}}, nil
	case OpLessThanEqual:
		return bson.M{field: bson.M{"$lte": filter.Value}}, nil
	case OpLike:
		// Convert SQL LIKE to MongoDB regex
		pattern := filter.Value.(string)
		pattern = strings.ReplaceAll(pattern, "%", ".*")
		pattern = strings.ReplaceAll(pattern, "_", ".")
		return bson.M{field: bson.M{"$regex": pattern, "$options": "i"}}, nil
	case OpILike:
		// Case-insensitive like
		pattern := filter.Value.(string)
		pattern = strings.ReplaceAll(pattern, "%", ".*")
		pattern = strings.ReplaceAll(pattern, "_", ".")
		return bson.M{field: bson.M{"$regex": pattern, "$options": "i"}}, nil
	case OpContains:
		// Contains substring
		pattern := filter.Value.(string)
		return bson.M{field: bson.M{"$regex": pattern, "$options": "i"}}, nil
	case OpNotContains:
		// Does not contain substring
		pattern := filter.Value.(string)
		return bson.M{field: bson.M{"$not": bson.M{"$regex": pattern, "$options": "i"}}}, nil
	case OpStartsWith:
		// Starts with
		pattern := filter.Value.(string)
		return bson.M{field: bson.M{"$regex": "^" + pattern, "$options": "i"}}, nil
	case OpEndsWith:
		// Ends with
		pattern := filter.Value.(string)
		return bson.M{field: bson.M{"$regex": pattern + "$", "$options": "i"}}, nil
	case OpNull:
		return bson.M{field: bson.M{"$exists": false}}, nil
	case OpNotNull:
		return bson.M{field: bson.M{"$exists": true}}, nil
	case OpJsonContains:
		// JSON contains
		return bson.M{field: bson.M{"$elemMatch": filter.Value}}, nil
	case OpJsonNotContains:
		// JSON does not contain
		return bson.M{field: bson.M{"$not": bson.M{"$elemMatch": filter.Value}}}, nil
	case OpJsonExists:
		// JSON path exists
		return bson.M{field + "." + filter.Options["path"].(string): bson.M{"$exists": true}}, nil
	case OpJsonNotExists:
		// JSON path does not exist
		return bson.M{field + "." + filter.Options["path"].(string): bson.M{"$exists": false}}, nil
	case OpArrayContains:
		// Array contains
		return bson.M{field: bson.M{"$elemMatch": bson.M{"$eq": filter.Value}}}, nil
	case OpArrayNotContains:
		// Array does not contain
		return bson.M{field: bson.M{"$not": bson.M{"$elemMatch": bson.M{"$eq": filter.Value}}}}, nil
	case OpArrayLength:
		// Array length
		if lengthOption, ok := filter.Options["length"].(int); ok {
			return bson.M{field: bson.M{"$size": lengthOption}}, nil
		}
		return nil, fmt.Errorf("array_length operator requires 'length' option")
	default:
		return nil, fmt.Errorf("unsupported operator: %s", filter.Operator)
	}
}

// parseArrayValue parses an array value for MongoDB
func (mqb *MongoQueryBuilderImpl) parseArrayValue(value interface{}) []interface{} {
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
