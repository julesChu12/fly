package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/application/usecase/auth"
	"github.com/julesChu12/fly/custos/internal/interface/http/middleware"
	"github.com/julesChu12/fly/custos/pkg/errors"
)

type AuthHandler struct {
	registerUC  *auth.RegisterUseCase
	loginUC     *auth.LoginUseCase
	refreshUC   *auth.RefreshUseCase
	logoutUC    *auth.LogoutUseCase
	logoutAllUC *auth.LogoutAllUseCase
}

func NewAuthHandler(registerUC *auth.RegisterUseCase, loginUC *auth.LoginUseCase, refreshUC *auth.RefreshUseCase, logoutUC *auth.LogoutUseCase, logoutAllUC *auth.LogoutAllUseCase) *AuthHandler {
	return &AuthHandler{
		registerUC:  registerUC,
		loginUC:     loginUC,
		refreshUC:   refreshUC,
		logoutUC:    logoutUC,
		logoutAllUC: logoutAllUC,
	}
}

// Register godoc
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 201 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	userInfo, err := h.registerUC.Execute(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, &dto.SuccessResponse{
		Data: userInfo,
	})
}

// Login godoc
// @Summary 用户登录
// @Description 使用用户名/邮箱和密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录凭证"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	meta := &dto.LoginMetadata{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	loginResp, err := h.loginUC.Execute(c.Request.Context(), &req, meta)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, &dto.SuccessResponse{
		Data: loginResp,
	})
}

// Refresh godoc
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "刷新令牌"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	resp, err := h.refreshUC.Execute(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, &dto.SuccessResponse{Data: resp})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := middleware.GetSessionID(c)
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, &dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Session context missing",
		})
		return
	}

	if err := h.logoutUC.Execute(c.Request.Context(), sessionID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, &dto.SuccessResponse{Data: gin.H{"status": "logged_out"}})
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, &dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	if err := h.logoutAllUC.Execute(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, &dto.SuccessResponse{Data: gin.H{"status": "all_sessions_revoked"}})
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	if domainErr, ok := err.(*errors.DomainError); ok {
		statusCode := h.getStatusCodeFromError(domainErr.Code)
		c.JSON(statusCode, &dto.ErrorResponse{
			Code:    domainErr.Code,
			Message: domainErr.Message,
			Fields:  domainErr.Fields,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, &dto.ErrorResponse{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "Internal server error",
	})
}

func (h *AuthHandler) getStatusCodeFromError(code string) int {
	switch code {
	case errors.CodeUserNotFound, errors.CodeInvalidCredentials:
		return http.StatusUnauthorized
	case errors.CodeUserAlreadyExists:
		return http.StatusConflict
	case errors.CodeInvalidPassword:
		return http.StatusBadRequest
	case errors.CodeTokenExpired, errors.CodeTokenInvalid:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
