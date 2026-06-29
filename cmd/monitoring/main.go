// File: /home/meninjar/goprint/service/cmd/api/main.go (Update)
package main

import (
	"context"
	"log"

	"service/internal/infrastructure/cache"
	"service/internal/infrastructure/config"
	"service/internal/infrastructure/database"
	"service/internal/infrastructure/monitoring"
	"service/pkg/logger"

	// Inisialisasi master/reference Ethnic Handler

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// ... existing config loading ...
	cfg := config.LoadConfig()
	// 2. Init Logger dengan konfigurasi dari config
	loggerConfig := logger.Config{
		Level:       cfg.Logger.Level,
		Format:      cfg.Logger.Format,
		Output:      "console",
		ServiceName: "service-general",
		Environment: cfg.Server.Mode,
	}
	logger.Init(loggerConfig)

	// Initialize Prometheus Metrics
	metrics := monitoring.NewMetrics("goprint-service")

	// 2. Init Database
	dbService := database.New(cfg)
	defer dbService.Close()

	// Run Database Migrations
	if err := dbService.Migrate(); err != nil {
		log.Printf("⚠️  Migration warning: %v", err)
	}

	// Get GORM DB Connection with metrics wrapper
	// 3. Init Cache using factory pattern
	cacheFactory := cache.NewFactory(cfg.Cache)
	cacheManager, err := cacheFactory.CreateManager()
	if err != nil {
		log.Fatalf("❌ Failed to initialize cache manager: %v", err)
	}
	defer cacheManager.Close()

	// Initialize Health Monitor
	healthMonitor := monitoring.NewHealthMonitor(metrics, dbService, cacheManager, cfg)

	// Start health monitoring in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go healthMonitor.Start(ctx)

	// 4. Init Layers (Repo & Usecase) - Now with wrapped DB
	// Update all repository initializations to use wrappedDB

	// Example:
	// ethnicRepo := ethnic.NewRepository(wrappedDB)
	// ethnicService := ethnic.NewService(ethnicRepo)
	// ethnicHandler := handlers.NewEthnicHandler(ethnicService)

	// 5. Init HTTP Server with metrics
	engine := gin.New()

	// Add metrics middleware
	engine.Use(monitoring.MetricsMiddleware(metrics))

	// Register Prometheus metrics endpoint
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Register health endpoint with metrics
	healthMonitor.RegisterHealthEndpoint(engine)

	// ... rest of the setup ...
}
