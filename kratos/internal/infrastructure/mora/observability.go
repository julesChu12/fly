package mora

import (
	"context"

	"github.com/julesChu12/fly/mora/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer      trace.Tracer
	cleanupFunc observability.CleanupFunc
)

// InitObservability initializes OpenTelemetry tracing
func InitObservability(serviceName, exporterURL, env string) error {
	cfg := observability.Config{
		ServiceName:  serviceName,
		ExporterURL:  exporterURL,
		Environment:  env,
		SampleRatio:  1.0,
		ExporterType: "otlp",
	}

	// If no exporter URL, use stdout for development
	if exporterURL == "" {
		cfg.ExporterType = "stdout"
	}

	cleanup, err := observability.Init(cfg)
	if err != nil {
		return err
	}

	cleanupFunc = cleanup
	tracer = otel.Tracer(serviceName)
	return nil
}

// GetTracer returns the global tracer instance
func GetTracer() trace.Tracer {
	if tracer == nil {
		tracer = otel.Tracer("kratos")
	}
	return tracer
}

// StartSpan starts a new span
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, name)
}

// ShutdownObservability gracefully shuts down observability
func ShutdownObservability(ctx context.Context) error {
	if cleanupFunc != nil {
		return cleanupFunc()
	}
	return nil
}
