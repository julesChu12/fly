package service

import (
	"context"
	"hermes/internal/domain/entity"
	"hermes/internal/domain/repository"
	"hermes/pkg/constants"
	"hermes/pkg/errors"
	"hermes/pkg/types"
)

type CustomerService interface {
	CreateCustomer(ctx context.Context, req *types.CreateCustomerRequest) (*types.CustomerResponse, error)
	GetCustomer(ctx context.Context, id uint) (*types.CustomerResponse, error)
	GetCustomerWithContacts(ctx context.Context, id uint) (*types.CustomerResponse, error)
	UpdateCustomer(ctx context.Context, id uint, req *types.UpdateCustomerRequest) (*types.CustomerResponse, error)
	DeleteCustomer(ctx context.Context, id uint) error
	ListCustomers(ctx context.Context, req *types.ListRequest) (*types.ListResponse, error)
}

type customerService struct {
	customerRepo repository.CustomerRepository
	contactRepo  repository.ContactRepository
}

func NewCustomerService(customerRepo repository.CustomerRepository, contactRepo repository.ContactRepository) CustomerService {
	return &customerService{
		customerRepo: customerRepo,
		contactRepo:  contactRepo,
	}
}

func (s *customerService) CreateCustomer(ctx context.Context, req *types.CreateCustomerRequest) (*types.CustomerResponse, error) {
	if req.Email != "" {
		_, err := s.customerRepo.GetByEmail(ctx, req.Email)
		if err == nil {
			return nil, errors.ErrDuplicateEmail
		}
		if err != errors.ErrCustomerNotFound {
			return nil, err
		}
	}

	customer := &entity.Customer{
		Name:  req.Name,
		Phone: req.Phone,
		Email: req.Email,
		Tags:  req.Tags,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, err
	}

	return s.toCustomerResponse(customer), nil
}

func (s *customerService) GetCustomer(ctx context.Context, id uint) (*types.CustomerResponse, error) {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toCustomerResponse(customer), nil
}

func (s *customerService) GetCustomerWithContacts(ctx context.Context, id uint) (*types.CustomerResponse, error) {
	customer, err := s.customerRepo.GetWithContacts(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := s.toCustomerResponse(customer)
	resp.Contacts = make([]types.ContactResponse, len(customer.Contacts))
	for i, contact := range customer.Contacts {
		resp.Contacts[i] = s.toContactResponse(&contact)
	}

	return resp, nil
}

func (s *customerService) UpdateCustomer(ctx context.Context, id uint, req *types.UpdateCustomerRequest) (*types.CustomerResponse, error) {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Email != "" && req.Email != customer.Email {
		_, err := s.customerRepo.GetByEmail(ctx, req.Email)
		if err == nil {
			return nil, errors.ErrDuplicateEmail
		}
		if err != errors.ErrCustomerNotFound {
			return nil, err
		}
	}

	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Email != "" {
		customer.Email = req.Email
	}
	if req.Tags != "" {
		customer.Tags = req.Tags
	}

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return nil, err
	}

	return s.toCustomerResponse(customer), nil
}

func (s *customerService) DeleteCustomer(ctx context.Context, id uint) error {
	_, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.customerRepo.Delete(ctx, id)
}

func (s *customerService) ListCustomers(ctx context.Context, req *types.ListRequest) (*types.ListResponse, error) {
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
	customers, err := s.customerRepo.List(ctx, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]types.CustomerResponse, len(customers))
	for i, customer := range customers {
		responses[i] = *s.toCustomerResponse(customer)
	}

	return &types.ListResponse{
		Data: responses,
		Page: req.Page,
		Size: len(responses),
	}, nil
}

func (s *customerService) toCustomerResponse(customer *entity.Customer) *types.CustomerResponse {
	return &types.CustomerResponse{
		ID:        customer.ID,
		Name:      customer.Name,
		Phone:     customer.Phone,
		Email:     customer.Email,
		Tags:      customer.Tags,
		CreatedAt: customer.CreatedAt,
		UpdatedAt: customer.UpdatedAt,
	}
}

func (s *customerService) toContactResponse(contact *entity.Contact) types.ContactResponse {
	return types.ContactResponse{
		ID:         contact.ID,
		CustomerID: contact.CustomerID,
		Type:       contact.Type,
		Value:      contact.Value,
		IsPrimary:  contact.IsPrimary,
		CreatedAt:  contact.CreatedAt,
		UpdatedAt:  contact.UpdatedAt,
	}
}