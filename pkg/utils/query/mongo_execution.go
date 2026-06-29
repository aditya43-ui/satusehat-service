package query

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ExecuteFind executes a MongoDB find query
func (mqb *MongoQueryBuilderImpl) ExecuteFind(ctx context.Context, collection *mongo.Collection, query DynamicQuery, dest interface{}) error {
	// Security check for collection name
	if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[collection.Name()] {
		return fmt.Errorf("disallowed collection: %s", collection.Name())
	}

	filter, findOptions, err := mqb.BuildFindQuery(query)
	if err != nil {
		return err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && mqb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(mqb.queryTimeout)*time.Second)
		defer cancel()
	}

	start := time.Now()
	cursor, err := collection.Find(ctx, filter.(bson.M), findOptions.(*options.FindOptions))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, dest)
	if mqb.enableQueryLogging {
		fmt.Printf("[DEBUG] MongoDB Find executed in %v\n", time.Since(start))
	}
	return err
}

// ExecuteAggregate executes a MongoDB aggregation pipeline
func (mqb *MongoQueryBuilderImpl) ExecuteAggregate(ctx context.Context, collection *mongo.Collection, query DynamicQuery, dest interface{}) error {
	// Security check for collection name
	if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[collection.Name()] {
		return fmt.Errorf("disallowed collection: %s", collection.Name())
	}

	pipeline, err := mqb.BuildAggregateQuery(query)
	if err != nil {
		return err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && mqb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(mqb.queryTimeout)*time.Second)
		defer cancel()
	}

	start := time.Now()
	cursor, err := collection.Aggregate(ctx, pipeline.([]bson.D))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, dest)
	if mqb.enableQueryLogging {
		fmt.Printf("[DEBUG] MongoDB Aggregate executed in %v\n", time.Since(start))
	}
	return err
}

// ExecuteCount executes a MongoDB count query
func (mqb *MongoQueryBuilderImpl) ExecuteCount(ctx context.Context, collection *mongo.Collection, query DynamicQuery) (int64, error) {
	// Security check for collection name
	if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[collection.Name()] {
		return 0, fmt.Errorf("disallowed collection: %s", collection.Name())
	}

	filter, _, err := mqb.BuildFindQuery(query)
	if err != nil {
		return 0, err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && mqb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(mqb.queryTimeout)*time.Second)
		defer cancel()
	}

	start := time.Now()
	count, err := collection.CountDocuments(ctx, filter.(bson.M))
	if mqb.enableQueryLogging {
		fmt.Printf("[DEBUG] MongoDB Count executed in %v\n", time.Since(start))
	}
	return count, err
}

// ExecuteInsert executes a MongoDB insert operation
func (mqb *MongoQueryBuilderImpl) ExecuteInsert(ctx context.Context, collection *mongo.Collection, data InsertData) (*mongo.InsertOneResult, error) {
	// Security check for collection name
	if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[collection.Name()] {
		return nil, fmt.Errorf("disallowed collection: %s", collection.Name())
	}

	document := bson.M{}
	for i, col := range data.Columns {
		if mqb.allowedFields != nil && !mqb.allowedFields[col] {
			return nil, fmt.Errorf("disallowed field: %s", col)
		}
		document[col] = data.Values[i]
	}

	// Handle JSON values
	for col, val := range data.JsonValues {
		if mqb.allowedFields != nil && !mqb.allowedFields[col] {
			return nil, fmt.Errorf("disallowed field: %s", col)
		}
		document[col] = val
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && mqb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(mqb.queryTimeout)*time.Second)
		defer cancel()
	}

	start := time.Now()
	result, err := collection.InsertOne(ctx, document)
	if mqb.enableQueryLogging {
		fmt.Printf("[DEBUG] MongoDB Insert executed in %v\n", time.Since(start))
	}
	return result, err
}

// ExecuteUpdate executes a MongoDB update operation
func (mqb *MongoQueryBuilderImpl) ExecuteUpdate(ctx context.Context, collection *mongo.Collection, updateData UpdateData, filters []FilterGroup) (*mongo.UpdateResult, error) {
	// Security check for collection name
	if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[collection.Name()] {
		return nil, fmt.Errorf("disallowed collection: %s", collection.Name())
	}

	filter, err := mqb.buildFilter(filters)
	if err != nil {
		return nil, err
	}

	update := bson.M{"$set": bson.M{}}
	for i, col := range updateData.Columns {
		if mqb.allowedFields != nil && !mqb.allowedFields[col] {
			return nil, fmt.Errorf("disallowed field: %s", col)
		}
		update["$set"].(bson.M)[col] = updateData.Values[i]
	}

	// Handle JSON updates
	for col, jsonUpdate := range updateData.JsonUpdates {
		if mqb.allowedFields != nil && !mqb.allowedFields[col] {
			return nil, fmt.Errorf("disallowed field: %s", col)
		}
		// Use dot notation for nested JSON updates
		update["$set"].(bson.M)[col+"."+jsonUpdate.Path] = jsonUpdate.Value
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && mqb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(mqb.queryTimeout)*time.Second)
		defer cancel()
	}

	start := time.Now()
	result, err := collection.UpdateMany(ctx, filter, update)
	if mqb.enableQueryLogging {
		fmt.Printf("[DEBUG] MongoDB Update executed in %v\n", time.Since(start))
	}
	return result, err
}

// ExecuteDelete executes a MongoDB delete operation
func (mqb *MongoQueryBuilderImpl) ExecuteDelete(ctx context.Context, collection *mongo.Collection, filters []FilterGroup) (*mongo.DeleteResult, error) {
	// Security check for collection name
	if mqb.enableSecurityChecks && len(mqb.allowedCollections) > 0 && !mqb.allowedCollections[collection.Name()] {
		return nil, fmt.Errorf("disallowed collection: %s", collection.Name())
	}

	filter, err := mqb.buildFilter(filters)
	if err != nil {
		return nil, err
	}

	// Set timeout if not already in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && mqb.queryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(mqb.queryTimeout)*time.Second)
		defer cancel()
	}

	start := time.Now()
	result, err := collection.DeleteMany(ctx, filter)
	if mqb.enableQueryLogging {
		fmt.Printf("[DEBUG] MongoDB Delete executed in %v\n", time.Since(start))
	}
	return result, err
}
