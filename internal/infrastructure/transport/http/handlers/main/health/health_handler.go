package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"service/internal/infrastructure/cache"
	"service/internal/infrastructure/config"
	"service/internal/infrastructure/database"
	"service/internal/interfaces/minio"

	"github.com/gin-gonic/gin"
	miniogo "github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db           *gorm.DB
	redisClient  *redis.Client
	config       *config.Config
	cacheManager *cache.Manager // Alternatif untuk Redis client
	dbManager    database.Service
}

// NewHealthHandlerWithCache creates a new health handler with cache manager
func NewHealthHandlerWithCache(db *gorm.DB, cacheManager *cache.Manager, config *config.Config, dbManager database.Service) *HealthHandler {
	handler := &HealthHandler{
		db:           db,
		config:       config,
		cacheManager: cacheManager,
		dbManager:    dbManager,
	}

	// Coba dapatkan Redis client dari cache manager
	if cacheManager != nil {
		if redisClientInterface := cacheManager.GetRedisClient(); redisClientInterface != nil {
			if client, ok := redisClientInterface.(*redis.Client); ok {
				handler.redisClient = client
			}
		}
	}

	return handler
}

// HealthCheckComplete performs comprehensive health check
func (h *HealthHandler) HealthCheckComplete(c *gin.Context) {
	startTime := time.Now()

	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(startTime).Milliseconds(),
		"version":   "1.0.0",
		"service":   "service",
	}

	// Check database
	dbStatus := h.checkDatabase()
	health["database"] = dbStatus

	// Check cache
	cacheStatus := h.checkCache()
	health["cache"] = cacheStatus

	// Check external services
	externalStatus := h.checkExternalServices()
	health["external_services"] = externalStatus

	// Check Minio
	minioStatus := h.checkMinio()
	health["minio"] = minioStatus

	// Determine overall status
	if dbStatus["status"] != "UP" || cacheStatus["status"] != "UP" || (minioStatus["status"] != "UP" && minioStatus["status"] != "DISABLED") {
		health["status"] = "DOWN"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	c.JSON(http.StatusOK, health)
}

// HealthCheckDatabase checks database connectivity
func (h *HealthHandler) HealthCheckDatabase(c *gin.Context) {
	status := h.checkDatabase()

	if status["status"] != "UP" {
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

// HealthCheckCache checks cache connectivity
func (h *HealthHandler) HealthCheckCache(c *gin.Context) {
	status := h.checkCache()

	if status["status"] != "UP" {
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

// HealthCheckExternal checks external service connectivity
func (h *HealthHandler) HealthCheckExternal(c *gin.Context) {
	status := h.checkExternalServices()

	// Hanya return 503 (Unavailable) jika status benar-benar DOWN, bukan saat sekadar DEGRADED.
	if status["status"] == "DOWN" {
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

// HealthCheckMinio checks Minio Object Storage connectivity
func (h *HealthHandler) HealthCheckMinio(c *gin.Context) {
	status := h.checkMinio()

	if status["status"] == "DOWN" {
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

// TestUploadMinio is a testing endpoint to upload a file to Minio
func (h *HealthHandler) TestUploadMinio(c *gin.Context) {
	if minio.I == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Minio client is not initialized or disconnected"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required (form-data key: 'file')"})
		return
	}

	// Gunakan bucket dari request body (jika ada), atau default ke config/nama statis
	bucketName := c.DefaultPostForm("bucket", "dev-test")
	ctx := c.Request.Context()

	// Ensure bucket exists
	exists, err := minio.I.BucketExists(ctx, bucketName)
	if err == nil && !exists {
		err = minio.I.MakeBucket(ctx, bucketName, miniogo.MakeBucketOptions{Region: h.config.Minio.Region})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bucket: " + err.Error()})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bucket status: " + err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file: " + err.Error()})
		return
	}
	defer src.Close()

	objectName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
	info, err := minio.I.PutObject(ctx, bucketName, objectName, src, file.Size, miniogo.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to Minio: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"data":    info,
	})
}

// checkDatabase checks database health
func (h *HealthHandler) checkDatabase() gin.H {
	if h.dbManager == nil {
		return gin.H{
			"status": "UNKNOWN",
			"error":  "Database manager not initialized",
		}
	}

	allDbInfo := h.dbManager.GetAllDatabasesInfo()
	overallStatus := "UP"

	// Iterate through the map and check each database
	for name, info := range allDbInfo {
		dbInfo, ok := info.(gin.H)
		if !ok {
			continue
		}

		// Standardize the status to "UP" if it's "connected" or "healthy"
		if status, ok := dbInfo["status"].(string); ok {
			if status == "connected" || status == "healthy" || status == "UP" {
				dbInfo["status"] = "UP"
			} else {
				overallStatus = "DOWN"
			}
		}
		allDbInfo[name] = dbInfo // Update the map with the checked info
	}

	return gin.H{
		"status":     overallStatus,
		"components": allDbInfo,
	}
}

// checkCache checks Redis/cache health
func (h *HealthHandler) checkCache() gin.H {
	if h.cacheManager == nil {
		return gin.H{"status": "UNKNOWN", "error": "Cache manager not initialized"}
	}

	if !h.config.Cache.Enabled {
		return gin.H{"status": "UP", "details": gin.H{"provider": "noop", "message": "Cache is disabled"}}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	err := h.cacheManager.Health(ctx)
	latency := time.Since(start)

	details := gin.H{
		"provider":   "redis",
		"latency":    latency.String(),
		"latency_ms": latency.Milliseconds(),
	}

	if err != nil {
		details["error"] = err.Error()
		return gin.H{"status": "DOWN", "details": details}
	}

	return gin.H{
		"status":  "UP",
		"details": details,
	}
}

// checkMinio checks Minio Object Storage health
func (h *HealthHandler) checkMinio() gin.H {
	if h.config == nil || h.config.Minio.Endpoint == "" {
		return gin.H{"status": "DISABLED", "message": "Minio is not configured in .env"}
	}

	if minio.I == nil {
		return gin.H{"status": "DOWN", "message": "Minio global client is nil"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	// Panggilan ringan API minio untuk memastikan server hidup & kredensial benar
	_, err := minio.I.ListBuckets(ctx)
	latency := time.Since(start)

	details := gin.H{
		"endpoint":   h.config.Minio.Endpoint,
		"latency_ms": latency.Milliseconds(),
	}

	if err != nil {
		details["error"] = err.Error()
		return gin.H{"status": "DOWN", "details": details}
	}

	return gin.H{"status": "UP", "details": details}
}

// checkExternalServices checks external service health
func (h *HealthHandler) checkExternalServices() gin.H {
	status := gin.H{
		"status":  "UP",
		"message": "All configured external services are reachable",
	}

	hasDegraded := false
	hasDown := false
	downServices := []string{}
	degradedServices := []string{}
	activeServices := 0

	// Check BPJS service if configured
	if h.config != nil {
		bpjsStatus := h.checkBPJSService()
		status["bpjs"] = bpjsStatus

		if s, ok := bpjsStatus["status"].(string); ok && s != "DISABLED" {
			activeServices++
			if s == "DOWN" {
				hasDown = true
				downServices = append(downServices, "BPJS")
			} else if s == "DEGRADED" {
				hasDegraded = true
				degradedServices = append(degradedServices, "BPJS")
			}
		}
	}

	// Check SatuSehat service if configured
	if h.config != nil {
		satuSehatStatus := h.checkSatuSehatService()
		status["satu_sehat"] = satuSehatStatus

		if s, ok := satuSehatStatus["status"].(string); ok && s != "DISABLED" {
			activeServices++
			if s == "DOWN" {
				hasDown = true
				downServices = append(downServices, "SatuSehat")
			} else if s == "DEGRADED" {
				hasDegraded = true
				degradedServices = append(degradedServices, "SatuSehat")
			}
		}
	}

	if activeServices == 0 {
		status["status"] = "UP"
		status["message"] = "All external services are disabled"
	} else if hasDown {
		status["status"] = "DOWN"
		status["message"] = fmt.Sprintf("Some external services are unreachable: %v", downServices)
	} else if hasDegraded {
		status["status"] = "DEGRADED"
		status["message"] = fmt.Sprintf("Some external services are degraded: %v", degradedServices)
	}

	return status
}

// checkBPJSService checks BPJS VClaim service
func (h *HealthHandler) checkBPJSService() gin.H {
	if !h.config.Bpjs.Enabled {
		return gin.H{
			"status":  "DISABLED",
			"message": "BPJS integration is disabled in configuration",
		}
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", h.config.Bpjs.BaseURL, nil)
	if err != nil {
		return gin.H{
			"status":  "DOWN",
			"message": "Failed to create BPJS request: " + err.Error(),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return gin.H{
			"status":  "DOWN",
			"message": "BPJS service unreachable: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	status := "UP"
	message := "BPJS service is reachable"

	switch resp.StatusCode {
	case http.StatusOK:
		message = "BPJS service is reachable and responding normally"
	case http.StatusUnauthorized:
		message = "BPJS service is reachable (Authentication required/Invalid Signature)"
	case http.StatusForbidden:
		status = "DEGRADED"
		message = "BPJS service is reachable, but access is forbidden (Check IP Whitelisting or Credentials)"
	case http.StatusNotFound:
		message = "BPJS service is reachable (Base URL responded with 404, server is UP)"
	default:
		if resp.StatusCode >= 500 {
			status = "DOWN"
			message = fmt.Sprintf("BPJS service returned server error: %s", resp.Status)
		} else {
			message = fmt.Sprintf("BPJS service is reachable (Status: %s)", resp.Status)
		}
	}

	return gin.H{
		"status":      status,
		"message":     message,
		"status_code": resp.StatusCode,
	}
}

// checkSatuSehatService checks SatuSehat FHIR service
func (h *HealthHandler) checkSatuSehatService() gin.H {
	if !h.config.SatuSehat.Enabled {
		return gin.H{
			"status":  "DISABLED",
			"message": "SatuSehat integration is disabled in configuration",
		}
	}

	client := &http.Client{
		Timeout: 2 * time.Second, // Dipercepat agar health check tidak blocking terlalu lama
	}

	checkURL := func(url string) (string, string, int) {
		if url == "" {
			return "DISABLED", "URL not configured", 0
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "DOWN", "Request creation failed: " + err.Error(), 0
		}
		resp, err := client.Do(req)
		if err != nil {
			return "DOWN", "Unreachable: " + err.Error(), 0
		}
		defer resp.Body.Close()
		return "UP", fmt.Sprintf("Reachable (Status: %s)", resp.Status), resp.StatusCode
	}

	authStatus, authMsg, authCode := checkURL(h.config.SatuSehat.AuthURL)
	baseStatus, baseMsg, baseCode := checkURL(h.config.SatuSehat.BaseURL)
	consentStatus, consentMsg, _ := checkURL(h.config.SatuSehat.ConsentURL)
	kfaStatus, kfaMsg, _ := checkURL(h.config.SatuSehat.KFAURL)

	overallStatus := "UP"
	overallMessage := "SatuSehat services are reachable"

	// Evaluasi status keseluruhan berdasarkan Base URL & Auth URL
	if baseStatus == "DOWN" || authStatus == "DOWN" {
		overallStatus = "DOWN"
		overallMessage = "One or more critical SatuSehat services are unreachable"
	} else if baseCode == http.StatusForbidden || authCode == http.StatusForbidden {
		overallStatus = "DEGRADED"
		overallMessage = "Services are reachable, but access is forbidden (Check IP/Permissions)"
	}

	return gin.H{
		"status":  overallStatus,
		"message": overallMessage,
		"details": gin.H{
			"org_id":       h.config.SatuSehat.OrgID,
			"fasyankes_id": h.config.SatuSehat.FasyakesID,
			"endpoints": gin.H{
				"auth":      gin.H{"url": h.config.SatuSehat.AuthURL, "status": authStatus, "message": authMsg, "status_code": authCode},
				"fhir_base": gin.H{"url": h.config.SatuSehat.BaseURL, "status": baseStatus, "message": baseMsg, "status_code": baseCode},
				"consent":   gin.H{"url": h.config.SatuSehat.ConsentURL, "status": consentStatus, "message": consentMsg},
				"kfa":       gin.H{"url": h.config.SatuSehat.KFAURL, "status": kfaStatus, "message": kfaMsg},
			},
		},
	}
}
