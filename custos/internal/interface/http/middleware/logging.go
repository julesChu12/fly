package middleware

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	moralogger "github.com/julesChu12/fly/mora/pkg/logger"
)

// LoggingMiddleware provides structured logging for all HTTP requests
func LoggingMiddleware(logger *moralogger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Extract user info from context if available
		var userID uint
		var username string
		if user, exists := param.Keys["user"]; exists {
			if u, ok := user.(*entity.User); ok {
				userID = u.ID
				username = u.Username
			}
		}

		// Extract trace ID
		traceID, exists := param.Keys["trace_id"]
		if !exists {
			traceID = ""
		}

		fields := map[string]interface{}{
			"timestamp":     param.TimeStamp.Format(time.RFC3339),
			"status":        param.StatusCode,
			"latency":       param.Latency.String(),
			"client_ip":     param.ClientIP,
			"method":        param.Method,
			"path":          param.Path,
			"user_agent":    param.Request.UserAgent(),
			"request_id":    traceID,
			"response_size": param.BodySize,
		}

		if userID > 0 {
			fields["user_id"] = userID
			fields["username"] = username
		}

		if param.ErrorMessage != "" {
			fields["error"] = param.ErrorMessage
		}

		logger.WithFields(fields).Info("HTTP %s %s - %d", param.Method, param.Path, param.StatusCode)

		return "" // We don't need Gin's default output since we're using structured logging
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to Gin context
		c.Set("trace_id", requestID)
		c.Set("request_id", requestID)

		// Add to Go context for downstream services
		ctx := context.WithValue(c.Request.Context(), "trace_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// AuditLogMiddleware logs important security and admin actions
func AuditLogMiddleware(logger *moralogger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only log certain sensitive endpoints
		if !shouldAuditPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()

		// Capture request body for audit (be careful with sensitive data)
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		c.Next()

		// Extract user info
		var userID uint
		var username string
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(*entity.User); ok {
				userID = u.ID
				username = u.Username
			}
		}

		// Log audit event
		fields := map[string]interface{}{
			"event_type":  "api_access",
			"timestamp":   start.Format(time.RFC3339),
			"user_id":     userID,
			"username":    username,
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
			"latency_ms":  time.Since(start).Milliseconds(),
			"request_id":  c.GetString("request_id"),
		}

		// Add query parameters for GET requests
		if c.Request.Method == "GET" && len(c.Request.URL.RawQuery) > 0 {
			fields["query_params"] = c.Request.URL.RawQuery
		}

		// Add form data size for POST/PUT requests (don't log actual content for security)
		if len(requestBody) > 0 {
			fields["request_body_size"] = len(requestBody)
		}

		logger.WithFields(fields).Info("AUDIT: %s %s by user %s (%d) - %d",
			c.Request.Method,
			c.Request.URL.Path,
			username,
			userID,
			c.Writer.Status(),
		)
	}
}

// shouldAuditPath determines if a path should be audited
func shouldAuditPath(path string) bool {
	auditPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/logout",
		"/api/v1/auth/register",
		"/api/v1/admin/",
		"/api/v1/oauth/",
	}

	for _, auditPath := range auditPaths {
		if len(path) >= len(auditPath) && path[:len(auditPath)] == auditPath {
			return true
		}
	}
	return false
}

// SecurityEventLogger logs security-related events
type SecurityEventLogger struct {
	logger *moralogger.Logger
}

func NewSecurityEventLogger(logger *moralogger.Logger) *SecurityEventLogger {
	return &SecurityEventLogger{logger: logger}
}

func (s *SecurityEventLogger) LogAuthAttempt(ctx context.Context, username, clientIP, userAgent string, success bool, reason string) {
	fields := map[string]interface{}{
		"event_type": "auth_attempt",
		"timestamp":  time.Now().Format(time.RFC3339),
		"username":   username,
		"client_ip":  clientIP,
		"user_agent": userAgent,
		"success":    success,
		"reason":     reason,
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["request_id"] = traceID
	}

	if success {
		s.logger.WithFields(fields).Info("SECURITY: Authentication SUCCESS for user %s from %s", username, clientIP)
	} else {
		s.logger.WithFields(fields).Warn("SECURITY: Authentication FAILED for user %s from %s", username, clientIP)
	}
}

func (s *SecurityEventLogger) LogTokenValidation(ctx context.Context, userID uint, success bool, reason string) {
	fields := map[string]interface{}{
		"event_type": "token_validation",
		"timestamp":  time.Now().Format(time.RFC3339),
		"user_id":    userID,
		"success":    success,
		"reason":     reason,
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["request_id"] = traceID
	}

	s.logger.WithFields(fields).Info("SECURITY: Token validation %s for user %d: %s",
		map[bool]string{true: "SUCCESS", false: "FAILED"}[success], userID, reason)
}

func (s *SecurityEventLogger) LogPermissionCheck(ctx context.Context, userID uint, resource, action string, allowed bool) {
	fields := map[string]interface{}{
		"event_type": "permission_check",
		"timestamp":  time.Now().Format(time.RFC3339),
		"user_id":    userID,
		"resource":   resource,
		"action":     action,
		"allowed":    allowed,
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["request_id"] = traceID
	}

	s.logger.WithFields(fields).Debug("SECURITY: Permission check for user %d on %s:%s - %s",
		userID, resource, action, map[bool]string{true: "ALLOWED", false: "DENIED"}[allowed])
}

func (s *SecurityEventLogger) LogAdminAction(ctx context.Context, adminID uint, adminUsername, action, targetType string, targetID uint) {
	fields := map[string]interface{}{
		"event_type":     "admin_action",
		"timestamp":      time.Now().Format(time.RFC3339),
		"admin_id":       adminID,
		"admin_username": adminUsername,
		"action":         action,
		"target_type":    targetType,
		"target_id":      targetID,
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["request_id"] = traceID
	}

	s.logger.WithFields(fields).Info("ADMIN: %s performed %s on %s %d",
		adminUsername, action, targetType, targetID)
}
