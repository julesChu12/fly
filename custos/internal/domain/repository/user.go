package repository

import (
	"context"
	"errors"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserOAuthNotFound = errors.New("user oauth binding not found")
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*entity.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Multi-tenant methods
	GetByIDWithTenant(ctx context.Context, id uint, tenantID uint) (*entity.User, error)
	GetByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (*entity.User, error)
	GetByEmailWithTenant(ctx context.Context, email string, tenantID uint) (*entity.User, error)
	ListByTenant(ctx context.Context, tenantID uint, limit, offset int) ([]*entity.User, error)
	ExistsByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (bool, error)
	ExistsByEmailWithTenant(ctx context.Context, email string, tenantID uint) (bool, error)
}
