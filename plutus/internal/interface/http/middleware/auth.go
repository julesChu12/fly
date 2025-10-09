package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/plutus/pkg/constants"
)

// TenantMiddleware extracts tenant_id from request headers and adds it to context
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDStr := c.GetHeader(constants.HeaderTenantID)
		if tenantIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing tenant ID header",
				"code":  "MISSING_TENANT_ID",
			})
			c.Abort()
			return
		}

		tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid tenant ID format",
				"code":  "INVALID_TENANT_ID",
			})
			c.Abort()
			return
		}

		// Add tenant_id to context
		ctx := context.WithValue(c.Request.Context(), constants.ContextKeyTenantID, uint(tenantID))
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// UserIDMiddleware extracts user_id from request headers and adds it to context (optional)
func UserIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader(constants.HeaderUserID)
		if userIDStr != "" {
			if userID, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
				ctx := context.WithValue(c.Request.Context(), constants.ContextKeyUserID, uint(userID))
				c.Request = c.Request.WithContext(ctx)
			}
		}

		c.Next()
	}
}

// TraceIDMiddleware extracts or generates trace_id for request tracing
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(constants.HeaderTraceID)
		if traceID == "" {
			// Generate a simple trace ID if not provided
			traceID = generateTraceID()
		}

		// Add trace_id to context
		ctx := context.WithValue(c.Request.Context(), constants.ContextKeyTraceID, traceID)
		c.Request = c.Request.WithContext(ctx)

		// Also set it in response header for client tracking
		c.Header(constants.HeaderTraceID, traceID)

		c.Next()
	}
}

// generateTraceID generates a simple trace ID (you may want to use UUID or more sophisticated method)
func generateTraceID() string {
	// Simple implementation - in production, use UUID or similar
	return strconv.FormatInt(int64(1000000000+len("trace")), 36)
}
