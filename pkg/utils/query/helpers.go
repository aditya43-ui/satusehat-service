package query

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// getDefaultPort returns default port for database type
func getDefaultPort(dbType string) int {
	switch dbType {
	case "postgres":
		return 5432
	case "mysql":
		return 3306
	case "sqlserver":
		return 1433
	case "mongodb":
		return 27017
	case "sqlite":
		return 0 // SQLite doesn't use port
	default:
		return 5432
	}
}

// getDefaultSchema returns default schema for database type
func getDefaultSchema(dbType string) string {
	switch dbType {
	case "postgres":
		return "public"
	case "mysql":
		return ""
	case "sqlserver":
		return "dbo"
	case "mongodb":
		return ""
	case "sqlite":
		return ""
	default:
		return "public"
	}
}

// getDefaultSSLMode returns default SSL mode for database type
func getDefaultSSLMode(dbType string) string {
	switch dbType {
	case "postgres":
		return "disable"
	case "mysql":
		return "false"
	case "sqlserver":
		return "false"
	case "mongodb":
		return "false"
	case "sqlite":
		return ""
	default:
		return "disable"
	}
}

// getDefaultMaxOpenConns returns default max open connections for database type
func getDefaultMaxOpenConns(dbType string) int {
	switch dbType {
	case "postgres":
		return 25
	case "mysql":
		return 25
	case "sqlserver":
		return 25
	case "mongodb":
		return 100
	case "sqlite":
		return 1 // SQLite only supports one writer at a time
	default:
		return 25
	}
}

// getDefaultMaxIdleConns returns default max idle connections for database type
func getDefaultMaxIdleConns(dbType string) int {
	switch dbType {
	case "postgres":
		return 25
	case "mysql":
		return 25
	case "sqlserver":
		return 25
	case "mongodb":
		return 10
	case "sqlite":
		return 1 // SQLite only supports one writer at a time
	default:
		return 25
	}
}

// getDefaultConnMaxLifetime returns default connection max lifetime for database type
func getDefaultConnMaxLifetime(dbType string) string {
	switch dbType {
	case "postgres":
		return "5m"
	case "mysql":
		return "5m"
	case "sqlserver":
		return "5m"
	case "mongodb":
		return "30m"
	case "sqlite":
		return "5m"
	default:
		return "5m"
	}
}

// getEnvFromMap gets value from map with default
func getEnvFromMap(config map[string]string, key, defaultValue string) string {
	if value, exists := config[key]; exists {
		return value
	}
	return defaultValue
}

// getEnvAsIntFromMap gets int value from map with default
func getEnvAsIntFromMap(config map[string]string, key string, defaultValue int) int {
	if value, exists := config[key]; exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBoolFromMap gets bool value from map with default
func getEnvAsBoolFromMap(config map[string]string, key string, defaultValue bool) bool {
	if value, exists := config[key]; exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// parseSchemes parses comma-separated schemes string into a slice
func parseSchemes(schemesStr string) []string {
	if schemesStr == "" {
		return []string{"http"}
	}

	schemes := strings.Split(schemesStr, ",")
	for i, scheme := range schemes {
		schemes[i] = strings.TrimSpace(scheme)
	}
	return schemes
}

// parseStaticTokens parses comma-separated static tokens string into a slice
func parseStaticTokens(tokensStr string) []string {
	if tokensStr == "" {
		return []string{}
	}

	tokens := strings.Split(tokensStr, ",")
	var result []string

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token != "" {
			result = append(result, token)
		}
	}
	return result
}

// parseOrigins parses comma-separated origins string into a slice
func parseOrigins(originsStr string) []string {
	if originsStr == "" {
		return []string{"http://localhost:8080"} // Default for development
	}
	origins := strings.Split(originsStr, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

// parseDuration parses duration string
func parseDuration(durationStr string) time.Duration {
	if duration, err := time.ParseDuration(durationStr); err == nil {
		return duration
	}
	return 5 * time.Minute
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as int with default
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool gets environment variable as bool with default
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
