package repository

import (
	"context"

	"github.com/julesChu12/fly/hermes/internal/domain/entity"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *entity.Customer) error
	GetByID(ctx context.Context, id uint) (*entity.Customer, error)
	GetByIDAndTenant(ctx context.Context, id uint, tenantID uint) (*entity.Customer, error)
	GetByEmail(ctx context.Context, email string) (*entity.Customer, error)
	GetByEmailAndTenant(ctx context.Context, email string, tenantID uint) (*entity.Customer, error)
	Update(ctx context.Context, customer *entity.Customer) error
	Delete(ctx context.Context, id uint) error
	DeleteByTenant(ctx context.Context, id uint, tenantID uint) error
	List(ctx context.Context, offset, limit int) ([]*entity.Customer, int64, error)
	ListByTenant(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Customer, int64, error)
	GetWithContacts(ctx context.Context, id uint) (*entity.Customer, error)
	GetWithContactsByTenant(ctx context.Context, id uint, tenantID uint) (*entity.Customer, error)
}

type ContactRepository interface {
	Create(ctx context.Context, contact *entity.Contact) error
	GetByID(ctx context.Context, id uint) (*entity.Contact, error)
	GetByIDAndTenant(ctx context.Context, id uint, tenantID uint) (*entity.Contact, error)
	GetByCustomerID(ctx context.Context, customerID uint) ([]*entity.Contact, error)
	GetByCustomerIDAndTenant(ctx context.Context, customerID uint, tenantID uint) ([]*entity.Contact, error)
	Update(ctx context.Context, contact *entity.Contact) error
	Delete(ctx context.Context, id uint) error
	DeleteByTenant(ctx context.Context, id uint, tenantID uint) error
	List(ctx context.Context, offset, limit int) ([]*entity.Contact, error)
	ListByTenant(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Contact, error)
}
