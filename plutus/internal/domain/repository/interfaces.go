package repository

import (
	"context"

	"github.com/julesChu12/fly/plutus/internal/domain/entity"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet *entity.Wallet) error
	GetByID(ctx context.Context, id uint) (*entity.Wallet, error)
	GetByCustomerID(ctx context.Context, tenantID, customerID uint) (*entity.Wallet, error)
	Update(ctx context.Context, wallet *entity.Wallet) error
	UpdateBalance(ctx context.Context, id uint, newBalance float64) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Wallet, int64, error)
	LockForUpdate(ctx context.Context, id uint) (*entity.Wallet, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *entity.Transaction) error
	GetByID(ctx context.Context, id uint) (*entity.Transaction, error)
	GetByIdempotencyKey(ctx context.Context, tenantID uint, key string) (*entity.Transaction, error)
	Update(ctx context.Context, transaction *entity.Transaction) error
	List(ctx context.Context, tenantID uint, walletID *uint, offset, limit int) ([]*entity.Transaction, int64, error)
	ListByOrderID(ctx context.Context, orderID uint) ([]*entity.Transaction, error)
	ListByWalletID(ctx context.Context, walletID uint, offset, limit int) ([]*entity.Transaction, int64, error)
}

type PaymentChannelRepository interface {
	Create(ctx context.Context, channel *entity.PaymentChannelEntity) error
	GetByID(ctx context.Context, id uint) (*entity.PaymentChannelEntity, error)
	GetByName(ctx context.Context, tenantID uint, name string) (*entity.PaymentChannelEntity, error)
	Update(ctx context.Context, channel *entity.PaymentChannelEntity) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.PaymentChannelEntity, int64, error)
	ListEnabled(ctx context.Context, tenantID uint) ([]*entity.PaymentChannelEntity, error)
}
