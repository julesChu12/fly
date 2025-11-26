package monitoring

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// MonitoringIntegration provides a complete monitoring setup for gRPC services
type MonitoringIntegration struct {
	interceptor  *GRPCInterceptor
	collector    MetricsCollector
	tracer       GRPCTracer
	recorder     RequestRecorder
	sampler      RequestSampler
	logger       *logger.Logger
}

// MonitoringConfig represents configuration for monitoring integration
type MonitoringConfig struct {
	Interceptor   *GRPCInterceptorConfig `yaml:"interceptor"`
	Metrics       *MetricsConfig          `yaml:"metrics"`
	Tracing       *TracingConfig          `yaml:"tracing"`
	Sampling      *SamplingConfig         `yaml:"sampling"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled         bool          `yaml:"enabled"`
	ExportInterval  time.Duration `yaml:"export_interval"`
	FlushInterval   time.Duration `yaml:"flush_interval"`
}

// TracingConfig represents tracing configuration
type TracingConfig struct {
	Enabled bool   `yaml:"enabled"`
	Service string `yaml:"service"`
	Version string `yaml:"version"`
}

// SamplingConfig represents sampling configuration
type SamplingConfig struct {
	Enabled    bool    `yaml:"enabled"`
	Rate       float64 `yaml:"rate"`
	Strategy   string  `yaml:"strategy"` // random, deterministic, adaptive
}

// DefaultMonitoringConfig returns default monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		Interceptor: DefaultGRPCInterceptorConfig(),
		Metrics: &MetricsConfig{
			Enabled:        true,
			ExportInterval: 30 * time.Second,
			FlushInterval:   30 * time.Second,
		},
		Tracing: &TracingConfig{
			Enabled: true,
			Service: "clotho",
			Version: "1.0.0",
		},
		Sampling: &SamplingConfig{
			Enabled:  true,
			Rate:     0.1, // 10%
			Strategy: "random",
		},
	}
}

// NewMonitoringIntegration creates a new monitoring integration
func NewMonitoringIntegration(config *MonitoringConfig) *MonitoringIntegration {
	if config == nil {
		config = DefaultMonitoringConfig()
	}

	log := logger.NewDefault()
	log.Info("Initializing gRPC monitoring integration")

	// Create monitoring components
	collector := NewMetricsCollector(log)
	tracer := NewGRPCTracer(config.Tracing.Enabled, log)
	recorder := NewRequestRecorder(
		config.Interceptor.EnableLogging,
		int64(config.Interceptor.MaxRequestSize),
		int64(config.Interceptor.MaxResponseSize),
		log,
	)
	sampler := NewRequestSampler(
		config.Sampling.Enabled,
		config.Sampling.Rate,
		log,
	)

	// Create interceptor with all components
	interceptor := NewGRPCInterceptor(config.Interceptor)

	// Replace components in interceptor with our instances
	interceptor.tracer = tracer
	interceptor.recorder = recorder
	interceptor.sampler = sampler

	integration := &MonitoringIntegration{
		interceptor: interceptor,
		collector:   collector,
		tracer:      tracer,
		recorder:    recorder,
		sampler:     sampler,
		logger:      log,
	}

	log.Info("gRPC monitoring integration initialized",
		"metrics_enabled", config.Metrics.Enabled,
		"tracing_enabled", config.Tracing.Enabled,
		"sampling_enabled", config.Sampling.Enabled,
		"sampling_rate", config.Sampling.Rate,
	)

	return integration
}

// GetUnaryServerInterceptor returns the unary server interceptor
func (m *MonitoringIntegration) GetUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return m.interceptor.UnaryServerInterceptor()
}

// GetUnaryClientInterceptor returns the unary client interceptor
func (m *MonitoringIntegration) GetUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return m.interceptor.UnaryClientInterceptor()
}

// GetStreamServerInterceptor returns the stream server interceptor
func (m *MonitoringIntegration) GetStreamServerInterceptor() grpc.StreamServerInterceptor {
	return m.interceptor.StreamServerInterceptor()
}

// GetStreamClientInterceptor returns the stream client interceptor
func (m *MonitoringIntegration) GetStreamClientInterceptor() grpc.StreamClientInterceptor {
	return m.interceptor.StreamClientInterceptor()
}

// GetDialOptions returns the dial options with monitoring interceptors
func (m *MonitoringIntegration) GetDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(m.interceptor.UnaryClientInterceptor()),
		grpc.WithChainStreamInterceptor(m.interceptor.StreamClientInterceptor()),
	}
}

// GetServerOptions returns the server options with monitoring interceptors
func (m *MonitoringIntegration) GetServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(m.interceptor.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(m.interceptor.StreamServerInterceptor()),
	}
}

// GetCollector returns the metrics collector
func (m *MonitoringIntegration) GetCollector() MetricsCollector {
	return m.collector
}

// GetInterceptor returns the gRPC interceptor
func (m *MonitoringIntegration) GetInterceptor() *GRPCInterceptor {
	return m.interceptor
}

// GetMetrics returns current metrics
func (m *MonitoringIntegration) GetMetrics() map[string]*GRPCMetrics {
	return m.interceptor.GetMetrics()
}

// ResetMetrics resets all metrics
func (m *MonitoringIntegration) ResetMetrics() {
	m.interceptor.ResetMetrics()
	m.collector.ResetMetrics()
}

// Start starts the monitoring integration
func (m *MonitoringIntegration) Start(ctx context.Context) error {
	m.logger.Info("Starting gRPC monitoring integration")

	// Start metrics collector
	if err := m.collector.StartMetricsExporter(ctx); err != nil {
		m.logger.Error("Failed to start metrics collector", "error", err.Error())
		return err
	}

	m.logger.Info("gRPC monitoring integration started")
	return nil
}

// Stop stops the monitoring integration
func (m *MonitoringIntegration) Stop() {
	m.logger.Info("Stopping gRPC monitoring integration")

	m.collector.StopMetricsExporter()

	m.logger.Info("gRPC monitoring integration stopped")
}

// GetHealthStatus returns the health status of the monitoring integration
func (m *MonitoringIntegration) GetHealthStatus() map[string]interface{} {
	metrics := m.collector.GetAllMetrics()

	status := map[string]interface{}{
		"interceptor_enabled": true,
		"collector_enabled":   true,
		"tracer_enabled":      true,
		"recorder_enabled":    true,
		"sampler_enabled":     true,
		"services_monitored":  len(metrics),
		"total_requests":      m.interceptor.getTotalRequests(),
		"total_errors":        m.interceptor.getTotalErrors(),
		"avg_success_rate":    m.interceptor.getAverageSuccessRate(),
	}

	if samplingStats := m.sampler.GetStats(); samplingStats != nil {
		status["sampling_stats"] = samplingStats
	}

	return status
}

// ServiceHealthCheck represents a health check for a specific service
type ServiceHealthCheck struct {
	ServiceName string    `json:"service_name"`
	Status      string    `json:"status"`
	LastCheck   time.Time `json:"last_check"`
	Error       string    `json:"error,omitempty"`
	Metrics     *ServiceMetrics `json:"metrics,omitempty"`
}

// GetServiceHealth returns health status for all monitored services
func (m *MonitoringIntegration) GetServiceHealth() []ServiceHealthCheck {
	metrics := m.collector.GetAllMetrics()
	healthChecks := make([]ServiceHealthCheck, 0, len(metrics))

	for serviceName, serviceMetrics := range metrics {
		healthCheck := ServiceHealthCheck{
			ServiceName: serviceName,
			Status:      "healthy", // Default status
			LastCheck:   time.Now(),
			Metrics:     serviceMetrics,
		}

		// Determine health status based on metrics
		if serviceMetrics.TotalRequests == 0 {
			healthCheck.Status = "no_requests"
		} else {
			errorRate := float64(serviceMetrics.ErrorRequests) / float64(serviceMetrics.TotalRequests)
			if errorRate > 0.5 { // 50% error rate threshold
				healthCheck.Status = "unhealthy"
			} else if errorRate > 0.1 { // 10% error rate threshold
				healthCheck.Status = "degraded"
			}

			// Check latency
			if serviceMetrics.AverageLatency > 5*time.Second {
				if healthCheck.Status == "healthy" {
					healthCheck.Status = "slow"
				}
			}
		}

		healthChecks = append(healthChecks, healthCheck)
	}

	return healthChecks
}