package types

import (
	"time"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
)

// Order Request/Response Types
type CreateOrderRequest struct {
	OrderNo     string                   `json:"order_no" binding:"required"`
	CustomerID  uint                     `json:"customer_id" binding:"required"`
	TotalAmount float64                  `json:"total_amount" binding:"required,min=0"`
	Currency    string                   `json:"currency"`
	Remark      string                   `json:"remark"`
	Items       []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

type CreateOrderItemRequest struct {
	ProductID   *uint   `json:"product_id"`
	ProductName string  `json:"product_name" binding:"required"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" binding:"required,min=0"`
}

type UpdateOrderStatusRequest struct {
	Status     entity.OrderStatus `json:"status" binding:"required"`
	Reason     string             `json:"reason"`
	OperatorID *uint              `json:"operator_id"`
}

type OrderResponse struct {
	ID          uint                     `json:"id"`
	TenantID    uint                     `json:"tenant_id"`
	OrderNo     string                   `json:"order_no"`
	CustomerID  uint                     `json:"customer_id"`
	TotalAmount float64                  `json:"total_amount"`
	Currency    string                   `json:"currency"`
	Status      entity.OrderStatus       `json:"status"`
	Remark      string                   `json:"remark"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Items       []OrderItemResponse      `json:"items,omitempty"`
	StatusLogs  []OrderStatusLogResponse `json:"status_logs,omitempty"`
	Audits      []OrderAuditResponse     `json:"audits,omitempty"`
}

type OrderItemResponse struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	OrderID     uint      `json:"order_id"`
	ProductID   *uint     `json:"product_id"`
	ProductName string    `json:"product_name"`
	SKU         string    `json:"sku"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	TotalPrice  float64   `json:"total_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrderStatusLogResponse struct {
	ID         uint                `json:"id"`
	TenantID   uint                `json:"tenant_id"`
	OrderID    uint                `json:"order_id"`
	FromStatus *entity.OrderStatus `json:"from_status"`
	ToStatus   entity.OrderStatus  `json:"to_status"`
	Reason     string              `json:"reason"`
	OperatorID *uint               `json:"operator_id"`
	CreatedAt  time.Time           `json:"created_at"`
}

type OrderAuditResponse struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	OrderID   uint      `json:"order_id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

// List Request/Response Types
type ListRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

type ListOrdersRequest struct {
	ListRequest
	CustomerID *uint               `form:"customer_id" json:"customer_id"`
	Status     *entity.OrderStatus `form:"status" json:"status"`
}

type ListResponse struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// Standard API Response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
