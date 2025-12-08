package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderv1 "github.com/julesChu12/fly/kratos/api/proto/order/v1"
	_ "github.com/julesChu12/fly/kratos/docs" // Import swagger docs
	"github.com/julesChu12/fly/kratos/internal/application/service"
	cacheinfra "github.com/julesChu12/fly/kratos/internal/infrastructure/cache"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/custos"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/database"
	grpcInterface "github.com/julesChu12/fly/kratos/internal/interface/grpc"
	httpInterface "github.com/julesChu12/fly/kratos/internal/interface/http"
	"github.com/julesChu12/fly/mora/pkg/config"
	moraDB "github.com/julesChu12/fly/mora/pkg/db"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Kratos HTTP and gRPC servers",
	Long:  `Start the Kratos HTTP and gRPC servers to handle order management requests.`,
	Run:   runServer,
}

var (
	configPath string
	httpPort   int
	grpcPort   int
	dbDSN      string
)

func init() {
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "configs/kratos.yaml", "Path to configuration file")
	serveCmd.Flags().IntVarP(&httpPort, "http-port", "p", 0, "HTTP server port (default: 8082 or from config)")
	serveCmd.Flags().IntVarP(&grpcPort, "grpc-port", "g", 0, "gRPC server port (default: 9092 or from config)")
	serveCmd.Flags().StringVarP(&dbDSN, "db-dsn", "d", "", "Database DSN (overrides config)")
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Override with command line flags
	if httpPort > 0 {
		config.Server.HTTPPort = httpPort
	}
	if grpcPort > 0 {
		config.Server.GRPCPort = grpcPort
	}
	if dbDSN != "" {
		config.Database.DSN = dbDSN
	}

	// Initialize logger using mora logger
	loggerCfg := logger.Config{
		Level:  config.Logger.Level,
		Format: config.Logger.Format,
	}

	l, err := logger.New(loggerCfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	l.Infof("Starting Kratos Order Service...")
	l.Infof("Server configuration - HTTP:%d, gRPC:%d", config.Server.HTTPPort, config.Server.GRPCPort)

	// Initialize database using mora db client
	dbConfig := moraDB.Config{
		Driver:          config.Database.Driver,
		DSN:             config.Database.DSN,
		MaxOpenConns:    config.Database.MaxOpenConns,
		MaxIdleConns:    config.Database.MaxIdleConns,
		ConnMaxLifetime: config.Database.ConnMaxLifetime,
		LogLevel:        config.Database.LogLevel,
	}

	dbClient, err := moraDB.New(dbConfig)
	if err != nil {
		l.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbClient.Close()

	l.Info("Database connected successfully")

	// Initialize repositories
	orderRepo := database.NewOrderRepository(dbClient.DB())
	orderItemRepo := database.NewOrderItemRepository(dbClient.DB())
	statusLogRepo := database.NewOrderStatusLogRepository(dbClient.DB())
	auditRepo := database.NewOrderAuditRepository(dbClient.DB())

	// Initialize optional cache
	var serviceOpts []service.Option
	var cacheCloser func() error
	if config.Cache.Enabled {
		orderCache, err := cacheinfra.NewOrderCache(config.Cache.Address, config.Cache.DB)
		if err != nil {
			l.Warnf("Failed to connect to cache: %v", err)
		} else {
			serviceOpts = append(serviceOpts, service.WithOrderCache(orderCache, config.Cache.TTL))
			cacheCloser = orderCache.Close
			l.Info("Cache connected successfully")
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
			l.Warnf("Failed to connect to Custos at %s: %v", config.Auth.CustosEndpoint, err)
			l.Warn("Running without authentication middleware")
		} else {
			l.Infof("Successfully connected to Custos at %s", config.Auth.CustosEndpoint)
			defer custosClient.Close()
		}
	}

	// Start servers
	httpServer, grpcServer := startServers(config, orderService, custosClient, l)

	// Wait for interrupt signal to gracefully shutdown the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down servers...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		l.Errorf("HTTP server forced to shutdown: %v", err)
	} else {
		l.Info("HTTP server stopped gracefully")
	}

	// Graceful stop gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		l.Warn("gRPC server shutdown timeout, forcing stop")
		grpcServer.Stop()
	case <-stopped:
		l.Info("gRPC server stopped gracefully")
	}

	l.Info("Servers exited")
}

func startServers(config *Config, orderService service.OrderService, custosClient *custos.Client, l *logger.Logger) (*http.Server, *grpc.Server) {
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
		l.Infof("Starting HTTP server on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", config.Server.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		l.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
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
		l.Infof("Starting gRPC server on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatalf("gRPC server failed: %v", err)
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
		Driver          string `mapstructure:"driver"`
		DSN             string `mapstructure:"dsn"`
		MaxOpenConns    int    `mapstructure:"max_open_conns"`
		MaxIdleConns    int    `mapstructure:"max_idle_conns"`
		ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
		LogLevel        string `mapstructure:"log_level"`
	} `mapstructure:"database"`
	Cache struct {
		Enabled bool          `mapstructure:"enabled"`
		Address string        `mapstructure:"address"`
		DB      int           `mapstructure:"db"`
		TTL     time.Duration `mapstructure:"ttl"`
	} `mapstructure:"cache"`
	Auth struct {
		JWTSecret      string   `mapstructure:"jwt_secret"`
		CustosEndpoint string   `mapstructure:"custos_endpoint"`
		SkipPaths      []string `mapstructure:"skip_paths"`
	} `mapstructure:"auth"`
	RateLimit struct {
		Enabled           bool    `mapstructure:"enabled"`
		RequestsPerSecond float64 `mapstructure:"requests_per_second"`
		Burst             int     `mapstructure:"burst"`
	} `mapstructure:"rate_limit"`
	Logger struct {
		Level        string `mapstructure:"level"`
		Format       string `mapstructure:"format"`
		OutputPath   string `mapstructure:"output_path"`
		MaxSize      int    `mapstructure:"max_size"`
		MaxBackups   int    `mapstructure:"max_backups"`
		MaxAge       int    `mapstructure:"max_age"`
		Compress     bool   `mapstructure:"compress"`
		EnableStdout bool   `mapstructure:"enable_stdout"`
		EnableFile   bool   `mapstructure:"enable_file"`
	} `mapstructure:"logger"`
	Observability struct {
		ServiceName string `mapstructure:"service_name"`
		Endpoint    string `mapstructure:"endpoint"`
	} `mapstructure:"observability"`
}

func loadConfig() (*Config, error) {
	v := config.New().
		WithYAML(configPath).
		WithEnvPrefix("KRATOS").
		MustLoad()

	// Set defaults
	v.SetDefault("server.http_port", 8082)
	v.SetDefault("server.grpc_port", 9092)
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.dsn", "root:password@tcp(localhost:3306)/kratos?charset=utf8mb4&parseTime=True&loc=Local")
	v.SetDefault("database.max_open_conns", 10)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 3600)
	v.SetDefault("database.log_level", "warn")
	v.SetDefault("cache.enabled", false)
	v.SetDefault("cache.address", "localhost:6379")
	v.SetDefault("cache.db", 0)
	v.SetDefault("cache.ttl", "5m")
	v.SetDefault("auth.jwt_secret", "")
	v.SetDefault("auth.custos_endpoint", "localhost:50051")
	v.SetDefault("auth.skip_paths", []string{"/health", "/ready", "/metrics"})
	v.SetDefault("rate_limit.enabled", false)
	v.SetDefault("rate_limit.requests_per_second", 1000.0)
	v.SetDefault("rate_limit.burst", 2000)
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.enable_stdout", true)
	v.SetDefault("logger.enable_file", false)
	v.SetDefault("logger.max_size", 100)
	v.SetDefault("logger.max_backups", 10)
	v.SetDefault("logger.max_age", 30)
	v.SetDefault("logger.compress", false)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
