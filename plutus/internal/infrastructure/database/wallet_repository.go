package database

import (
	"context"

	"github.com/julesChu12/fly/plutus/internal/domain/entity"
	"github.com/julesChu12/fly/plutus/internal/domain/repository"
	"github.com/julesChu12/fly/plutus/pkg/errors"
	"gorm.io/gorm"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) repository.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(ctx context.Context, wallet *entity.Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *walletRepository) GetByID(ctx context.Context, id uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.WithContext(ctx).First(&wallet, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrWalletNotFound
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByCustomerID(ctx context.Context, tenantID, customerID uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).First(&wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrWalletNotFound
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) Update(ctx context.Context, wallet *entity.Wallet) error {
	return r.db.WithContext(ctx).Save(wallet).Error
}

func (r *walletRepository) UpdateBalance(ctx context.Context, id uint, newBalance float64) error {
	return r.db.WithContext(ctx).Model(&entity.Wallet{}).Where("id = ?", id).Update("balance", newBalance).Error
}

func (r *walletRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Wallet{}, id).Error
}

func (r *walletRepository) List(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Wallet, int64, error) {
	var wallets []*entity.Wallet
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Wallet{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&wallets).Error
	return wallets, total, err
}

func (r *walletRepository) LockForUpdate(ctx context.Context, id uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").First(&wallet, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrWalletNotFound
		}
		return nil, err
	}
	return &wallet, nil
}
