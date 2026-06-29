package cache

import (
	"context"
	"fmt"
	"time"
)

// Manager provides high-level cache operations
type Manager struct {
	cache  Cache
	config CacheConfig
}

// RedisClientProvider interface untuk cache yang menyediakan Redis client
type RedisClientProvider interface {
	GetRedisClient() interface{}
}

// NewManager creates a new cache manager
func NewManager(cache Cache, config CacheConfig) *Manager {
	return &Manager{
		cache:  cache,
		config: config,
	}
}

// Session operations
func (m *Manager) SetSession(ctx context.Context, sessionID string, userData interface{}) error {
	key := KeyPrefixSession + sessionID
	return m.cache.SetStruct(ctx, key, userData, m.config.SessionTTL)
}

func (m *Manager) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	key := KeyPrefixSession + sessionID
	return m.cache.GetStruct(ctx, key, dest)
}

func (m *Manager) DeleteSession(ctx context.Context, sessionID string) error {
	key := KeyPrefixSession + sessionID
	return m.cache.Delete(ctx, key)
}

// Rate limiting operations
func (m *Manager) IncrementRateLimit(ctx context.Context, identifier string) (int64, error) {
	key := KeyPrefixRateLimit + identifier
	return m.cache.Increment(ctx, key, 1)
}

func (m *Manager) GetRateLimit(ctx context.Context, identifier string) (int64, error) {
	key := KeyPrefixRateLimit + identifier
	// Try to get as string first, then convert
	value, err := m.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	var count int64
	_, err = fmt.Sscanf(value, "%d", &count)
	return count, err
}

func (m *Manager) ResetRateLimit(ctx context.Context, identifier string) error {
	key := KeyPrefixRateLimit + identifier
	return m.cache.Delete(ctx, key)
}

func (m *Manager) SetRateLimit(ctx context.Context, identifier string, count int64) error {
	key := KeyPrefixRateLimit + identifier
	value := fmt.Sprintf("%d", count)
	return m.cache.Set(ctx, key, value, m.config.RateLimitTTL)
}

// User cache operations
func (m *Manager) SetUser(ctx context.Context, userID string, userData interface{}) error {
	key := KeyPrefixUser + userID
	return m.cache.SetStruct(ctx, key, userData, m.config.DefaultTTL)
}

func (m *Manager) GetUser(ctx context.Context, userID string, dest interface{}) error {
	key := KeyPrefixUser + userID
	return m.cache.GetStruct(ctx, key, dest)
}

func (m *Manager) DeleteUser(ctx context.Context, userID string) error {
	key := KeyPrefixUser + userID
	return m.cache.Delete(ctx, key)
}

// Token operations
func (m *Manager) SetToken(ctx context.Context, token string, tokenData interface{}, expiration time.Duration) error {
	key := KeyPrefixToken + token
	return m.cache.SetStruct(ctx, key, tokenData, expiration)
}

func (m *Manager) GetToken(ctx context.Context, token string, dest interface{}) error {
	key := KeyPrefixToken + token
	return m.cache.GetStruct(ctx, key, dest)
}

func (m *Manager) DeleteToken(ctx context.Context, token string) error {
	key := KeyPrefixToken + token
	return m.cache.Delete(ctx, key)
}

// Generic cache operations with TTL
func (m *Manager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return m.cache.SetStruct(ctx, key, value, expiration)
}

func (m *Manager) Get(ctx context.Context, key string, dest interface{}) error {
	return m.cache.GetStruct(ctx, key, dest)
}

func (m *Manager) Delete(ctx context.Context, key string) error {
	return m.cache.Delete(ctx, key)
}

func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	return m.cache.Exists(ctx, key)
}

// Health check
func (m *Manager) Health(ctx context.Context) error {
	return m.cache.Health(ctx)
}

// GetRedisClient mendapatkan Redis client jika tersedia
func (m *Manager) GetRedisClient() interface{} {
	if redisProvider, ok := m.cache.(RedisClientProvider); ok {
		return redisProvider.GetRedisClient()
	}
	return nil
}

// Close the cache connection
func (m *Manager) Close() error {
	return m.cache.Close()
}
