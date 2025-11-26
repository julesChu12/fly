package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	hermesv1 "github.com/julesChu12/fly/hermes/api/proto"
	"github.com/julesChu12/fly/mora/pkg/discovery"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// CustomerGRPCClient represents a gRPC client for the Hermes (Customer) service
type CustomerGRPCClient struct {
	conn           *grpc.ClientConn
	customerClient hermesv1.CustomerServiceClient
	contactClient  hermesv1.ContactServiceClient
	logger         *logger.Logger
}

// CustomerGRPCClientConfig represents configuration for the gRPC client
type CustomerGRPCClientConfig struct {
	MaxRecvMsgSize      int           `yaml:"max_recv_msg_size"`
	MaxSendMsgSize      int           `yaml:"max_send_msg_size"`
	ConnectTimeout      time.Duration `yaml:"connect_timeout"`
	KeepaliveTime       time.Duration `yaml:"keepalive_time"`
	KeepaliveTimeout    time.Duration `yaml:"keepalive_timeout"`
	EnableRetry         bool          `yaml:"enable_retry"`
	MaxRetries          int           `yaml:"max_retries"`
	PermitWithoutStream bool          `yaml:"permit_without_stream"`
}

// DefaultCustomerGRPCClientConfig returns default configuration for the gRPC client
func DefaultCustomerGRPCClientConfig() *CustomerGRPCClientConfig {
	return &CustomerGRPCClientConfig{
		MaxRecvMsgSize:      4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:      4 * 1024 * 1024, // 4MB
		ConnectTimeout:      10 * time.Second,
		KeepaliveTime:       30 * time.Second,
		KeepaliveTimeout:    5 * time.Second,
		EnableRetry:         true,
		MaxRetries:          3,
		PermitWithoutStream: true,
	}
}

// NewCustomerGRPCClient creates a new Customer gRPC client with optimized settings
func NewCustomerGRPCClient(target string, config *CustomerGRPCClientConfig) (*CustomerGRPCClient, error) {
	if config == nil {
		config = DefaultCustomerGRPCClientConfig()
	}

	log := logger.NewDefault()
	log.Info("Creating optimized Customer gRPC client", "target", target)

	// Configure gRPC dial options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(config.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(config.MaxSendMsgSize),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                config.KeepaliveTime,
			Timeout:             config.KeepaliveTimeout,
			PermitWithoutStream: config.PermitWithoutStream,
		}),
	}

	// Connect to gRPC server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		log.Error("Failed to connect to Customer gRPC service", "error", err.Error())
		return nil, fmt.Errorf("failed to connect to Customer gRPC service: %w", err)
	}

	customerClient := hermesv1.NewCustomerServiceClient(conn)
	contactClient := hermesv1.NewContactServiceClient(conn)

	log.Info("Successfully connected to Customer gRPC service", "target", target)

	return &CustomerGRPCClient{
		conn:           conn,
		customerClient: customerClient,
		contactClient:  contactClient,
		logger:         log,
	}, nil
}

// NewCustomerGRPCClientWithDiscovery uses service discovery to create a Customer gRPC client
func NewCustomerGRPCClientWithDiscovery(disc discovery.Discovery, config *CustomerGRPCClientConfig) (*CustomerGRPCClient, error) {
	if config == nil {
		config = DefaultCustomerGRPCClientConfig()
	}

	log := logger.NewDefault()

	// Get Hermes service address from discovery
	instance, err := disc.GetService(context.Background(), "hermes")
	if err != nil {
		log.Error("Failed to discover Hermes service", "error", err.Error())
		return nil, fmt.Errorf("failed to discover Hermes service: %w", err)
	}

	address := instance.Address()
	target := fmt.Sprintf("%s:%d", address, 9080) // Assuming gRPC port 9080
	log.Info("Discovered Hermes gRPC service", "address", target, "instance_id", instance.ID)

	return NewCustomerGRPCClient(target, config)
}

// === Customer Service Methods ===

// CreateCustomer creates a new customer via gRPC
func (c *CustomerGRPCClient) CreateCustomer(ctx context.Context, name, phone, email, tags string) (*hermesv1.Customer, error) {
	c.logger.Debug("Creating customer via gRPC", "name", name, "email", email)

	req := &hermesv1.CreateCustomerRequest{
		Name:  name,
		Phone: phone,
		Email: email,
		Tags:  tags,
	}

	resp, err := c.customerClient.CreateCustomer(ctx, req)
	if err != nil {
		c.logger.Error("Failed to create customer via gRPC", "error", err.Error())
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	c.logger.Info("Customer created successfully via gRPC", "customer_id", resp.Customer.Id, "name", resp.Customer.Name)
	return resp.Customer, nil
}

// GetCustomer retrieves a customer by ID via gRPC
func (c *CustomerGRPCClient) GetCustomer(ctx context.Context, customerID uint32) (*hermesv1.Customer, error) {
	c.logger.Debug("Getting customer via gRPC", "customer_id", customerID)

	req := &hermesv1.GetCustomerRequest{
		Id: customerID,
	}

	resp, err := c.customerClient.GetCustomer(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get customer via gRPC", "customer_id", customerID, "error", err.Error())
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	c.logger.Debug("Customer retrieved successfully via gRPC", "customer_id", customerID)
	return resp.Customer, nil
}

// GetCustomerWithContacts retrieves a customer with contacts via gRPC
func (c *CustomerGRPCClient) GetCustomerWithContacts(ctx context.Context, customerID uint32) (*hermesv1.Customer, error) {
	c.logger.Debug("Getting customer with contacts via gRPC", "customer_id", customerID)

	req := &hermesv1.GetCustomerRequest{
		Id: customerID,
	}

	resp, err := c.customerClient.GetCustomerWithContacts(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get customer with contacts via gRPC", "customer_id", customerID, "error", err.Error())
		return nil, fmt.Errorf("failed to get customer with contacts: %w", err)
	}

	c.logger.Debug("Customer with contacts retrieved successfully via gRPC", "customer_id", customerID, "contacts_count", len(resp.Customer.Contacts))
	return resp.Customer, nil
}

// ListCustomers retrieves a paginated list of customers via gRPC
func (c *CustomerGRPCClient) ListCustomers(ctx context.Context, page, pageSize int32) (*hermesv1.ListCustomersResponse, error) {
	c.logger.Debug("Listing customers via gRPC", "page", page, "page_size", pageSize)

	req := &hermesv1.ListCustomersRequest{
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := c.customerClient.ListCustomers(ctx, req)
	if err != nil {
		c.logger.Error("Failed to list customers via gRPC", "error", err.Error())
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	c.logger.Debug("Customers listed successfully via gRPC", "total", resp.Total, "page", resp.Page)
	return resp, nil
}

// UpdateCustomer updates a customer via gRPC
func (c *CustomerGRPCClient) UpdateCustomer(ctx context.Context, customerID uint32, name, phone, email, tags string) (*hermesv1.Customer, error) {
	c.logger.Debug("Updating customer via gRPC", "customer_id", customerID, "name", name)

	req := &hermesv1.UpdateCustomerRequest{
		Id:    customerID,
		Name:  name,
		Phone: phone,
		Email: email,
		Tags:  tags,
	}

	resp, err := c.customerClient.UpdateCustomer(ctx, req)
	if err != nil {
		c.logger.Error("Failed to update customer via gRPC", "customer_id", customerID, "error", err.Error())
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	c.logger.Info("Customer updated successfully via gRPC", "customer_id", customerID, "new_name", resp.Customer.Name)
	return resp.Customer, nil
}

// DeleteCustomer deletes a customer via gRPC
func (c *CustomerGRPCClient) DeleteCustomer(ctx context.Context, customerID uint32) error {
	c.logger.Debug("Deleting customer via gRPC", "customer_id", customerID)

	req := &hermesv1.DeleteCustomerRequest{
		Id: customerID,
	}

	_, err := c.customerClient.DeleteCustomer(ctx, req)
	if err != nil {
		c.logger.Error("Failed to delete customer via gRPC", "customer_id", customerID, "error", err.Error())
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	c.logger.Info("Customer deleted successfully via gRPC", "customer_id", customerID)
	return nil
}

// === Contact Service Methods ===

// CreateContact creates a new contact via gRPC
func (c *CustomerGRPCClient) CreateContact(ctx context.Context, customerID uint32, contactType, value string, isPrimary bool) (*hermesv1.Contact, error) {
	c.logger.Debug("Creating contact via gRPC", "customer_id", customerID, "type", contactType)

	req := &hermesv1.CreateContactRequest{
		CustomerId: customerID,
		Type:       contactType,
		Value:      value,
		IsPrimary:  isPrimary,
	}

	resp, err := c.contactClient.CreateContact(ctx, req)
	if err != nil {
		c.logger.Error("Failed to create contact via gRPC", "customer_id", customerID, "error", err.Error())
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	c.logger.Info("Contact created successfully via gRPC", "contact_id", resp.Contact.Id, "customer_id", customerID)
	return resp.Contact, nil
}

// GetContact retrieves a contact by ID via gRPC
func (c *CustomerGRPCClient) GetContact(ctx context.Context, contactID uint32) (*hermesv1.Contact, error) {
	c.logger.Debug("Getting contact via gRPC", "contact_id", contactID)

	req := &hermesv1.GetContactRequest{
		Id: contactID,
	}

	resp, err := c.contactClient.GetContact(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get contact via gRPC", "contact_id", contactID, "error", err.Error())
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	c.logger.Debug("Contact retrieved successfully via gRPC", "contact_id", contactID)
	return resp.Contact, nil
}

// UpdateContact updates a contact via gRPC
func (c *CustomerGRPCClient) UpdateContact(ctx context.Context, contactID uint32, contactType, value string, isPrimary bool) (*hermesv1.Contact, error) {
	c.logger.Debug("Updating contact via gRPC", "contact_id", contactID, "type", contactType)

	req := &hermesv1.UpdateContactRequest{
		Id:        contactID,
		Type:      contactType,
		Value:     value,
		IsPrimary: isPrimary,
	}

	resp, err := c.contactClient.UpdateContact(ctx, req)
	if err != nil {
		c.logger.Error("Failed to update contact via gRPC", "contact_id", contactID, "error", err.Error())
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	c.logger.Info("Contact updated successfully via gRPC", "contact_id", contactID, "new_type", resp.Contact.Type)
	return resp.Contact, nil
}

// DeleteContact deletes a contact via gRPC
func (c *CustomerGRPCClient) DeleteContact(ctx context.Context, contactID uint32) error {
	c.logger.Debug("Deleting contact via gRPC", "contact_id", contactID)

	req := &hermesv1.DeleteContactRequest{
		Id: contactID,
	}

	_, err := c.contactClient.DeleteContact(ctx, req)
	if err != nil {
		c.logger.Error("Failed to delete contact via gRPC", "contact_id", contactID, "error", err.Error())
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	c.logger.Info("Contact deleted successfully via gRPC", "contact_id", contactID)
	return nil
}

// ListContactsByCustomer retrieves contacts for a customer via gRPC
func (c *CustomerGRPCClient) ListContactsByCustomer(ctx context.Context, customerID uint32) ([]*hermesv1.Contact, error) {
	c.logger.Debug("Listing contacts by customer via gRPC", "customer_id", customerID)

	req := &hermesv1.ListContactsByCustomerRequest{
		CustomerId: customerID,
	}

	resp, err := c.contactClient.ListContactsByCustomer(ctx, req)
	if err != nil {
		c.logger.Error("Failed to list contacts by customer via gRPC", "customer_id", customerID, "error", err.Error())
		return nil, fmt.Errorf("failed to list contacts by customer: %w", err)
	}

	c.logger.Debug("Contacts listed successfully via gRPC", "customer_id", customerID, "contacts_count", len(resp.Contacts))
	return resp.Contacts, nil
}

// Close closes the gRPC connection
func (c *CustomerGRPCClient) Close() error {
	if c.conn != nil {
		c.logger.Info("Closing Customer gRPC client connection")
		return c.conn.Close()
	}
	return nil
}

// GetStats returns connection statistics for monitoring
func (c *CustomerGRPCClient) GetStats() map[string]interface{} {
	if c.conn != nil {
		state := c.conn.GetState()
		return map[string]interface{}{
			"connection_state": state.String(),
			"target":           c.conn.Target(),
		}
	}
	return map[string]interface{}{
		"connection_state": "disconnected",
	}
}