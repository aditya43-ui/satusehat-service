package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"service/pkg/logger" // Tambahkan import ini

	"github.com/google/uuid"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	// Development mode - izinkan semua origin (hanya untuk dev!)
	if os.Getenv("APP_ENV") == "development" && os.Getenv("CORS_ALLOW_ALL") == "true" {
		log.Println("WARNING: CORS allowing all origins (development mode)")

		// Gunakan config khusus untuk allow all origins
		config := cors.DefaultConfig()
		config.AllowAllOrigins = true
		config.AllowCredentials = false // Tidak bisa digunakan dengan AllowAllOrigins
		config.AllowMethods = []string{
			"GET", "POST", "PUT", "PATCH", "DELETE",
			"HEAD", "OPTIONS",
		}
		config.AllowHeaders = []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-API-Key",
			"X-CSRF-Token",
			"X-Custom-Header",
			"Accept",
			"Accept-Language",
			"Accept-Encoding",
			"Access-Control-Request-Headers",
			"Access-Control-Request-Method",
			// Headers tambahan untuk Nuxt 3
			"x-use-fetch",
			"x-nuxt-base-url",
			"x-forwarded-for",
			"x-forwarded-proto",
			"x-forwarded-host",
		}
		config.MaxAge = 12 * time.Hour

		return cors.New(config)
	}

	// Config untuk specific origins
	config := cors.DefaultConfig()

	// Baca allowed origins dari environment variable
	// Format: CORS_ORIGINS=http://localhost:3000,http://localhost:3001,https://myapp.com
	originsEnv := os.Getenv("CORS_ORIGINS")
	if originsEnv != "" {
		// Split by comma dan trim spaces
		origins := strings.Split(originsEnv, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		config.AllowOrigins = origins
		log.Printf("CORS: Using origins from environment: %v", config.AllowOrigins)
	} else {
		// Default origins untuk Nuxt 3 development
		config.AllowOrigins = []string{
			"http://localhost:3000",            // Nuxt 3 default
			"http://localhost:3001",            // Nuxt 3 alternatif
			"http://localhost:3002",            // Nuxt 3 alternatif
			"http://localhost:3005",            // Nuxt 3 port Anda
			"http://localhost:8080",            // Common dev port
			"http://localhost:5173",            // Vite default port
			"http://localhost:5174",            // Vite alternatif
			"https://localhost:3000",           // HTTPS Nuxt
			"https://localhost:8080",           // HTTPS common
			"http://meninjar.dev.rssa.id:8094", // Domain production Anda
		}
		log.Printf("CORS: Using default origins: %v", config.AllowOrigins)
	}

	// Method yang diizinkan untuk Nuxt 3 + TypeScript
	config.AllowMethods = []string{
		"GET", "POST", "PUT", "PATCH", "DELETE",
		"HEAD", "OPTIONS",
	}

	// Headers yang diizinkan untuk Nuxt 3 + Axios
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"X-Requested-With",
		"X-API-Key",
		"X-CSRF-Token",
		"X-Custom-Header",
		"Accept",
		"Accept-Language",
		"Accept-Encoding",
		"Access-Control-Request-Headers",
		"Access-Control-Request-Method",
		// Headers tambahan untuk Nuxt 3
		"x-use-fetch",
		"x-nuxt-base-url",
		"x-forwarded-for",
		"x-forwarded-proto",
		"x-forwarded-host",
	}

	// Izinkan credentials (penting untuk Nuxt 3)
	config.AllowCredentials = true

	// Preflight cache duration
	config.MaxAge = 12 * time.Hour

	return cors.New(config)
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Ambil atau Generate Request ID untuk Tracing (Correlation ID)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Injeksi request_id ke dalam context request
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", requestID)

		// Process request
		c.Next()

		// Log using custom logger
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		fields := []logger.Field{
			logger.String("ip", clientIP),
			logger.String("method", method),
			logger.String("path", path),
			logger.Int("status", statusCode),
			logger.Duration("latency", latency),
			logger.String("user_agent", c.Request.UserAgent()),
		}

		if raw != "" {
			fields = append(fields, logger.String("query", raw))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, logger.String("error", c.Errors.String()))
		}

		logCtx := logger.Default().WithContext(ctx)
		// Use appropriate log level based on status code
		if statusCode >= 500 {
			logCtx.Error("HTTP Request", fields...)
		} else if statusCode >= 400 {
			logCtx.Warn("HTTP Request", fields...)
		} else {
			logCtx.Info("HTTP Request", fields...)
		}
	}
}

func ErrorMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Default().WithContext(c.Request.Context()).Error("Panic recovered", logger.Any("panic", recovered))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	})
}

func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. HSTS (Strict-Transport-Security)
		// Memaksa browser hanya menggunakan HTTPS selama 1 tahun, termasuk subdomain.
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// 2. Content Security Policy (CSP) - Code Injection Protection
		// Kita longgarkan khusus untuk path /swagger agar UI bisa melakukan load JavaScript & CSS inline bawaannya.
		if strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; object-src 'none'; frame-ancestors 'none'")
		} else {
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none'; frame-ancestors 'none'; upgrade-insecure-requests; block-all-mixed-content")
		}

		// 3. X-Content-Type-Options
		// Mencegah browser menebak (sniffing) MIME type.
		c.Header("X-Content-Type-Options", "nosniff")

		// 4. X-Frame-Options (Clickjacking Protection)
		// Mencegah website di-embed dalam iframe orang lain.
		c.Header("X-Frame-Options", "DENY")

		// 5. X-XSS-Protection
		// Layer pertahanan lama untuk browser lama (Legacy), tapi tetap bagus untuk ada.
		c.Header("X-XSS-Protection", "1; mode=block")

		// 6. Referrer-Policy
		// Menjaga privasi user saat klik link keluar dari aplikasi Anda.
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 7. Permissions-Policy (Feature Policy)
		// Mematikan fitur browser yang tidak dipakai (kamera, mic, lokasi) untuk mengurangi attack vector.
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		// Remove Information Leakage
		c.Header("Server", "Unknown") // Atau hapus total
		c.Header("X-Powered-By", "")

		c.Next()
	}
}
