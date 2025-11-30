package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	itemsv1 "github.com/julesChu12/fly/items/api/proto/items/v1"
	"github.com/julesChu12/fly/items/internal/application/service"
	grpcHandlers "github.com/julesChu12/fly/items/internal/interface/grpc/handlers"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// Server gRPC 服务器
type Server struct {
	itemService     service.ItemService
	categoryService *grpcHandlers.CategoryServer
	grpcServer      *grpc.Server
	port           string
	logger         *logger.Logger
}

// NewServer 创建新的 gRPC 服务器
func NewServer(
	itemService service.ItemService,
	categoryService *grpcHandlers.CategoryServer,
	port string,
	logger *logger.Logger,
) *Server {
	return &Server{
		itemService:     itemService,
		categoryService: categoryService,
		port:           port,
		logger:         logger,
	}
}

// Start 启动 gRPC 服务器
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	// 配置 gRPC 服务器选项
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.MaxRecvMsgSize(4 * 1024 * 1024),  // 4MB
	grpc.MaxSendMsgSize(4 * 1024 * 1024),  // 4MB
	}

	s.grpcServer = grpc.NewServer(opts...)

	// 注册服务处理器
	itemHandler := grpcHandlers.NewItemServer(s.itemService, s.logger)
	itemsv1.RegisterItemServiceServer(s.grpcServer, itemHandler)

	if s.categoryService != nil {
		itemsv1.RegisterCategoryServiceServer(s.grpcServer, s.categoryService)
	}

	// 启用反射（用于开发调试）
	reflection.Register(s.grpcServer)

	s.logger.Infof("Starting gRPC server on port %s", s.port)

	return s.grpcServer.Serve(lis)
}

// Stop 停止 gRPC 服务器
func (s *Server) Stop() error {
	if s.grpcServer == nil {
		return nil
	}

	s.logger.Info("Stopping gRPC server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.grpcServer.GracefulStop()
	<-ctx.Done()

	s.logger.Info("gRPC server stopped")
	return nil
}

// GetServer 返回 grpc.Server 实例
func (s *Server) GetServer() *grpc.Server {
	return s.grpcServer
}