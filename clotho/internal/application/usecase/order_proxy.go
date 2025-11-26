package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// OrderProxy represents the use case for order operations
type OrderProxy struct {
	orderClient *client.OrderHTTPClient
	logger      *logger.Logger
}

// NewOrderProxy creates a new OrderProxy instance
func NewOrderProxy(orderClient *client.OrderHTTPClient, logger *logger.Logger) *OrderProxy {
	return &OrderProxy{
		orderClient: orderClient,
		logger:      logger,
	}
}

// ListOrders retrieves a list of orders with filters
func (p *OrderProxy) ListOrders(ctx context.Context, filter *client.OrderFilter) (*client.OrderListResponse, error) {
	p.logger.Info("Retrieving order list", "filter", filter)

	// Set defaults for filter
	if filter != nil {
		if filter.Page <= 0 {
			filter.Page = 1
		}
		if filter.PageSize <= 0 {
			filter.PageSize = 20
		}
	}

	start := time.Now()
	orders, err := p.orderClient.ListOrders(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve order list", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve order list: %w", err)
	}

	p.logger.Info("Successfully retrieved order list", "count", len(orders.Data), "total", orders.Total, "duration", duration)
	return orders, nil
}

// CreateOrder creates a new order
func (p *OrderProxy) CreateOrder(ctx context.Context, req *client.CreateOrderRequestHTTP) (*client.Order, error) {
	p.logger.Info("Creating new order", "order_no", req.OrderNo, "customer_id", req.CustomerID)

	// Validate request
	if err := p.validateCreateOrderRequest(req); err != nil {
		p.logger.Error("Invalid order creation request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	order, err := p.orderClient.CreateOrder(ctx, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to create order", "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	p.logger.Info("Successfully created order", "order_id", order.ID, "order_no", order.OrderNo, "duration", duration)
	return order, nil
}

// GetOrder retrieves an order by ID
func (p *OrderProxy) GetOrder(ctx context.Context, orderID uint) (*client.Order, error) {
	p.logger.Info("Retrieving order", "order_id", orderID)

	if orderID == 0 {
		return nil, fmt.Errorf("order ID is required")
	}

	start := time.Now()
	order, err := p.orderClient.GetOrder(ctx, orderID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve order", "order_id", orderID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve order: %w", err)
	}

	p.logger.Info("Successfully retrieved order", "order_id", order.ID, "order_no", order.OrderNo, "duration", duration)
	return order, nil
}

// GetOrderWithItems retrieves an order with all items by ID
func (p *OrderProxy) GetOrderWithItems(ctx context.Context, orderID uint) (*client.Order, error) {
	p.logger.Info("Retrieving order with items", "order_id", orderID)

	if orderID == 0 {
		return nil, fmt.Errorf("order ID is required")
	}

	start := time.Now()
	order, err := p.orderClient.GetOrderWithItems(ctx, orderID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve order with items", "order_id", orderID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve order with items: %w", err)
	}

	p.logger.Info("Successfully retrieved order with items", "order_id", order.ID, "order_no", order.OrderNo, "items_count", len(order.Items), "duration", duration)
	return order, nil
}

// GetOrderLogs retrieves status change logs for an order
func (p *OrderProxy) GetOrderLogs(ctx context.Context, orderID uint) ([]client.OrderStatusLog, error) {
	p.logger.Info("Retrieving order logs", "order_id", orderID)

	if orderID == 0 {
		return nil, fmt.Errorf("order ID is required")
	}

	start := time.Now()
	logs, err := p.orderClient.GetOrderLogs(ctx, orderID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve order logs", "order_id", orderID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve order logs: %w", err)
	}

	p.logger.Info("Successfully retrieved order logs", "order_id", orderID, "logs_count", len(logs), "duration", duration)
	return logs, nil
}

// UpdateOrderStatus updates the status of an existing order
func (p *OrderProxy) UpdateOrderStatus(ctx context.Context, orderID uint, req *client.UpdateOrderStatusRequestHTTP) (*client.Order, error) {
	p.logger.Info("Updating order status", "order_id", orderID, "status", req.Status)

	if orderID == 0 {
		return nil, fmt.Errorf("order ID is required")
	}

	// Validate that order exists first
	_, err := p.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		p.logger.Error("Order not found for status update", "order_id", orderID, "error", err.Error())
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Validate request
	if err := p.validateUpdateOrderStatusRequest(req); err != nil {
		p.logger.Error("Invalid order status update request", "error", err.Error())
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	updatedOrder, err := p.orderClient.UpdateOrderStatus(ctx, orderID, req)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to update order status", "order_id", orderID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	p.logger.Info("Successfully updated order status", "order_id", updatedOrder.ID, "order_no", updatedOrder.OrderNo, "new_status", updatedOrder.Status, "duration", duration)
	return updatedOrder, nil
}

// DeleteOrder deletes an order
func (p *OrderProxy) DeleteOrder(ctx context.Context, orderID uint) error {
	p.logger.Info("Deleting order", "order_id", orderID)

	if orderID == 0 {
		return fmt.Errorf("order ID is required")
	}

	// Validate that order exists first
	_, err := p.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		p.logger.Error("Order not found for deletion", "order_id", orderID, "error", err.Error())
		return fmt.Errorf("order not found: %w", err)
	}

	start := time.Now()
	err = p.orderClient.DeleteOrder(ctx, orderID)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to delete order", "order_id", orderID, "error", err.Error(), "duration", duration)
		return fmt.Errorf("failed to delete order: %w", err)
	}

	p.logger.Info("Successfully deleted order", "order_id", orderID, "duration", duration)
	return nil
}

// GetOrdersByCustomer retrieves orders for a specific customer
func (p *OrderProxy) GetOrdersByCustomer(ctx context.Context, customerID uint, page, pageSize int) (*client.OrderListResponse, error) {
	p.logger.Info("Retrieving orders for customer", "customer_id", customerID)

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

	filter := &client.OrderFilter{
		CustomerID: &customerID,
		Page:       page,
		PageSize:   pageSize,
	}

	start := time.Now()
	orders, err := p.orderClient.ListOrders(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve customer orders", "customer_id", customerID, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve customer orders: %w", err)
	}

	p.logger.Info("Successfully retrieved customer orders", "customer_id", customerID, "count", len(orders.Data), "total", orders.Total, "duration", duration)
	return orders, nil
}

// GetOrdersByStatus retrieves orders by status
func (p *OrderProxy) GetOrdersByStatus(ctx context.Context, status client.OrderStatus, page, pageSize int) (*client.OrderListResponse, error) {
	p.logger.Info("Retrieving orders by status", "status", status)

	// Set defaults
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	filter := &client.OrderFilter{
		Status:   &status,
		Page:     page,
		PageSize: pageSize,
	}

	start := time.Now()
	orders, err := p.orderClient.ListOrders(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("Failed to retrieve orders by status", "status", status, "error", err.Error(), "duration", duration)
		return nil, fmt.Errorf("failed to retrieve orders by status: %w", err)
	}

	p.logger.Info("Successfully retrieved orders by status", "status", status, "count", len(orders.Data), "total", orders.Total, "duration", duration)
	return orders, nil
}

// Validation helper functions

func (p *OrderProxy) validateCreateOrderRequest(req *client.CreateOrderRequestHTTP) error {
	if req.OrderNo == "" {
		return fmt.Errorf("order number is required")
	}
	if req.CustomerID == 0 {
		return fmt.Errorf("customer ID is required")
	}
	if req.TotalAmount <= 0 {
		return fmt.Errorf("total amount must be greater than 0")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("order items are required")
	}
	for i, item := range req.Items {
		if item.ProductName == "" {
			return fmt.Errorf("item %d: product name is required", i+1)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be greater than 0", i+1)
		}
		if item.UnitPrice <= 0 {
			return fmt.Errorf("item %d: unit price must be greater than 0", i+1)
		}
	}
	return nil
}

func (p *OrderProxy) validateUpdateOrderStatusRequest(req *client.UpdateOrderStatusRequestHTTP) error {
	if req.Status == "" {
		return fmt.Errorf("status is required")
	}

	// Validate status is a valid order status
	validStatuses := map[client.OrderStatus]bool{
		client.OrderStatusPending:    true,
		client.OrderStatusConfirmed:  true,
		client.OrderStatusProcessing: true,
		client.OrderStatusShipped:    true,
		client.OrderStatusDelivered:  true,
		client.OrderStatusCancelled:  true,
		client.OrderStatusRefunded:   true,
	}

	if !validStatuses[req.Status] {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	return nil
}