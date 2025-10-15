package mysql

import (
	"context"
	"fmt"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	moradb "github.com/julesChu12/fly/mora/pkg/db"
)

type userOAuthRepository struct {
	client *moradb.Client
}

func NewUserOAuthRepository(client *moradb.Client) repository.UserOAuthRepository {
	return &userOAuthRepository{client: client}
}

func (r *userOAuthRepository) Create(ctx context.Context, userOAuth *entity.UserOAuth) error {
	if err := r.client.Create(ctx, userOAuth); err != nil {
		return fmt.Errorf("failed to create user OAuth binding: %w", err)
	}
	return nil
}

func (r *userOAuthRepository) GetByProviderUID(ctx context.Context, provider, providerUID string) (*entity.UserOAuth, error) {
	var userOAuth entity.UserOAuth
	if err := r.client.DB().WithContext(ctx).
		Where("provider = ? AND provider_uid = ?", provider, providerUID).
		First(&userOAuth).Error; err != nil {
		return nil, nil // Return nil for not found (gorm.ErrRecordNotFound)
	}
	return &userOAuth, nil
}

func (r *userOAuthRepository) GetByUserID(ctx context.Context, userID uint) ([]*entity.UserOAuth, error) {
	var bindings []*entity.UserOAuth
	if err := r.client.DB().WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("failed to get user OAuth bindings: %w", err)
	}
	return bindings, nil
}

func (r *userOAuthRepository) GetByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*entity.UserOAuth, error) {
	var userOAuth entity.UserOAuth
	if err := r.client.DB().WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		First(&userOAuth).Error; err != nil {
		return nil, nil // Return nil for not found
	}
	return &userOAuth, nil
}

func (r *userOAuthRepository) Update(ctx context.Context, userOAuth *entity.UserOAuth) error {
	if err := r.client.Save(ctx, userOAuth); err != nil {
		return fmt.Errorf("failed to update user OAuth binding: %w", err)
	}
	return nil
}

func (r *userOAuthRepository) Delete(ctx context.Context, id uint) error {
	if err := r.client.Delete(ctx, &entity.UserOAuth{}, id); err != nil {
		return fmt.Errorf("failed to delete user OAuth binding: %w", err)
	}
	return nil
}

func (r *userOAuthRepository) UnbindProvider(ctx context.Context, userID uint, provider string) error {
	if err := r.client.DB().WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		Delete(&entity.UserOAuth{}).Error; err != nil {
		return fmt.Errorf("failed to unbind OAuth provider: %w", err)
	}
	return nil
}
