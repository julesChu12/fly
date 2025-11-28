package grpc

import (
	"context"
	"fmt"
	"time"

	appointmentsv1 "github.com/julesChu12/fly/appointments/api/proto/appointments/v1"
	"github.com/julesChu12/fly/appointments/internal/application/service"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AppointmentServer implements the gRPC AppointmentService
type AppointmentServer struct {
	appointmentsv1.UnimplementedAppointmentServiceServer
	appointmentService service.AppointmentService
}

// NewAppointmentServer creates a new AppointmentServer
func NewAppointmentServer(appointmentService service.AppointmentService) *AppointmentServer {
	return &AppointmentServer{
		appointmentService: appointmentService,
	}
}

// CreateAppointment creates a new appointment
func (s *AppointmentServer) CreateAppointment(ctx context.Context, req *appointmentsv1.CreateAppointmentRequest) (*appointmentsv1.CreateAppointmentResponse, error) {
	// Convert protobuf request to DTO
	createReq := &dto.CreateAppointmentRequest{
		CustomerID:     req.CustomerId,
		StaffID:        req.StaffId,
		ServiceID:      req.ServiceId,
		StartTime:      req.StartTime.AsTime(),
		EndTime:        req.EndTime.AsTime(),
		Status:         entity.AppointmentStatus(req.Status),
		Notes:          req.Notes,
		CustomerName:   req.CustomerName,
		StaffName:      req.StaffName,
		ServiceName:    req.ServiceName,
		ServiceDuration: int(req.ServiceDuration),
		ServicePrice:    req.ServicePrice,
	}

	// Call service
	appointment, err := s.appointmentService.CreateAppointment(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create appointment: %v", err)
	}

	// Convert response to protobuf
	return &appointmentsv1.CreateAppointmentResponse{
		Appointment: toProtoAppointment(appointment),
	}, nil
}

// GetAppointment retrieves an appointment by ID
func (s *AppointmentServer) GetAppointment(ctx context.Context, req *appointmentsv1.GetAppointmentRequest) (*appointmentsv1.GetAppointmentResponse, error) {
	appointment, err := s.appointmentService.GetAppointment(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Appointment not found: %v", err)
	}

	return &appointmentsv1.GetAppointmentResponse{
		Appointment: toProtoAppointment(appointment),
	}, nil
}

// ListAppointments retrieves a paginated list of appointments
func (s *AppointmentServer) ListAppointments(ctx context.Context, req *appointmentsv1.ListAppointmentsRequest) (*appointmentsv1.ListAppointmentsResponse, error) {
	// Convert request
	filter := &dto.AppointmentFilter{
		Page:  int(req.Page),
		Limit: int(req.PageSize),
	}

	if req.CustomerId != 0 {
		filter.CustomerID = req.CustomerId
	}
	if req.StaffId != 0 {
		filter.StaffID = req.StaffId
	}
	if req.ServiceId != 0 {
		filter.ServiceID = req.ServiceId
	}
	if req.Status != "" {
		filter.Status = entity.AppointmentStatus(req.Status)
	}
	if req.StartDate != nil {
		filter.StartDate = req.StartDate.AsTime()
	}
	if req.EndDate != nil {
		filter.EndDate = req.EndDate.AsTime()
	}

	// Call service
	result, err := s.appointmentService.ListAppointments(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list appointments: %v", err)
	}

	// Convert response
	appointments := make([]*appointmentsv1.Appointment, len(result.Appointments))
	for i, apt := range result.Appointments {
		appointments[i] = toProtoAppointment(&apt)
	}

	return &appointmentsv1.ListAppointmentsResponse{
		Appointments: appointments,
		Total:        result.Total,
		Page:         int32(result.Page),
		PageSize:     int32(result.Limit),
	}, nil
}

// UpdateAppointment updates an existing appointment
func (s *AppointmentServer) UpdateAppointment(ctx context.Context, req *appointmentsv1.UpdateAppointmentRequest) (*appointmentsv1.UpdateAppointmentResponse, error) {
	// Convert protobuf request to DTO
	updateReq := &dto.UpdateAppointmentRequest{
		StartTime:      req.StartTime.AsTime(),
		EndTime:        req.EndTime.AsTime(),
		Status:         entity.AppointmentStatus(req.Status),
		Notes:          req.Notes,
		CustomerName:   req.CustomerName,
		StaffName:      req.StaffName,
		ServiceName:    req.ServiceName,
		ServiceDuration: int(req.ServiceDuration),
		ServicePrice:    req.ServicePrice,
	}

	// Call service
	appointment, err := s.appointmentService.UpdateAppointment(ctx, req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update appointment: %v", err)
	}

	return &appointmentsv1.UpdateAppointmentResponse{
		Appointment: toProtoAppointment(appointment),
	}, nil
}

// DeleteAppointment deletes an appointment
func (s *AppointmentServer) DeleteAppointment(ctx context.Context, req *appointmentsv1.DeleteAppointmentRequest) (*appointmentsv1.DeleteAppointmentResponse, error) {
	err := s.appointmentService.DeleteAppointment(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete appointment: %v", err)
	}

	return &appointmentsv1.DeleteAppointmentResponse{
		Success: true,
		Message: "Appointment deleted successfully",
	}, nil
}

// UpdateAppointmentStatus updates the status of an appointment
func (s *AppointmentServer) UpdateAppointmentStatus(ctx context.Context, req *appointmentsv1.UpdateAppointmentStatusRequest) (*appointmentsv1.UpdateAppointmentStatusResponse, error) {
	err := s.appointmentService.UpdateAppointmentStatus(ctx, req.Id, entity.AppointmentStatus(req.Status))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update appointment status: %v", err)
	}

	return &appointmentsv1.UpdateAppointmentStatusResponse{
		Success: true,
		Message: "Appointment status updated successfully",
	}, nil
}

// ConfirmAppointment confirms an appointment
func (s *AppointmentServer) ConfirmAppointment(ctx context.Context, req *appointmentsv1.ConfirmAppointmentRequest) (*appointmentsv1.ConfirmAppointmentResponse, error) {
	err := s.appointmentService.ConfirmAppointment(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to confirm appointment: %v", err)
	}

	return &appointmentsv1.ConfirmAppointmentResponse{
		Success: true,
		Message: "Appointment confirmed successfully",
	}, nil
}

// CancelAppointment cancels an appointment
func (s *AppointmentServer) CancelAppointment(ctx context.Context, req *appointmentsv1.CancelAppointmentRequest) (*appointmentsv1.CancelAppointmentResponse, error) {
	err := s.appointmentService.CancelAppointment(ctx, req.Id, req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to cancel appointment: %v", err)
	}

	return &appointmentsv1.CancelAppointmentResponse{
		Success: true,
		Message: "Appointment cancelled successfully",
	}, nil
}

// CompleteAppointment marks an appointment as completed
func (s *AppointmentServer) CompleteAppointment(ctx context.Context, req *appointmentsv1.CompleteAppointmentRequest) (*appointmentsv1.CompleteAppointmentResponse, error) {
	err := s.appointmentService.CompleteAppointment(ctx, req.Id, req.Notes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to complete appointment: %v", err)
	}

	return &appointmentsv1.CompleteAppointmentResponse{
		Success: true,
		Message: "Appointment completed successfully",
	}, nil
}

// CheckAvailability checks availability for a time slot
func (s *AppointmentServer) CheckAvailability(ctx context.Context, req *appointmentsv1.CheckAvailabilityRequest) (*appointmentsv1.CheckAvailabilityResponse, error) {
	available, err := s.appointmentService.CheckAvailability(ctx, req.StaffId, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check availability: %v", err)
	}

	return &appointmentsv1.CheckAvailabilityResponse{
		Available: available,
	}, nil
}

// CheckConflict checks for appointment conflicts
func (s *AppointmentServer) CheckConflict(ctx context.Context, req *appointmentsv1.CheckConflictRequest) (*appointmentsv1.CheckConflictResponse, error) {
	hasConflict, err := s.appointmentService.CheckConflict(ctx, req.StaffId, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check conflict: %v", err)
	}

	return &appointmentsv1.CheckConflictResponse{
		HasConflict: hasConflict,
	}, nil
}

// GetAvailableSlots gets available time slots
func (s *AppointmentServer) GetAvailableSlots(ctx context.Context, req *appointmentsv1.GetAvailableSlotsRequest) (*appointmentsv1.GetAvailableSlotsResponse, error) {
	slots, err := s.appointmentService.GetAvailableSlots(ctx, req.StaffId, req.Date.AsTime(), int(req.Duration))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get available slots: %v", err)
	}

	// Convert time slots to protobuf
	protoSlots := make([]*appointmentsv1.TimeSlot, len(slots))
	for i, slot := range slots {
		protoSlots[i] = &appointmentsv1.TimeSlot{
			StartTime: timestamppb.New(slot.StartTime),
			EndTime:   timestamppb.New(slot.EndTime),
		}
	}

	return &appointmentsv1.GetAvailableSlotsResponse{
		Slots: protoSlots,
	}, nil
}

// GetAppointmentsByCustomer gets appointments by customer ID
func (s *AppointmentServer) GetAppointmentsByCustomer(ctx context.Context, req *appointmentsv1.GetAppointmentsByCustomerRequest) (*appointmentsv1.GetAppointmentsByCustomerResponse, error) {
	appointments, err := s.appointmentService.GetAppointmentsByCustomer(ctx, req.CustomerId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get appointments by customer: %v", err)
	}

	protoAppointments := make([]*appointmentsv1.Appointment, len(appointments))
	for i, apt := range appointments {
		protoAppointments[i] = toProtoAppointment(&apt)
	}

	return &appointmentsv1.GetAppointmentsByCustomerResponse{
		Appointments: protoAppointments,
	}, nil
}

// GetAppointmentsByEmployee gets appointments by employee ID
func (s *AppointmentServer) GetAppointmentsByEmployee(ctx context.Context, req *appointmentsv1.GetAppointmentsByEmployeeRequest) (*appointmentsv1.GetAppointmentsByEmployeeResponse, error) {
	appointments, err := s.appointmentService.GetAppointmentsByEmployee(ctx, req.StaffId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get appointments by employee: %v", err)
	}

	protoAppointments := make([]*appointmentsv1.Appointment, len(appointments))
	for i, apt := range appointments {
		protoAppointments[i] = toProtoAppointment(&apt)
	}

	return &appointmentsv1.GetAppointmentsByEmployeeResponse{
		Appointments: protoAppointments,
	}, nil
}

// GetEmployeeSchedule gets employee schedule
func (s *AppointmentServer) GetEmployeeSchedule(ctx context.Context, req *appointmentsv1.GetEmployeeScheduleRequest) (*appointmentsv1.GetEmployeeScheduleResponse, error) {
	schedule, err := s.appointmentService.GetEmployeeSchedule(ctx, req.StaffId, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get employee schedule: %v", err)
	}

	return &appointmentsv1.GetEmployeeScheduleResponse{
		Schedule: toProtoSchedule(schedule),
	}, nil
}

// GetDailySchedule gets daily schedule
func (s *AppointmentServer) GetDailySchedule(ctx context.Context, req *appointmentsv1.GetDailyScheduleRequest) (*appointmentsv1.GetDailyScheduleResponse, error) {
	schedule, err := s.appointmentService.GetDailySchedule(ctx, req.Date.AsTime())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get daily schedule: %v", err)
	}

	return &appointmentsv1.GetDailyScheduleResponse{
		Schedule: toProtoSchedule(schedule),
	}, nil
}

// BatchCreateAppointments creates multiple appointments
func (s *AppointmentServer) BatchCreateAppointments(ctx context.Context, req *appointmentsv1.BatchCreateAppointmentsRequest) (*appointmentsv1.BatchCreateAppointmentsResponse, error) {
	// Convert protobuf requests to DTO
	createReqs := make([]*dto.CreateAppointmentRequest, len(req.Appointments))
	for i, apt := range req.Appointments {
		createReqs[i] = &dto.CreateAppointmentRequest{
			CustomerID:     apt.CustomerId,
			StaffID:        apt.StaffId,
			ServiceID:      apt.ServiceId,
			StartTime:      apt.StartTime.AsTime(),
			EndTime:        apt.EndTime.AsTime(),
			Status:         entity.AppointmentStatus(apt.Status),
			Notes:          apt.Notes,
			CustomerName:   apt.CustomerName,
			StaffName:      apt.StaffName,
			ServiceName:    apt.ServiceName,
			ServiceDuration: int(apt.ServiceDuration),
			ServicePrice:    apt.ServicePrice,
		}
	}

	// Call service
	appointments, err := s.appointmentService.BatchCreateAppointments(ctx, createReqs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to batch create appointments: %v", err)
	}

	// Convert response to protobuf
	protoAppointments := make([]*appointmentsv1.Appointment, len(appointments))
	for i, apt := range appointments {
		protoAppointments[i] = toProtoAppointment(&apt)
	}

	return &appointmentsv1.BatchCreateAppointmentsResponse{
		Appointments: protoAppointments,
		Success:      true,
		Message:      fmt.Sprintf("Successfully created %d appointments", len(appointments)),
	}, nil
}

// Helper functions for converting between protobuf and internal types

func toProtoAppointment(apt *entity.Appointment) *appointmentsv1.Appointment {
	return &appointmentsv1.Appointment{
		Id:              apt.ID,
		CustomerId:      apt.CustomerID,
		StaffId:         apt.StaffID,
		ServiceId:       apt.ServiceID,
		StartTime:       timestamppb.New(apt.StartTime),
		EndTime:         timestamppb.New(apt.EndTime),
		Status:          string(apt.Status),
		Notes:           apt.Notes,
		CustomerName:    apt.CustomerName,
		StaffName:       apt.StaffName,
		ServiceName:     apt.ServiceName,
		ServiceDuration: int32(apt.ServiceDuration),
		ServicePrice:    apt.ServicePrice,
		CreatedAt:       timestamppb.New(apt.CreatedAt),
		UpdatedAt:       timestamppb.New(apt.UpdatedAt),
	}
}

func toProtoSchedule(schedule *dto.ScheduleResponse) *appointmentsv1.Schedule {
	appointments := make([]*appointmentsv1.Appointment, len(schedule.Appointments))
	for i, apt := range schedule.Appointments {
		appointments[i] = toProtoAppointment(&apt)
	}

	return &appointmentsv1.Schedule{
		Date:         timestamppb.New(schedule.Date),
		TotalCount:   int32(schedule.TotalCount),
		Appointments: appointments,
	}
}