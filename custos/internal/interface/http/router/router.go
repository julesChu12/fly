package router

import (
	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/interface/http/handler"
	"github.com/julesChu12/fly/custos/internal/interface/http/middleware"
)

type Router struct {
	authHandler   *handler.AuthHandler
	userHandler   *handler.UserHandler
	oauthHandler  *handler.OAuthHandler
	adminHandler  *handler.AdminHandler
	healthHandler *handler.HealthHandler
	authMW        *middleware.AuthMiddleware
}

func NewRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	oauthHandler *handler.OAuthHandler,
	adminHandler *handler.AdminHandler,
	healthHandler *handler.HealthHandler,
	authMW *middleware.AuthMiddleware,
) *Router {
	return &Router{
		authHandler:   authHandler,
		userHandler:   userHandler,
		oauthHandler:  oauthHandler,
		adminHandler:  adminHandler,
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

		// OAuth routes
		oauth := v1.Group("/oauth")
		{
			oauth.GET("/:provider/login", r.oauthHandler.GetOAuthURL)
			oauth.GET("/:provider/callback", r.oauthHandler.HandleOAuthCallback)
		}

		oauthProtected := v1.Group("/oauth")
		oauthProtected.Use(r.authMW.RequireAuth())
		{
			oauthProtected.POST("/:provider/bind", r.oauthHandler.BindOAuthProvider)
			oauthProtected.DELETE("/:provider/unbind", r.oauthHandler.UnbindOAuthProvider)
			oauthProtected.GET("/bindings", r.oauthHandler.GetUserOAuthBindings)
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
			admin.GET("/users", r.adminHandler.ListUsers)
			admin.GET("/users/:id", r.adminHandler.GetUser)
			admin.PATCH("/users/:id/status", r.adminHandler.UpdateUserStatus)
			admin.PATCH("/users/:id/role", r.adminHandler.UpdateUserRole)
			admin.POST("/users/:id/force-logout", r.adminHandler.ForceLogoutUser)
			admin.GET("/stats", r.adminHandler.GetSystemStats)
		}
	}

	return router
}
