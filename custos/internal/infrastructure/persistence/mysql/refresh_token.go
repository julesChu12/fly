package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	if err := r.db.WithContext(ctx).
		Where("token_hash = ? AND is_used = false AND expires_at > ?", tokenHash, time.Now()).
		First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return &token, nil
}

func (r *refreshTokenRepository) GetByUserID(ctx context.Context, userID uint) ([]*entity.RefreshToken, error) {
	var tokens []*entity.RefreshToken
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to get refresh tokens: %w", err)
	}
	return tokens, nil
}

func (r *refreshTokenRepository) Update(ctx context.Context, token *entity.RefreshToken) error {
	if err := r.db.WithContext(ctx).Save(token).Error; err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entity.RefreshToken{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&entity.RefreshToken{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).
		Model(&entity.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("is_used", true).Error; err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}
	return nil
}
