package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"gorm.io/gorm"
)

type sessionRepositoryNew struct {
	db *gorm.DB
}

func NewSessionRepositoryNew(db *gorm.DB) repository.SessionRepository {
	return &sessionRepositoryNew{db: db}
}

func (r *sessionRepositoryNew) Create(ctx context.Context, session *entity.Session) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *sessionRepositoryNew) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	var session entity.Session
	if err := r.db.WithContext(ctx).
		Where("session_id = ? AND revoked = false", id).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

func (r *sessionRepositoryNew) GetByRefreshTokenHash(ctx context.Context, hash string) (*entity.Session, error) {
	// TODO: Implement refresh token hash lookup when RefreshToken entity is properly integrated
	return nil, fmt.Errorf("not implemented: GetByRefreshTokenHash")
}

func (r *sessionRepositoryNew) UpdateRefreshToken(ctx context.Context, id string, newHash string, expiresAt time.Time, lastUsed time.Time) error {
	// TODO: Implement refresh token update when RefreshToken entity is properly integrated
	return fmt.Errorf("not implemented: UpdateRefreshToken")
}

func (r *sessionRepositoryNew) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	if err := r.db.WithContext(ctx).
		Model(&entity.Session{}).
		Where("session_id = ?", id).
		Update("revoked", true).Error; err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

func (r *sessionRepositoryNew) RevokeByUser(ctx context.Context, userID uint, revokedAt time.Time) error {
	if err := r.db.WithContext(ctx).
		Model(&entity.Session{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error; err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}
	return nil
}

func (r *sessionRepositoryNew) ListActiveByUser(ctx context.Context, userID uint, now time.Time) ([]*entity.Session, error) {
	var sessions []*entity.Session
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked = false", userID).
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	return sessions, nil
}

func (r *sessionRepositoryNew) CleanupExpired(ctx context.Context, olderThan time.Time) error {
	if err := r.db.WithContext(ctx).
		Where("revoked = true AND created_at < ?", olderThan).
		Delete(&entity.Session{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}
