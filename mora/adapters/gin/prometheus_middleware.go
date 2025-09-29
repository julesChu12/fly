package gin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/observability"
)

// PrometheusMiddleware returns a Gin middleware that collects Prometheus metrics
func PrometheusMiddleware(metricsRegistry *observability.MetricsRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Collect metrics after request completion
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()

		// If path is empty (e.g., 404), use a generic label
		if path == "" {
			path = "unknown"
		}

		// Record HTTP metrics
		metricsRegistry.ObserveHistogram("http_request_duration_seconds", duration,
			method, path, status)
		metricsRegistry.IncrementCounter("http_requests_total",
			method, path, status)
	}
}

// SetupPrometheusMetrics creates a metrics registry with standard HTTP metrics
func SetupPrometheusMetrics(serviceName string) *observability.MetricsRegistry {
	registry := observability.NewMetricsRegistry(serviceName)
	registry.RegisterStandardMetrics()
	return registry
}