package repository

import (
	"context"
	"time"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
)

// SessionRepository 定义了登录会话及其刷新令牌的持久化操作。
type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	GetByID(ctx context.Context, id string) (*entity.Session, error)
	GetByRefreshTokenHash(ctx context.Context, hash string) (*entity.Session, error)
	UpdateRefreshToken(ctx context.Context, id string, newHash string, expiresAt time.Time, lastUsed time.Time) error
	UpdateLastSeen(ctx context.Context, sessionID string, lastSeenAt time.Time) error
	Revoke(ctx context.Context, id string, revokedAt time.Time) error
	RevokeByUser(ctx context.Context, userID uint, revokedAt time.Time) error
	ListActiveByUser(ctx context.Context, userID uint, now time.Time) ([]*entity.Session, error)
	CleanupExpired(ctx context.Context, olderThan time.Time) error
}
