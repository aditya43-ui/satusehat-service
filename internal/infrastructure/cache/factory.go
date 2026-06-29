// File: /home/meninjar/goprint/service-general/internal/infrastructure/cache/factory.go
package cache

import (
	"fmt"

	"service/internal/infrastructure/config"
)

// Factory creates cache instances based on configuration
type Factory struct {
	config config.CacheConfig
}

// NewFactory creates a new cache factory
func NewFactory(cfg config.CacheConfig) *Factory {
	return &Factory{
		config: cfg,
	}
}

// Create creates a cache instance based on configuration
func (f *Factory) Create() (Cache, error) {
	if !f.config.Enabled {
		return NewNoOpCache(), nil
	}

	// For now, only Redis is supported
	return NewRedisCache(CacheConfig{
		Enabled:         f.config.Enabled,
		DefaultTTL:      f.config.DefaultTTL,
		SessionTTL:      f.config.SessionTTL,
		RateLimitTTL:    f.config.RateLimitTTL,
		CleanupInterval: f.config.CleanupInterval,
		MaxRetries:      f.config.MaxRetries,
		RetryDelay:      f.config.RetryDelay,
		Redis: RedisConfig{
			Host:     f.config.Redis.Host,
			Port:     f.config.Redis.Port,
			Password: f.config.Redis.Password,
			DB:       f.config.Redis.DB,
		},
	})
}

// CreateManager creates a cache manager with the configured cache
func (f *Factory) CreateManager() (*Manager, error) {
	cache, err := f.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return NewManager(cache, CacheConfig{
		Enabled:         f.config.Enabled,
		DefaultTTL:      f.config.DefaultTTL,
		SessionTTL:      f.config.SessionTTL,
		RateLimitTTL:    f.config.RateLimitTTL,
		CleanupInterval: f.config.CleanupInterval,
		MaxRetries:      f.config.MaxRetries,
		RetryDelay:      f.config.RetryDelay,
		Redis: RedisConfig{
			Host:     f.config.Redis.Host,
			Port:     f.config.Redis.Port,
			Password: f.config.Redis.Password,
			DB:       f.config.Redis.DB,
		},
	}), nil
}
