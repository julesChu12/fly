package service

import (
	"context"
	"testing"
	"time"

	"github.com/julesChu12/fly/hermes/internal/domain/entity"
	"github.com/julesChu12/fly/hermes/pkg/constants"
	"github.com/julesChu12/fly/hermes/pkg/errors"
	"github.com/julesChu12/fly/hermes/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCustomerRepository is a mock implementation of CustomerRepository
type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) GetByID(ctx context.Context, id uint) (*entity.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByIDAndTenant(ctx context.Context, id uint, tenantID uint) (*entity.Customer, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByEmail(ctx context.Context, email string) (*entity.Customer, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetByEmailAndTenant(ctx context.Context, email string, tenantID uint) (*entity.Customer, error) {
	args := m.Called(ctx, email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Update(ctx context.Context, customer *entity.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomerRepository) DeleteByTenant(ctx context.Context, id uint, tenantID uint) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockCustomerRepository) List(ctx context.Context, offset, limit int) ([]*entity.Customer, int64, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Customer), args.Get(1).(int64), args.Error(2)
}

func (m *MockCustomerRepository) ListByTenant(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Customer, int64, error) {
	args := m.Called(ctx, tenantID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Customer), args.Get(1).(int64), args.Error(2)
}

func (m *MockCustomerRepository) GetWithContacts(ctx context.Context, id uint) (*entity.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetWithContactsByTenant(ctx context.Context, id uint, tenantID uint) (*entity.Customer, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

// MockContactRepository is a mock implementation of ContactRepository
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) Create(ctx context.Context, contact *entity.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactRepository) GetByID(ctx context.Context, id uint) (*entity.Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Contact), args.Error(1)
}

func (m *MockContactRepository) GetByIDAndTenant(ctx context.Context, id uint, tenantID uint) (*entity.Contact, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Contact), args.Error(1)
}

func (m *MockContactRepository) GetByCustomerID(ctx context.Context, customerID uint) ([]*entity.Contact, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Contact), args.Error(1)
}

func (m *MockContactRepository) GetByCustomerIDAndTenant(ctx context.Context, customerID uint, tenantID uint) ([]*entity.Contact, error) {
	args := m.Called(ctx, customerID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Contact), args.Error(1)
}

func (m *MockContactRepository) Update(ctx context.Context, contact *entity.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactRepository) DeleteByTenant(ctx context.Context, id uint, tenantID uint) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockContactRepository) List(ctx context.Context, offset, limit int) ([]*entity.Contact, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Contact), args.Error(1)
}

func (m *MockContactRepository) ListByTenant(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Contact, error) {
	args := m.Called(ctx, tenantID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Contact), args.Error(1)
}

// Test CustomerService

func TestCustomerService_CreateCustomer(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	req := &types.CreateCustomerRequest{
		Name:  "Test Customer",
		Phone: "1234567890",
		Email: "test@example.com",
		Tags:  "vip",
	}

	// Mock: Email doesn't exist
	customerRepo.On("GetByEmail", ctx, req.Email).Return(nil, errors.ErrCustomerNotFound)

	// Mock: Create succeeds
	customerRepo.On("Create", ctx, mock.AnythingOfType("*entity.Customer")).Return(nil).Run(func(args mock.Arguments) {
		customer := args.Get(1).(*entity.Customer)
		customer.ID = 1
		customer.CreatedAt = time.Now()
		customer.UpdatedAt = time.Now()
	})

	// Execute
	resp, err := service.CreateCustomer(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, req.Name, resp.Name)
	assert.Equal(t, req.Email, resp.Email)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_CreateCustomer_DuplicateEmail(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	req := &types.CreateCustomerRequest{
		Name:  "Test Customer",
		Email: "duplicate@example.com",
	}

	existingCustomer := &entity.Customer{
		ID:    1,
		Email: req.Email,
	}

	// Mock: Email already exists
	customerRepo.On("GetByEmail", ctx, req.Email).Return(existingCustomer, nil)

	// Execute
	resp, err := service.CreateCustomer(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrDuplicateEmail, err)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_GetCustomer(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	customer := &entity.Customer{
		ID:    1,
		Name:  "Test Customer",
		Email: "test@example.com",
	}

	// Mock
	customerRepo.On("GetByID", ctx, uint(1)).Return(customer, nil)

	// Execute
	resp, err := service.GetCustomer(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, customer.ID, resp.ID)
	assert.Equal(t, customer.Name, resp.Name)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_GetCustomer_NotFound(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	// Mock
	customerRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.ErrCustomerNotFound)

	// Execute
	resp, err := service.GetCustomer(ctx, 999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrCustomerNotFound, err)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_GetCustomerWithContacts(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	now := time.Now()
	customer := &entity.Customer{
		ID:    1,
		Name:  "Test Customer",
		Email: "test@example.com",
		Contacts: []entity.Contact{
			{ID: 1, CustomerID: 1, Type: "phone", Value: "123", CreatedAt: now, UpdatedAt: now},
			{ID: 2, CustomerID: 1, Type: "email", Value: "contact@example.com", CreatedAt: now, UpdatedAt: now},
		},
	}

	// Mock
	customerRepo.On("GetWithContacts", ctx, uint(1)).Return(customer, nil)

	// Execute
	resp, err := service.GetCustomerWithContacts(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, customer.ID, resp.ID)
	assert.Len(t, resp.Contacts, 2)
	assert.Equal(t, "phone", resp.Contacts[0].Type)
	assert.Equal(t, "email", resp.Contacts[1].Type)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_UpdateCustomer(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	existingCustomer := &entity.Customer{
		ID:    1,
		Name:  "Old Name",
		Email: "old@example.com",
		Phone: "111",
	}

	req := &types.UpdateCustomerRequest{
		Name:  "New Name",
		Phone: "999",
		Email: "new@example.com",
	}

	// Mock: Get existing customer
	customerRepo.On("GetByID", ctx, uint(1)).Return(existingCustomer, nil)

	// Mock: New email doesn't exist
	customerRepo.On("GetByEmail", ctx, req.Email).Return(nil, errors.ErrCustomerNotFound)

	// Mock: Update succeeds
	customerRepo.On("Update", ctx, mock.AnythingOfType("*entity.Customer")).Return(nil)

	// Execute
	resp, err := service.UpdateCustomer(ctx, 1, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.Name, resp.Name)
	assert.Equal(t, req.Phone, resp.Phone)
	assert.Equal(t, req.Email, resp.Email)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_UpdateCustomer_DuplicateEmail(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	existingCustomer := &entity.Customer{
		ID:    1,
		Email: "old@example.com",
	}

	anotherCustomer := &entity.Customer{
		ID:    2,
		Email: "taken@example.com",
	}

	req := &types.UpdateCustomerRequest{
		Email: "taken@example.com",
	}

	// Mock: Get existing customer
	customerRepo.On("GetByID", ctx, uint(1)).Return(existingCustomer, nil)

	// Mock: Email already taken by another customer
	customerRepo.On("GetByEmail", ctx, req.Email).Return(anotherCustomer, nil)

	// Execute
	resp, err := service.UpdateCustomer(ctx, 1, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrDuplicateEmail, err)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_UpdateCustomer_PartialUpdate(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	existingCustomer := &entity.Customer{
		ID:    1,
		Name:  "Original Name",
		Email: "original@example.com",
		Phone: "111",
		Tags:  "old-tag",
	}

	req := &types.UpdateCustomerRequest{
		Name: "Updated Name",
		// Email and Phone not provided, should remain unchanged
	}

	// Mock: Get existing customer
	customerRepo.On("GetByID", ctx, uint(1)).Return(existingCustomer, nil)

	// Mock: Update succeeds
	customerRepo.On("Update", ctx, mock.AnythingOfType("*entity.Customer")).Return(nil)

	// Execute
	resp, err := service.UpdateCustomer(ctx, 1, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Updated Name", resp.Name)
	assert.Equal(t, "original@example.com", resp.Email) // Should remain unchanged
	assert.Equal(t, "111", resp.Phone)                  // Should remain unchanged
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_DeleteCustomer(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	customer := &entity.Customer{ID: 1, Name: "Test"}

	// Mock: Customer exists
	customerRepo.On("GetByID", ctx, uint(1)).Return(customer, nil)

	// Mock: Delete succeeds
	customerRepo.On("Delete", ctx, uint(1)).Return(nil)

	// Execute
	err := service.DeleteCustomer(ctx, 1)

	// Assert
	assert.NoError(t, err)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_DeleteCustomer_NotFound(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	// Mock: Customer doesn't exist
	customerRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.ErrCustomerNotFound)

	// Execute
	err := service.DeleteCustomer(ctx, 999)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrCustomerNotFound, err)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_ListCustomers(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	customers := []*entity.Customer{
		{ID: 1, Name: "Customer 1"},
		{ID: 2, Name: "Customer 2"},
		{ID: 3, Name: "Customer 3"},
	}

	req := &types.ListRequest{
		Page:     1,
		PageSize: 10,
	}

	// Mock
	customerRepo.On("List", ctx, 0, 10).Return(customers, int64(3), nil)

	// Execute
	resp, err := service.ListCustomers(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(3), resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Len(t, resp.Data.([]types.CustomerResponse), 3)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_ListCustomers_DefaultPagination(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	req := &types.ListRequest{
		Page:     0, // Invalid, should default to 1
		PageSize: 0, // Should default to DefaultPageSize
	}

	customers := []*entity.Customer{}

	// Mock: Should use default page size (20)
	customerRepo.On("List", ctx, 0, constants.DefaultPageSize).Return(customers, int64(0), nil)

	// Execute
	resp, err := service.ListCustomers(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Page) // Should be corrected to 1
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_ListCustomers_MaxPageSize(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	req := &types.ListRequest{
		Page:     1,
		PageSize: 200, // Exceeds max, should be capped to MaxPageSize
	}

	customers := []*entity.Customer{}

	// Mock: Should use MaxPageSize (100)
	customerRepo.On("List", ctx, 0, constants.MaxPageSize).Return(customers, int64(0), nil)

	// Execute
	resp, err := service.ListCustomers(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	customerRepo.AssertExpectations(t)
}

func TestCustomerService_ListCustomers_Pagination(t *testing.T) {
	customerRepo := new(MockCustomerRepository)
	contactRepo := new(MockContactRepository)
	service := NewCustomerService(customerRepo, contactRepo)
	ctx := context.Background()

	// Create 25 customers for testing pagination
	allCustomers := make([]*entity.Customer, 25)
	for i := 0; i < 25; i++ {
		allCustomers[i] = &entity.Customer{
			ID:   uint(i + 1),
			Name: "Customer",
		}
	}

	// Test page 2 with page size 10
	req := &types.ListRequest{
		Page:     2,
		PageSize: 10,
	}

	// Mock: Page 2 should have offset 10, limit 10
	customerRepo.On("List", ctx, 10, 10).Return(allCustomers[10:20], int64(25), nil)

	// Execute
	resp, err := service.ListCustomers(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(25), resp.Total)
	assert.Equal(t, 2, resp.Page)
	assert.Len(t, resp.Data.([]types.CustomerResponse), 10)
	customerRepo.AssertExpectations(t)
}
