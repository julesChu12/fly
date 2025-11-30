package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/utils"
)

// TraceIDMiddleware 追踪ID中间件
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否已有trace_id
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = utils.GenerateTraceID()
		}

		// 设置trace_id到响应头和上下文
		c.Header("X-Trace-ID", traceID)
		c.Set("trace_id", traceID)

		c.Next()
	}
}
