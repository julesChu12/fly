package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// MetricsCollector collects and manages gRPC metrics
type MetricsCollector interface {
	// RecordRequest records a request metric
	RecordRequest(service, method string, duration time.Duration, success bool, err error)

	// RecordStream records a stream metric
	RecordStream(service, method string, duration time.Duration, success bool)

	// GetServiceMetrics returns metrics for a specific service
	GetServiceMetrics(service string) *ServiceMetrics

	// GetAllMetrics returns all collected metrics
	GetAllMetrics() map[string]*ServiceMetrics

	// ResetMetrics resets all metrics
	ResetMetrics()

	// StartMetricsExporter starts the background metrics exporter
	StartMetricsExporter(ctx context.Context) error

	// StopMetricsExporter stops the metrics exporter
	StopMetricsExporter()
}

// ServiceMetrics represents metrics for a specific service
type ServiceMetrics struct {
	ServiceName      string                    `json:"service_name"`
	Methods          map[string]*MethodMetrics `json:"methods"`
	TotalRequests    int64                     `json:"total_requests"`
	SuccessRequests  int64                     `json:"success_requests"`
	ErrorRequests    int64                     `json:"error_requests"`
	TotalDuration    time.Duration             `json:"total_duration"`
	AverageLatency   time.Duration             `json:"average_latency"`
	MinLatency       time.Duration             `json:"min_latency"`
	MaxLatency       time.Duration             `json:"max_latency"`
	LastRequestTime  time.Time                 `json:"last_request_time"`
	ErrorRates       map[string]int64          `json:"error_rates"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
}

// MethodMetrics represents metrics for a specific method
type MethodMetrics struct {
	MethodName       string        `json:"method_name"`
	RequestCount     int64         `json:"request_count"`
	SuccessCount     int64         `json:"success_count"`
	ErrorCount       int64         `json:"error_count"`
	TotalDuration    time.Duration `json:"total_duration"`
	AverageLatency   time.Duration `json:"average_latency"`
	MinLatency       time.Duration `json:"min_latency"`
	MaxLatency       time.Duration `json:"max_latency"`
	LastRequestTime  time.Time     `json:"last_request_time"`
	ErrorCodes       map[string]int64 `json:"error_codes"`
	StreamMetrics    *StreamMetrics `json:"stream_metrics,omitempty"`
}

// StreamMetrics represents metrics for streaming operations
type StreamMetrics struct {
	StreamCount      int64         `json:"stream_count"`
	SuccessStreams   int64         `json:"success_streams"`
	ErrorStreams     int64         `json:"error_streams"`
	AverageDuration  time.Duration `json:"average_duration"`
	MinDuration      time.Duration `json:"min_duration"`
	MaxDuration      time.Duration `json:"max_duration"`
	BytesReceived    int64         `json:"bytes_received"`
	BytesSent        int64         `json:"bytes_sent"`
	LastStreamTime   time.Time     `json:"last_stream_time"`
}

// DefaultMetricsCollector represents a default implementation of MetricsCollector
type DefaultMetricsCollector struct {
	metrics map[string]*ServiceMetrics
	mu      sync.RWMutex
	logger  *logger.Logger

	// Exporter configuration
	exportInterval time.Duration
	exporterCtx    context.Context
	exporterCancel context.CancelFunc
	exporterActive bool
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(log *logger.Logger) MetricsCollector {
	return &DefaultMetricsCollector{
		metrics:        make(map[string]*ServiceMetrics),
		logger:         log,
		exportInterval: 30 * time.Second, // Default export interval
	}
}

// RecordRequest records a request metric
func (c *DefaultMetricsCollector) RecordRequest(service, method string, duration time.Duration, success bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get or create service metrics
	serviceMetrics, exists := c.metrics[service]
	if !exists {
		serviceMetrics = &ServiceMetrics{
			ServiceName:     service,
			Methods:         make(map[string]*MethodMetrics),
			TotalDuration:   0,
			MinLatency:      duration,
			MaxLatency:      duration,
			AverageLatency:  duration,
			ErrorRates:      make(map[string]int64),
			CreatedAt:       time.Now(),
		}
		c.metrics[service] = serviceMetrics
	}

	// Get or create method metrics
	methodMetrics, exists := serviceMetrics.Methods[method]
	if !exists {
		methodMetrics = &MethodMetrics{
			MethodName:     method,
			MinLatency:     duration,
			MaxLatency:     duration,
			AverageLatency: duration,
			ErrorCodes:     make(map[string]int64),
		}
		serviceMetrics.Methods[method] = methodMetrics
	}

	// Update service metrics
	serviceMetrics.TotalRequests++
	if success {
		serviceMetrics.SuccessRequests++
	} else {
		serviceMetrics.ErrorRequests++
		if err != nil {
			errorType := "unknown"
			// You can categorize errors here based on error type
			serviceMetrics.ErrorRates[errorType]++
		}
	}

	// Update service latency metrics
	if duration < serviceMetrics.MinLatency || serviceMetrics.TotalRequests == 1 {
		serviceMetrics.MinLatency = duration
	}
	if duration > serviceMetrics.MaxLatency {
		serviceMetrics.MaxLatency = duration
	}

	// Calculate moving average for service
	serviceMetrics.TotalDuration += duration
	serviceMetrics.AverageLatency = time.Duration(int64(serviceMetrics.TotalDuration) / serviceMetrics.TotalRequests)
	serviceMetrics.LastRequestTime = time.Now()
	serviceMetrics.UpdatedAt = time.Now()

	// Update method metrics
	methodMetrics.RequestCount++
	if success {
		methodMetrics.SuccessCount++
	} else {
		methodMetrics.ErrorCount++
		if err != nil {
			errorType := "unknown"
			methodMetrics.ErrorCodes[errorType]++
		}
	}

	// Update method latency metrics
	if duration < methodMetrics.MinLatency || methodMetrics.RequestCount == 1 {
		methodMetrics.MinLatency = duration
	}
	if duration > methodMetrics.MaxLatency {
		methodMetrics.MaxLatency = duration
	}

	// Calculate moving average for method
	methodMetrics.TotalDuration += duration
	methodMetrics.AverageLatency = time.Duration(int64(methodMetrics.TotalDuration) / methodMetrics.RequestCount)
	methodMetrics.LastRequestTime = time.Now()
}

// RecordStream records a stream metric
func (c *DefaultMetricsCollector) RecordStream(service, method string, duration time.Duration, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get or create service metrics
	serviceMetrics, exists := c.metrics[service]
	if !exists {
		serviceMetrics = &ServiceMetrics{
			ServiceName: service,
			Methods:     make(map[string]*MethodMetrics),
			ErrorRates:  make(map[string]int64),
			CreatedAt:   time.Now(),
		}
		c.metrics[service] = serviceMetrics
	}

	// Get or create method metrics
	methodMetrics, exists := serviceMetrics.Methods[method]
	if !exists {
		methodMetrics = &MethodMetrics{
			MethodName: method,
			ErrorCodes: make(map[string]int64),
		}
		serviceMetrics.Methods[method] = methodMetrics
	}

	// Get or create stream metrics
	if methodMetrics.StreamMetrics == nil {
		methodMetrics.StreamMetrics = &StreamMetrics{
			MinDuration: duration,
			MaxDuration: duration,
		}
	}
	streamMetrics := methodMetrics.StreamMetrics

	// Update stream metrics
	streamMetrics.StreamCount++
	if success {
		streamMetrics.SuccessStreams++
	} else {
		streamMetrics.ErrorStreams++
	}

	// Update stream duration metrics
	if duration < streamMetrics.MinDuration || streamMetrics.StreamCount == 1 {
		streamMetrics.MinDuration = duration
	}
	if duration > streamMetrics.MaxDuration {
		streamMetrics.MaxDuration = duration
	}

	// Calculate moving average for streams
	if streamMetrics.StreamCount == 1 {
		streamMetrics.AverageDuration = duration
	} else {
		// Simple moving average
		streamMetrics.AverageDuration = time.Duration(
			(int64(streamMetrics.AverageDuration)*9 + int64(duration)) / 10,
		)
	}

	streamMetrics.LastStreamTime = time.Now()
	serviceMetrics.UpdatedAt = time.Now()
}

// GetServiceMetrics returns metrics for a specific service
func (c *DefaultMetricsCollector) GetServiceMetrics(service string) *ServiceMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if serviceMetrics, exists := c.metrics[service]; exists {
		// Return a copy to avoid concurrent modifications
		return c.copyServiceMetrics(serviceMetrics)
	}

	return nil
}

// GetAllMetrics returns all collected metrics
func (c *DefaultMetricsCollector) GetAllMetrics() map[string]*ServiceMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*ServiceMetrics)
	for service, metrics := range c.metrics {
		result[service] = c.copyServiceMetrics(metrics)
	}

	return result
}

// ResetMetrics resets all metrics
func (c *DefaultMetricsCollector) ResetMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = make(map[string]*ServiceMetrics)
	c.logger.Info("All gRPC metrics reset")
}

// StartMetricsExporter starts the background metrics exporter
func (c *DefaultMetricsCollector) StartMetricsExporter(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.exporterActive {
		return nil // Already running
	}

	c.exporterCtx, c.exporterCancel = context.WithCancel(ctx)
	c.exporterActive = true

	go c.metricsExporter()

	c.logger.Info("gRPC metrics exporter started", "interval", c.exportInterval)
	return nil
}

// StopMetricsExporter stops the metrics exporter
func (c *DefaultMetricsCollector) StopMetricsExporter() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.exporterActive {
		return
	}

	c.exporterCancel()
	c.exporterActive = false
	c.logger.Info("gRPC metrics exporter stopped")
}

// metricsExporter runs the background metrics export loop
func (c *DefaultMetricsCollector) metricsExporter() {
	ticker := time.NewTicker(c.exportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.exporterCtx.Done():
			return
		case <-ticker.C:
			c.exportMetrics()
		}
	}
}

// exportMetrics exports current metrics
func (c *DefaultMetricsCollector) exportMetrics() {
	c.mu.RLock()
	metrics := make(map[string]*ServiceMetrics)
	for service, serviceMetrics := range c.metrics {
		metrics[service] = c.copyServiceMetrics(serviceMetrics)
	}
	c.mu.RUnlock()

	if len(metrics) == 0 {
		return
	}

	// Calculate overall statistics
	totalRequests := int64(0)
	totalErrors := int64(0)

	for _, serviceMetrics := range metrics {
		totalRequests += serviceMetrics.TotalRequests
		totalErrors += serviceMetrics.ErrorRequests
	}

	errorRate := float64(0)
	if totalRequests > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests) * 100
	}

	c.logger.Info("gRPC metrics export",
		"services", len(metrics),
		"total_requests", totalRequests,
		"total_errors", totalErrors,
		"error_rate_percent", errorRate,
	)

	// Log detailed metrics for each service
	for service, serviceMetrics := range metrics {
		if serviceMetrics.TotalRequests > 0 {
			c.logger.Debug("Service metrics",
				"service", service,
				"requests", serviceMetrics.TotalRequests,
				"success_rate", float64(serviceMetrics.SuccessRequests)/float64(serviceMetrics.TotalRequests)*100,
				"avg_latency_ms", serviceMetrics.AverageLatency.Milliseconds(),
			)
		}
	}
}

// copyServiceMetrics creates a deep copy of service metrics
func (c *DefaultMetricsCollector) copyServiceMetrics(original *ServiceMetrics) *ServiceMetrics {
	copy := &ServiceMetrics{
		ServiceName:      original.ServiceName,
		TotalRequests:    original.TotalRequests,
		SuccessRequests:  original.SuccessRequests,
		ErrorRequests:    original.ErrorRequests,
		TotalDuration:    original.TotalDuration,
		AverageLatency:   original.AverageLatency,
		MinLatency:       original.MinLatency,
		MaxLatency:       original.MaxLatency,
		LastRequestTime:  original.LastRequestTime,
		ErrorRates:       make(map[string]int64),
		CreatedAt:        original.CreatedAt,
		UpdatedAt:        original.UpdatedAt,
	}

	// Copy error rates
	for errorType, count := range original.ErrorRates {
		copy.ErrorRates[errorType] = count
	}

	// Copy methods
	copy.Methods = make(map[string]*MethodMetrics)
	for method, methodMetrics := range original.Methods {
		methodCopy := &MethodMetrics{
			MethodName:       methodMetrics.MethodName,
			RequestCount:     methodMetrics.RequestCount,
			SuccessCount:     methodMetrics.SuccessCount,
			ErrorCount:       methodMetrics.ErrorCount,
			TotalDuration:    methodMetrics.TotalDuration,
			AverageLatency:   methodMetrics.AverageLatency,
			MinLatency:       methodMetrics.MinLatency,
			MaxLatency:       methodMetrics.MaxLatency,
			LastRequestTime:  methodMetrics.LastRequestTime,
			ErrorCodes:       make(map[string]int64),
		}

		// Copy error codes
		for code, count := range methodMetrics.ErrorCodes {
			methodCopy.ErrorCodes[code] = count
		}

		// Copy stream metrics if exists
		if methodMetrics.StreamMetrics != nil {
			streamMetrics := methodMetrics.StreamMetrics
			methodCopy.StreamMetrics = &StreamMetrics{
				StreamCount:     streamMetrics.StreamCount,
				SuccessStreams:  streamMetrics.SuccessStreams,
				ErrorStreams:    streamMetrics.ErrorStreams,
				AverageDuration: streamMetrics.AverageDuration,
				MinDuration:     streamMetrics.MinDuration,
				MaxDuration:     streamMetrics.MaxDuration,
				BytesReceived:   streamMetrics.BytesReceived,
				BytesSent:       streamMetrics.BytesSent,
				LastStreamTime:  streamMetrics.LastStreamTime,
			}
		}

		copy.Methods[method] = methodCopy
	}

	return copy
}