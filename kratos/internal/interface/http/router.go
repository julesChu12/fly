package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/kratos/internal/application/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	orderHandler *OrderHandler
}

func NewRouter(orderService service.OrderService) *Router {
	return &Router{
		orderHandler: NewOrderHandler(orderService),
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Middleware
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(r.corsMiddleware())

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Readiness check
	engine.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Prometheus metrics
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := engine.Group("/api")
	{
		orders := api.Group("/orders")
		{
			orders.POST("", r.orderHandler.CreateOrder)
			orders.GET("", r.orderHandler.ListOrders)
			orders.GET("/:id", r.orderHandler.GetOrder)
			orders.GET("/:id/items", r.orderHandler.GetOrderWithItems)
			orders.PATCH("/:id/status", r.orderHandler.UpdateOrderStatus)
			orders.DELETE("/:id", r.orderHandler.DeleteOrder)
		}
	}

	return engine
}

func (r *Router) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Tenant-ID, X-User-ID, X-Trace-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
