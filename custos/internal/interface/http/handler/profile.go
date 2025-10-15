package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/application/service"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
)

type ProfileHandler struct {
	profileService *service.UserProfileService
}

func NewProfileHandler(userRepo repository.UserRepository, userProfileRepo repository.UserProfileRepository) *ProfileHandler {
	return &ProfileHandler{
		profileService: service.NewUserProfileService(userRepo, userProfileRepo),
	}
}

// GetProfile gets the current user's profile
// GET /api/v1/profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	// Get authenticated user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
		return
	}

	// Get user profile using service
	response, err := h.profileService.GetProfile(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProfile updates the current user's profile
// PUT /api/v1/profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	// Get authenticated user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update profile using service
	if err := h.profileService.UpdateProfile(c.Request.Context(), uid, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	response := dto.UpdateProfileResponse{
		UserID:  uid,
		Message: "profile updated successfully",
	}

	c.JSON(http.StatusOK, response)
}
