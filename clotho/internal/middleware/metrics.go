package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/mora/pkg/observability"
)

// MetricsMiddleware provides detailed metrics collection for HTTP requests
type MetricsMiddleware struct {
	registry *observability.MetricsRegistry
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware() *MetricsMiddleware {
	registry := observability.NewMetricsRegistry("clotho")

	// Register standard metrics
	registry.RegisterStandardMetrics()

	// Register additional custom metrics
	registry.RegisterCounter("rate_limit_exceeded_total",
		"Total number of rate limit exceeded events",
		[]string{"limiter_type", "client_ip"})

	registry.RegisterGauge("circuit_breaker_state",
		"Current state of circuit breakers (0=closed, 1=half-open, 2=open)",
		[]string{"endpoint"})

	registry.RegisterCounter("circuit_breaker_trips_total",
		"Total number of circuit breaker trips",
		[]string{"endpoint"})

	registry.RegisterCounter("user_logins_total",
		"Total number of user login attempts",
		[]string{"user_type", "status"})

	registry.RegisterCounter("profile_updates_total",
		"Total number of profile updates",
		[]string{"update_type"})

	registry.RegisterCounter("api_calls_total",
		"Total number of API calls",
		[]string{"api_name", "status"})

	return &MetricsMiddleware{
		registry: registry,
	}
}

// Middleware returns a Gin middleware for metrics collection
func (m *MetricsMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Add metrics registry to context for use in handlers
		c.Set("metrics", m.registry)

		// Process request
		c.Next()

		// Collect metrics after request completion
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()

		// If path is empty (e.g., 404), use the raw path
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record HTTP metrics
		m.registry.ObserveHistogram("http_request_duration_seconds", duration,
			method, path, status)
		m.registry.IncrementCounter("http_requests_total",
			method, path, status)

		// Log detailed request info for tracing
		log := logger.NewDefault().WithContext(c.Request.Context())
		log.Info("HTTP request completed",
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", duration*1000,
			"client_ip", c.ClientIP(),
			"user_agent", c.GetHeader("User-Agent"))
	}
}

// RecordRateLimitExceeded records rate limit exceeded events
func (m *MetricsMiddleware) RecordRateLimitExceeded(limiterType, clientIP string) {
	m.registry.IncrementCounter("rate_limit_exceeded_total", limiterType, clientIP)
}

// RecordCircuitBreakerState records circuit breaker state changes
func (m *MetricsMiddleware) RecordCircuitBreakerState(endpoint string, state int) {
	m.registry.SetGauge("circuit_breaker_state", float64(state), endpoint)
}

// RecordCircuitBreakerTrip records circuit breaker trips
func (m *MetricsMiddleware) RecordCircuitBreakerTrip(endpoint string) {
	m.registry.IncrementCounter("circuit_breaker_trips_total", endpoint)
}

// RecordGRPCClientRequest records gRPC client request metrics
func (m *MetricsMiddleware) RecordGRPCClientRequest(service, method, status string, duration time.Duration) {
	durationSeconds := duration.Seconds()
	m.registry.ObserveHistogram("grpc_request_duration_seconds", durationSeconds,
		service, method, status)
	m.registry.IncrementCounter("grpc_requests_total",
		service, method, status)
}

// GetMetricsRegistry returns the metrics registry for direct use
func (m *MetricsMiddleware) GetMetricsRegistry() *observability.MetricsRegistry {
	return m.registry
}

// Business Metrics Helper Functions

// RecordUserLogin records user login events
func RecordUserLogin(c *gin.Context, userType string, success bool) {
	if registry, exists := c.Get("metrics"); exists {
		m := registry.(*observability.MetricsRegistry)
		status := "failure"
		if success {
			status = "success"
		}
		m.IncrementCounter("user_logins_total", userType, status)
	}
}

// RecordProfileUpdate records profile update events
func RecordProfileUpdate(c *gin.Context, updateType string) {
	if registry, exists := c.Get("metrics"); exists {
		m := registry.(*observability.MetricsRegistry)
		m.IncrementCounter("profile_updates_total", updateType)
	}
}

// RecordAPICall records specific API call metrics
func RecordAPICall(c *gin.Context, apiName string, success bool) {
	if registry, exists := c.Get("metrics"); exists {
		m := registry.(*observability.MetricsRegistry)
		status := "failure"
		if success {
			status = "success"
		}
		m.IncrementCounter("api_calls_total", apiName, status)
	}
}