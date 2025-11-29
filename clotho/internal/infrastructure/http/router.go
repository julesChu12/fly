package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/http/handler"
	"github.com/julesChu12/fly/clotho/internal/interface/http/auth"
	"github.com/julesChu12/fly/clotho/internal/middleware"
	ginAdapter "github.com/julesChu12/fly/mora/adapters/gin"
	"github.com/julesChu12/fly/mora/pkg/logger"
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
	cfg.SetDefault("services.custos.http.address", "http://localhost:8081")
	cfg.SetDefault("services.hermes.address", "http://localhost:8080")
	cfg.SetDefault("services.plutus.address", "http://localhost:8085")
	cfg.SetDefault("services.kratos.address", "http://localhost:8082")
	cfg.SetDefault("services.appointments.address", "http://localhost:8083")
	cfg.SetDefault("services.staff.address", "http://localhost:8084")
	cfg.SetDefault("services.items.address", "http://localhost:8086")

	// Rate limiter configuration
	cfg.SetDefault("rate_limiter.global_rps", 1000.0)
	cfg.SetDefault("rate_limiter.global_burst", 2000)
	cfg.SetDefault("rate_limiter.per_ip_rps", 10.0)
	cfg.SetDefault("rate_limiter.per_ip_burst", 20)
	cfg.SetDefault("rate_limiter.per_user_rps", 100.0)
	cfg.SetDefault("rate_limiter.per_user_burst", 200)
	cfg.SetDefault("rate_limiter.cleanup_interval", "5m")
	cfg.SetDefault("rate_limiter.max_idle_time", "10m")

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
		CleanupInterval: cfg.GetDuration("rate_limiter.cleanup_interval"),
		MaxIdleTime:     cfg.GetDuration("rate_limiter.max_idle_time"),
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

	// Initialize logger for auth module
	authLogger := logger.NewDefault()

	// Create Custos HTTP client for auth endpoints
	custosHTTPAddress := cfg.GetString("services.custos.http.address")
	custosTimeout := 30 * time.Second
	custosHTTPClient := client.NewCustosHTTPClient(custosHTTPAddress, custosTimeout, authLogger)

	// Initialize auth handler
	authHandler := auth.NewAuthHandler(custosHTTPClient, authLogger)

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

	// Create staff proxy with HTTP client
	staffAddress := cfg.GetString("services.staff.address")
	staffTimeout := 10 * time.Second
	staffClient := client.NewStaffHTTPClient(staffAddress, staffTimeout)
	staffLogger := logger.NewDefault()
	staffProxy := usecase.NewStaffProxy(staffClient, staffLogger)
	staffHandler := handler.NewStaffHandler(staffProxy)

	// Create appointment proxy with HTTP client
	appointmentsAddress := cfg.GetString("services.appointments.address")
	appointmentsTimeout := 10 * time.Second
	appointmentsClient := client.NewAppointmentHTTPClient(appointmentsAddress, appointmentsTimeout)
	appointmentsLogger := logger.NewDefault()
	appointmentsProxy := usecase.NewAppointmentProxy(appointmentsClient, appointmentsLogger)
	appointmentsHandler := handler.NewAppointmentHandler(appointmentsProxy)

	// Create customer proxy with HTTP client
	hermesAddress := cfg.GetString("services.hermes.address")
	hermesTimeout := 10 * time.Second
	customerClient := client.NewCustomerHTTPClient(hermesAddress, hermesTimeout)
	customerLogger := logger.NewDefault()
	customerProxy := usecase.NewCustomerProxy(customerClient, customerLogger)
	customerHandler := handler.NewCustomerHandler(customerProxy)

	// Create order proxy with HTTP client
	kratosAddress := cfg.GetString("services.kratos.address")
	kratosTimeout := 10 * time.Second
	orderClient := client.NewOrderHTTPClient(kratosAddress, kratosTimeout)
	orderLogger := logger.NewDefault()
	orderProxy := usecase.NewOrderProxy(orderClient, orderLogger)
	orderHandler := handler.NewOrderHandler(orderProxy)

	// Create payment proxy with HTTP client
	plutusAddress := cfg.GetString("services.plutus.address")
	plutusTimeout := 10 * time.Second
	paymentClient := client.NewPaymentHTTPClient(plutusAddress, plutusTimeout)
	paymentLogger := logger.NewDefault()
	paymentProxy := usecase.NewPaymentProxy(paymentClient, paymentLogger)
	paymentHandler := handler.NewPaymentHandler(paymentProxy)

	// Create items proxy with HTTP client
	itemsAddress := cfg.GetString("services.items.address")
	itemsTimeout := 10 * time.Second
	itemsClient := client.NewItemsHTTPClient(itemsAddress, itemsTimeout, logger.NewDefault())
	itemsHandler := handler.NewItemsHandler(itemsClient, logger.NewDefault())
	categoriesHandler := handler.NewCategoriesHandler(itemsClient, logger.NewDefault())

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

		// Service route groups
		customers := v1.Group("/customers")
		{
			// Customer service routes - proxy to hermes service
			customerHandler.RegisterRoutes(customers)
		}

		appointments := v1.Group("/appointments")
		{
			// Appointments service routes - proxy to appointments service
			appointmentsHandler.RegisterRoutes(appointments)
		}

		employees := v1.Group("/employees")
		{
			// Staff service routes - proxy to staff service
			staffHandler.RegisterRoutes(employees)
		}

		orders := v1.Group("/orders")
		{
			// Order service routes - proxy to kratos service
			orderHandler.RegisterRoutes(orders)
		}

		payments := v1.Group("/payments")
		{
			// Payment service routes - proxy to plutus service
			paymentHandler.RegisterRoutes(payments)
		}

		// Items service routes - proxy to items service
		items := v1.Group("/items")
		{
			// Item management routes
			itemsHandler.RegisterRoutes(items)
		}

		// Categories service routes - proxy to items service
		categories := v1.Group("/categories")
		{
			// Category management routes
			categoriesHandler.RegisterRoutes(categories)
		}
	}

	// Register auth routes (mix of public and authenticated)
	auth.RegisterAuthRoutes(router, authHandler, authMiddleware, authLogger)

	return router
}
