package database

import (
	"context"

	"github.com/julesChu12/fly/plutus/internal/domain/entity"
	"github.com/julesChu12/fly/plutus/internal/domain/repository"
	"github.com/julesChu12/fly/plutus/pkg/errors"
	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) repository.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

func (r *transactionRepository) GetByID(ctx context.Context, id uint) (*entity.Transaction, error) {
	var transaction entity.Transaction
	err := r.db.WithContext(ctx).First(&transaction, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrTransactionNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByIdempotencyKey(ctx context.Context, tenantID uint, key string) (*entity.Transaction, error) {
	var transaction entity.Transaction
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND idempotency_key = ?", tenantID, key).First(&transaction).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrTransactionNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *entity.Transaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

func (r *transactionRepository) List(ctx context.Context, tenantID uint, walletID *uint, offset, limit int) ([]*entity.Transaction, int64, error) {
	var transactions []*entity.Transaction
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Transaction{}).Where("tenant_id = ?", tenantID)

	if walletID != nil {
		query = query.Where("wallet_id = ?", *walletID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&transactions).Error
	return transactions, total, err
}

func (r *transactionRepository) ListByOrderID(ctx context.Context, orderID uint) ([]*entity.Transaction, error) {
	var transactions []*entity.Transaction
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at ASC").Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) ListByWalletID(ctx context.Context, walletID uint, offset, limit int) ([]*entity.Transaction, int64, error) {
	var transactions []*entity.Transaction
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Transaction{}).Where("wallet_id = ?", walletID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&transactions).Error
	return transactions, total, err
}

// Payment Channel Repository
type paymentChannelRepository struct {
	db *gorm.DB
}

func NewPaymentChannelRepository(db *gorm.DB) repository.PaymentChannelRepository {
	return &paymentChannelRepository{db: db}
}

func (r *paymentChannelRepository) Create(ctx context.Context, channel *entity.PaymentChannelEntity) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

func (r *paymentChannelRepository) GetByID(ctx context.Context, id uint) (*entity.PaymentChannelEntity, error) {
	var channel entity.PaymentChannelEntity
	err := r.db.WithContext(ctx).First(&channel, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrPaymentChannelNotFound
		}
		return nil, err
	}
	return &channel, nil
}

func (r *paymentChannelRepository) GetByName(ctx context.Context, tenantID uint, name string) (*entity.PaymentChannelEntity, error) {
	var channel entity.PaymentChannelEntity
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND name = ?", tenantID, name).First(&channel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrPaymentChannelNotFound
		}
		return nil, err
	}
	return &channel, nil
}

func (r *paymentChannelRepository) Update(ctx context.Context, channel *entity.PaymentChannelEntity) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

func (r *paymentChannelRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.PaymentChannelEntity{}, id).Error
}

func (r *paymentChannelRepository) List(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.PaymentChannelEntity, int64, error) {
	var channels []*entity.PaymentChannelEntity
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.PaymentChannelEntity{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&channels).Error
	return channels, total, err
}

func (r *paymentChannelRepository) ListEnabled(ctx context.Context, tenantID uint) ([]*entity.PaymentChannelEntity, error) {
	var channels []*entity.PaymentChannelEntity
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND is_enabled = ?", tenantID, true).Order("created_at ASC").Find(&channels).Error
	return channels, err
}
