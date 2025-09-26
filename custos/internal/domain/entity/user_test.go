package entity

import (
	"testing"

	"github.com/julesChu12/custos/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	username := "testuser"
	email := "test@example.com"
	password := "hashedpassword123"

	user := NewUser(username, email, password)

	assert.Equal(t, username, user.Username)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, password, user.Password)
	assert.Equal(t, types.UserStatusActive, user.Status)
	assert.Equal(t, types.UserRoleUser, user.Role)
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   types.UserStatus
		expected bool
	}{
		{
			name:     "active user",
			status:   types.UserStatusActive,
			expected: true,
		},
		{
			name:     "inactive user",
			status:   types.UserStatusInactive,
			expected: false,
		},
		{
			name:     "frozen user",
			status:   types.UserStatusFrozen,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Status: tt.status}
			assert.Equal(t, tt.expected, user.IsActive())
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     types.UserRole
		expected bool
	}{
		{
			name:     "admin user",
			role:     types.UserRoleAdmin,
			expected: true,
		},
		{
			name:     "regular user",
			role:     types.UserRoleUser,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			assert.Equal(t, tt.expected, user.IsAdmin())
		})
	}
}

func TestUser_TableName(t *testing.T) {
	user := &User{}
	assert.Equal(t, "users", user.TableName())
}
