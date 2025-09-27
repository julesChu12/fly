package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/http/handler"
	"github.com/julesChu12/fly/clotho/internal/middleware"
	ginAdapter "github.com/julesChu12/fly/mora/adapters/gin"
	"github.com/spf13/viper"
)

// SetupRouter initializes and configures the Gin router with all routes and middleware
func SetupRouter(cfg *viper.Viper) *gin.Engine {
	// Set Gin mode based on configuration
	mode := cfg.GetString("app.mode")
	if mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router
	router := gin.New()

	// Add global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add OpenTelemetry observability middleware
	serviceName := cfg.GetString("observability.service_name")
	if serviceName == "" {
		serviceName = "clotho"
	}
	router.Use(ginAdapter.ObservabilityMiddleware(serviceName))

	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Health check endpoint (no auth required)
	router.GET("/health", handler.HealthCheck)

	// Initialize dependencies (defer gRPC connection until needed)
	custosAddress := cfg.GetString("services.custos.address")
	if custosAddress == "" {
		custosAddress = "localhost:50051" // default
	}

	// Create user proxy with lazy gRPC client initialization
	userProxy := usecase.NewUserProxyUseCase(nil, 30*time.Second)
	userHandler := handler.NewUserHandler(userProxy)

	// TODO: Initialize gRPC clients when first needed or use connection pool

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.GetString("jwt.secret"))

	// API v1 routes (auth required)
	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware.ValidateToken())
	{
		// User routes
		users := v1.Group("/users")
		{
			users.GET("/me", handler.GetCurrentUser)
			users.GET("/:id", userHandler.GetUserByID)
		}

		// Future route groups for orders, payments, etc.
		// orders := v1.Group("/orders")
		// payments := v1.Group("/payments")
	}

	return router
}