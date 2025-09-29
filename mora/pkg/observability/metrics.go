package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsRegistry manages Prometheus metrics for a service
type MetricsRegistry struct {
	registry  *prometheus.Registry
	namespace string

	// Pre-defined metric instances
	histograms map[string]*prometheus.HistogramVec
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
}

// NewMetricsRegistry creates a new metrics registry for the given service
func NewMetricsRegistry(serviceName string) *MetricsRegistry {
	registry := prometheus.NewRegistry()

	return &MetricsRegistry{
		registry:   registry,
		namespace:  serviceName,
		histograms: make(map[string]*prometheus.HistogramVec),
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
	}
}

// RegisterHistogram registers a new histogram metric
func (m *MetricsRegistry) RegisterHistogram(name, help string, labels []string, buckets ...float64) {
	defaultBuckets := prometheus.DefBuckets
	if len(buckets) > 0 {
		defaultBuckets = buckets
	}

	histogram := promauto.With(m.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.namespace,
			Name:      name,
			Help:      help,
			Buckets:   defaultBuckets,
		},
		labels,
	)

	m.histograms[name] = histogram
}

// RegisterCounter registers a new counter metric
func (m *MetricsRegistry) RegisterCounter(name, help string, labels []string) {
	counter := promauto.With(m.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.namespace,
			Name:      name,
			Help:      help,
		},
		labels,
	)

	m.counters[name] = counter
}

// RegisterGauge registers a new gauge metric
func (m *MetricsRegistry) RegisterGauge(name, help string, labels []string) {
	gauge := promauto.With(m.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.namespace,
			Name:      name,
			Help:      help,
		},
		labels,
	)

	m.gauges[name] = gauge
}

// ObserveHistogram records a value for a histogram metric
func (m *MetricsRegistry) ObserveHistogram(name string, value float64, labelValues ...string) {
	if histogram, exists := m.histograms[name]; exists {
		histogram.WithLabelValues(labelValues...).Observe(value)
	}
}

// IncrementCounter increments a counter metric
func (m *MetricsRegistry) IncrementCounter(name string, labelValues ...string) {
	if counter, exists := m.counters[name]; exists {
		counter.WithLabelValues(labelValues...).Inc()
	}
}

// AddToCounter adds a value to a counter metric
func (m *MetricsRegistry) AddToCounter(name string, value float64, labelValues ...string) {
	if counter, exists := m.counters[name]; exists {
		counter.WithLabelValues(labelValues...).Add(value)
	}
}

// SetGauge sets a gauge metric value
func (m *MetricsRegistry) SetGauge(name string, value float64, labelValues ...string) {
	if gauge, exists := m.gauges[name]; exists {
		gauge.WithLabelValues(labelValues...).Set(value)
	}
}

// IncGauge increments a gauge metric
func (m *MetricsRegistry) IncGauge(name string, labelValues ...string) {
	if gauge, exists := m.gauges[name]; exists {
		gauge.WithLabelValues(labelValues...).Inc()
	}
}

// DecGauge decrements a gauge metric
func (m *MetricsRegistry) DecGauge(name string, labelValues ...string) {
	if gauge, exists := m.gauges[name]; exists {
		gauge.WithLabelValues(labelValues...).Dec()
	}
}

// AddToGauge adds a value to a gauge metric
func (m *MetricsRegistry) AddToGauge(name string, value float64, labelValues ...string) {
	if gauge, exists := m.gauges[name]; exists {
		gauge.WithLabelValues(labelValues...).Add(value)
	}
}

// GetRegistry returns the underlying Prometheus registry
func (m *MetricsRegistry) GetRegistry() *prometheus.Registry {
	return m.registry
}

// RegisterStandardMetrics registers common application metrics
func (m *MetricsRegistry) RegisterStandardMetrics() {
	// HTTP request metrics
	m.RegisterHistogram("http_request_duration_seconds",
		"Duration of HTTP requests in seconds",
		[]string{"method", "path", "status"},
		0.001, 0.01, 0.1, 0.5, 1, 2.5, 5, 10)

	m.RegisterCounter("http_requests_total",
		"Total number of HTTP requests",
		[]string{"method", "path", "status"})

	// gRPC metrics
	m.RegisterHistogram("grpc_request_duration_seconds",
		"Duration of gRPC requests in seconds",
		[]string{"service", "method", "status"},
		0.001, 0.01, 0.1, 0.5, 1, 2.5, 5, 10)

	m.RegisterCounter("grpc_requests_total",
		"Total number of gRPC requests",
		[]string{"service", "method", "status"})

	// Database metrics
	m.RegisterHistogram("db_query_duration_seconds",
		"Duration of database queries in seconds",
		[]string{"operation", "table"},
		0.001, 0.01, 0.1, 0.5, 1, 2.5, 5, 10)

	m.RegisterCounter("db_queries_total",
		"Total number of database queries",
		[]string{"operation", "table", "status"})

	// Cache metrics
	m.RegisterCounter("cache_operations_total",
		"Total number of cache operations",
		[]string{"operation", "status"})

	m.RegisterHistogram("cache_operation_duration_seconds",
		"Duration of cache operations in seconds",
		[]string{"operation"},
		0.001, 0.01, 0.1, 0.5, 1, 2.5, 5)
}