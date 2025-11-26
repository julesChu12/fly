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

// StaffHTTPClient represents an HTTP client for the Staff service
type StaffHTTPClient struct {
	baseURL    string
	client     *OptimizedHTTPClient
	httpClient *http.Client // Fallback for service discovery
	logger     *logger.Logger
	discovery  discovery.Discovery // 服务发现客户端
}

// Staff represents a staff member from Staff service
type Staff struct {
	ID          string     `json:"id"`
	UserID      *string    `json:"user_id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Gender      string     `json:"gender"`
	Birthday    *string    `json:"birthday"`
	Avatar      *string    `json:"avatar"`
	Department  string     `json:"department"`
	Position    string     `json:"position"`
	RoleID      string     `json:"role_id"`
	Status      string     `json:"status"`
	HireDate    *string    `json:"hire_date"`
	Salary      *float64   `json:"salary"`
	Address     *string    `json:"address"`
	EmergencyContact *string `json:"emergency_contact"`
	Notes       *string    `json:"notes"`
	Skills      []string   `json:"skills"`
	WorkingHours *interface{} `json:"working_hours"`
	IsAvailable bool       `json:"is_available"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// StaffRole represents a staff role
type StaffRole struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description *string   `json:"description"`
	Permissions []string  `json:"permissions"`
	IsDefault   bool      `json:"is_default"`
	SortOrder   int       `json:"sort_order"`
	Status      string    `json:"status"`
	StaffCount  int       `json:"staff_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StaffAvailability represents staff availability
type StaffAvailability struct {
	ID          string    `json:"id"`
	StaffID     string    `json:"staff_id"`
	DayOfWeek   int       `json:"day_of_week"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
	IsAvailable bool      `json:"is_available"`
	Notes       *string   `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateStaffRequestHTTP represents a request to create staff via HTTP
type CreateStaffRequestHTTP struct {
	Name             string      `json:"name"`
	Email            string      `json:"email"`
	Phone            string      `json:"phone"`
	Gender           string      `json:"gender"`
	Birthday         *string     `json:"birthday"`
	Avatar           *string     `json:"avatar"`
	Department       string      `json:"department"`
	Position         string      `json:"position"`
	RoleID           string      `json:"role_id"`
	Status           string      `json:"status"`
	HireDate         *string     `json:"hire_date"`
	Salary           *float64    `json:"salary"`
	Address          *string     `json:"address"`
	EmergencyContact *string     `json:"emergency_contact"`
	Notes            *string     `json:"notes"`
	Skills           []string    `json:"skills"`
	WorkingHours     *interface{} `json:"working_hours"`
	IsAvailable      bool        `json:"is_available"`
}

// UpdateStaffRequestHTTP represents a request to update staff via HTTP
type UpdateStaffRequestHTTP struct {
	Name             *string     `json:"name"`
	Email            *string     `json:"email"`
	Phone            *string     `json:"phone"`
	Gender           *string     `json:"gender"`
	Birthday         *string     `json:"birthday"`
	Avatar           *string     `json:"avatar"`
	Department       *string     `json:"department"`
	Position         *string     `json:"position"`
	RoleID           *string     `json:"role_id"`
	Status           *string     `json:"status"`
	HireDate         *string     `json:"hire_date"`
	Salary           *float64    `json:"salary"`
	Address          *string     `json:"address"`
	EmergencyContact *string     `json:"emergency_contact"`
	Notes            *string     `json:"notes"`
	Skills           []string    `json:"skills"`
	WorkingHours     *interface{} `json:"working_hours"`
	IsAvailable      *bool       `json:"is_available"`
}

// UpdateStatusRequestHTTP represents a request to update staff status via HTTP
type UpdateStatusRequestHTTP struct {
	Status string  `json:"status"`
	Reason *string `json:"reason"`
}

// AvailabilityRequest represents a request to set availability
type AvailabilityRequest struct {
	StaffID       string                 `json:"staff_id"`
	Availabilities []AvailabilityItem      `json:"availabilities"`
}

// AvailabilityItem represents an availability item
type AvailabilityItem struct {
	DayOfWeek   int     `json:"day_of_week"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	IsAvailable bool    `json:"is_available"`
	Notes       *string `json:"notes"`
}

// StaffFilter represents filter parameters for staff queries
type StaffFilter struct {
	Search      *string `json:"search"`
	Department  *string `json:"department"`
	RoleID      *string `json:"role_id"`
	Status      *string `json:"status"`
	IsAvailable *bool   `json:"is_available"`
	MinAge      *int    `json:"min_age"`
	MaxAge      *int    `json:"max_age"`
	Gender      *string `json:"gender"`
	Skills      []string `json:"skills"`
	Page        int     `json:"page"`
	Limit       int     `json:"limit"`
	Sort        string  `json:"sort"`
	Order       string  `json:"order"`
}

// NewStaffHTTPClient creates a new Staff HTTP client with optimized settings
func NewStaffHTTPClient(baseURL string, timeout time.Duration) *StaffHTTPClient {
	log := logger.NewDefault()
	log.Info("Creating optimized Staff HTTP client", "base_url", baseURL)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient(baseURL, config, log)

	return &StaffHTTPClient{
		baseURL:    baseURL,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
	}
}

// NewStaffHTTPClientWithDiscovery uses service discovery to create a Staff HTTP client
func NewStaffHTTPClientWithDiscovery(disc discovery.Discovery, timeout time.Duration) (*StaffHTTPClient, error) {
	log := logger.NewDefault()

	// 从服务发现获取 staff 服务地址
	instance, err := disc.GetService(context.Background(), "staff")
	if err != nil {
		log.Error("Failed to discover Staff service", "error", err.Error())
		return nil, fmt.Errorf("failed to discover Staff service: %w", err)
	}

	address := instance.Address()
	log.Info("Discovered Staff service", "address", address, "instance_id", instance.ID)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient("http://"+address, config, log)

	return &StaffHTTPClient{
		baseURL:    "http://" + address,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
		discovery: disc,
	}, nil
}

// ListStaff retrieves a list of staff members
func (c *StaffHTTPClient) ListStaff(ctx context.Context, filter *StaffFilter) ([]Staff, int64, error) {
	// 构建查询参数
	params := make(map[string]string)
	if filter.Page > 0 {
		params["page"] = fmt.Sprintf("%d", filter.Page)
	}
	if filter.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", filter.Limit)
	}
	if filter.Search != nil {
		params["search"] = *filter.Search
	}
	if filter.Department != nil {
		params["department"] = *filter.Department
	}
	if filter.RoleID != nil {
		params["role_id"] = *filter.RoleID
	}
	if filter.Status != nil {
		params["status"] = *filter.Status
	}
	if filter.IsAvailable != nil {
		params["is_available"] = fmt.Sprintf("%t", *filter.IsAvailable)
	}
	if filter.Sort != "" {
		params["sort"] = filter.Sort
	}
	if filter.Order != "" {
		params["order"] = filter.Order
	}

	url := c.buildURL("/api/v1/staff", params)

	var response struct {
		Code    int `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Staff []Staff `json:"staff"`
			Total int64   `json:"total"`
			Page  int     `json:"page"`
			Limit int     `json:"limit"`
		} `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to list staff", "error", err.Error())
		return nil, 0, err
	}

	if response.Code != 200 {
		return nil, 0, fmt.Errorf("staff service error: %s", response.Message)
	}

	return response.Data.Staff, response.Data.Total, nil
}

// CreateStaff creates a new staff member
func (c *StaffHTTPClient) CreateStaff(ctx context.Context, req *CreateStaffRequestHTTP) (*Staff, error) {
	url := c.buildURL("/api/v1/staff", nil)

	var response struct {
		Code    int   `json:"code"`
		Message string `json:"message"`
		Data    Staff `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to create staff", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("staff service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetStaff retrieves a staff member by ID
func (c *StaffHTTPClient) GetStaff(ctx context.Context, staffID string) (*Staff, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/staff/%s", staffID), nil)

	var response struct {
		Code    int   `json:"code"`
		Message string `json:"message"`
		Data    Staff `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get staff", "staff_id", staffID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("staff service error: %s", response.Message)
	}

	return &response.Data, nil
}

// UpdateStaff updates a staff member
func (c *StaffHTTPClient) UpdateStaff(ctx context.Context, staffID string, req *UpdateStaffRequestHTTP) (*Staff, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/staff/%s", staffID), nil)

	var response struct {
		Code    int   `json:"code"`
		Message string `json:"message"`
		Data    Staff `json:"data"`
	}

	err := c.doRequest(ctx, "PUT", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to update staff", "staff_id", staffID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("staff service error: %s", response.Message)
	}

	return &response.Data, nil
}

// DeleteStaff deletes a staff member
func (c *StaffHTTPClient) DeleteStaff(ctx context.Context, staffID string) error {
	url := c.buildURL(fmt.Sprintf("/api/v1/staff/%s", staffID), nil)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	err := c.doRequest(ctx, "DELETE", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to delete staff", "staff_id", staffID, "error", err.Error())
		return err
	}

	if response.Code != 200 {
		return fmt.Errorf("staff service error: %s", response.Message)
	}

	return nil
}

// UpdateStaffStatus updates staff status
func (c *StaffHTTPClient) UpdateStaffStatus(ctx context.Context, staffID string, req *UpdateStatusRequestHTTP) (*Staff, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/staff/%s/status", staffID), nil)

	var response struct {
		Code    int   `json:"code"`
		Message string `json:"message"`
		Data    Staff `json:"data"`
	}

	err := c.doRequest(ctx, "PUT", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to update staff status", "staff_id", staffID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("staff service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetAvailableStaff retrieves available staff members
func (c *StaffHTTPClient) GetAvailableStaff(ctx context.Context, filter *StaffFilter) ([]Staff, error) {
	// 构建查询参数
	params := make(map[string]string)
	if filter.Department != nil {
		params["department"] = *filter.Department
	}
	if len(filter.Skills) > 0 {
		// 注意：这里需要处理技能数组
		params["skills"] = filter.Skills[0] // 简化实现，只传递第一个技能
	}

	available := true
	params["is_available"] = fmt.Sprintf("%t", available)

	url := c.buildURL("/api/v1/staff/available", params)

	var response struct {
		Code    int     `json:"code"`
		Message string  `json:"message"`
		Data    []Staff `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get available staff", "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("staff service error: %s", response.Message)
	}

	return response.Data, nil
}

// ListRoles retrieves a list of staff roles
func (c *StaffHTTPClient) ListRoles(ctx context.Context) ([]StaffRole, error) {
	url := c.buildURL("/api/v1/staff/roles", nil)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    []StaffRole `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to list roles", "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("staff service error: %s", response.Message)
	}

	return response.Data, nil
}

// CreateRole creates a new staff role
func (c *StaffHTTPClient) CreateRole(ctx context.Context, req map[string]interface{}) (*StaffRole, error) {
	url := c.buildURL("/api/v1/staff/roles", nil)

	var response struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    StaffRole `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to create role", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("staff service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetAvailability retrieves staff availability
func (c *StaffHTTPClient) GetAvailability(ctx context.Context, staffID string) ([]StaffAvailability, error) {
	url := c.buildURL(fmt.Sprintf("/api/v1/staff/availability/%s", staffID), nil)

	var response struct {
		Code    int                   `json:"code"`
		Message string                `json:"message"`
		Data    []StaffAvailability    `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get availability", "staff_id", staffID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("staff service error: %s", response.Message)
	}

	return response.Data, nil
}

// SetAvailability sets staff availability
func (c *StaffHTTPClient) SetAvailability(ctx context.Context, staffID string, req *AvailabilityRequest) error {
	url := c.buildURL(fmt.Sprintf("/api/v1/staff/availability/%s", staffID), nil)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	err := c.doRequest(ctx, "PUT", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to set availability", "staff_id", staffID, "error", err.Error())
		return err
	}

	if response.Code != 200 {
		return fmt.Errorf("staff service error: %s", response.Message)
	}

	return nil
}

// Helper functions

func (c *StaffHTTPClient) buildURL(path string, params map[string]string) string {
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

func (c *StaffHTTPClient) doRequest(ctx context.Context, method, url string, body interface{}, response interface{}) error {
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