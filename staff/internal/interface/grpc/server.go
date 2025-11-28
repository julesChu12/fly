package grpc

import (
	"fmt"
	"net"
	"time"

	staffv1 "github.com/julesChu12/fly/staff/api/proto/staff/v1"
	"github.com/julesChu12/fly/staff/internal/application/service"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// Server represents the gRPC server
type Server struct {
	staffService service.StaffService
	grpcServer   *grpc.Server
	port         string
	logger       *logger.Logger
}

// NewServer creates a new gRPC server instance
func NewServer(staffService service.StaffService, port string, logger *logger.Logger) *Server {
	return &Server{
		staffService: staffService,
		port:         port,
		logger:       logger,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Create listener
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	// Configure gRPC server options
	opts := []grpc.ServerOption{
		// Keepalive parameters
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second, // Kick out idle connections after 15 seconds
			MaxConnectionAge:      30 * time.Second, // Max age of a connection
			MaxConnectionAgeGrace: 5 * time.Second,  // Grace period for connection age
			Time:                  5 * time.Second,  // Ping interval for checking connection health
			Timeout:               1 * time.Second,  // Ping timeout
		}),
		// Max message size
		grpc.MaxRecvMsgSize(4 * 1024 * 1024), // 4MB
		grpc.MaxSendMsgSize(4 * 1024 * 1024), // 4MB
	}

	// Create gRPC server
	s.grpcServer = grpc.NewServer(opts...)

	// Create and register staff service handler
	staffHandler := NewStaffServer(s.staffService)
	staffv1.RegisterStaffServiceServer(s.grpcServer, staffHandler)

	// Enable reflection for debugging and development
	reflection.Register(s.grpcServer)

	s.logger.Infof("Starting gRPC server on port %s", s.port)

	// Start serving
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.logger.Info("Stopping gRPC server...")
		s.grpcServer.GracefulStop()
	}
}

// StopWithTimeout stops the gRPC server with timeout
func (s *Server) StopWithTimeout(timeout time.Duration) error {
	if s.grpcServer == nil {
		return nil
	}

	s.logger.Infof("Stopping gRPC server with timeout %v", timeout)

	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("gRPC server stopped gracefully")
		return nil
	case <-time.After(timeout):
		s.logger.Warn("gRPC server shutdown timeout, forcing stop")
		s.grpcServer.Stop()
		return fmt.Errorf("gRPC server shutdown timeout")
	}
}

// GetPort returns the server port
func (s *Server) GetPort() string {
	return s.port
}

// IsRunning returns true if the server is running
func (s *Server) IsRunning() bool {
	return s.grpcServer != nil
}