package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ListWithFilter(ctx context.Context, filter *repository.UserListFilter, limit, offset int) ([]*entity.User, int64, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountByRole(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockUserRepository) CountByType(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockUserRepository) CountNewUsers(ctx context.Context, since string) (int64, error) {
	args := m.Called(ctx, since)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetByIDWithTenant(ctx context.Context, id uint, tenantID uint) (*entity.User, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (*entity.User, error) {
	args := m.Called(ctx, username, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmailWithTenant(ctx context.Context, email string, tenantID uint) (*entity.User, error) {
	args := m.Called(ctx, email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) ListByTenant(ctx context.Context, tenantID uint, limit, offset int) ([]*entity.User, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsernameWithTenant(ctx context.Context, username string, tenantID uint) (bool, error) {
	args := m.Called(ctx, username, tenantID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmailWithTenant(ctx context.Context, email string, tenantID uint) (bool, error) {
	args := m.Called(ctx, email, tenantID)
	return args.Bool(0), args.Error(1)
}

type MockUserProfileRepository struct {
	mock.Mock
}

func (m *MockUserProfileRepository) Create(ctx context.Context, profile *entity.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockUserProfileRepository) GetByUserID(ctx context.Context, userID uint) (*entity.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) Update(ctx context.Context, profile *entity.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockUserProfileRepository) Delete(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserProfileRepository) Exists(ctx context.Context, userID uint) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// Test fixtures
func createTestUser() *entity.User {
	return &entity.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}
}

func createTestProfile() *entity.UserProfile {
	birthday := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	return &entity.UserProfile{
		UserID:   1,
		Nickname: "Test User",
		Avatar:   "https://example.com/avatar.jpg",
		Gender:   "male",
		Birthday: &birthday,
		Extra:    `{"interests": ["coding", "reading"]}`,
	}
}

// Tests for GetProfile
func TestUserProfileService_GetProfile_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(1)
	testUser := createTestUser()
	testProfile := createTestProfile()

	mockUserRepo.On("GetByID", ctx, userID).Return(testUser, nil)
	mockProfileRepo.On("GetByUserID", ctx, userID).Return(testProfile, nil)

	response, err := service.GetProfile(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, testUser.Username, response.Username)
	assert.Equal(t, testUser.Email, response.Email)
	assert.Equal(t, testProfile.Nickname, response.Nickname)
	assert.Equal(t, testProfile.Avatar, response.Avatar)
	assert.Equal(t, testProfile.Gender, response.Gender)
	assert.Equal(t, "1990-01-01", response.Birthday)

	mockUserRepo.AssertExpectations(t)
	mockProfileRepo.AssertExpectations(t)
}

func TestUserProfileService_GetProfile_AutoCreateProfile(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(1)
	testUser := createTestUser()

	mockUserRepo.On("GetByID", ctx, userID).Return(testUser, nil)
	mockProfileRepo.On("GetByUserID", ctx, userID).Return(nil, repository.ErrUserProfileNotFound)
	mockProfileRepo.On("Create", ctx, mock.AnythingOfType("*entity.UserProfile")).Return(nil)

	response, err := service.GetProfile(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, userID, response.UserID)
	assert.Equal(t, "other", response.Gender) // Default gender

	mockUserRepo.AssertExpectations(t)
	mockProfileRepo.AssertExpectations(t)
}

func TestUserProfileService_GetProfile_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(999)

	mockUserRepo.On("GetByID", ctx, userID).Return(nil, repository.ErrUserNotFound)

	response, err := service.GetProfile(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "user not found")

	mockUserRepo.AssertExpectations(t)
}

// Tests for UpdateProfile
func TestUserProfileService_UpdateProfile_ExistingProfile(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(1)
	testUser := createTestUser()
	testProfile := createTestProfile()

	req := &dto.UpdateProfileRequest{
		Nickname: "Updated Name",
		Avatar:   "https://example.com/new-avatar.jpg",
		Gender:   "female",
		Birthday: "1995-05-05",
	}

	mockUserRepo.On("GetByID", ctx, userID).Return(testUser, nil)
	mockProfileRepo.On("GetByUserID", ctx, userID).Return(testProfile, nil)
	mockProfileRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserProfile")).Return(nil)

	err := service.UpdateProfile(ctx, userID, req)

	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockProfileRepo.AssertExpectations(t)
}

func TestUserProfileService_UpdateProfile_CreateNewProfile(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(1)
	testUser := createTestUser()

	req := &dto.UpdateProfileRequest{
		Nickname: "New User",
		Gender:   "other",
	}

	mockUserRepo.On("GetByID", ctx, userID).Return(testUser, nil)
	mockProfileRepo.On("GetByUserID", ctx, userID).Return(nil, repository.ErrUserProfileNotFound)
	mockProfileRepo.On("Create", ctx, mock.AnythingOfType("*entity.UserProfile")).Return(nil)

	err := service.UpdateProfile(ctx, userID, req)

	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockProfileRepo.AssertExpectations(t)
}

func TestUserProfileService_UpdateProfile_InvalidBirthdayFormat(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(1)
	testUser := createTestUser()
	testProfile := createTestProfile()

	req := &dto.UpdateProfileRequest{
		Birthday: "invalid-date",
	}

	mockUserRepo.On("GetByID", ctx, userID).Return(testUser, nil)
	mockProfileRepo.On("GetByUserID", ctx, userID).Return(testProfile, nil)

	err := service.UpdateProfile(ctx, userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid birthday format")

	mockUserRepo.AssertExpectations(t)
	mockProfileRepo.AssertExpectations(t)
}

// Tests for DeleteProfile
func TestUserProfileService_DeleteProfile_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(1)

	mockProfileRepo.On("Exists", ctx, userID).Return(true, nil)
	mockProfileRepo.On("Delete", ctx, userID).Return(nil)

	err := service.DeleteProfile(ctx, userID)

	assert.NoError(t, err)

	mockProfileRepo.AssertExpectations(t)
}

func TestUserProfileService_DeleteProfile_NotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(999)

	mockProfileRepo.On("Exists", ctx, userID).Return(false, nil)

	err := service.DeleteProfile(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, repository.ErrUserProfileNotFound, err)

	mockProfileRepo.AssertExpectations(t)
}

// Tests for CreateProfile
func TestUserProfileService_CreateProfile_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(1)
	testUser := createTestUser()

	req := &dto.UpdateProfileRequest{
		Nickname: "New User",
		Gender:   "male",
	}

	mockUserRepo.On("GetByID", ctx, userID).Return(testUser, nil)
	mockProfileRepo.On("Exists", ctx, userID).Return(false, nil)
	mockProfileRepo.On("Create", ctx, mock.AnythingOfType("*entity.UserProfile")).Return(nil)

	err := service.CreateProfile(ctx, userID, req)

	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockProfileRepo.AssertExpectations(t)
}

func TestUserProfileService_CreateProfile_AlreadyExists(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(1)
	testUser := createTestUser()

	req := &dto.UpdateProfileRequest{
		Nickname: "New User",
	}

	mockUserRepo.On("GetByID", ctx, userID).Return(testUser, nil)
	mockProfileRepo.On("Exists", ctx, userID).Return(true, nil)

	err := service.CreateProfile(ctx, userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profile already exists")

	mockUserRepo.AssertExpectations(t)
	mockProfileRepo.AssertExpectations(t)
}

func TestUserProfileService_CreateProfile_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProfileRepo := new(MockUserProfileRepository)
	service := NewUserProfileService(mockUserRepo, mockProfileRepo)

	ctx := context.Background()
	userID := uint(999)

	req := &dto.UpdateProfileRequest{
		Nickname: "New User",
	}

	mockUserRepo.On("GetByID", ctx, userID).Return(nil, repository.ErrUserNotFound)

	err := service.CreateProfile(ctx, userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")

	mockUserRepo.AssertExpectations(t)
}

// Tests for applyProfileUpdates
func TestUserProfileService_ApplyProfileUpdates_AllFields(t *testing.T) {
	service := &UserProfileService{}
	profile := entity.NewUserProfile(1)

	req := &dto.UpdateProfileRequest{
		Nickname: "Test User",
		Avatar:   "https://example.com/avatar.jpg",
		Gender:   "female",
		Birthday: "1990-01-01",
		Extra:    `{"key": "value"}`,
	}

	err := service.applyProfileUpdates(profile, req)

	assert.NoError(t, err)
	assert.Equal(t, "Test User", profile.Nickname)
	assert.Equal(t, "https://example.com/avatar.jpg", profile.Avatar)
	assert.Equal(t, "female", profile.Gender)
	assert.NotNil(t, profile.Birthday)
	assert.Equal(t, `{"key": "value"}`, profile.Extra)
}

func TestUserProfileService_ApplyProfileUpdates_PartialFields(t *testing.T) {
	service := &UserProfileService{}
	profile := createTestProfile()

	originalNickname := profile.Nickname
	req := &dto.UpdateProfileRequest{
		Avatar: "https://example.com/new-avatar.jpg",
	}

	err := service.applyProfileUpdates(profile, req)

	assert.NoError(t, err)
	assert.Equal(t, originalNickname, profile.Nickname) // Should remain unchanged
	assert.Equal(t, "https://example.com/new-avatar.jpg", profile.Avatar)
}

func TestUserProfileService_ApplyProfileUpdates_InvalidBirthday(t *testing.T) {
	service := &UserProfileService{}
	profile := entity.NewUserProfile(1)

	req := &dto.UpdateProfileRequest{
		Birthday: "invalid-date",
	}

	err := service.applyProfileUpdates(profile, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid birthday format")
}

// Test entity methods
func TestUserProfile_SetGender_Valid(t *testing.T) {
	profile := entity.NewUserProfile(1)

	tests := []struct {
		name   string
		gender string
	}{
		{"Male", "male"},
		{"Female", "female"},
		{"Other", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := profile.SetGender(tt.gender)
			assert.NoError(t, err)
			assert.Equal(t, tt.gender, profile.Gender)
		})
	}
}

func TestUserProfile_SetGender_Invalid(t *testing.T) {
	profile := entity.NewUserProfile(1)

	err := profile.SetGender("invalid")
	assert.Error(t, err)
	assert.Equal(t, entity.ErrInvalidGender, err)
}

func TestUserProfile_SetExtra_Success(t *testing.T) {
	profile := entity.NewUserProfile(1)
	data := map[string]interface{}{
		"interests": []string{"coding", "reading"},
		"age":       30,
	}

	err := profile.SetExtra(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, profile.Extra)
}

func TestUserProfile_GetExtra_Success(t *testing.T) {
	profile := entity.NewUserProfile(1)
	originalData := map[string]interface{}{
		"interests": []string{"coding", "reading"},
	}

	err := profile.SetExtra(originalData)
	assert.NoError(t, err)

	var retrievedData map[string]interface{}
	err = profile.GetExtra(&retrievedData)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedData["interests"])
}

func TestUserProfile_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		profile  *entity.UserProfile
		expected bool
	}{
		{
			name: "Complete profile",
			profile: &entity.UserProfile{
				Nickname: "Test User",
				Avatar:   "https://example.com/avatar.jpg",
			},
			expected: true,
		},
		{
			name: "Missing nickname",
			profile: &entity.UserProfile{
				Avatar: "https://example.com/avatar.jpg",
			},
			expected: false,
		},
		{
			name: "Missing avatar",
			profile: &entity.UserProfile{
				Nickname: "Test User",
			},
			expected: false,
		},
		{
			name:     "Empty profile",
			profile:  &entity.UserProfile{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.IsComplete()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserProfile_Age(t *testing.T) {
	tests := []struct {
		name     string
		birthday *time.Time
		expected int
	}{
		{
			name:     "30 years old",
			birthday: func() *time.Time { bd := time.Now().AddDate(-30, 0, -1); return &bd }(),
			expected: 30,
		},
		{
			name:     "Birthday today",
			birthday: func() *time.Time { bd := time.Now().AddDate(-25, 0, 0); return &bd }(),
			expected: 25,
		},
		{
			name:     "No birthday",
			birthday: nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &entity.UserProfile{
				Birthday: tt.birthday,
			}
			age := profile.Age()
			assert.Equal(t, tt.expected, age)
		})
	}
}

// Test error handling
func TestUserProfileService_GetProfile_RepositoryErrors(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockUserRepository, *MockUserProfileRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "User repository error",
			setupMocks: func(userRepo *MockUserRepository, profileRepo *MockUserProfileRepository) {
				userRepo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.New("database error"))
			},
			expectError: true,
			errorMsg:    "failed to get user",
		},
		{
			name: "Profile repository error",
			setupMocks: func(userRepo *MockUserRepository, profileRepo *MockUserProfileRepository) {
				userRepo.On("GetByID", mock.Anything, uint(1)).Return(createTestUser(), nil)
				profileRepo.On("GetByUserID", mock.Anything, uint(1)).Return(nil, errors.New("database error"))
			},
			expectError: true,
			errorMsg:    "failed to get profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockProfileRepo := new(MockUserProfileRepository)
			service := NewUserProfileService(mockUserRepo, mockProfileRepo)

			tt.setupMocks(mockUserRepo, mockProfileRepo)

			response, err := service.GetProfile(context.Background(), 1)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}

			mockUserRepo.AssertExpectations(t)
			mockProfileRepo.AssertExpectations(t)
		})
	}
}
