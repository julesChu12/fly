package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// AppointmentProxy represents the use case for appointment operations
type AppointmentProxy struct {
	appointmentClient *client.AppointmentHTTPClient
	logger            *logger.Logger
}

// NewAppointmentProxy creates a new AppointmentProxy instance
func NewAppointmentProxy(appointmentClient *client.AppointmentHTTPClient, logger *logger.Logger) *AppointmentProxy {
	return &AppointmentProxy{
		appointmentClient: appointmentClient,
		logger:            logger,
	}
}

// ListAppointments retrieves a list of appointments
func (p *AppointmentProxy) ListAppointments(ctx context.Context, filter *client.AppointmentFilter) ([]client.Appointment, int64, error) {
	p.logger.Info("Retrieving appointment list", "filter", filter)

	// Set defaults for filter
	if filter != nil {
		if filter.Page <= 0 {
			filter.Page = 1
		}
		if filter.Limit <= 0 {
			filter.Limit = 20
		}
	}

	start := time.Now()
	appointments, total, err := p.appointmentClient.ListAppointments(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve appointment list", "error", err.Error(), "duration", duration)
		return nil, 0, fmt.Errorf("failed to retrieve appointment list: %w", err)
	}

	p.logger.Info("Successfully retrieved appointment list", "count", len(appointments), "total", total, "duration", duration)
	return appointments, total, nil
}

// CreateAppointment creates a new appointment
func (p *AppointmentProxy) CreateAppointment(ctx context.Context, req *client.CreateAppointmentRequestHTTP) (*client.Appointment, error) {
	p.logger.Info("Creating new appointment", "customer_id", req.CustomerID, "staff_id", req.StaffID, "service_id", req.ServiceID)

	// Validate request
	if err := p.validateCreateAppointmentRequest(req); err != nil {
		p.logger.Error("Invalid appointment creation request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	appointment, err := p.appointmentClient.CreateAppointment(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to create appointment", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to create appointment: %w", err)
	}

	p.logger.Info("Successfully created appointment", "appointment_id", appointment.ID, "customer_id", appointment.CustomerID, "duration", duration)
	return appointment, nil
}

// GetAppointment retrieves an appointment by ID
func (p *AppointmentProxy) GetAppointment(ctx context.Context, appointmentID string) (*client.Appointment, error) {
	p.logger.Info("Retrieving appointment", "appointment_id", appointmentID)

	if appointmentID == "" {
		return nil, fmt.Errorf("appointment ID is required")
	}

	start := time.Now()
	appointment, err := p.appointmentClient.GetAppointment(ctx, appointmentID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve appointment", "appointment_id", appointmentID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve appointment: %w", err)
	}

	p.logger.Info("Successfully retrieved appointment", "appointment_id", appointment.ID, "customer_id", appointment.CustomerID, "duration", duration)
	return appointment, nil
}

// UpdateAppointment updates an existing appointment
func (p *AppointmentProxy) UpdateAppointment(ctx context.Context, appointmentID string, req *client.UpdateAppointmentRequestHTTP) (*client.Appointment, error) {
	p.logger.Info("Updating appointment", "appointment_id", appointmentID)

	if appointmentID == "" {
		return nil, fmt.Errorf("appointment ID is required")
	}

	// Validate that appointment exists first
	_, err := p.appointmentClient.GetAppointment(ctx, appointmentID)
	if err != nil {
		p.logger.Error("Appointment not found for update", "appointment_id", appointmentID, "error", err.Error())
		return nil, fmt.Errorf("appointment not found: %w", err)
	}

	start := time.Now()
	updatedAppointment, err := p.appointmentClient.UpdateAppointment(ctx, appointmentID, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to update appointment", "appointment_id", appointmentID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to update appointment: %w", err)
	}

	p.logger.Info("Successfully updated appointment", "appointment_id", updatedAppointment.ID, "duration", duration)
	return updatedAppointment, nil
}

// DeleteAppointment deletes an appointment
func (p *AppointmentProxy) DeleteAppointment(ctx context.Context, appointmentID string) error {
	p.logger.Info("Deleting appointment", "appointment_id", appointmentID)

	if appointmentID == "" {
		return fmt.Errorf("appointment ID is required")
	}

	// Validate that appointment exists first
	_, err := p.appointmentClient.GetAppointment(ctx, appointmentID)
	if err != nil {
		p.logger.Error("Appointment not found for deletion", "appointment_id", appointmentID, "error", err.Error())
		return fmt.Errorf("appointment not found: %w", err)
	}

	start := time.Now()
	err = p.appointmentClient.DeleteAppointment(ctx, appointmentID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to delete appointment", "appointment_id", appointmentID, "error", err.Error(), "duration", duration)
		return fmt.Errorf("failed to delete appointment: %w", err)
	}

	p.logger.Info("Successfully deleted appointment", "appointment_id", appointmentID, "duration", duration)
	return nil
}

// UpdateAppointmentStatus updates appointment status
func (p *AppointmentProxy) UpdateAppointmentStatus(ctx context.Context, appointmentID string, req *client.UpdateAppointmentStatusRequestHTTP) (*client.Appointment, error) {
	p.logger.Info("Updating appointment status", "appointment_id", appointmentID, "status", req.Status)

	if appointmentID == "" {
		return nil, fmt.Errorf("appointment ID is required")
	}

	// Validate status transition
	if err := p.validateStatusTransition(req.Status); err != nil {
		p.logger.Error("Invalid status transition", "appointment_id", appointmentID, "from_status", req.Status, "error", err.Error())
		return nil, fmt.Errorf("invalid status transition: %w", err)
	}

	// Validate that appointment exists first
	_, err := p.appointmentClient.GetAppointment(ctx, appointmentID)
	if err != nil {
		p.logger.Error("Appointment not found for status update", "appointment_id", appointmentID, "error", err.Error())
		return nil, fmt.Errorf("appointment not found: %w", err)
	}

	start := time.Now()
	updatedAppointment, err := p.appointmentClient.UpdateAppointmentStatus(ctx, appointmentID, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to update appointment status", "appointment_id", appointmentID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to update appointment status: %w", err)
	}

	p.logger.Info("Successfully updated appointment status", "appointment_id", updatedAppointment.ID, "new_status", updatedAppointment.Status, "duration", duration)
	return updatedAppointment, nil
}

// CheckAvailability checks availability for a given employee and date
func (p *AppointmentProxy) CheckAvailability(ctx context.Context, req *client.AppointmentAvailabilityRequest) (*client.AvailabilityResponse, error) {
	p.logger.Info("Checking availability", "staff_id", req.StaffID, "date", req.Date)

	// Validate request
	if err := p.validateAvailabilityRequest(req); err != nil {
		p.logger.Error("Invalid availability request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	availability, err := p.appointmentClient.CheckAvailability(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to check availability", "staff_id", req.StaffID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	p.logger.Info("Successfully checked availability", "staff_id", availability.StaffID, "slots_count", len(availability.Slots), "duration", duration)
	return availability, nil
}

// CheckConflict checks for appointment conflicts
func (p *AppointmentProxy) CheckConflict(ctx context.Context, req *client.ConflictCheckRequest) (*client.ConflictInfo, error) {
	p.logger.Info("Checking appointment conflict", "staff_id", req.StaffID, "start_time", req.StartTime, "end_time", req.EndTime)

	// Validate request
	if err := p.validateConflictCheckRequest(req); err != nil {
		p.logger.Error("Invalid conflict check request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	conflictInfo, err := p.appointmentClient.CheckConflict(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to check conflict", "staff_id", req.StaffID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to check conflict: %w", err)
	}

	p.logger.Info("Successfully checked conflict", "staff_id", req.StaffID, "conflict", conflictInfo.Conflict, "conflict_count", conflictInfo.ConflictCount, "duration", duration)
	return conflictInfo, nil
}

// GetAppointmentsByCustomer retrieves appointments for a specific customer
func (p *AppointmentProxy) GetAppointmentsByCustomer(ctx context.Context, customerID string, filter *client.AppointmentFilter) ([]client.Appointment, error) {
	p.logger.Info("Retrieving appointments for customer", "customer_id", customerID)

	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}

	start := time.Now()
	appointments, err := p.appointmentClient.GetAppointmentsByCustomer(ctx, customerID, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve appointments for customer", "customer_id", customerID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve appointments for customer: %w", err)
	}

	p.logger.Info("Successfully retrieved appointments for customer", "customer_id", customerID, "count", len(appointments), "duration", duration)
	return appointments, nil
}

// GetAppointmentsByEmployee retrieves appointments for a specific employee
func (p *AppointmentProxy) GetAppointmentsByEmployee(ctx context.Context, employeeID string, filter *client.AppointmentFilter) ([]client.Appointment, error) {
	p.logger.Info("Retrieving appointments for employee", "staff_id", employeeID)

	if employeeID == "" {
		return nil, fmt.Errorf("employee ID is required")
	}

	start := time.Now()
	appointments, err := p.appointmentClient.GetAppointmentsByEmployee(ctx, employeeID, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve appointments for employee", "staff_id", employeeID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve appointments for employee: %w", err)
	}

	p.logger.Info("Successfully retrieved appointments for employee", "staff_id", employeeID, "count", len(appointments), "duration", duration)
	return appointments, nil
}

// Validation helper functions

func (p *AppointmentProxy) validateCreateAppointmentRequest(req *client.CreateAppointmentRequestHTTP) error {
	if req.CustomerID == "" {
		return fmt.Errorf("customer ID is required")
	}
	if req.StaffID == "" {
		return fmt.Errorf("employee ID is required")
	}
	if req.ServiceID == "" {
		return fmt.Errorf("service ID is required")
	}
	if req.StartTime.IsZero() {
		return fmt.Errorf("start time is required")
	}
	if req.EndTime.IsZero() {
		return fmt.Errorf("end time is required")
	}
	if req.EndTime.Before(req.StartTime) {
		return fmt.Errorf("end time must be after start time")
	}
	if req.EndTime.Equal(req.StartTime) {
		return fmt.Errorf("end time must be different from start time")
	}
	return nil
}

func (p *AppointmentProxy) validateAvailabilityRequest(req *client.AppointmentAvailabilityRequest) error {
	if req.StaffID == "" {
		return fmt.Errorf("employee ID is required")
	}
	if req.Date.IsZero() {
		return fmt.Errorf("date is required")
	}
	if req.ServiceDuration <= 0 {
		return fmt.Errorf("service duration must be positive")
	}
	return nil
}

func (p *AppointmentProxy) validateConflictCheckRequest(req *client.ConflictCheckRequest) error {
	if req.StaffID == "" {
		return fmt.Errorf("employee ID is required")
	}
	if req.StartTime.IsZero() {
		return fmt.Errorf("start time is required")
	}
	if req.EndTime.IsZero() {
		return fmt.Errorf("end time is required")
	}
	if req.EndTime.Before(req.StartTime) {
		return fmt.Errorf("end time must be after start time")
	}
	if req.EndTime.Equal(req.StartTime) {
		return fmt.Errorf("end time must be different from start time")
	}
	return nil
}

func (p *AppointmentProxy) validateStatusTransition(status string) error {
	validStatuses := map[string]bool{
		"pending":      true,
		"confirmed":    true,
		"in_progress":  true,
		"completed":    true,
		"cancelled":    true,
		"no_show":      true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	return nil
}