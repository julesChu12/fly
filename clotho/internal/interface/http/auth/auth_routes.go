package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/middleware"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// RegisterAuthRoutes 注册认证相关路由
func RegisterAuthRoutes(r *gin.Engine, authHandler *AuthHandler, authMiddleware *middleware.AuthMiddleware, logger *logger.Logger) {
	// 认证路由组
	authGroup := r.Group("/api/v1/auth")
	{
		// 公开路由（无需认证）
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)

		// 需要认证的路由
		authGroup.POST("/logout", authMiddleware.ValidateToken(), authHandler.Logout)
		authGroup.POST("/logout-all", authMiddleware.ValidateToken(), authHandler.LogoutAll)
	}

	logger.Info("Auth routes registered successfully")
}