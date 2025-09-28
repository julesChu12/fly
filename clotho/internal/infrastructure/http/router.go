package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/http/handler"
	"github.com/julesChu12/fly/clotho/internal/middleware"
	ginAdapter "github.com/julesChu12/fly/mora/adapters/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Import docs for swagger generation
	_ "github.com/julesChu12/fly/clotho/docs"
)

// setDefaults sets default values for all configuration keys
func setDefaults(cfg *viper.Viper) {
	// App configuration
	cfg.SetDefault("app.mode", "development")

	// Observability configuration
	cfg.SetDefault("observability.service_name", "clotho")

	// Service configuration
	cfg.SetDefault("services.custos.address", "localhost:50051")

	// Rate limiter configuration
	cfg.SetDefault("rate_limiter.global_rps", 1000.0)
	cfg.SetDefault("rate_limiter.global_burst", 2000)
	cfg.SetDefault("rate_limiter.per_ip_rps", 10.0)
	cfg.SetDefault("rate_limiter.per_ip_burst", 20)
	cfg.SetDefault("rate_limiter.per_user_rps", 100.0)
	cfg.SetDefault("rate_limiter.per_user_burst", 200)

	// Circuit breaker configuration
	cfg.SetDefault("circuit_breaker.max_requests", 5)
	cfg.SetDefault("circuit_breaker.failure_threshold", 5)
	cfg.SetDefault("circuit_breaker.failure_ratio", 0.6)
	cfg.SetDefault("circuit_breaker.min_requests", 10)

	// JWT configuration
	cfg.SetDefault("jwt.secret", "your-secret-key")
}

// SetupRouter initializes and configures the Gin router with all routes and middleware
func SetupRouter(cfg *viper.Viper) *gin.Engine {
	// Set default values for configuration
	setDefaults(cfg)

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
	router.Use(ginAdapter.ObservabilityMiddleware(serviceName))

	// Add enhanced metrics middleware
	metricsMiddleware := middleware.NewMetricsMiddleware()
	router.Use(metricsMiddleware.Middleware())

	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Add rate limiting middleware
	rateLimiterConfig := middleware.RateLimiterConfig{
		GlobalRPS:    cfg.GetFloat64("rate_limiter.global_rps"),
		GlobalBurst:  cfg.GetInt("rate_limiter.global_burst"),
		PerIPRPS:     cfg.GetFloat64("rate_limiter.per_ip_rps"),
		PerIPBurst:   cfg.GetInt("rate_limiter.per_ip_burst"),
		PerUserRPS:   cfg.GetFloat64("rate_limiter.per_user_rps"),
		PerUserBurst: cfg.GetInt("rate_limiter.per_user_burst"),
	}

	rateLimiter := middleware.NewRateLimiter(rateLimiterConfig)
	router.Use(rateLimiter.Middleware())

	// Add circuit breaker middleware
	circuitBreakerConfig := middleware.CircuitBreakerConfig{
		MaxRequests:      uint32(cfg.GetInt("circuit_breaker.max_requests")),
		FailureThreshold: uint32(cfg.GetInt("circuit_breaker.failure_threshold")),
		FailureRatio:     cfg.GetFloat64("circuit_breaker.failure_ratio"),
		MinRequests:      uint32(cfg.GetInt("circuit_breaker.min_requests")),
	}

	circuitBreaker := middleware.NewCircuitBreakerManager(circuitBreakerConfig)
	router.Use(circuitBreaker.Middleware())

	// Health check endpoint (no auth required)
	router.GET("/health", handler.HealthCheck)

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Prometheus metrics endpoint
	registry := metricsMiddleware.GetMetricsRegistry().GetRegistry()
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	// Create user proxy with lazy gRPC client initialization
	userProxy := usecase.NewUserProxyUseCase(nil, 30*time.Second)
	custosAddress := cfg.GetString("services.custos.address")
	custosDialTimeout := 5 * time.Second
	userProxy.SetCustosClientFactory(func() (*client.CustosClient, error) {
		return client.NewCustosClient(custosAddress, custosDialTimeout)
	})
	userHandler := handler.NewUserHandler(userProxy)
	profileHandler := handler.NewProfileHandler(userProxy)
	monitoringHandler := handler.NewMiddlewareStatsHandler(rateLimiter, circuitBreaker, metricsMiddleware.GetMetricsRegistry())

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

		// Profile routes - comprehensive user profile management
		profile := v1.Group("/profile")
		{
			profile.GET("/", profileHandler.GetProfile)                   // GET /api/v1/profile - Get current user's complete profile
			profile.PUT("/", profileHandler.UpdateProfile)                // PUT /api/v1/profile - Update current user's profile
			profile.PUT("/preferences", profileHandler.UpdatePreferences) // PUT /api/v1/profile/preferences - Update user preferences
			profile.GET("/users/:id", profileHandler.GetUserProfile)      // GET /api/v1/profile/users/:id - Get another user's public profile
		}

		// Monitoring routes - middleware statistics and management
		monitoring := v1.Group("/monitoring")
		{
			monitoring.GET("/stats", monitoringHandler.GetAllStats)                           // GET /api/v1/monitoring/stats - All middleware stats
			monitoring.GET("/rate-limiter", monitoringHandler.GetRateLimiterStats)            // GET /api/v1/monitoring/rate-limiter - Rate limiter stats
			monitoring.GET("/circuit-breaker", monitoringHandler.GetCircuitBreakerStats)      // GET /api/v1/monitoring/circuit-breaker - Circuit breaker stats
			monitoring.POST("/circuit-breaker/reset", monitoringHandler.ResetCircuitBreakers) // POST /api/v1/monitoring/circuit-breaker/reset - Reset circuit breakers
		}

		// Future route groups for orders, payments, etc.
		// orders := v1.Group("/orders")
		// payments := v1.Group("/payments")
	}

	return router
}
