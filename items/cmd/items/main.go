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

	"go.uber.org/zap"

	"github.com/julesChu12/fly/items/internal/infrastructure/http/router"
	"github.com/julesChu12/fly/mora/pkg/config"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// @title Items Service API
// @version 1.0
// @description 统一商品管理服务 API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8086
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token for authentication
func main() {
	// Initialize configuration
	cfgLoader := config.New().WithYAML("configs/items.yaml")
	cfg, err := cfgLoader.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	zapLogger := logger.NewDefault()
	defer zapLogger.Sync()

	zapLogger.Info("Starting Items Service")

	// Setup Gin router
	ginRouter := router.SetupRouter(cfg)

	// Get server port
	port := cfg.GetString("server.port")
	if port == "" {
		port = "8086"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      ginRouter,
		ReadTimeout:  time.Duration(cfg.GetInt("server.read_timeout")) * time.Second,
		WriteTimeout: time.Duration(cfg.GetInt("server.write_timeout")) * time.Second,
		IdleTimeout:  time.Duration(cfg.GetInt("server.idle_timeout")) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		zapLogger.Info("Starting HTTP server", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLogger.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited")
	fmt.Println("Items Service stopped gracefully")
}