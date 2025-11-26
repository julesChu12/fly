package usecase

import (
	"context"
	"fmt"

	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/clotho/internal/middleware"
	hermesv1 "github.com/julesChu12/fly/hermes/api/proto"
	orderv1 "github.com/julesChu12/fly/kratos/api/proto/order/v1"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// CustomerListResult wraps customer list results for uniform return types
type CustomerListResult struct {
	Customers []client.Customer `json:"customers"`
	Total     int64            `json:"total"`
}

// ProtocolAwareProxy represents a service proxy that can use both HTTP and gRPC
type ProtocolAwareProxy struct {
	serviceName     string
	protocolSelector *middleware.ProtocolSelector
	httpClient       interface{}
	grpcClient       interface{}
	logger           *logger.Logger
}

// NewProtocolAwareProxy creates a new protocol-aware service proxy
func NewProtocolAwareProxy(
	serviceName string,
	httpClient interface{},
	grpcClient interface{},
	protocolSelector *middleware.ProtocolSelector,
) *ProtocolAwareProxy {
	logger := logger.NewDefault()

	// Register clients with protocol selector
	if httpClient != nil {
		protocolSelector.RegisterHTTPClient(serviceName, httpClient)
	}
	if grpcClient != nil {
		protocolSelector.RegisterGRPCClient(serviceName, grpcClient)
	}

	proxy := &ProtocolAwareProxy{
		serviceName:      serviceName,
		protocolSelector: protocolSelector,
		httpClient:        httpClient,
		grpcClient:        grpcClient,
		logger:            logger,
	}

	logger.Info("Created protocol-aware proxy",
		"service", serviceName,
		"http_client", httpClient != nil,
		"grpc_client", grpcClient != nil)

	return proxy
}

// SmartCustomerProxy represents a protocol-aware customer service proxy
type SmartCustomerProxy struct {
	*ProtocolAwareProxy
	httpCustomerClient *client.CustomerHTTPClient
	grpcCustomerClient *client.CustomerGRPCClient
}

// NewSmartCustomerProxy creates a new protocol-aware customer proxy
func NewSmartCustomerProxy(
	httpClient *client.CustomerHTTPClient,
	grpcClient *client.CustomerGRPCClient,
	protocolSelector *middleware.ProtocolSelector,
) *SmartCustomerProxy {
	baseProxy := NewProtocolAwareProxy("hermes", httpClient, grpcClient, protocolSelector)

	return &SmartCustomerProxy{
		ProtocolAwareProxy: baseProxy,
		httpCustomerClient: httpClient,
		grpcCustomerClient: grpcClient,
	}
}

// CreateCustomer creates a customer using the optimal protocol
func (p *SmartCustomerProxy) CreateCustomer(
	ctx context.Context,
	name, phone, email, tags string,
	requestSize int64,
) (*client.Customer, error) {
	result, err := p.protocolSelector.WrapExecution(
		ctx,
		p.serviceName,
		"/api/customers",
		requestSize,
		func(ctx context.Context, protocol middleware.ProtocolType) (interface{}, error) {
			switch protocol {
			case middleware.ProtocolGRPC:
				if p.grpcCustomerClient == nil {
					return nil, fmt.Errorf("gRPC client not available")
				}
				return p.grpcCustomerClient.CreateCustomer(ctx, name, phone, email, tags)
			case middleware.ProtocolHTTP:
				if p.httpCustomerClient == nil {
					return nil, fmt.Errorf("HTTP client not available")
				}
				req := &client.CreateCustomerRequestHTTP{
					Name:  name,
					Phone: phone,
					Email: email,
					Tags:  tags,
				}
				return p.httpCustomerClient.CreateCustomer(ctx, req)
			default:
				return nil, fmt.Errorf("unsupported protocol: %s", protocol)
			}
		},
	)

	if err != nil {
		return nil, err
	}

	// Type assertion based on what was actually used
	switch v := result.(type) {
	case *hermesv1.Customer:
		return convertHermesCustomerToHTTPCustomer(v), nil
	case *client.Customer:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
}

// GetCustomer retrieves a customer using the optimal protocol
func (p *SmartCustomerProxy) GetCustomer(
	ctx context.Context,
	customerID uint,
	requestSize int64,
) (*client.Customer, error) {
	result, err := p.protocolSelector.WrapExecution(
		ctx,
		p.serviceName,
		fmt.Sprintf("/api/customers/%d", customerID),
		requestSize,
		func(ctx context.Context, protocol middleware.ProtocolType) (interface{}, error) {
			switch protocol {
			case middleware.ProtocolGRPC:
				if p.grpcCustomerClient == nil {
					return nil, fmt.Errorf("gRPC client not available")
				}
				return p.grpcCustomerClient.GetCustomer(ctx, uint32(customerID))
			case middleware.ProtocolHTTP:
				if p.httpCustomerClient == nil {
					return nil, fmt.Errorf("HTTP client not available")
				}
				return p.httpCustomerClient.GetCustomer(ctx, customerID)
			default:
				return nil, fmt.Errorf("unsupported protocol: %s", protocol)
			}
		},
	)

	if err != nil {
		return nil, err
	}

	// Type assertion based on what was actually used
	switch v := result.(type) {
	case *hermesv1.Customer:
		return convertHermesCustomerToHTTPCustomer(v), nil
	case *client.Customer:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
}

// ListCustomers retrieves customers using the optimal protocol
func (p *SmartCustomerProxy) ListCustomers(
	ctx context.Context,
	page, limit int,
	searchTerm string,
	requestSize int64,
) ([]client.Customer, int64, error) {
	result, err := p.protocolSelector.WrapExecution(
		ctx,
		p.serviceName,
		"/api/customers",
		requestSize,
		func(ctx context.Context, protocol middleware.ProtocolType) (interface{}, error) {
			switch protocol {
			case middleware.ProtocolGRPC:
				if p.grpcCustomerClient == nil {
					return nil, fmt.Errorf("gRPC client not available")
				}
				return p.grpcCustomerClient.ListCustomers(ctx, int32(page), int32(limit))
			case middleware.ProtocolHTTP:
				if p.httpCustomerClient == nil {
					return nil, fmt.Errorf("HTTP client not available")
				}
				filter := &client.CustomerFilter{
					Search: searchTerm,
					Page:   page,
					Limit:  limit,
				}
				customers, total, err := p.httpCustomerClient.ListCustomers(ctx, filter)
				if err != nil {
					return nil, err
				}
				return &CustomerListResult{Customers: customers, Total: total}, nil
			default:
				return nil, fmt.Errorf("unsupported protocol: %s", protocol)
			}
		},
	)

	if err != nil {
		return nil, 0, err
	}

	// Type assertion and conversion
	switch v := result.(type) {
	case *hermesv1.ListCustomersResponse:
		customers := make([]client.Customer, len(v.Customers))
		for i, grpcCustomer := range v.Customers {
			customers[i] = *convertHermesCustomerToHTTPCustomer(grpcCustomer)
		}
		return customers, v.Total, nil
	case *CustomerListResult:
		// HTTP client wrapped result
		return v.Customers, v.Total, nil
	default:
		return nil, 0, fmt.Errorf("unexpected result type: %T", result)
	}
}

// SmartOrderProxy represents a protocol-aware order service proxy
type SmartOrderProxy struct {
	*ProtocolAwareProxy
	httpOrderClient *client.OrderHTTPClient
	grpcOrderClient *client.OrderGRPCClient
}

// NewSmartOrderProxy creates a new protocol-aware order proxy
func NewSmartOrderProxy(
	httpClient *client.OrderHTTPClient,
	grpcClient *client.OrderGRPCClient,
	protocolSelector *middleware.ProtocolSelector,
) *SmartOrderProxy {
	baseProxy := NewProtocolAwareProxy("kratos", httpClient, grpcClient, protocolSelector)

	return &SmartOrderProxy{
		ProtocolAwareProxy: baseProxy,
		httpOrderClient:    httpClient,
		grpcOrderClient:    grpcClient,
	}
}

// CreateOrder creates an order using the optimal protocol
func (p *SmartOrderProxy) CreateOrder(
	ctx context.Context,
	orderNo string,
	customerID uint,
	totalAmount float64,
	currency string,
	remark string,
	items []client.CreateOrderItemRequestHTTP,
	requestSize int64,
) (*client.Order, error) {
	result, err := p.protocolSelector.WrapExecution(
		ctx,
		p.serviceName,
		"/api/orders",
		requestSize,
		func(ctx context.Context, protocol middleware.ProtocolType) (interface{}, error) {
			switch protocol {
			case middleware.ProtocolGRPC:
				if p.grpcOrderClient == nil {
					return nil, fmt.Errorf("gRPC client not available")
				}
				// Convert items
				grpcItems := make([]*orderv1.CreateOrderItemRequest, len(items))
				for i, item := range items {
					grpcItems[i] = &orderv1.CreateOrderItemRequest{
						ProductName: item.ProductName,
						Sku:         item.SKU,
						Quantity:    int32(item.Quantity),
						UnitPrice:   item.UnitPrice,
					}
					if item.ProductID != nil {
						productID := uint64(*item.ProductID)
						grpcItems[i].ProductId = &productID
					}
				}

				req := &orderv1.CreateOrderRequest{
					OrderNo:     orderNo,
					CustomerId:  uint64(customerID),
					TotalAmount: totalAmount,
					Currency:    currency,
					Remark:      remark,
					Items:       grpcItems,
				}
				return p.grpcOrderClient.CreateOrder(ctx, req)

			case middleware.ProtocolHTTP:
				if p.httpOrderClient == nil {
					return nil, fmt.Errorf("HTTP client not available")
				}
				req := &client.CreateOrderRequestHTTP{
					OrderNo:     orderNo,
					CustomerID:  customerID,
					TotalAmount: totalAmount,
					Currency:    currency,
					Remark:      remark,
					Items:       items,
				}
				return p.httpOrderClient.CreateOrder(ctx, req)

			default:
				return nil, fmt.Errorf("unsupported protocol: %s", protocol)
			}
		},
	)

	if err != nil {
		return nil, err
	}

	// Type assertion based on what was actually used
	switch v := result.(type) {
	case *orderv1.Order:
		return convertKratosOrderToHTTPOrder(v), nil
	case *client.Order:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
}

// GetOrder retrieves an order using the optimal protocol
func (p *SmartOrderProxy) GetOrder(
	ctx context.Context,
	orderID uint,
	requestSize int64,
) (*client.Order, error) {
	result, err := p.protocolSelector.WrapExecution(
		ctx,
		p.serviceName,
		fmt.Sprintf("/api/orders/%d", orderID),
		requestSize,
		func(ctx context.Context, protocol middleware.ProtocolType) (interface{}, error) {
			switch protocol {
			case middleware.ProtocolGRPC:
				if p.grpcOrderClient == nil {
					return nil, fmt.Errorf("gRPC client not available")
				}
				return p.grpcOrderClient.GetOrder(ctx, uint64(orderID))
			case middleware.ProtocolHTTP:
				if p.httpOrderClient == nil {
					return nil, fmt.Errorf("HTTP client not available")
				}
				return p.httpOrderClient.GetOrder(ctx, orderID)
			default:
				return nil, fmt.Errorf("unsupported protocol: %s", protocol)
			}
		},
	)

	if err != nil {
		return nil, err
	}

	// Type assertion based on what was actually used
	switch v := result.(type) {
	case *orderv1.Order:
		return convertKratosOrderToHTTPOrder(v), nil
	case *client.Order:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
}

// Helper conversion functions

// convertHermesCustomerToHTTPCustomer converts Hermes gRPC Customer to HTTP Customer
func convertHermesCustomerToHTTPCustomer(grpcCustomer *hermesv1.Customer) *client.Customer {
	if grpcCustomer == nil {
		return nil
	}

	httpCustomer := &client.Customer{
		ID:        uint(grpcCustomer.Id),
		Name:      grpcCustomer.Name,
		Phone:     grpcCustomer.Phone,
		Email:     grpcCustomer.Email,
		Tags:      grpcCustomer.Tags,
		CreatedAt: grpcCustomer.CreatedAt.AsTime(),
		UpdatedAt: grpcCustomer.UpdatedAt.AsTime(),
	}

	// Convert contacts
	if grpcCustomer.Contacts != nil {
		httpCustomer.Contacts = make([]client.Contact, len(grpcCustomer.Contacts))
		for i, grpcContact := range grpcCustomer.Contacts {
			httpCustomer.Contacts[i] = client.Contact{
				ID:         uint(grpcContact.Id),
				CustomerID: uint(grpcContact.CustomerId),
				Type:       grpcContact.Type,
				Value:      grpcContact.Value,
				IsPrimary:  grpcContact.IsPrimary,
				CreatedAt:  grpcContact.CreatedAt.AsTime(),
				UpdatedAt:  grpcContact.UpdatedAt.AsTime(),
			}
		}
	}

	return httpCustomer
}

// convertKratosOrderToHTTPOrder converts Kratos gRPC Order to HTTP Order
func convertKratosOrderToHTTPOrder(grpcOrder *orderv1.Order) *client.Order {
	if grpcOrder == nil {
		return nil
	}

	httpOrder := &client.Order{
		ID:          uint(grpcOrder.Id),
		TenantID:    uint(grpcOrder.TenantId),
		OrderNo:     grpcOrder.OrderNo,
		CustomerID:  uint(grpcOrder.CustomerId),
		TotalAmount: grpcOrder.TotalAmount,
		Currency:    grpcOrder.Currency,
		Remark:      grpcOrder.Remark,
		CreatedAt:   grpcOrder.CreatedAt.AsTime(),
		UpdatedAt:   grpcOrder.UpdatedAt.AsTime(),
		Status:      client.OrderStatus(convertKratosOrderStatus(grpcOrder.Status)),
	}

	return httpOrder
}

// convertKratosOrderStatus converts Kratos gRPC OrderStatus to HTTP OrderStatus
func convertKratosOrderStatus(grpcStatus orderv1.OrderStatus) client.OrderStatus {
	switch grpcStatus {
	case orderv1.OrderStatus_ORDER_STATUS_PENDING:
		return client.OrderStatusPending
	case orderv1.OrderStatus_ORDER_STATUS_PAID:
		return client.OrderStatusConfirmed
	case orderv1.OrderStatus_ORDER_STATUS_FULFILLED:
		return client.OrderStatusShipped
	case orderv1.OrderStatus_ORDER_STATUS_CANCELED:
		return client.OrderStatusCancelled
	default:
		return client.OrderStatusPending
	}
}

// GetProtocolMetrics returns current protocol metrics for monitoring
func (p *ProtocolAwareProxy) GetProtocolMetrics() map[middleware.ProtocolType]*middleware.ProtocolMetrics {
	return p.protocolSelector.GetMetrics()
}

// GetHealthStatus returns current health status
func (p *ProtocolAwareProxy) GetHealthStatus() map[middleware.ProtocolType]bool {
	return p.protocolSelector.GetHealthStatus()
}