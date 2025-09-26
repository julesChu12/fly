package repository

import (
	"context"
	"time"

	"github.com/julesChu12/custos/internal/domain/entity"
)

// SessionRepository defines persistence operations for login sessions and their refresh tokens.
type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	GetByID(ctx context.Context, id string) (*entity.Session, error)
	GetByRefreshTokenHash(ctx context.Context, hash string) (*entity.Session, error)
	UpdateRefreshToken(ctx context.Context, id, newHash string, expiresAt time.Time, lastUsed time.Time) error
	Revoke(ctx context.Context, id string, revokedAt time.Time) error
	RevokeByUser(ctx context.Context, userID uint, revokedAt time.Time) error
	ListActiveByUser(ctx context.Context, userID uint, now time.Time) ([]*entity.Session, error)
	CleanupExpired(ctx context.Context, olderThan time.Time) error
}
