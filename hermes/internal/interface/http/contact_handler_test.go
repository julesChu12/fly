package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/hermes/pkg/errors"
	"github.com/julesChu12/fly/hermes/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockContactService is a mock implementation of ContactService
type MockContactService struct {
	mock.Mock
}

func (m *MockContactService) CreateContact(ctx context.Context, req *types.CreateContactRequest) (*types.ContactResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactResponse), args.Error(1)
}

func (m *MockContactService) GetContact(ctx context.Context, id uint) (*types.ContactResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactResponse), args.Error(1)
}

func (m *MockContactService) GetContactsByCustomerID(ctx context.Context, customerID uint) ([]types.ContactResponse, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.ContactResponse), args.Error(1)
}

func (m *MockContactService) UpdateContact(ctx context.Context, id uint, req *types.UpdateContactRequest) (*types.ContactResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactResponse), args.Error(1)
}

func (m *MockContactService) DeleteContact(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockContactService) ListContacts(ctx context.Context, req *types.ListRequest) (*types.ContactListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ContactListResponse), args.Error(1)
}

func TestContactHandler_CreateContact(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	req := &types.CreateContactRequest{
		TenantID:   1,
		CustomerID: 1,
		Type:       "phone",
		Value:      "1234567890",
		IsPrimary:  true,
	}

	resp := &types.ContactResponse{
		ID:         1,
		TenantID:   req.TenantID,
		CustomerID: req.CustomerID,
		Type:       req.Type,
		Value:      req.Value,
		IsPrimary:  req.IsPrimary,
	}

	mockService.On("CreateContact", mock.Anything, req).Return(resp, nil)

	// Create request
	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/contacts", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateContact(c)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestContactHandler_CreateContact_InvalidJSON(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	// Invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/contacts", bytes.NewBufferString("{invalid"))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateContact(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContactHandler_GetContact(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	resp := &types.ContactResponse{
		ID:         1,
		TenantID:   1,
		CustomerID: 1,
		Type:       "email",
		Value:      "test@example.com",
		IsPrimary:  true,
	}

	mockService.On("GetContact", mock.Anything, uint(1)).Return(resp, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/contacts/1", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Execute
	handler.GetContact(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestContactHandler_GetContact_InvalidID(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	// Create request with invalid ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/contacts/invalid", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	// Execute
	handler.GetContact(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContactHandler_GetContact_NotFound(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	mockService.On("GetContact", mock.Anything, uint(999)).Return(nil, errors.ErrContactNotFound)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/contacts/999", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

	// Execute
	handler.GetContact(c)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestContactHandler_UpdateContact(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	isPrimary := true
	req := &types.UpdateContactRequest{
		Type:      "mobile",
		Value:     "9876543210",
		IsPrimary: &isPrimary,
	}

	resp := &types.ContactResponse{
		ID:         1,
		TenantID:   1,
		CustomerID: 1,
		Type:       req.Type,
		Value:      req.Value,
		IsPrimary:  *req.IsPrimary,
	}

	mockService.On("UpdateContact", mock.Anything, uint(1), req).Return(resp, nil)

	// Create request
	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/contacts/1", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Execute
	handler.UpdateContact(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestContactHandler_UpdateContact_InvalidID(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	req := &types.UpdateContactRequest{Type: "phone"}
	reqBody, _ := json.Marshal(req)

	// Create request with invalid ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/contacts/invalid", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	// Execute
	handler.UpdateContact(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContactHandler_UpdateContact_InvalidJSON(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	// Invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/contacts/1", bytes.NewBufferString("{invalid"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Execute
	handler.UpdateContact(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContactHandler_DeleteContact(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	mockService.On("DeleteContact", mock.Anything, uint(1)).Return(nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/contacts/1", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// Execute
	handler.DeleteContact(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestContactHandler_DeleteContact_InvalidID(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	// Create request with invalid ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/contacts/invalid", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	// Execute
	handler.DeleteContact(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContactHandler_DeleteContact_NotFound(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	mockService.On("DeleteContact", mock.Anything, uint(999)).Return(errors.ErrContactNotFound)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/contacts/999", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

	// Execute
	handler.DeleteContact(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestContactHandler_ListContacts(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	resp := &types.ContactListResponse{
		Data: []types.ContactResponse{
			{ID: 1, Type: "phone", Value: "123"},
			{ID: 2, Type: "email", Value: "test@example.com"},
		},
		Page: 1,
		Size: 2,
	}

	mockService.On("ListContacts", mock.Anything, mock.AnythingOfType("*types.ListRequest")).Return(resp, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/contacts?page=1&page_size=20", nil)

	// Execute
	handler.ListContacts(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestContactHandler_ListContacts_InvalidQuery(t *testing.T) {
	mockService := new(MockContactService)
	handler := NewContactHandler(mockService)

	// Create request with invalid query params
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/contacts?page=invalid", nil)

	// Execute
	handler.ListContacts(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
