package errors

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"service/pkg/logger"
	"sync"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	MaxDelay    time.Duration
	Backoff     BackoffStrategy
	RetryIf     func(Error) bool
	OnRetry     func(attempt int, err Error)
}

// BackoffStrategy defines backoff strategy
type BackoffStrategy int

const (
	LinearBackoff BackoffStrategy = iota
	ExponentialBackoff
	ExponentialBackoffWithJitter
	FixedBackoff
)

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 3,
		Delay:       time.Second,
		MaxDelay:    30 * time.Second,
		Backoff:     ExponentialBackoff,
		RetryIf:     DefaultRetryCondition,
		OnRetry:     nil,
	}
}

// DefaultRetryCondition determines if error is retryable
func DefaultRetryCondition(err Error) bool {
	return IsRetryable(err.Code())
}

// Retry executes function with retry logic
func Retry(fn func() (interface{}, error), config *RetryConfig) (interface{}, error) {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr Error
	log := logger.Default()

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		appErr := FromError(err)
		if appErr == nil {
			appErr = NewWithCode(ErrCodeInternalError, err.Error())
		}

		// Check if error is retryable
		if config.RetryIf != nil && !config.RetryIf(appErr) {
			return nil, appErr
		}

		lastErr = appErr

		// Don't wait after last attempt
		if attempt < config.MaxAttempts-1 {
			delay := calculateDelay(attempt, config)

			if config.OnRetry != nil {
				config.OnRetry(attempt+1, appErr)
			}

			log.Warn(appErr.Error(),
				logger.Int("attempt", attempt+1),
				logger.Int("max_attempts", config.MaxAttempts),
				logger.String("retry_delay", delay.String()),
				logger.Bool("will_retry", attempt < config.MaxAttempts-1),
			)

			time.Sleep(delay)
		}
	}

	return nil, lastErr
}

// RetryWithContext executes function with retry logic and context
func RetryWithContext(ctx context.Context, fn func() (interface{}, error), config *RetryConfig) (interface{}, error) {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr Error
	log := logger.Default()

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, Wrap(ctx.Err(), ErrCodeTimeout, "Context cancelled during retry")
		default:
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		appErr := FromError(err)
		if appErr == nil {
			appErr = NewWithCode(ErrCodeInternalError, err.Error())
		}

		// Check if error is retryable
		if config.RetryIf != nil && !config.RetryIf(appErr) {
			return nil, appErr
		}

		lastErr = appErr

		// Don't wait after last attempt
		if attempt < config.MaxAttempts-1 {
			delay := calculateDelay(attempt, config)

			if config.OnRetry != nil {
				config.OnRetry(attempt+1, appErr)
			}

			log.Warn(appErr.Error(),
				logger.Int("attempt", attempt+1),
				logger.Int("max_attempts", config.MaxAttempts),
				logger.String("retry_delay", delay.String()),
				logger.Bool("will_retry", attempt < config.MaxAttempts-1),
			)

			// Wait with context cancellation
			select {
			case <-ctx.Done():
				return nil, Wrap(ctx.Err(), ErrCodeTimeout, "Context cancelled during retry delay")
			case <-time.After(delay):
			}
		}
	}

	return nil, lastErr
}

// calculateDelay calculates delay based on backoff strategy
func calculateDelay(attempt int, config *RetryConfig) time.Duration {
	var delay time.Duration

	switch config.Backoff {
	case LinearBackoff:
		delay = time.Duration(attempt+1) * config.Delay
	case ExponentialBackoff:
		delay = config.Delay * time.Duration(math.Pow(2, float64(attempt)))
	case ExponentialBackoffWithJitter:
		delay = config.Delay * time.Duration(math.Pow(2, float64(attempt)))
		// Add jitter
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
		delay += jitter
	case FixedBackoff:
		delay = config.Delay
	default:
		delay = config.Delay
	}

	// Apply max delay limit
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	return delay
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	name         string
	maxFailures  int
	resetTimeout time.Duration
	state        CircuitState
	failures     int
	lastFailTime time.Time
	mu           sync.RWMutex
}

// CircuitState represents circuit breaker state
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// Execute executes function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	log := logger.Default()

	// Check if circuit is open
	if cb.state == CircuitOpen {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			log.Info("Circuit breaker transitioning to half-open", logger.String("circuit", cb.name))
		} else {
			return nil, NewWithCode(ErrCodeServiceUnavailable,
				fmt.Sprintf("Circuit breaker '%s' is open", cb.name))
		}
	}

	// Execute function
	result, err := fn()
	if err != nil {
		cb.onFailure()
		return nil, err
	}

	cb.onSuccess()
	return result, nil
}

// onSuccess handles successful execution
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		log := logger.Default()
		log.Info("Circuit breaker closed", logger.String("circuit", cb.name))
	}
}

// onFailure handles failed execution
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = CircuitOpen
		log := logger.Default()
		log.Warn("Circuit breaker opened",
			logger.String("circuit", cb.name),
			logger.Int("failures", cb.failures),
			logger.Int("max_failures", cb.maxFailures),
		)
	}
}

// GetState returns current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailures returns current failure count
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// Reset resets circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitClosed
	cb.failures = 0
	cb.lastFailTime = time.Time{}

	log := logger.Default()
	log.Info("Circuit breaker reset", logger.String("circuit", cb.name))
}

// RecoveryManager manages error recovery strategies
type RecoveryManager struct {
	circuitBreakers map[string]*CircuitBreaker
	retryConfigs    map[string]*RetryConfig
	mu              sync.RWMutex
}

// NewRecoveryManager creates new recovery manager
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		circuitBreakers: make(map[string]*CircuitBreaker),
		retryConfigs:    make(map[string]*RetryConfig),
	}
}

// RegisterCircuitBreaker registers circuit breaker
func (rm *RecoveryManager) RegisterCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.circuitBreakers[name] = NewCircuitBreaker(name, maxFailures, resetTimeout)
}

// RegisterRetryConfig registers retry configuration
func (rm *RecoveryManager) RegisterRetryConfig(name string, config *RetryConfig) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.retryConfigs[name] = config
}

// ExecuteWithRecovery executes function with recovery strategies
func (rm *RecoveryManager) ExecuteWithRecovery(operationName string, fn func() (interface{}, error)) (interface{}, error) {
	rm.mu.RLock()
	cb, hasCB := rm.circuitBreakers[operationName]
	config, hasConfig := rm.retryConfigs[operationName]
	rm.mu.RUnlock()

	// Execute with circuit breaker if available
	if hasCB {
		return cb.Execute(func() (interface{}, error) {
			// Execute with retry if available
			if hasConfig {
				return Retry(fn, config)
			}
			return fn()
		})
	}

	// Execute with retry only
	if hasConfig {
		return Retry(fn, config)
	}

	// Execute normally
	return fn()
}

// GetCircuitBreaker returns circuit breaker by name
func (rm *RecoveryManager) GetCircuitBreaker(name string) (*CircuitBreaker, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	cb, exists := rm.circuitBreakers[name]
	return cb, exists
}

// GetRetryConfig returns retry config by name
func (rm *RecoveryManager) GetRetryConfig(name string) (*RetryConfig, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	config, exists := rm.retryConfigs[name]
	return config, exists
}

// Global recovery manager
var globalRecoveryManager = NewRecoveryManager()

// GetRecoveryManager returns global recovery manager
func GetRecoveryManager() *RecoveryManager {
	return globalRecoveryManager
}

// ExecuteWithGlobalRecovery executes with global recovery manager
func ExecuteWithGlobalRecovery(operationName string, fn func() (interface{}, error)) (interface{}, error) {
	return globalRecoveryManager.ExecuteWithRecovery(operationName, fn)
}
