package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// HTTPClientConfig represents the configuration for the optimized HTTP client
type HTTPClientConfig struct {
	// Connection pool settings
	MaxIdleConns        int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host"`
	MaxConnsPerHost     int           `yaml:"max_conns_per_host"`
	IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout"`

	// Request settings
	Timeout         time.Duration `yaml:"timeout"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout"`
	KeepAlive       time.Duration `yaml:"keep_alive"`

	// Retry settings
	MaxRetries      int           `yaml:"max_retries"`
	RetryDelay      time.Duration `yaml:"retry_delay"`
	BackoffMultiplier float64      `yaml:"backoff_multiplier"`
	MaxRetryDelay   time.Duration `yaml:"max_retry_delay"`

	// Circuit breaker settings
	CircuitBreakerEnabled bool          `yaml:"circuit_breaker_enabled"`
	FailureThreshold      int           `yaml:"failure_threshold"`
	RecoveryTimeout       time.Duration `yaml:"recovery_timeout"`
	RequestTimeout        time.Duration `yaml:"request_timeout"`
}

// DefaultHTTPClientConfig returns a default configuration for the HTTP client
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		// Connection pool settings (optimized for performance)
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,

		// Request settings
		Timeout:        30 * time.Second,
		ConnectTimeout: 10 * time.Second,
		KeepAlive:     30 * time.Second,

		// Retry settings
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxRetryDelay:    5 * time.Second,

		// Circuit breaker settings
		CircuitBreakerEnabled: true,
		FailureThreshold:     5,
		RecoveryTimeout:      30 * time.Second,
		RequestTimeout:       15 * time.Second,
	}
}

// OptimizedHTTPClient represents an optimized HTTP client with connection pooling, retry logic, and circuit breaker
type OptimizedHTTPClient struct {
	client    *http.Client
	config    *HTTPClientConfig
	logger    *logger.Logger

	// Circuit breaker state
	failureCount int
	lastFailure  time.Time
	state        CircuitState
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	CircuitStateClosed CircuitState = iota
	CircuitStateOpen
	CircuitStateHalfOpen
)

// NewOptimizedHTTPClient creates a new optimized HTTP client with the given configuration
func NewOptimizedHTTPClient(baseURL string, config *HTTPClientConfig, logger *logger.Logger) *OptimizedHTTPClient {
	if config == nil {
		config = DefaultHTTPClientConfig()
	}

	// Create transport with optimized connection pool settings
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		MaxConnsPerHost:     config.MaxConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,

		// Enable HTTP/2 for better performance
		ForceAttemptHTTP2: true,

		// Connection settings
		DialContext: (&net.Dialer{
			Timeout:   config.ConnectTimeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,

		// Response headers timeout
		ResponseHeaderTimeout: config.Timeout,

		// Expect continue timeout for large uploads
		ExpectContinueTimeout: 1 * time.Second,

		// Enable proxy support
		Proxy: http.ProxyFromEnvironment,
	}

	// Create HTTP client with optimized settings
	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	// Disable follow redirects to have more control
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &OptimizedHTTPClient{
		client: client,
		config: config,
		logger: logger,
		state:  CircuitStateClosed,
	}
}

// Do executes an HTTP request with retry logic and circuit breaker
func (c *OptimizedHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Check circuit breaker state
	if c.config.CircuitBreakerEnabled && !c.canExecuteRequest() {
		return nil, fmt.Errorf("circuit breaker is open, rejecting request")
	}

	// Add common headers
	c.addCommonHeaders(req)

	var lastErr error

	// Implement retry logic
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := time.Duration(c.config.RetryDelay) *
				time.Duration(1<<(attempt-1)) * time.Duration(c.config.BackoffMultiplier)

			if delay > c.config.MaxRetryDelay {
				delay = c.config.MaxRetryDelay
			}

			// Log retry attempt
			c.logger.Warn("HTTP request retrying",
				"attempt", attempt+1,
				"max_retries", c.config.MaxRetries+1,
				"delay", delay,
				"method", req.Method,
				"url", req.URL.String(),
				"last_error", lastErr.Error(),
			)

			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		// Clone the request body for retry (since body can only be read once)
		var reqBodyCopy io.ReadCloser
		if req.Body != nil {
			// Read the body
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				lastErr = fmt.Errorf("failed to read request body: %w", err)
				c.recordFailure()
				continue
			}

			// Create a new reader from the bytes
			reqBodyCopy = io.NopCloser(strings.NewReader(string(bodyBytes)))
			req.Body = reqBodyCopy
		}

		// Execute the request
		resp, err := c.client.Do(req.WithContext(ctx))
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			c.recordFailure()

			// Check if this is a retryable error
			if !c.isRetryableError(err) {
				break
			}
			continue
		}

		// Check if response indicates success
		if c.isSuccessResponse(resp) {
			c.recordSuccess()
			return resp, nil
		}

		// Read response body for logging
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		lastErr = fmt.Errorf("HTTP request returned error status: %d, body: %s",
			resp.StatusCode, string(bodyBytes))
		c.recordFailure()

		// Don't retry on certain status codes
		if !c.isRetryableStatus(resp.StatusCode) {
			break
		}

		// Close response body before retry
		resp.Body.Close()
	}

	// All retries failed
	return nil, lastErr
}

// Get performs an HTTP GET request
func (c *OptimizedHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	return c.Do(ctx, req)
}

// Post performs an HTTP POST request
func (c *OptimizedHTTPClient) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return c.Do(ctx, req)
}

// Put performs an HTTP PUT request
func (c *OptimizedHTTPClient) Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create PUT request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return c.Do(ctx, req)
}

// Delete performs an HTTP DELETE request
func (c *OptimizedHTTPClient) Delete(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DELETE request: %w", err)
	}

	return c.Do(ctx, req)
}

// addCommonHeaders adds common headers to the request
func (c *OptimizedHTTPClient) addCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Clotho-Orchestration/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Request-ID", generateRequestID())

	// Add connection keep-alive header
	req.Header.Set("Connection", "keep-alive")
}

// canExecuteRequest checks if the request can be executed based on circuit breaker state
func (c *OptimizedHTTPClient) canExecuteRequest() bool {
	switch c.state {
	case CircuitStateClosed:
		return true
	case CircuitStateOpen:
		// Check if recovery timeout has passed
		if time.Since(c.lastFailure) > c.config.RecoveryTimeout {
			c.state = CircuitStateHalfOpen
			c.logger.Info("Circuit breaker transitioning to half-open state")
			return true
		}
		return false
	case CircuitStateHalfOpen:
		return true
	default:
		return false
	}
}

// recordFailure records a failure and updates circuit breaker state
func (c *OptimizedHTTPClient) recordFailure() {
	c.failureCount++
	c.lastFailure = time.Now()

	if c.failureCount >= c.config.FailureThreshold {
		c.state = CircuitStateOpen
		c.logger.Warn("Circuit breaker opened due to failure threshold",
			"failure_count", c.failureCount,
			"threshold", c.config.FailureThreshold,
		)
	}
}

// recordSuccess records a success and updates circuit breaker state
func (c *OptimizedHTTPClient) recordSuccess() {
	if c.state == CircuitStateHalfOpen {
		// Reset failure count and close circuit
		c.failureCount = 0
		c.state = CircuitStateClosed
		c.logger.Info("Circuit breaker closed after successful request")
	} else if c.state == CircuitStateClosed {
		// Reset failure count on success in closed state
		c.failureCount = 0
	}
}

// isRetryableError checks if an error is retryable
func (c *OptimizedHTTPClient) isRetryableError(err error) bool {
	if err == nil {
		return true
	}

	// Check for network errors that are typically retryable
	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"connection timed out",
		"temporary failure",
		"network is unreachable",
		"no such host",
		"timeout",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryableErr) {
			return true
		}
	}

	return false
}

// isSuccessResponse checks if the response indicates success
func (c *OptimizedHTTPClient) isSuccessResponse(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// isRetryableStatus checks if the status code is retryable
func (c *OptimizedHTTPClient) isRetryableStatus(statusCode int) bool {
	// Don't retry on client errors (4xx) except specific cases
	retryableStatuses := []int{
		408, // Request Timeout
		429, // Too Many Requests
		500, // Internal Server Error
		502, // Bad Gateway
		503, // Service Unavailable
		504, // Gateway Timeout
		507, // Insufficient Storage
	}

	for _, retryableStatus := range retryableStatuses {
		if statusCode == retryableStatus {
			return true
		}
	}

	return false
}

// GetStats returns client statistics for monitoring
func (c *OptimizedHTTPClient) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"circuit_breaker_state": c.state.String(),
		"failure_count":         c.failureCount,
		"last_failure_time":      c.lastFailure,
		"max_idle_conns":        c.config.MaxIdleConns,
		"max_idle_conns_per_host": c.config.MaxIdleConnsPerHost,
	}
}

// String returns string representation of circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitStateClosed:
		return "closed"
	case CircuitStateOpen:
		return "open"
	case CircuitStateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// generateRequestID generates a unique request ID for tracing
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}