package mysql

import (
	"context"
	"time"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"gorm.io/gorm"
)

type sessionRepository struct {
	db               *gorm.DB
	refreshTokenRepo repository.RefreshTokenRepository
}

func NewSessionRepository(db *gorm.DB) repository.SessionRepository {
	refreshTokenRepo := NewRefreshTokenRepository(db)
	return &sessionRepository{
		db:               db,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (r *sessionRepository) Create(ctx context.Context, session *entity.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	var session entity.Session
	if err := r.db.WithContext(ctx).
		Where("session_id = ? AND revoked = false", id).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetByRefreshTokenHash(ctx context.Context, hash string) (*entity.Session, error) {
	// First find the refresh token by hash
	refreshToken, err := r.refreshTokenRepo.GetByTokenHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if refreshToken == nil {
		return nil, nil // Refresh token not found
	}

	// Find the session associated with this refresh token
	var session entity.Session
	if err := r.db.WithContext(ctx).
		Where("refresh_token_id = ? AND revoked = false", refreshToken.ID).
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) UpdateRefreshToken(ctx context.Context, id, newHash string, expiresAt time.Time, lastUsed time.Time) error {
	// Start a transaction to ensure consistency
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the session first
	var session entity.Session
	if err := tx.Where("session_id = ?", id).First(&session).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound
		}
		return err
	}

	// If session has an existing refresh token, mark it as used
	if session.RefreshTokenID != nil {
		if err := tx.Model(&entity.RefreshToken{}).
			Where("id = ?", *session.RefreshTokenID).
			Update("is_used", true).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Create a new refresh token
	newRefreshToken := &entity.RefreshToken{
		UserID:    session.UserID,
		TokenHash: newHash,
		ExpiresAt: expiresAt,
	}
	if err := tx.Create(newRefreshToken).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update the session with the new refresh token ID and last seen time
	if err := tx.Model(&session).
		Updates(map[string]interface{}{
			"refresh_token_id": newRefreshToken.ID,
			"last_seen_at":     lastUsed,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}

func (r *sessionRepository) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&entity.Session{}).
		Where("session_id = ?", id).
		Update("revoked", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *sessionRepository) RevokeByUser(ctx context.Context, userID uint, revokedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.Session{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}

func (r *sessionRepository) ListActiveByUser(ctx context.Context, userID uint, now time.Time) ([]*entity.Session, error) {
	var sessions []*entity.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked = false", userID).
		Order("last_seen_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *sessionRepository) UpdateLastSeen(ctx context.Context, sessionID string, lastSeenAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&entity.Session{}).
		Where("session_id = ?", sessionID).
		Update("last_seen_at", lastSeenAt)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *sessionRepository) CleanupExpired(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("revoked = true AND created_at < ?", olderThan).
		Delete(&entity.Session{}).Error
}
