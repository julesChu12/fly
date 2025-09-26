package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware creates a Gin middleware that logs HTTP requests using zap logger
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get response status
		statusCode := c.Writer.Status()

		// Build log fields
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		}

		// Add user context if available
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		if tenantID, exists := c.Get("tenant_id"); exists {
			fields = append(fields, zap.Any("tenant_id", tenantID))
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.Error("HTTP request completed with server error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("HTTP request completed with client error", fields...)
		} else {
			logger.Info("HTTP request completed", fields...)
		}
	}
}