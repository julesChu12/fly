package user

import (
	"context"

	"github.com/julesChu12/custos/internal/domain/entity"
	"github.com/julesChu12/custos/internal/domain/repository"
	"github.com/julesChu12/custos/pkg/errors"
)

type UserUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(userRepo repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
	}
}

// GetUser retrieves a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, userID uint) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewUserNotFoundError()
	}
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (uc *UserUseCase) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errors.NewUserNotFoundError()
	}
	return user, nil
}

// UpdateUser updates user information
func (uc *UserUseCase) UpdateUser(ctx context.Context, userID uint, updates map[string]interface{}) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.NewUserNotFoundError()
	}

	// Update user fields based on the updates map
	for field, value := range updates {
		switch field {
		case "email":
			if email, ok := value.(string); ok {
				user.Email = email
			}
		case "username":
			if username, ok := value.(string); ok {
				user.Username = username
			}
		}
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeactivateUser deactivates a user account
func (uc *UserUseCase) DeactivateUser(ctx context.Context, userID uint) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.NewUserNotFoundError()
	}

	user.Deactivate()
	return uc.userRepo.Update(ctx, user)
}

// ActivateUser activates a user account
func (uc *UserUseCase) ActivateUser(ctx context.Context, userID uint) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.NewUserNotFoundError()
	}

	user.Activate()
	return uc.userRepo.Update(ctx, user)
}

// ListUsers retrieves a list of users with pagination
func (uc *UserUseCase) ListUsers(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	return uc.userRepo.List(ctx, limit, offset)
}
