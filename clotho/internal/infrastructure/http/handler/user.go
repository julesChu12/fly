package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	TenantID int64  `json:"tenant_id,omitempty"`
}

func GetCurrentUser(c *gin.Context) {
	// Extract user information from middleware context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User ID not found in token",
		})
		return
	}

	userType, _ := c.Get("user_type")
	tenantID, _ := c.Get("tenant_id")

	// TODO: In the future, this should call Custos gRPC service to get full user details
	// For now, return basic information from JWT claims
	response := UserResponse{
		ID:       userID.(int64),
		Username: "user", // This would come from Custos service
		Email:    "user@example.com", // This would come from Custos service
		UserType: userType.(string),
		TenantID: tenantID.(int64),
	}

	c.JSON(http.StatusOK, response)
}