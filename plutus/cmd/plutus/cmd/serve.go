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

	pb "github.com/julesChu12/fly/plutus/api/proto"
	"github.com/julesChu12/fly/plutus/internal/application/service"
	"github.com/julesChu12/fly/plutus/internal/infrastructure/database"
	grpcInterface "github.com/julesChu12/fly/plutus/internal/interface/grpc"
	httpInterface "github.com/julesChu12/fly/plutus/internal/interface/http"
	"github.com/julesChu12/fly/plutus/pkg/observability"
	"github.com/julesChu12/fly/mora/pkg/config"
	moraDB "github.com/julesChu12/fly/mora/pkg/db"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Plutus HTTP and gRPC servers",
	Long:  `Start the Plutus HTTP and gRPC servers to handle payment and wallet requests.`,
	Run:   runServer,
}

var (
	configPath string
	httpPort   int
	grpcPort   int
	dbDSN      string
)

func init() {
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "configs/plutus.yaml", "Path to configuration file")
	serveCmd.Flags().IntVarP(&httpPort, "http-port", "p", 0, "HTTP server port (default: 8085 or from config)")
	serveCmd.Flags().IntVarP(&grpcPort, "grpc-port", "g", 0, "gRPC server port (default: 9085 or from config)")
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

	l.Infof("Starting Plutus Payment Service...")
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
	walletRepo := database.NewWalletRepository(dbClient.DB())
	transactionRepo := database.NewTransactionRepository(dbClient.DB())
	channelRepo := database.NewPaymentChannelRepository(dbClient.DB())

	// Initialize services
	walletService := service.NewWalletService(dbClient.DB(), walletRepo, transactionRepo, channelRepo)

	// Initialize HTTP router
	router := httpInterface.NewRouter(walletService)
	engine := router.SetupRoutes()

	// Start health check server (disabled - HTTP server has /health endpoint)
	// Temporary comment: health check runs on separate port which conflicts with other services
	// Using HTTP server's health endpoint instead

	// Start metrics server
	metricsServer := observability.NewMetricsServer(9090)
	go func() {
		if err := metricsServer.Start(); err != nil {
			l.Warnf("Metrics server error: %v", err)
		}
	}()

	// Start servers
	httpServer, grpcServerInstance, grpcListener := startServers(config, engine, walletService, l)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down Plutus server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		l.Errorf("HTTP server forced to shutdown: %v", err)
	} else {
		l.Info("HTTP server stopped gracefully")
	}

	// Shutdown gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServerInstance.GracefulStop()
		grpcListener.Close()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		l.Warn("gRPC server shutdown timeout, forcing stop")
		grpcServerInstance.Stop()
	case <-stopped:
		l.Info("gRPC server stopped gracefully")
	}

	l.Info("Plutus server stopped")
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
}

func loadConfig() (*Config, error) {
	v := config.New().
		WithYAML(configPath).
		WithEnvPrefix("PLUTUS").
		MustLoad()

	// Set defaults
	v.SetDefault("server.http_port", 8085)
	v.SetDefault("server.grpc_port", 9085)
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.dsn", "root:password@tcp(localhost:3306)/plutus?charset=utf8mb4&parseTime=True&loc=Local")
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 3600)
	v.SetDefault("database.log_level", "info")
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

func startServers(config *Config, engine http.Handler, walletService service.WalletService, l *logger.Logger) (*http.Server, *grpc.Server, net.Listener) {
	// Start HTTP server
	httpServer := &http.Server{
		Addr:         formatPort(config.Server.HTTPPort),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		l.Infof("Starting Plutus HTTP server on port %d", config.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start gRPC server
	lis, err := net.Listen("tcp", formatPort(config.Server.GRPCPort))
	if err != nil {
		l.Fatalf("Failed to listen on port %d: %v", config.Server.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	walletHandler := grpcInterface.NewWalletGRPCHandler(walletService)
	pb.RegisterWalletServiceServer(grpcServer, walletHandler)

	go func() {
		l.Infof("Starting gRPC server on port %d", config.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatalf("gRPC server failed: %v", err)
		}
	}()

	return httpServer, grpcServer, lis
}

func formatPort(port int) string {
	if port == 0 {
		return ":8085"
	}
	return fmt.Sprintf(":%d", port)
}
