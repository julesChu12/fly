package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// CustomerProxy represents the use case for customer operations
type CustomerProxy struct {
	customerClient *client.CustomerHTTPClient
	logger         *logger.Logger
}

// NewCustomerProxy creates a new CustomerProxy instance
func NewCustomerProxy(customerClient *client.CustomerHTTPClient, logger *logger.Logger) *CustomerProxy {
	return &CustomerProxy{
		customerClient: customerClient,
		logger:         logger,
	}
}

// ListCustomers retrieves a list of customers
func (p *CustomerProxy) ListCustomers(ctx context.Context, filter *client.CustomerFilter) ([]client.Customer, int64, error) {
	p.logger.Info("Retrieving customer list", "filter", filter)

	// Set defaults for filter
	if filter != nil {
		if filter.Page <= 0 {
			filter.Page = 1
		}
		if filter.Limit <= 0 {
			filter.Limit = 20
		}
	}

	start := time.Now()
	customers, total, err := p.customerClient.ListCustomers(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve customer list", "error", err.Error(), "duration", duration)
		return nil, 0, fmt.Errorf("failed to retrieve customer list: %w", err)
	}

	p.logger.Info("Successfully retrieved customer list", "count", len(customers), "total", total, "duration", duration)
	return customers, total, nil
}

// CreateCustomer creates a new customer
func (p *CustomerProxy) CreateCustomer(ctx context.Context, req *client.CreateCustomerRequestHTTP) (*client.Customer, error) {
	p.logger.Info("Creating new customer", "name", req.Name, "email", req.Email)

	// Validate request
	if err := p.validateCreateCustomerRequest(req); err != nil {
		p.logger.Error("Invalid customer creation request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	customer, err := p.customerClient.CreateCustomer(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to create customer", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	p.logger.Info("Successfully created customer", "customer_id", customer.ID, "name", customer.Name, "duration", duration)
	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (p *CustomerProxy) GetCustomer(ctx context.Context, customerID uint) (*client.Customer, error) {
	p.logger.Info("Retrieving customer", "customer_id", customerID)

	if customerID == 0 {
		return nil, fmt.Errorf("customer ID is required")
	}

	start := time.Now()
	customer, err := p.customerClient.GetCustomer(ctx, customerID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve customer", "customer_id", customerID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve customer: %w", err)
	}

	p.logger.Info("Successfully retrieved customer", "customer_id", customer.ID, "name", customer.Name, "duration", duration)
	return customer, nil
}

// UpdateCustomer updates an existing customer
func (p *CustomerProxy) UpdateCustomer(ctx context.Context, customerID uint, req *client.UpdateCustomerRequestHTTP) (*client.Customer, error) {
	p.logger.Info("Updating customer", "customer_id", customerID)

	if customerID == 0 {
		return nil, fmt.Errorf("customer ID is required")
	}

	// Validate that customer exists first
	_, err := p.customerClient.GetCustomer(ctx, customerID)
	if err != nil {
		p.logger.Error("Customer not found for update", "customer_id", customerID, "error", err.Error())
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Validate request
	if err := p.validateUpdateCustomerRequest(req); err != nil {
		p.logger.Error("Invalid customer update request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	updatedCustomer, err := p.customerClient.UpdateCustomer(ctx, customerID, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to update customer", "customer_id", customerID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	p.logger.Info("Successfully updated customer", "customer_id", updatedCustomer.ID, "name", updatedCustomer.Name, "duration", duration)
	return updatedCustomer, nil
}

// DeleteCustomer deletes a customer
func (p *CustomerProxy) DeleteCustomer(ctx context.Context, customerID uint) error {
	p.logger.Info("Deleting customer", "customer_id", customerID)

	if customerID == 0 {
		return fmt.Errorf("customer ID is required")
	}

	// Validate that customer exists first
	_, err := p.customerClient.GetCustomer(ctx, customerID)
	if err != nil {
		p.logger.Error("Customer not found for deletion", "customer_id", customerID, "error", err.Error())
		return fmt.Errorf("customer not found: %w", err)
	}

	start := time.Now()
	err = p.customerClient.DeleteCustomer(ctx, customerID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to delete customer", "customer_id", customerID, "error", err.Error(), "duration", duration)
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	p.logger.Info("Successfully deleted customer", "customer_id", customerID, "duration", duration)
	return nil
}

// CreateContact creates a new contact for a customer
func (p *CustomerProxy) CreateContact(ctx context.Context, req *client.CreateContactRequestHTTP) (*client.Contact, error) {
	p.logger.Info("Creating new contact", "customer_id", req.CustomerID, "type", req.Type)

	// Validate request
	if err := p.validateCreateContactRequest(req); err != nil {
		p.logger.Error("Invalid contact creation request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	contact, err := p.customerClient.CreateContact(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to create contact", "customer_id", req.CustomerID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	p.logger.Info("Successfully created contact", "contact_id", contact.ID, "customer_id", contact.CustomerID, "duration", duration)
	return contact, nil
}

// GetContacts retrieves contacts for a specific customer
func (p *CustomerProxy) GetContacts(ctx context.Context, customerID uint, page, pageSize int) ([]client.Contact, error) {
	p.logger.Info("Retrieving contacts for customer", "customer_id", customerID)

	if customerID == 0 {
		return nil, fmt.Errorf("customer ID is required")
	}

	// Set defaults
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	start := time.Now()
	contacts, err := p.customerClient.GetContacts(ctx, customerID, page, pageSize)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve contacts", "customer_id", customerID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve contacts: %w", err)
	}

	p.logger.Info("Successfully retrieved contacts", "customer_id", customerID, "count", len(contacts), "duration", duration)
	return contacts, nil
}

// UpdateContact updates an existing contact
func (p *CustomerProxy) UpdateContact(ctx context.Context, contactID uint, req *client.UpdateContactRequestHTTP) (*client.Contact, error) {
	p.logger.Info("Updating contact", "contact_id", contactID)

	if contactID == 0 {
		return nil, fmt.Errorf("contact ID is required")
	}

	// Validate request
	if err := p.validateUpdateContactRequest(req); err != nil {
		p.logger.Error("Invalid contact update request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	updatedContact, err := p.customerClient.UpdateContact(ctx, contactID, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to update contact", "contact_id", contactID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	p.logger.Info("Successfully updated contact", "contact_id", updatedContact.ID, "customer_id", updatedContact.CustomerID, "duration", duration)
	return updatedContact, nil
}

// DeleteContact deletes a contact
func (p *CustomerProxy) DeleteContact(ctx context.Context, contactID uint) error {
	p.logger.Info("Deleting contact", "contact_id", contactID)

	if contactID == 0 {
		return fmt.Errorf("contact ID is required")
	}

	start := time.Now()
	err := p.customerClient.DeleteContact(ctx, contactID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to delete contact", "contact_id", contactID, "error", err.Error(), "duration", duration)
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	p.logger.Info("Successfully deleted contact", "contact_id", contactID, "duration", duration)
	return nil
}

// SearchCustomers searches for customers by name, phone, or email
func (p *CustomerProxy) SearchCustomers(ctx context.Context, searchTerm string, page, limit int) ([]client.Customer, int64, error) {
	p.logger.Info("Searching customers", "search_term", searchTerm, "page", page, "limit", limit)

	if searchTerm == "" {
		return nil, 0, fmt.Errorf("search term is required")
	}

	// Set defaults
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	start := time.Now()
	customers, total, err := p.customerClient.SearchCustomers(ctx, searchTerm, page, limit)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to search customers", "search_term", searchTerm, "error", err.Error(), "duration", duration)
		return nil, 0, fmt.Errorf("failed to search customers: %w", err)
	}

	p.logger.Info("Successfully searched customers", "search_term", searchTerm, "count", len(customers), "total", total, "duration", duration)
	return customers, total, nil
}

// Validation helper functions

func (p *CustomerProxy) validateCreateCustomerRequest(req *client.CreateCustomerRequestHTTP) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}

func (p *CustomerProxy) validateUpdateCustomerRequest(req *client.UpdateCustomerRequestHTTP) error {
	// For update, allow partial updates but ensure at least one field is provided
	if req.Name == "" && req.Phone == "" && req.Email == "" && req.Tags == "" {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

func (p *CustomerProxy) validateCreateContactRequest(req *client.CreateContactRequestHTTP) error {
	if req.CustomerID == 0 {
		return fmt.Errorf("customer ID is required")
	}
	if req.Type == "" {
		return fmt.Errorf("contact type is required")
	}
	if req.Value == "" {
		return fmt.Errorf("contact value is required")
	}
	return nil
}

func (p *CustomerProxy) validateUpdateContactRequest(req *client.UpdateContactRequestHTTP) error {
	// For update, allow partial updates but ensure at least one field is provided
	if req.Type == "" && req.Value == "" {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}