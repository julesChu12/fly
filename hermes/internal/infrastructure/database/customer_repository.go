package database

import (
	"context"

	"github.com/julesChu12/fly/hermes/internal/domain/entity"
	"github.com/julesChu12/fly/hermes/internal/domain/repository"
	"github.com/julesChu12/fly/hermes/pkg/errors"
	"gorm.io/gorm"
)

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) repository.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	if err := r.db.WithContext(ctx).Create(customer).Error; err != nil {
		return err
	}
	return nil
}

func (r *customerRepository) GetByID(ctx context.Context, id uint) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).First(&customer, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCustomerNotFound
		}
		return nil, err
	}
	return &customer, nil
}

// GetByIDAndTenant gets a customer by ID with tenant isolation
func (r *customerRepository) GetByIDAndTenant(ctx context.Context, id uint, tenantID uint) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&customer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCustomerNotFound
		}
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&customer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCustomerNotFound
		}
		return nil, err
	}
	return &customer, nil
}

// GetByEmailAndTenant gets a customer by email with tenant isolation
func (r *customerRepository) GetByEmailAndTenant(ctx context.Context, email string, tenantID uint) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).Where("email = ? AND tenant_id = ?", email, tenantID).First(&customer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCustomerNotFound
		}
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) Update(ctx context.Context, customer *entity.Customer) error {
	return r.db.WithContext(ctx).Save(customer).Error
}

func (r *customerRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Customer{}, id).Error
}

// DeleteByTenant deletes a customer with tenant isolation
func (r *customerRepository) DeleteByTenant(ctx context.Context, id uint, tenantID uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&entity.Customer{}, id).Error
}

func (r *customerRepository) List(ctx context.Context, offset, limit int) ([]*entity.Customer, int64, error) {
	var customers []*entity.Customer
	var total int64

	if err := r.db.WithContext(ctx).Model(&entity.Customer{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&customers).Error
	return customers, total, err
}

// ListByTenant lists customers with tenant isolation
func (r *customerRepository) ListByTenant(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Customer, int64, error) {
	var customers []*entity.Customer
	var total int64

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if err := query.Model(&entity.Customer{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&customers).Error
	return customers, total, err
}

func (r *customerRepository) GetWithContacts(ctx context.Context, id uint) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).Preload("Contacts").First(&customer, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCustomerNotFound
		}
		return nil, err
	}
	return &customer, nil
}

// GetWithContactsByTenant gets a customer with contacts with tenant isolation
func (r *customerRepository) GetWithContactsByTenant(ctx context.Context, id uint, tenantID uint) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Preload("Contacts", "tenant_id = ?", tenantID).First(&customer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrCustomerNotFound
		}
		return nil, err
	}
	return &customer, nil
}
