package service

import (
	"context"

	"github.com/julesChu12/fly/hermes/internal/domain/entity"
	"github.com/julesChu12/fly/hermes/internal/domain/repository"
	"github.com/julesChu12/fly/hermes/pkg/constants"
	"github.com/julesChu12/fly/hermes/pkg/types"
)

type ContactService interface {
	CreateContact(ctx context.Context, req *types.CreateContactRequest) (*types.ContactResponse, error)
	GetContact(ctx context.Context, id uint) (*types.ContactResponse, error)
	GetContactsByCustomerID(ctx context.Context, customerID uint) ([]types.ContactResponse, error)
	UpdateContact(ctx context.Context, id uint, req *types.UpdateContactRequest) (*types.ContactResponse, error)
	DeleteContact(ctx context.Context, id uint) error
	ListContacts(ctx context.Context, req *types.ListRequest) (*types.ContactListResponse, error)
}

type contactService struct {
	contactRepo repository.ContactRepository
}

func NewContactService(contactRepo repository.ContactRepository) ContactService {
	return &contactService{
		contactRepo: contactRepo,
	}
}

func (s *contactService) CreateContact(ctx context.Context, req *types.CreateContactRequest) (*types.ContactResponse, error) {
	contact := &entity.Contact{
		TenantID:   req.TenantID,
		CustomerID: req.CustomerID,
		Type:       req.Type,
		Value:      req.Value,
		IsPrimary:  req.IsPrimary,
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, err
	}

	return s.toContactResponse(contact), nil
}

func (s *contactService) GetContact(ctx context.Context, id uint) (*types.ContactResponse, error) {
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toContactResponse(contact), nil
}

func (s *contactService) GetContactsByCustomerID(ctx context.Context, customerID uint) ([]types.ContactResponse, error) {
	contacts, err := s.contactRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	responses := make([]types.ContactResponse, len(contacts))
	for i, contact := range contacts {
		responses[i] = *s.toContactResponse(contact)
	}

	return responses, nil
}

func (s *contactService) UpdateContact(ctx context.Context, id uint, req *types.UpdateContactRequest) (*types.ContactResponse, error) {
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Type != "" {
		contact.Type = req.Type
	}
	if req.Value != "" {
		contact.Value = req.Value
	}
	if req.IsPrimary != nil {
		contact.IsPrimary = *req.IsPrimary
	}

	if err := s.contactRepo.Update(ctx, contact); err != nil {
		return nil, err
	}

	return s.toContactResponse(contact), nil
}

func (s *contactService) DeleteContact(ctx context.Context, id uint) error {
	_, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.contactRepo.Delete(ctx, id)
}

func (s *contactService) ListContacts(ctx context.Context, req *types.ListRequest) (*types.ContactListResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = constants.DefaultPageSize
	}
	if req.PageSize > constants.MaxPageSize {
		req.PageSize = constants.MaxPageSize
	}
	if req.Page < 1 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.PageSize
	contacts, err := s.contactRepo.List(ctx, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]types.ContactResponse, len(contacts))
	for i, contact := range contacts {
		responses[i] = *s.toContactResponse(contact)
	}

	return &types.ContactListResponse{
		Data: responses,
		Page: req.Page,
		Size: len(responses),
	}, nil
}

func (s *contactService) toContactResponse(contact *entity.Contact) *types.ContactResponse {
	return &types.ContactResponse{
		ID:         contact.ID,
		TenantID:   contact.TenantID,
		CustomerID: contact.CustomerID,
		Type:       contact.Type,
		Value:      contact.Value,
		IsPrimary:  contact.IsPrimary,
		CreatedAt:  contact.CreatedAt,
		UpdatedAt:  contact.UpdatedAt,
	}
}
