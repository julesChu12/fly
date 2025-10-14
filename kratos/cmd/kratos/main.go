package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julesChu12/fly/kratos/api/proto/order/v1"
	"github.com/julesChu12/fly/kratos/internal/application/service"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/database"
	grpcInterface "github.com/julesChu12/fly/kratos/internal/interface/grpc"
	httpInterface "github.com/julesChu12/fly/kratos/internal/interface/http"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title Kratos Order Service API
// @version 1.0
// @description Order service for Fly platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8082
// @BasePath /

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := initDatabase(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	orderRepo := database.NewOrderRepository(db)
	orderItemRepo := database.NewOrderItemRepository(db)
	statusLogRepo := database.NewOrderStatusLogRepository(db)
	auditRepo := database.NewOrderAuditRepository(db)

	// Initialize services
	orderService := service.NewOrderService(orderRepo, orderItemRepo, statusLogRepo, auditRepo)

	// Start servers
	httpServer, grpcServer := startServers(config, orderService)

	// Wait for interrupt signal to gracefully shutdown the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	// Graceful stop gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		log.Println("gRPC server shutdown timeout, forcing stop")
		grpcServer.Stop()
	case <-stopped:
		log.Println("gRPC server stopped gracefully")
	}

	log.Println("Servers exited")
}

func startServers(config *Config, orderService service.OrderService) (*http.Server, *grpc.Server) {
	// Start HTTP server
	router := httpInterface.NewRouter(orderService)
	engine := router.SetupRoutes()

	httpAddr := fmt.Sprintf(":%d", config.Server.HTTPPort)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: engine,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", config.Server.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcInterface.UnaryServerInterceptor(),
			grpcInterface.ContextInjectorInterceptor(),
		),
	)

	// Register gRPC services
	orderv1.RegisterOrderServiceServer(grpcServer, grpcInterface.NewOrderServiceServer(orderService))

	// Register reflection service on gRPC server (for grpcurl, etc.)
	reflection.Register(grpcServer)

	go func() {
		log.Printf("Starting gRPC server on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	return httpServer, grpcServer
}

type Config struct {
	Server struct {
		HTTPPort int `mapstructure:"http_port"`
		GRPCPort int `mapstructure:"grpc_port"`
	} `mapstructure:"server"`
	Database struct {
		Driver string `mapstructure:"driver"`
		DSN    string `mapstructure:"dsn"`
	} `mapstructure:"database"`
	Redis struct {
		Addr string `mapstructure:"addr"`
		DB   int    `mapstructure:"db"`
	} `mapstructure:"redis"`
	Observability struct {
		ServiceName string `mapstructure:"service_name"`
		Endpoint    string `mapstructure:"endpoint"`
	} `mapstructure:"observability"`
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("kratos")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	// Set defaults
	viper.SetDefault("server.http_port", 8082)
	viper.SetDefault("server.grpc_port", 9092)
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.dsn", "root:password@tcp(localhost:3306)/kratos?charset=utf8mb4&parseTime=True&loc=Local")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v. Using defaults.", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func initDatabase(config *Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(config.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")
	return db, nil
}
