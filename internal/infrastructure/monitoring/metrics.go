// File: /home/meninjar/goprint/service/internal/infrastructure/monitoring/metrics.go
package monitoring

import (
	"time"

	"service/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics defines all Prometheus metrics for the GoPrint service
type Metrics struct {
	// HTTP metrics
	HttpRequestsTotal    *prometheus.CounterVec
	HttpRequestDuration  *prometheus.HistogramVec
	HttpResponseSize     *prometheus.HistogramVec
	HttpRequestsInFlight prometheus.Gauge

	// Database metrics
	DbConnections       prometheus.Gauge
	DbQueryDuration     *prometheus.HistogramVec
	DbQueryErrors       *prometheus.CounterVec
	DbActiveConnections *prometheus.GaugeVec

	// Cache metrics
	CacheHits       prometheus.Counter
	CacheMisses     prometheus.Counter
	CacheOperations *prometheus.CounterVec
	CacheDuration   *prometheus.HistogramVec

	// Business logic metrics
	EthnicOperations *prometheus.CounterVec
	AuthOperations   *prometheus.CounterVec
	PersonOperations *prometheus.CounterVec

	// Service health metrics
	ServiceUptime         prometheus.Counter
	ServiceHealthStatus   prometheus.Gauge
	ExternalServiceStatus *prometheus.GaugeVec

	// Resource usage metrics
	MemoryUsage prometheus.Gauge
	CPUUsage    prometheus.Gauge
	Goroutines  prometheus.Gauge

	// Logger untuk monitoring
	logger      logger.Logger
	serviceName string
	startTime   time.Time
}

// NewMetrics creates a new instance of Prometheus metrics
func NewMetrics(serviceName string) *Metrics {
	log := logger.Default().WithFields(
		logger.String("component", "monitoring"),
		logger.String("service", serviceName),
	)

	labels := []string{"method", "endpoint", "status"}
	dbLabels := []string{"operation", "table", "database"}
	cacheLabels := []string{"operation", "result"}
	businessLabels := []string{"operation", "status", "entity"}
	externalLabels := []string{"service", "status"}

	metrics := &Metrics{
		serviceName: serviceName,
		logger:      log,
		startTime:   time.Now(),

		// HTTP metrics
		HttpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_requests_total",
				Help:        "Total number of HTTP requests",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			labels,
		),
		HttpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "http_request_duration_seconds",
				Help:        "HTTP request duration in seconds",
				ConstLabels: prometheus.Labels{"service": serviceName},
				Buckets:     prometheus.DefBuckets,
			},
			labels,
		),
		HttpResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "http_response_size_bytes",
				Help:        "HTTP response size in bytes",
				ConstLabels: prometheus.Labels{"service": serviceName},
				Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
			},
			labels,
		),
		HttpRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name:        "http_requests_in_flight",
				Help:        "Number of HTTP requests currently being processed",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),

		// Database metrics
		DbConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name:        "db_connections_active",
				Help:        "Number of active database connections",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
		DbActiveConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "db_connections_active_by_db",
				Help:        "Number of active database connections by database",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			[]string{"database"},
		),
		DbQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "db_query_duration_seconds",
				Help:        "Database query duration in seconds",
				ConstLabels: prometheus.Labels{"service": serviceName},
				Buckets:     prometheus.DefBuckets,
			},
			dbLabels,
		),
		DbQueryErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "db_query_errors_total",
				Help:        "Total number of database query errors",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			dbLabels,
		),

		// Cache metrics
		CacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Name:        "cache_hits_total",
				Help:        "Total number of cache hits",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
		CacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Name:        "cache_misses_total",
				Help:        "Total number of cache misses",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
		CacheOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "cache_operations_total",
				Help:        "Total number of cache operations",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			cacheLabels,
		),
		CacheDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        "cache_operation_duration_seconds",
				Help:        "Cache operation duration in seconds",
				ConstLabels: prometheus.Labels{"service": serviceName},
				Buckets:     prometheus.DefBuckets,
			},
			cacheLabels,
		),

		// Business logic metrics
		EthnicOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "ethnic_operations_total",
				Help:        "Total number of ethnic operations",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			businessLabels,
		),
		AuthOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "auth_operations_total",
				Help:        "Total number of authentication operations",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			businessLabels,
		),
		PersonOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "person_operations_total",
				Help:        "Total number of person operations",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			businessLabels,
		),

		// Service health metrics
		ServiceUptime: promauto.NewCounter(
			prometheus.CounterOpts{
				Name:        "service_uptime_seconds",
				Help:        "Service uptime in seconds",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
		ServiceHealthStatus: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name:        "service_health_status",
				Help:        "Service health status (1 = healthy, 0 = unhealthy)",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
		ExternalServiceStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "external_service_status",
				Help:        "External service status (1 = up, 0 = down)",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
			externalLabels,
		),

		// Resource usage metrics
		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name:        "memory_usage_bytes",
				Help:        "Memory usage in bytes",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name:        "cpu_usage_percent",
				Help:        "CPU usage percentage",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
		Goroutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name:        "goroutines_count",
				Help:        "Number of goroutines",
				ConstLabels: prometheus.Labels{"service": serviceName},
			},
		),
	}

	log.Info("Prometheus metrics initialized successfully",
		logger.String("service", serviceName),
	)

	return metrics
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(method, endpoint, status string, duration time.Duration, responseSize int64) {
	labels := prometheus.Labels{
		"method":   method,
		"endpoint": endpoint,
		"status":   status,
	}

	m.HttpRequestsTotal.With(labels).Inc()
	m.HttpRequestDuration.With(labels).Observe(duration.Seconds())
	m.HttpResponseSize.With(labels).Observe(float64(responseSize))

	m.logger.Debug("HTTP request recorded",
		logger.String("method", method),
		logger.String("endpoint", endpoint),
		logger.String("status", status),
		logger.Float64("duration_seconds", duration.Seconds()),
		logger.Int64("response_size", responseSize),
	)
}

// RecordDBQuery records database query metrics
func (m *Metrics) RecordDBQuery(operation, table, database string, duration time.Duration, err error) {
	labels := prometheus.Labels{
		"operation": operation,
		"table":     table,
		"database":  database,
	}

	m.DbQueryDuration.With(labels).Observe(duration.Seconds())
	if err != nil {
		m.DbQueryErrors.With(labels).Inc()
		m.logger.Error("Database query error recorded",
			logger.String("operation", operation),
			logger.String("table", table),
			logger.String("database", database),
			logger.ErrorField(err),
		)
	} else {
		m.logger.Debug("Database query recorded",
			logger.String("operation", operation),
			logger.String("table", table),
			logger.String("database", database),
			logger.Float64("duration_seconds", duration.Seconds()),
		)
	}
}

// RecordCacheOperation records cache operation metrics
func (m *Metrics) RecordCacheOperation(operation string, hit bool, duration time.Duration) {
	if hit {
		m.CacheHits.Inc()
	} else {
		m.CacheMisses.Inc()
	}

	result := "hit"
	if !hit {
		result = "miss"
	}

	labels := prometheus.Labels{
		"operation": operation,
		"result":    result,
	}

	m.CacheOperations.With(labels).Inc()
	m.CacheDuration.With(labels).Observe(duration.Seconds())

	m.logger.Debug("Cache operation recorded",
		logger.String("operation", operation),
		logger.String("result", result),
		logger.Float64("duration_seconds", duration.Seconds()),
	)
}

// RecordBusinessOperation records business logic operation metrics
func (m *Metrics) RecordBusinessOperation(operationType, operation, status string) {
	var counter *prometheus.CounterVec

	switch operationType {
	case "ethnic":
		counter = m.EthnicOperations
	case "auth":
		counter = m.AuthOperations
	case "person":
		counter = m.PersonOperations
	default:
		m.logger.Warn("Unknown business operation type",
			logger.String("type", operationType),
			logger.String("operation", operation),
		)
		return
	}

	labels := prometheus.Labels{
		"operation": operation,
		"status":    status,
		"entity":    operationType,
	}

	counter.With(labels).Inc()

	m.logger.Debug("Business operation recorded",
		logger.String("type", operationType),
		logger.String("operation", operation),
		logger.String("status", status),
	)
}

// UpdateServiceHealth updates service health status
func (m *Metrics) UpdateServiceHealth(healthy bool) {
	if healthy {
		m.ServiceHealthStatus.Set(1)
		m.logger.Info("Service health status updated to healthy")
	} else {
		m.ServiceHealthStatus.Set(0)
		m.logger.Error("Service health status updated to unhealthy")
	}
}

// UpdateExternalServiceStatus updates external service status
func (m *Metrics) UpdateExternalServiceStatus(service string, status bool) {
	value := 0.0
	statusStr := "down"
	if status {
		value = 1.0
		statusStr = "up"
	}

	labels := prometheus.Labels{
		"service": service,
		"status":  statusStr,
	}

	m.ExternalServiceStatus.With(labels).Set(value)

	m.logger.Info("External service status updated",
		logger.String("service", service),
		logger.String("status", statusStr),
	)
}

// UpdateResourceUsage updates resource usage metrics
func (m *Metrics) UpdateResourceUsage(memory, cpu float64, goroutines int) {
	m.MemoryUsage.Set(memory)
	m.CPUUsage.Set(cpu)
	m.Goroutines.Set(float64(goroutines))

	m.logger.Debug("Resource usage updated",
		logger.Float64("memory_mb", memory),
		logger.Float64("cpu_percent", cpu),
		logger.Int("goroutines", goroutines),
	)
}

// IncrementUptime increments service uptime
func (m *Metrics) IncrementUptime(seconds float64) {
	m.ServiceUptime.Add(seconds)
}

// GetUptime returns current uptime in seconds
func (m *Metrics) GetUptime() float64 {
	return time.Since(m.startTime).Seconds()
}
