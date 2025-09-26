package repository

import (
	"context"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
)

// UserOAuthRepository defines methods for OAuth user binding operations
type UserOAuthRepository interface {
	Create(ctx context.Context, userOAuth *entity.UserOAuth) error
	GetByProviderUID(ctx context.Context, provider, providerUID string) (*entity.UserOAuth, error)
	GetByUserID(ctx context.Context, userID uint) ([]*entity.UserOAuth, error)
	GetByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*entity.UserOAuth, error)
	Update(ctx context.Context, userOAuth *entity.UserOAuth) error
	Delete(ctx context.Context, id uint) error
	UnbindProvider(ctx context.Context, userID uint, provider string) error
}

// UserProfileRepository defines methods for user profile operations
type UserProfileRepository interface {
	Create(ctx context.Context, profile *entity.UserProfile) error
	GetByUserID(ctx context.Context, userID uint) (*entity.UserProfile, error)
	Update(ctx context.Context, profile *entity.UserProfile) error
	Delete(ctx context.Context, userID uint) error
}

// RefreshTokenRepository defines methods for refresh token operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	GetByUserID(ctx context.Context, userID uint) ([]*entity.RefreshToken, error)
	Update(ctx context.Context, token *entity.RefreshToken) error
	Delete(ctx context.Context, id uint) error
	DeleteExpired(ctx context.Context) (int64, error)
	RevokeByUserID(ctx context.Context, userID uint) error
}

// JWKKeyRepository defines methods for JWK key operations
type JWKKeyRepository interface {
	Create(ctx context.Context, key *entity.JWKKey) error
	GetByKid(ctx context.Context, kid string) (*entity.JWKKey, error)
	GetActiveKeys(ctx context.Context) ([]*entity.JWKKey, error)
	GetAllKeys(ctx context.Context) ([]*entity.JWKKey, error)
	Update(ctx context.Context, key *entity.JWKKey) error
	Delete(ctx context.Context, kid string) error
	RotateKey(ctx context.Context, kid string) error
	RetireKey(ctx context.Context, kid string) error
}
