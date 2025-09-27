package rbac

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/pkg/types"
)

// RBACService handles role-based access control using Casbin
type RBACService struct {
	enforcer *casbin.Enforcer
}

// NewRBACService creates a new RBAC service with Casbin
func NewRBACService(db *gorm.DB, modelPath string) (*RBACService, error) {
	// Initialize Gorm adapter for Casbin
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create gorm adapter: %w", err)
	}

	// Create Casbin enforcer
	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// Load policy from database
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	service := &RBACService{
		enforcer: enforcer,
	}

	// Initialize default policies
	if err := service.initializeDefaultPolicies(); err != nil {
		return nil, fmt.Errorf("failed to initialize default policies: %w", err)
	}

	return service, nil
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *RBACService) CheckPermission(ctx context.Context, user *entity.User, resource, action string) bool {
	userSubject := fmt.Sprintf("user:%d", user.ID)

	// Check direct permission
	allowed, err := s.enforcer.Enforce(userSubject, resource, action)
	if err != nil {
		return false
	}

	return allowed
}

// CheckResourceAccess checks if a user can access a specific resource
func (s *RBACService) CheckResourceAccess(ctx context.Context, user *entity.User, resourceType, resourceID, action string) bool {
	userSubject := fmt.Sprintf("user:%d", user.ID)
	resource := fmt.Sprintf("%s:%s", resourceType, resourceID)

	allowed, err := s.enforcer.Enforce(userSubject, resource, action)
	if err != nil {
		return false
	}

	return allowed
}

// AssignRole assigns a role to a user
func (s *RBACService) AssignRole(ctx context.Context, userID uint, role string) error {
	userSubject := fmt.Sprintf("user:%d", userID)

	// Remove existing roles first
	if err := s.RemoveAllRoles(ctx, userID); err != nil {
		return err
	}

	// Assign new role
	_, err := s.enforcer.AddRoleForUser(userSubject, role)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// RemoveRole removes a role from a user
func (s *RBACService) RemoveRole(ctx context.Context, userID uint, role string) error {
	userSubject := fmt.Sprintf("user:%d", userID)

	_, err := s.enforcer.DeleteRoleForUser(userSubject, role)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// RemoveAllRoles removes all roles from a user
func (s *RBACService) RemoveAllRoles(ctx context.Context, userID uint) error {
	userSubject := fmt.Sprintf("user:%d", userID)

	_, err := s.enforcer.DeleteRolesForUser(userSubject)
	if err != nil {
		return fmt.Errorf("failed to remove all roles: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// GetUserRoles returns all roles for a user
func (s *RBACService) GetUserRoles(ctx context.Context, userID uint) ([]string, error) {
	userSubject := fmt.Sprintf("user:%d", userID)
	roles, err := s.enforcer.GetRolesForUser(userSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	return roles, nil
}

// GetUserPermissions returns all permissions for a user
func (s *RBACService) GetUserPermissions(ctx context.Context, user *entity.User) []string {
	userSubject := fmt.Sprintf("user:%d", user.ID)
	permissions, err := s.enforcer.GetPermissionsForUser(userSubject)
	if err != nil {
		return []string{}
	}

	var result []string
	for _, perm := range permissions {
		if len(perm) >= 3 {
			result = append(result, fmt.Sprintf("%s:%s", perm[1], perm[2]))
		}
	}

	return result
}

// AddPolicy adds a policy rule
func (s *RBACService) AddPolicy(ctx context.Context, subject, object, action string) error {
	_, err := s.enforcer.AddPolicy(subject, object, action)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// RemovePolicy removes a policy rule
func (s *RBACService) RemovePolicy(ctx context.Context, subject, object, action string) error {
	_, err := s.enforcer.RemovePolicy(subject, object, action)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// initializeDefaultPolicies sets up default roles and policies
func (s *RBACService) initializeDefaultPolicies() error {
	// Define default role policies
	defaultPolicies := [][]string{
		// Admin role - full access
		{"admin", "*", "*"},

		// User role - limited access
		{"user", "profile", "read"},
		{"user", "profile", "update"},
		{"user", "sessions", "read"},
		{"user", "sessions", "delete"},

		// Guest role - read only
		{"guest", "profile", "read"},
	}

	// Add policies if they don't exist
	for _, policy := range defaultPolicies {
		if len(policy) == 3 {
			exists, err := s.enforcer.HasPolicy(policy[0], policy[1], policy[2])
			if err != nil {
				return fmt.Errorf("failed to check policy existence %v: %w", policy, err)
			}
			if !exists {
				if _, err := s.enforcer.AddPolicy(policy[0], policy[1], policy[2]); err != nil {
					return fmt.Errorf("failed to add default policy %v: %w", policy, err)
				}
			}
		}
	}

	// Save policies to database
	return s.enforcer.SavePolicy()
}

// SyncUserRole synchronizes user role with Casbin based on entity role
func (s *RBACService) SyncUserRole(ctx context.Context, user *entity.User) error {
	var casbinRole string

	switch user.Role {
	case types.UserRoleAdmin:
		casbinRole = "admin"
	case types.UserRoleUser:
		casbinRole = "user"
	case types.UserRoleGuest:
		casbinRole = "guest"
	default:
		casbinRole = "guest"
	}

	return s.AssignRole(ctx, user.ID, casbinRole)
}
