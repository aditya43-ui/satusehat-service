// File: /home/meninjar/goprint/service-general/internal/infrastructure/cache/cache.go
package cache

import (
	"context"
	"time"
)

// Cache interface defines the contract for cache implementations
type Cache interface {
	// Basic operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Struct operations
	GetStruct(ctx context.Context, key string, dest interface{}) error
	SetStruct(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Numeric operations
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)

	// Utility operations
	TTL(ctx context.Context, key string) (time.Duration, error)
	Close() error
	Health(ctx context.Context) error

	// Batch operations
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	MSet(ctx context.Context, items map[string]interface{}, expiration time.Duration) error
}

// CacheConfig holds configuration for cache
type CacheConfig struct {
	Enabled         bool
	DefaultTTL      time.Duration
	SessionTTL      time.Duration
	RateLimitTTL    time.Duration
	CleanupInterval time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	Redis           RedisConfig
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Common cache key prefixes
const (
	KeyPrefixSession   = "session:"
	KeyPrefixUser      = "user:"
	KeyPrefixToken     = "token:"
	KeyPrefixRateLimit = "rate_limit:"
	KeyPrefixCache     = "cache:"
	KeyPrefixTemp      = "temp:"
)

// Common expiration times
const (
	ExpirationSession  = 24 * time.Hour
	ExpirationShort    = 5 * time.Minute
	ExpirationMedium   = 30 * time.Minute
	ExpirationLong     = 2 * time.Hour
	ExpirationVeryLong = 24 * time.Hour
)
