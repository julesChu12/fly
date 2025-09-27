package observability

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestObservabilityInit(t *testing.T) {
	cfg := Config{
		ServiceName:  "test-service",
		ExporterType: "stdout",
		SampleRatio:  1.0,
		Environment:  "test",
	}

	cleanup, err := Init(cfg)
	if err != nil {
		t.Fatalf("failed to initialize observability: %v", err)
	}
	defer cleanup()

	// Test that tracer is available
	tracer := otel.Tracer("test")
	if tracer == nil {
		t.Fatal("tracer should not be nil")
	}

	// Test tracing
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Test trace extraction
	traceID, spanID := WithTrace(ctx)
	if traceID == "" {
		t.Error("trace ID should not be empty")
	}
	if spanID == "" {
		t.Error("span ID should not be empty")
	}
}

func TestWithTrace(t *testing.T) {
	// Test with nil context
	traceID, spanID := WithTrace(nil)
	if traceID != "" || spanID != "" {
		t.Error("trace and span IDs should be empty for nil context")
	}

	// Test with context without span
	ctx := context.Background()
	traceID, spanID = WithTrace(ctx)
	if traceID != "" || spanID != "" {
		t.Error("trace and span IDs should be empty for context without span")
	}
}

func TestGetTraceID(t *testing.T) {
	// Test with nil context
	traceID := GetTraceID(nil)
	if traceID != "" {
		t.Error("trace ID should be empty for nil context")
	}

	// Test with context without span
	ctx := context.Background()
	traceID = GetTraceID(ctx)
	if traceID != "" {
		t.Error("trace ID should be empty for context without span")
	}
}

func TestGetSpanID(t *testing.T) {
	// Test with nil context
	spanID := GetSpanID(nil)
	if spanID != "" {
		t.Error("span ID should be empty for nil context")
	}

	// Test with context without span
	ctx := context.Background()
	spanID = GetSpanID(ctx)
	if spanID != "" {
		t.Error("span ID should be empty for context without span")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ServiceName == "" {
		t.Error("default service name should not be empty")
	}
	if cfg.ExporterURL == "" {
		t.Error("default exporter URL should not be empty")
	}
	if cfg.SampleRatio <= 0 || cfg.SampleRatio > 1 {
		t.Error("default sample ratio should be between 0 and 1")
	}
}