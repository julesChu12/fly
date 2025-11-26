package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/staff/internal/domain/entity"
	"github.com/julesChu12/fly/staff/internal/domain/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStaffRepository is a mock implementation of StaffRepository
type MockStaffRepository struct {
	mock.Mock
}

func (m *MockStaffRepository) Create(staff *entity.Staff) error {
	args := m.Called(staff)
	return args.Error(0)
}

func (m *MockStaffRepository) GetByID(id string) (*entity.Staff, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Staff), args.Error(1)
}

func (m *MockStaffRepository) Update(staff *entity.Staff) error {
	args := m.Called(staff)
	return args.Error(0)
}

func (m *MockStaffRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStaffRepository) SoftDelete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStaffRepository) List(filter *dto.StaffFilter) ([]*entity.Staff, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Staff), args.Error(1)
}

func (m *MockStaffRepository) Count(filter *dto.StaffFilter) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStaffRepository) GetByEmail(email string) (*entity.Staff, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Staff), args.Error(1)
}

func (m *MockStaffRepository) GetByPhone(phone string) (*entity.Staff, error) {
	args := m.Called(phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Staff), args.Error(1)
}

func (m *MockStaffRepository) GetByUserID(userID string) (*entity.Staff, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Staff), args.Error(1)
}

func (m *MockStaffRepository) GetByRoleID(roleID string, filter *dto.StaffFilter) ([]*entity.Staff, error) {
	args := m.Called(roleID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Staff), args.Error(1)
}

func (m *MockStaffRepository) GetByDepartment(department string, filter *dto.StaffFilter) ([]*entity.Staff, error) {
	args := m.Called(department, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Staff), args.Error(1)
}

func (m *MockStaffRepository) GetAvailableStaff(filter *dto.StaffFilter) ([]*entity.Staff, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Staff), args.Error(1)
}

func (m *MockStaffRepository) UpdateStatus(id string, status entity.StaffStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockStaffRepository) GetByStatus(status entity.StaffStatus, filter *dto.StaffFilter) ([]*entity.Staff, error) {
	args := m.Called(status, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Staff), args.Error(1)
}

func (m *MockStaffRepository) GetStats() (*dto.StaffStats, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StaffStats), args.Error(1)
}

func (m *MockStaffRepository) GetDepartmentStats() ([]*dto.DepartmentStats, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.DepartmentStats), args.Error(1)
}

func (m *MockStaffRepository) GetRoleStats() ([]*dto.RoleStats, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.RoleStats), args.Error(1)
}

func (m *MockStaffRepository) ExistByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

// MockStaffRoleRepository is a mock implementation of StaffRoleRepository
type MockStaffRoleRepository struct {
	mock.Mock
}

func (m *MockStaffRoleRepository) Create(role *entity.StaffRole) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockStaffRoleRepository) GetByID(id string) (*entity.StaffRole, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StaffRole), args.Error(1)
}

func (m *MockStaffRoleRepository) Update(role *entity.StaffRole) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockStaffRoleRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStaffRoleRepository) SoftDelete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStaffRoleRepository) List(filter *dto.RoleFilter) ([]*entity.StaffRole, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StaffRole), args.Error(1)
}

func (m *MockStaffRoleRepository) Count(filter *dto.RoleFilter) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStaffRoleRepository) GetByCode(code string) (*entity.StaffRole, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StaffRole), args.Error(1)
}

func (m *MockStaffRoleRepository) GetByName(name string) (*entity.StaffRole, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StaffRole), args.Error(1)
}

func (m *MockStaffRoleRepository) GetDefaultRole() (*entity.StaffRole, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StaffRole), args.Error(1)
}

func (m *MockStaffRoleRepository) GetActiveRoles() ([]*entity.StaffRole, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StaffRole), args.Error(1)
}

func (m *MockStaffRoleRepository) UpdateStatus(id string, status entity.StaffStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockStaffRoleRepository) ExistByName(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

// MockStaffAvailabilityRepository is a mock implementation of StaffAvailabilityRepository
type MockStaffAvailabilityRepository struct {
	mock.Mock
}

func (m *MockStaffAvailabilityRepository) Create(availability *entity.StaffAvailability) error {
	args := m.Called(availability)
	return args.Error(0)
}

func (m *MockStaffAvailabilityRepository) GetByID(id string) (*entity.StaffAvailability, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StaffAvailability), args.Error(1)
}

func (m *MockStaffAvailabilityRepository) Update(availability *entity.StaffAvailability) error {
	args := m.Called(availability)
	return args.Error(0)
}

func (m *MockStaffAvailabilityRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStaffAvailabilityRepository) GetByStaffID(staffID string) ([]*entity.StaffAvailability, error) {
	args := m.Called(staffID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StaffAvailability), args.Error(1)
}

func (m *MockStaffAvailabilityRepository) GetByStaffIDAndDay(staffID string, dayOfWeek int) (*entity.StaffAvailability, error) {
	args := m.Called(staffID, dayOfWeek)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StaffAvailability), args.Error(1)
}

func (m *MockStaffAvailabilityRepository) GetAvailableStaff(dateTime time.Time) ([]*entity.Staff, error) {
	args := m.Called(dateTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Staff), args.Error(1)
}

func (m *MockStaffAvailabilityRepository) BatchCreate(availabilities []*entity.StaffAvailability) error {
	args := m.Called(availabilities)
	return args.Error(0)
}

func (m *MockStaffAvailabilityRepository) BatchUpdate(availabilities []*entity.StaffAvailability) error {
	args := m.Called(availabilities)
	return args.Error(0)
}

func (m *MockStaffAvailabilityRepository) DeleteByStaffID(staffID string) error {
	args := m.Called(staffID)
	return args.Error(0)
}

func TestNewStaffService(t *testing.T) {
	mockStaffRepo := &MockStaffRepository{}
	mockRoleRepo := &MockStaffRoleRepository{}
	mockAvailabilityRepo := &MockStaffAvailabilityRepository{}

	service := NewStaffService(mockStaffRepo, mockRoleRepo, mockAvailabilityRepo)

	assert.NotNil(t, service)
}

func TestStaffService_CreateStaff(t *testing.T) {
	mockStaffRepo := &MockStaffRepository{}
	mockRoleRepo := &MockStaffRoleRepository{}
	mockAvailabilityRepo := &MockStaffAvailabilityRepository{}

	service := NewStaffService(mockStaffRepo, mockRoleRepo, mockAvailabilityRepo)

	roleID := uuid.New().String()
	req := &dto.CreateStaffRequest{
		Name:       "John Doe",
		Email:      "john@example.com",
		Department: "Engineering",
		Position:   "Developer",
		RoleID:     roleID,
		Skills:     []string{"Go", "Docker"},
		IsAvailable: true,
	}

	// Mock calls
	mockStaffRepo.On("ExistByEmail", req.Email).Return(false, nil)
	mockRoleRepo.On("GetByID", roleID).Return(&entity.StaffRole{
		ID:   uuid.MustParse(roleID),
		Name: "Developer",
	}, nil)
	mockStaffRepo.On("Create", mock.AnythingOfType("*entity.Staff")).Return(nil)
	mockStaffRepo.On("GetByID", mock.AnythingOfType("string")).Return(&entity.Staff{
		ID:         uuid.New(),
		Name:       req.Name,
		Email:      req.Email,
		Department: req.Department,
		Position:   req.Position,
	}, nil)

	staff, err := service.CreateStaff(req)

	assert.NoError(t, err)
	assert.NotNil(t, staff)
	assert.Equal(t, req.Name, staff.Name)
	assert.Equal(t, req.Email, staff.Email)

	mockStaffRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestStaffService_CreateStaff_EmailExists(t *testing.T) {
	mockStaffRepo := &MockStaffRepository{}
	mockRoleRepo := &MockStaffRoleRepository{}
	mockAvailabilityRepo := &MockStaffAvailabilityRepository{}

	service := NewStaffService(mockStaffRepo, mockRoleRepo, mockAvailabilityRepo)

	req := &dto.CreateStaffRequest{
		Name:       "John Doe",
		Email:      "john@example.com",
		Department: "Engineering",
		Position:   "Developer",
		RoleID:     uuid.New().String(),
	}

	mockStaffRepo.On("ExistByEmail", req.Email).Return(true, nil)

	_, err := service.CreateStaff(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "邮箱已存在")

	mockStaffRepo.AssertExpectations(t)
}

func TestStaffService_GetAvailability(t *testing.T) {
	mockStaffRepo := &MockStaffRepository{}
	mockRoleRepo := &MockStaffRoleRepository{}
	mockAvailabilityRepo := &MockStaffAvailabilityRepository{}

	service := NewStaffService(mockStaffRepo, mockRoleRepo, mockAvailabilityRepo)

	staffID := uuid.New().String()

	// Mock calls
	mockStaffRepo.On("GetByID", staffID).Return(&entity.Staff{
		ID:   uuid.MustParse(staffID),
		Name: "John Doe",
	}, nil)

	availabilities := []*entity.StaffAvailability{
		{
			ID:         uuid.New(),
			StaffID:    uuid.MustParse(staffID),
			DayOfWeek:  1,
			StartTime:  "09:00",
			EndTime:    "17:00",
			IsAvailable: true,
		},
	}

	mockAvailabilityRepo.On("GetByStaffID", staffID).Return(availabilities, nil)

	response, err := service.GetAvailability(staffID)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, staffID, response.StaffID)
	assert.Len(t, response.Availabilities, 1)

	mockStaffRepo.AssertExpectations(t)
	mockAvailabilityRepo.AssertExpectations(t)
}

func TestStaffService_SetAvailability(t *testing.T) {
	mockStaffRepo := &MockStaffRepository{}
	mockRoleRepo := &MockStaffRoleRepository{}
	mockAvailabilityRepo := &MockStaffAvailabilityRepository{}

	service := NewStaffService(mockStaffRepo, mockRoleRepo, mockAvailabilityRepo)

	staffID := uuid.New().String()
	req := &dto.AvailabilityRequest{
		StaffID: staffID,
		Availabilities: []dto.AvailabilityItem{
			{
				DayOfWeek:   1,
				StartTime:   "09:00",
				EndTime:     "17:00",
				IsAvailable: true,
			},
		},
	}

	// Mock calls
	mockStaffRepo.On("GetByID", staffID).Return(&entity.Staff{
		ID:   uuid.MustParse(staffID),
		Name: "John Doe",
	}, nil)
	mockAvailabilityRepo.On("DeleteByStaffID", staffID).Return(nil)
	mockAvailabilityRepo.On("BatchCreate", mock.AnythingOfType("[]*entity.StaffAvailability")).Return(nil)

	err := service.SetAvailability(staffID, req)

	assert.NoError(t, err)

	mockStaffRepo.AssertExpectations(t)
	mockAvailabilityRepo.AssertExpectations(t)
}