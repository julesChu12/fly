package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/hermes/pkg/errors"
	"github.com/julesChu12/fly/hermes/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCustomerService is a mock implementation of CustomerService
type MockCustomerService struct {
	mock.Mock
}

func (m *MockCustomerService) CreateCustomer(ctx context.Context, req *types.CreateCustomerRequest) (*types.CustomerResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.CustomerResponse), args.Error(1)
}

func (m *MockCustomerService) GetCustomer(ctx context.Context, id uint) (*types.CustomerResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.CustomerResponse), args.Error(1)
}

func (m *MockCustomerService) GetCustomerWithContacts(ctx context.Context, id uint) (*types.CustomerResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.CustomerResponse), args.Error(1)
}

func (m *MockCustomerService) UpdateCustomer(ctx context.Context, id uint, req *types.UpdateCustomerRequest) (*types.CustomerResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.CustomerResponse), args.Error(1)
}

func (m *MockCustomerService) DeleteCustomer(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomerService) ListCustomers(ctx context.Context, req *types.ListRequest) (*types.ListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ListResponse), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestCustomerHandler_CreateCustomer(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	req := &types.CreateCustomerRequest{
		Name:  "Test Customer",
		Email: "test@example.com",
		Phone: "1234567890",
	}

	resp := &types.CustomerResponse{
		ID:        1,
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockService.On("CreateCustomer", mock.Anything, req).Return(resp, nil)

	// Create request
	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/customers", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateCustomer(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	mockService.AssertExpectations(t)
}

func TestCustomerHandler_CreateCustomer_InvalidJSON(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	// Invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/customers", bytes.NewBufferString("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateCustomer(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCustomerHandler_CreateCustomer_ServiceError(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	req := &types.CreateCustomerRequest{
		Name:  "Test Customer",
		Email: "duplicate@example.com",
	}

	mockService.On("CreateCustomer", mock.Anything, req).Return(nil, errors.ErrDuplicateEmail)

	// Create request
	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/customers", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateCustomer(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestCustomerHandler_GetCustomer(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	resp := &types.CustomerResponse{
		ID:    1,
		Name:  "Test Customer",
		Email: "test@example.com",
		Phone: "1234567890",
	}

	mockService.On("GetCustomer", mock.Anything, uint(1)).Return(resp, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/customers/1", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Execute
	handler.GetCustomer(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	mockService.AssertExpectations(t)
}

func TestCustomerHandler_GetCustomer_InvalidID(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	// Create request with invalid ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/customers/invalid", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	// Execute
	handler.GetCustomer(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCustomerHandler_GetCustomer_NotFound(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	mockService.On("GetCustomer", mock.Anything, uint(999)).Return(nil, errors.ErrCustomerNotFound)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/customers/999", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

	// Execute
	handler.GetCustomer(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestCustomerHandler_GetCustomerWithContacts(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	resp := &types.CustomerResponse{
		ID:    1,
		Name:  "Test Customer",
		Email: "test@example.com",
		Contacts: []types.ContactResponse{
			{ID: 1, Type: "phone", Value: "123"},
			{ID: 2, Type: "email", Value: "contact@example.com"},
		},
	}

	mockService.On("GetCustomerWithContacts", mock.Anything, uint(1)).Return(resp, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/customers/1/contacts", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Execute
	handler.GetCustomerWithContacts(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCustomerHandler_UpdateCustomer(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	req := &types.UpdateCustomerRequest{
		Name:  "Updated Name",
		Email: "updated@example.com",
	}

	resp := &types.CustomerResponse{
		ID:    1,
		Name:  req.Name,
		Email: req.Email,
	}

	mockService.On("UpdateCustomer", mock.Anything, uint(1), req).Return(resp, nil)

	// Create request
	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/customers/1", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Execute
	handler.UpdateCustomer(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCustomerHandler_UpdateCustomer_InvalidID(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	req := &types.UpdateCustomerRequest{Name: "Test"}
	reqBody, _ := json.Marshal(req)

	// Create request with invalid ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/customers/invalid", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	// Execute
	handler.UpdateCustomer(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCustomerHandler_UpdateCustomer_InvalidJSON(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	// Invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/customers/1", bytes.NewBufferString("{invalid"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Execute
	handler.UpdateCustomer(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCustomerHandler_DeleteCustomer(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	mockService.On("DeleteCustomer", mock.Anything, uint(1)).Return(nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/customers/1", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Execute
	handler.DeleteCustomer(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCustomerHandler_DeleteCustomer_InvalidID(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	// Create request with invalid ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/customers/invalid", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	// Execute
	handler.DeleteCustomer(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCustomerHandler_DeleteCustomer_NotFound(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	mockService.On("DeleteCustomer", mock.Anything, uint(999)).Return(errors.ErrCustomerNotFound)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/customers/999", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

	// Execute
	handler.DeleteCustomer(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestCustomerHandler_ListCustomers(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	resp := &types.ListResponse{
		Data: []types.CustomerResponse{
			{ID: 1, Name: "Customer 1"},
			{ID: 2, Name: "Customer 2"},
		},
		Total: 2,
		Page:  1,
		Size:  2,
	}

	mockService.On("ListCustomers", mock.Anything, mock.AnythingOfType("*types.ListRequest")).Return(resp, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/customers?page=1&page_size=20", nil)

	// Execute
	handler.ListCustomers(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCustomerHandler_ListCustomers_InvalidQuery(t *testing.T) {
	mockService := new(MockCustomerService)
	handler := NewCustomerHandler(mockService)

	// Create request with invalid query params
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/customers?page=invalid", nil)

	// Execute
	handler.ListCustomers(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
