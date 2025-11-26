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

// OrderHTTPClient represents an HTTP client for the Kratos (Order) service
type OrderHTTPClient struct {
	baseURL    string
	client     *OptimizedHTTPClient
	httpClient *http.Client // Fallback for service discovery
	logger     *logger.Logger
	discovery  discovery.Discovery // 服务发现客户端
}

// Order represents an order from Kratos service
type Order struct {
	ID          uint                  `json:"id"`
	TenantID    uint                  `json:"tenant_id"`
	OrderNo     string                `json:"order_no"`
	CustomerID  uint                  `json:"customer_id"`
	TotalAmount float64               `json:"total_amount"`
	Currency    string                `json:"currency"`
	Status      OrderStatus           `json:"status"`
	Remark      string                `json:"remark"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Items       []OrderItem           `json:"items,omitempty"`
	StatusLogs  []OrderStatusLog      `json:"status_logs,omitempty"`
	Audits      []OrderAudit          `json:"audits,omitempty"`
}

// OrderItem represents an order item
type OrderItem struct {
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

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// OrderStatusLog represents an order status change log
type OrderStatusLog struct {
	ID         uint        `json:"id"`
	TenantID   uint        `json:"tenant_id"`
	OrderID    uint        `json:"order_id"`
	FromStatus *OrderStatus `json:"from_status"`
	ToStatus   OrderStatus `json:"to_status"`
	Reason     string      `json:"reason"`
	OperatorID *uint       `json:"operator_id"`
	CreatedAt  time.Time   `json:"created_at"`
}

// OrderAudit represents an order audit log
type OrderAudit struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	OrderID   uint      `json:"order_id"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateOrderRequestHTTP represents a request to create an order via HTTP
type CreateOrderRequestHTTP struct {
	OrderNo     string                        `json:"order_no"`
	CustomerID  uint                          `json:"customer_id"`
	TotalAmount float64                       `json:"total_amount"`
	Currency    string                        `json:"currency"`
	Remark      string                        `json:"remark"`
	Items       []CreateOrderItemRequestHTTP   `json:"items"`
}

// CreateOrderItemRequestHTTP represents a request to create an order item via HTTP
type CreateOrderItemRequestHTTP struct {
	ProductID   *uint   `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

// UpdateOrderStatusRequestHTTP represents a request to update order status via HTTP
type UpdateOrderStatusRequestHTTP struct {
	Status     OrderStatus `json:"status"`
	Reason     string      `json:"reason"`
	OperatorID *uint       `json:"operator_id"`
}

// OrderFilter represents filter parameters for order queries
type OrderFilter struct {
	CustomerID *uint        `form:"customer_id" json:"customer_id"`
	Status     *OrderStatus `form:"status" json:"status"`
	Page       int          `form:"page" json:"page"`
	PageSize   int          `form:"page_size" json:"page_size"`
}

// OrderListResponse represents the response for order list queries
type OrderListResponse struct {
	Data  []Order `json:"data"`
	Total int64   `json:"total"`
	Page  int     `json:"page"`
	Size  int     `json:"size"`
}

// NewOrderHTTPClient creates a new Order HTTP client with optimized settings
func NewOrderHTTPClient(baseURL string, timeout time.Duration) *OrderHTTPClient {
	log := logger.NewDefault()
	log.Info("Creating optimized Order HTTP client", "base_url", baseURL)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient(baseURL, config, log)

	return &OrderHTTPClient{
		baseURL:    baseURL,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
	}
}

// NewOrderHTTPClientWithDiscovery uses service discovery to create an Order HTTP client
func NewOrderHTTPClientWithDiscovery(disc discovery.Discovery, timeout time.Duration) (*OrderHTTPClient, error) {
	log := logger.NewDefault()

	// 从服务发现获取 kratos 服务地址
	instance, err := disc.GetService(context.Background(), "kratos")
	if err != nil {
		log.Error("Failed to discover Kratos service", "error", err.Error())
		return nil, fmt.Errorf("failed to discover Kratos service: %w", err)
	}

	address := instance.Address()
	log.Info("Discovered Kratos service", "address", address, "instance_id", instance.ID)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient("http://"+address, config, log)

	return &OrderHTTPClient{
		baseURL:    "http://" + address,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
		discovery: disc,
	}, nil
}

// ListOrders retrieves a list of orders with filters
func (c *OrderHTTPClient) ListOrders(ctx context.Context, filter *OrderFilter) (*OrderListResponse, error) {
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
	if filter.Status != nil {
		params["status"] = string(*filter.Status)
	}

	url := c.buildURL("/api/orders", params)

	var response struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Data    OrderListResponse `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to list orders", "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("kratos service error: %s", response.Message)
	}

	return &response.Data, nil
}

// CreateOrder creates a new order
func (c *OrderHTTPClient) CreateOrder(ctx context.Context, req *CreateOrderRequestHTTP) (*Order, error) {
	url := c.buildURL("/api/orders", nil)

	var response struct {
		Code    int   `json:"code"`
		Message string `json:"message"`
		Data    Order `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to create order", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("kratos service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetOrder retrieves an order by ID
func (c *OrderHTTPClient) GetOrder(ctx context.Context, orderID uint) (*Order, error) {
	url := c.buildURL(fmt.Sprintf("/api/orders/%d", orderID), nil)

	var response struct {
		Code    int   `json:"code"`
		Message string `json:"message"`
		Data    Order `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get order", "order_id", orderID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("kratos service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetOrderWithItems retrieves an order with all items by ID
func (c *OrderHTTPClient) GetOrderWithItems(ctx context.Context, orderID uint) (*Order, error) {
	url := c.buildURL(fmt.Sprintf("/api/orders/%d/items", orderID), nil)

	var response struct {
		Code    int   `json:"code"`
		Message string `json:"message"`
		Data    Order `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get order with items", "order_id", orderID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("kratos service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetOrderLogs retrieves status change logs for an order
func (c *OrderHTTPClient) GetOrderLogs(ctx context.Context, orderID uint) ([]OrderStatusLog, error) {
	url := c.buildURL(fmt.Sprintf("/api/orders/%d/logs", orderID), nil)

	var response struct {
		Code    int             `json:"code"`
		Message string         `json:"message"`
		Data    []OrderStatusLog `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get order logs", "order_id", orderID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("kratos service error: %s", response.Message)
	}

	return response.Data, nil
}

// UpdateOrderStatus updates the status of an existing order
func (c *OrderHTTPClient) UpdateOrderStatus(ctx context.Context, orderID uint, req *UpdateOrderStatusRequestHTTP) (*Order, error) {
	url := c.buildURL(fmt.Sprintf("/api/orders/%d/status", orderID), nil)

	var response struct {
		Code    int   `json:"code"`
		Message string `json:"message"`
		Data    Order `json:"data"`
	}

	err := c.doRequest(ctx, "PATCH", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to update order status", "order_id", orderID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("kratos service error: %s", response.Message)
	}

	return &response.Data, nil
}

// DeleteOrder deletes an order
func (c *OrderHTTPClient) DeleteOrder(ctx context.Context, orderID uint) error {
	url := c.buildURL(fmt.Sprintf("/api/orders/%d", orderID), nil)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	err := c.doRequest(ctx, "DELETE", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to delete order", "order_id", orderID, "error", err.Error())
		return err
	}

	if response.Code != 200 {
		return fmt.Errorf("kratos service error: %s", response.Message)
	}

	return nil
}

// Helper functions

func (c *OrderHTTPClient) buildURL(path string, params map[string]string) string {
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

func (c *OrderHTTPClient) doRequest(ctx context.Context, method, url string, body interface{}, response interface{}) error {
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