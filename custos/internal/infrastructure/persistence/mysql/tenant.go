package mysql

import (
	"context"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	moradb "github.com/julesChu12/fly/mora/pkg/db"
)

type TenantRepository struct {
	client *moradb.Client
}

func NewTenantRepository(client *moradb.Client) repository.TenantRepository {
	return &TenantRepository{client: client}
}

func (r *TenantRepository) Create(ctx context.Context, tenant *entity.Tenant) error {
	return r.client.Create(ctx, tenant)
}

func (r *TenantRepository) GetByID(ctx context.Context, id uint) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.client.First(ctx, &tenant, id)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.client.DB().WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) GetByDomain(ctx context.Context, domain string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.client.DB().WithContext(ctx).Where("domain = ?", domain).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) Update(ctx context.Context, tenant *entity.Tenant) error {
	return r.client.Save(ctx, tenant)
}

func (r *TenantRepository) Delete(ctx context.Context, id uint) error {
	return r.client.Delete(ctx, &entity.Tenant{}, id)
}

func (r *TenantRepository) List(ctx context.Context, limit, offset int) ([]*entity.Tenant, error) {
	var tenants []*entity.Tenant
	err := r.client.Paginate(ctx, &tenants, (offset/limit)+1, limit)
	return tenants, err
}

func (r *TenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return r.client.Exists(ctx, &entity.Tenant{}, "slug = ?", slug)
}

func (r *TenantRepository) ExistsByDomain(ctx context.Context, domain string) (bool, error) {
	return r.client.Exists(ctx, &entity.Tenant{}, "domain = ?", domain)
}

func (r *TenantRepository) GetUserCount(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	err := r.client.Count(ctx, &entity.User{}, &count, "tenant_id = ?", tenantID)
	return count, err
}
