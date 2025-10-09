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

func (r *contactRepository) GetByCustomerID(ctx context.Context, customerID uint) ([]*entity.Contact, error) {
	var contacts []*entity.Contact
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Find(&contacts).Error
	return contacts, err
}

func (r *contactRepository) Update(ctx context.Context, contact *entity.Contact) error {
	return r.db.WithContext(ctx).Save(contact).Error
}

func (r *contactRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Contact{}, id).Error
}

func (r *contactRepository) List(ctx context.Context, offset, limit int) ([]*entity.Contact, error) {
	var contacts []*entity.Contact
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&contacts).Error
	return contacts, err
}
