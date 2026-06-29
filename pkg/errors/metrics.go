package errors

import (
	"sync"
)

// Metrics holds error metrics
type Metrics struct {
	TotalErrors      int
	ErrorsByCode     map[string]int
	ErrorsByCategory map[string]int
}

// ErrorMetrics represents error metrics for external use
type ErrorMetrics struct {
	TotalErrors      int            `json:"total_errors"`
	ErrorsByCode     map[string]int `json:"errors_by_code"`
	ErrorsByCategory map[string]int `json:"errors_by_category"`
}

var (
	globalMetrics = &Metrics{
		ErrorsByCode:     make(map[string]int),
		ErrorsByCategory: make(map[string]int),
	}
	metricsMutex sync.RWMutex
)

// RecordError records an error in metrics
func RecordError(err Error) {
	metricsMutex.Lock()
	defer metricsMutex.Unlock()

	globalMetrics.TotalErrors++
	globalMetrics.ErrorsByCode[err.Code()]++
	globalMetrics.ErrorsByCategory[err.Category()]++
}

// GetMetrics returns a copy of current metrics
func GetMetrics() *Metrics {
	metricsMutex.RLock()
	defer metricsMutex.RUnlock()

	// Create a deep copy
	copy := &Metrics{
		TotalErrors:      globalMetrics.TotalErrors,
		ErrorsByCode:     make(map[string]int),
		ErrorsByCategory: make(map[string]int),
	}

	for k, v := range globalMetrics.ErrorsByCode {
		copy.ErrorsByCode[k] = v
	}

	for k, v := range globalMetrics.ErrorsByCategory {
		copy.ErrorsByCategory[k] = v
	}

	return copy
}
