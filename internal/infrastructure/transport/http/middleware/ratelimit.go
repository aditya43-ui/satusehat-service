// internal/infrastructure/transport/http/middleware/rate_limit.go

package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"service/internal/infrastructure/cache"
	"service/pkg/logger" // Pastikan import ini benar

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware creates rate limiting middleware using Redis cache
func RateLimitMiddleware(cacheManager *cache.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Increment rate limit counter
		count, err := cacheManager.IncrementRateLimit(ctx, clientIP)
		if err != nil {
			// PERBAIKAN: Gunakan logger baru dengan konteks dan field terstruktur
			logger.Default().WithContext(ctx).
				Error("Failed to increment rate limit for IP",
					logger.ErrorField(err),
					logger.String("client_ip", clientIP),
				)
			// Allow request to proceed if cache is unavailable
			c.Next()
			return
		}

		// Set TTL on first request
		if count == 1 {
			if err := cacheManager.SetRateLimit(ctx, clientIP, count); err != nil {
				// PERBAIKAN: Gunakan logger baru
				logger.Default().WithContext(ctx).
					Error("Failed to set rate limit TTL for IP",
						logger.ErrorField(err),
						logger.String("client_ip", clientIP),
					)
			}
		}

		// Check if rate limit exceeded (60 requests per minute)
		if count > 60 {
			// PERBAIKAN: Tambahkan log saat rate limit terlampaui
			logger.Default().WithContext(ctx).
				Warn("Rate limit exceeded for IP",
					logger.String("client_ip", clientIP),
					logger.Int64("count", count),
				)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByToken creates rate limiting middleware based on auth token
func RateLimitByToken(cacheManager *cache.Manager, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.Next()
			return
		}

		// Extract token from Bearer format
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Increment rate limit counter
		count, err := cacheManager.IncrementRateLimit(ctx, "token:"+token)
		if err != nil {
			// PERBAIKAN: Gunakan logger baru
			logger.Default().WithContext(ctx).
				Error("Failed to increment token rate limit",
					logger.ErrorField(err),
					logger.String("token_prefix", token[:minLen(len(token), 10)]+"..."), // Jangan log token utuh
				)
			c.Next()
			return
		}

		// Set TTL on first request
		if count == 1 {
			if err := cacheManager.SetRateLimit(ctx, "token:"+token, count); err != nil {
				// PERBAIKAN: Gunakan logger baru
				logger.Default().WithContext(ctx).
					Error("Failed to set token rate limit TTL",
						logger.ErrorField(err),
						logger.String("token_prefix", token[:minLen(len(token), 10)]+"..."),
					)
			}
		}

		// Check if rate limit exceeded
		if count > int64(requestsPerMinute) {
			// PERBAIKAN: Tambahkan log saat rate limit token terlampaui
			logger.Default().WithContext(ctx).
				Warn("Rate limit exceeded for token",
					logger.String("token_prefix", token[:minLen(len(token), 10)]+"..."),
					logger.Int64("count", count),
				)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests for this token",
				"code":        "TOKEN_RATE_LIMIT_EXCEEDED",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function untuk menghindari error jika token lebih pendek dari 10 karakter
func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Variabel global untuk menyimpan limiter per-client IP/Identifier untuk memory rate limiter
var (
	visitors = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

// getVisitor mengambil atau membuat limiter baru untuk client identifier tertentu
func getVisitor(identifier string, r rate.Limit, b int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[identifier]
	if !exists {
		limiter = rate.NewLimiter(r, b)
		visitors[identifier] = limiter
	}
	return limiter
}

// MemoryRateLimitMiddleware membatasi jumlah request per client secara lokal di memori.
// requestsPerSecond: jumlah hit yang diizinkan per detik.
// burstSize: jumlah hit maksimal dalam satu waktu (burst).
func MemoryRateLimitMiddleware(requestsPerSecond float64, burstSize int) gin.HandlerFunc {
	limit := rate.Limit(requestsPerSecond)
	return func(c *gin.Context) {
		clientIdentifier := c.ClientIP()

		limiter := getVisitor(clientIdentifier, limit, burstSize)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "Terlalu banyak permintaan ke API Satu Sehat. Silakan coba beberapa saat lagi.",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
