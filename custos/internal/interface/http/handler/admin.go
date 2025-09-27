package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/domain/service/rbac"
)

type AdminHandler struct {
	userRepo repository.UserRepository
	rbacSvc  *rbac.RBACService
}

func NewAdminHandler(userRepo repository.UserRepository, rbacSvc *rbac.RBACService) *AdminHandler {
	return &AdminHandler{
		userRepo: userRepo,
		rbacSvc:  rbacSvc,
	}
}

// AssignRole assigns a role to a user
// POST /api/v1/admin/users/:id/roles
func (h *AdminHandler) AssignRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	validRoles := []string{"admin", "user", "guest"}
	isValidRole := false
	for _, role := range validRoles {
		if req.Role == role {
			isValidRole = true
			break
		}
	}

	if !isValidRole {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	// Check if user exists
	_, err = h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Assign role
	if err := h.rbacSvc.AssignRole(c.Request.Context(), uint(userID), req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role assigned successfully"})
}

// GetUserRoles gets all roles for a user
// GET /api/v1/admin/users/:id/roles
func (h *AdminHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Check if user exists
	user, err := h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get roles
	roles, err := h.rbacSvc.GetUserRoles(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user roles"})
		return
	}

	// Get permissions
	permissions := h.rbacSvc.GetUserPermissions(c.Request.Context(), user)

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"roles":       roles,
		"permissions": permissions,
	})
}

// AddPolicy adds a new policy rule
// POST /api/v1/admin/policies
func (h *AdminHandler) AddPolicy(c *gin.Context) {
	var req struct {
		Subject string `json:"subject" binding:"required"`
		Object  string `json:"object" binding:"required"`
		Action  string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add policy
	if err := h.rbacSvc.AddPolicy(c.Request.Context(), req.Subject, req.Object, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "policy added successfully"})
}

// RemovePolicy removes a policy rule
// DELETE /api/v1/admin/policies
func (h *AdminHandler) RemovePolicy(c *gin.Context) {
	var req struct {
		Subject string `json:"subject" binding:"required"`
		Object  string `json:"object" binding:"required"`
		Action  string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove policy
	if err := h.rbacSvc.RemovePolicy(c.Request.Context(), req.Subject, req.Object, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "policy removed successfully"})
}

// ListUsers placeholder (admin only)
func (h *AdminHandler) ListUsers(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "list users not implemented"})
}

// GetUser placeholder (admin only)
func (h *AdminHandler) GetUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "get user not implemented"})
}

// UpdateUserStatus placeholder (admin only)
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "update user status not implemented"})
}

// UpdateUserRole placeholder (admin only)
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "update user role not implemented"})
}

// ForceLogoutUser placeholder (admin only)
func (h *AdminHandler) ForceLogoutUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "force logout user not implemented"})
}

// GetSystemStats placeholder (admin only)
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "get system stats not implemented"})
}