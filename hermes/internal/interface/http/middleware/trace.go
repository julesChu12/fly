package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/julesChu12/fly/hermes/pkg/constants"
)

// TraceIDMiddleware extracts or generates trace_id for request tracing
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(constants.HeaderTraceID)
		if traceID == "" {
			// Generate UUID as trace ID if not provided
			traceID = uuid.New().String()
		}

		// Add trace_id to context
		ctx := context.WithValue(c.Request.Context(), constants.ContextKeyTraceID, traceID)
		c.Request = c.Request.WithContext(ctx)

		// Also set it in response header for client tracking
		c.Header(constants.HeaderTraceID, traceID)

		c.Next()
	}
}
