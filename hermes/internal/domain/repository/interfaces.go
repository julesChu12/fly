package repository

import (
	"context"
	"hermes/internal/domain/entity"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *entity.Customer) error
	GetByID(ctx context.Context, id uint) (*entity.Customer, error)
	GetByEmail(ctx context.Context, email string) (*entity.Customer, error)
	Update(ctx context.Context, customer *entity.Customer) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]*entity.Customer, error)
	GetWithContacts(ctx context.Context, id uint) (*entity.Customer, error)
}

type ContactRepository interface {
	Create(ctx context.Context, contact *entity.Contact) error
	GetByID(ctx context.Context, id uint) (*entity.Contact, error)
	GetByCustomerID(ctx context.Context, customerID uint) ([]*entity.Contact, error)
	Update(ctx context.Context, contact *entity.Contact) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]*entity.Contact, error)
}