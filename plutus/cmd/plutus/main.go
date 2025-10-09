package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julesChu12/fly/plutus/internal/application/service"
	"github.com/julesChu12/fly/plutus/internal/infrastructure/database"
	httpInterface "github.com/julesChu12/fly/plutus/internal/interface/http"
	"github.com/spf13/viper"
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
	walletService := service.NewWalletService(walletRepo, transactionRepo, channelRepo)

	// Initialize HTTP router
	router := httpInterface.NewRouter(walletService)
	engine := router.SetupRoutes()

	// Start server
	serverAddr := fmt.Sprintf(":%d", config.Server.HTTPPort)
	log.Printf("Starting Plutus server on %s", serverAddr)

	if err := http.ListenAndServe(serverAddr, engine); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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
