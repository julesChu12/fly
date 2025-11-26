package monitoring

import (
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// RequestRecorder records gRPC request information
type RequestRecorder interface {
	// LogRequestStart logs the start of a request
	LogRequestStart(requestInfo *RequestInfo)

	// LogRequestEnd logs the end of a request
	LogRequestEnd(requestInfo *RequestInfo)

	// LogStreamStart logs the start of a streaming request
	LogStreamStart(requestInfo *RequestInfo, direction string)

	// LogStreamEnd logs the end of a streaming request
	LogStreamEnd(requestInfo *RequestInfo)

	// LogStreamError logs a streaming request error
	LogStreamError(requestInfo *RequestInfo, err error)
}

// DefaultRequestRecorder represents a default implementation of RequestRecorder
type DefaultRequestRecorder struct {
	enabled           bool
	logger            *logger.Logger
	maxRequestSize    int64
	maxResponseSize   int64
	truncateThreshold int64
}

// NewRequestRecorder creates a new request recorder
func NewRequestRecorder(enabled bool, maxRequestSize, maxResponseSize int64, logger *logger.Logger) RequestRecorder {
	return &DefaultRequestRecorder{
		enabled:           enabled,
		logger:            logger,
		maxRequestSize:    maxRequestSize,
		maxResponseSize:   maxResponseSize,
		truncateThreshold: 1024, // Truncate after 1KB
	}
}

// LogRequestStart logs the start of a request
func (r *DefaultRequestRecorder) LogRequestStart(requestInfo *RequestInfo) {
	if !r.enabled {
		return
	}

	fields := map[string]interface{}{
		"request_id":   requestInfo.RequestID,
		"trace_id":     requestInfo.TraceID,
		"service":      requestInfo.ServiceName,
		"method":       requestInfo.MethodName,
		"full_method":  requestInfo.FullMethod,
		"is_client":    requestInfo.IsClient,
		"request_size": requestInfo.RequestSize,
	}

	// Add client IP if available
	if requestInfo.ClientIP != "" {
		fields["client_ip"] = requestInfo.ClientIP
	}

	// Add user agent if available
	if requestInfo.UserAgent != "" {
		fields["user_agent"] = r.truncateString(requestInfo.UserAgent, r.truncateThreshold)
	}

	r.logger.Info("gRPC request started", fields)
}

// LogRequestEnd logs the end of a request
func (r *DefaultRequestRecorder) LogRequestEnd(requestInfo *RequestInfo) {
	if !r.enabled {
		return
	}

	fields := map[string]interface{}{
		"request_id":    requestInfo.RequestID,
		"trace_id":      requestInfo.TraceID,
		"service":       requestInfo.ServiceName,
		"method":        requestInfo.MethodName,
		"full_method":   requestInfo.FullMethod,
		"is_client":     requestInfo.IsClient,
		"duration_ms":   requestInfo.Duration.Milliseconds(),
		"success":       requestInfo.Success,
		"request_size":  requestInfo.RequestSize,
		"response_size": requestInfo.ResponseSize,
	}

	if !requestInfo.Success && requestInfo.Error != nil {
		fields["error"] = requestInfo.Error.Error()
	}

	if !requestInfo.Success {
		r.logger.Error("gRPC request completed", fields)
	} else {
		r.logger.Info("gRPC request completed", fields)
	}
}

// LogStreamStart logs the start of a streaming request
func (r *DefaultRequestRecorder) LogStreamStart(requestInfo *RequestInfo, direction string) {
	if !r.enabled {
		return
	}

	fields := map[string]interface{}{
		"request_id":  requestInfo.RequestID,
		"trace_id":    requestInfo.TraceID,
		"service":     requestInfo.ServiceName,
		"method":      requestInfo.MethodName,
		"full_method": requestInfo.FullMethod,
		"is_client":   requestInfo.IsClient,
		"direction":   direction,
		"stream_type": "start",
	}

	r.logger.Info("gRPC stream started", fields)
}

// LogStreamEnd logs the end of a streaming request
func (r *DefaultRequestRecorder) LogStreamEnd(requestInfo *RequestInfo) {
	if !r.enabled {
		return
	}

	fields := map[string]interface{}{
		"request_id":  requestInfo.RequestID,
		"trace_id":    requestInfo.TraceID,
		"service":     requestInfo.ServiceName,
		"method":      requestInfo.MethodName,
		"full_method": requestInfo.FullMethod,
		"is_client":   requestInfo.IsClient,
		"duration_ms": requestInfo.Duration.Milliseconds(),
		"stream_type": "end",
	}

	r.logger.Info("gRPC stream completed", fields)
}

// LogStreamError logs a streaming request error
func (r *DefaultRequestRecorder) LogStreamError(requestInfo *RequestInfo, err error) {
	if !r.enabled {
		return
	}

	fields := map[string]interface{}{
		"request_id":  requestInfo.RequestID,
		"trace_id":    requestInfo.TraceID,
		"service":     requestInfo.ServiceName,
		"method":      requestInfo.MethodName,
		"full_method": requestInfo.FullMethod,
		"is_client":   requestInfo.IsClient,
		"duration_ms": requestInfo.Duration.Milliseconds(),
		"error":       err.Error(),
		"stream_type": "error",
	}

	r.logger.Error("gRPC stream error", fields)
}

// truncateString truncates a string to the specified maximum length
func (r *DefaultRequestRecorder) truncateString(s string, maxLen int64) string {
	if int64(len(s)) <= maxLen {
		return s
	}

	if maxLen < 20 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}

// NoopRequestRecorder represents a no-operation request recorder
type NoopRequestRecorder struct{}

// NewNoopRequestRecorder creates a new no-op request recorder
func NewNoopRequestRecorder() RequestRecorder {
	return &NoopRequestRecorder{}
}

// LogRequestStart logs the start of a request (no-op)
func (r *NoopRequestRecorder) LogRequestStart(requestInfo *RequestInfo) {
	// No-op
}

// LogRequestEnd logs the end of a request (no-op)
func (r *NoopRequestRecorder) LogRequestEnd(requestInfo *RequestInfo) {
	// No-op
}

// LogStreamStart logs the start of a streaming request (no-op)
func (r *NoopRequestRecorder) LogStreamStart(requestInfo *RequestInfo, direction string) {
	// No-op
}

// LogStreamEnd logs the end of a streaming request (no-op)
func (r *NoopRequestRecorder) LogStreamEnd(requestInfo *RequestInfo) {
	// No-op
}

// LogStreamError logs a streaming request error (no-op)
func (r *NoopRequestRecorder) LogStreamError(requestInfo *RequestInfo, err error) {
	// No-op
}