package router

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/julesChu12/fly/items/docs"
	"github.com/julesChu12/fly/items/internal/container"
	"github.com/julesChu12/fly/items/internal/infrastructure/http/handler"
	"github.com/julesChu12/fly/items/internal/infrastructure/http/middleware"
	"github.com/julesChu12/fly/mora/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// setDefaults sets default values for all configuration keys
func setDefaults(cfg *viper.Viper) {
	// App configuration
	cfg.SetDefault("app.mode", "development")

	// Server configuration
	cfg.SetDefault("server.port", "8086")
	cfg.SetDefault("server.host", "0.0.0.0")
	cfg.SetDefault("server.read_timeout", 30)
	cfg.SetDefault("server.write_timeout", 30)
	cfg.SetDefault("server.idle_timeout", 60)

	// Database configuration
	cfg.SetDefault("database.driver", "mysql")
	cfg.SetDefault("database.host", "localhost")
	cfg.SetDefault("database.port", 3306)
	cfg.SetDefault("database.username", "root")
	cfg.SetDefault("database.password", "")
	cfg.SetDefault("database.database", "items_db")
	cfg.SetDefault("database.charset", "utf8mb4")
	cfg.SetDefault("database.parse_time", true)
	cfg.SetDefault("database.loc", "Local")
	cfg.SetDefault("database.max_idle_conns", 10)
	cfg.SetDefault("database.max_open_conns", 100)
	cfg.SetDefault("database.conn_max_lifetime", 3600)

	// Logging configuration
	cfg.SetDefault("logging.level", "info")
	cfg.SetDefault("logging.format", "json")
	cfg.SetDefault("logging.output", "stdout")
}

// SetupRouter initializes and configures the Gin router with all routes and middleware
func SetupRouter(cfg *viper.Viper) (*gin.Engine, *container.Container) {
	// Set default values for configuration
	setDefaults(cfg)

	// Set Gin mode based on configuration
	mode := cfg.GetString("app.mode")
	if mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize dependency container
	cont, err := container.NewContainer(cfg)
	if err != nil {
		panic("Failed to initialize container: " + err.Error())
	}

	// Create router
	router := gin.New()

	// Add global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add CORS middleware
	router.Use(middleware.CORS())

	// Initialize logger
	logger := logger.NewDefault()

	// Health check endpoint (no auth required)
	router.GET("/health", handler.HealthCheck)

	// Swagger documentation endpoint
	docs.SwaggerInfo.BasePath = "/api"  // Initialize docs package
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Item routes
		items := v1.Group("/items")
		{
			items.POST("", cont.ItemHandler.CreateItem)
			items.GET("", cont.ItemHandler.GetItems)
			items.GET("/:id", cont.ItemHandler.GetItemByID)
			items.PUT("/:id", cont.ItemHandler.UpdateItem)
			items.DELETE("/:id", cont.ItemHandler.DeleteItem)
			items.PATCH("/:id/status", cont.ItemHandler.UpdateItemStatus)
		}

		// Category routes
		categories := v1.Group("/categories")
		{
			categories.POST("", cont.CategoryHandler.CreateCategory)
			categories.GET("", cont.CategoryHandler.GetCategories)
			categories.GET("/tree", cont.CategoryHandler.GetCategoryTree)
			categories.GET("/:id", cont.CategoryHandler.GetCategoryByID)
			categories.PUT("/:id", cont.CategoryHandler.UpdateCategory)
			categories.DELETE("/:id", cont.CategoryHandler.DeleteCategory)
		}

		// Search routes
		search := v1.Group("/search")
		{
			search.GET("/items", cont.SearchHandler.SearchItems)
		}

		// Statistics routes
		stats := v1.Group("/stats")
		{
			stats.GET("/overview", cont.StatsHandler.GetOverviewStats)
			stats.GET("/items", cont.StatsHandler.GetItemStats)
		}
	}

	logger.Info("Router initialized successfully")
	return router, cont
}