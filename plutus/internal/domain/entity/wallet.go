package entity

import (
	"time"
)

type WalletStatus string
type TransactionType string
type PaymentChannel string
type TransactionStatus string

const (
	WalletStatusActive WalletStatus = "active"
	WalletStatusFrozen WalletStatus = "frozen"
)

const (
	TransactionTypeRecharge TransactionType = "recharge"
	TransactionTypeConsume  TransactionType = "consume"
	TransactionTypeRefund   TransactionType = "refund"
)

const (
	ChannelWallet PaymentChannel = "wallet"
	ChannelWechat PaymentChannel = "wechat"
	ChannelAlipay PaymentChannel = "alipay"
	ChannelStripe PaymentChannel = "stripe"
	ChannelPaypal PaymentChannel = "paypal"
	ChannelBank   PaymentChannel = "bank"
	ChannelOther  PaymentChannel = "other"
)

const (
	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusSuccess TransactionStatus = "success"
	TransactionStatusFailed  TransactionStatus = "failed"
)

type Wallet struct {
	ID         uint          `json:"id" gorm:"primaryKey"`
	TenantID   uint          `json:"tenant_id" gorm:"not null;index"`
	CustomerID uint          `json:"customer_id" gorm:"not null;index"`
	Balance    float64       `json:"balance" gorm:"type:decimal(14,2);default:0"`
	Currency   string        `json:"currency" gorm:"size:3;default:CNY"`
	Status     WalletStatus  `json:"status" gorm:"type:enum('active','frozen');default:active"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`

	// Relations
	Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:WalletID"`
}

type Transaction struct {
	ID              uint              `json:"id" gorm:"primaryKey"`
	TenantID        uint              `json:"tenant_id" gorm:"not null;index"`
	WalletID        uint              `json:"wallet_id" gorm:"not null;index"`
	OrderID         *uint             `json:"order_id"`
	Type            TransactionType   `json:"type" gorm:"type:enum('recharge','consume','refund');not null"`
	Amount          float64           `json:"amount" gorm:"type:decimal(14,2);not null"`
	Currency        string            `json:"currency" gorm:"size:3;default:CNY"`
	Channel         PaymentChannel    `json:"channel" gorm:"type:enum('wallet','wechat','alipay','stripe','paypal','bank','other');default:wallet"`
	Status          TransactionStatus `json:"status" gorm:"type:enum('pending','success','failed');default:pending"`
	IdempotencyKey  *string           `json:"idempotency_key" gorm:"size:128"`
	ReferenceNo     *string           `json:"reference_no" gorm:"size:128"`
	Meta            string            `json:"meta" gorm:"type:json"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`

	// Relations
	Wallet Wallet `json:"wallet,omitempty" gorm:"foreignKey:WalletID"`
}

type PaymentChannelEntity struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TenantID  uint      `json:"tenant_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"size:128;not null"`
	Type      string    `json:"type" gorm:"size:32;not null"`
	Config    string    `json:"config" gorm:"type:json"`
	IsEnabled bool      `json:"is_enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}