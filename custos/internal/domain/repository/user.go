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

// UserListFilter defines filter options for listing users
type UserListFilter struct {
	Status   *string
	Role     *string
	UserType *string
	TenantID *uint
	Keyword  *string // Search by username or email
}

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

	// Advanced list with filters
	ListWithFilter(ctx context.Context, filter *UserListFilter, limit, offset int) ([]*entity.User, int64, error)

	// Statistics methods
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountByRole(ctx context.Context) (map[string]int64, error)
	CountByType(ctx context.Context) (map[string]int64, error)
	CountNewUsers(ctx context.Context, since string) (int64, error) // since format: "2006-01-02"
	CountTotal(ctx context.Context) (int64, error)

	// Multi-tenant methods
	GetByIDWithTenant(ctx context.Context, id uint, tenantID uint) (*entity.User, error)
	GetByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (*entity.User, error)
	GetByEmailWithTenant(ctx context.Context, email string, tenantID uint) (*entity.User, error)
	ListByTenant(ctx context.Context, tenantID uint, limit, offset int) ([]*entity.User, error)
	ExistsByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (bool, error)
	ExistsByEmailWithTenant(ctx context.Context, email string, tenantID uint) (bool, error)
}
