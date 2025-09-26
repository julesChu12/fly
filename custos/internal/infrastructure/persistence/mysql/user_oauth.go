package mysql

import (
	"context"
	"fmt"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"gorm.io/gorm"
)

type userOAuthRepository struct {
	db *gorm.DB
}

func NewUserOAuthRepository(db *gorm.DB) repository.UserOAuthRepository {
	return &userOAuthRepository{db: db}
}

func (r *userOAuthRepository) Create(ctx context.Context, userOAuth *entity.UserOAuth) error {
	if err := r.db.WithContext(ctx).Create(userOAuth).Error; err != nil {
		return fmt.Errorf("failed to create user OAuth binding: %w", err)
	}
	return nil
}

func (r *userOAuthRepository) GetByProviderUID(ctx context.Context, provider, providerUID string) (*entity.UserOAuth, error) {
	var userOAuth entity.UserOAuth
	if err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_uid = ?", provider, providerUID).
		First(&userOAuth).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user OAuth binding: %w", err)
	}
	return &userOAuth, nil
}

func (r *userOAuthRepository) GetByUserID(ctx context.Context, userID uint) ([]*entity.UserOAuth, error) {
	var bindings []*entity.UserOAuth
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("failed to get user OAuth bindings: %w", err)
	}
	return bindings, nil
}

func (r *userOAuthRepository) GetByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*entity.UserOAuth, error) {
	var userOAuth entity.UserOAuth
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		First(&userOAuth).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user OAuth binding: %w", err)
	}
	return &userOAuth, nil
}

func (r *userOAuthRepository) Update(ctx context.Context, userOAuth *entity.UserOAuth) error {
	if err := r.db.WithContext(ctx).Save(userOAuth).Error; err != nil {
		return fmt.Errorf("failed to update user OAuth binding: %w", err)
	}
	return nil
}

func (r *userOAuthRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entity.UserOAuth{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user OAuth binding: %w", err)
	}
	return nil
}

func (r *userOAuthRepository) UnbindProvider(ctx context.Context, userID uint, provider string) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		Delete(&entity.UserOAuth{}).Error; err != nil {
		return fmt.Errorf("failed to unbind OAuth provider: %w", err)
	}
	return nil
}
