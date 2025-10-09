package repository

import (
	"context"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
)

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id uint) (*entity.Order, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error)
	GetWithItems(ctx context.Context, id uint) (*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
	UpdateStatus(ctx context.Context, id uint, status entity.OrderStatus, reason string, operatorID *uint) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Order, int64, error)
	ListByCustomer(ctx context.Context, tenantID uint, customerID uint, offset, limit int) ([]*entity.Order, int64, error)
	ListByStatus(ctx context.Context, tenantID uint, status entity.OrderStatus, offset, limit int) ([]*entity.Order, int64, error)
}

type OrderItemRepository interface {
	Create(ctx context.Context, item *entity.OrderItem) error
	GetByID(ctx context.Context, id uint) (*entity.OrderItem, error)
	GetByOrderID(ctx context.Context, orderID uint) ([]*entity.OrderItem, error)
	Update(ctx context.Context, item *entity.OrderItem) error
	Delete(ctx context.Context, id uint) error
	DeleteByOrderID(ctx context.Context, orderID uint) error
}

type OrderStatusLogRepository interface {
	Create(ctx context.Context, log *entity.OrderStatusLog) error
	GetByOrderID(ctx context.Context, orderID uint) ([]*entity.OrderStatusLog, error)
	List(ctx context.Context, offset, limit int) ([]*entity.OrderStatusLog, error)
}

type OrderAuditRepository interface {
	Create(ctx context.Context, audit *entity.OrderAudit) error
	GetByOrderID(ctx context.Context, orderID uint) ([]*entity.OrderAudit, error)
	List(ctx context.Context, offset, limit int) ([]*entity.OrderAudit, error)
}
