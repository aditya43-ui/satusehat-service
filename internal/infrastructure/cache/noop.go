// File: /home/meninjar/goprint/service-general/internal/infrastructure/cache/noop.go
package cache

import (
	"context"
	"errors"
	"time"
)

// NoOpCache is a no-operation cache implementation for development
type NoOpCache struct{}

// NewNoOpCache creates a new no-op cache instance
func NewNoOpCache() Cache {
	return &NoOpCache{}
}

// Get always returns empty string and error
func (n *NoOpCache) Get(ctx context.Context, key string) (string, error) {
	return "", errors.New("cache disabled")
}

// Set always succeeds but does nothing
func (n *NoOpCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return nil
}

// Delete always succeeds but does nothing
func (n *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

// Exists always returns false
func (n *NoOpCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// GetStruct always returns error
func (n *NoOpCache) GetStruct(ctx context.Context, key string, dest interface{}) error {
	return errors.New("cache disabled")
}

// SetStruct always succeeds but does nothing
func (n *NoOpCache) SetStruct(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

// Increment always returns 0 and error
func (n *NoOpCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, errors.New("cache disabled")
}

// Decrement always returns 0 and error
func (n *NoOpCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return 0, errors.New("cache disabled")
}

// TTL always returns 0 and error
func (n *NoOpCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, errors.New("cache disabled")
}

// Close always succeeds
func (n *NoOpCache) Close() error {
	return nil
}

// Health always succeeds
func (n *NoOpCache) Health(ctx context.Context) error {
	return nil
}

// MGet always returns empty slice and error
func (n *NoOpCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return []interface{}{}, errors.New("cache disabled")
}

// MSet always succeeds but does nothing
func (n *NoOpCache) MSet(ctx context.Context, items map[string]interface{}, expiration time.Duration) error {
	return nil
}
