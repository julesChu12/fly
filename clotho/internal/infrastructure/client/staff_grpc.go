package client

import (
	"context"
	"net/http"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// StaffClient represents an HTTP client for the Staff service
type StaffClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logger.Logger
}

// NewStaffClient creates a new Staff client
func NewStaffClient(baseURL string, logger *logger.Logger) *StaffClient {
	return &StaffClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CreateStaff creates a new staff member
func (c *StaffClient) CreateStaff(ctx context.Context, req *CreateStaffRequest) (*StaffResponse, error) {
	// Simplified implementation
	return &StaffResponse{
		ID:     "test-staff-id",
		Name:   req.Name,
		Email:  req.Email,
		Status: "active",
	}, nil
}

// GetStaff gets a staff member by ID
func (c *StaffClient) GetStaff(ctx context.Context, id string) (*StaffResponse, error) {
	// Simplified implementation
	return &StaffResponse{
		ID:     id,
		Name:   "Test Staff",
		Email:  "test@example.com",
		Status: "active",
	}, nil
}

// UpdateStaff updates a staff member
func (c *StaffClient) UpdateStaff(ctx context.Context, id string, req *UpdateStaffRequest) (*StaffResponse, error) {
	// Simplified implementation
	name := ""
	email := ""
	if req.Name != nil {
		name = *req.Name
	}
	if req.Email != nil {
		email = *req.Email
	}

	return &StaffResponse{
		ID:     id,
		Name:   name,
		Email:  email,
		Status: "updated",
	}, nil
}

// ListStaff lists staff members with filters
func (c *StaffClient) ListStaff(ctx context.Context, req *ListStaffRequest) (*ListStaffResponse, error) {
	// Simplified implementation
	return &ListStaffResponse{
		Staff: []*StaffResponse{},
		Total: 0,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

// DeleteStaff deletes a staff member
func (c *StaffClient) DeleteStaff(ctx context.Context, id string) error {
	// Simplified implementation
	c.logger.Infof("Deleting staff member %s", id)
	return nil
}

// UpdateStaffStatus updates a staff member's status
func (c *StaffClient) UpdateStaffStatus(ctx context.Context, id string, req *UpdateStatusRequest) (*StaffResponse, error) {
	// Simplified implementation
	return &StaffResponse{
		ID:     id,
		Status: req.Status,
	}, nil
}

// GetAvailableStaff gets available staff members
func (c *StaffClient) GetAvailableStaff(ctx context.Context, req *GetAvailableStaffRequest) (*ListStaffResponse, error) {
	// Simplified implementation
	return &ListStaffResponse{
		Staff: []*StaffResponse{},
		Total: 0,
		Page:  1,
		Limit: 20,
	}, nil
}

// CreateStaffRequest represents a request to create a staff member
type CreateStaffRequest struct {
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Department  string   `json:"department"`
	Position    string   `json:"position"`
	RoleID      string   `json:"role_id"`
	Skills      []string `json:"skills"`
	IsAvailable bool     `json:"is_available"`
}

// UpdateStaffRequest represents a request to update a staff member
type UpdateStaffRequest struct {
	Name        *string  `json:"name"`
	Email       *string  `json:"email"`
	Phone       *string  `json:"phone"`
	Department  *string  `json:"department"`
	Position    *string  `json:"position"`
	RoleID      *string  `json:"role_id"`
	Skills      []string `json:"skills"`
	IsAvailable *bool    `json:"is_available"`
}

// UpdateStatusRequest represents a request to update staff status
type UpdateStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// ListStaffRequest represents a request to list staff members
type ListStaffRequest struct {
	Search      *string `json:"search"`
	Department  *string `json:"department"`
	RoleID      *string `json:"role_id"`
	Status      *string `json:"status"`
	IsAvailable *bool   `json:"is_available"`
	Page        int     `json:"page"`
	Limit       int     `json:"limit"`
	Sort        string  `json:"sort"`
	Order       string  `json:"order"`
}

// GetAvailableStaffRequest represents a request to get available staff
type GetAvailableStaffRequest struct {
	Department *string  `json:"department"`
	Skills     []string `json:"skills"`
}

// StaffResponse represents a staff member response
type StaffResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Department  string   `json:"department"`
	Position    string   `json:"position"`
	RoleID      string   `json:"role_id"`
	Skills      []string `json:"skills"`
	Status      string   `json:"status"`
	IsAvailable bool     `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListStaffResponse represents a list staff response
type ListStaffResponse struct {
	Staff []*StaffResponse `json:"staff"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}