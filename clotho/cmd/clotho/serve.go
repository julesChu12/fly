package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	httpRouter "github.com/julesChu12/fly/clotho/internal/infrastructure/http"
	"github.com/julesChu12/fly/mora/pkg/config"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/mora/pkg/observability"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Clotho HTTP server",
	Long:  `Start the Clotho HTTP server to handle API orchestration requests.`,
	Run:   runServer,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("config", "c", "configs/clotho.yaml", "Path to configuration file")
	serveCmd.Flags().StringP("port", "p", "8080", "Port to run the server on")
}

func runServer(cmd *cobra.Command, args []string) {
	// 加载配置文件
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatalf("无法获取配置文件路径: %v", err)
	}
	cfg, err := config.New().WithYAML(configPath).Load()
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// Initialize logger
	loggerConfig := logger.Config{
		Level:  cfg.GetString("logging.level"),
		Format: cfg.GetString("logging.format"),
	}
	logger, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize OpenTelemetry observability
	observabilityConfig := observability.Config{
		ServiceName:  cfg.GetString("observability.service_name"),
		ExporterURL:  cfg.GetString("observability.exporter_url"),
		SampleRatio:  cfg.GetFloat64("observability.sample_ratio"),
		Environment:  cfg.GetString("observability.environment"),
		ExporterType: cfg.GetString("observability.exporter_type"),
	}

	// Set defaults if not configured
	if observabilityConfig.ServiceName == "" {
		observabilityConfig.ServiceName = "clotho"
	}
	if observabilityConfig.ExporterURL == "" {
		observabilityConfig.ExporterURL = "http://localhost:4317"
	}
	if observabilityConfig.SampleRatio == 0 {
		observabilityConfig.SampleRatio = 1.0
	}
	if observabilityConfig.Environment == "" {
		observabilityConfig.Environment = "development"
	}
	if observabilityConfig.ExporterType == "" {
		observabilityConfig.ExporterType = "stdout"
	}

	cleanup, err := observability.Init(observabilityConfig)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to initialize observability: %v", err))
	}
	defer cleanup()

	logger.Info("OpenTelemetry observability initialized")

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router using the router package
	router := httpRouter.SetupRouter(cfg)

	// Get port from command line or config
	port, _ := cmd.Flags().GetString("port")
	if port == "" {
		port = cfg.GetString("server.port")
	}
	if port == "" {
		port = "8080" // default
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	logger.Info(fmt.Sprintf("Starting Clotho server on port %s", port))

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(fmt.Sprintf("Failed to start server: %v", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	logger.Info("Server exited")
}
