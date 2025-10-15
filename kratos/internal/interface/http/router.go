package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/kratos/internal/application/service"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/custos"
	"github.com/julesChu12/fly/kratos/internal/interface/http/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

// RouterConfig holds HTTP transport specific settings.
type RouterConfig struct {
	CustosClient     *custos.Client
	SkipAuthPaths    []string
	RateLimitEnabled bool
	RateLimitRPS     float64
	RateLimitBurst   int
}

type Router struct {
	orderHandler *OrderHandler
	cfg          RouterConfig
	rateLimiter  *rate.Limiter
}

func NewRouter(orderService service.OrderService, cfg RouterConfig) *Router {
	var limiter *rate.Limiter
	if cfg.RateLimitEnabled && cfg.RateLimitRPS > 0 {
		burst := cfg.RateLimitBurst
		if burst <= 0 {
			burst = int(cfg.RateLimitRPS)
			if burst == 0 {
				burst = 1
			}
		}
		limiter = rate.NewLimiter(rate.Limit(cfg.RateLimitRPS), burst)
	}

	return &Router{
		orderHandler: NewOrderHandler(orderService),
		cfg:          cfg,
		rateLimiter:  limiter,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(r.corsMiddleware())

	if r.rateLimiter != nil {
		engine.Use(r.rateLimitMiddleware())
	}

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().UTC()})
	})

	engine.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := engine.Group("/api")

	// Apply authentication middleware if Custos client is available
	if r.cfg.CustosClient != nil {
		skipPaths := append([]string(nil), r.cfg.SkipAuthPaths...)
		// Ensure health-style endpoints are always accessible.
		skipDefaults := []string{"/health", "/ready", "/metrics", "/swagger/*"}
		skipPaths = append(skipPaths, skipDefaults...)

		authMW := middleware.NewAuthMiddleware(r.cfg.CustosClient, skipPaths)
		api.Use(authMW.RequireAuth())
	}

	orders := api.Group("/orders")
	{
		orders.POST("", r.orderHandler.CreateOrder)
		orders.GET("", r.orderHandler.ListOrders)
		orders.GET("/:id", r.orderHandler.GetOrder)
		orders.GET("/:id/items", r.orderHandler.GetOrderWithItems)
		orders.GET("/:id/logs", r.orderHandler.GetOrderLogs)
		orders.PATCH("/:id/status", r.orderHandler.UpdateOrderStatus)
		orders.DELETE("/:id", r.orderHandler.DeleteOrder)
	}

	return engine
}

func (r *Router) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Tenant-ID, X-User-ID, X-Trace-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (r *Router) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		if !r.rateLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "too_many_requests",
				"message": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
