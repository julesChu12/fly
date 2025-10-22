package handler

import (
	"log"
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
// @Summary 获取个人资料
// @Description 获取当前登录用户的个人资料信息
// @Tags 用户配置
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.GetProfileResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /profile [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	// Get authenticated user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, &dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		log.Printf("[ProfileHandler.GetProfile] Invalid user ID type: %T, value: %v", userID, userID)
		c.JSON(http.StatusInternalServerError, &dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Invalid user ID format",
		})
		return
	}

	log.Printf("[ProfileHandler.GetProfile] Fetching profile for user ID: %d", uid)

	// Get user profile using service
	response, err := h.profileService.GetProfile(c.Request.Context(), uid)
	if err != nil {
		log.Printf("[ProfileHandler.GetProfile] Failed to get profile for user %d: %v", uid, err)
		c.JSON(http.StatusInternalServerError, &dto.ErrorResponse{
			Code:    "PROFILE_FETCH_FAILED",
			Message: "Failed to get profile",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProfile updates the current user's profile
// @Summary 更新个人资料
// @Description 更新当前登录用户的个人资料信息
// @Tags 用户配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "资料信息"
// @Success 200 {object} dto.UpdateProfileResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /profile [put]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	// Get authenticated user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, &dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		log.Printf("[ProfileHandler.UpdateProfile] Invalid user ID type: %T, value: %v", userID, userID)
		c.JSON(http.StatusInternalServerError, &dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Invalid user ID format",
		})
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	log.Printf("[ProfileHandler.UpdateProfile] Updating profile for user ID: %d", uid)

	// Update profile using service
	if err := h.profileService.UpdateProfile(c.Request.Context(), uid, &req); err != nil {
		log.Printf("[ProfileHandler.UpdateProfile] Failed to update profile for user %d: %v", uid, err)
		c.JSON(http.StatusInternalServerError, &dto.ErrorResponse{
			Code:    "PROFILE_UPDATE_FAILED",
			Message: "Failed to update profile",
		})
		return
	}

	response := dto.UpdateProfileResponse{
		UserID:  uid,
		Message: "profile updated successfully",
	}

	c.JSON(http.StatusOK, response)
}
