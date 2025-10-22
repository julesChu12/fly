package router

import (
	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/interface/http/handler"
	"github.com/julesChu12/fly/custos/internal/interface/http/middleware"
)

type Router struct {
	authHandler    *handler.AuthHandler
	oauthHandler   *handler.OAuthHandler
	adminHandler   *handler.AdminHandler
	profileHandler *handler.ProfileHandler
	healthHandler  *handler.HealthHandler
	authMW         *middleware.AuthMiddleware
}

func NewRouter(
	authHandler *handler.AuthHandler,
	oauthHandler *handler.OAuthHandler,
	adminHandler *handler.AdminHandler,
	profileHandler *handler.ProfileHandler,
	healthHandler *handler.HealthHandler,
	authMW *middleware.AuthMiddleware,
) *Router {
	return &Router{
		authHandler:    authHandler,
		oauthHandler:   oauthHandler,
		adminHandler:   adminHandler,
		profileHandler: profileHandler,
		healthHandler:  healthHandler,
		authMW:         authMW,
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

		// Profile routes (user profile management)
		profile := v1.Group("/profile")
		profile.Use(r.authMW.RequireAuth())
		{
			profile.GET("", r.profileHandler.GetProfile)
			profile.PUT("", r.profileHandler.UpdateProfile)
		}

		admin := v1.Group("/admin")
		admin.Use(r.authMW.RequireAuth())
		admin.Use(r.authMW.RequireRole("admin"))
		{
			// User management
			admin.GET("/users", r.adminHandler.ListUsers)
			admin.GET("/users/:id", r.adminHandler.GetUser)
			admin.PATCH("/users/:id/status", r.adminHandler.UpdateUserStatus)
			admin.POST("/users/:id/force-logout", r.adminHandler.ForceLogoutUser)

			// Role management
			admin.POST("/users/:id/roles", r.adminHandler.AssignRole)
			admin.GET("/users/:id/roles", r.adminHandler.GetUserRoles)
			admin.PATCH("/users/:id/role", r.adminHandler.UpdateUserRole) // Backward compatibility

			// Policy management
			admin.POST("/policies", r.adminHandler.AddPolicy)
			admin.DELETE("/policies", r.adminHandler.RemovePolicy)

			// System stats
			admin.GET("/stats", r.adminHandler.GetSystemStats)
		}
	}

	return router
}
