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
	"github.com/julesChu12/fly/mora/pkg/config"
	moraDB "github.com/julesChu12/fly/mora/pkg/db"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/staff/internal/application/service"
	"github.com/julesChu12/fly/staff/internal/infrastructure/database"
	grpcInterface "github.com/julesChu12/fly/staff/internal/interface/grpc"
	httpInterface "github.com/julesChu12/fly/staff/internal/interface/http"
	"github.com/julesChu12/fly/staff/internal/interface/http/middleware"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Staff HTTP and gRPC servers",
	Long:  `Start the Staff HTTP and gRPC servers to handle staff management requests.`,
	Run:   runServer,
}

var (
	configPath string
	httpPort   int
	grpcPort   int
	dbDSN      string
)

func init() {
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "configs/staff.yaml", "Path to configuration file")
	serveCmd.Flags().IntVarP(&httpPort, "http-port", "p", 0, "HTTP server port (default: 8084 or from config)")
	serveCmd.Flags().IntVarP(&grpcPort, "grpc-port", "g", 0, "gRPC server port (default: 9084 or from config)")
	serveCmd.Flags().StringVarP(&dbDSN, "db-dsn", "d", "", "Database DSN (overrides config)")
}

func runServer(cmd *cobra.Command, args []string) {
	// 如果命令行指定了不同的配置文件，通过环境变量传递给配置加载器
	if configPath != "configs/staff.yaml" {
		fmt.Printf("Using custom config file: %s\n", configPath)
		os.Setenv("CONFIG_PATH", configPath)
	}

	// Load configuration using mora config loader
	cfg, err := loadConfig()
	if err != nil {
		log := logger.NewDefault()
		log.Error("Failed to load config", "error", err)
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
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	l.Infof("Starting Staff Service...")
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
	staffRepo := database.NewStaffRepository(dbClient.DB())
	roleRepo := database.NewStaffRoleRepository(dbClient.DB())
	availabilityRepo := database.NewStaffAvailabilityRepository(dbClient.DB())

	// Initialize Service layer
	staffService := service.NewStaffService(staffRepo, roleRepo, availabilityRepo)

	// Start servers
	httpServer, grpcServer := startServers(cfg, staffService, l)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down Staff service...")

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
		grpcServer.Stop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		l.Warn("gRPC server shutdown timeout, forcing stop")
	case <-stopped:
		l.Info("gRPC server stopped gracefully")
	}

	l.Info("Staff service stopped")
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
	// 使用环境变量CONFIG_PATH或默认路径
	cfgPath := configPath
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}

	v := config.New().
		WithYAML(cfgPath).
		WithEnvPrefix("STAFF").
		MustLoad()

	// Set defaults
	v.SetDefault("server.http_port", 8084)
	v.SetDefault("server.grpc_port", 9084)
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
			database = "staff"
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

// startServers starts HTTP and gRPC servers
func startServers(
	cfg *Config,
	staffService service.StaffService,
	l *logger.Logger,
) (*http.Server, *grpcInterface.Server) {
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
			"service": "staff",
		})
	})

	// API routes
	api := r.Group("/api")
	staffHandler := httpInterface.NewStaffHandler(staffService)
	staffHandler.RegisterRoutes(api)

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

	// Start HTTP server
	go func() {
		l.Infof("Starting HTTP server on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	// Start gRPC server
	grpcServer := grpcInterface.NewServer(staffService, fmt.Sprintf("%d", cfg.Server.GRPCPort), l)
	go func() {
		if err := grpcServer.Start(); err != nil {
			l.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	return httpServer, grpcServer
}
