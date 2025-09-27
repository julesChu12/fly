package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	TenantID int64  `json:"tenant_id,omitempty"`
}

// UserHandler contains dependencies for user-related handlers
type UserHandler struct {
	userProxy *usecase.UserProxyUseCase
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userProxy *usecase.UserProxyUseCase) *UserHandler {
	return &UserHandler{
		userProxy: userProxy,
	}
}

func GetCurrentUser(c *gin.Context) {
	// Log request with trace context
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Getting current user information")

	// Extract user information from middleware context
	userID, exists := c.Get("user_id")
	if !exists {
		log.Warn("User ID not found in token")
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

	log.Info("Current user information retrieved successfully")
	c.JSON(http.StatusOK, response)
}

// GetUserByID retrieves user information by ID using the UserHandler
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// Log request with trace context
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Getting user by ID")

	// Parse user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Warn("Invalid user ID provided", "user_id", userIDStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_user_id",
			"message": "User ID must be a valid integer",
		})
		return
	}

	log.Info("Calling user proxy to get user information", "user_id", userID)

	// Call use case to get user information
	userInfo, err := h.userProxy.GetUserByID(userID)
	if err != nil {
		log.Error("Failed to retrieve user information", "user_id", userID, "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_server_error",
			"message": "Failed to retrieve user information",
		})
		return
	}

	// Convert to response format
	response := UserResponse{
		ID:       userInfo.ID,
		Username: userInfo.Username,
		Email:    userInfo.Email,
		UserType: userInfo.UserType,
		TenantID: userInfo.TenantID,
	}

	log.Info("User information retrieved successfully", "user_id", userID)
	c.JSON(http.StatusOK, response)
}