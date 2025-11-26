package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// ProtocolType represents the communication protocol type
type ProtocolType string

const (
	ProtocolHTTP ProtocolType = "http"
	ProtocolGRPC ProtocolType = "grpc"
)

// SelectionStrategy defines the strategy for protocol selection
type SelectionStrategy string

const (
	StrategyAuto       SelectionStrategy = "auto"       // Automatic selection based on metrics
	StrategyHTTP       SelectionStrategy = "http"       // Force HTTP
	StrategyGRPC       SelectionStrategy = "grpc"       // Force gRPC
	StrategyPerformance SelectionStrategy = "performance" // Choose based on performance metrics
	StrategyLatency    SelectionStrategy = "latency"     // Choose based on latency
	StrategyReliability SelectionStrategy = "reliability" // Choose based on reliability
)

// ProtocolMetrics tracks performance metrics for each protocol
type ProtocolMetrics struct {
	RequestCount    int64         `json:"request_count"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	AvgLatency      time.Duration `json:"avg_latency"`
	MinLatency      time.Duration `json:"min_latency"`
	MaxLatency      time.Duration `json:"max_latency"`
	LastUpdated     time.Time     `json:"last_updated"`
	ResponseSize    int64         `json:"response_size"`
	ErrorRate       float64       `json:"error_rate"`
	SuccessRate     float64       `json:"success_rate"`
}

// ProtocolSelectorConfig represents configuration for protocol selector
type ProtocolSelectorConfig struct {
	Strategy            SelectionStrategy `yaml:"strategy"`
	EnableMetrics       bool              `yaml:"enable_metrics"`
	MetricsWindow       time.Duration     `yaml:"metrics_window"`
	MinSampleSize       int               `yaml:"min_sample_size"`
	LatencyThreshold    time.Duration     `yaml:"latency_threshold"`
	ErrorRateThreshold  float64           `yaml:"error_rate_threshold"`
	EnableFallback      bool              `yaml:"enable_fallback"`
	FallbackAfterErrors int               `yaml:"fallback_after_errors"`
	HealthCheckInterval time.Duration     `yaml:"health_check_interval"`
	PreferHTTPFor       []string          `yaml:"prefer_http_for"`  // Endpoint patterns that prefer HTTP
	PreferGRPCFor       []string          `yaml:"prefer_grpc_for"`  // Endpoint patterns that prefer gRPC
}

// DefaultProtocolSelectorConfig returns default configuration
func DefaultProtocolSelectorConfig() *ProtocolSelectorConfig {
	return &ProtocolSelectorConfig{
		Strategy:            StrategyAuto,
		EnableMetrics:       true,
		MetricsWindow:       5 * time.Minute,
		MinSampleSize:       10,
		LatencyThreshold:    100 * time.Millisecond,
		ErrorRateThreshold:  0.05, // 5%
		EnableFallback:      true,
		FallbackAfterErrors: 3,
		HealthCheckInterval: 30 * time.Second,
		PreferHTTPFor:       []string{"/api/docs", "/api/health", "/api/metrics"},
		PreferGRPCFor:       []string{"/api/customers", "/api/orders", "/api/payments"},
	}
}

// ProtocolSelector intelligent protocol selection middleware
type ProtocolSelector struct {
	config                *ProtocolSelectorConfig
	logger                *logger.Logger
	httpClients           map[string]interface{} // HTTP clients by service name
	grpcClients           map[string]interface{} // gRPC clients by service name
	metrics               map[ProtocolType]*ProtocolMetrics
	mu                    sync.RWMutex
	healthCheckStatus     map[ProtocolType]bool
	healthCheckMu         sync.RWMutex
	requestBuffer         []RequestRecord
	healthCheckTicker      *time.Ticker
	stopChan              chan struct{}
}

// RequestRecord records a request for metric collection
type RequestRecord struct {
	Protocol     ProtocolType
	Service      string
	Endpoint     string
	Latency      time.Duration
	Success      bool
	ResponseSize int64
	Timestamp    time.Time
}

// NewProtocolSelector creates a new protocol selector
func NewProtocolSelector(config *ProtocolSelectorConfig) *ProtocolSelector {
	if config == nil {
		config = DefaultProtocolSelectorConfig()
	}

	logger := logger.NewDefault()
	logger.Info("Initializing Protocol Selector", "strategy", config.Strategy)

	ps := &ProtocolSelector{
		config:            config,
		logger:            logger,
		httpClients:       make(map[string]interface{}),
		grpcClients:       make(map[string]interface{}),
		metrics:           make(map[ProtocolType]*ProtocolMetrics),
		healthCheckStatus: make(map[ProtocolType]bool),
		requestBuffer:     make([]RequestRecord, 0, 1000),
		stopChan:          make(chan struct{}),
	}

	// Initialize metrics
	ps.metrics[ProtocolHTTP] = &ProtocolMetrics{}
	ps.metrics[ProtocolGRPC] = &ProtocolMetrics{}

	// Start background tasks if metrics are enabled
	if config.EnableMetrics {
		go ps.metricsCollector()
		go ps.healthChecker()
	}

	return ps
}

// RegisterHTTPClient registers an HTTP client for a service
func (ps *ProtocolSelector) RegisterHTTPClient(serviceName string, client interface{}) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.httpClients[serviceName] = client
	ps.logger.Debug("Registered HTTP client", "service", serviceName)
}

// RegisterGRPCClient registers a gRPC client for a service
func (ps *ProtocolSelector) RegisterGRPCClient(serviceName string, client interface{}) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.grpcClients[serviceName] = client
	ps.logger.Debug("Registered gRPC client", "service", serviceName)
}

// SelectProtocol intelligently selects the best protocol for a request
func (ps *ProtocolSelector) SelectProtocol(serviceName, endpoint string, requestSize int64) ProtocolType {
	switch ps.config.Strategy {
	case StrategyHTTP:
		return ProtocolHTTP
	case StrategyGRPC:
		return ProtocolGRPC
	case StrategyPerformance:
		return ps.selectByPerformance(serviceName, endpoint)
	case StrategyLatency:
		return ps.selectByLatency(serviceName, endpoint)
	case StrategyReliability:
		return ps.selectByReliability(serviceName, endpoint)
	case StrategyAuto:
		return ps.selectAuto(serviceName, endpoint, requestSize)
	default:
		return ps.selectAuto(serviceName, endpoint, requestSize)
	}
}

// selectAuto implements automatic protocol selection logic
func (ps *ProtocolSelector) selectAuto(serviceName, endpoint string, requestSize int64) ProtocolType {
	// Check if endpoint has preference
	if ps.hasEndpointPreference(endpoint, ProtocolHTTP) {
		return ProtocolHTTP
	}
	if ps.hasEndpointPreference(endpoint, ProtocolGRPC) {
		return ProtocolGRPC
	}

	// Check health status
	if !ps.isProtocolHealthy(ProtocolGRPC) {
		ps.logger.Warn("gRPC unhealthy, falling back to HTTP", "service", serviceName)
		return ProtocolHTTP
	}

	if !ps.isProtocolHealthy(ProtocolHTTP) {
		ps.logger.Warn("HTTP unhealthy, using gRPC", "service", serviceName)
		return ProtocolGRPC
	}

	// Check if both clients are available
	ps.mu.RLock()
	httpAvailable := ps.httpClients[serviceName] != nil
	grpcAvailable := ps.grpcClients[serviceName] != nil
	ps.mu.RUnlock()

	if httpAvailable && !grpcAvailable {
		return ProtocolHTTP
	}
	if grpcAvailable && !httpAvailable {
		return ProtocolGRPC
	}

	// Default to gRPC for better performance when both are available
	if httpAvailable && grpcAvailable {
		// For small requests (<1KB), HTTP might be faster due to lower connection overhead
		if requestSize < 1024 {
			return ProtocolHTTP
		}
		// For larger requests or batch operations, prefer gRPC
		return ProtocolGRPC
	}

	// Fallback to HTTP
	return ProtocolHTTP
}

// selectByPerformance selects protocol based on performance metrics
func (ps *ProtocolSelector) selectByPerformance(serviceName, endpoint string) ProtocolType {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	httpMetrics := ps.metrics[ProtocolHTTP]
	grpcMetrics := ps.metrics[ProtocolGRPC]

	// Not enough data, use default logic
	if httpMetrics.RequestCount < int64(ps.config.MinSampleSize) &&
	   grpcMetrics.RequestCount < int64(ps.config.MinSampleSize) {
		return ps.selectAuto(serviceName, endpoint, 0)
	}

	// Compare success rates and latency
	httpScore := ps.calculatePerformanceScore(httpMetrics)
	grpcScore := ps.calculatePerformanceScore(grpcMetrics)

	ps.logger.Debug("Performance scores calculated",
		"http_score", httpScore,
		"grpc_score", grpcScore)

	if grpcScore > httpScore {
		return ProtocolGRPC
	}
	return ProtocolHTTP
}

// selectByLatency selects protocol based on latency
func (ps *ProtocolSelector) selectByLatency(serviceName, endpoint string) ProtocolType {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	httpMetrics := ps.metrics[ProtocolHTTP]
	grpcMetrics := ps.metrics[ProtocolGRPC]

	// Not enough data, use default logic
	if httpMetrics.RequestCount < int64(ps.config.MinSampleSize) &&
	   grpcMetrics.RequestCount < int64(ps.config.MinSampleSize) {
		return ps.selectAuto(serviceName, endpoint, 0)
	}

	// Compare average latency
	if grpcMetrics.AvgLatency < httpMetrics.AvgLatency-ps.config.LatencyThreshold {
		return ProtocolGRPC
	}
	if httpMetrics.AvgLatency < grpcMetrics.AvgLatency-ps.config.LatencyThreshold {
		return ProtocolHTTP
	}

	// If latency is similar, prefer gRPC for its other benefits
	return ProtocolGRPC
}

// selectByReliability selects protocol based on error rates
func (ps *ProtocolSelector) selectByReliability(serviceName, endpoint string) ProtocolType {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	httpMetrics := ps.metrics[ProtocolHTTP]
	grpcMetrics := ps.metrics[ProtocolGRPC]

	// Not enough data, use default logic
	if httpMetrics.RequestCount < int64(ps.config.MinSampleSize) &&
	   grpcMetrics.RequestCount < int64(ps.config.MinSampleSize) {
		return ps.selectAuto(serviceName, endpoint, 0)
	}

	// Check error rates
	if grpcMetrics.ErrorRate > ps.config.ErrorRateThreshold &&
	   httpMetrics.ErrorRate <= ps.config.ErrorRateThreshold {
		return ProtocolHTTP
	}
	if httpMetrics.ErrorRate > ps.config.ErrorRateThreshold &&
	   grpcMetrics.ErrorRate <= ps.config.ErrorRateThreshold {
		return ProtocolGRPC
	}

	// If both are reliable, prefer gRPC
	return ProtocolGRPC
}

// hasEndpointPreference checks if an endpoint has a protocol preference
func (ps *ProtocolSelector) hasEndpointPreference(endpoint string, protocol ProtocolType) bool {
	switch protocol {
	case ProtocolHTTP:
		for _, pattern := range ps.config.PreferHTTPFor {
			if matchEndpointPattern(endpoint, pattern) {
				return true
			}
		}
	case ProtocolGRPC:
		for _, pattern := range ps.config.PreferGRPCFor {
			if matchEndpointPattern(endpoint, pattern) {
				return true
			}
		}
	}
	return false
}

// matchEndpointPattern checks if endpoint matches a pattern
func matchEndpointPattern(endpoint, pattern string) bool {
	// Simple pattern matching - can be enhanced with regex if needed
	if pattern == endpoint {
		return true
	}
	// Prefix matching
	if len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
		return len(endpoint) >= len(pattern) && endpoint[:len(pattern)] == pattern
	}
	return false
}

// isProtocolHealthy checks if a protocol is healthy
func (ps *ProtocolSelector) isProtocolHealthy(protocol ProtocolType) bool {
	ps.healthCheckMu.RLock()
	defer ps.healthCheckMu.RUnlock()

	return ps.healthCheckStatus[protocol]
}

// calculatePerformanceScore calculates a performance score for protocol metrics
func (ps *ProtocolSelector) calculatePerformanceScore(metrics *ProtocolMetrics) float64 {
	if metrics.RequestCount == 0 {
		return 0.0
	}

	// Weight factors
	latencyWeight := 0.4
	reliabilityWeight := 0.6

	// Normalize latency (lower is better, max 1 second)
	latencyScore := 1.0 - float64(metrics.AvgLatency.Seconds())

	// Reliability score (success rate)
	reliabilityScore := metrics.SuccessRate

	// Weighted score
	score := latencyWeight*latencyScore + reliabilityWeight*reliabilityScore

	// Ensure score is between 0 and 1
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// RecordRequest records a request for metric collection
func (ps *ProtocolSelector) RecordRequest(protocol ProtocolType, serviceName, endpoint string, latency time.Duration, success bool, responseSize int64) {
	if !ps.config.EnableMetrics {
		return
	}

	record := RequestRecord{
		Protocol:     protocol,
		Service:      serviceName,
		Endpoint:     endpoint,
		Latency:      latency,
		Success:      success,
		ResponseSize: responseSize,
		Timestamp:    time.Now(),
	}

	// Add to buffer for background processing
	ps.mu.Lock()
	ps.requestBuffer = append(ps.requestBuffer, record)

	// Prevent buffer from growing too large
	if len(ps.requestBuffer) > 1000 {
		// Process oldest records
		ps.processRequestRecords(ps.requestBuffer[:500])
		ps.requestBuffer = ps.requestBuffer[500:]
	}
	ps.mu.Unlock()
}

// metricsCollector runs in background to process request metrics
func (ps *ProtocolSelector) metricsCollector() {
	ticker := time.NewTicker(ps.config.MetricsWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ps.processMetrics()
		case <-ps.stopChan:
			return
		}
	}
}

// processMetrics processes accumulated request records and updates metrics
func (ps *ProtocolSelector) processMetrics() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(ps.requestBuffer) == 0 {
		return
	}

	ps.processRequestRecords(ps.requestBuffer)
	ps.requestBuffer = ps.requestBuffer[:0]
}

// processRequestRecords processes a batch of request records
func (ps *ProtocolSelector) processRequestRecords(records []RequestRecord) {
	// Group records by protocol
	protocolRecords := map[ProtocolType][]RequestRecord{
		ProtocolHTTP: {},
		ProtocolGRPC: {},
	}

	for _, record := range records {
		protocolRecords[record.Protocol] = append(protocolRecords[record.Protocol], record)
	}

	// Update metrics for each protocol
	for protocol, protoRecords := range protocolRecords {
		if len(protoRecords) == 0 {
			continue
		}

		metrics := ps.metrics[protocol]
		if metrics == nil {
			metrics = &ProtocolMetrics{}
			ps.metrics[protocol] = metrics
		}

		// Calculate new metrics
		var totalLatency time.Duration
		var totalResponseSize int64
		successCount := int64(0)
		minLatency := protoRecords[0].Latency
		maxLatency := protoRecords[0].Latency

		for _, record := range protoRecords {
			totalLatency += record.Latency
			totalResponseSize += record.ResponseSize
			if record.Success {
				successCount++
			}
			if record.Latency < minLatency {
				minLatency = record.Latency
			}
			if record.Latency > maxLatency {
				maxLatency = record.Latency
			}
		}

		// Update metrics with exponential moving average
		weight := float64(len(protoRecords)) / float64(metrics.RequestCount+int64(len(protoRecords)))

		metrics.RequestCount += int64(len(protoRecords))
		metrics.SuccessCount += successCount
		metrics.ErrorCount += int64(len(protoRecords)) - successCount
		avgLatency := time.Duration(totalLatency) / time.Duration(len(protoRecords))
		metrics.AvgLatency = time.Duration(float64(metrics.AvgLatency)*(1-weight) + float64(avgLatency)*weight)
		metrics.MinLatency = minLatency
		metrics.MaxLatency = maxLatency
		metrics.ResponseSize += totalResponseSize
		metrics.LastUpdated = time.Now()

		if metrics.RequestCount > 0 {
			metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.RequestCount)
			metrics.ErrorRate = float64(metrics.ErrorCount) / float64(metrics.RequestCount)
		}

		ps.logger.Debug("Updated protocol metrics",
			"protocol", protocol,
			"request_count", metrics.RequestCount,
			"success_rate", metrics.SuccessRate,
			"avg_latency", metrics.AvgLatency)
	}
}

// healthChecker runs background health checks for protocols
func (ps *ProtocolSelector) healthChecker() {
	if ps.config.HealthCheckInterval <= 0 {
		return
	}

	ticker := time.NewTicker(ps.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ps.performHealthChecks()
		case <-ps.stopChan:
			return
		}
	}
}

// performHealthChecks performs health checks for all protocols
func (ps *ProtocolSelector) performHealthChecks() {
	// Check HTTP health
	httpHealthy := ps.checkHTTPHealth()

	// Check gRPC health
	grpcHealthy := ps.checkGRPCHealth()

	ps.healthCheckMu.Lock()
	ps.healthCheckStatus[ProtocolHTTP] = httpHealthy
	ps.healthCheckStatus[ProtocolGRPC] = grpcHealthy
	ps.healthCheckMu.Unlock()

	ps.logger.Debug("Health check completed",
		"http_healthy", httpHealthy,
		"grpc_healthy", grpcHealthy)
}

// checkHTTPHealth checks HTTP protocol health
func (ps *ProtocolSelector) checkHTTPHealth() bool {
	// Simple health check - can be enhanced with actual ping/heartbeat
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Check if any HTTP client is registered
	for _, client := range ps.httpClients {
		if client != nil {
			return true
		}
	}
	return false
}

// checkGRPCHealth checks gRPC protocol health
func (ps *ProtocolSelector) checkGRPCHealth() bool {
	// Simple health check - can be enhanced with actual ping/heartbeat
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Check if any gRPC client is registered
	for _, client := range ps.grpcClients {
		if client != nil {
			return true
		}
	}
	return false
}

// GetMetrics returns current protocol metrics
func (ps *ProtocolSelector) GetMetrics() map[ProtocolType]*ProtocolMetrics {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	metrics := make(map[ProtocolType]*ProtocolMetrics)
	for protocol, metric := range ps.metrics {
		// Return a copy to avoid concurrent modifications
		metricCopy := *metric
		metrics[protocol] = &metricCopy
	}
	return metrics
}

// GetHealthStatus returns current health status
func (ps *ProtocolSelector) GetHealthStatus() map[ProtocolType]bool {
	ps.healthCheckMu.RLock()
	defer ps.healthCheckMu.RUnlock()

	status := make(map[ProtocolType]bool)
	for protocol, healthy := range ps.healthCheckStatus {
		status[protocol] = healthy
	}
	return status
}

// Stop stops the protocol selector background tasks
func (ps *ProtocolSelector) Stop() {
	close(ps.stopChan)
	if ps.healthCheckTicker != nil {
		ps.healthCheckTicker.Stop()
	}
	ps.logger.Info("Protocol selector stopped")
}

// ProtocolSelectionMiddleware creates a Gin middleware for protocol selection
func (ps *ProtocolSelector) ProtocolSelectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get service name from context or request path
		serviceName := c.Param("service")
		if serviceName == "" {
			// Extract service name from path
			path := c.Request.URL.Path
			if len(path) > 1 {
				// Simple path parsing - can be enhanced
				parts := splitPath(path)
				if len(parts) > 1 {
					serviceName = parts[1]
				}
			}
		}

		// Select protocol for this request
		requestSize := c.Request.ContentLength
		selectedProtocol := ps.SelectProtocol(serviceName, c.Request.URL.Path, requestSize)

		// Store selection in context
		c.Set("selected_protocol", selectedProtocol)
		c.Set("service_name", serviceName)

		// Log the selection
		ps.logger.Debug("Protocol selected",
			"service", serviceName,
			"endpoint", c.Request.URL.Path,
			"protocol", selectedProtocol,
			"request_size", requestSize)

		c.Next()
	}
}

// splitPath splits URL path into parts
func splitPath(path string) []string {
	parts := make([]string, 0)
	start := 0

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}

	if start < len(path) {
		parts = append(parts, path[start:])
	}

	return parts
}

// WrapExecution wraps request execution with protocol selection and metric recording
func (ps *ProtocolSelector) WrapExecution(
	ctx context.Context,
	serviceName, endpoint string,
	requestSize int64,
	executeFunc func(ctx context.Context, protocol ProtocolType) (interface{}, error),
) (interface{}, error) {
	// Select protocol
	protocol := ps.SelectProtocol(serviceName, endpoint, requestSize)

	// Record start time
	startTime := time.Now()

	// Execute the request
	result, err := executeFunc(ctx, protocol)

	// Record metrics
	latency := time.Since(startTime)
	success := err == nil
	responseSize := int64(0)

	// Estimate response size if possible
	if result != nil {
		// This is a simplified estimation
		responseSize = 1024 // 1KB default
	}

	ps.RecordRequest(protocol, serviceName, endpoint, latency, success, responseSize)

	return result, err
}