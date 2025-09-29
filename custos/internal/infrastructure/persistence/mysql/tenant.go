package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) repository.TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) Create(ctx context.Context, tenant *entity.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *TenantRepository) GetByID(ctx context.Context, id uint) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.db.WithContext(ctx).First(&tenant, id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) GetByDomain(ctx context.Context, domain string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) Update(ctx context.Context, tenant *entity.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *TenantRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Tenant{}, id).Error
}

func (r *TenantRepository) List(ctx context.Context, limit, offset int) ([]*entity.Tenant, error) {
	var tenants []*entity.Tenant
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&tenants).Error
	return tenants, err
}

func (r *TenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Tenant{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *TenantRepository) ExistsByDomain(ctx context.Context, domain string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Tenant{}).Where("domain = ?", domain).Count(&count).Error
	return count > 0, err
}

func (r *TenantRepository) GetUserCount(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}