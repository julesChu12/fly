package http

import (
	"github.com/julesChu12/fly/plutus/internal/application/service"
	"github.com/julesChu12/fly/plutus/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	walletHandler *WalletHandler
}

func NewRouter(walletService service.WalletService) *Router {
	return &Router{
		walletHandler: NewWalletHandler(walletService),
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// Global middleware
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(r.corsMiddleware())
	engine.Use(middleware.TraceIDMiddleware())

	// Health check (no auth required)
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger documentation (no auth required)
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes with authentication
	api := engine.Group("/api")
	api.Use(middleware.TenantMiddleware())
	api.Use(middleware.UserIDMiddleware())
	{
		// Wallet routes
		wallets := api.Group("/wallets")
		{
			wallets.POST("", r.walletHandler.CreateWallet)
			wallets.GET("", r.walletHandler.ListWallets)
			wallets.GET("/:id", r.walletHandler.GetWallet)
			wallets.GET("/customer/:customer_id", r.walletHandler.GetWalletByCustomerID)
		}

		// Transaction routes
		transactions := api.Group("/transactions")
		{
			transactions.POST("/recharge", r.walletHandler.Recharge)
			transactions.POST("/consume", r.walletHandler.Consume)
			transactions.POST("/refund", r.walletHandler.Refund)
			transactions.GET("", r.walletHandler.ListTransactions)
			transactions.GET("/:id", r.walletHandler.GetTransaction)
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
