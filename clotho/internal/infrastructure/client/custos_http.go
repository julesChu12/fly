package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// CustosHTTPClient 封装对 Custos HTTP API 的调用
type CustosHTTPClient struct {
	client  *http.Client
	baseURL string
	logger  *logger.Logger
}

// NewCustosHTTPClient 创建新的 Custos HTTP 客户端
func NewCustosHTTPClient(baseURL string, timeout time.Duration, logger *logger.Logger) *CustosHTTPClient {
	return &CustosHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		logger:  logger,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CustosLoginResponse Custos服务返回的原始响应结构
type CustosLoginResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    LoginData   `json:"data"`
}

// LoginData 登录数据
type LoginData struct {
	AccessToken          string      `json:"access_token"`
	TokenType            string      `json:"token_type"`
	ExpiresIn            int         `json:"expires_in"`
	RefreshToken         string      `json:"refresh_token"`
	RefreshExpiresIn     int         `json:"refresh_expires_in"`
	SessionID            string      `json:"session_id"`
	User                 HTTPUserInfo `json:"user"`
}

// LoginResponse 登录响应 - 兼容原有接口
type LoginResponse struct {
	AccessToken          string      `json:"access_token"`
	TokenType            string      `json:"token_type"`
	ExpiresIn            int         `json:"expires_in"`
	RefreshToken         string      `json:"refresh_token"`
	RefreshExpiresIn     int         `json:"refresh_expires_in"`
	SessionID            string      `json:"session_id"`
	User                 HTTPUserInfo `json:"user"`
}

// HTTPUserInfo HTTP API 用户信息
type HTTPUserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User HTTPUserInfo `json:"user"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	SessionID    string `json:"session_id"`
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// Login 用户登录
func (c *CustosHTTPClient) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 先解析为 Custos 原始响应格式
	custosResp := &CustosLoginResponse{}
	err := c.postRequest(ctx, "/api/v1/auth/login", req, custosResp)
	if err != nil {
		return nil, err
	}

	// 转换为标准 LoginResponse 格式
	resp := &LoginResponse{
		AccessToken:      custosResp.Data.AccessToken,
		TokenType:        custosResp.Data.TokenType,
		ExpiresIn:        custosResp.Data.ExpiresIn,
		RefreshToken:     custosResp.Data.RefreshToken,
		RefreshExpiresIn: custosResp.Data.RefreshExpiresIn,
		SessionID:        custosResp.Data.SessionID,
		User:             custosResp.Data.User,
	}

	return resp, nil
}

// Register 用户注册
func (c *CustosHTTPClient) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	resp := &RegisterResponse{}
	err := c.postRequest(ctx, "/api/v1/auth/register", req, resp)
	return resp, err
}

// RefreshToken 刷新令牌
func (c *CustosHTTPClient) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*TokenResponse, error) {
	resp := &TokenResponse{}
	err := c.postRequest(ctx, "/api/v1/auth/refresh", req, resp)
	return resp, err
}

// Logout 用户登出
func (c *CustosHTTPClient) Logout(ctx context.Context, sessionID, accessToken string) error {
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + accessToken
	headers["X-Session-ID"] = sessionID

	return c.deleteRequest(ctx, "/api/v1/auth/logout", headers)
}

// LogoutAll 登出所有设备
func (c *CustosHTTPClient) LogoutAll(ctx context.Context, accessToken string) error {
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + accessToken

	return c.deleteRequest(ctx, "/api/v1/auth/logout-all", headers)
}

// ForgotPassword 忘记密码
func (c *CustosHTTPClient) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	return c.postRequest(ctx, "/api/v1/auth/forgot-password", req, nil)
}

// ResetPassword 重置密码
func (c *CustosHTTPClient) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	return c.postRequest(ctx, "/api/v1/auth/reset-password", req, nil)
}

// postRequest 发送 POST 请求
func (c *CustosHTTPClient) postRequest(ctx context.Context, path string, reqBody interface{}, respBody interface{}) error {
	// 序列化请求体
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("Failed to marshal request body", "error", err)
		return fmt.Errorf("marshal request: %w", err)
	}

	// 创建请求
	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Error("Failed to create request", "error", err)
		return fmt.Errorf("create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	c.logger.Info("Sending HTTP request", "method", "POST", "url", fullURL)
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Failed to send request", "error", err)
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", "error", err)
		return fmt.Errorf("read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Error("HTTP request failed", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 如果需要解析响应体
	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			c.logger.Error("Failed to unmarshal response", "error", err, "body", string(body))
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	c.logger.Info("HTTP request successful", "status", resp.StatusCode)
	return nil
}

// deleteRequest 发送 DELETE 请求
func (c *CustosHTTPClient) deleteRequest(ctx context.Context, path string, headers map[string]string) error {
	// 创建请求
	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "DELETE", fullURL, nil)
	if err != nil {
		c.logger.Error("Failed to create request", "error", err)
		return fmt.Errorf("create request: %w", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	c.logger.Info("Sending HTTP request", "method", "DELETE", "url", fullURL)
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Failed to send request", "error", err)
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("HTTP request failed", "status", resp.StatusCode, "body", string(body))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Info("HTTP request successful", "status", resp.StatusCode)
	return nil
}