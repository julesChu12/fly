package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	custosClient *client.CustosHTTPClient
	logger       *logger.Logger
}

// NewAuthHandler 创建新的认证处理器
func NewAuthHandler(custosClient *client.CustosHTTPClient, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		custosClient: custosClient,
		logger:       logger,
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户使用用户名和密码登录系统
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body client.LoginRequest true "登录信息"
// @Success 200 {object} handler.SuccessResponse{data=client.LoginResponse}
// @Failure 400 {object} handler.SuccessResponse
// @Failure 500 {object} handler.SuccessResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req client.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid login request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	// 调用 Custos 服务进行登录
	resp, err := h.custosClient.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Login failed", "error", err, "username", req.Username)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "LOGIN_FAILED",
			"message": "Login failed",
		})
		return
	}

	h.logger.Info("User logged in successfully", "username", req.Username, "userID", resp.User.ID)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Login successful",
		"data":    resp,
	})
}

// Register 用户注册
// @Summary 用户注册
// @Description 新用户注册账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body client.RegisterRequest true "注册信息"
// @Success 200 {object} handler.SuccessResponse{data=client.RegisterResponse}
// @Failure 400 {object} handler.SuccessResponse
// @Failure 500 {object} handler.SuccessResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req client.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid register request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	// 调用 Custos 服务进行注册
	resp, err := h.custosClient.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Registration failed", "error", err, "username", req.Username, "email", req.Email)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "REGISTRATION_FAILED",
			"message": "Registration failed",
		})
		return
	}

	h.logger.Info("User registered successfully", "username", req.Username, "email", req.Email, "userID", resp.User.ID)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Registration successful",
		"data":    resp,
	})
}

// RefreshToken 刷新访问令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body client.RefreshTokenRequest true "刷新令牌信息"
// @Success 200 {object} handler.SuccessResponse{data=client.TokenResponse}
// @Failure 400 {object} handler.SuccessResponse
// @Failure 500 {object} handler.SuccessResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req client.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid refresh token request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	// 调用 Custos 服务刷新令牌
	resp, err := h.custosClient.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Token refresh failed", "error", err, "sessionID", req.SessionID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "TOKEN_REFRESH_FAILED",
			"message": "Token refresh failed",
		})
		return
	}

	h.logger.Info("Token refreshed successfully", "sessionID", req.SessionID)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Token refreshed successfully",
		"data":    resp,
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出当前设备
// @Tags 认证
// @Security ApiKeyAuth
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.SuccessResponse
// @Failure 500 {object} handler.SuccessResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从头部获取认证信息
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "MISSING_TOKEN",
			"message": "Access token is required",
		})
		return
	}

	// 移除 Bearer 前缀
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "MISSING_SESSION_ID",
			"message": "Session ID is required",
		})
		return
	}

	// 调用 Custos 服务登出
	if err := h.custosClient.Logout(c.Request.Context(), sessionID, accessToken); err != nil {
		h.logger.Error("Logout failed", "error", err, "sessionID", sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "LOGOUT_FAILED",
			"message": "Logout failed",
		})
		return
	}

	h.logger.Info("User logged out successfully", "sessionID", sessionID)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Logged out successfully",
		"data":    gin.H{"message": "Logged out successfully"},
	})
}

// LogoutAll 登出所有设备
// @Summary 登出所有设备
// @Description 用户登出所有设备
// @Tags 认证
// @Security ApiKeyAuth
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.SuccessResponse
// @Failure 500 {object} handler.SuccessResponse
// @Router /api/v1/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// 从头部获取认证信息
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "MISSING_TOKEN",
			"message": "Access token is required",
		})
		return
	}

	// 移除 Bearer 前缀
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	// 调用 Custos 服务登出所有设备
	if err := h.custosClient.LogoutAll(c.Request.Context(), accessToken); err != nil {
		h.logger.Error("Logout all devices failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "LOGOUT_ALL_FAILED",
			"message": "Logout all devices failed",
		})
		return
	}

	h.logger.Info("User logged out from all devices successfully")
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Logged out from all devices successfully",
		"data":    gin.H{"message": "Logged out from all devices successfully"},
	})
}

// ForgotPassword 忘记密码
// @Summary 忘记密码
// @Description 发送密码重置邮件
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body client.ForgotPasswordRequest true "邮箱信息"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.SuccessResponse
// @Failure 500 {object} handler.SuccessResponse
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req client.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid forgot password request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	// 调用 Custos 服务发送重置密码邮件
	if err := h.custosClient.ForgotPassword(c.Request.Context(), &req); err != nil {
		h.logger.Error("Forgot password failed", "error", err, "email", req.Email)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "FORGOT_PASSWORD_FAILED",
			"message": "Failed to send reset password email",
		})
		return
	}

	h.logger.Info("Password reset email sent successfully", "email", req.Email)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Password reset email sent successfully",
		"data":    gin.H{"message": "Password reset email sent successfully"},
	})
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 使用重置令牌设置新密码
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body client.ResetPasswordRequest true "重置密码信息"
// @Success 200 {object} handler.SuccessResponse
// @Failure 400 {object} handler.SuccessResponse
// @Failure 500 {object} handler.SuccessResponse
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req client.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid reset password request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	// 调用 Custos 服务重置密码
	if err := h.custosClient.ResetPassword(c.Request.Context(), &req); err != nil {
		h.logger.Error("Reset password failed", "error", err, "token", req.Token)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "RESET_PASSWORD_FAILED",
			"message": "Failed to reset password",
		})
		return
	}

	h.logger.Info("Password reset successfully", "token", req.Token)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Password reset successfully",
		"data":    gin.H{"message": "Password reset successfully"},
	})
}