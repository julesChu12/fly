package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/custos/internal/application/dto"
	"github.com/julesChu12/custos/internal/application/usecase/auth"
	"github.com/julesChu12/custos/internal/interface/http/middleware"
	"github.com/julesChu12/custos/pkg/errors"
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
		})
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
		})
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

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
		})
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
