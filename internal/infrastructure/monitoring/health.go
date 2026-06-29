// File: /home/meninjar/goprint/service/internal/infrastructure/monitoring/health.go
package monitoring

import (
	"context"
	"runtime"
	"time"

	"service/internal/infrastructure/cache"
	"service/internal/infrastructure/config"
	"service/internal/infrastructure/database"
	"service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// HealthMonitor handles health monitoring and metrics collection
type HealthMonitor struct {
	metrics      *Metrics
	dbService    database.Service
	cacheManager *cache.Manager
	config       *config.Config
	startTime    time.Time
	logger       logger.Logger
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(metrics *Metrics, dbService database.Service, cacheManager *cache.Manager, config *config.Config) *HealthMonitor {
	log := logger.Default().WithFields(
		logger.String("component", "health_monitor"),
		logger.String("service", "goprint"),
	)

	return &HealthMonitor{
		metrics:      metrics,
		dbService:    dbService,
		cacheManager: cacheManager,
		config:       config,
		startTime:    time.Now(),
		logger:       log,
	}
}

// Start starts the health monitoring routine
func (h *HealthMonitor) Start(ctx context.Context) {
	log := h.logger.WithFields(logger.String("action", "start_health_monitoring"))
	log.Info("Starting health monitoring service")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial health check
	h.performHealthCheck()

	for {
		select {
		case <-ticker.C:
			h.performHealthCheck()
		case <-ctx.Done():
			log.Info("Health monitoring service stopped")
			return
		}
	}
}

// performHealthCheck performs comprehensive health check
func (h *HealthMonitor) performHealthCheck() {
	log := h.logger.WithFields(logger.String("action", "health_check"))

	// Check all databases
	dbHealthy := h.checkAllDatabases()

	// Check cache
	cacheHealthy := h.checkCache()

	// Check external services
	externalHealthy := h.checkExternalServices()

	// Update service health
	overallHealthy := dbHealthy && cacheHealthy
	h.metrics.UpdateServiceHealth(overallHealthy)

	// Update resource usage
	h.updateResourceUsage()

	// Update uptime
	uptime := time.Since(h.startTime).Seconds()
	h.metrics.IncrementUptime(uptime)

	log.Info("Health check completed",
		logger.Bool("database_healthy", dbHealthy),
		logger.Bool("cache_healthy", cacheHealthy),
		logger.Bool("external_healthy", externalHealthy),
		logger.Bool("overall_healthy", overallHealthy),
		logger.Float64("uptime_seconds", uptime),
	)
}

// checkAllDatabases checks all configured databases
func (h *HealthMonitor) checkAllDatabases() bool {
	log := h.logger.WithFields(logger.String("action", "check_databases"))

	dbList := h.dbService.ListDBs()
	if len(dbList) == 0 {
		log.Warn("No databases configured")
		return false
	}

	allHealthy := true
	for _, dbName := range dbList {
		if !h.checkDatabase(dbName) {
			allHealthy = false
		}
	}

	return allHealthy
}

// checkDatabase checks specific database connectivity
func (h *HealthMonitor) checkDatabase(dbName string) bool {
	log := h.logger.WithFields(
		logger.String("action", "check_database"),
		logger.String("database", dbName),
	)

	// Get database type
	dbType, err := h.dbService.GetDBType(dbName)
	if err != nil {
		log.Error("Failed to get database type", logger.ErrorField(err))
		h.metrics.UpdateExternalServiceStatus(dbName, false)
		return false
	}

	// Check based on database type
	switch dbType {
	case database.Postgres, database.MySQL, database.SQLServer, database.SQLite:
		return h.checkSQLDatabase(dbName, log)
	case database.MongoDB:
		return h.checkMongoDatabase(dbName, log)
	default:
		log.Error("Unknown database type", logger.String("type", string(dbType)))
		return false
	}
}

// checkSQLDatabase checks SQL database connectivity
func (h *HealthMonitor) checkSQLDatabase(dbName string, log logger.Logger) bool {
	// Get GORM DB
	gormDB, err := h.dbService.GetGormDB(dbName)
	if err != nil {
		log.Error("Failed to get GORM database", logger.ErrorField(err))
		h.metrics.UpdateExternalServiceStatus(dbName, false)
		return false
	}

	// Get SQL DB from GORM
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error("Failed to get SQL database from GORM", logger.ErrorField(err))
		h.metrics.UpdateExternalServiceStatus(dbName, false)
		return false
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sqlDB.PingContext(ctx)
	if err != nil {
		log.Error("Database ping failed", logger.ErrorField(err))
		h.metrics.UpdateExternalServiceStatus(dbName, false)
		h.metrics.DbActiveConnections.With(prometheus.Labels{"database": dbName}).Set(0)
		return false
	}

	// Get connection stats
	stats := sqlDB.Stats()
	h.metrics.DbActiveConnections.With(prometheus.Labels{"database": dbName}).Set(float64(stats.OpenConnections))

	log.Info("Database connection healthy",
		logger.Int("open_connections", stats.OpenConnections),
		logger.Int("idle_connections", stats.Idle),
		logger.Int("in_use_connections", stats.InUse),
	)

	h.metrics.UpdateExternalServiceStatus(dbName, true)
	return true
}

// checkMongoDatabase checks MongoDB connectivity
func (h *HealthMonitor) checkMongoDatabase(dbName string, log logger.Logger) bool {
	client, err := h.dbService.GetMongoClient(dbName)
	if err != nil {
		log.Error("Failed to get MongoDB client", logger.ErrorField(err))
		h.metrics.UpdateExternalServiceStatus(dbName, false)
		return false
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Error("MongoDB ping failed", logger.ErrorField(err))
		h.metrics.UpdateExternalServiceStatus(dbName, false)
		return false
	}

	log.Info("MongoDB connection healthy")
	h.metrics.UpdateExternalServiceStatus(dbName, true)
	return true
}

// checkCache checks cache connectivity
func (h *HealthMonitor) checkCache() bool {
	log := h.logger.WithFields(logger.String("action", "check_cache"))

	if h.cacheManager == nil {
		log.Warn("Cache manager not available")
		return true // Cache is optional
	}

	// Test cache health directly using manager
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.cacheManager.Health(ctx)
	if err != nil {
		log.Error("Cache health check failed", logger.ErrorField(err))
		h.metrics.UpdateExternalServiceStatus("redis", false)
		return false
	}

	log.Info("Cache connection healthy")
	h.metrics.UpdateExternalServiceStatus("redis", true)
	return true
}

// checkExternalServices checks external service connectivity
func (h *HealthMonitor) checkExternalServices() bool {
	allHealthy := true

	// Check BPJS service if configured
	if h.config != nil && h.config.Bpjs.BaseURL != "" {
		bpjsHealthy := h.checkBPJSService()
		if !bpjsHealthy {
			allHealthy = false
		}
	}

	// Check SatuSehat service if configured
	if h.config != nil && h.config.SatuSehat.BaseURL != "" {
		satuSehatHealthy := h.checkSatuSehatService()
		if !satuSehatHealthy {
			allHealthy = false
		}
	}

	return allHealthy
}

// checkBPJSService checks BPJS service connectivity
func (h *HealthMonitor) checkBPJSService() bool {
	log := h.logger.WithFields(
		logger.String("action", "check_bpjs"),
		logger.String("service", "bpjs"),
	)

	// Implement BPJS health check logic here
	// For now, return true as placeholder
	log.Info("BPJS service health check (placeholder)")
	h.metrics.UpdateExternalServiceStatus("bpjs", true)
	return true
}

// checkSatuSehatService checks SatuSehat service connectivity
func (h *HealthMonitor) checkSatuSehatService() bool {
	log := h.logger.WithFields(
		logger.String("action", "check_satu_sehat"),
		logger.String("service", "satu_sehat"),
	)

	// Implement SatuSehat health check logic here
	// For now, return true as placeholder
	log.Info("SatuSehat service health check (placeholder)")
	h.metrics.UpdateExternalServiceStatus("satu_sehat", true)
	return true
}

// updateResourceUsage updates resource usage metrics
func (h *HealthMonitor) updateResourceUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Memory usage in MB
	memoryMB := float64(m.Alloc) / 1024 / 1024

	// Number of goroutines
	goroutines := runtime.NumGoroutine()

	// CPU usage would require more complex implementation
	// For now, we'll set it to 0 as placeholder
	cpuUsage := 0.0

	h.metrics.UpdateResourceUsage(memoryMB, cpuUsage, goroutines)
}

// RegisterHealthEndpoint registers health check endpoint with metrics
func (h *HealthMonitor) RegisterHealthEndpoint(router *gin.Engine) {
	router.GET("/metrics/health", func(c *gin.Context) {
		health := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"uptime":    time.Since(h.startTime).String(),
			"version":   "1.0.0",
			"service":   "service-general",
		}

		// Check database
		dbHealthy := h.checkAllDatabases()
		health["database"] = gin.H{
			"status":    "healthy",
			"connected": dbHealthy,
			"databases": h.getDatabaseStatuses(),
		}

		// Check cache
		cacheHealthy := h.checkCache()
		health["cache"] = gin.H{
			"status":    "healthy",
			"connected": cacheHealthy,
		}

		// Check external services
		externalHealthy := h.checkExternalServices()
		health["external_services"] = gin.H{
			"status":    "healthy",
			"connected": externalHealthy,
			"services":  h.getExternalServiceStatuses(),
		}

		// Determine overall status
		overallHealthy := dbHealthy && cacheHealthy && externalHealthy
		if !overallHealthy {
			health["status"] = "unhealthy"
			c.JSON(503, health)
			return
		}

		c.JSON(200, health)
	})
}

// getDatabaseStatuses returns detailed database statuses
func (h *HealthMonitor) getDatabaseStatuses() map[string]interface{} {
	statuses := make(map[string]interface{})

	dbList := h.dbService.ListDBs()
	for _, dbName := range dbList {
		dbType, _ := h.dbService.GetDBType(dbName)
		healthy := h.checkDatabase(dbName)

		statuses[dbName] = gin.H{
			"type":      string(dbType),
			"connected": healthy,
		}
	}

	return statuses
}

// getExternalServiceStatuses returns external service statuses
func (h *HealthMonitor) getExternalServiceStatuses() map[string]interface{} {
	statuses := make(map[string]interface{})

	if h.config != nil {
		if h.config.Bpjs.BaseURL != "" {
			statuses["bpjs"] = gin.H{
				"connected": true, // Placeholder
				"url":       h.config.Bpjs.BaseURL,
			}
		}

		if h.config.SatuSehat.BaseURL != "" {
			statuses["satu_sehat"] = gin.H{
				"connected": true, // Placeholder
				"url":       h.config.SatuSehat.BaseURL,
			}
		}
	}

	return statuses
}
