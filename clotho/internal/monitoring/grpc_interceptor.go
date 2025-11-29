package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/metadata"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// GRPCInterceptorConfig represents configuration for gRPC monitoring
type GRPCInterceptorConfig struct {
	EnableTracing     bool          `yaml:"enable_tracing"`
	EnableMetrics     bool          `yaml:"enable_metrics"`
	EnableLogging      bool          `yaml:"enable_logging"`
	MaxRequestSize    int           `yaml:"max_request_size"`
	MaxResponseSize   int           `yaml:"max_response_size"`
	LogLevel          string        `yaml:"log_level"`
	MetricsFlushInterval time.Duration `yaml:"metrics_flush_interval"`
	EnableSampling     bool          `yaml:"enable_sampling"`
	SamplingRate       float64       `yaml:"sampling_rate"`
}

// DefaultGRPCInterceptorConfig returns default configuration
func DefaultGRPCInterceptorConfig() *GRPCInterceptorConfig {
	return &GRPCInterceptorConfig{
		EnableTracing:      true,
	EnableMetrics:      true,
		EnableLogging:      true,
	MaxRequestSize:     1024 * 1024, // 1MB
	MaxResponseSize:    1024 * 1024, // 1MB
		LogLevel:           "info",
		MetricsFlushInterval: 30 * time.Second,
		EnableSampling:     true,
		SamplingRate:        0.1, // 10%
	}
}

// GRPCMetrics holds gRPC performance metrics
type GRPCMetrics struct {
	mu                sync.RWMutex
	ServiceName       string
	MethodName        string
	RequestCount      int64
	SuccessCount      int64
	ErrorCount        int64
	TotalDuration     time.Duration
	MinDuration       time.Duration
	MaxDuration       time.Duration
	AverageDuration   time.Duration
	LastRequestTime   time.Time
	TotalRequestSize  int64
	TotalResponseSize int64
	ErrorCodes        map[codes.Code]int64
	LastUpdated       time.Time
}

// RequestInfo holds information about a gRPC request
type RequestInfo struct {
	StartTime       time.Time
	ServiceName     string
	MethodName      string
	FullMethod      string
	IsClient        bool
	Success         bool
	Error           error
	Duration        time.Duration
	RequestSize     int64
	ResponseSize    int64
	ClientIP        string
	UserAgent       string
	RequestID       string
	TraceID         string
}

// GRPCInterceptor provides comprehensive gRPC monitoring
type GRPCInterceptor struct {
	config     *GRPCInterceptorConfig
	logger     *logger.Logger
	metrics    map[string]*GRPCMetrics // key: service.method
	tracer     GRPCTracer
	recorder    RequestRecorder
	sampler    RequestSampler
	mu         sync.RWMutex
}

// NewGRPCInterceptor creates a new gRPC interceptor
func NewGRPCInterceptor(config *GRPCInterceptorConfig) *GRPCInterceptor {
	if config == nil {
		config = DefaultGRPCInterceptorConfig()
	}

	logger := logger.NewDefault()
	logger.Info("Initializing gRPC interceptor", "config", fmt.Sprintf("%+v", config))

	interceptor := &GRPCInterceptor{
		config:  config,
		logger:  logger,
		metrics: make(map[string]*GRPCMetrics),
		tracer:  NewGRPCTracer(config.EnableTracing, logger),
		recorder: NewRequestRecorder(config.EnableLogging, int64(config.MaxRequestSize), int64(config.MaxResponseSize), logger),
		sampler: NewRequestSampler(config.EnableSampling, config.SamplingRate, logger),
	}

	// Start background metrics flusher if enabled
	if config.EnableMetrics {
		go interceptor.metricsFlusher()
	}

	return interceptor
}

// UnaryClientInterceptor intercepts unary client calls
func (i *GRPCInterceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Skip sampling if not enabled or if sample rate is 0
		if i.config.EnableSampling && !i.sampler.ShouldSample(method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		requestInfo := &RequestInfo{
			StartTime:   time.Now(),
			FullMethod:  method,
			IsClient:    true,
			RequestID:   i.extractRequestID(ctx),
			RequestSize:  i.estimateSize(req),
		}

		// Extract service and method names
		requestInfo.ServiceName, requestInfo.MethodName = i.parseMethod(method)

		// Log request start
		if i.config.EnableLogging {
			i.recorder.LogRequestStart(requestInfo)
		}

		// Start tracing
		span := i.tracer.StartSpan(ctx, requestInfo)
		ctx = i.tracer.SetSpanContext(ctx, span)

		// Execute the call
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Complete request info
		requestInfo.Duration = time.Since(requestInfo.StartTime)
		requestInfo.Success = err == nil
		requestInfo.Error = err
		requestInfo.ResponseSize = i.estimateSize(reply)

		// Log request completion
		if i.config.EnableLogging {
			i.recorder.LogRequestEnd(requestInfo)
		}

		// Record metrics
		if i.config.EnableMetrics {
			i.recordMetrics(requestInfo)
		}

		// Complete tracing
		i.tracer.EndSpan(span, err)

		return err
	}
}

// UnaryServerInterceptor intercepts unary server calls
func (i *GRPCInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip sampling if not enabled or if sample rate is 0
		if i.config.EnableSampling && !i.sampler.ShouldSample(info.FullMethod) {
			return handler(ctx, req)
		}

		requestInfo := &RequestInfo{
			StartTime:   time.Now(),
			FullMethod:  info.FullMethod,
			IsClient:    false,
			RequestID:   i.extractRequestID(ctx),
			RequestSize:  i.estimateSize(req),
		}

		// Extract service and method names
		requestInfo.ServiceName, requestInfo.MethodName = i.parseMethod(info.FullMethod)

		// Log request start
		if i.config.EnableLogging {
			i.recorder.LogRequestStart(requestInfo)
		}

		// Start tracing
		span := i.tracer.StartSpan(ctx, requestInfo)
		ctx = i.tracer.SetSpanContext(ctx, span)

		// Execute the handler
		resp, err := handler(ctx, req)

		// Complete request info
		requestInfo.Duration = time.Since(requestInfo.StartTime)
		requestInfo.Success = err == nil
		requestInfo.Error = err
		requestInfo.ResponseSize = i.estimateSize(resp)

		// Log request completion
		if i.config.EnableLogging {
			i.recorder.LogRequestEnd(requestInfo)
		}

		// Record metrics
		if i.config.EnableMetrics {
			i.recordMetrics(requestInfo)
		}

		// Complete tracing
		i.tracer.EndSpan(span, err)

		return resp, err
	}
}

// StreamClientInterceptor intercepts streaming client calls
func (i *GRPCInterceptor) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// Skip sampling if not enabled or if sample rate is 0
		if i.config.EnableSampling && !i.sampler.ShouldSample(method) {
			return streamer(ctx, desc, cc, method, opts...)
		}

		requestInfo := &RequestInfo{
			StartTime:   time.Now(),
			FullMethod:  method,
			IsClient:    true,
			RequestID:   i.extractRequestID(ctx),
		}

		// Extract service and method names
		requestInfo.ServiceName, requestInfo.MethodName = i.parseMethod(method)

		// Log request start
		if i.config.EnableLogging {
			i.recorder.LogStreamStart(requestInfo, "client")
		}

		// Start tracing
		span := i.tracer.StartSpan(ctx, requestInfo)
		ctx = i.tracer.SetSpanContext(ctx, span)

		// Create the stream
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			requestInfo.Success = false
			requestInfo.Error = err
			requestInfo.Duration = time.Since(requestInfo.StartTime)

			// Log error
			if i.config.EnableLogging {
				i.recorder.LogStreamError(requestInfo, err)
			}

			// Record error metrics
			if i.config.EnableMetrics {
				i.recordMetrics(requestInfo)
			}

			// Complete tracing
			i.tracer.EndSpan(span, err)

			return nil, err
		}

		// Wrap the stream for monitoring
		return &MonitoredClientStream{
			ClientStream:    clientStream,
			RequestInfo:     requestInfo,
			Interceptor:     i,
			Span:           span,
			ctx:            ctx,
		}, nil
	}
}

// StreamServerInterceptor intercepts streaming server calls
func (i *GRPCInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip sampling if not enabled or if sample rate is 0
		if i.config.EnableSampling && !i.sampler.ShouldSample(info.FullMethod) {
			return handler(srv, ss)
		}

		requestInfo := &RequestInfo{
			StartTime:   time.Now(),
			FullMethod:  info.FullMethod,
			IsClient:    false,
			RequestID:   i.extractRequestID(ss.Context()),
		}

		// Extract service and method names
		requestInfo.ServiceName, requestInfo.MethodName = i.parseMethod(info.FullMethod)

		// Log request start
		if i.config.EnableLogging {
			i.recorder.LogStreamStart(requestInfo, "server")
		}

		// Start tracing
		span := i.tracer.StartSpan(ss.Context(), requestInfo)
		ctx := i.tracer.SetSpanContext(ss.Context(), span)

		// Wrap the stream for monitoring
		wrappedStream := &MonitoredServerStream{
			ServerStream:  ss,
			RequestInfo:    requestInfo,
			Interceptor:    i,
			Span:          span,
			ctx:           ctx,
		}

		// Execute the handler
		err := handler(srv, wrappedStream)

		// Complete request info
		requestInfo.Duration = time.Since(requestInfo.StartTime)
		requestInfo.Success = err == nil
		requestInfo.Error = err

		// Log completion
		if i.config.EnableLogging {
			if err != nil {
				i.recorder.LogStreamError(requestInfo, err)
			} else {
				i.recorder.LogStreamEnd(requestInfo)
			}
		}

		// Record metrics
		if i.config.EnableMetrics {
			i.recordMetrics(requestInfo)
		}

		// Complete tracing
		i.tracer.EndSpan(span, err)

		return err
	}
}

// MonitoredClientStream wraps a client stream for monitoring
type MonitoredClientStream struct {
	grpc.ClientStream
	RequestInfo *RequestInfo
	Interceptor *GRPCInterceptor
	Span       Span
	ctx         context.Context
}

func (m *MonitoredClientStream) SendMsg(msg interface{}) error {
	size := m.Interceptor.estimateSize(msg)
	m.RequestInfo.ResponseSize += size
	return m.ClientStream.SendMsg(msg)
}

func (m *MonitoredClientStream) RecvMsg(msg interface{}) error {
	size := m.Interceptor.estimateSize(msg)
	m.RequestInfo.RequestSize += size
	return m.ClientStream.RecvMsg(msg)
}

func (m *MonitoredClientStream) CloseSend() error {
	return m.ClientStream.CloseSend()
}

func (m *MonitoredClientStream) Header() (metadata.MD, error) {
	return m.ClientStream.Header()
}

func (m *MonitoredClientStream) Trailer() metadata.MD {
	return m.ClientStream.Trailer()
}

func (m *MonitoredClientStream) Context() context.Context {
	return m.ctx
}

// MonitoredServerStream wraps a server stream for monitoring
type MonitoredServerStream struct {
	grpc.ServerStream
	RequestInfo *RequestInfo
	Interceptor *GRPCInterceptor
	Span       Span
	ctx         context.Context
}

func (m *MonitoredServerStream) SendMsg(msg interface{}) error {
	size := m.Interceptor.estimateSize(msg)
	m.RequestInfo.ResponseSize += size
	return m.ServerStream.SendMsg(msg)
}

func (m *MonitoredServerStream) RecvMsg(msg interface{}) error {
	size := m.Interceptor.estimateSize(msg)
	m.RequestInfo.RequestSize += size
	return m.ServerStream.RecvMsg(msg)
}

func (m *MonitoredServerStream) SetHeader(md metadata.MD) error {
	return m.ServerStream.SetHeader(md)
}

func (m *MonitoredServerStream) SendHeader(md metadata.MD) error {
	return m.ServerStream.SendHeader(md)
}

func (m *MonitoredServerStream) SetTrailer(md metadata.MD) {
	m.ServerStream.SetTrailer(md)
}

func (m *MonitoredServerStream) Context() context.Context {
	return m.ctx
}

// parseMethod extracts service and method names from full gRPC method
func (i *GRPCInterceptor) parseMethod(fullMethod string) (service, method string) {
	// Parse "/package.Service/Method" format
	parts := make([]string, 0)
	start := 0
	for i := 1; i < len(fullMethod); i++ {
		if fullMethod[i] == '/' {
			if start > 0 {
				parts = append(parts, fullMethod[start:i])
			}
			start = i + 1
		}
	}
	if start < len(fullMethod) {
		parts = append(parts, fullMethod[start:])
	}

	if len(parts) >= 3 {
		return parts[1], parts[2]
	}
	return "unknown", "unknown"
}

// estimateSize estimates the size of a message
func (i *GRPCInterceptor) estimateSize(msg interface{}) int64 {
	if msg == nil {
		return 0
	}

	// This is a simplified estimation - in production you might want to use protobuf.Size()
	switch v := msg.(type) {
	case []byte:
		return int64(len(v))
	case string:
		return int64(len(v))
	default:
		return 1024 // Default 1KB estimate
	}
}

// extractRequestID extracts or generates a request ID from context
func (i *GRPCInterceptor) extractRequestID(ctx context.Context) string {
	// Try to get request ID from context
	if ctx != nil {
		// In a real implementation, you might get this from headers, metadata, etc.
		// For now, generate a simple ID
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// recordMetrics records gRPC call metrics
func (i *GRPCInterceptor) recordMetrics(info *RequestInfo) {
	key := fmt.Sprintf("%s.%s", info.ServiceName, info.MethodName)

	i.mu.Lock()
	defer i.mu.Unlock()

	metrics, exists := i.metrics[key]
	if !exists {
		metrics = &GRPCMetrics{
			ServiceName: info.ServiceName,
			MethodName:  info.MethodName,
			ErrorCodes:  make(map[codes.Code]int64),
		}
		i.metrics[key] = metrics
	}

	// Update metrics
	metrics.RequestCount++
	if info.Success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
		if info.Error != nil {
			if grpcErr, ok := status.FromError(info.Error); ok {
				metrics.ErrorCodes[grpcErr.Code()]++
			}
		}
	}

	// Update duration metrics
	if metrics.RequestCount == 1 {
		metrics.MinDuration = info.Duration
		metrics.MaxDuration = info.Duration
		metrics.AverageDuration = info.Duration
	} else {
		if info.Duration < metrics.MinDuration {
			metrics.MinDuration = info.Duration
		}
		if info.Duration > metrics.MaxDuration {
			metrics.MaxDuration = info.Duration
		}
		// Calculate moving average
		metrics.AverageDuration = time.Duration(
			(float64(metrics.AverageDuration)*0.9 + float64(info.Duration)*0.1),
		)
	}

	metrics.TotalDuration += info.Duration
	metrics.TotalRequestSize += info.RequestSize
	metrics.TotalResponseSize += info.ResponseSize
	metrics.LastRequestTime = info.StartTime
	metrics.LastUpdated = time.Now()
}

// metricsFlusher periodically flushes metrics to external systems
func (i *GRPCInterceptor) metricsFlusher() {
	ticker := time.NewTicker(i.config.MetricsFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.flushMetrics()
		}
	}
}

// flushMetrics flushes current metrics to logging
func (i *GRPCInterceptor) flushMetrics() {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if len(i.metrics) == 0 {
		return
	}

	i.logger.Info("gRPC Metrics Report",
		"total_services", len(i.metrics),
		"total_requests", i.getTotalRequests(),
		"total_errors", i.getTotalErrors(),
		"avg_success_rate", i.getAverageSuccessRate(),
	)

	for _, metrics := range i.metrics {
		if metrics.RequestCount > 0 {
			successRate := float64(metrics.SuccessCount) / float64(metrics.RequestCount) * 100
			avgLatency := metrics.AverageDuration.Milliseconds()

			i.logger.Debug("Service Metrics",
				"service", metrics.ServiceName,
				"method", metrics.MethodName,
				"requests", metrics.RequestCount,
				"success_rate", fmt.Sprintf("%.2f%%", successRate),
				"avg_latency_ms", avgLatency,
				"total_duration_ms", metrics.TotalDuration.Milliseconds(),
			)
		}
	}
}

// getTotalRequests returns total requests across all services
func (i *GRPCInterceptor) getTotalRequests() int64 {
	i.mu.RLock()
	defer i.mu.RUnlock()

	total := int64(0)
	for _, metrics := range i.metrics {
		total += metrics.RequestCount
	}
	return total
}

// getTotalErrors returns total errors across all services
func (i *GRPCInterceptor) getTotalErrors() int64 {
	i.mu.RLock()
	defer i.mu.RUnlock()

	total := int64(0)
	for _, metrics := range i.metrics {
		total += metrics.ErrorCount
	}
	return total
}

// getAverageSuccessRate returns average success rate across all services
func (i *GRPCInterceptor) getAverageSuccessRate() float64 {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if len(i.metrics) == 0 {
		return 0.0
	}

	totalSuccess := int64(0)
	totalRequests := int64(0)

	for _, metrics := range i.metrics {
		totalSuccess += metrics.SuccessCount
		totalRequests += metrics.RequestCount
	}

	if totalRequests == 0 {
		return 0.0
	}

	return float64(totalSuccess) / float64(totalRequests) * 100
}

// GetMetrics returns a copy of current metrics
func (i *GRPCInterceptor) GetMetrics() map[string]*GRPCMetrics {
	i.mu.RLock()
	defer i.mu.RUnlock()

	metrics := make(map[string]*GRPCMetrics)
	for key, metric := range i.metrics {
		// Create a new copy without copying the mutex
		metric.mu.RLock()
		metricCopy := &GRPCMetrics{
			ServiceName:       metric.ServiceName,
			MethodName:        metric.MethodName,
			RequestCount:      metric.RequestCount,
			SuccessCount:      metric.SuccessCount,
			ErrorCount:        metric.ErrorCount,
			TotalDuration:     metric.TotalDuration,
			MinDuration:       metric.MinDuration,
			MaxDuration:       metric.MaxDuration,
			AverageDuration:   metric.AverageDuration,
			LastRequestTime:   metric.LastRequestTime,
			TotalRequestSize:  metric.TotalRequestSize,
			TotalResponseSize: metric.TotalResponseSize,
			ErrorCodes:        make(map[codes.Code]int64),
			LastUpdated:       metric.LastUpdated,
		}
		// Copy error codes
		for code, count := range metric.ErrorCodes {
			metricCopy.ErrorCodes[code] = count
		}
		metric.mu.RUnlock()
		metrics[key] = metricCopy
	}
	return metrics
}

// ResetMetrics resets all metrics
func (i *GRPCInterceptor) ResetMetrics() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.metrics = make(map[string]*GRPCMetrics)
	i.logger.Info("All gRPC metrics reset")
}