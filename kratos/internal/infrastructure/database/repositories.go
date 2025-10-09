package database

import (
	"context"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/internal/domain/repository"
	"github.com/julesChu12/fly/kratos/pkg/errors"
	"gorm.io/gorm"
)

type orderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) repository.OrderItemRepository {
	return &orderItemRepository{db: db}
}

func (r *orderItemRepository) Create(ctx context.Context, item *entity.OrderItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *orderItemRepository) GetByID(ctx context.Context, id uint) (*entity.OrderItem, error) {
	var item entity.OrderItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrOrderItemNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *orderItemRepository) GetByOrderID(ctx context.Context, orderID uint) ([]*entity.OrderItem, error) {
	var items []*entity.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

func (r *orderItemRepository) Update(ctx context.Context, item *entity.OrderItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *orderItemRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.OrderItem{}, id).Error
}

func (r *orderItemRepository) DeleteByOrderID(ctx context.Context, orderID uint) error {
	return r.db.WithContext(ctx).Where("order_id = ?", orderID).Delete(&entity.OrderItem{}).Error
}

// Order Status Log Repository
type orderStatusLogRepository struct {
	db *gorm.DB
}

func NewOrderStatusLogRepository(db *gorm.DB) repository.OrderStatusLogRepository {
	return &orderStatusLogRepository{db: db}
}

func (r *orderStatusLogRepository) Create(ctx context.Context, log *entity.OrderStatusLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *orderStatusLogRepository) GetByOrderID(ctx context.Context, orderID uint) ([]*entity.OrderStatusLog, error) {
	var logs []*entity.OrderStatusLog
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at ASC").Find(&logs).Error
	return logs, err
}

func (r *orderStatusLogRepository) List(ctx context.Context, offset, limit int) ([]*entity.OrderStatusLog, error) {
	var logs []*entity.OrderStatusLog
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// Order Audit Repository
type orderAuditRepository struct {
	db *gorm.DB
}

func NewOrderAuditRepository(db *gorm.DB) repository.OrderAuditRepository {
	return &orderAuditRepository{db: db}
}

func (r *orderAuditRepository) Create(ctx context.Context, audit *entity.OrderAudit) error {
	return r.db.WithContext(ctx).Create(audit).Error
}

func (r *orderAuditRepository) GetByOrderID(ctx context.Context, orderID uint) ([]*entity.OrderAudit, error) {
	var audits []*entity.OrderAudit
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at ASC").Find(&audits).Error
	return audits, err
}

func (r *orderAuditRepository) List(ctx context.Context, offset, limit int) ([]*entity.OrderAudit, error) {
	var audits []*entity.OrderAudit
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&audits).Error
	return audits, err
}
