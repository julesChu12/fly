package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAppointmentRepository is a mock implementation of AppointmentRepository
type MockAppointmentRepository struct {
	mock.Mock
}

func (m *MockAppointmentRepository) Create(appointment *entity.Appointment) error {
	args := m.Called(appointment)
	return args.Error(0)
}

func (m *MockAppointmentRepository) GetByID(id string) (*entity.Appointment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) Update(appointment *entity.Appointment) error {
	args := m.Called(appointment)
	return args.Error(0)
}

func (m *MockAppointmentRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAppointmentRepository) SoftDelete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAppointmentRepository) List(filter *dto.AppointmentFilter) ([]*entity.Appointment, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) Count(filter *dto.AppointmentFilter) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAppointmentRepository) GetByCustomerID(customerID string, filter *dto.AppointmentFilter) ([]*entity.Appointment, error) {
	args := m.Called(customerID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) GetByEmployeeID(employeeID string, filter *dto.AppointmentFilter) ([]*entity.Appointment, error) {
	args := m.Called(employeeID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) GetByDateRange(startDate, endDate time.Time, filter *dto.AppointmentFilter) ([]*entity.Appointment, error) {
	args := m.Called(startDate, endDate, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) CheckConflict(employeeID string, startTime, endTime time.Time, excludeID *string) ([]*entity.Appointment, error) {
	args := m.Called(employeeID, startTime, endTime, excludeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) GetAvailableSlots(employeeID string, date time.Time, serviceDuration time.Duration) ([]*time.Time, error) {
	args := m.Called(employeeID, date, serviceDuration)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*time.Time), args.Error(1)
}

func (m *MockAppointmentRepository) GetByStatus(status entity.AppointmentStatus, filter *dto.AppointmentFilter) ([]*entity.Appointment, error) {
	args := m.Called(status, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentRepository) UpdateStatus(id string, status entity.AppointmentStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockAppointmentRepository) GetPendingReminders(beforeTime time.Time) ([]*entity.Appointment, error) {
	args := m.Called(beforeTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func TestAppointmentService_CreateAppointment(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	now := time.Now()
	req := &dto.CreateAppointmentRequest{
		CustomerID: "550e8400-e29b-41d4-a716-446655440001",
		EmployeeID: "550e8400-e29b-41d4-a716-446655440002",
		ServiceID:  "550e8400-e29b-41d4-a716-446655440003",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
		Notes:      stringPtr("Test appointment"),
	}

	// Test successful creation
	mockRepo.On("CheckConflict", req.EmployeeID, req.StartTime, req.EndTime, (*string)(nil)).Return([]*entity.Appointment{}, nil)
	mockRepo.On("Create", mock.AnythingOfType("*entity.Appointment")).Return(nil)

	appointment, err := service.CreateAppointment(req)
	assert.NoError(t, err)
	assert.NotNil(t, appointment)
	assert.Equal(t, uuid.MustParse(req.CustomerID), appointment.CustomerID)
	assert.Equal(t, uuid.MustParse(req.EmployeeID), appointment.EmployeeID)
	assert.Equal(t, uuid.MustParse(req.ServiceID), appointment.ServiceID)
	assert.Equal(t, entity.AppointmentStatusPending, appointment.Status)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_CreateAppointment_Conflict(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	now := time.Now()
	req := &dto.CreateAppointmentRequest{
		CustomerID: "550e8400-e29b-41d4-a716-446655440001",
		EmployeeID: "550e8400-e29b-41d4-a716-446655440002",
		ServiceID:  "550e8400-e29b-41d4-a716-446655440003",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
	}

	// Test conflict detection
	existingAppointment := &entity.Appointment{
		ID:        uuid.New(),
		EmployeeID: uuid.MustParse(req.EmployeeID),
		StartTime: now,
		EndTime:   now.Add(time.Hour),
		Status:    entity.AppointmentStatusConfirmed,
	}

	mockRepo.On("CheckConflict", req.EmployeeID, req.StartTime, req.EndTime, (*string)(nil)).Return([]*entity.Appointment{existingAppointment}, nil)

	appointment, err := service.CreateAppointment(req)
	assert.Error(t, err)
	assert.Nil(t, appointment)
	assert.Contains(t, err.Error(), "时间冲突")

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_GetAppointmentByID(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	appointmentID := uuid.New().String()
	expectedAppointment := &entity.Appointment{
		ID:         uuid.MustParse(appointmentID),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		EmployeeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Status:     entity.AppointmentStatusConfirmed,
	}

	// Test successful retrieval
	mockRepo.On("GetByID", appointmentID).Return(expectedAppointment, nil)

	appointment, err := service.GetAppointmentByID(appointmentID)
	assert.NoError(t, err)
	assert.NotNil(t, appointment)
	assert.Equal(t, expectedAppointment.ID, appointment.ID)

	mockRepo.AssertExpectations(t)

	// Test not found
	mockRepo.On("GetByID", "non-existent-id").Return((*entity.Appointment)(nil), assert.AnError)

	appointment, err = service.GetAppointmentByID("non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, appointment)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_UpdateAppointment(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	appointmentID := uuid.New().String()
	existingAppointment := &entity.Appointment{
		ID:         uuid.MustParse(appointmentID),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		EmployeeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Status:     entity.AppointmentStatusPending,
	}

	req := &dto.UpdateAppointmentRequest{
		Notes: stringPtr("Updated notes"),
	}

	// Test successful update
	mockRepo.On("GetByID", appointmentID).Return(existingAppointment, nil)
	mockRepo.On("CheckConflict", existingAppointment.EmployeeID.String(), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("*string")).Return([]*entity.Appointment{}, nil)
	mockRepo.On("Update", mock.AnythingOfType("*entity.Appointment")).Return(nil)

	appointment, err := service.UpdateAppointment(appointmentID, req)
	assert.NoError(t, err)
	assert.NotNil(t, appointment)
	assert.Equal(t, "Updated notes", *appointment.Notes)

	mockRepo.AssertExpectations(t)

	// Test appointment not found
	mockRepo.On("GetByID", "non-existent-id").Return((*entity.Appointment)(nil), assert.AnError)

	appointment, err = service.UpdateAppointment("non-existent-id", req)
	assert.Error(t, err)
	assert.Nil(t, appointment)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_DeleteAppointment(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	appointmentID := uuid.New().String()

	// Test successful deletion
	mockRepo.On("SoftDelete", appointmentID).Return(nil)

	err := service.DeleteAppointment(appointmentID)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_ListAppointments(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	filter := &dto.AppointmentFilter{
		Page:  1,
		Limit: 10,
	}

	expectedAppointments := []*entity.Appointment{
		{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			EmployeeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Status:     entity.AppointmentStatusConfirmed,
		},
		{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
			EmployeeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
			Status:     entity.AppointmentStatusPending,
		},
	}

	// Test successful list
	mockRepo.On("List", mock.AnythingOfType("*dto.AppointmentFilter")).Return(expectedAppointments, nil)
	mockRepo.On("Count", mock.AnythingOfType("*dto.AppointmentFilter")).Return(int64(2), nil)

	appointments, total, err := service.ListAppointments(filter)
	assert.NoError(t, err)
	assert.Len(t, appointments, 2)
	assert.Equal(t, int64(2), total)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_UpdateAppointmentStatus(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	appointmentID := uuid.New().String()
	existingAppointment := &entity.Appointment{
		ID:         uuid.MustParse(appointmentID),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		EmployeeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Status:     entity.AppointmentStatusPending,
	}

	req := &dto.UpdateStatusRequest{
		Status: "confirmed",
	}

	// Test successful status update
	mockRepo.On("GetByID", appointmentID).Return(existingAppointment, nil)
	mockRepo.On("Update", mock.AnythingOfType("*entity.Appointment")).Return(nil)

	appointment, err := service.UpdateAppointmentStatus(appointmentID, req)
	assert.NoError(t, err)
	assert.NotNil(t, appointment)
	assert.Equal(t, entity.AppointmentStatusConfirmed, appointment.Status)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_CancelAppointment(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	appointmentID := uuid.New().String()
	existingAppointment := &entity.Appointment{
		ID:         uuid.MustParse(appointmentID),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		EmployeeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Status:     entity.AppointmentStatusPending,
	}

	reason := stringPtr("Customer requested cancellation")

	// Test successful cancellation
	mockRepo.On("GetByID", appointmentID).Return(existingAppointment, nil)
	mockRepo.On("Update", mock.AnythingOfType("*entity.Appointment")).Return(nil)

	appointment, err := service.CancelAppointment(appointmentID, reason)
	assert.NoError(t, err)
	assert.NotNil(t, appointment)
	assert.Equal(t, entity.AppointmentStatusCancelled, appointment.Status)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_GetCalendarView(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	req := &dto.CalendarViewRequest{
		StartDate: time.Now().Truncate(24 * time.Hour),
		EndDate:   time.Now().Add(7 * 24 * time.Hour).Truncate(24 * time.Hour),
	}

	expectedAppointments := []*entity.Appointment{
		{
			ID:        uuid.New(),
			EmployeeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			StartTime: req.StartDate.Add(9 * time.Hour),
			EndTime:   req.StartDate.Add(10 * time.Hour),
			Status:    entity.AppointmentStatusConfirmed,
		},
	}

	// Test successful calendar view
	mockRepo.On("GetByDateRange", req.StartDate, req.EndDate, mock.AnythingOfType("*dto.AppointmentFilter")).Return(expectedAppointments, nil)

	events, err := service.GetCalendarView(req)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, expectedAppointments[0].ID.String(), events[0].ID)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_CheckAvailability(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	date := time.Now().Truncate(24 * time.Hour)
	req := &dto.AvailabilityRequest{
		EmployeeID:      "550e8400-e29b-41d4-a716-446655440002",
		Date:            date,
		ServiceDuration: 30 * time.Minute,
	}

	expectedAppointments := []*entity.Appointment{}

	// Test availability check
	employeeID := "550e8400-e29b-41d4-a716-446655440002"
	mockRepo.On("GetByEmployeeID", employeeID, mock.AnythingOfType("*dto.AppointmentFilter")).Return(expectedAppointments, nil)

	response, err := service.CheckAvailability(req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.EmployeeID, response.EmployeeID)
	assert.Equal(t, date, response.Date)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_CheckConflict(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	req := &dto.ConflictCheckRequest{
		EmployeeID: "550e8400-e29b-41d4-a716-446655440002",
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(time.Hour),
	}

	existingAppointment := &entity.Appointment{
		ID:        uuid.New(),
		EmployeeID: uuid.MustParse(req.EmployeeID),
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    entity.AppointmentStatusConfirmed,
	}

	// Test conflict detection
	mockRepo.On("CheckConflict", req.EmployeeID, req.StartTime, req.EndTime, (*string)(nil)).Return([]*entity.Appointment{existingAppointment}, nil)

	response, err := service.CheckConflict(req)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Conflict)
	assert.Equal(t, 1, response.ConflictCount)

	mockRepo.AssertExpectations(t)

	// Test no conflict - create a new mock service instance to avoid state issues
	mockRepo2 := &MockAppointmentRepository{}
	service2 := NewAppointmentService(mockRepo2)

	mockRepo2.On("CheckConflict", req.EmployeeID, req.StartTime, req.EndTime, (*string)(nil)).Return([]*entity.Appointment{}, nil)

	response, err = service2.CheckConflict(req)
	assert.NoError(t, err)
	assert.False(t, response.Conflict)
	assert.Equal(t, 0, response.ConflictCount)

	mockRepo2.AssertExpectations(t)
}

func TestAppointmentService_GetAppointmentsByCustomerID(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	customerID := "550e8400-e29b-41d4-a716-446655440001"
	filter := &dto.AppointmentFilter{
		Page:  1,
		Limit: 10,
	}

	expectedAppointments := []*entity.Appointment{
		{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse(customerID),
			EmployeeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Status:     entity.AppointmentStatusConfirmed,
		},
	}

	// Test successful retrieval by customer ID
	mockRepo.On("GetByCustomerID", customerID, mock.AnythingOfType("*dto.AppointmentFilter")).Return(expectedAppointments, nil)

	appointments, err := service.GetAppointmentsByCustomerID(customerID, filter)
	assert.NoError(t, err)
	assert.Len(t, appointments, 1)
	assert.Equal(t, uuid.MustParse(customerID), appointments[0].CustomerID)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_GetAppointmentsByEmployeeID(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	employeeID := "550e8400-e29b-41d4-a716-446655440002"
	filter := &dto.AppointmentFilter{
		Page:  1,
		Limit: 10,
	}

	expectedAppointments := []*entity.Appointment{
		{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			EmployeeID: uuid.MustParse(employeeID),
			Status:     entity.AppointmentStatusConfirmed,
		},
	}

	// Test successful retrieval by employee ID
	mockRepo.On("GetByEmployeeID", employeeID, mock.AnythingOfType("*dto.AppointmentFilter")).Return(expectedAppointments, nil)

	appointments, err := service.GetAppointmentsByEmployeeID(employeeID, filter)
	assert.NoError(t, err)
	assert.Len(t, appointments, 1)
	assert.Equal(t, uuid.MustParse(employeeID), appointments[0].EmployeeID)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_GetUpcomingAppointments(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	employeeID := "550e8400-e29b-41d4-a716-446655440002"
	limit := 10

	expectedAppointments := []*entity.Appointment{
		{
			ID:         uuid.New(),
			EmployeeID: uuid.MustParse(employeeID),
			Status:     entity.AppointmentStatusConfirmed,
			StartTime:  time.Now().Add(time.Hour),
		},
	}

	// Test successful retrieval of upcoming appointments
	mockRepo.On("List", mock.AnythingOfType("*dto.AppointmentFilter")).Return(expectedAppointments, nil)

	appointments, err := service.GetUpcomingAppointments(employeeID, limit)
	assert.NoError(t, err)
	assert.Len(t, appointments, 1)
	assert.Equal(t, uuid.MustParse(employeeID), appointments[0].EmployeeID)

	mockRepo.AssertExpectations(t)
}

func TestAppointmentService_GetPendingReminders(t *testing.T) {
	mockRepo := &MockAppointmentRepository{}
	service := NewAppointmentService(mockRepo)

	beforeTime := time.Now()

	expectedAppointments := []*entity.Appointment{
		{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			EmployeeID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Status:     entity.AppointmentStatusConfirmed,
		},
	}

	// Test successful retrieval of pending reminders
	mockRepo.On("GetPendingReminders", beforeTime).Return(expectedAppointments, nil)

	appointments, err := service.GetPendingReminders(beforeTime)
	assert.NoError(t, err)
	assert.Len(t, appointments, 1)

	mockRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}