package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// StaffProxy represents the use case for staff operations
type StaffProxy struct {
	staffClient *client.StaffHTTPClient
	logger      *logger.Logger
}

// NewStaffProxy creates a new StaffProxy instance
func NewStaffProxy(staffClient *client.StaffHTTPClient, logger *logger.Logger) *StaffProxy {
	return &StaffProxy{
		staffClient: staffClient,
		logger:      logger,
	}
}

// ListStaff retrieves a list of staff members
func (p *StaffProxy) ListStaff(ctx context.Context, filter *client.StaffFilter) ([]client.Staff, int64, error) {
	p.logger.Info("Retrieving staff list", "filter", filter)

	start := time.Now()
	staff, total, err := p.staffClient.ListStaff(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve staff list", "error", err.Error(), "duration", duration)
		return nil, 0, fmt.Errorf("failed to retrieve staff list: %w", err)
	}

	p.logger.Info("Successfully retrieved staff list", "count", len(staff), "total", total, "duration", duration)
	return staff, total, nil
}

// CreateStaff creates a new staff member
func (p *StaffProxy) CreateStaff(ctx context.Context, req *client.CreateStaffRequestHTTP) (*client.Staff, error) {
	p.logger.Info("Creating new staff member", "name", req.Name, "email", req.Email, "department", req.Department)

	// Validate request
	if err := p.validateCreateStaffRequest(req); err != nil {
		p.logger.Error("Invalid staff creation request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	staff, err := p.staffClient.CreateStaff(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to create staff member", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to create staff member: %w", err)
	}

	p.logger.Info("Successfully created staff member", "staff_id", staff.ID, "name", staff.Name, "duration", duration)
	return staff, nil
}

// GetStaff retrieves a staff member by ID
func (p *StaffProxy) GetStaff(ctx context.Context, staffID string) (*client.Staff, error) {
	p.logger.Info("Retrieving staff member", "staff_id", staffID)

	start := time.Now()
	staff, err := p.staffClient.GetStaff(ctx, staffID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve staff member", "staff_id", staffID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve staff member: %w", err)
	}

	p.logger.Info("Successfully retrieved staff member", "staff_id", staffID, "name", staff.Name, "duration", duration)
	return staff, nil
}

// UpdateStaff updates a staff member
func (p *StaffProxy) UpdateStaff(ctx context.Context, staffID string, req *client.UpdateStaffRequestHTTP) (*client.Staff, error) {
	p.logger.Info("Updating staff member", "staff_id", staffID)

	// Validate that staff exists first
	_, err := p.staffClient.GetStaff(ctx, staffID)
	if err != nil {
		p.logger.Error("Staff member not found for update", "staff_id", staffID, "error", err.Error())
		return nil, fmt.Errorf("staff member not found: %w", err)
	}

	start := time.Now()
	updatedStaff, err := p.staffClient.UpdateStaff(ctx, staffID, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to update staff member", "staff_id", staffID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to update staff member: %w", err)
	}

	p.logger.Info("Successfully updated staff member", "staff_id", staffID, "name", updatedStaff.Name, "duration", duration)
	return updatedStaff, nil
}

// DeleteStaff deletes a staff member
func (p *StaffProxy) DeleteStaff(ctx context.Context, staffID string) error {
	p.logger.Info("Deleting staff member", "staff_id", staffID)

	// Validate that staff exists first
	_, err := p.staffClient.GetStaff(ctx, staffID)
	if err != nil {
		p.logger.Error("Staff member not found for deletion", "staff_id", staffID, "error", err.Error())
		return fmt.Errorf("staff member not found: %w", err)
	}

	start := time.Now()
	err = p.staffClient.DeleteStaff(ctx, staffID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to delete staff member", "staff_id", staffID, "error", err.Error(), "duration", duration)
		return fmt.Errorf("failed to delete staff member: %w", err)
	}

	p.logger.Info("Successfully deleted staff member", "staff_id", staffID, "duration", duration)
	return nil
}

// UpdateStaffStatus updates staff status
func (p *StaffProxy) UpdateStaffStatus(ctx context.Context, staffID string, req *client.UpdateStatusRequestHTTP) (*client.Staff, error) {
	p.logger.Info("Updating staff status", "staff_id", staffID, "status", req.Status)

	// Validate that staff exists first
	_, err := p.staffClient.GetStaff(ctx, staffID)
	if err != nil {
		p.logger.Error("Staff member not found for status update", "staff_id", staffID, "error", err.Error())
		return nil, fmt.Errorf("staff member not found: %w", err)
	}

	// Validate status transition
	if err := p.validateStatusTransition(req.Status); err != nil {
		p.logger.Error("Invalid status transition", "staff_id", staffID, "from_status", req.Status, "error", err.Error())
		return nil, fmt.Errorf("invalid status transition: %w", err)
	}

	start := time.Now()
	updatedStaff, err := p.staffClient.UpdateStaffStatus(ctx, staffID, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to update staff status", "staff_id", staffID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to update staff status: %w", err)
	}

	p.logger.Info("Successfully updated staff status", "staff_id", staffID, "status", req.Status, "duration", duration)
	return updatedStaff, nil
}

// GetAvailableStaff retrieves available staff members
func (p *StaffProxy) GetAvailableStaff(ctx context.Context, filter *client.StaffFilter) ([]client.Staff, error) {
	p.logger.Info("Retrieving available staff", "filter", filter)

	start := time.Now()
	staff, err := p.staffClient.GetAvailableStaff(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve available staff", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve available staff: %w", err)
	}

	p.logger.Info("Successfully retrieved available staff", "count", len(staff), "duration", duration)
	return staff, nil
}

// ListRoles retrieves a list of staff roles
func (p *StaffProxy) ListRoles(ctx context.Context) ([]client.StaffRole, error) {
	p.logger.Info("Retrieving staff roles list")

	start := time.Now()
	roles, err := p.staffClient.ListRoles(ctx)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve staff roles", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve staff roles: %w", err)
	}

	p.logger.Info("Successfully retrieved staff roles", "count", len(roles), "duration", duration)
	return roles, nil
}

// CreateRole creates a new staff role
func (p *StaffProxy) CreateRole(ctx context.Context, req map[string]interface{}) (*client.StaffRole, error) {
	p.logger.Info("Creating new staff role", "name", req["name"])

	// Validate request
	if err := p.validateCreateRoleRequest(req); err != nil {
		p.logger.Error("Invalid role creation request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	role, err := p.staffClient.CreateRole(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to create staff role", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to create staff role: %w", err)
	}

	p.logger.Info("Successfully created staff role", "role_id", role.ID, "name", role.Name, "duration", duration)
	return role, nil
}

// GetAvailability retrieves staff availability
func (p *StaffProxy) GetAvailability(ctx context.Context, staffID string) ([]client.StaffAvailability, error) {
	p.logger.Info("Retrieving staff availability", "staff_id", staffID)

	// Validate that staff exists first
	_, err := p.staffClient.GetStaff(ctx, staffID)
	if err != nil {
		p.logger.Error("Staff member not found for availability retrieval", "staff_id", staffID, "error", err.Error())
		return nil, fmt.Errorf("staff member not found: %w", err)
	}

	start := time.Now()
	availability, err := p.staffClient.GetAvailability(ctx, staffID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve staff availability", "staff_id", staffID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve staff availability: %w", err)
	}

	p.logger.Info("Successfully retrieved staff availability", "staff_id", staffID, "count", len(availability), "duration", duration)
	return availability, nil
}

// SetAvailability sets staff availability
func (p *StaffProxy) SetAvailability(ctx context.Context, staffID string, req *client.AvailabilityRequest) error {
	p.logger.Info("Setting staff availability", "staff_id", staffID, "availability_count", len(req.Availabilities))

	// Validate that staff exists first
	_, err := p.staffClient.GetStaff(ctx, staffID)
	if err != nil {
		p.logger.Error("Staff member not found for availability setting", "staff_id", staffID, "error", err.Error())
		return fmt.Errorf("staff member not found: %w", err)
	}

	// Validate availability request
	if err := p.validateAvailabilityRequest(req); err != nil {
		p.logger.Error("Invalid availability request", "staff_id", staffID, "error", err.Error())
		return fmt.Errorf("invalid availability request: %w", err)
	}

	start := time.Now()
	err = p.staffClient.SetAvailability(ctx, staffID, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to set staff availability", "staff_id", staffID, "error", err.Error(), "duration", duration)
		return fmt.Errorf("failed to set staff availability: %w", err)
	}

	p.logger.Info("Successfully set staff availability", "staff_id", staffID, "availability_count", len(req.Availabilities), "duration", duration)
	return nil
}

// Validation helper functions

func (p *StaffProxy) validateCreateStaffRequest(req *client.CreateStaffRequestHTTP) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Department == "" {
		return fmt.Errorf("department is required")
	}
	if req.Position == "" {
		return fmt.Errorf("position is required")
	}
	if req.RoleID == "" {
		return fmt.Errorf("role_id is required")
	}
	return nil
}

func (p *StaffProxy) validateStatusTransition(status string) error {
	validStatuses := map[string]bool{
		"active":     true,
		"inactive":   true,
		"on_leave":   true,
		"suspended":  true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	return nil
}

func (p *StaffProxy) validateCreateRoleRequest(req map[string]interface{}) error {
	if name, ok := req["name"].(string); !ok || name == "" {
		return fmt.Errorf("role name is required")
	}
	if code, ok := req["code"].(string); !ok || code == "" {
		return fmt.Errorf("role code is required")
	}
	return nil
}

func (p *StaffProxy) validateAvailabilityRequest(req *client.AvailabilityRequest) error {
	if req.StaffID == "" {
		return fmt.Errorf("staff_id is required")
	}
	if len(req.Availabilities) == 0 {
		return fmt.Errorf("at least one availability item is required")
	}

	// Validate each availability item
	for i, item := range req.Availabilities {
		if item.DayOfWeek < 0 || item.DayOfWeek > 6 {
			return fmt.Errorf("invalid day_of_week at index %d: must be 0-6", i)
		}
		if item.StartTime == "" {
			return fmt.Errorf("start_time is required at index %d", i)
		}
		if item.EndTime == "" {
			return fmt.Errorf("end_time is required at index %d", i)
		}
		// Note: Additional time format validation could be added here
	}

	return nil
}

// BatchDeleteStaff batch deletes staff members by IDs
func (p *StaffProxy) BatchDeleteStaff(ctx context.Context, staffIDs []uint) error {
	p.logger.Info("Batch deleting staff", "count", len(staffIDs))

	start := time.Now()
	err := p.staffClient.BatchDeleteStaff(ctx, staffIDs)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to batch delete staff", "error", err.Error(), "duration", duration)
		return fmt.Errorf("failed to batch delete staff: %w", err)
	}

	p.logger.Info("Successfully batch deleted staff", "count", len(staffIDs), "duration", duration)
	return nil
}