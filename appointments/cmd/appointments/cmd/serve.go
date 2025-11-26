package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/appointments/internal/application/service"
	"github.com/julesChu12/fly/appointments/internal/domain/repository"
	"github.com/julesChu12/fly/appointments/internal/infrastructure/database"
	httpInterface "github.com/julesChu12/fly/appointments/internal/interface/http"
	"github.com/julesChu12/fly/appointments/internal/interface/http/middleware"
	"github.com/julesChu12/fly/mora/pkg/config"
	moraDB "github.com/julesChu12/fly/mora/pkg/db"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Appointments HTTP and gRPC servers",
	Long:  `Start the Appointments HTTP and gRPC servers to handle appointment management requests.`,
	Run:   runServer,
}

var (
	configPath string
	httpPort   int
	grpcPort   int
	dbDSN      string
)

func init() {
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "configs/appointments.yaml", "Path to configuration file")
	serveCmd.Flags().IntVarP(&httpPort, "http-port", "p", 0, "HTTP server port (default: 8083 or from config)")
	serveCmd.Flags().IntVarP(&grpcPort, "grpc-port", "g", 0, "gRPC server port (default: 9083 or from config)")
	serveCmd.Flags().StringVarP(&dbDSN, "db-dsn", "d", "", "Database DSN (overrides config)")
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration using mora config loader
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Override with command line flags
	if httpPort > 0 {
		cfg.Server.HTTPPort = httpPort
	}
	if grpcPort > 0 {
		cfg.Server.GRPCPort = grpcPort
	}
	if dbDSN != "" {
		cfg.Database.DSN = dbDSN
	}

	// Initialize logger using mora logger
	loggerCfg := logger.Config{
		Level:        cfg.Logger.Level,
		Format:       cfg.Logger.Format,
		OutputPath:   cfg.Logger.OutputPath,
		MaxSize:      cfg.Logger.MaxSize,
		MaxBackups:   cfg.Logger.MaxBackups,
		MaxAge:       cfg.Logger.MaxAge,
		Compress:     cfg.Logger.Compress,
		EnableStdout: cfg.Logger.EnableStdout,
		EnableFile:   cfg.Logger.EnableFile,
	}

	l, err := logger.New(loggerCfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	l.Infof("Starting Appointments Service...")
	l.Infof("Server configuration - HTTP:%d, gRPC:%d", cfg.Server.HTTPPort, cfg.Server.GRPCPort)

	// Initialize database using mora db client
	dbConfig := moraDB.Config{
		Driver:          cfg.Database.Driver,
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		LogLevel:        cfg.Database.LogLevel,
	}

	dbClient, err := moraDB.New(dbConfig)
	if err != nil {
		l.Fatalf("Failed to connect database: %v", err)
	}
	defer dbClient.Close()

	l.Info("Database connected successfully")

	// Initialize Repository layer
	appointmentRepo := database.NewAppointmentRepository(dbClient.DB())

	// Initialize Service layer
	appointmentService := service.NewAppointmentService(appointmentRepo)

	// Start servers
	httpServer := startServers(cfg, appointmentService, appointmentRepo, l)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down Appointments service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		l.Errorf("HTTP server forced to shutdown: %v", err)
	} else {
		l.Info("HTTP server stopped gracefully")
	}

	l.Info("Appointments service stopped")
}

// Config holds the application configuration
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

// loadConfig loads configuration from file and environment
func loadConfig() (*Config, error) {
	v := config.New().
		WithYAML(configPath).
		WithEnvPrefix("APPOINTMENTS").
		MustLoad()

	// Set defaults
	v.SetDefault("server.http_port", 8083)
	v.SetDefault("server.grpc_port", 9083)
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.max_open_conns", 10)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 3600)
	v.SetDefault("database.log_level", "warn")
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.enable_stdout", true)
	v.SetDefault("logger.enable_file", false)
	v.SetDefault("logger.max_size", 100)
	v.SetDefault("logger.max_backups", 10)
	v.SetDefault("logger.max_age", 30)
	v.SetDefault("logger.compress", false)

	// Build database DSN from config if not already set
	if !v.IsSet("database.dsn") {
		host := v.GetString("database.host")
		port := v.GetString("database.port")
		username := v.GetString("database.username")
		password := v.GetString("database.password")
		database := v.GetString("database.database")
		charset := v.GetString("database.charset")

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "3306"
		}
		if username == "" {
			username = "root"
		}
		if password == "" {
			password = "password"
		}
		if database == "" {
			database = "appointments"
		}
		if charset == "" {
			charset = "utf8mb4"
		}

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
			username, password, host, port, database, charset)
		v.Set("database.dsn", dsn)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// startServers starts HTTP server
func startServers(
	cfg *Config,
	appointmentService service.AppointmentService,
	appointmentRepo repository.AppointmentRepository,
	l *logger.Logger,
) *http.Server {
	// Start HTTP server
	r := gin.New()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Add TraceID middleware (MUST be early in chain)
	r.Use(middleware.TraceIDMiddleware())

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "appointments",
		})
	})

	// API routes
	api := r.Group("/api")
	appointmentHandler := httpInterface.NewAppointmentHandler(appointmentService)
	appointmentHandler.RegisterRoutes(api)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	httpAddr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	l.Infof("HTTP server configured on %s", httpAddr)

	return httpServer
}