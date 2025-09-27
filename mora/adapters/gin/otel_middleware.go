package gin

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// ObservabilityMiddleware returns a Gin middleware that adds OpenTelemetry tracing
func ObservabilityMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
