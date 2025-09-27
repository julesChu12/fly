package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

const (
	// TraceIDKey is the key used to store trace ID in context
	TraceIDKey = "trace_id"
)

// GetTraceIDFromContext extracts trace ID from context
// First tries to get from OpenTelemetry context, then falls back to custom context value
func GetTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try to get trace ID from OpenTelemetry context first
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}

	// Fall back to custom context value
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// GetSpanIDFromContext extracts span ID from OpenTelemetry context
func GetSpanIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.SpanID().String()
	}
	return ""
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}
