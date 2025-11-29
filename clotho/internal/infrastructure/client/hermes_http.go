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

// CustomerHTTPClient represents an HTTP client for the Hermes (Customer) service
type CustomerHTTPClient struct {
	baseURL    string
	client     *OptimizedHTTPClient
	httpClient *http.Client // Fallback for service discovery
	logger     *logger.Logger
	discovery  discovery.Discovery // 服务发现客户端
}

// Customer represents a customer from Hermes service
type Customer struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Tags      string    `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Contacts  []Contact `json:"contacts,omitempty"`
}

// Contact represents a contact record
type Contact struct {
	ID         uint      `json:"id"`
	TenantID   uint      `json:"tenant_id"`
	CustomerID uint      `json:"customer_id"`
	Type       string    `json:"type"`
	Value      string    `json:"value"`
	IsPrimary  bool      `json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateCustomerRequestHTTP represents a request to create a customer via HTTP
type CreateCustomerRequestHTTP struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Tags  string `json:"tags"`
}

// UpdateCustomerRequestHTTP represents a request to update a customer via HTTP
type UpdateCustomerRequestHTTP struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Tags  string `json:"tags"`
}

// CreateContactRequestHTTP represents a request to create a contact via HTTP
type CreateContactRequestHTTP struct {
	TenantID   uint   `json:"tenant_id"`
	CustomerID uint   `json:"customer_id"`
	Type       string `json:"type"`
	Value      string `json:"value"`
	IsPrimary  bool   `json:"is_primary"`
}

// UpdateContactRequestHTTP represents a request to update a contact via HTTP
type UpdateContactRequestHTTP struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	IsPrimary *bool  `json:"is_primary"`
}

// CustomerFilter represents filter parameters for customer queries
type CustomerFilter struct {
	Search string `json:"search"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

// NewCustomerHTTPClient creates a new Customer HTTP client with optimized settings
func NewCustomerHTTPClient(baseURL string, timeout time.Duration) *CustomerHTTPClient {
	log := logger.NewDefault()
	log.Info("Creating optimized Customer HTTP client", "base_url", baseURL)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient(baseURL, config, log)

	return &CustomerHTTPClient{
		baseURL:    baseURL,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
	}
}

// NewCustomerHTTPClientWithDiscovery uses service discovery to create a Customer HTTP client
func NewCustomerHTTPClientWithDiscovery(disc discovery.Discovery, timeout time.Duration) (*CustomerHTTPClient, error) {
	log := logger.NewDefault()

	// 从服务发现获取 hermes 服务地址
	instance, err := disc.GetService(context.Background(), "hermes")
	if err != nil {
		log.Error("Failed to discover Hermes service", "error", err.Error())
		return nil, fmt.Errorf("failed to discover Hermes service: %w", err)
	}

	address := instance.Address()
	log.Info("Discovered Hermes service", "address", address, "instance_id", instance.ID)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient("http://"+address, config, log)

	return &CustomerHTTPClient{
		baseURL:    "http://" + address,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
		discovery: disc,
	}, nil
}

// ListCustomers retrieves a list of customers
func (c *CustomerHTTPClient) ListCustomers(ctx context.Context, filter *CustomerFilter) ([]Customer, int64, error) {
	// 构建查询参数
	params := make(map[string]string)
	if filter.Page > 0 {
		params["page"] = fmt.Sprintf("%d", filter.Page)
	}
	if filter.Limit > 0 {
		params["page_size"] = fmt.Sprintf("%d", filter.Limit)
	}
	if filter.Search != "" {
		params["search"] = filter.Search
	}

	url := c.buildURL("/api/customers", params)

	var response struct {
		Code    int `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Customers []Customer `json:"data"`
			Total     int64       `json:"total"`
			Page      int         `json:"page"`
			Size      int         `json:"size"`
		} `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to list customers", "error", err.Error())
		return nil, 0, err
	}

	if response.Code != 200 {
		return nil, 0, fmt.Errorf("hermes service error: %s", response.Message)
	}

	return response.Data.Customers, response.Data.Total, nil
}

// CreateCustomer creates a new customer
func (c *CustomerHTTPClient) CreateCustomer(ctx context.Context, req *CreateCustomerRequestHTTP) (*Customer, error) {
	url := c.buildURL("/api/customers", nil)

	var response struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    Customer `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to create customer", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("hermes service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetCustomer retrieves a customer by ID
func (c *CustomerHTTPClient) GetCustomer(ctx context.Context, customerID uint) (*Customer, error) {
	url := c.buildURL(fmt.Sprintf("/api/customers/%d", customerID), nil)

	var response struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    Customer `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get customer", "customer_id", customerID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("hermes service error: %s", response.Message)
	}

	return &response.Data, nil
}

// UpdateCustomer updates an existing customer
func (c *CustomerHTTPClient) UpdateCustomer(ctx context.Context, customerID uint, req *UpdateCustomerRequestHTTP) (*Customer, error) {
	url := c.buildURL(fmt.Sprintf("/api/customers/%d", customerID), nil)

	var response struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    Customer `json:"data"`
	}

	err := c.doRequest(ctx, "PUT", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to update customer", "customer_id", customerID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("hermes service error: %s", response.Message)
	}

	return &response.Data, nil
}

// DeleteCustomer deletes a customer
func (c *CustomerHTTPClient) DeleteCustomer(ctx context.Context, customerID uint) error {
	url := c.buildURL(fmt.Sprintf("/api/customers/%d", customerID), nil)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	err := c.doRequest(ctx, "DELETE", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to delete customer", "customer_id", customerID, "error", err.Error())
		return err
	}

	if response.Code != 200 {
		return fmt.Errorf("hermes service error: %s", response.Message)
	}

	return nil
}

// CreateContact creates a new contact
func (c *CustomerHTTPClient) CreateContact(ctx context.Context, req *CreateContactRequestHTTP) (*Contact, error) {
	url := c.buildURL("/api/contacts", nil)

	var response struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    Contact `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to create contact", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("hermes service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetContacts retrieves contacts for a customer
func (c *CustomerHTTPClient) GetContacts(ctx context.Context, customerID uint, page, pageSize int) ([]Contact, error) {
	params := make(map[string]string)
	if page > 0 {
		params["page"] = fmt.Sprintf("%d", page)
	}
	if pageSize > 0 {
		params["page_size"] = fmt.Sprintf("%d", pageSize)
	}

	url := c.buildURL(fmt.Sprintf("/api/customers/%d/contacts", customerID), params)

	var response struct {
		Code    int `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Contacts []Contact `json:"data"`
			Page     int         `json:"page"`
			Size     int         `json:"size"`
		} `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get contacts", "customer_id", customerID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("hermes service error: %s", response.Message)
	}

	return response.Data.Contacts, nil
}

// UpdateContact updates an existing contact
func (c *CustomerHTTPClient) UpdateContact(ctx context.Context, contactID uint, req *UpdateContactRequestHTTP) (*Contact, error) {
	url := c.buildURL(fmt.Sprintf("/api/contacts/%d", contactID), nil)

	var response struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    Contact `json:"data"`
	}

	err := c.doRequest(ctx, "PUT", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to update contact", "contact_id", contactID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("hermes service error: %s", response.Message)
	}

	return &response.Data, nil
}

// DeleteContact deletes a contact
func (c *CustomerHTTPClient) DeleteContact(ctx context.Context, contactID uint) error {
	url := c.buildURL(fmt.Sprintf("/api/contacts/%d", contactID), nil)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	err := c.doRequest(ctx, "DELETE", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to delete contact", "contact_id", contactID, "error", err.Error())
		return err
	}

	if response.Code != 200 {
		return fmt.Errorf("hermes service error: %s", response.Message)
	}

	return nil
}

// SearchCustomers searches for customers by name, phone, or email
func (c *CustomerHTTPClient) SearchCustomers(ctx context.Context, searchTerm string, page, limit int) ([]Customer, int64, error) {
	filter := &CustomerFilter{
		Search: searchTerm,
		Page:   page,
		Limit:  limit,
	}

	return c.ListCustomers(ctx, filter)
}

// Helper functions

func (c *CustomerHTTPClient) buildURL(path string, params map[string]string) string {
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

func (c *CustomerHTTPClient) doRequest(ctx context.Context, method, url string, body interface{}, response interface{}) error {
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

// BatchDeleteCustomers batch deletes customers by IDs
func (c *CustomerHTTPClient) BatchDeleteCustomers(ctx context.Context, customerIDs []uint) error {
	url := c.buildURL("/api/customers/batch", nil)

	requestBody := map[string][]uint{
		"ids": customerIDs,
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	err := c.doRequest(ctx, "DELETE", url, requestBody, &response)
	if err != nil {
		c.logger.Error("Failed to batch delete customers", "count", len(customerIDs), "error", err.Error())
		return err
	}

	if response.Code != 200 {
		return fmt.Errorf("hermes service error: %s", response.Message)
	}

	c.logger.Info("Successfully batch deleted customers", "count", len(customerIDs))
	return nil
}
// GetCustomerStats retrieves customer statistics from the service
func (c *CustomerHTTPClient) GetCustomerStats(ctx context.Context) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/customers/stats", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get customer stats failed: status %d", resp.StatusCode)
	}

	var response struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Code != http.StatusOK {
		return nil, fmt.Errorf("get customer stats failed: %s", response.Message)
	}

	return response.Data, nil
}
