// internal/infrastructure/container/container.go

package container

import (
	"fmt"

	"service/internal/infrastructure/cache"
	"service/internal/infrastructure/config"
	"service/pkg/logger" // Pastikan import ini benar
)

// Container holds all application dependencies
type Container struct {
	Config       *config.Config
	CacheManager *cache.Manager
}

// NewContainer creates a new container with all dependencies
func NewContainer(cfg *config.Config) (*Container, error) {
	// Create cache factory
	cacheFactory := cache.NewFactory(cfg.Cache)

	// Create cache manager
	cacheManager, err := cacheFactory.CreateManager()
	if err != nil {
		// PERBAIKAN: Gunakan logger baru yang terstruktur
		logger.Default().Fatal("Failed to create cache manager", logger.ErrorField(err))
	}

	// PERBAIKAN: Gunakan logger baru yang terstruktur
	logger.Default().Info("Cache manager initialized successfully")

	return &Container{
		Config:       cfg,
		CacheManager: cacheManager,
	}, nil
}

// Close closes all container resources
func (c *Container) Close() error {
	if c.CacheManager != nil {
		if err := c.CacheManager.Close(); err != nil {
			// PERBAIKAN: Gunakan logger baru, log error sebelum mengembalikan
			logger.Default().Error("Failed to close cache manager", logger.ErrorField(err))
			return fmt.Errorf("failed to close cache manager: %w", err)
		}
	}
	return nil
}
