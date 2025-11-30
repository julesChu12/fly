package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"github.com/julesChu12/fly/mora/pkg/observability"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// Config 可观测性配置
type Config struct {
	ServiceName    string  `yaml:"service_name" json:"service_name"`
	Environment    string  `yaml:"environment" json:"environment"`
	ExporterType   string  `yaml:"exporter_type" json:"exporter_type"`
	ExporterURL    string  `yaml:"exporter_url" json:"exporter_url"`
	SampleRatio    float64 `yaml:"sample_ratio" json:"sample_ratio"`
	EnableMetrics  bool    `yaml:"enable_metrics" json:"enable_metrics"`
	EnableTracing  bool    `yaml:"enable_tracing" json:"enable_tracing"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		ServiceName:   "appointments-service",
		Environment:   "development",
		ExporterType:  "stdout", // stdout, otlp
		ExporterURL:   "http://localhost:4317",
		SampleRatio:   1.0,
		EnableMetrics: true,
		EnableTracing: true,
	}
}

// Manager 可观测性管理器
type Manager struct {
	config   *observability.Config
	cleanup  observability.CleanupFunc
	logger   *logger.Logger
	tracer   trace.Tracer
}

// NewManager 创建可观测性管理器
func NewManager(config *Config, logger *logger.Logger) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 转换配置
	obsConfig := observability.Config{
		ServiceName:   config.ServiceName,
		Environment:   config.Environment,
		ExporterType:  config.ExporterType,
		ExporterURL:   config.ExporterURL,
		SampleRatio:   config.SampleRatio,
	}

	// 初始化 OpenTelemetry
	cleanup, err := observability.Init(obsConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化可观测性失败: %w", err)
	}

	// 获取 Tracer
	tracer := observability.GetTracer(config.ServiceName)

	logger.Info("可观测性初始化成功",
		map[string]interface{}{
			"service_name":  config.ServiceName,
			"environment":   config.Environment,
			"exporter_type": config.ExporterType,
			"sample_ratio":  config.SampleRatio,
		})

	return &Manager{
		config:  &obsConfig,
		cleanup: cleanup,
		logger:  logger,
		tracer:  tracer,
	}, nil
}

// GetTracer 获取 Tracer
func (m *Manager) GetTracer() trace.Tracer {
	return m.tracer
}

// StartSpan 开始 Span
func (m *Manager) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return m.tracer.Start(ctx, name)
}

// StartSpanWithAttributes 开始带属性的 Span
func (m *Manager) StartSpanWithAttributes(ctx context.Context, name string, attrs map[string]string) (context.Context, trace.Span) {
	ctx, span := m.tracer.Start(ctx, name)

	// 添加属性
	for key, value := range attrs {
		span.SetAttributes(attribute.String(key, value))
	}

	return ctx, span
}

// RecordError 记录错误到 Span
func (m *Manager) RecordError(span trace.Span, err error, message string) {
	if span != nil && err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("error.message", err.Error()),
		)
		if message != "" {
			span.SetAttributes(
				attribute.String("error.description", message),
			)
		}
		m.logger.Error("Span 错误记录",
			map[string]interface{}{
				"error":   err.Error(),
				"message": message,
			})
	}
}

// SetSpanAttributes 设置 Span 属性
func (m *Manager) SetSpanAttributes(span trace.Span, attrs map[string]interface{}) {
	if span != nil {
		for key, value := range attrs {
			span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
		}
	}
}

// AddEvent 添加事件到 Span
func (m *Manager) AddEvent(span trace.Span, name string, attrs map[string]interface{}) {
	if span != nil {
		attributes := make([]attribute.KeyValue, 0, len(attrs))
		for key, value := range attrs {
			attributes = append(attributes, attribute.String(key, fmt.Sprintf("%v", value)))
		}
		span.AddEvent(name, trace.WithAttributes(attributes...))
	}
}

// Close 关闭可观测性管理器
func (m *Manager) Close() error {
	if m.cleanup != nil {
		return m.cleanup()
	}
	return nil
}

// GetContextSpan 获取当前上下文的 Span
func GetContextSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// WithSpan 创建带 Span 的函数执行器
func WithSpan[T any](ctx context.Context, manager *Manager, name string, fn func(context.Context) (T, error)) (T, error) {
	var zero T

	ctx, span := manager.StartSpan(ctx, name)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		manager.RecordError(span, err, "function execution failed")
		return zero, err
	}

	return result, nil
}

// WithSpanWithTimeout 创建带超时的 Span 执行器
func WithSpanWithTimeout[T any](ctx context.Context, manager *Manager, name string, timeout time.Duration, fn func(context.Context) (T, error)) (T, error) {
	var zero T

	ctx, span := manager.StartSpan(ctx, name)
	defer span.End()

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := fn(ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			manager.RecordError(span, ctx.Err(), "operation timeout")
		} else {
			manager.RecordError(span, err, "function execution failed")
		}
		return zero, err
	}

	return result, nil
}