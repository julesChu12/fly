package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/interface/http/middleware"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)
	userRole := middleware.GetUserRole(c)

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, &dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	userInfo := &dto.UserInfo{
		ID:       userID,
		Username: username,
		Role:     userRole,
	}

	c.JSON(http.StatusOK, &dto.SuccessResponse{
		Data: userInfo,
	})
}
