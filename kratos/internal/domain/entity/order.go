package entity

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusFulfilled OrderStatus = "fulfilled"
	OrderStatusCanceled  OrderStatus = "canceled"
)

type Order struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	TenantID     uint        `json:"tenant_id" gorm:"not null;index"`
	OrderNo      string      `json:"order_no" gorm:"uniqueIndex;size:64;not null"`
	CustomerID   uint        `json:"customer_id" gorm:"not null;index"`
	TotalAmount  float64     `json:"total_amount" gorm:"type:decimal(12,2);not null"`
	Currency     string      `json:"currency" gorm:"size:3;default:CNY"`
	Status       OrderStatus `json:"status" gorm:"type:enum('pending','paid','fulfilled','canceled');default:pending;index"`
	Remark       string      `json:"remark" gorm:"size:255"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`

	// Relations
	Items      []OrderItem      `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	StatusLogs []OrderStatusLog `json:"status_logs,omitempty" gorm:"foreignKey:OrderID"`
	Audits     []OrderAudit     `json:"audits,omitempty" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TenantID    uint      `json:"tenant_id" gorm:"not null;index"`
	OrderID     uint      `json:"order_id" gorm:"not null;index"`
	ProductID   *uint     `json:"product_id"`
	ProductName string    `json:"product_name" gorm:"size:256;not null"`
	SKU         string    `json:"sku" gorm:"size:128"`
	Quantity    int       `json:"quantity" gorm:"default:1;not null"`
	UnitPrice   float64   `json:"unit_price" gorm:"type:decimal(12,2);not null"`
	TotalPrice  float64   `json:"total_price" gorm:"type:decimal(12,2) GENERATED ALWAYS AS (quantity * unit_price) STORED"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrderStatusLog struct {
	ID         uint        `json:"id" gorm:"primaryKey"`
	TenantID   uint        `json:"tenant_id" gorm:"not null;index"`
	OrderID    uint        `json:"order_id" gorm:"not null;index"`
	FromStatus *OrderStatus `json:"from_status" gorm:"type:enum('pending','paid','fulfilled','canceled')"`
	ToStatus   OrderStatus `json:"to_status" gorm:"type:enum('pending','paid','fulfilled','canceled');not null"`
	Reason     string      `json:"reason" gorm:"size:255"`
	OperatorID *uint       `json:"operator_id"`
	CreatedAt  time.Time   `json:"created_at"`
}

type OrderAudit struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TenantID  uint      `json:"tenant_id" gorm:"not null;index"`
	OrderID   uint      `json:"order_id" gorm:"not null;index"`
	Action    string    `json:"action" gorm:"size:64;not null"`
	Actor     string    `json:"actor" gorm:"size:128"`
	Payload   string    `json:"payload" gorm:"type:json"`
	CreatedAt time.Time `json:"created_at"`
}