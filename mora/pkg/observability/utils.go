package observability

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// WithTrace extracts trace context information and returns trace ID and span ID
func WithTrace(ctx context.Context) (traceID, spanID string) {
	if ctx == nil {
		return "", ""
	}

	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return "", ""
	}

	return spanCtx.TraceID().String(), spanCtx.SpanID().String()
}

// GetTraceID extracts only the trace ID from context
func GetTraceID(ctx context.Context) string {
	traceID, _ := WithTrace(ctx)
	return traceID
}

// GetSpanID extracts only the span ID from context
func GetSpanID(ctx context.Context) string {
	_, spanID := WithTrace(ctx)
	return spanID
}
