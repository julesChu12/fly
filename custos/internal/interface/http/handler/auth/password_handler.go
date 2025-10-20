package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/domain/service/password"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
)

// PasswordHandler handles password-related HTTP requests
type PasswordHandler struct {
	passwordService *password.PasswordService
	userRepo        repository.UserRepository
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(passwordService *password.PasswordService, userRepo repository.UserRepository) *PasswordHandler {
	return &PasswordHandler{
		passwordService: passwordService,
		userRepo:        userRepo,
	}
}

// ValidatePasswordRequest represents a password validation request
type ValidatePasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// ValidatePasswordResponse represents a password validation response
type ValidatePasswordResponse struct {
	IsValid         bool     `json:"is_valid"`
	Errors          []string `json:"errors,omitempty"`
	Score           int      `json:"score"`
	StrengthMessage string   `json:"strength_message"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// PasswordPolicyResponse represents the current password policy
type PasswordPolicyResponse struct {
	MinLength           int  `json:"min_length"`
	RequireUppercase    bool `json:"require_uppercase"`
	RequireLowercase    bool `json:"require_lowercase"`
	RequireNumbers      bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
	ForbidCommonPasswords bool `json:"forbid_common_passwords"`
	MaxRepeatingChars   int  `json:"max_repeating_chars"`
}

// ValidatePassword validates a password without storing it
// @Summary Validate password strength
// @Description Validates a password against the current security policy
// @Tags password
// @Accept json
// @Produce json
// @Param request body ValidatePasswordRequest true "Password to validate"
// @Success 200 {object} ValidatePasswordResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/v1/password/validate [post]
func (h *PasswordHandler) ValidatePassword(c *gin.Context) {
	var req ValidatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// Validate password
	result := h.passwordService.ValidatePassword(req.Password)

	response := ValidatePasswordResponse{
		IsValid:         result.IsValid,
		Errors:          result.Errors,
		Score:           result.Score,
		StrengthMessage: password.GenerateStrengthMessage(result.Score),
	}

	c.JSON(http.StatusOK, response)
}

// ChangePassword changes the current user's password
// @Summary Change user password
// @Description Changes the authenticated user's password
// @Tags password
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Password change details"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/password/change [post]
func (h *PasswordHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, &dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User ID not found in context",
		})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, &dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Invalid user ID format",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:   "User not found",
			Message: "User does not exist",
		})
		return
	}

	// Verify current password
	valid, err := h.passwordService.VerifyPassword(req.CurrentPassword, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:   "Internal server error",
			Message: "Failed to verify current password",
		})
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:   "Invalid current password",
			Message: "The current password is incorrect",
		})
		return
	}

	// Validate and hash new password
	newPasswordHash, validationResult, err := h.passwordService.ValidateAndHash(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:   "Internal server error",
			Message: "Failed to process new password",
		})
		return
	}

	if !validationResult.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid new password",
			"message": "New password does not meet security requirements",
			"validation_errors": validationResult.Errors,
		})
		return
	}

	// Update password in database
	user.Password = newPasswordHash
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:   "Internal server error",
			Message: "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, &dto.SuccessResponse{
		Data: "Password changed successfully",
	})
}

// GetPasswordPolicy returns the current password policy
// @Summary Get password policy
// @Description Returns the current password security policy
// @Tags password
// @Produce json
// @Success 200 {object} PasswordPolicyResponse
// @Router /api/v1/password/policy [get]
func (h *PasswordHandler) GetPasswordPolicy(c *gin.Context) {
	policy := h.passwordService.GetValidator().GetPolicy()

	response := PasswordPolicyResponse{
		MinLength:           policy.MinLength,
		RequireUppercase:    policy.RequireUppercase,
		RequireLowercase:    policy.RequireLowercase,
		RequireNumbers:      policy.RequireNumbers,
		RequireSpecialChars: policy.RequireSpecialChars,
		ForbidCommonPasswords: policy.ForbidCommonPasswords,
		MaxRepeatingChars:   policy.MaxRepeatingChars,
	}

	c.JSON(http.StatusOK, response)
}

// CheckPasswordStrength checks password strength for a specific user (admin only)
// @Summary Check password strength for user
// @Description Checks the password strength for a specific user (admin access required)
// @Tags password
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Param request body ValidatePasswordRequest true "Password to check"
// @Success 200 {object} ValidatePasswordResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/admin/users/{user_id}/password/check [post]
func (h *PasswordHandler) CheckPasswordStrength(c *gin.Context) {
	// Get user ID from URL parameter
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:   "Invalid user ID",
			Message: "User ID must be a valid number",
		})
		return
	}

	// Check if the user exists
	_, err = h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:   "User not found",
			Message: "The specified user does not exist",
		})
		return
	}

	var req ValidatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		})
		return
	}

	// Validate password
	result := h.passwordService.ValidatePassword(req.Password)

	response := ValidatePasswordResponse{
		IsValid:         result.IsValid,
		Errors:          result.Errors,
		Score:           result.Score,
		StrengthMessage: password.GenerateStrengthMessage(result.Score),
	}

	c.JSON(http.StatusOK, response)
}

// RegisterPasswordRoutes registers password-related routes
func (h *PasswordHandler) RegisterPasswordRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	// Public routes
	passwordGroup := r.Group("/password")
	{
		passwordGroup.POST("/validate", h.ValidatePassword)
		passwordGroup.GET("/policy", h.GetPasswordPolicy)
	}

	// Protected routes (require authentication)
	protectedGroup := r.Group("/password", authMiddleware)
	{
		protectedGroup.POST("/change", h.ChangePassword)
	}

	// Admin routes (require admin privileges)
	adminGroup := r.Group("/admin", authMiddleware, adminMiddleware)
	{
		adminGroup.POST("/users/:user_id/password/check", h.CheckPasswordStrength)
	}
}