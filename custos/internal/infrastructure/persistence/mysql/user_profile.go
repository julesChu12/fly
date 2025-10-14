package mysql

import (
	"context"
	"fmt"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"gorm.io/gorm"
)

type userProfileRepository struct {
	db *gorm.DB
}

// NewUserProfileRepository creates a new user profile repository
func NewUserProfileRepository(db *gorm.DB) repository.UserProfileRepository {
	return &userProfileRepository{db: db}
}

func (r *userProfileRepository) Create(ctx context.Context, profile *entity.UserProfile) error {
	if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
		return fmt.Errorf("failed to create user profile: %w", err)
	}
	return nil
}

func (r *userProfileRepository) GetByUserID(ctx context.Context, userID uint) (*entity.UserProfile, error) {
	var profile entity.UserProfile
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, repository.ErrUserProfileNotFound
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return &profile, nil
}

func (r *userProfileRepository) Update(ctx context.Context, profile *entity.UserProfile) error {
	if err := r.db.WithContext(ctx).
		Model(&entity.UserProfile{}).
		Where("user_id = ?", profile.UserID).
		Updates(profile).Error; err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}
	return nil
}

func (r *userProfileRepository) Delete(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.UserProfile{}).Error; err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}
	return nil
}

func (r *userProfileRepository) Exists(ctx context.Context, userID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.UserProfile{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check profile existence: %w", err)
	}
	return count > 0, nil
}
