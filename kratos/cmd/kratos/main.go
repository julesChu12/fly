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

	_ "github.com/julesChu12/fly/kratos/docs" // Import swagger docs
	"github.com/julesChu12/fly/kratos/api/proto/order/v1"
	"github.com/julesChu12/fly/kratos/internal/application/service"
	cacheinfra "github.com/julesChu12/fly/kratos/internal/infrastructure/cache"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/custos"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/database"
	grpcInterface "github.com/julesChu12/fly/kratos/internal/interface/grpc"
	httpInterface "github.com/julesChu12/fly/kratos/internal/interface/http"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
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

	// Initialize optional cache
	var serviceOpts []service.Option
	var cacheCloser func() error
	if config.Cache.Enabled {
		orderCache, err := cacheinfra.NewOrderCache(config.Cache.Address, config.Cache.DB)
		if err != nil {
			log.Printf("Warning: failed to connect to cache: %v", err)
		} else {
			serviceOpts = append(serviceOpts, service.WithOrderCache(orderCache, config.Cache.TTL))
			cacheCloser = orderCache.Close
		}
	}

	// Initialize services
	orderService := service.NewOrderService(orderRepo, orderItemRepo, statusLogRepo, auditRepo, serviceOpts...)

	if cacheCloser != nil {
		defer cacheCloser()
	}

	// Initialize Custos client (if custos_endpoint is configured)
	var custosClient *custos.Client
	if config.Auth.CustosEndpoint != "" {
		custosClient, err = custos.NewClient(config.Auth.CustosEndpoint)
		if err != nil {
			log.Printf("Warning: failed to connect to Custos at %s: %v", config.Auth.CustosEndpoint, err)
			log.Println("Running without authentication middleware")
		} else {
			log.Printf("Successfully connected to Custos at %s", config.Auth.CustosEndpoint)
			defer custosClient.Close()
		}
	}

	// Start servers
	httpServer, grpcServer := startServers(config, orderService, custosClient)

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

func startServers(config *Config, orderService service.OrderService, custosClient *custos.Client) (*http.Server, *grpc.Server) {
	// HTTP router with auth and rate-limit configuration
	routerCfg := httpInterface.RouterConfig{
		CustosClient:     custosClient,
		SkipAuthPaths:    config.Auth.SkipPaths,
		RateLimitEnabled: config.RateLimit.Enabled,
		RateLimitRPS:     config.RateLimit.RequestsPerSecond,
		RateLimitBurst:   config.RateLimit.Burst,
	}
	router := httpInterface.NewRouter(orderService, routerCfg)
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

	var grpcLimiter *rate.Limiter
	if config.RateLimit.Enabled && config.RateLimit.RequestsPerSecond > 0 {
		burst := config.RateLimit.Burst
		if burst <= 0 {
			burst = int(config.RateLimit.RequestsPerSecond)
			if burst == 0 {
				burst = 1
			}
		}
		grpcLimiter = rate.NewLimiter(rate.Limit(config.RateLimit.RequestsPerSecond), burst)
	}

	var grpcOpts []grpc.ServerOption
	var interceptors []grpc.UnaryServerInterceptor

	// Add logging interceptor
	interceptors = append(interceptors, grpcInterface.UnaryServerInterceptor())

	// Add rate limit interceptor if enabled
	if grpcLimiter != nil {
		interceptors = append(interceptors, grpcInterface.RateLimitInterceptor(grpcLimiter))
	}

	// Add auth interceptor if Custos client is available
	if custosClient != nil {
		authInterceptor := grpcInterface.NewGRPCAuthInterceptor(custosClient)
		interceptors = append(interceptors, authInterceptor.UnaryInterceptor())
	}

	// Add context injector interceptor
	interceptors = append(interceptors, grpcInterface.ContextInjectorInterceptor())

	grpcOpts = append(grpcOpts, grpc.ChainUnaryInterceptor(interceptors...))
	grpcServer := grpc.NewServer(grpcOpts...)

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
	Cache struct {
		Enabled bool          `mapstructure:"enabled"`
		Address string        `mapstructure:"address"`
		DB      int           `mapstructure:"db"`
		TTL     time.Duration `mapstructure:"ttl"`
	} `mapstructure:"cache"`
	Auth struct {
		JWTSecret       string   `mapstructure:"jwt_secret"`
		CustosEndpoint  string   `mapstructure:"custos_endpoint"`
		SkipPaths       []string `mapstructure:"skip_paths"`
	} `mapstructure:"auth"`
	RateLimit struct {
		Enabled           bool    `mapstructure:"enabled"`
		RequestsPerSecond float64 `mapstructure:"requests_per_second"`
		Burst             int     `mapstructure:"burst"`
	} `mapstructure:"rate_limit"`
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
	viper.SetDefault("cache.enabled", false)
	viper.SetDefault("cache.address", "localhost:6379")
	viper.SetDefault("cache.db", 0)
	viper.SetDefault("cache.ttl", "5m")
	viper.SetDefault("auth.jwt_secret", "")
	viper.SetDefault("auth.custos_endpoint", "localhost:50051")
	viper.SetDefault("auth.skip_paths", []string{"/health", "/ready", "/metrics"})
	viper.SetDefault("rate_limit.enabled", false)
	viper.SetDefault("rate_limit.requests_per_second", 1000.0)
	viper.SetDefault("rate_limit.burst", 2000)

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
