package router

import (
	"github.com/gin-gonic/gin"
	"github.com/julesChu12/custos/internal/interface/http/handler"
	"github.com/julesChu12/custos/internal/interface/http/middleware"
)

type Router struct {
	authHandler   *handler.AuthHandler
	userHandler   *handler.UserHandler
	healthHandler *handler.HealthHandler
	authMW        *middleware.AuthMiddleware
}

func NewRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	healthHandler *handler.HealthHandler,
	authMW *middleware.AuthMiddleware,
) *Router {
	return &Router{
		authHandler:   authHandler,
		userHandler:   userHandler,
		healthHandler: healthHandler,
		authMW:        authMW,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS())

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", r.healthHandler.Check)

		auth := v1.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)
			auth.POST("/refresh", r.authHandler.Refresh)
		}

		authProtected := v1.Group("/auth")
		authProtected.Use(r.authMW.RequireAuth())
		{
			authProtected.POST("/logout", r.authHandler.Logout)
			authProtected.POST("/logout-all", r.authHandler.LogoutAll)
		}

		user := v1.Group("/user")
		user.Use(r.authMW.RequireAuth())
		{
			user.GET("/profile", r.userHandler.GetProfile)
		}

		admin := v1.Group("/admin")
		admin.Use(r.authMW.RequireAuth())
		admin.Use(r.authMW.RequireRole("admin"))
		{
		}
	}

	return router
}
