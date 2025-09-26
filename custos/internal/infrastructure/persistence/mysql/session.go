package mysql

import (
	"context"
	"time"

	"github.com/julesChu12/custos/internal/domain/entity"
	"github.com/julesChu12/custos/internal/domain/repository"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) repository.SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *entity.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	var session entity.Session
	if err := r.db.WithContext(ctx).First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) GetByRefreshTokenHash(ctx context.Context, hash string) (*entity.Session, error) {
	var session entity.Session
	if err := r.db.WithContext(ctx).Where("refresh_token_hash = ?", hash).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) UpdateRefreshToken(ctx context.Context, id, newHash string, expiresAt time.Time, lastUsed time.Time) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.Session{}).Where("id = ?", id).Updates(map[string]interface{}{
		"refresh_token_hash":       newHash,
		"refresh_token_expires_at": expiresAt,
		"last_used_at":             lastUsed,
		"revoked_at":               nil,
		"updated_at":               now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.Session{}).Where("id = ?", id).Updates(map[string]interface{}{
		"revoked_at":   revokedAt,
		"updated_at":   now,
		"last_used_at": revokedAt,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *SessionRepository) RevokeByUser(ctx context.Context, userID uint, revokedAt time.Time) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.Session{}).Where("user_id = ? AND revoked_at IS NULL", userID).Updates(map[string]interface{}{
		"revoked_at":   revokedAt,
		"updated_at":   now,
		"last_used_at": revokedAt,
	})
	return result.Error
}

func (r *SessionRepository) ListActiveByUser(ctx context.Context, userID uint, now time.Time) ([]*entity.Session, error) {
	var sessions []*entity.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND (revoked_at IS NULL) AND refresh_token_expires_at > ?", userID, now).
		Order("last_used_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *SessionRepository) CleanupExpired(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("refresh_token_expires_at < ?", olderThan).
		Delete(&entity.Session{}).Error
}
