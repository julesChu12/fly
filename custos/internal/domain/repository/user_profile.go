package repository

import (
	"context"
	"errors"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
)

// ErrUserProfileNotFound is returned when a user profile is not found
var ErrUserProfileNotFound = errors.New("user profile not found")

// UserProfileRepository defines the interface for user profile persistence
type UserProfileRepository interface {
	// Create creates a new user profile
	Create(ctx context.Context, profile *entity.UserProfile) error

	// GetByUserID retrieves a user profile by user ID
	GetByUserID(ctx context.Context, userID uint) (*entity.UserProfile, error)

	// Update updates an existing user profile
	Update(ctx context.Context, profile *entity.UserProfile) error

	// Delete deletes a user profile by user ID
	Delete(ctx context.Context, userID uint) error

	// Exists checks if a profile exists for a user
	Exists(ctx context.Context, userID uint) (bool, error)
}
