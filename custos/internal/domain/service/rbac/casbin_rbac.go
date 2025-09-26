package rbac

import (
	"context"
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"gorm.io/gorm"
)

// CasbinRBACService implements RBAC using Casbin
type CasbinRBACService struct {
	enforcer *casbin.Enforcer
	adapter  *gormadapter.Adapter
}

// NewCasbinRBACService creates a new Casbin-based RBAC service
func NewCasbinRBACService(db *gorm.DB, modelPath string) (*CasbinRBACService, error) {
	// Initialize Casbin adapter
	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "casbin", "casbin_rule")
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	// Initialize Casbin enforcer
	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Load policy from database
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load Casbin policy: %w", err)
	}

	return &CasbinRBACService{
		enforcer: enforcer,
		adapter:  adapter,
	}, nil
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *CasbinRBACService) CheckPermission(ctx context.Context, user *entity.User, resource, action string) bool {
	userID := strconv.FormatUint(uint64(user.ID), 10)

	// For multi-tenancy, include tenant in the subject
	subject := userID
	if user.TenantID != nil {
		subject = fmt.Sprintf("%s:tenant:%d", userID, *user.TenantID)
	}

	allowed, err := s.enforcer.Enforce(subject, resource, action)
	if err != nil {
		// Log error and deny by default
		return false
	}
	return allowed
}

// CheckResourceAccess checks if a user can access a specific resource instance
func (s *CasbinRBACService) CheckResourceAccess(ctx context.Context, user *entity.User, resourceID string) bool {
	userID := strconv.FormatUint(uint64(user.ID), 10)

	subject := userID
	if user.TenantID != nil {
		subject = fmt.Sprintf("%s:tenant:%d", userID, *user.TenantID)
	}

	// Check if user owns the resource or has admin access
	allowed, err := s.enforcer.Enforce(subject, resourceID, "read")
	if err != nil {
		return false
	}
	return allowed
}

// AssignRole assigns a role to a user
func (s *CasbinRBACService) AssignRole(ctx context.Context, userID uint, role string, tenantID *uint) error {
	subject := strconv.FormatUint(uint64(userID), 10)
	if tenantID != nil {
		subject = fmt.Sprintf("%s:tenant:%d", subject, *tenantID)
		role = fmt.Sprintf("%s:tenant:%d", role, *tenantID)
	}

	_, err := s.enforcer.AddRoleForUser(subject, role)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// RemoveRole removes a role from a user
func (s *CasbinRBACService) RemoveRole(ctx context.Context, userID uint, role string, tenantID *uint) error {
	subject := strconv.FormatUint(uint64(userID), 10)
	if tenantID != nil {
		subject = fmt.Sprintf("%s:tenant:%d", subject, *tenantID)
		role = fmt.Sprintf("%s:tenant:%d", role, *tenantID)
	}

	_, err := s.enforcer.DeleteRoleForUser(subject, role)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// GetUserRoles gets all roles for a user
func (s *CasbinRBACService) GetUserRoles(ctx context.Context, userID uint, tenantID *uint) ([]string, error) {
	subject := strconv.FormatUint(uint64(userID), 10)
	if tenantID != nil {
		subject = fmt.Sprintf("%s:tenant:%d", subject, *tenantID)
	}

	roles, err := s.enforcer.GetRolesForUser(subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}

// GetUserPermissions gets all permissions for a user (derived from roles)
func (s *CasbinRBACService) GetUserPermissions(ctx context.Context, user *entity.User) ([]string, error) {
	userID := strconv.FormatUint(uint64(user.ID), 10)

	subject := userID
	if user.TenantID != nil {
		subject = fmt.Sprintf("%s:tenant:%d", userID, *user.TenantID)
	}

	permissions, err := s.enforcer.GetPermissionsForUser(subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	var result []string
	for _, perm := range permissions {
		if len(perm) >= 3 {
			result = append(result, fmt.Sprintf("%s:%s", perm[1], perm[2])) // resource:action
		}
	}

	return result, nil
}

// AddPolicy adds a policy rule
func (s *CasbinRBACService) AddPolicy(ctx context.Context, role, resource, action string, tenantID *uint) error {
	if tenantID != nil {
		role = fmt.Sprintf("%s:tenant:%d", role, *tenantID)
	}

	_, err := s.enforcer.AddPolicy(role, resource, action)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// RemovePolicy removes a policy rule
func (s *CasbinRBACService) RemovePolicy(ctx context.Context, role, resource, action string, tenantID *uint) error {
	if tenantID != nil {
		role = fmt.Sprintf("%s:tenant:%d", role, *tenantID)
	}

	_, err := s.enforcer.RemovePolicy(role, resource, action)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// InitializeDefaultPolicies sets up default roles and policies
func (s *CasbinRBACService) InitializeDefaultPolicies(ctx context.Context, tenantID *uint) error {
	policies := [][]string{
		// Admin role has all permissions
		{"admin", "users", "create"},
		{"admin", "users", "read"},
		{"admin", "users", "update"},
		{"admin", "users", "delete"},
		{"admin", "roles", "create"},
		{"admin", "roles", "read"},
		{"admin", "roles", "update"},
		{"admin", "roles", "delete"},
		{"admin", "admin", "access"},

		// Staff role has limited permissions
		{"staff", "users", "read"},
		{"staff", "users", "update"},

		// Customer role has basic permissions
		{"customer", "profile", "read"},
		{"customer", "profile", "update"},
	}

	for _, policy := range policies {
		role, resource, action := policy[0], policy[1], policy[2]
		if tenantID != nil {
			role = fmt.Sprintf("%s:tenant:%d", role, *tenantID)
		}

		if _, err := s.enforcer.AddPolicy(role, resource, action); err != nil {
			return fmt.Errorf("failed to add default policy %v: %w", policy, err)
		}
	}

	return s.enforcer.SavePolicy()
}

// SyncUserRole synchronizes user role in Casbin with user entity role
func (s *CasbinRBACService) SyncUserRole(ctx context.Context, user *entity.User) error {
	userID := strconv.FormatUint(uint64(user.ID), 10)

	subject := userID
	role := string(user.Role)
	if user.TenantID != nil {
		subject = fmt.Sprintf("%s:tenant:%d", userID, *user.TenantID)
		role = fmt.Sprintf("%s:tenant:%d", role, *user.TenantID)
	}

	// Remove all existing roles for the user
	if _, err := s.enforcer.DeleteRolesForUser(subject); err != nil {
		return fmt.Errorf("failed to clear user roles: %w", err)
	}

	// Add the current role
	if _, err := s.enforcer.AddRoleForUser(subject, role); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return s.enforcer.SavePolicy()
}

// ReloadPolicy reloads policy from database
func (s *CasbinRBACService) ReloadPolicy() error {
	return s.enforcer.LoadPolicy()
}
