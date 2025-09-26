package rbac

import (
	"context"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/pkg/types"
)

// RBACService handles role-based access control
type RBACService struct {
	// Future: Casbin enforcer will be added here
}

// NewRBACService creates a new RBAC service
func NewRBACService() *RBACService {
	return &RBACService{}
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *RBACService) CheckPermission(ctx context.Context, user *entity.User, resource, action string) bool {
	// Basic role-based permission check
	switch user.Role {
	case types.UserRoleAdmin:
		return true // Admin has all permissions
	case types.UserRoleUser:
		// Regular users can only access their own resources
		return action == "read" || action == "update"
	case types.UserRoleGuest:
		// Guests have very limited permissions
		return action == "read"
	default:
		return false
	}
}

// CheckResourceAccess checks if a user can access a specific resource
func (s *RBACService) CheckResourceAccess(ctx context.Context, user *entity.User, resourceID string) bool {
	// For now, users can only access their own resources
	// This will be enhanced with Casbin policies
	return true // Placeholder implementation
}

// GetUserPermissions returns all permissions for a user
func (s *RBACService) GetUserPermissions(ctx context.Context, user *entity.User) []string {
	permissions := []string{}

	switch user.Role {
	case types.UserRoleAdmin:
		permissions = []string{"*"} // All permissions
	case types.UserRoleUser:
		permissions = []string{"read", "update", "create"}
	case types.UserRoleGuest:
		permissions = []string{"read"}
	}

	return permissions
}
