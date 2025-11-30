package client

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// PlutusClient Plutus支付服务客户端
type PlutusClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
	logger     *logger.Logger
	config     *PlutusClientConfig
}

// PaymentMethod 支付方式
type PaymentMethod string

const (
	PaymentMethodWeChatPay PaymentMethod = "wechat_pay"
	PaymentMethodAlipay    PaymentMethod = "alipay"
	PaymentMethodUnionPay  PaymentMethod = "union_pay"
	PaymentMethodBalance   PaymentMethod = "balance"
)

// Payment 支付记录结构
type Payment struct {
	ID            string        `json:"id"`
	OrderID       string        `json:"order_id"`
	AppointmentID string        `json:"appointment_id"`
	CustomerID    string        `json:"customer_id"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Status        PaymentStatus `json:"status"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	TransactionID string        `json:"transaction_id"`
	ExternalID    string        `json:"external_id,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderID       string        `json:"order_id" validate:"required"`
	AppointmentID string        `json:"appointment_id" validate:"required"`
	CustomerID    string        `json:"customer_id" validate:"required"`
	Amount        float64       `json:"amount" validate:"required,gt=0"`
	Currency      string        `json:"currency"`
	PaymentMethod PaymentMethod `json:"payment_method" validate:"required"`
	Description   string        `json:"description"`
	NotifyURL     string        `json:"notify_url,omitempty"`
	ReturnURL     string        `json:"return_url,omitempty"`
	ExpireTime    *time.Time    `json:"expire_time,omitempty"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	PaymentID  string  `json:"payment_id" validate:"required"`
	Amount     float64 `json:"amount" validate:"required,gt=0"`
	Reason     string  `json:"reason"`
	RefundID   string  `json:"refund_id,omitempty"`
	ExternalID string  `json:"external_id,omitempty"`
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	Success bool           `json:"success"`
	Data    *Payment       `json:"data"`
	Error   *ErrorResponse `json:"error"`
}

// PaymentStatusQuery 支付状态查询响应
type PaymentStatusQuery struct {
	PaymentID    string `json:"payment_id"`
	Status       string `json:"status"`
	Amount       string `json:"amount"`
	PaidAmount   string `json:"paid_amount"`
	RefundAmount string `json:"refund_amount"`
	CreatedAt    string `json:"created_at"`
	CompletedAt  string `json:"completed_at"`
}

// NewPlutusClient 创建Plutus客户端
func NewPlutusClient(config *PlutusClientConfig, logger *logger.Logger) *PlutusClient {
	if config == nil {
		config = DefaultPlutusClientConfig()
	}

	return &PlutusClient{
		baseURL:   config.BaseURL,
		apiKey:    config.APIKey,
		apiSecret: config.APISecret,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
		config: config,
	}
}

// CreatePayment 创建支付记录
func (c *PlutusClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error) {
	c.logger.Debug("开始创建支付记录",
		map[string]interface{}{
			"order_id":       req.OrderID,
			"appointment_id": req.AppointmentID,
			"amount":         req.Amount,
			"payment_method": req.PaymentMethod,
		})

	var lastErr error

	// 重试机制
	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("重试创建支付记录",
				map[string]interface{}{
					"attempt":  attempt + 1,
					"order_id": req.OrderID,
				})
			time.Sleep(c.config.RetryDelay)
		}

		payment, err := c.createPaymentOnce(ctx, req)
		if err == nil {
			c.logger.Info("支付记录创建成功",
				map[string]interface{}{
					"payment_id":     payment.ID,
					"order_id":       payment.OrderID,
					"transaction_id": payment.TransactionID,
					"amount":         payment.Amount,
				})
			return payment, nil
		}

		lastErr = err

		// 判断是否应该重试
		if !c.shouldRetry(err) {
			break
		}
	}

	return nil, fmt.Errorf("创建支付记录失败，已重试%d次: %w", c.config.MaxRetries, lastErr)
}

// createPaymentOnce 单次创建支付记录
func (c *PlutusClient) createPaymentOnce(ctx context.Context, req *CreatePaymentRequest) (*Payment, error) {
	url := fmt.Sprintf("%s/api/v1/payments", c.baseURL)

	// 设置默认值
	if req.Currency == "" {
		req.Currency = "CNY"
	}

	// 生成签名
	signature, timestamp, err := c.generateSignature(req)
	if err != nil {
		return nil, fmt.Errorf("生成签名失败: %w", err)
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

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("X-Timestamp", timestamp)
	httpReq.Header.Set("X-Signature", signature)

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

	var response PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !response.Success || response.Data == nil {
		return nil, fmt.Errorf("创建支付记录失败: %s", response.Error.Message)
	}

	return response.Data, nil
}

// GetPayment 获取支付信息
func (c *PlutusClient) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	url := fmt.Sprintf("%s/api/v1/payments/%s", c.baseURL, paymentID)

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := c.generateSimpleSignature("GET", url, timestamp)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("X-Timestamp", timestamp)
	httpReq.Header.Set("X-Signature", signature)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var response PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !response.Success || response.Data == nil {
		return nil, fmt.Errorf("获取支付信息失败: %s", response.Error.Message)
	}

	return response.Data, nil
}

// QueryPaymentStatus 查询支付状态
func (c *PlutusClient) QueryPaymentStatus(ctx context.Context, paymentID string) (*PaymentStatusQuery, error) {
	url := fmt.Sprintf("%s/api/v1/payments/%s/status", c.baseURL, paymentID)

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := c.generateSimpleSignature("GET", url, timestamp)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("X-Timestamp", timestamp)
	httpReq.Header.Set("X-Signature", signature)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var statusQuery PaymentStatusQuery
	if err := json.NewDecoder(resp.Body).Decode(&statusQuery); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &statusQuery, nil
}

// RefundPayment 退款
func (c *PlutusClient) RefundPayment(ctx context.Context, req *RefundRequest) error {
	url := fmt.Sprintf("%s/api/v1/payments/refund", c.baseURL)

	// 生成签名
	signature, timestamp, err := c.generateSignature(req)
	if err != nil {
		return fmt.Errorf("生成签名失败: %w", err)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("X-Timestamp", timestamp)
	httpReq.Header.Set("X-Signature", signature)

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

// generateSignature 生成请求签名
func (c *PlutusClient) generateSignature(req interface{}) (string, string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// 将请求转换为map并排序
	reqMap, err := structToMap(req)
	if err != nil {
		return "", "", err
	}

	// 添加时间戳
	reqMap["timestamp"] = timestamp

	// 按键名排序
	var keys []string
	for k := range reqMap {
		if k != "signature" { // 排除签名字段
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signParts []string
	for _, k := range keys {
		signParts = append(signParts, fmt.Sprintf("%s=%v", k, reqMap[k]))
	}
	signString := strings.Join(signParts, "&")

	// 生成HMAC-SHA256签名
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(signString))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature, timestamp, nil
}

// generateSimpleSignature 生成简单签名（用于GET请求）
func (c *PlutusClient) generateSimpleSignature(method, url, timestamp string) string {
	signString := fmt.Sprintf("%s\n%s\n%s", method, url, timestamp)

	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(signString))
	return hex.EncodeToString(h.Sum(nil))
}

// structToMap 将结构体转换为map
func structToMap(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// shouldRetry 判断是否应该重试
func (c *PlutusClient) shouldRetry(err error) bool {
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

// handleErrorResponse 处理错误响应
func (c *PlutusClient) handleErrorResponse(resp *http.Response) error {
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

// GetBaseURL 获取基础URL（用于健康检查）
func (c *PlutusClient) GetBaseURL() string {
	return c.baseURL
}

// CreatePaymentRequestFromOrder 从订单创建支付请求
func CreatePaymentRequestFromOrder(order *Order, paymentMethod PaymentMethod) *CreatePaymentRequest {
	return &CreatePaymentRequest{
		OrderID:       order.ID,
		AppointmentID: order.AppointmentID,
		CustomerID:    order.CustomerID,
		Amount:        order.Amount,
		Currency:      order.Currency,
		PaymentMethod: paymentMethod,
		Description:   fmt.Sprintf("预约服务订单 - %s", order.OrderNumber),
		ExpireTime:    &order.PaymentDeadline,
	}
}
