package monitoring

import (
	"context"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// Span represents a distributed tracing span
type Span interface {
	// Context returns the span context
	Context() SpanContext

	// SetOperationName sets the operation name
	SetOperationName(name string)

	// SetTag sets a tag on the span
	SetTag(key string, value interface{})

	// SetBaggageItem sets a baggage item
	SetBaggageItem(key, value string)

	// Finish finishes the span
	Finish()

	// FinishWithOptions finishes the span with options
	FinishWithOptions(opts FinishOptions)
}

// SpanContext represents a span context
type SpanContext struct {
	TraceID      string
	SpanID       string
	Sampled      bool
	Baggage      map[string]string
}

// FinishOptions represents options for finishing a span
type FinishOptions struct {
	FinishTime time.Time
	LogRecords []LogRecord
}

// LogRecord represents a log record
type LogRecord struct {
	Timestamp time.Time
	Level     string
	Message   string
	Fields    map[string]interface{}
}

// GRPCTracer represents a distributed tracer for gRPC
type GRPCTracer interface {
	// StartSpan starts a new span
	StartSpan(ctx context.Context, requestInfo *RequestInfo) Span

	// EndSpan ends a span with optional error
	EndSpan(span Span, err error)

	// SetSpanContext sets the span context in the request context
	SetSpanContext(ctx context.Context, span Span) context.Context

	// ExtractSpanContext extracts span context from request context
	ExtractSpanContext(ctx context.Context) *SpanContext

	// InjectSpanContext injects span context into request context
	InjectSpanContext(ctx context.Context, spanCtx *SpanContext) context.Context
}

// DefaultGRPCTracer represents a default implementation of GRPCTracer
type DefaultGRPCTracer struct {
	enabled bool
	logger  *logger.Logger
}

// NewGRPCTracer creates a new gRPC tracer
func NewGRPCTracer(enabled bool, logger *logger.Logger) GRPCTracer {
	return &DefaultGRPCTracer{
		enabled: enabled,
		logger:  logger,
	}
}

// StartSpan starts a new span
func (t *DefaultGRPCTracer) StartSpan(ctx context.Context, requestInfo *RequestInfo) Span {
	if !t.enabled {
		return &NoopSpan{}
	}

	spanCtx := &SpanContext{
		TraceID: requestInfo.TraceID,
		SpanID:  requestInfo.RequestID,
		Sampled: true,
		Baggage: make(map[string]string),
	}

	span := &DefaultSpan{
		context:    spanCtx,
		operation:  requestInfo.FullMethod,
		startTime:  time.Now(),
		logger:     t.logger,
		requestInfo: requestInfo,
	}

	t.logger.Debug("Started gRPC span",
		"trace_id", spanCtx.TraceID,
		"span_id", spanCtx.SpanID,
		"operation", requestInfo.FullMethod,
	)

	return span
}

// SetSpanContext sets the span context in the request context
func (t *DefaultGRPCTracer) SetSpanContext(ctx context.Context, span Span) context.Context {
	if !t.enabled {
		return ctx
	}

	// In a real implementation, this would integrate with opentelemetry or similar
	// For now, we'll store the span context in the context using a simple key
	if defaultSpan, ok := span.(*DefaultSpan); ok {
		return context.WithValue(ctx, "span_context", defaultSpan.context)
	}

	return ctx
}

// ExtractSpanContext extracts span context from request context
func (t *DefaultGRPCTracer) ExtractSpanContext(ctx context.Context) *SpanContext {
	if !t.enabled {
		return &SpanContext{}
	}

	if spanCtx, ok := ctx.Value("span_context").(*SpanContext); ok {
		return spanCtx
	}

	return &SpanContext{}
}

// InjectSpanContext injects span context into request context
func (t *DefaultGRPCTracer) InjectSpanContext(ctx context.Context, spanCtx *SpanContext) context.Context {
	if !t.enabled {
		return ctx
	}

	return context.WithValue(ctx, "span_context", spanCtx)
}

// EndSpan ends a span with optional error
func (t *DefaultGRPCTracer) EndSpan(span Span, err error) {
	if span == nil {
		return
	}

	// Set error information if provided
	if err != nil {
		if defaultSpan, ok := span.(*DefaultSpan); ok {
			defaultSpan.SetTag("error", true)
			defaultSpan.SetTag("error.message", err.Error())
		}
	}

	span.Finish()
}

// DefaultSpan represents a default span implementation
type DefaultSpan struct {
	context     *SpanContext
	operation   string
	startTime   time.Time
	logger      *logger.Logger
	requestInfo *RequestInfo
	tags        map[string]interface{}
	logs        []LogRecord
	finished    bool
}

// Context returns the span context
func (s *DefaultSpan) Context() SpanContext {
	if s.context == nil {
		return SpanContext{}
	}
	return *s.context
}

// SetOperationName sets the operation name
func (s *DefaultSpan) SetOperationName(name string) {
	s.operation = name
}

// SetTag sets a tag on the span
func (s *DefaultSpan) SetTag(key string, value interface{}) {
	if s.tags == nil {
		s.tags = make(map[string]interface{})
	}
	s.tags[key] = value
}

// SetBaggageItem sets a baggage item
func (s *DefaultSpan) SetBaggageItem(key, value string) {
	if s.context != nil && s.context.Baggage != nil {
		s.context.Baggage[key] = value
	}
}

// Finish finishes the span
func (s *DefaultSpan) Finish() {
	s.FinishWithOptions(FinishOptions{
		FinishTime: time.Now(),
	})
}

// FinishWithOptions finishes the span with options
func (s *DefaultSpan) FinishWithOptions(opts FinishOptions) {
	if s.finished {
		return
	}

	s.finished = true
	finishTime := opts.FinishTime
	if finishTime.IsZero() {
		finishTime = time.Now()
	}

	duration := finishTime.Sub(s.startTime)

	// Add completion log record
	logRecord := LogRecord{
		Timestamp: finishTime,
		Level:     "info",
		Message:   "Span completed",
		Fields: map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
			"success":     s.requestInfo.Success,
		},
	}

	if s.requestInfo.Error != nil {
		logRecord.Fields["error"] = s.requestInfo.Error.Error()
		logRecord.Level = "error"
		logRecord.Message = "Span completed with error"
	}

	s.logs = append(s.logs, logRecord)

	// Log span completion
	if !s.requestInfo.Success {
		s.logger.Error("gRPC span completed",
			"trace_id", s.context.TraceID,
			"span_id", s.context.SpanID,
			"operation", s.operation,
			"duration_ms", duration.Milliseconds(),
			"success", s.requestInfo.Success,
			"tags", s.tags,
		)
	} else {
		s.logger.Debug("gRPC span completed",
			"trace_id", s.context.TraceID,
			"span_id", s.context.SpanID,
			"operation", s.operation,
			"duration_ms", duration.Milliseconds(),
			"success", s.requestInfo.Success,
			"tags", s.tags,
		)
	}
}

// NoopSpan represents a no-operation span
type NoopSpan struct{}

// Context returns the span context
func (s *NoopSpan) Context() SpanContext {
	return SpanContext{}
}

// SetOperationName sets the operation name
func (s *NoopSpan) SetOperationName(name string) {
	// No-op
}

// SetTag sets a tag on the span
func (s *NoopSpan) SetTag(key string, value interface{}) {
	// No-op
}

// SetBaggageItem sets a baggage item
func (s *NoopSpan) SetBaggageItem(key, value string) {
	// No-op
}

// Finish finishes the span
func (s *NoopSpan) Finish() {
	// No-op
}

// FinishWithOptions finishes the span with options
func (s *NoopSpan) FinishWithOptions(opts FinishOptions) {
	// No-op
}

// NoopGRPCTracer represents a no-operation gRPC tracer
type NoopGRPCTracer struct{}

// NewNoopGRPCTracer creates a new no-op gRPC tracer
func NewNoopGRPCTracer() GRPCTracer {
	return &NoopGRPCTracer{}
}

// StartSpan creates a no-op span
func (t *NoopGRPCTracer) StartSpan(ctx context.Context, requestInfo *RequestInfo) Span {
	return &NoopSpan{}
}

// EndSpan does nothing
func (t *NoopGRPCTracer) EndSpan(span Span, err error) {
	// No-op
}

// SetSpanContext does nothing
func (t *NoopGRPCTracer) SetSpanContext(ctx context.Context, span Span) context.Context {
	return ctx
}

// ExtractSpanContext returns empty context
func (t *NoopGRPCTracer) ExtractSpanContext(ctx context.Context) *SpanContext {
	return &SpanContext{}
}

// InjectSpanContext does nothing
func (t *NoopGRPCTracer) InjectSpanContext(ctx context.Context, spanCtx *SpanContext) context.Context {
	return ctx
}