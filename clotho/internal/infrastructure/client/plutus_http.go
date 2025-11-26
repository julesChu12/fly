package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/julesChu12/fly/mora/pkg/discovery"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// PaymentHTTPClient represents an HTTP client for the Plutus (Payment/Wallet) service
type PaymentHTTPClient struct {
	baseURL    string
	client     *OptimizedHTTPClient
	httpClient *http.Client // Fallback for service discovery
	logger     *logger.Logger
	discovery  discovery.Discovery // 服务发现客户端
}

// Wallet represents a wallet from Plutus service
type Wallet struct {
	ID         uint          `json:"id"`
	TenantID   uint          `json:"tenant_id"`
	CustomerID uint          `json:"customer_id"`
	Balance    float64       `json:"balance"`
	Currency   string        `json:"currency"`
	Status     WalletStatus  `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// WalletStatus represents the status of a wallet
type WalletStatus string

const (
	WalletStatusActive  WalletStatus = "active"
	WalletStatusFrozen  WalletStatus = "frozen"
)

// Transaction represents a transaction from Plutus service
type Transaction struct {
	ID             uint              `json:"id"`
	TenantID       uint              `json:"tenant_id"`
	WalletID       uint              `json:"wallet_id"`
	OrderID        *uint             `json:"order_id"`
	Type           TransactionType   `json:"type"`
	Amount         float64           `json:"amount"`
	Currency       string            `json:"currency"`
	Channel        PaymentChannel    `json:"channel"`
	Status         TransactionStatus `json:"status"`
	IdempotencyKey *string           `json:"idempotency_key"`
	ReferenceNo    *string           `json:"reference_no"`
	Meta           string            `json:"meta"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// TransactionType represents the type of a transaction
type TransactionType string

const (
	TransactionTypeRecharge TransactionType = "recharge"
	TransactionTypeConsume  TransactionType = "consume"
	TransactionTypeRefund   TransactionType = "refund"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusSuccess TransactionStatus = "success"
	TransactionStatusFailed  TransactionStatus = "failed"
)

// PaymentChannel represents a payment channel
type PaymentChannel string

const (
	PaymentChannelWallet  PaymentChannel = "wallet"
	PaymentChannelWeChat  PaymentChannel = "wechat"
	PaymentChannelAlipay  PaymentChannel = "alipay"
	PaymentChannelStripe  PaymentChannel = "stripe"
	PaymentChannelPaypal  PaymentChannel = "paypal"
	PaymentChannelBank    PaymentChannel = "bank"
	PaymentChannelOther   PaymentChannel = "other"
)

// PaymentChannelConfig represents a payment channel configuration
type PaymentChannelConfig struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Config    string    `json:"config"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateWalletRequestHTTP represents a request to create a wallet via HTTP
type CreateWalletRequestHTTP struct {
	CustomerID uint   `json:"customer_id"`
	Currency   string `json:"currency"`
}

// RechargeRequestHTTP represents a request to recharge a wallet via HTTP
type RechargeRequestHTTP struct {
	CustomerID     uint                   `json:"customer_id"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	Channel        PaymentChannel        `json:"channel"`
	IdempotencyKey *string                `json:"idempotency_key"`
	ReferenceNo    *string                `json:"reference_no"`
	Meta           map[string]interface{} `json:"meta"`
}

// ConsumeRequestHTTP represents a request to consume from a wallet via HTTP
type ConsumeRequestHTTP struct {
	CustomerID     uint                   `json:"customer_id"`
	OrderID        *uint                  `json:"order_id"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	IdempotencyKey *string                `json:"idempotency_key"`
	Meta           map[string]interface{} `json:"meta"`
}

// RefundRequestHTTP represents a request to refund to a wallet via HTTP
type RefundRequestHTTP struct {
	CustomerID     uint                   `json:"customer_id"`
	OrderID        *uint                  `json:"order_id"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	IdempotencyKey *string                `json:"idempotency_key"`
	ReferenceNo    *string                `json:"reference_no"`
	Meta           map[string]interface{} `json:"meta"`
}

// WalletFilter represents filter parameters for wallet queries
type WalletFilter struct {
	CustomerID *uint `form:"customer_id" json:"customer_id"`
	Page       int   `form:"page" json:"page"`
	PageSize   int   `form:"page_size" json:"page_size"`
}

// TransactionFilter represents filter parameters for transaction queries
type TransactionFilter struct {
	WalletID *uint             `form:"wallet_id" json:"wallet_id"`
	OrderID  *uint             `form:"order_id" json:"order_id"`
	Type     *TransactionType  `form:"type" json:"type"`
	Status   *TransactionStatus `form:"status" json:"status"`
	Page     int               `form:"page" json:"page"`
	PageSize int               `form:"page_size" json:"page_size"`
}

// ListResponse represents the response for list queries
type ListResponse struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// NewPaymentHTTPClient creates a new Payment HTTP client with optimized settings
func NewPaymentHTTPClient(baseURL string, timeout time.Duration) *PaymentHTTPClient {
	log := logger.NewDefault()
	log.Info("Creating optimized Payment HTTP client", "base_url", baseURL)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient(baseURL, config, log)

	return &PaymentHTTPClient{
		baseURL:    baseURL,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
	}
}

// NewPaymentHTTPClientWithDiscovery uses service discovery to create a Payment HTTP client
func NewPaymentHTTPClientWithDiscovery(disc discovery.Discovery, timeout time.Duration) (*PaymentHTTPClient, error) {
	log := logger.NewDefault()

	// 从服务发现获取 plutus 服务地址
	instance, err := disc.GetService(context.Background(), "plutus")
	if err != nil {
		log.Error("Failed to discover Plutus service", "error", err.Error())
		return nil, fmt.Errorf("failed to discover Plutus service: %w", err)
	}

	address := instance.Address()
	log.Info("Discovered Plutus service", "address", address, "instance_id", instance.ID)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient("http://"+address, config, log)

	return &PaymentHTTPClient{
		baseURL:    "http://" + address,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
		discovery:  disc,
	}, nil
}

// ListWallets retrieves a list of wallets with filters
func (c *PaymentHTTPClient) ListWallets(ctx context.Context, filter *WalletFilter) (*ListResponse, error) {
	// 构建查询参数
	params := make(map[string]string)
	if filter.Page > 0 {
		params["page"] = fmt.Sprintf("%d", filter.Page)
	}
	if filter.PageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", filter.PageSize)
	}
	if filter.CustomerID != nil {
		params["customer_id"] = fmt.Sprintf("%d", *filter.CustomerID)
	}

	url := c.buildURL("/api/wallets", params)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    ListResponse `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to list wallets", "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("plutus service error: %s", response.Message)
	}

	return &response.Data, nil
}

// CreateWallet creates a new wallet
func (c *PaymentHTTPClient) CreateWallet(ctx context.Context, req *CreateWalletRequestHTTP) (*Wallet, error) {
	url := c.buildURL("/api/wallets", nil)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    Wallet `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to create wallet", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("plutus service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetWallet retrieves a wallet by ID
func (c *PaymentHTTPClient) GetWallet(ctx context.Context, walletID uint) (*Wallet, error) {
	url := c.buildURL(fmt.Sprintf("/api/wallets/%d", walletID), nil)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    Wallet `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get wallet", "wallet_id", walletID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("plutus service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetWalletByCustomerID retrieves a wallet by customer ID
func (c *PaymentHTTPClient) GetWalletByCustomerID(ctx context.Context, customerID uint) (*Wallet, error) {
	url := c.buildURL(fmt.Sprintf("/api/wallets/customer/%d", customerID), nil)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    Wallet `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get wallet by customer", "customer_id", customerID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("plutus service error: %s", response.Message)
	}

	return &response.Data, nil
}

// Recharge adds funds to a wallet
func (c *PaymentHTTPClient) Recharge(ctx context.Context, req *RechargeRequestHTTP) (*Transaction, error) {
	url := c.buildURL("/api/transactions/recharge", nil)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    Transaction `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to recharge wallet", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("plutus service error: %s", response.Message)
	}

	return &response.Data, nil
}

// Consume deducts funds from a wallet
func (c *PaymentHTTPClient) Consume(ctx context.Context, req *ConsumeRequestHTTP) (*Transaction, error) {
	url := c.buildURL("/api/transactions/consume", nil)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    Transaction `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to consume from wallet", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("plutus service error: %s", response.Message)
	}

	return &response.Data, nil
}

// Refund returns funds to a wallet
func (c *PaymentHTTPClient) Refund(ctx context.Context, req *RefundRequestHTTP) (*Transaction, error) {
	url := c.buildURL("/api/transactions/refund", nil)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    Transaction `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to refund to wallet", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("plutus service error: %s", response.Message)
	}

	return &response.Data, nil
}

// ListTransactions retrieves a list of transactions with filters
func (c *PaymentHTTPClient) ListTransactions(ctx context.Context, filter *TransactionFilter) (*ListResponse, error) {
	// 构建查询参数
	params := make(map[string]string)
	if filter.Page > 0 {
		params["page"] = fmt.Sprintf("%d", filter.Page)
	}
	if filter.PageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", filter.PageSize)
	}
	if filter.WalletID != nil {
		params["wallet_id"] = fmt.Sprintf("%d", *filter.WalletID)
	}
	if filter.OrderID != nil {
		params["order_id"] = fmt.Sprintf("%d", *filter.OrderID)
	}
	if filter.Type != nil {
		params["type"] = string(*filter.Type)
	}
	if filter.Status != nil {
		params["status"] = string(*filter.Status)
	}

	url := c.buildURL("/api/transactions", params)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    ListResponse `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to list transactions", "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("plutus service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetTransactionsByOrderID retrieves transactions for a specific order
func (c *PaymentHTTPClient) GetTransactionsByOrderID(ctx context.Context, orderID uint) (*ListResponse, error) {
	filter := &TransactionFilter{
		OrderID:  &orderID,
		Page:     1,
		PageSize: 50, // Default limit
	}
	return c.ListTransactions(ctx, filter)
}

// Helper functions

func (c *PaymentHTTPClient) buildURL(path string, params map[string]string) string {
	url := c.baseURL + path

	if len(params) > 0 {
		url += "?"
		first := true
		for key, value := range params {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", key, value)
			first = false
		}
	}

	return url
}

func (c *PaymentHTTPClient) doRequest(ctx context.Context, method, url string, body interface{}, response interface{}) error {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	c.logger.Debug("Making optimized HTTP request", "method", method, "url", url)

	var resp *http.Response
	var err error

	// Try to use optimized client first
	if c.client != nil {
		switch method {
		case "GET":
			resp, err = c.client.Get(ctx, url)
		case "POST":
			resp, err = c.client.Post(ctx, url, "application/json", reqBody)
		case "PUT":
			resp, err = c.client.Put(ctx, url, "application/json", reqBody)
		case "DELETE":
			resp, err = c.client.Delete(ctx, url)
		default:
			// Fallback to creating request manually for other methods
			req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err = c.client.Do(ctx, req)
		}
	} else {
		// Fallback to legacy HTTP client
		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err = c.httpClient.Do(req)
	}

	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response for debugging
	c.logger.Debug("HTTP response received",
		"method", method,
		"url", url,
		"status_code", resp.StatusCode,
		"response_size", len(respBody))

	// Allow certain error status codes for specific business logic
	if resp.StatusCode < 200 || resp.StatusCode >= 600 {
		return fmt.Errorf("HTTP error: %d, body: %s", resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(respBody))
	}

	return nil
}