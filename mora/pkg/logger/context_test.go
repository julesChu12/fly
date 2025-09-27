package logger

import (
	"context"
	"testing"
)

func TestGetTraceIDFromContext(t *testing.T) {
	t.Run("returns trace ID when present", func(t *testing.T) {
		traceID := "test-trace-123"
		ctx := context.WithValue(context.Background(), TraceIDKey, traceID)

		result := GetTraceIDFromContext(ctx)
		if result != traceID {
			t.Errorf("GetTraceIDFromContext() = %v, want %v", result, traceID)
		}
	})

	t.Run("returns empty string when not present", func(t *testing.T) {
		ctx := context.Background()
		result := GetTraceIDFromContext(ctx)
		if result != "" {
			t.Errorf("GetTraceIDFromContext() = %v, want empty string", result)
		}
	})

	t.Run("returns empty string with nil context", func(t *testing.T) {
		result := GetTraceIDFromContext(nil)
		if result != "" {
			t.Errorf("GetTraceIDFromContext() = %v, want empty string", result)
		}
	})

	t.Run("returns empty string when value is not string", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), TraceIDKey, 123)
		result := GetTraceIDFromContext(ctx)
		if result != "" {
			t.Errorf("GetTraceIDFromContext() = %v, want empty string", result)
		}
	})

	t.Run("returns empty string when key is different", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "different_key", "trace-123")
		result := GetTraceIDFromContext(ctx)
		if result != "" {
			t.Errorf("GetTraceIDFromContext() = %v, want empty string", result)
		}
	})
}

func TestWithTraceID(t *testing.T) {
	t.Run("adds trace ID to context", func(t *testing.T) {
		traceID := "test-trace-456"
		ctx := context.Background()

		newCtx := WithTraceID(ctx, traceID)
		if newCtx == nil {
			t.Fatal("WithTraceID() should return context")
		}

		result := GetTraceIDFromContext(newCtx)
		if result != traceID {
			t.Errorf("GetTraceIDFromContext() = %v, want %v", result, traceID)
		}
	})

	t.Run("overwrites existing trace ID", func(t *testing.T) {
		originalTraceID := "original-trace"
		newTraceID := "new-trace"

		ctx := WithTraceID(context.Background(), originalTraceID)
		newCtx := WithTraceID(ctx, newTraceID)

		result := GetTraceIDFromContext(newCtx)
		if result != newTraceID {
			t.Errorf("GetTraceIDFromContext() = %v, want %v", result, newTraceID)
		}
	})

	t.Run("works with empty trace ID", func(t *testing.T) {
		ctx := WithTraceID(context.Background(), "")
		result := GetTraceIDFromContext(ctx)
		if result != "" {
			t.Errorf("GetTraceIDFromContext() = %v, want empty string", result)
		}
	})

	t.Run("works with nil parent context", func(t *testing.T) {
		traceID := "test-trace"

		// Note: context.WithValue will panic with nil parent in Go
		// This test documents the expected behavior
		defer func() {
			if r := recover(); r == nil {
				t.Error("WithTraceID() with nil context should panic")
			}
		}()

		WithTraceID(nil, traceID)
	})
}

func TestTraceIDKey(t *testing.T) {
	t.Run("constant has expected value", func(t *testing.T) {
		expected := "trace_id"
		if TraceIDKey != expected {
			t.Errorf("TraceIDKey = %v, want %v", TraceIDKey, expected)
		}
	})
}

func TestContextIntegration(t *testing.T) {
	t.Run("full round trip", func(t *testing.T) {
		traceID := "integration-test-789"

		// Create context with trace ID
		ctx := WithTraceID(context.Background(), traceID)

		// Retrieve trace ID
		retrievedTraceID := GetTraceIDFromContext(ctx)

		// Verify
		if retrievedTraceID != traceID {
			t.Errorf("Round trip failed: got %v, want %v", retrievedTraceID, traceID)
		}
	})

	t.Run("context chain", func(t *testing.T) {
		// Create a chain of contexts
		baseCtx := context.Background()

		// Add some other context value
		ctx1 := context.WithValue(baseCtx, "other_key", "other_value")

		// Add trace ID
		traceID := "chain-test-abc"
		ctx2 := WithTraceID(ctx1, traceID)

		// Add another context value
		ctx3 := context.WithValue(ctx2, "another_key", "another_value")

		// Verify trace ID is still accessible
		retrievedTraceID := GetTraceIDFromContext(ctx3)
		if retrievedTraceID != traceID {
			t.Errorf("Trace ID lost in context chain: got %v, want %v", retrievedTraceID, traceID)
		}

		// Verify other values are still accessible
		if otherValue := ctx3.Value("other_key"); otherValue != "other_value" {
			t.Errorf("Other context value lost: got %v, want other_value", otherValue)
		}

		if anotherValue := ctx3.Value("another_key"); anotherValue != "another_value" {
			t.Errorf("Another context value lost: got %v, want another_value", anotherValue)
		}
	})

	t.Run("multiple trace IDs", func(t *testing.T) {
		traceID1 := "trace-1"
		traceID2 := "trace-2"

		ctx1 := WithTraceID(context.Background(), traceID1)
		ctx2 := WithTraceID(context.Background(), traceID2)

		// Verify they are independent
		result1 := GetTraceIDFromContext(ctx1)
		result2 := GetTraceIDFromContext(ctx2)

		if result1 != traceID1 {
			t.Errorf("Context 1 trace ID = %v, want %v", result1, traceID1)
		}

		if result2 != traceID2 {
			t.Errorf("Context 2 trace ID = %v, want %v", result2, traceID2)
		}

		if result1 == result2 {
			t.Error("Different contexts should have different trace IDs")
		}
	})
}
