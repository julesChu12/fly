package database

import (
	"context"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/internal/domain/repository"
	"github.com/julesChu12/fly/kratos/pkg/errors"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) GetByID(ctx context.Context, id uint) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetWithItems(ctx context.Context, id uint) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).Preload("Items").First(&order, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint, status entity.OrderStatus, reason string, operatorID *uint) error {
	updates := map[string]interface{}{
		"status": status,
	}

	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *orderRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Order{}, id).Error
}

func (r *orderRepository) List(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Order{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}

func (r *orderRepository) ListByCustomer(ctx context.Context, tenantID uint, customerID uint, offset, limit int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Order{}).Where("tenant_id = ? AND customer_id = ?", tenantID, customerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}

func (r *orderRepository) ListByStatus(ctx context.Context, tenantID uint, status entity.OrderStatus, offset, limit int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Order{}).Where("tenant_id = ? AND status = ?", tenantID, status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}
