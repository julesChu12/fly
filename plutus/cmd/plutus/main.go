package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/julesChu12/fly/plutus/api/proto"
	"github.com/julesChu12/fly/plutus/internal/application/service"
	"github.com/julesChu12/fly/plutus/internal/infrastructure/database"
	grpcInterface "github.com/julesChu12/fly/plutus/internal/interface/grpc"
	httpInterface "github.com/julesChu12/fly/plutus/internal/interface/http"
	"github.com/julesChu12/fly/plutus/pkg/observability"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title Plutus Payment & Wallet Service API
// @version 1.0
// @description Payment and wallet service for Fly platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8085
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
	walletRepo := database.NewWalletRepository(db)
	transactionRepo := database.NewTransactionRepository(db)
	channelRepo := database.NewPaymentChannelRepository(db)

	// Initialize services
	walletService := service.NewWalletService(db, walletRepo, transactionRepo, channelRepo)

	// Initialize HTTP router
	router := httpInterface.NewRouter(walletService)
	engine := router.SetupRoutes()

	// Start health check server in background
	healthServer := observability.NewHealthCheckServer(db, 8081)
	go func() {
		if err := healthServer.Start(); err != nil {
			log.Printf("Health check server error: %v", err)
		}
	}()

	// Start metrics server in background
	metricsServer := observability.NewMetricsServer(9090)
	go func() {
		if err := metricsServer.Start(); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Start gRPC server in background
	go func() {
		if err := startGRPCServer(config, walletService); err != nil {
			log.Fatal("gRPC server failed:", err)
		}
	}()

	// Start main HTTP server in background
	go func() {
		serverAddr := fmt.Sprintf(":%d", config.Server.HTTPPort)
		log.Printf("Starting Plutus HTTP server on %s", serverAddr)
		if err := http.ListenAndServe(serverAddr, engine); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Plutus server...")
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

// startGRPCServer starts the gRPC server
// 启动gRPC服务器
func startGRPCServer(config *Config, walletService service.WalletService) error {
	port := config.Server.GRPCPort
	if port == 0 {
		port = 9085
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	s := grpc.NewServer()

	// Register gRPC services
	walletHandler := grpcInterface.NewWalletGRPCHandler(walletService)
	pb.RegisterWalletServiceServer(s, walletHandler)

	log.Printf("gRPC server starting on port %d", port)
	return s.Serve(lis)
}
