package cmd

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

	pb "github.com/julesChu12/fly/plutus/api/proto"
	"github.com/julesChu12/fly/plutus/internal/application/service"
	"github.com/julesChu12/fly/plutus/internal/infrastructure/database"
	grpcInterface "github.com/julesChu12/fly/plutus/internal/interface/grpc"
	httpInterface "github.com/julesChu12/fly/plutus/internal/interface/http"
	"github.com/julesChu12/fly/plutus/pkg/observability"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
		log.Fatalf("Failed to load config: %v", err)
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

	// Initialize database
	db, err := initDatabase(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	walletRepo := database.NewWalletRepository(db)
	transactionRepo := database.NewTransactionRepository(db)
	channelRepo := database.NewPaymentChannelRepository(db)

	// Initialize services
	walletService := service.NewWalletService(db, walletRepo, transactionRepo, channelRepo)

	// Initialize HTTP router
	router := httpInterface.NewRouter(walletService)
	engine := router.SetupRoutes()

	// Start health check server
	healthServer := observability.NewHealthCheckServer(db, 8081)
	go func() {
		if err := healthServer.Start(); err != nil {
			log.Printf("Health check server error: %v", err)
		}
	}()

	// Start metrics server
	metricsServer := observability.NewMetricsServer(9090)
	go func() {
		if err := metricsServer.Start(); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Start servers
	httpServer, grpcServerInstance, grpcListener := startServers(config, engine, walletService)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Plutus server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
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
		log.Println("gRPC server shutdown timeout, forcing stop")
		grpcServerInstance.Stop()
	case <-stopped:
		log.Println("gRPC server stopped gracefully")
	}

	log.Println("Plutus server stopped")
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
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("plutus")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	// Set defaults
	viper.SetDefault("server.http_port", 8085)
	viper.SetDefault("server.grpc_port", 9085)
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.dsn", "root:password@tcp(localhost:3306)/plutus?charset=utf8mb4&parseTime=True&loc=Local")

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
		return nil, err
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connection established successfully")
	return db, nil
}

func startServers(config *Config, engine http.Handler, walletService service.WalletService) (*http.Server, *grpc.Server, net.Listener) {
	// Start HTTP server
	httpServer := &http.Server{
		Addr:         formatPort(config.Server.HTTPPort),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Starting Plutus HTTP server on port %d", config.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start gRPC server
	lis, err := net.Listen("tcp", formatPort(config.Server.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", config.Server.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()
	walletHandler := grpcInterface.NewWalletGRPCHandler(walletService)
	pb.RegisterWalletServiceServer(grpcServer, walletHandler)

	go func() {
		log.Printf("Starting gRPC server on port %d", config.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
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
