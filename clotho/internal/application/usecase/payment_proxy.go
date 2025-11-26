package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// PaymentProxy represents the use case for payment and wallet operations
type PaymentProxy struct {
	paymentClient *client.PaymentHTTPClient
	logger        *logger.Logger
}

// NewPaymentProxy creates a new PaymentProxy instance
func NewPaymentProxy(paymentClient *client.PaymentHTTPClient, logger *logger.Logger) *PaymentProxy {
	return &PaymentProxy{
		paymentClient: paymentClient,
		logger:        logger,
	}
}

// ListWallets retrieves a list of wallets with filters
func (p *PaymentProxy) ListWallets(ctx context.Context, filter *client.WalletFilter) (*client.ListResponse, error) {
	p.logger.Info("Retrieving wallet list", "filter", filter)

	// Set defaults for filter
	if filter != nil {
		if filter.Page <= 0 {
			filter.Page = 1
		}
		if filter.PageSize <= 0 {
			filter.PageSize = 20
		}
	}

	start := time.Now()
	wallets, err := p.paymentClient.ListWallets(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve wallet list", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve wallet list: %w", err)
	}

	p.logger.Info("Successfully retrieved wallet list", "count", len(wallets.Data.([]client.Wallet)), "total", wallets.Total, "duration", duration)
	return wallets, nil
}

// CreateWallet creates a new wallet
func (p *PaymentProxy) CreateWallet(ctx context.Context, req *client.CreateWalletRequestHTTP) (*client.Wallet, error) {
	p.logger.Info("Creating new wallet", "customer_id", req.CustomerID)

	// Validate request
	if err := p.validateCreateWalletRequest(req); err != nil {
		p.logger.Error("Invalid wallet creation request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	wallet, err := p.paymentClient.CreateWallet(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to create wallet", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	p.logger.Info("Successfully created wallet", "wallet_id", wallet.ID, "customer_id", wallet.CustomerID, "duration", duration)
	return wallet, nil
}

// GetWallet retrieves a wallet by ID
func (p *PaymentProxy) GetWallet(ctx context.Context, walletID uint) (*client.Wallet, error) {
	p.logger.Info("Retrieving wallet", "wallet_id", walletID)

	if walletID == 0 {
		return nil, fmt.Errorf("wallet ID is required")
	}

	start := time.Now()
	wallet, err := p.paymentClient.GetWallet(ctx, walletID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve wallet", "wallet_id", walletID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve wallet: %w", err)
	}

	p.logger.Info("Successfully retrieved wallet", "wallet_id", wallet.ID, "customer_id", wallet.CustomerID, "balance", wallet.Balance, "duration", duration)
	return wallet, nil
}

// GetWalletByCustomerID retrieves a wallet by customer ID
func (p *PaymentProxy) GetWalletByCustomerID(ctx context.Context, customerID uint) (*client.Wallet, error) {
	p.logger.Info("Retrieving wallet by customer", "customer_id", customerID)

	if customerID == 0 {
		return nil, fmt.Errorf("customer ID is required")
	}

	start := time.Now()
	wallet, err := p.paymentClient.GetWalletByCustomerID(ctx, customerID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve wallet by customer", "customer_id", customerID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve wallet by customer: %w", err)
	}

	p.logger.Info("Successfully retrieved wallet by customer", "customer_id", customerID, "wallet_id", wallet.ID, "balance", wallet.Balance, "duration", duration)
	return wallet, nil
}

// Recharge adds funds to a wallet
func (p *PaymentProxy) Recharge(ctx context.Context, req *client.RechargeRequestHTTP) (*client.Transaction, error) {
	p.logger.Info("Recharging wallet", "customer_id", req.CustomerID, "amount", req.Amount, "channel", req.Channel)

	// Validate request
	if err := p.validateRechargeRequest(req); err != nil {
		p.logger.Error("Invalid recharge request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	transaction, err := p.paymentClient.Recharge(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to recharge wallet", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to recharge wallet: %w", err)
	}

	p.logger.Info("Successfully recharged wallet", "transaction_id", transaction.ID, "customer_id", req.CustomerID, "amount", req.Amount, "duration", duration)
	return transaction, nil
}

// Consume deducts funds from a wallet
func (p *PaymentProxy) Consume(ctx context.Context, req *client.ConsumeRequestHTTP) (*client.Transaction, error) {
	p.logger.Info("Consuming from wallet", "customer_id", req.CustomerID, "amount", req.Amount)

	if req.OrderID != nil {
		p.logger.Debug("Consume with order", "order_id", *req.OrderID)
	}

	// Validate request
	if err := p.validateConsumeRequest(req); err != nil {
		p.logger.Error("Invalid consume request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	transaction, err := p.paymentClient.Consume(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to consume from wallet", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to consume from wallet: %w", err)
	}

	p.logger.Info("Successfully consumed from wallet", "transaction_id", transaction.ID, "customer_id", req.CustomerID, "amount", req.Amount, "duration", duration)
	return transaction, nil
}

// Refund returns funds to a wallet
func (p *PaymentProxy) Refund(ctx context.Context, req *client.RefundRequestHTTP) (*client.Transaction, error) {
	p.logger.Info("Refunding to wallet", "customer_id", req.CustomerID, "amount", req.Amount)

	if req.OrderID != nil {
		p.logger.Debug("Refund for order", "order_id", *req.OrderID)
	}

	// Validate request
	if err := p.validateRefundRequest(req); err != nil {
		p.logger.Error("Invalid refund request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	transaction, err := p.paymentClient.Refund(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to refund to wallet", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to refund to wallet: %w", err)
	}

	p.logger.Info("Successfully refunded to wallet", "transaction_id", transaction.ID, "customer_id", req.CustomerID, "amount", req.Amount, "duration", duration)
	return transaction, nil
}

// ListTransactions retrieves a list of transactions with filters
func (p *PaymentProxy) ListTransactions(ctx context.Context, filter *client.TransactionFilter) (*client.ListResponse, error) {
	p.logger.Info("Retrieving transaction list", "filter", filter)

	// Set defaults for filter
	if filter != nil {
		if filter.Page <= 0 {
			filter.Page = 1
		}
		if filter.PageSize <= 0 {
			filter.PageSize = 20
		}
	}

	start := time.Now()
	transactions, err := p.paymentClient.ListTransactions(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve transaction list", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve transaction list: %w", err)
	}

	p.logger.Info("Successfully retrieved transaction list", "count", len(transactions.Data.([]client.Transaction)), "total", transactions.Total, "duration", duration)
	return transactions, nil
}

// GetTransactionsByOrderID retrieves transactions for a specific order
func (p *PaymentProxy) GetTransactionsByOrderID(ctx context.Context, orderID uint) (*client.ListResponse, error) {
	p.logger.Info("Retrieving transactions for order", "order_id", orderID)

	if orderID == 0 {
		return nil, fmt.Errorf("order ID is required")
	}

	start := time.Now()
	transactions, err := p.paymentClient.GetTransactionsByOrderID(ctx, orderID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve order transactions", "order_id", orderID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve order transactions: %w", err)
	}

	p.logger.Info("Successfully retrieved order transactions", "order_id", orderID, "count", len(transactions.Data.([]client.Transaction)), "duration", duration)
	return transactions, nil
}

// GetWalletBalance gets the current balance of a customer's wallet
func (p *PaymentProxy) GetWalletBalance(ctx context.Context, customerID uint) (float64, error) {
	p.logger.Info("Getting wallet balance", "customer_id", customerID)

	if customerID == 0 {
		return 0, fmt.Errorf("customer ID is required")
	}

	wallet, err := p.GetWalletByCustomerID(ctx, customerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get wallet for balance: %w", err)
	}

	p.logger.Info("Successfully retrieved wallet balance", "customer_id", customerID, "balance", wallet.Balance)
	return wallet.Balance, nil
}

// ProcessPaymentForOrder processes payment for an order (consume operation)
func (p *PaymentProxy) ProcessPaymentForOrder(ctx context.Context, customerID, orderID uint, amount float64, currency string) (*client.Transaction, error) {
	p.logger.Info("Processing payment for order", "customer_id", customerID, "order_id", orderID, "amount", amount, "currency", currency)

	idempotencyKey := fmt.Sprintf("order_%d_%d", orderID, customerID)
	req := &client.ConsumeRequestHTTP{
		CustomerID:     customerID,
		OrderID:        &orderID,
		Amount:         amount,
		Currency:       currency,
		IdempotencyKey: &idempotencyKey,
		Meta: map[string]interface{}{
			"order_id":    orderID,
			"customer_id": customerID,
			"source":      "order_payment",
		},
	}

	transaction, err := p.Consume(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to process payment for order: %w", err)
	}

	p.logger.Info("Successfully processed payment for order", "transaction_id", transaction.ID, "order_id", orderID, "amount", amount)
	return transaction, nil
}

// ProcessRefundForOrder processes refund for an order
func (p *PaymentProxy) ProcessRefundForOrder(ctx context.Context, customerID, orderID uint, amount float64, currency string, reason string) (*client.Transaction, error) {
	p.logger.Info("Processing refund for order", "customer_id", customerID, "order_id", orderID, "amount", amount, "reason", reason)

	idempotencyKey := fmt.Sprintf("refund_order_%d_%d", orderID, customerID)
	req := &client.RefundRequestHTTP{
		CustomerID:     customerID,
		OrderID:        &orderID,
		Amount:         amount,
		Currency:       currency,
		IdempotencyKey: &idempotencyKey,
		Meta: map[string]interface{}{
			"order_id":    orderID,
			"customer_id": customerID,
			"reason":      reason,
			"source":      "order_refund",
		},
	}

	transaction, err := p.Refund(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund for order: %w", err)
	}

	p.logger.Info("Successfully processed refund for order", "transaction_id", transaction.ID, "order_id", orderID, "amount", amount)
	return transaction, nil
}

// Validation helper functions

func (p *PaymentProxy) validateCreateWalletRequest(req *client.CreateWalletRequestHTTP) error {
	if req.CustomerID == 0 {
		return fmt.Errorf("customer ID is required")
	}
	return nil
}

func (p *PaymentProxy) validateRechargeRequest(req *client.RechargeRequestHTTP) error {
	if req.CustomerID == 0 {
		return fmt.Errorf("customer ID is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if req.Channel == "" {
		return fmt.Errorf("payment channel is required")
	}

	// Validate payment channel is a valid channel
	validChannels := map[client.PaymentChannel]bool{
		client.PaymentChannelWallet: true,
		client.PaymentChannelWeChat: true,
		client.PaymentChannelAlipay: true,
		client.PaymentChannelStripe: true,
		client.PaymentChannelPaypal: true,
		client.PaymentChannelBank:   true,
		client.PaymentChannelOther:  true,
	}

	if !validChannels[req.Channel] {
		return fmt.Errorf("invalid payment channel: %s", req.Channel)
	}

	return nil
}

func (p *PaymentProxy) validateConsumeRequest(req *client.ConsumeRequestHTTP) error {
	if req.CustomerID == 0 {
		return fmt.Errorf("customer ID is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	return nil
}

func (p *PaymentProxy) validateRefundRequest(req *client.RefundRequestHTTP) error {
	if req.CustomerID == 0 {
		return fmt.Errorf("customer ID is required")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	return nil
}