package grpc

import (
	"context"
	"time"

	appointmentsv1 "github.com/julesChu12/fly/appointments/api/proto/appointments/v1"
	"github.com/julesChu12/fly/appointments/internal/application/service"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
	var notes *string
	if req.Notes != nil {
		notes = &req.Notes.Value
	}

	createReq := &dto.CreateAppointmentRequest{
		CustomerID:   req.CustomerId,
		StaffID:      req.StaffId,
		ServiceID:    req.ServiceId,
		StartTime:    req.StartTime.AsTime(),
		EndTime:      req.EndTime.AsTime(),
		Notes:        notes,
		Reminder:     req.Reminder,
		ReminderTime: func() *time.Time {
			if req.ReminderTime != nil {
				t := req.ReminderTime.AsTime()
				return &t
			}
			return nil
		}(),
	}

	// Call service
	appointment, err := s.appointmentService.CreateAppointment(createReq)
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
	appointment, err := s.appointmentService.GetAppointmentByID(req.Id)
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

	if req.CustomerId != nil {
		filter.CustomerID = &req.CustomerId.Value
	}
	if req.StaffId != nil {
		filter.StaffID = &req.StaffId.Value
	}
	if req.ServiceId != nil {
		filter.ServiceID = &req.ServiceId.Value
	}
	if req.Status != 0 {
		status := convertStatusFromProto(req.Status)
		filter.Status = &status
	}
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		filter.StartDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		filter.EndDate = &t
	}

	// Call service
	appointments, total, err := s.appointmentService.ListAppointments(filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list appointments: %v", err)
	}

	// Convert response
	protoAppointments := make([]*appointmentsv1.Appointment, len(appointments))
	for i, apt := range appointments {
		protoAppointments[i] = toProtoAppointment(apt)
	}

	return &appointmentsv1.ListAppointmentsResponse{
		Appointments: protoAppointments,
		Total:        total,
		Page:         int32(req.Page),
		PageSize:     int32(req.PageSize),
	}, nil
}

// UpdateAppointment updates an existing appointment
func (s *AppointmentServer) UpdateAppointment(ctx context.Context, req *appointmentsv1.UpdateAppointmentRequest) (*appointmentsv1.UpdateAppointmentResponse, error) {
	// Convert protobuf request to DTO
	updateReq := &dto.UpdateAppointmentRequest{}

	if req.CustomerId != nil {
		updateReq.CustomerID = &req.CustomerId.Value
	}
	if req.StaffId != nil {
		updateReq.StaffID = &req.StaffId.Value
	}
	if req.ServiceId != nil {
		updateReq.ServiceID = &req.ServiceId.Value
	}
	if req.StartTime != nil {
		updateReq.StartTime = func() *time.Time {
			t := req.StartTime.AsTime()
			return &t
		}()
	}
	if req.EndTime != nil {
		updateReq.EndTime = func() *time.Time {
			t := req.EndTime.AsTime()
			return &t
		}()
	}
	if req.Notes != nil {
		updateReq.Notes = &req.Notes.Value
	}
	if req.Status != 0 {
		status := string(convertStatusFromProto(req.Status))
		updateReq.Status = &status
	}
	if req.Reminder != nil {
		updateReq.Reminder = &req.Reminder.Value
	}
	if req.ReminderTime != nil {
		t := req.ReminderTime.AsTime()
		updateReq.ReminderTime = &t
	}

	// Call service
	appointment, err := s.appointmentService.UpdateAppointment(req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update appointment: %v", err)
	}

	return &appointmentsv1.UpdateAppointmentResponse{
		Appointment: toProtoAppointment(appointment),
	}, nil
}

// DeleteAppointment deletes an appointment
func (s *AppointmentServer) DeleteAppointment(ctx context.Context, req *appointmentsv1.DeleteAppointmentRequest) (*appointmentsv1.DeleteAppointmentResponse, error) {
	err := s.appointmentService.DeleteAppointment(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete appointment: %v", err)
	}

	return &appointmentsv1.DeleteAppointmentResponse{}, nil
}

// UpdateAppointmentStatus updates the status of an appointment
func (s *AppointmentServer) UpdateAppointmentStatus(ctx context.Context, req *appointmentsv1.UpdateAppointmentStatusRequest) (*appointmentsv1.UpdateAppointmentStatusResponse, error) {
	updateReq := &dto.UpdateStatusRequest{
		Status: string(convertStatusFromProto(req.Status)),
	}

	if req.CompletionNotes != nil {
		updateReq.CompletionNotes = &req.CompletionNotes.Value
	}
	// Note: CancellationReason not in UpdateStatusRequest, would need to add to DTO if needed

	appointment, err := s.appointmentService.UpdateAppointmentStatus(req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update appointment status: %v", err)
	}

	return &appointmentsv1.UpdateAppointmentStatusResponse{
		Appointment: toProtoAppointment(appointment),
	}, nil
}

// ConfirmAppointment confirms an appointment
func (s *AppointmentServer) ConfirmAppointment(ctx context.Context, req *appointmentsv1.ConfirmAppointmentRequest) (*appointmentsv1.ConfirmAppointmentResponse, error) {
	var notes *string
	if req.Notes != nil {
		notes = &req.Notes.Value
	}

	updateReq := &dto.UpdateStatusRequest{
		Status:           string(entity.AppointmentStatusConfirmed),
		CompletionNotes:  notes,
	}

	appointment, err := s.appointmentService.UpdateAppointmentStatus(req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to confirm appointment: %v", err)
	}

	return &appointmentsv1.ConfirmAppointmentResponse{
		Appointment: toProtoAppointment(appointment),
	}, nil
}

// CancelAppointment cancels an appointment
func (s *AppointmentServer) CancelAppointment(ctx context.Context, req *appointmentsv1.CancelAppointmentRequest) (*appointmentsv1.CancelAppointmentResponse, error) {
	var reason *string
	if req.Reason != nil {
		reason = &req.Reason.Value
	}

	appointment, err := s.appointmentService.CancelAppointment(req.Id, reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to cancel appointment: %v", err)
	}

	return &appointmentsv1.CancelAppointmentResponse{
		Appointment: toProtoAppointment(appointment),
	}, nil
}

// CompleteAppointment marks an appointment as completed
func (s *AppointmentServer) CompleteAppointment(ctx context.Context, req *appointmentsv1.CompleteAppointmentRequest) (*appointmentsv1.CompleteAppointmentResponse, error) {
	var notes *string
	if req.CompletionNotes != nil {
		notes = &req.CompletionNotes.Value
	}

	updateReq := &dto.UpdateStatusRequest{
		Status:          string(entity.AppointmentStatusCompleted),
		CompletionNotes: notes,
	}

	appointment, err := s.appointmentService.UpdateAppointmentStatus(req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to complete appointment: %v", err)
	}

	return &appointmentsv1.CompleteAppointmentResponse{
		Appointment: toProtoAppointment(appointment),
	}, nil
}

// CheckAvailability checks availability for a time slot
func (s *AppointmentServer) CheckAvailability(ctx context.Context, req *appointmentsv1.CheckAvailabilityRequest) (*appointmentsv1.CheckAvailabilityResponse, error) {
	availabilityReq := &dto.AvailabilityRequest{
		StaffID:         req.StaffId,
		Date:            req.Date.AsTime(),
		ServiceDuration: req.ServiceDuration.AsDuration(),
	}

	response, err := s.appointmentService.CheckAvailability(availabilityReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check availability: %v", err)
	}

	// Convert slots to protobuf
	slots := make([]*appointmentsv1.AvailableSlot, len(response.Slots))
	for i, slot := range response.Slots {
		slots[i] = &appointmentsv1.AvailableSlot{
			StartTime: timestamppb.New(slot.StartTime),
			EndTime:   timestamppb.New(slot.EndTime),
			Available: slot.Available,
		}
	}

	return &appointmentsv1.CheckAvailabilityResponse{
		StaffId:            response.StaffID,
		Date:               timestamppb.New(response.Date),
		Slots:              slots,
		HasAvailability:    len(slots) > 0,
		AvailableSlotsCount: int32(len(slots)),
	}, nil
}

// CheckConflict checks for appointment conflicts
func (s *AppointmentServer) CheckConflict(ctx context.Context, req *appointmentsv1.CheckConflictRequest) (*appointmentsv1.CheckConflictResponse, error) {
	var excludeID *string
	if req.ExcludeAppointmentId != nil {
		excludeID = &req.ExcludeAppointmentId.Value
	}

	conflictReq := &dto.ConflictCheckRequest{
		StaffID:   req.StaffId,
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
		ExcludeID: excludeID,
	}

	conflictInfo, err := s.appointmentService.CheckConflict(conflictReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check conflict: %v", err)
	}

	// Convert suggestions to protobuf
	suggestedTimes := make([]*timestamppb.Timestamp, len(conflictInfo.Suggestions))
	for i, t := range conflictInfo.Suggestions {
		suggestedTimes[i] = timestamppb.New(t)
	}

	// Convert alternative slots
	alternativeSlots := make([]*appointmentsv1.AvailableSlot, len(conflictInfo.Suggestions))
	for i, t := range conflictInfo.Suggestions {
		// Create a 1-hour slot starting at suggestion time
		alternativeSlots[i] = &appointmentsv1.AvailableSlot{
			StartTime: timestamppb.New(t),
			EndTime:   timestamppb.New(t.Add(time.Hour)),
			Available: true,
		}
	}

	return &appointmentsv1.CheckConflictResponse{
		HasConflict:               conflictInfo.Conflict,
		ConflictingAppointmentIds: conflictInfo.ConflictIDs,
		ConflictCount:             int32(conflictInfo.ConflictCount),
		SuggestedTimes:            suggestedTimes,
		AlternativeSlots:          alternativeSlots,
	}, nil
}

// GetAppointmentsByCustomer gets appointments by customer ID
func (s *AppointmentServer) GetAppointmentsByCustomer(ctx context.Context, req *appointmentsv1.GetAppointmentsByCustomerRequest) (*appointmentsv1.GetAppointmentsByCustomerResponse, error) {
	filter := &dto.AppointmentFilter{
		Page:  int(req.Page),
		Limit: int(req.PageSize),
	}

	if req.Status != 0 {
		status := convertStatusFromProto(req.Status)
		filter.Status = &status
	}
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		filter.StartDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		filter.EndDate = &t
	}

	appointments, err := s.appointmentService.GetAppointmentsByCustomerID(req.CustomerId, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get appointments by customer: %v", err)
	}

	protoAppointments := make([]*appointmentsv1.Appointment, len(appointments))
	for i, apt := range appointments {
		protoAppointments[i] = toProtoAppointment(apt)
	}

	return &appointmentsv1.GetAppointmentsByCustomerResponse{
		Appointments: protoAppointments,
		Total:        int64(len(appointments)),
		Page:         int32(req.Page),
		PageSize:     int32(req.PageSize),
	}, nil
}

// GetAppointmentsByEmployee gets appointments by employee ID
func (s *AppointmentServer) GetAppointmentsByEmployee(ctx context.Context, req *appointmentsv1.GetAppointmentsByEmployeeRequest) (*appointmentsv1.GetAppointmentsByEmployeeResponse, error) {
	filter := &dto.AppointmentFilter{
		Page:  int(req.Page),
		Limit: int(req.PageSize),
	}

	if req.Status != 0 {
		status := convertStatusFromProto(req.Status)
		filter.Status = &status
	}
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		filter.StartDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		filter.EndDate = &t
	}

	appointments, err := s.appointmentService.GetAppointmentsByStaffID(req.StaffId, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get appointments by employee: %v", err)
	}

	protoAppointments := make([]*appointmentsv1.Appointment, len(appointments))
	for i, apt := range appointments {
		protoAppointments[i] = toProtoAppointment(apt)
	}

	return &appointmentsv1.GetAppointmentsByEmployeeResponse{
		Appointments: protoAppointments,
		Total:        int64(len(appointments)),
		Page:         int32(req.Page),
		PageSize:     int32(req.PageSize),
	}, nil
}

// GetEmployeeSchedule gets employee schedule
func (s *AppointmentServer) GetEmployeeSchedule(ctx context.Context, req *appointmentsv1.GetEmployeeScheduleRequest) (*appointmentsv1.GetEmployeeScheduleResponse, error) {
	filter := &dto.AppointmentFilter{
		StaffID:  &req.StaffId,
		StartDate: func() *time.Time { t := req.StartDate.AsTime(); return &t }(),
		EndDate:   func() *time.Time { t := req.EndDate.AsTime(); return &t }(),
	}

	appointments, _, err := s.appointmentService.ListAppointments(filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get employee schedule: %v", err)
	}

	protoAppointments := make([]*appointmentsv1.Appointment, len(appointments))
	for i, apt := range appointments {
		protoAppointments[i] = toProtoAppointment(apt)
	}

	return &appointmentsv1.GetEmployeeScheduleResponse{
		StaffId:      req.StaffId,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Appointments: protoAppointments,
	}, nil
}

// GetDailySchedule gets daily schedule for multiple employees
func (s *AppointmentServer) GetDailySchedule(ctx context.Context, req *appointmentsv1.GetDailyScheduleRequest) (*appointmentsv1.GetDailyScheduleResponse, error) {
	date := req.Date.AsTime()
	employeeSchedules := make([]*appointmentsv1.DailyEmployeeSchedule, len(req.StaffIds))

	for i, staffID := range req.StaffIds {
		filter := &dto.AppointmentFilter{
			StaffID: func() *string { id := staffID; return &id }(),
			Date:    func() *time.Time { d := date; return &d }(),
		}

		appointments, _, err := s.appointmentService.ListAppointments(filter)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to get daily schedule: %v", err)
		}

		protoAppointments := make([]*appointmentsv1.Appointment, len(appointments))
		for j, apt := range appointments {
			protoAppointments[j] = toProtoAppointment(apt)
		}

		employeeSchedules[i] = &appointmentsv1.DailyEmployeeSchedule{
			StaffId:      staffID,
			Appointments: protoAppointments,
		}
	}

	return &appointmentsv1.GetDailyScheduleResponse{
		Date:              req.Date,
		EmployeeSchedules: employeeSchedules,
	}, nil
}

// BatchCreateAppointments creates multiple appointments in batch
func (s *AppointmentServer) BatchCreateAppointments(ctx context.Context, req *appointmentsv1.BatchCreateAppointmentsRequest) (*appointmentsv1.BatchCreateAppointmentsResponse, error) {
	// Convert requests
	createReqs := make([]*dto.CreateAppointmentRequest, len(req.Appointments))
	for i, apt := range req.Appointments {
		var notes *string
		if apt.Notes != nil {
			notes = &apt.Notes.Value
		}

		createReqs[i] = &dto.CreateAppointmentRequest{
			CustomerID:   apt.CustomerId,
			StaffID:      apt.StaffId,
			ServiceID:    apt.ServiceId,
			StartTime:    apt.StartTime.AsTime(),
			EndTime:      apt.EndTime.AsTime(),
			Notes:        notes,
			Reminder:     apt.Reminder,
			ReminderTime: func() *time.Time {
				if apt.ReminderTime != nil {
					t := apt.ReminderTime.AsTime()
					return &t
				}
				return nil
			}(),
		}
	}

	// For simplicity, create appointments one by one
	successfulAppointments := make([]*appointmentsv1.Appointment, 0)
	batchErrors := make([]*appointmentsv1.BatchError, 0)

	for i, createReq := range createReqs {
		appointment, err := s.appointmentService.CreateAppointment(createReq)
		if err != nil {
			if req.FailOnError {
				return nil, status.Errorf(codes.Internal, "Failed to create appointment at index %d: %v", i, err)
			}
			batchErrors = append(batchErrors, &appointmentsv1.BatchError{
				Index:        int32(i),
				ErrorMessage: err.Error(),
				Appointment:  req.Appointments[i],
			})
			continue
		}
		successfulAppointments = append(successfulAppointments, toProtoAppointment(appointment))
	}

	return &appointmentsv1.BatchCreateAppointmentsResponse{
		SuccessfulAppointments: successfulAppointments,
		Errors:                  batchErrors,
		TotalProcessed:          int32(len(req.Appointments)),
		SuccessfulCount:         int32(len(successfulAppointments)),
		FailedCount:             int32(len(batchErrors)),
	}, nil
}

// Helper functions for converting between protobuf and internal types

func toProtoAppointment(apt *entity.Appointment) *appointmentsv1.Appointment {
	var notes *wrapperspb.StringValue
	if apt.Notes != nil {
		notes = wrapperspb.String(*apt.Notes)
	}

	var reminderTime *timestamppb.Timestamp
	if apt.ReminderTime != nil {
		reminderTime = timestamppb.New(*apt.ReminderTime)
	}

	return &appointmentsv1.Appointment{
		Id:           apt.ID.String(),
		CustomerId:   apt.CustomerID.String(),
		StaffId:      apt.StaffID.String(),
		ServiceId:    apt.ServiceID.String(),
		StartTime:    timestamppb.New(apt.StartTime),
		EndTime:      timestamppb.New(apt.EndTime),
		Status:       convertStatusToProto(apt.Status),
		Notes:        notes,
		Reminder:     apt.Reminder,
		ReminderTime: reminderTime,
		CreatedAt:    timestamppb.New(apt.CreatedAt),
		UpdatedAt:    timestamppb.New(apt.UpdatedAt),
	}
}

func convertStatusToProto(status entity.AppointmentStatus) appointmentsv1.AppointmentStatus {
	switch status {
	case entity.AppointmentStatusPending:
		return appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_PENDING
	case entity.AppointmentStatusConfirmed:
		return appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_CONFIRMED
	case entity.AppointmentStatusInProgress:
		return appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_IN_PROGRESS
	case entity.AppointmentStatusCompleted:
		return appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_COMPLETED
	case entity.AppointmentStatusCancelled:
		return appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_CANCELLED
	case entity.AppointmentStatusNoShow:
		return appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_NO_SHOW
	default:
		return appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_UNSPECIFIED
	}
}

func convertStatusFromProto(status appointmentsv1.AppointmentStatus) entity.AppointmentStatus {
	switch status {
	case appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_PENDING:
		return entity.AppointmentStatusPending
	case appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_CONFIRMED:
		return entity.AppointmentStatusConfirmed
	case appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_IN_PROGRESS:
		return entity.AppointmentStatusInProgress
	case appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_COMPLETED:
		return entity.AppointmentStatusCompleted
	case appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_CANCELLED:
		return entity.AppointmentStatusCancelled
	case appointmentsv1.AppointmentStatus_APPOINTMENT_STATUS_NO_SHOW:
		return entity.AppointmentStatusNoShow
	default:
		return entity.AppointmentStatusPending
	}
}