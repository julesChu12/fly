package gozero

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

// NewServerStatsHandler creates a new server stats handler for OpenTelemetry tracing
func NewServerStatsHandler(opts ...otelgrpc.Option) stats.Handler {
	return otelgrpc.NewServerHandler(opts...)
}

// NewClientStatsHandler creates a new client stats handler for OpenTelemetry tracing
func NewClientStatsHandler(opts ...otelgrpc.Option) stats.Handler {
	return otelgrpc.NewClientHandler(opts...)
}

// ServerOption returns a gRPC server option with OpenTelemetry stats handler
func ServerOption(opts ...otelgrpc.Option) grpc.ServerOption {
	return grpc.StatsHandler(otelgrpc.NewServerHandler(opts...))
}

// ClientOption returns a gRPC dial option with OpenTelemetry stats handler
func ClientOption(opts ...otelgrpc.Option) grpc.DialOption {
	return grpc.WithStatsHandler(otelgrpc.NewClientHandler(opts...))
}
