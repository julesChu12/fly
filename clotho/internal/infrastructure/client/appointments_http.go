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

// AppointmentHTTPClient represents an HTTP client for the Appointments service
type AppointmentHTTPClient struct {
	baseURL    string
	client     *OptimizedHTTPClient
	httpClient *http.Client // Fallback for service discovery
	logger     *logger.Logger
	discovery  discovery.Discovery // 服务发现客户端
}

// AppointmentStatus represents appointment status
type AppointmentStatus string

const (
	AppointmentStatusPending     AppointmentStatus = "pending"
	AppointmentStatusConfirmed   AppointmentStatus = "confirmed"
	AppointmentStatusInProgress AppointmentStatus = "in_progress"
	AppointmentStatusCompleted   AppointmentStatus = "completed"
	AppointmentStatusCancelled   AppointmentStatus = "cancelled"
	AppointmentStatusNoShow      AppointmentStatus = "no_show"
)

// Appointment represents an appointment from Appointments service
type Appointment struct {
	ID            string          `json:"id"`
	CustomerID    string          `json:"customer_id"`
	StaffID    string          `json:"staff_id"`
	ServiceID     string          `json:"service_id"`
	StartTime     time.Time       `json:"start_time"`
	EndTime       time.Time       `json:"end_time"`
	Notes         *string         `json:"notes"`
	Status        AppointmentStatus `json:"status"`
	Reminder      bool            `json:"reminder"`
	ReminderTime  *time.Time      `json:"reminder_time"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// CreateAppointmentRequestHTTP represents a request to create an appointment via HTTP
type CreateAppointmentRequestHTTP struct {
	CustomerID    string     `json:"customer_id"`
	StaffID    string     `json:"staff_id"`
	ServiceID     string     `json:"service_id"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       time.Time  `json:"end_time"`
	Notes         *string    `json:"notes"`
	Reminder      bool       `json:"reminder"`
	ReminderTime  *time.Time `json:"reminder_time"`
}

// UpdateAppointmentRequestHTTP represents a request to update an appointment via HTTP
type UpdateAppointmentRequestHTTP struct {
	CustomerID     *string    `json:"customer_id"`
	StaffID     *string    `json:"staff_id"`
	ServiceID      *string    `json:"service_id"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Notes          *string    `json:"notes"`
	Status         *string    `json:"status"`
	Reminder       *bool      `json:"reminder"`
	ReminderTime   *time.Time `json:"reminder_time"`
}

// UpdateAppointmentStatusRequestHTTP represents a request to update appointment status via HTTP
type UpdateAppointmentStatusRequestHTTP struct {
	Status          string  `json:"status"`
	CompletionNotes *string `json:"completion_notes"`
}

// AppointmentFilter represents filter parameters for appointment queries
type AppointmentFilter struct {
	CustomerID  *string           `json:"customer_id"`
	StaffID  *string           `json:"staff_id"`
	ServiceID   *string           `json:"service_id"`
	Status      *AppointmentStatus `json:"status"`
	StartDate   *time.Time        `json:"start_date"`
	EndDate     *time.Time        `json:"end_date"`
	Date        *time.Time        `json:"date"`
	MinDuration *time.Duration    `json:"min_duration"`
	MaxDuration *time.Duration    `json:"max_duration"`
	Reminder    *bool             `json:"reminder"`
	Page        int               `json:"page"`
	Limit       int               `json:"limit"`
	Sort        string            `json:"sort"`
	Order       string            `json:"order"`
}

// AppointmentAvailabilityRequest represents a request to check availability
type AppointmentAvailabilityRequest struct {
	StaffID      string        `json:"staff_id"`
	Date            time.Time     `json:"date"`
	ServiceDuration time.Duration `json:"service_duration"`
}

// AvailableSlot represents an available time slot
type AvailableSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
}

// AvailabilityResponse represents availability check response
type AvailabilityResponse struct {
	StaffID string          `json:"staff_id"`
	Date       time.Time       `json:"date"`
	Slots      []AvailableSlot `json:"slots"`
}

// ConflictCheckRequest represents a request to check appointment conflicts
type ConflictCheckRequest struct {
	StaffID string    `json:"staff_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	ExcludeID  *string   `json:"exclude_id"`
}

// ConflictInfo represents conflict information
type ConflictInfo struct {
	Conflict      bool      `json:"conflict"`
	ConflictIDs   []string  `json:"conflict_ids,omitempty"`
	ConflictCount int       `json:"conflict_count"`
	Suggestions   []time.Time `json:"suggestions,omitempty"`
}

// NewAppointmentHTTPClient creates a new Appointment HTTP client with optimized settings
func NewAppointmentHTTPClient(baseURL string, timeout time.Duration) *AppointmentHTTPClient {
	log := logger.NewDefault()
	log.Info("Creating optimized Appointment HTTP client", "base_url", baseURL)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient(baseURL, config, log)

	return &AppointmentHTTPClient{
		baseURL:    baseURL,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
	}
}

// NewAppointmentHTTPClientWithDiscovery uses service discovery to create an Appointment HTTP client
func NewAppointmentHTTPClientWithDiscovery(disc discovery.Discovery, timeout time.Duration) (*AppointmentHTTPClient, error) {
	log := logger.NewDefault()

	// 从服务发现获取 appointments 服务地址
	instance, err := disc.GetService(context.Background(), "appointments")
	if err != nil {
		log.Error("Failed to discover Appointments service", "error", err.Error())
		return nil, fmt.Errorf("failed to discover Appointments service: %w", err)
	}

	address := instance.Address()
	log.Info("Discovered Appointments service", "address", address, "instance_id", instance.ID)

	// Create optimized client configuration
	config := DefaultHTTPClientConfig()
	config.Timeout = timeout

	// Create optimized HTTP client
	optimizedClient := NewOptimizedHTTPClient("http://"+address, config, log)

	return &AppointmentHTTPClient{
		baseURL:    "http://" + address,
		client:     optimizedClient,
		httpClient: &http.Client{Timeout: timeout}, // Fallback
		logger:     log,
		discovery:  disc,
	}, nil
}

// ListAppointments retrieves a list of appointments
func (c *AppointmentHTTPClient) ListAppointments(ctx context.Context, filter *AppointmentFilter) ([]Appointment, int64, error) {
	// 构建查询参数
	params := make(map[string]string)
	if filter.Page > 0 {
		params["page"] = fmt.Sprintf("%d", filter.Page)
	}
	if filter.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", filter.Limit)
	}
	if filter.CustomerID != nil {
		params["customer_id"] = *filter.CustomerID
	}
	if filter.StaffID != nil {
		params["staff_id"] = *filter.StaffID
	}
	if filter.ServiceID != nil {
		params["service_id"] = *filter.ServiceID
	}
	if filter.Status != nil {
		params["status"] = string(*filter.Status)
	}
	if filter.Sort != "" {
		params["sort"] = filter.Sort
	}
	if filter.Order != "" {
		params["order"] = filter.Order
	}

	url := c.buildURL("/api/appointments", params)

	var response struct {
		Code    int           `json:"code"`
		Message string        `json:"message"`
		Data    struct {
			Appointments []Appointment `json:"appointments"`
			Total       int64         `json:"total"`
			Page        int           `json:"page"`
			Limit       int           `json:"limit"`
		} `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to list appointments", "error", err.Error())
		return nil, 0, err
	}

	if response.Code != 200 {
		return nil, 0, fmt.Errorf("appointments service error: %s", response.Message)
	}

	return response.Data.Appointments, response.Data.Total, nil
}

// CreateAppointment creates a new appointment
func (c *AppointmentHTTPClient) CreateAppointment(ctx context.Context, req *CreateAppointmentRequestHTTP) (*Appointment, error) {
	url := c.buildURL("/api/appointments", nil)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    Appointment `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to create appointment", "error", err.Error())
		return nil, err
	}

	if response.Code != 201 {
		return nil, fmt.Errorf("appointments service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetAppointment retrieves an appointment by ID
func (c *AppointmentHTTPClient) GetAppointment(ctx context.Context, appointmentID string) (*Appointment, error) {
	url := c.buildURL(fmt.Sprintf("/api/appointments/%s", appointmentID), nil)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    Appointment `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to get appointment", "appointment_id", appointmentID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("appointments service error: %s", response.Message)
	}

	return &response.Data, nil
}

// UpdateAppointment updates an appointment
func (c *AppointmentHTTPClient) UpdateAppointment(ctx context.Context, appointmentID string, req *UpdateAppointmentRequestHTTP) (*Appointment, error) {
	url := c.buildURL(fmt.Sprintf("/api/appointments/%s", appointmentID), nil)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    Appointment `json:"data"`
	}

	err := c.doRequest(ctx, "PUT", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to update appointment", "appointment_id", appointmentID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("appointments service error: %s", response.Message)
	}

	return &response.Data, nil
}

// DeleteAppointment deletes an appointment
func (c *AppointmentHTTPClient) DeleteAppointment(ctx context.Context, appointmentID string) error {
	url := c.buildURL(fmt.Sprintf("/api/appointments/%s", appointmentID), nil)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	err := c.doRequest(ctx, "DELETE", url, nil, &response)
	if err != nil {
		c.logger.Error("Failed to delete appointment", "appointment_id", appointmentID, "error", err.Error())
		return err
	}

	if response.Code != 200 {
		return fmt.Errorf("appointments service error: %s", response.Message)
	}

	return nil
}

// UpdateAppointmentStatus updates appointment status
func (c *AppointmentHTTPClient) UpdateAppointmentStatus(ctx context.Context, appointmentID string, req *UpdateAppointmentStatusRequestHTTP) (*Appointment, error) {
	url := c.buildURL(fmt.Sprintf("/api/appointments/%s/status", appointmentID), nil)

	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    Appointment `json:"data"`
	}

	err := c.doRequest(ctx, "PUT", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to update appointment status", "appointment_id", appointmentID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("appointments service error: %s", response.Message)
	}

	return &response.Data, nil
}

// CheckAvailability checks availability for a given employee and date
func (c *AppointmentHTTPClient) CheckAvailability(ctx context.Context, req *AppointmentAvailabilityRequest) (*AvailabilityResponse, error) {
	url := c.buildURL("/api/appointments/availability", nil)

	var response struct {
		Code    int                 `json:"code"`
		Message string              `json:"message"`
		Data    AvailabilityResponse `json:"data"`
	}

	err := c.doRequest(ctx, "GET", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to check availability", "staff_id", req.StaffID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("appointments service error: %s", response.Message)
	}

	return &response.Data, nil
}

// CheckConflict checks for appointment conflicts
func (c *AppointmentHTTPClient) CheckConflict(ctx context.Context, req *ConflictCheckRequest) (*ConflictInfo, error) {
	url := c.buildURL("/api/appointments/conflict-check", nil)

	var response struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Data    ConflictInfo `json:"data"`
	}

	err := c.doRequest(ctx, "POST", url, req, &response)
	if err != nil {
		c.logger.Error("Failed to check conflict", "staff_id", req.StaffID, "error", err.Error())
		return nil, err
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("appointments service error: %s", response.Message)
	}

	return &response.Data, nil
}

// GetAppointmentsByCustomer retrieves appointments for a specific customer
func (c *AppointmentHTTPClient) GetAppointmentsByCustomer(ctx context.Context, customerID string, filter *AppointmentFilter) ([]Appointment, error) {
	if filter == nil {
		filter = &AppointmentFilter{}
	}
	filter.CustomerID = &customerID
	appointments, _, err := c.ListAppointments(ctx, filter)
	return appointments, err
}

// GetAppointmentsByEmployee retrieves appointments for a specific employee
func (c *AppointmentHTTPClient) GetAppointmentsByEmployee(ctx context.Context, employeeID string, filter *AppointmentFilter) ([]Appointment, error) {
	if filter == nil {
		filter = &AppointmentFilter{}
	}
	filter.StaffID = &employeeID
	appointments, _, err := c.ListAppointments(ctx, filter)
	return appointments, err
}

// Helper functions

func (c *AppointmentHTTPClient) buildURL(path string, params map[string]string) string {
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

func (c *AppointmentHTTPClient) doRequest(ctx context.Context, method, url string, body interface{}, response interface{}) error {
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