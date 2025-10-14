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

// Test ContactService

func TestContactService_CreateContact(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	req := &types.CreateContactRequest{
		TenantID:   1,
		CustomerID: 1,
		Type:       "phone",
		Value:      "1234567890",
		IsPrimary:  true,
	}

	// Mock: Create succeeds
	contactRepo.On("Create", ctx, mock.AnythingOfType("*entity.Contact")).Return(nil).Run(func(args mock.Arguments) {
		contact := args.Get(1).(*entity.Contact)
		contact.ID = 1
		contact.CreatedAt = time.Now()
		contact.UpdatedAt = time.Now()
	})

	// Execute
	resp, err := service.CreateContact(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, req.TenantID, resp.TenantID)
	assert.Equal(t, req.CustomerID, resp.CustomerID)
	assert.Equal(t, req.Type, resp.Type)
	assert.Equal(t, req.Value, resp.Value)
	assert.True(t, resp.IsPrimary)
	contactRepo.AssertExpectations(t)
}

func TestContactService_GetContact(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	contact := &entity.Contact{
		ID:         1,
		TenantID:   1,
		CustomerID: 1,
		Type:       "email",
		Value:      "test@example.com",
		IsPrimary:  true,
	}

	// Mock
	contactRepo.On("GetByID", ctx, uint(1)).Return(contact, nil)

	// Execute
	resp, err := service.GetContact(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, contact.ID, resp.ID)
	assert.Equal(t, contact.Type, resp.Type)
	assert.Equal(t, contact.Value, resp.Value)
	contactRepo.AssertExpectations(t)
}

func TestContactService_GetContact_NotFound(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	// Mock
	contactRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.ErrContactNotFound)

	// Execute
	resp, err := service.GetContact(ctx, 999)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrContactNotFound, err)
	contactRepo.AssertExpectations(t)
}

func TestContactService_GetContactsByCustomerID(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	now := time.Now()
	contacts := []*entity.Contact{
		{ID: 1, TenantID: 1, CustomerID: 1, Type: "phone", Value: "123", CreatedAt: now, UpdatedAt: now},
		{ID: 2, TenantID: 1, CustomerID: 1, Type: "email", Value: "test@example.com", CreatedAt: now, UpdatedAt: now},
	}

	// Mock
	contactRepo.On("GetByCustomerID", ctx, uint(1)).Return(contacts, nil)

	// Execute
	resp, err := service.GetContactsByCustomerID(ctx, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, "phone", resp[0].Type)
	assert.Equal(t, "email", resp[1].Type)
	contactRepo.AssertExpectations(t)
}

func TestContactService_GetContactsByCustomerID_Empty(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	// Mock: No contacts for this customer
	contactRepo.On("GetByCustomerID", ctx, uint(999)).Return([]*entity.Contact{}, nil)

	// Execute
	resp, err := service.GetContactsByCustomerID(ctx, 999)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 0)
	contactRepo.AssertExpectations(t)
}

func TestContactService_UpdateContact(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	existingContact := &entity.Contact{
		ID:         1,
		TenantID:   1,
		CustomerID: 1,
		Type:       "phone",
		Value:      "111",
		IsPrimary:  false,
	}

	isPrimary := true
	req := &types.UpdateContactRequest{
		Type:      "mobile",
		Value:     "999",
		IsPrimary: &isPrimary,
	}

	// Mock: Get existing contact
	contactRepo.On("GetByID", ctx, uint(1)).Return(existingContact, nil)

	// Mock: Update succeeds
	contactRepo.On("Update", ctx, mock.AnythingOfType("*entity.Contact")).Return(nil)

	// Execute
	resp, err := service.UpdateContact(ctx, 1, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "mobile", resp.Type)
	assert.Equal(t, "999", resp.Value)
	assert.True(t, resp.IsPrimary)
	contactRepo.AssertExpectations(t)
}

func TestContactService_UpdateContact_PartialUpdate(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	existingContact := &entity.Contact{
		ID:         1,
		TenantID:   1,
		CustomerID: 1,
		Type:       "phone",
		Value:      "original-value",
		IsPrimary:  true,
	}

	req := &types.UpdateContactRequest{
		Type: "mobile",
		// Value and IsPrimary not provided, should remain unchanged
	}

	// Mock: Get existing contact
	contactRepo.On("GetByID", ctx, uint(1)).Return(existingContact, nil)

	// Mock: Update succeeds
	contactRepo.On("Update", ctx, mock.AnythingOfType("*entity.Contact")).Return(nil)

	// Execute
	resp, err := service.UpdateContact(ctx, 1, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "mobile", resp.Type)
	assert.Equal(t, "original-value", resp.Value) // Should remain unchanged
	assert.True(t, resp.IsPrimary)                // Should remain unchanged
	contactRepo.AssertExpectations(t)
}

func TestContactService_UpdateContact_SetIsPrimaryToFalse(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	existingContact := &entity.Contact{
		ID:         1,
		TenantID:   1,
		CustomerID: 1,
		Type:       "phone",
		Value:      "123",
		IsPrimary:  true,
	}

	isPrimary := false
	req := &types.UpdateContactRequest{
		IsPrimary: &isPrimary, // Explicitly set to false
	}

	// Mock: Get existing contact
	contactRepo.On("GetByID", ctx, uint(1)).Return(existingContact, nil)

	// Mock: Update succeeds
	contactRepo.On("Update", ctx, mock.AnythingOfType("*entity.Contact")).Return(nil)

	// Execute
	resp, err := service.UpdateContact(ctx, 1, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.IsPrimary) // Should be updated to false
	contactRepo.AssertExpectations(t)
}

func TestContactService_UpdateContact_NotFound(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	req := &types.UpdateContactRequest{
		Type: "mobile",
	}

	// Mock: Contact not found
	contactRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.ErrContactNotFound)

	// Execute
	resp, err := service.UpdateContact(ctx, 999, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrContactNotFound, err)
	contactRepo.AssertExpectations(t)
}

func TestContactService_DeleteContact(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	contact := &entity.Contact{ID: 1, TenantID: 1, CustomerID: 1}

	// Mock: Contact exists
	contactRepo.On("GetByID", ctx, uint(1)).Return(contact, nil)

	// Mock: Delete succeeds
	contactRepo.On("Delete", ctx, uint(1)).Return(nil)

	// Execute
	err := service.DeleteContact(ctx, 1)

	// Assert
	assert.NoError(t, err)
	contactRepo.AssertExpectations(t)
}

func TestContactService_DeleteContact_NotFound(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	// Mock: Contact doesn't exist
	contactRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.ErrContactNotFound)

	// Execute
	err := service.DeleteContact(ctx, 999)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, errors.ErrContactNotFound, err)
	contactRepo.AssertExpectations(t)
}

func TestContactService_ListContacts(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	now := time.Now()
	contacts := []*entity.Contact{
		{ID: 1, TenantID: 1, CustomerID: 1, Type: "phone", Value: "123", CreatedAt: now, UpdatedAt: now},
		{ID: 2, TenantID: 1, CustomerID: 1, Type: "email", Value: "test@example.com", CreatedAt: now, UpdatedAt: now},
		{ID: 3, TenantID: 1, CustomerID: 2, Type: "address", Value: "123 Main St", CreatedAt: now, UpdatedAt: now},
	}

	req := &types.ListRequest{
		Page:     1,
		PageSize: 10,
	}

	// Mock
	contactRepo.On("List", ctx, 0, 10).Return(contacts, nil)

	// Execute
	resp, err := service.ListContacts(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Page)
	assert.Len(t, resp.Data, 3)
	assert.Equal(t, 3, resp.Size)
	contactRepo.AssertExpectations(t)
}

func TestContactService_ListContacts_DefaultPagination(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	req := &types.ListRequest{
		Page:     0, // Invalid, should default to 1
		PageSize: 0, // Should default to DefaultPageSize
	}

	contacts := []*entity.Contact{}

	// Mock: Should use default page size (20)
	contactRepo.On("List", ctx, 0, constants.DefaultPageSize).Return(contacts, nil)

	// Execute
	resp, err := service.ListContacts(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Page) // Should be corrected to 1
	contactRepo.AssertExpectations(t)
}

func TestContactService_ListContacts_MaxPageSize(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	req := &types.ListRequest{
		Page:     1,
		PageSize: 200, // Exceeds max, should be capped to MaxPageSize
	}

	contacts := []*entity.Contact{}

	// Mock: Should use MaxPageSize (100)
	contactRepo.On("List", ctx, 0, constants.MaxPageSize).Return(contacts, nil)

	// Execute
	resp, err := service.ListContacts(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	contactRepo.AssertExpectations(t)
}

func TestContactService_ListContacts_Pagination(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	// Create 25 contacts for testing pagination
	now := time.Now()
	allContacts := make([]*entity.Contact, 25)
	for i := 0; i < 25; i++ {
		allContacts[i] = &entity.Contact{
			ID:         uint(i + 1),
			TenantID:   1,
			CustomerID: 1,
			Type:       "phone",
			Value:      "123",
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	}

	// Test page 2 with page size 10
	req := &types.ListRequest{
		Page:     2,
		PageSize: 10,
	}

	// Mock: Page 2 should have offset 10, limit 10
	contactRepo.On("List", ctx, 10, 10).Return(allContacts[10:20], nil)

	// Execute
	resp, err := service.ListContacts(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, resp.Page)
	assert.Len(t, resp.Data, 10)
	assert.Equal(t, 10, resp.Size)
	contactRepo.AssertExpectations(t)
}

func TestContactService_ListContacts_EmptyResult(t *testing.T) {
	contactRepo := new(MockContactRepository)
	service := NewContactService(contactRepo)
	ctx := context.Background()

	req := &types.ListRequest{
		Page:     1,
		PageSize: 20,
	}

	// Mock: No contacts
	contactRepo.On("List", ctx, 0, 20).Return([]*entity.Contact{}, nil)

	// Execute
	resp, err := service.ListContacts(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Data, 0)
	assert.Equal(t, 0, resp.Size)
	contactRepo.AssertExpectations(t)
}
