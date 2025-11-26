package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/utils"
)

// TraceIDMiddleware 添加追踪ID中间件
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = utils.GenerateTraceID()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}