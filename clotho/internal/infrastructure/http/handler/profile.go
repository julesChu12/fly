package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/clotho/internal/validation"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error   string                     `json:"error"`
	Message string                     `json:"message"`
	Details []validation.ValidationError `json:"details,omitempty"`
}

// ProfileRequest represents the request payload for updating user profile
type ProfileRequest struct {
	Username    string            `json:"username,omitempty"`
	Email       string            `json:"email,omitempty"`
	FirstName   string            `json:"first_name,omitempty"`
	LastName    string            `json:"last_name,omitempty"`
	Avatar      string            `json:"avatar,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	Phone       string            `json:"phone,omitempty"`
	Location    string            `json:"location,omitempty"`
	Website     string            `json:"website,omitempty"`
	Preferences map[string]string `json:"preferences,omitempty"`
}

// ProfileResponse represents the user profile information
type ProfileResponse struct {
	ID          int64             `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	FirstName   string            `json:"first_name,omitempty"`
	LastName    string            `json:"last_name,omitempty"`
	Avatar      string            `json:"avatar,omitempty"`
	Bio         string            `json:"bio,omitempty"`
	Phone       string            `json:"phone,omitempty"`
	Location    string            `json:"location,omitempty"`
	Website     string            `json:"website,omitempty"`
	UserType    string            `json:"user_type"`
	TenantID    int64             `json:"tenant_id,omitempty"`
	Status      string            `json:"status"`
	Preferences map[string]string `json:"preferences,omitempty"`
	Statistics  map[string]int64  `json:"statistics,omitempty"`
	CreatedAt   string            `json:"created_at,omitempty"`
	UpdatedAt   string            `json:"updated_at,omitempty"`
}

// ProfileHandler contains dependencies for profile-related handlers
type ProfileHandler struct {
	userProxy *usecase.UserProxyUseCase
}

// NewProfileHandler creates a new ProfileHandler instance
func NewProfileHandler(userProxy *usecase.UserProxyUseCase) *ProfileHandler {
	return &ProfileHandler{
		userProxy: userProxy,
	}
}

// GetProfile godoc
// @Summary Get current user's complete profile
// @Description Retrieve the complete profile information for the authenticated user, including preferences and statistics
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProfileResponse "User profile retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/profile [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Getting user profile")

	// Extract user ID from middleware context
	userID, exists := c.Get("user_id")
	if !exists {
		log.Warn("User ID not found in token")
		c.JSON(http.StatusUnauthorized, ValidationErrorResponse{
			Error:   "unauthorized",
			Message: "User ID not found in token",
		})
		return
	}

	// Get complete user profile using orchestration use case
	profile, err := h.userProxy.GetCurrentUserProfile(userID.(int64))
	if err != nil {
		log.Error("Failed to retrieve user profile", "user_id", userID, "error", err.Error())
		c.JSON(http.StatusInternalServerError, ValidationErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve user profile",
		})
		return
	}

	// Convert to response format
	response := ProfileResponse{
		ID:          profile.User.ID,
		Username:    profile.User.Username,
		Email:       profile.User.Email,
		FirstName:   profile.User.FirstName,
		LastName:    profile.User.LastName,
		Avatar:      profile.User.Avatar,
		Bio:         profile.User.Bio,
		Phone:       profile.User.Phone,
		Location:    profile.User.Location,
		Website:     profile.User.Website,
		UserType:    profile.User.UserType,
		TenantID:    profile.User.TenantID,
		Status:      profile.User.Status,
		Preferences: profile.Preferences,
		Statistics:  profile.Statistics,
		CreatedAt:   profile.User.CreatedAt,
		UpdatedAt:   profile.User.UpdatedAt,
	}

	log.Info("User profile retrieved successfully", "user_id", userID)
	c.JSON(http.StatusOK, response)
}

// UpdateProfile godoc
// @Summary Update current user's profile
// @Description Update the profile information for the authenticated user
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body ProfileRequest true "Profile update request"
// @Success 200 {object} ProfileResponse "Profile updated successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid payload"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/profile [put]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Updating user profile")

	// Extract user ID from middleware context
	userID, exists := c.Get("user_id")
	if !exists {
		log.Warn("User ID not found in token")
		c.JSON(http.StatusUnauthorized, ValidationErrorResponse{
			Error:   "unauthorized",
			Message: "User ID not found in token",
		})
		return
	}

	// Parse request body
	var req ProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid request payload", "error", err.Error())
		c.JSON(http.StatusBadRequest, ValidationErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request payload",
		})
		return
	}

	// Validate profile update request
	validationRequest := validation.ProfileUpdateRequest{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Avatar:    req.Avatar,
		Bio:       req.Bio,
		Phone:     req.Phone,
		Location:  req.Location,
		Website:   req.Website,
	}

	if validationErrors := validation.ValidateProfileUpdate(validationRequest); len(validationErrors) > 0 {
		log.Warn("Profile validation failed", "errors", validationErrors.Error())
		c.JSON(http.StatusBadRequest, ValidationErrorResponse{
			Error:   "validation_failed",
			Message: "Profile validation failed",
			Details: validationErrors,
		})
		return
	}

	// Prepare updates map
	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Website != "" {
		updates["website"] = req.Website
	}

	// Update profile through use case
	updatedProfile, err := h.userProxy.UpdateUserProfile(userID.(int64), updates)
	if err != nil {
		log.Error("Failed to update user profile", "user_id", userID, "error", err.Error())
		c.JSON(http.StatusInternalServerError, ValidationErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to update user profile",
		})
		return
	}

	// Update preferences if provided
	if req.Preferences != nil && len(req.Preferences) > 0 {
		err = h.userProxy.UpdateUserPreferences(userID.(int64), req.Preferences)
		if err != nil {
			log.Warn("Failed to update user preferences", "user_id", userID, "error", err.Error())
			// Don't fail the entire request for preference update failure
		}
	}

	// Return updated profile
	response := ProfileResponse{
		ID:          updatedProfile.User.ID,
		Username:    updatedProfile.User.Username,
		Email:       updatedProfile.User.Email,
		FirstName:   updatedProfile.User.FirstName,
		LastName:    updatedProfile.User.LastName,
		Avatar:      updatedProfile.User.Avatar,
		Bio:         updatedProfile.User.Bio,
		Phone:       updatedProfile.User.Phone,
		Location:    updatedProfile.User.Location,
		Website:     updatedProfile.User.Website,
		UserType:    updatedProfile.User.UserType,
		TenantID:    updatedProfile.User.TenantID,
		Status:      updatedProfile.User.Status,
		Preferences: updatedProfile.Preferences,
		Statistics:  updatedProfile.Statistics,
		CreatedAt:   updatedProfile.User.CreatedAt,
		UpdatedAt:   updatedProfile.User.UpdatedAt,
	}

	log.Info("User profile updated successfully", "user_id", userID)
	c.JSON(http.StatusOK, response)
}

// GetUserProfile godoc
// @Summary Get public profile of a user
// @Description Retrieve public profile information for a specific user by ID
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} ProfileResponse "User profile retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/profile/users/{id} [get]
func (h *ProfileHandler) GetUserProfile(c *gin.Context) {
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Getting user profile by ID")

	// Parse user ID from URL parameter
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Warn("Invalid user ID provided", "user_id", userIDStr)
		c.JSON(http.StatusBadRequest, ValidationErrorResponse{
			Error:   "invalid_user_id",
			Message: "User ID must be a valid integer",
		})
		return
	}

	// Get user profile (public information only)
	profile, err := h.userProxy.GetCurrentUserProfile(userID)
	if err != nil {
		log.Error("Failed to retrieve user profile", "user_id", userID, "error", err.Error())
		c.JSON(http.StatusInternalServerError, ValidationErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve user profile",
		})
		return
	}

	// Return only public profile information
	response := ProfileResponse{
		ID:       profile.User.ID,
		Username: profile.User.Username,
		UserType: profile.User.UserType,
		Status:   profile.User.Status,
		// Note: Email and other sensitive info excluded for public profiles
	}

	log.Info("User profile retrieved successfully", "user_id", userID)
	c.JSON(http.StatusOK, response)
}

// UpdatePreferences godoc
// @Summary Update user preferences
// @Description Update the preferences for the authenticated user
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param preferences body object true "User preferences as key-value pairs"
// @Success 200 {object} object "Preferences updated successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid preferences payload"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing token"
// @Router /api/v1/profile/preferences [put]
func (h *ProfileHandler) UpdatePreferences(c *gin.Context) {
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Updating user preferences")

	// Extract user ID from middleware context
	userID, exists := c.Get("user_id")
	if !exists {
		log.Warn("User ID not found in token")
		c.JSON(http.StatusUnauthorized, ValidationErrorResponse{
			Error:   "unauthorized",
			Message: "User ID not found in token",
		})
		return
	}

	// Parse preferences from request
	var preferences map[string]string
	if err := c.ShouldBindJSON(&preferences); err != nil {
		log.Warn("Invalid preferences payload", "error", err.Error())
		c.JSON(http.StatusBadRequest, ValidationErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid preferences payload",
		})
		return
	}

	// Validate preferences
	if validationErrors := validation.ValidatePreferences(preferences); len(validationErrors) > 0 {
		log.Warn("Preferences validation failed", "errors", validationErrors.Error())
		c.JSON(http.StatusBadRequest, ValidationErrorResponse{
			Error:   "validation_failed",
			Message: "Preferences validation failed",
			Details: validationErrors,
		})
		return
	}

	// Update preferences through use case
	updateErr := h.userProxy.UpdateUserPreferences(userID.(int64), preferences)
	if updateErr != nil {
		log.Error("Failed to update user preferences", "user_id", userID, "error", updateErr.Error())
		c.JSON(http.StatusInternalServerError, ValidationErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to update user preferences",
		})
		return
	}

	log.Info("User preferences updated successfully", "user_id", userID, "preferences_count", len(preferences))

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":     "Preferences updated successfully",
		"preferences": preferences,
	})
}