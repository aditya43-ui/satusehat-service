// File: /home/meninjar/goprint/service/internal/infrastructure/monitoring/repository_wrapper.go
package monitoring

import (
	"context"
	"time"

	"service/pkg/logger"

	"gorm.io/gorm"
)

// GormDBWrapper wraps GORM DB with metrics collection
type GormDBWrapper struct {
	*gorm.DB
	metrics  *Metrics
	database string
	logger   logger.Logger
}

// NewGormDBWrapper creates a new GORM wrapper with metrics
func NewGormDBWrapper(db *gorm.DB, metrics *Metrics, database string) *GormDBWrapper {
	log := logger.Default().WithFields(
		logger.String("component", "gorm_wrapper"),
		logger.String("database", database),
	)

	return &GormDBWrapper{
		DB:       db,
		metrics:  metrics,
		database: database,
		logger:   log,
	}
}

// Create overrides GORM Create with metrics
func (w *GormDBWrapper) Create(value interface{}) *gorm.DB {
	start := time.Now()
	result := w.DB.Create(value)
	duration := time.Since(start)

	table := w.DB.Statement.Table
	status := "success"
	if result.Error != nil {
		status = "error"
	}

	w.metrics.RecordDBQuery("create", table, w.database, duration, result.Error)

	w.logger.Debug("Database create operation",
		logger.String("table", table),
		logger.String("status", status),
		logger.Float64("duration_seconds", duration.Seconds()),
		logger.ErrorField(result.Error),
	)

	return result
}

// Find overrides GORM Find with metrics
func (w *GormDBWrapper) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	start := time.Now()
	result := w.DB.Find(dest, conds...)
	duration := time.Since(start)

	table := w.DB.Statement.Table
	status := "success"
	if result.Error != nil {
		status = "error"
	}

	w.metrics.RecordDBQuery("find", table, w.database, duration, result.Error)

	w.logger.Debug("Database find operation",
		logger.String("table", table),
		logger.String("status", status),
		logger.Float64("duration_seconds", duration.Seconds()),
		logger.ErrorField(result.Error),
	)

	return result
}

// First overrides GORM First with metrics
func (w *GormDBWrapper) First(dest interface{}, conds ...interface{}) *gorm.DB {
	start := time.Now()
	result := w.DB.First(dest, conds...)
	duration := time.Since(start)

	table := w.DB.Statement.Table
	status := "success"
	if result.Error != nil {
		status = "error"
	}

	w.metrics.RecordDBQuery("first", table, w.database, duration, result.Error)

	w.logger.Debug("Database first operation",
		logger.String("table", table),
		logger.String("status", status),
		logger.Float64("duration_seconds", duration.Seconds()),
		logger.ErrorField(result.Error),
	)

	return result
}

// Update overrides GORM Update with metrics
func (w *GormDBWrapper) Update(column string, value interface{}) *gorm.DB {
	start := time.Now()
	result := w.DB.Update(column, value)
	duration := time.Since(start)

	table := w.DB.Statement.Table
	status := "success"
	if result.Error != nil {
		status = "error"
	}

	w.metrics.RecordDBQuery("update", table, w.database, duration, result.Error)

	w.logger.Debug("Database update operation",
		logger.String("table", table),
		logger.String("status", status),
		logger.Float64("duration_seconds", duration.Seconds()),
		logger.ErrorField(result.Error),
	)

	return result
}

// Delete overrides GORM Delete with metrics
func (w *GormDBWrapper) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	start := time.Now()
	result := w.DB.Delete(value, conds...)
	duration := time.Since(start)

	table := w.DB.Statement.Table
	status := "success"
	if result.Error != nil {
		status = "error"
	}

	w.metrics.RecordDBQuery("delete", table, w.database, duration, result.Error)

	w.logger.Debug("Database delete operation",
		logger.String("table", table),
		logger.String("status", status),
		logger.Float64("duration_seconds", duration.Seconds()),
		logger.ErrorField(result.Error),
	)

	return result
}

// WithContext overrides GORM WithContext to maintain wrapper
func (w *GormDBWrapper) WithContext(ctx context.Context) *GormDBWrapper {
	return &GormDBWrapper{
		DB:       w.DB.WithContext(ctx),
		metrics:  w.metrics,
		database: w.database,
		logger:   w.logger,
	}
}

// Model overrides GORM Model to maintain wrapper
func (w *GormDBWrapper) Model(value interface{}) *GormDBWrapper {
	return &GormDBWrapper{
		DB:       w.DB.Model(value),
		metrics:  w.metrics,
		database: w.database,
		logger:   w.logger,
	}
}

// Table overrides GORM Table to maintain wrapper
func (w *GormDBWrapper) Table(name string, args ...interface{}) *GormDBWrapper {
	return &GormDBWrapper{
		DB:       w.DB.Table(name, args...),
		metrics:  w.metrics,
		database: w.database,
		logger:   w.logger,
	}
}
