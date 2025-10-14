package database

import (
	"context"

	"github.com/julesChu12/fly/hermes/internal/domain/entity"
	"github.com/julesChu12/fly/hermes/internal/domain/repository"
	"github.com/julesChu12/fly/hermes/pkg/errors"
	"gorm.io/gorm"
)

type contactRepository struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) repository.ContactRepository {
	return &contactRepository{db: db}
}

func (r *contactRepository) Create(ctx context.Context, contact *entity.Contact) error {
	return r.db.WithContext(ctx).Create(contact).Error
}

func (r *contactRepository) GetByID(ctx context.Context, id uint) (*entity.Contact, error) {
	var contact entity.Contact
	err := r.db.WithContext(ctx).First(&contact, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrContactNotFound
		}
		return nil, err
	}
	return &contact, nil
}

// GetByIDAndTenant gets a contact by ID with tenant isolation
func (r *contactRepository) GetByIDAndTenant(ctx context.Context, id uint, tenantID uint) (*entity.Contact, error) {
	var contact entity.Contact
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrContactNotFound
		}
		return nil, err
	}
	return &contact, nil
}

func (r *contactRepository) GetByCustomerID(ctx context.Context, customerID uint) ([]*entity.Contact, error) {
	var contacts []*entity.Contact
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Find(&contacts).Error
	return contacts, err
}

// GetByCustomerIDAndTenant gets contacts by customer ID with tenant isolation
func (r *contactRepository) GetByCustomerIDAndTenant(ctx context.Context, customerID uint, tenantID uint) ([]*entity.Contact, error) {
	var contacts []*entity.Contact
	err := r.db.WithContext(ctx).Where("customer_id = ? AND tenant_id = ?", customerID, tenantID).Find(&contacts).Error
	return contacts, err
}

func (r *contactRepository) Update(ctx context.Context, contact *entity.Contact) error {
	return r.db.WithContext(ctx).Save(contact).Error
}

func (r *contactRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Contact{}, id).Error
}

// DeleteByTenant deletes a contact with tenant isolation
func (r *contactRepository) DeleteByTenant(ctx context.Context, id uint, tenantID uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&entity.Contact{}, id).Error
}

func (r *contactRepository) List(ctx context.Context, offset, limit int) ([]*entity.Contact, error) {
	var contacts []*entity.Contact
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&contacts).Error
	return contacts, err
}

// ListByTenant lists contacts with tenant isolation
func (r *contactRepository) ListByTenant(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Contact, error) {
	var contacts []*entity.Contact
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Offset(offset).Limit(limit).Find(&contacts).Error
	return contacts, err
}
