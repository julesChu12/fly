package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/service/rbac"
)

// RBACMiddleware creates middleware for role-based access control
func RBACMiddleware(rbacService *rbac.RBACService, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (should be set by auth middleware)
		userVal, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		user, ok := userVal.(*entity.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
			c.Abort()
			return
		}

		// Check permission
		if !rbacService.CheckPermission(c.Request.Context(), user, resource, action) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RBACResourceMiddleware creates middleware for resource-specific access control
func RBACResourceMiddleware(rbacService *rbac.RBACService, resourceType, action string, resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userVal, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		user, ok := userVal.(*entity.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
			c.Abort()
			return
		}

		// Get resource ID from URL parameter
		resourceID := c.Param(resourceIDParam)
		if resourceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resource ID is required"})
			c.Abort()
			return
		}

		// Check resource access
		if !rbacService.CheckResourceAccess(c.Request.Context(), user, resourceType, resourceID, action) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this resource"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole creates middleware that requires a specific role
func RequireRole(rbacService *rbac.RBACService, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userVal, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		user, ok := userVal.(*entity.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
			c.Abort()
			return
		}

		// Get user roles
		roles, err := rbacService.GetUserRoles(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user roles"})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, role := range roles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "required role not found"})
			c.Abort()
			return
		}

		c.Next()
	}
}