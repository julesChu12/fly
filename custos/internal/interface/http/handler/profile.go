package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
)

type ProfileHandler struct {
	userRepo        repository.UserRepository
	userProfileRepo repository.UserProfileRepository
}

func NewProfileHandler(userRepo repository.UserRepository, userProfileRepo repository.UserProfileRepository) *ProfileHandler {
	return &ProfileHandler{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
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

	// Get user profile
	profile, err := h.userProfileRepo.GetByUserID(c.Request.Context(), uid)
	if err != nil {
		if err == repository.ErrUserProfileNotFound {
			// Profile doesn't exist, create a default one
			profile = entity.NewUserProfile(uid)
			if err := h.userProfileRepo.Create(c.Request.Context(), profile); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile"})
			return
		}
	}

	// Convert to DTO
	response := dto.GetProfileResponse{
		UserID:   profile.UserID,
		Nickname: profile.Nickname,
		Avatar:   profile.Avatar,
		Gender:   profile.Gender,
		Extra:    profile.Extra,
	}

	if profile.Birthday != nil {
		response.Birthday = profile.Birthday.Format("2006-01-02")
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

	// Get existing profile or create new one
	profile, err := h.userProfileRepo.GetByUserID(c.Request.Context(), uid)
	if err != nil {
		if err == repository.ErrUserProfileNotFound {
			// Create new profile
			profile = entity.NewUserProfile(uid)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get profile"})
			return
		}
	}

	// Update profile fields
	if req.Nickname != "" {
		profile.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		profile.Avatar = req.Avatar
	}
	if req.Gender != "" {
		profile.Gender = req.Gender
	}
	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birthday format, use YYYY-MM-DD"})
			return
		}
		profile.Birthday = &birthday
	}
	if req.Extra != "" {
		profile.Extra = req.Extra
	}

	// Save profile
	if err == repository.ErrUserProfileNotFound {
		// Create
		if err := h.userProfileRepo.Create(c.Request.Context(), profile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile"})
			return
		}
	} else {
		// Update
		if err := h.userProfileRepo.Update(c.Request.Context(), profile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
			return
		}
	}

	response := dto.UpdateProfileResponse{
		UserID:  uid,
		Message: "profile updated successfully",
	}

	c.JSON(http.StatusOK, response)
}
