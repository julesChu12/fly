package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/domain/service/rbac"
	"github.com/julesChu12/fly/custos/pkg/types"
)

type AdminHandler struct {
	userRepo    repository.UserRepository
	rbacService *rbac.RBACService
}

func NewAdminHandler(userRepo repository.UserRepository, rbacService *rbac.RBACService) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		rbacService: rbacService,
	}
}

// ListUsers lists all users (admin only)
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	// Get current user from context (set by auth middleware)
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user := currentUser.(*entity.User)

	// Check admin permission
	if !h.rbacService.CheckPermission(c.Request.Context(), user, "users", "read") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// Get pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100 // Max 100 users per page
	}

	users, err := h.userRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	// Convert to response DTOs
	var userResponses []dto.UserInfo
	for _, u := range users {
		userResponses = append(userResponses, dto.UserInfo{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Nickname: u.Nickname,
			Avatar:   u.Avatar,
			Role:     string(u.Role),
			Status:   string(u.Status),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": userResponses,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUser gets a specific user by ID (admin only)
// GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUser(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user := currentUser.(*entity.User)

	if !h.rbacService.CheckPermission(c.Request.Context(), user, "users", "read") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	targetUser, err := h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		}
		return
	}

	response := dto.UserInfo{
		ID:       targetUser.ID,
		Username: targetUser.Username,
		Email:    targetUser.Email,
		Nickname: targetUser.Nickname,
		Avatar:   targetUser.Avatar,
		Role:     string(targetUser.Role),
		Status:   string(targetUser.Status),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateUserStatus updates user status (admin only)
// PATCH /api/v1/admin/users/:id/status
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user := currentUser.(*entity.User)

	if !h.rbacService.CheckPermission(c.Request.Context(), user, "users", "update") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive frozen disabled locked"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetUser, err := h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		}
		return
	}

	// Update status
	targetUser.Status = types.UserStatus(req.Status)

	if err := h.userRepo.Update(c.Request.Context(), targetUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user status updated successfully",
		"data": gin.H{
			"id":     targetUser.ID,
			"status": string(targetUser.Status),
		},
	})
}

// UpdateUserRole updates user role (admin only)
// PATCH /api/v1/admin/users/:id/role
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user := currentUser.(*entity.User)

	if !h.rbacService.CheckPermission(c.Request.Context(), user, "roles", "update") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=admin user guest"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetUser, err := h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		}
		return
	}

	// Update role
	targetUser.Role = types.UserRole(req.Role)

	if err := h.userRepo.Update(c.Request.Context(), targetUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user role updated successfully",
		"data": gin.H{
			"id":   targetUser.ID,
			"role": string(targetUser.Role),
		},
	})
}

// ForceLogoutUser forces logout for a specific user (admin only)
// POST /api/v1/admin/users/:id/force-logout
func (h *AdminHandler) ForceLogoutUser(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user := currentUser.(*entity.User)

	if !h.rbacService.CheckPermission(c.Request.Context(), user, "users", "update") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	targetUser, err := h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		}
		return
	}

	// Increment token version to invalidate all tokens
	targetUser.IncrementTokenVersion()

	if err := h.userRepo.Update(c.Request.Context(), targetUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to force logout user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user forced logout successfully",
		"data": gin.H{
			"user_id":       targetUser.ID,
			"token_version": targetUser.TokenVersion,
		},
	})
}

// GetSystemStats gets system statistics (admin only)
// GET /api/v1/admin/stats
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user := currentUser.(*entity.User)

	if !h.rbacService.CheckPermission(c.Request.Context(), user, "admin", "access") {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// Get basic stats (simplified implementation)
	users, err := h.userRepo.List(c.Request.Context(), 1000, 0) // Get sample to count
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	stats := map[string]interface{}{
		"total_users": len(users),
		"active_users": func() int {
			count := 0
			for _, u := range users {
				if u.Status == types.UserStatusActive {
					count++
				}
			}
			return count
		}(),
		"admin_users": func() int {
			count := 0
			for _, u := range users {
				if u.Role == types.UserRoleAdmin {
					count++
				}
			}
			return count
		}(),
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}