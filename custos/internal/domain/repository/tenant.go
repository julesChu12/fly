package repository

import (
	"context"
	"errors"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrTenantSlugExists = errors.New("tenant slug already exists")
	ErrTenantDomainExists = errors.New("tenant domain already exists")
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *entity.Tenant) error
	GetByID(ctx context.Context, id uint) (*entity.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*entity.Tenant, error)
	Update(ctx context.Context, tenant *entity.Tenant) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*entity.Tenant, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ExistsByDomain(ctx context.Context, domain string) (bool, error)
	GetUserCount(ctx context.Context, tenantID uint) (int64, error)
}