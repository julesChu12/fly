package utils

import (
	"context"
	"errors"
)

// Context keys for user information
const (
	UserIDKey   = "user_id"
	UsernameKey = "username"
	TenantIDKey = "tenant_id"
	UserTypeKey = "user_type"
	EmailKey    = "email"
)

var (
	// ErrUserIDNotFound is returned when user ID is not found in context
	ErrUserIDNotFound = errors.New("user_id not found in context")
	// ErrTenantIDNotFound is returned when tenant ID is not found in context
	ErrTenantIDNotFound = errors.New("tenant_id not found in context")
)

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (uint, error) {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return 0, ErrUserIDNotFound
	}

	userID, ok := value.(uint)
	if !ok {
		return 0, ErrUserIDNotFound
	}

	return userID, nil
}

// MustGetUserIDFromContext extracts user ID from context and panics if not found
func MustGetUserIDFromContext(ctx context.Context) uint {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		panic(err)
	}
	return userID
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (uint, error) {
	value := ctx.Value(TenantIDKey)
	if value == nil {
		return 0, nil // Tenant ID is optional, return 0 without error
	}

	tenantID, ok := value.(uint)
	if !ok {
		return 0, nil
	}

	return tenantID, nil
}

// GetUsernameFromContext extracts username from context
func GetUsernameFromContext(ctx context.Context) string {
	value := ctx.Value(UsernameKey)
	if value == nil {
		return ""
	}

	username, ok := value.(string)
	if !ok {
		return ""
	}

	return username
}

// GetUserTypeFromContext extracts user type from context
func GetUserTypeFromContext(ctx context.Context) string {
	value := ctx.Value(UserTypeKey)
	if value == nil {
		return ""
	}

	userType, ok := value.(string)
	if !ok {
		return ""
	}

	return userType
}

// GetEmailFromContext extracts email from context
func GetEmailFromContext(ctx context.Context) string {
	value := ctx.Value(EmailKey)
	if value == nil {
		return ""
	}

	email, ok := value.(string)
	if !ok {
		return ""
	}

	return email
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithTenantID adds tenant ID to context
func WithTenantID(ctx context.Context, tenantID uint) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// WithUsername adds username to context
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, UsernameKey, username)
}

// WithUserType adds user type to context
func WithUserType(ctx context.Context, userType string) context.Context {
	return context.WithValue(ctx, UserTypeKey, userType)
}

// WithEmail adds email to context
func WithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, EmailKey, email)
}
