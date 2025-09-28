package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/sony/gobreaker"
)

// CircuitBreakerManager manages multiple circuit breakers for different endpoints
type CircuitBreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	mu       sync.RWMutex
	config   CircuitBreakerConfig
}

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	// MaxRequests is the maximum number of requests allowed to pass through
	// when the circuit breaker is half-open
	MaxRequests uint32 `mapstructure:"max_requests" default:"5"`

	// Interval is the cyclic period of the closed state
	// for the circuit breaker to clear the internal counts
	Interval time.Duration `mapstructure:"interval" default:"60s"`

	// Timeout is the period of the open state,
	// after which the state of the circuit breaker becomes half-open
	Timeout time.Duration `mapstructure:"timeout" default:"30s"`

	// ReadyToTrip is called with a copy of Counts whenever a request fails in the closed state.
	// If ReadyToTrip returns true, the circuit breaker will be placed into the open state.
	FailureThreshold uint32 `mapstructure:"failure_threshold" default:"5"`
	FailureRatio     float64 `mapstructure:"failure_ratio" default:"0.6"`

	// MinRequests is the minimum number of requests to evaluate failure ratio
	MinRequests uint32 `mapstructure:"min_requests" default:"10"`
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(config CircuitBreakerConfig) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		config:   config,
	}
}

// getBreaker gets or creates a circuit breaker for the given endpoint
func (cbm *CircuitBreakerManager) getBreaker(endpoint string) *gobreaker.CircuitBreaker {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[endpoint]
	cbm.mu.RUnlock()

	if exists {
		return breaker
	}

	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := cbm.breakers[endpoint]; exists {
		return breaker
	}

	settings := gobreaker.Settings{
		Name:        endpoint,
		MaxRequests: cbm.config.MaxRequests,
		Interval:    cbm.config.Interval,
		Timeout:     cbm.config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip if we have enough requests and either:
			// 1. Total failures exceed threshold, OR
			// 2. Failure ratio exceeds threshold
			if counts.Requests < cbm.config.MinRequests {
				return false
			}

			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.TotalFailures >= cbm.config.FailureThreshold ||
				failureRatio >= cbm.config.FailureRatio
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log := logger.NewDefault()
			log.Info("Circuit breaker state changed",
				"endpoint", name,
				"from", from.String(),
				"to", to.String())
		},
	}

	breaker = gobreaker.NewCircuitBreaker(settings)
	cbm.breakers[endpoint] = breaker
	return breaker
}

// Middleware returns a Gin middleware for circuit breaking
func (cbm *CircuitBreakerManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.NewDefault().WithContext(c.Request.Context())

		// Create endpoint identifier (method + path)
		endpoint := c.Request.Method + " " + c.FullPath()
		breaker := cbm.getBreaker(endpoint)

		// Execute request through circuit breaker
		result, err := breaker.Execute(func() (interface{}, error) {
			// Continue with the request
			c.Next()

			// Check if the request was successful based on status code
			statusCode := c.Writer.Status()
			if statusCode >= 500 {
				// Consider 5xx as failures
				return nil, &CircuitBreakerError{
					StatusCode: statusCode,
					Message:    "Server error detected",
				}
			}

			return nil, nil
		})

		// Handle circuit breaker errors
		if err != nil {
			switch err {
			case gobreaker.ErrOpenState:
				log.Warn("Circuit breaker is open", "endpoint", endpoint)
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "service_unavailable",
					"message": "Service is temporarily unavailable. Please try again later.",
				})
				c.Abort()
				return

			case gobreaker.ErrTooManyRequests:
				log.Warn("Circuit breaker: too many requests", "endpoint", endpoint)
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "too_many_requests",
					"message": "Too many requests. Please try again later.",
				})
				c.Abort()
				return

			default:
				// This is an actual error from the request execution
				if _, ok := err.(*CircuitBreakerError); ok {
					// This was a server error, already handled in the request
					log.Warn("Request failed", "endpoint", endpoint, "error", err.Error())
				} else {
					// Unexpected error
					log.Error("Circuit breaker execution failed", "endpoint", endpoint, "error", err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "internal_server_error",
						"message": "An unexpected error occurred",
					})
					c.Abort()
				}
				return
			}
		}

		_ = result // Suppress unused variable warning
	}
}

// CircuitBreakerError represents an error that should trip the circuit breaker
type CircuitBreakerError struct {
	StatusCode int
	Message    string
}

func (e *CircuitBreakerError) Error() string {
	return e.Message
}

// GetStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerManager) GetStats() map[string]interface{} {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	stats := make(map[string]interface{})

	for endpoint, breaker := range cbm.breakers {
		counts := breaker.Counts()
		stats[endpoint] = map[string]interface{}{
			"state":           breaker.State().String(),
			"requests":        counts.Requests,
			"total_successes": counts.TotalSuccesses,
			"total_failures":  counts.TotalFailures,
			"consecutive_successes": counts.ConsecutiveSuccesses,
			"consecutive_failures":  counts.ConsecutiveFailures,
		}
	}

	return stats
}

// Reset resets all circuit breakers (Note: gobreaker doesn't have Reset method)
func (cbm *CircuitBreakerManager) Reset() {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	// Clear all existing breakers - they will be recreated when needed
	cbm.breakers = make(map[string]*gobreaker.CircuitBreaker)
}