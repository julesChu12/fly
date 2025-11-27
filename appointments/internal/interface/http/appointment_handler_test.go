package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAppointmentService is a mock implementation of AppointmentService
type MockAppointmentService struct {
	mock.Mock
}

func (m *MockAppointmentService) CreateAppointment(req *dto.CreateAppointmentRequest) (*entity.Appointment, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentService) GetAppointmentByID(id string) (*entity.Appointment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentService) UpdateAppointment(id string, req *dto.UpdateAppointmentRequest) (*entity.Appointment, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentService) DeleteAppointment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAppointmentService) ListAppointments(filter *dto.AppointmentFilter) ([]*entity.Appointment, int64, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, int64(0), args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Get(1).(int64), args.Error(2)
}

func (m *MockAppointmentService) UpdateAppointmentStatus(id string, req *dto.UpdateStatusRequest) (*entity.Appointment, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentService) CancelAppointment(id string, reason *string) (*entity.Appointment, error) {
	args := m.Called(id, reason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentService) CheckAvailability(req *dto.AvailabilityRequest) (*dto.AvailabilityResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AvailabilityResponse), args.Error(1)
}

func (m *MockAppointmentService) CheckConflict(req *dto.ConflictCheckRequest) (*dto.ConflictInfo, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ConflictInfo), args.Error(1)
}

func (m *MockAppointmentService) GetCalendarView(req *dto.CalendarViewRequest) ([]*dto.CalendarEvent, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.CalendarEvent), args.Error(1)
}

func (m *MockAppointmentService) GetAppointmentsByCustomerID(customerID string, filter *dto.AppointmentFilter) ([]*entity.Appointment, error) {
	args := m.Called(customerID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentService) GetAppointmentsByStaffID(employeeID string, filter *dto.AppointmentFilter) ([]*entity.Appointment, error) {
	args := m.Called(employeeID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentService) GetUpcomingAppointments(employeeID string, limit int) ([]*entity.Appointment, error) {
	args := m.Called(employeeID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func (m *MockAppointmentService) GetPendingReminders(beforeTime time.Time) ([]*entity.Appointment, error) {
	args := m.Called(beforeTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Appointment), args.Error(1)
}

func setupTestRouter() (*gin.Engine, *MockAppointmentService) {
	gin.SetMode(gin.TestMode)
	mockService := &MockAppointmentService{}
	handler := NewAppointmentHandler(mockService)

	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	return router, mockService
}

func TestAppointmentHandler_CreateAppointment(t *testing.T) {
	router, mockService := setupTestRouter()

	now := time.Now()
	req := dto.CreateAppointmentRequest{
		CustomerID: "550e8400-e29b-41d4-a716-446655440001",
		StaffID: "550e8400-e29b-41d4-a716-446655440002",
		ServiceID:  "550e8400-e29b-41d4-a716-446655440003",
		StartTime:  now,
		EndTime:    now.Add(time.Hour),
		Notes:      stringPtr("Test appointment"),
	}

	expectedAppointment := &entity.Appointment{
		ID:         uuid.New(),
		CustomerID: uuid.MustParse(req.CustomerID),
		StaffID: uuid.MustParse(req.StaffID),
		ServiceID:  uuid.MustParse(req.ServiceID),
		Status:     entity.AppointmentStatusPending,
	}

	body, _ := json.Marshal(req)
	mockService.On("CreateAppointment", &req).Return(expectedAppointment, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/v1/appointments", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(201), response["code"])
	assert.Equal(t, "创建成功", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_GetAppointment(t *testing.T) {
	router, mockService := setupTestRouter()

	appointmentID := uuid.New().String()
	expectedAppointment := &entity.Appointment{
		ID:         uuid.MustParse(appointmentID),
		CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Status:     entity.AppointmentStatusConfirmed,
	}

	mockService.On("GetAppointmentByID", appointmentID).Return(expectedAppointment, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/appointments/"+appointmentID, nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "获取成功", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_GetAppointment_NotFound(t *testing.T) {
	router, mockService := setupTestRouter()

	appointmentID := uuid.New().String()
	mockService.On("GetAppointmentByID", appointmentID).Return((*entity.Appointment)(nil), assert.AnError)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/appointments/"+appointmentID, nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(404), response["code"])
	assert.Equal(t, "预约不存在", response["message"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_ListAppointments(t *testing.T) {
	router, mockService := setupTestRouter()

	expectedAppointments := []*entity.Appointment{
		{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Status:     entity.AppointmentStatusConfirmed,
		},
	}

	mockService.On("ListAppointments", mock.AnythingOfType("*dto.AppointmentFilter")).Return(expectedAppointments, int64(1), nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/appointments?page=1&limit=20", nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "获取成功", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["total"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(20), data["limit"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_UpdateStatus(t *testing.T) {
	router, mockService := setupTestRouter()

	appointmentID := uuid.New().String()
	req := dto.UpdateStatusRequest{
		Status: "confirmed",
	}

	expectedAppointment := &entity.Appointment{
		ID:     uuid.MustParse(appointmentID),
		Status: entity.AppointmentStatusConfirmed,
	}

	body, _ := json.Marshal(req)
	mockService.On("UpdateAppointmentStatus", appointmentID, &req).Return(expectedAppointment, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/api/v1/appointments/"+appointmentID+"/status", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "更新成功", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_GetCalendarView(t *testing.T) {
	router, mockService := setupTestRouter()

	startDate := time.Now().Truncate(24 * time.Hour).Format("2006-01-02")
	endDate := time.Now().Add(7 * 24 * time.Hour).Truncate(24 * time.Hour).Format("2006-01-02")

	expectedEvents := []*dto.CalendarEvent{
		{
			ID:        uuid.New().String(),
			StartTime: time.Now().Add(9 * time.Hour),
			EndTime:   time.Now().Add(10 * time.Hour),
		},
	}

	mockService.On("GetCalendarView", mock.AnythingOfType("*dto.CalendarViewRequest")).Return(expectedEvents, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/appointments/calendar?start_date="+startDate+"&end_date="+endDate, nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "获取成功", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_CheckAvailability(t *testing.T) {
	router, mockService := setupTestRouter()

	date := time.Now().Truncate(24 * time.Hour).Format("2006-01-02")

	expectedResponse := &dto.AvailabilityResponse{
		StaffID: "550e8400-e29b-41d4-a716-446655440002",
		Date:       time.Now().Truncate(24 * time.Hour),
		Slots: []dto.AvailableSlot{
			{
				StartTime: time.Now().Add(9 * time.Hour),
				EndTime:   time.Now().Add(9*time.Hour + 30*time.Minute),
			},
		},
	}

	mockService.On("CheckAvailability", mock.AnythingOfType("*dto.AvailabilityRequest")).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/appointments/availability?staff_id=550e8400-e29b-41d4-a716-446655440002&date="+date+"&service_duration=30", nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "获取成功", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_CheckConflict(t *testing.T) {
	router, mockService := setupTestRouter()

	req := dto.ConflictCheckRequest{
		StaffID: "550e8400-e29b-41d4-a716-446655440002",
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(time.Hour),
	}

	expectedResponse := &dto.ConflictInfo{
		Conflict:      true,
		ConflictCount: 1,
		ConflictIDs:   []string{uuid.New().String()},
	}

	body, _ := json.Marshal(req)
	mockService.On("CheckConflict", &req).Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/v1/appointments/conflict-check", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "检查成功", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_GetAppointmentsByCustomer(t *testing.T) {
	router, mockService := setupTestRouter()

	customerID := "550e8400-e29b-41d4-a716-446655440001"
	expectedAppointments := []*entity.Appointment{
		{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse(customerID),
			StaffID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Status:     entity.AppointmentStatusConfirmed,
		},
	}

	mockService.On("GetAppointmentsByCustomerID", customerID, mock.AnythingOfType("*dto.AppointmentFilter")).Return(expectedAppointments, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/appointments/customer/"+customerID, nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "获取成功", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_GetAppointmentsByEmployee(t *testing.T) {
	router, mockService := setupTestRouter()

	employeeID := "550e8400-e29b-41d4-a716-446655440002"
	expectedAppointments := []*entity.Appointment{
		{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			StaffID: uuid.MustParse(employeeID),
			Status:     entity.AppointmentStatusConfirmed,
		},
	}

	mockService.On("GetAppointmentsByStaffID", employeeID, mock.AnythingOfType("*dto.AppointmentFilter")).Return(expectedAppointments, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/appointments/employee/"+employeeID, nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "获取成功", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func TestAppointmentHandler_GetUpcomingAppointments(t *testing.T) {
	router, mockService := setupTestRouter()

	employeeID := "550e8400-e29b-41d4-a716-446655440002"
	expectedAppointments := []*entity.Appointment{
		{
			ID:         uuid.New(),
			CustomerID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			StaffID: uuid.MustParse(employeeID),
			Status:     entity.AppointmentStatusConfirmed,
			StartTime:  time.Now().Add(time.Hour),
		},
	}

	mockService.On("GetUpcomingAppointments", employeeID, 10).Return(expectedAppointments, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/appointments/employee/"+employeeID+"/upcoming?limit=10", nil)
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "获取成功", response["message"])
	assert.NotNil(t, response["data"])

	mockService.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}