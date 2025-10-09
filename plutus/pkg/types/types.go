package types

import (
	"time"

	"github.com/julesChu12/fly/plutus/internal/domain/entity"
)

// Wallet Request/Response Types
type CreateWalletRequest struct {
	CustomerID uint   `json:"customer_id" binding:"required"`
	Currency   string `json:"currency"`
}

type WalletResponse struct {
	ID         uint                `json:"id"`
	TenantID   uint                `json:"tenant_id"`
	CustomerID uint                `json:"customer_id"`
	Balance    float64             `json:"balance"`
	Currency   string              `json:"currency"`
	Status     entity.WalletStatus `json:"status"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

// Transaction Request/Response Types
type RechargeRequest struct {
	CustomerID     uint                   `json:"customer_id" binding:"required"`
	Amount         float64                `json:"amount" binding:"required,min=0.01"`
	Currency       string                 `json:"currency"`
	Channel        entity.PaymentChannel  `json:"channel" binding:"required"`
	IdempotencyKey *string                `json:"idempotency_key"`
	ReferenceNo    *string                `json:"reference_no"`
	Meta           map[string]interface{} `json:"meta"`
}

type ConsumeRequest struct {
	CustomerID     uint                   `json:"customer_id" binding:"required"`
	OrderID        *uint                  `json:"order_id"`
	Amount         float64                `json:"amount" binding:"required,min=0.01"`
	Currency       string                 `json:"currency"`
	IdempotencyKey *string                `json:"idempotency_key"`
	Meta           map[string]interface{} `json:"meta"`
}

type RefundRequest struct {
	CustomerID     uint                   `json:"customer_id" binding:"required"`
	OrderID        *uint                  `json:"order_id"`
	Amount         float64                `json:"amount" binding:"required,min=0.01"`
	Currency       string                 `json:"currency"`
	IdempotencyKey *string                `json:"idempotency_key"`
	ReferenceNo    *string                `json:"reference_no"`
	Meta           map[string]interface{} `json:"meta"`
}

type TransactionResponse struct {
	ID             uint                     `json:"id"`
	TenantID       uint                     `json:"tenant_id"`
	WalletID       uint                     `json:"wallet_id"`
	OrderID        *uint                    `json:"order_id"`
	Type           entity.TransactionType   `json:"type"`
	Amount         float64                  `json:"amount"`
	Currency       string                   `json:"currency"`
	Channel        entity.PaymentChannel    `json:"channel"`
	Status         entity.TransactionStatus `json:"status"`
	IdempotencyKey *string                  `json:"idempotency_key"`
	ReferenceNo    *string                  `json:"reference_no"`
	Meta           string                   `json:"meta"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
}

// Payment Channel Request/Response Types
type CreatePaymentChannelRequest struct {
	Name      string                 `json:"name" binding:"required"`
	Type      string                 `json:"type" binding:"required"`
	Config    map[string]interface{} `json:"config"`
	IsEnabled bool                   `json:"is_enabled"`
}

type PaymentChannelResponse struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Config    string    `json:"config"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// List Request/Response Types
type ListRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

type ListWalletsRequest struct {
	ListRequest
	CustomerID *uint `form:"customer_id" json:"customer_id"`
}

type ListTransactionsRequest struct {
	ListRequest
	WalletID *uint                     `form:"wallet_id" json:"wallet_id"`
	OrderID  *uint                     `form:"order_id" json:"order_id"`
	Type     *entity.TransactionType   `form:"type" json:"type"`
	Status   *entity.TransactionStatus `form:"status" json:"status"`
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
