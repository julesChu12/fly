package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// KratosClient Kratos订单服务客户端
type KratosClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logger.Logger
	config     *KratosClientConfig
}

// KratosClientConfig Kratos客户端配置
type KratosClientConfig struct {
	BaseURL        string        `yaml:"base_url"`
	Timeout        time.Duration `yaml:"timeout"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
	EnableCircuit  bool          `yaml:"enable_circuit"`
	CircuitTimeout time.Duration `yaml:"circuit_timeout"`
}

// OrderStatus 订单状态枚举
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"    // 待支付
	OrderStatusPaid      OrderStatus = "paid"       // 已支付
	OrderStatusCompleted OrderStatus = "completed"  // 已完成
	OrderStatusCancelled OrderStatus = "cancelled"  // 已取消
	OrderStatusExpired   OrderStatus = "expired"    // 已过期
)

// PaymentStatus 支付状态枚举
type PaymentStatus string

const (
	PaymentStatusPending        PaymentStatus = "pending"         // 待支付
	PaymentStatusProcessing     PaymentStatus = "processing"      // 处理中
	PaymentStatusPaid           PaymentStatus = "paid"            // 已支付
	PaymentStatusFailed         PaymentStatus = "failed"          // 支付失败
	PaymentStatusRefunded       PaymentStatus = "refunded"        // 已退款
	PaymentStatusPartialRefunded PaymentStatus = "partial_refunded" // 部分退款
)

// Order 订单结构
type Order struct {
	ID              string       `json:"id"`
	OrderNumber     string       `json:"order_number"`
	CustomerID      string       `json:"customer_id"`
	AppointmentID   string       `json:"appointment_id"`
	StaffID         string       `json:"staff_id"`
	ServiceID       string       `json:"service_id"`
	Amount          float64      `json:"amount"`
	Currency        string       `json:"currency"`
	Status          OrderStatus  `json:"status"`
	PaymentStatus   PaymentStatus `json:"payment_status"`
	Description     string       `json:"description"`
	OrderTime       time.Time    `json:"order_time"`
	PaymentDeadline time.Time    `json:"payment_deadline"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	CustomerID    string  `json:"customer_id" validate:"required"`
	AppointmentID string  `json:"appointment_id" validate:"required"`
	StaffID       string  `json:"staff_id" validate:"required"`
	ServiceID     string  `json:"service_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	Currency      string  `json:"currency"`
	Description   string  `json:"description"`
	OrderTime     time.Time `json:"order_time"`
}

// UpdateOrderStatusRequest 更新订单状态请求
type UpdateOrderStatusRequest struct {
	Status        string `json:"status" validate:"required"`
	PaymentStatus string `json:"payment_status"`
	Reason        string `json:"reason"`
}

// OrderResponse 订单响应
type OrderResponse struct {
	Success bool         `json:"success"`
	Data    *Order       `json:"data"`
	Error   *ErrorResponse `json:"error"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

// DefaultKratosClientConfig 默认配置
func DefaultKratosClientConfig() *KratosClientConfig {
	return &KratosClientConfig{
		BaseURL:        "http://localhost:8081", // Kratos服务地址
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		EnableCircuit:  true,
		CircuitTimeout: 60 * time.Second,
	}
}

// NewKratosClient 创建Kratos客户端
func NewKratosClient(config *KratosClientConfig, logger *logger.Logger) *KratosClient {
	if config == nil {
		config = DefaultKratosClientConfig()
	}

	return &KratosClient{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
		config: config,
	}
}

// CreateOrder 创建订单
func (c *KratosClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	c.logger.Debug("开始创建订单",
		map[string]interface{}{
			"customer_id":    req.CustomerID,
			"appointment_id": req.AppointmentID,
			"amount":         req.Amount,
		})

	var lastErr error

	// 重试机制
	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("重试创建订单",
				map[string]interface{}{
					"attempt": attempt + 1,
					"order_id": req.AppointmentID,
				})
			time.Sleep(c.config.RetryDelay)
		}

		order, err := c.createOrderOnce(ctx, req)
		if err == nil {
			c.logger.Info("订单创建成功",
				map[string]interface{}{
					"order_id": order.ID,
					"order_number": order.OrderNumber,
					"amount": order.Amount,
				})
			return order, nil
		}

		lastErr = err

		// 判断是否应该重试
		if !c.shouldRetry(err) {
			break
		}
	}

	return nil, fmt.Errorf("创建订单失败，已重试%d次: %w", c.config.MaxRetries, lastErr)
}

// createOrderOnce 单次创建订单
func (c *KratosClient) createOrderOnce(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	url := fmt.Sprintf("%s/api/v1/orders", c.baseURL)

	// 设置默认值
	if req.Currency == "" {
		req.Currency = "CNY"
	}
	if req.OrderTime.IsZero() {
		req.OrderTime = time.Now()
	}

	// 序列化请求
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 处理响应
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var response OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !response.Success || response.Data == nil {
		return nil, fmt.Errorf("订单创建失败: %s", response.Error.Message)
	}

	return response.Data, nil
}

// GetOrder 获取订单信息
func (c *KratosClient) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	url := fmt.Sprintf("%s/api/v1/orders/%s", c.baseURL, orderID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var response OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !response.Success || response.Data == nil {
		return nil, fmt.Errorf("获取订单失败: %s", response.Error.Message)
	}

	return response.Data, nil
}

// UpdateOrderStatus 更新订单状态
func (c *KratosClient) UpdateOrderStatus(ctx context.Context, orderID string, req *UpdateOrderStatusRequest) error {
	url := fmt.Sprintf("%s/api/v1/orders/%s/status", c.baseURL, orderID)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// CancelOrder 取消订单
func (c *KratosClient) CancelOrder(ctx context.Context, orderID, reason string) error {
	req := &UpdateOrderStatusRequest{
		Status: string(OrderStatusCancelled),
		Reason: reason,
	}

	return c.UpdateOrderStatus(ctx, orderID, req)
}

// shouldRetry 判断是否应该重试
func (c *KratosClient) shouldRetry(err error) bool {
	// HTTP错误码判断
	if httpErr, ok := err.(*httpError); ok {
		// 5xx服务器错误可以重试，4xx客户端错误（除429外）不重试
		return httpErr.StatusCode >= 500 || httpErr.StatusCode == http.StatusTooManyRequests
	}

	// 网络超时错误可以重试
	if isTimeoutError(err) {
		return true
	}

	return false
}

// httpError HTTP错误类型
type httpError struct {
	StatusCode int
	Message    string
}

func (e *httpError) Error() string {
	return e.Message
}

// handleErrorResponse 处理错误响应
func (c *KratosClient) handleErrorResponse(resp *http.Response) error {
	var errorResp ErrorResponse
	bodyBytes, _ := json.Marshal(resp.Body)

	if err := json.Unmarshal(bodyBytes, &errorResp); err != nil {
		return &httpError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: 无法解析错误响应", resp.StatusCode),
		}
	}

	return &httpError{
		StatusCode: resp.StatusCode,
		Message:    errorResp.Message,
	}
}

// isTimeoutError 判断是否为超时错误
func isTimeoutError(err error) bool {
	// 检查是否为context超时
	if err == context.DeadlineExceeded {
		return true
	}

	// 检查是否为网络超时
	errStr := err.Error()
	return contains(errStr, "timeout") || contains(errStr, "deadline")
}

// contains 字符串包含检查（简单实现，生产环境可使用strings.Contains）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		   (len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		   (len(s) > len(substr) && findSubstring(s, substr))
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 从预约请求创建订单请求的辅助函数
func CreateOrderRequestFromAppointment(appointment *entity.Appointment, amount float64) *CreateOrderRequest {
	return &CreateOrderRequest{
		CustomerID:    appointment.CustomerID.String(),
		AppointmentID: appointment.ID.String(),
		StaffID:       appointment.StaffID.String(),
		ServiceID:     appointment.ServiceID.String(),
		Amount:        amount,
		Currency:      "CNY",
		Description:   fmt.Sprintf("预约服务 - 员工:%s, 时间:%s",
			appointment.StaffID.String(),
			appointment.StartTime.Format("2006-01-02 15:04")),
		OrderTime: time.Now(),
	}
}