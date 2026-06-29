// File: /home/meninjar/goprint/service/internal/infrastructure/monitoring/middleware.go
package monitoring

import (
	"strconv"
	"time"

	"service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MetricsMiddleware creates Gin middleware for collecting HTTP metrics
func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		// Add request ID to context for logging
		log := logger.Default().WithFields(
			logger.String("request_id", requestID),
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("client_ip", c.ClientIP()),
		)

		// Start timer
		start := time.Now()

		// Increment requests in flight
		metrics.HttpRequestsInFlight.Inc()
		defer metrics.HttpRequestsInFlight.Dec()

		// Log request start
		log.Info("HTTP request started")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get response size
		responseSize := c.Writer.Size()
		if responseSize < 0 {
			responseSize = 0
		}

		// Get status code
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		metrics.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			status,
			duration,
			int64(responseSize),
		)

		// Log request completion
		log.Info("HTTP request completed",
			logger.String("status", status),
			logger.Float64("duration_seconds", duration.Seconds()),
			logger.Int("response_size", responseSize),
		)

		// Add request ID to response
		c.Header("X-Request-ID", requestID)
	}
}

// DatabaseMetricsMiddleware creates middleware for database metrics
func DatabaseMetricsMiddleware(metrics *Metrics) func(operation, table, database string) func(error) {
	return func(operation, table, database string) func(error) {
		start := time.Now()
		return func(err error) {
			duration := time.Since(start)
			metrics.RecordDBQuery(operation, table, database, duration, err)
		}
	}
}

// CacheMetricsMiddleware creates middleware for cache metrics
func CacheMetricsMiddleware(metrics *Metrics) func(operation string, hit bool) func() {
	return func(operation string, hit bool) func() {
		start := time.Now()
		return func() {
			duration := time.Since(start)
			metrics.RecordCacheOperation(operation, hit, duration)
		}
	}
}
