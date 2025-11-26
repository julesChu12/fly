package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	orderv1 "github.com/julesChu12/fly/kratos/api/proto/order/v1"
	"github.com/julesChu12/fly/mora/pkg/discovery"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// OrderGRPCClient represents a gRPC client for the Kratos (Order) service
type OrderGRPCClient struct {
	conn   *grpc.ClientConn
	client orderv1.OrderServiceClient
	logger *logger.Logger
}

// OrderGRPCClientConfig represents configuration for the gRPC client
type OrderGRPCClientConfig struct {
	MaxRecvMsgSize    int           `yaml:"max_recv_msg_size"`
	MaxSendMsgSize    int           `yaml:"max_send_msg_size"`
	ConnectTimeout    time.Duration `yaml:"connect_timeout"`
	KeepaliveTime     time.Duration `yaml:"keepalive_time"`
	KeepaliveTimeout  time.Duration `yaml:"keepalive_timeout"`
	EnableRetry       bool          `yaml:"enable_retry"`
	MaxRetries        int           `yaml:"max_retries"`
	PermitWithoutStream bool        `yaml:"permit_without_stream"`
}

// DefaultOrderGRPCClientConfig returns default configuration for the gRPC client
func DefaultOrderGRPCClientConfig() *OrderGRPCClientConfig {
	return &OrderGRPCClientConfig{
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

// NewOrderGRPCClient creates a new Order gRPC client with optimized settings
func NewOrderGRPCClient(target string, config *OrderGRPCClientConfig) (*OrderGRPCClient, error) {
	if config == nil {
		config = DefaultOrderGRPCClientConfig()
	}

	log := logger.NewDefault()
	log.Info("Creating optimized Order gRPC client", "target", target)

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
		log.Error("Failed to connect to Order gRPC service", "error", err.Error())
		return nil, fmt.Errorf("failed to connect to Order gRPC service: %w", err)
	}

	client := orderv1.NewOrderServiceClient(conn)

	log.Info("Successfully connected to Order gRPC service", "target", target)

	return &OrderGRPCClient{
		conn:   conn,
		client: client,
		logger: log,
	}, nil
}

// NewOrderGRPCClientWithDiscovery uses service discovery to create an Order gRPC client
func NewOrderGRPCClientWithDiscovery(disc discovery.Discovery, config *OrderGRPCClientConfig) (*OrderGRPCClient, error) {
	if config == nil {
		config = DefaultOrderGRPCClientConfig()
	}

	log := logger.NewDefault()

	// Get Kratos service address from discovery
	instance, err := disc.GetService(context.Background(), "kratos")
	if err != nil {
		log.Error("Failed to discover Kratos service", "error", err.Error())
		return nil, fmt.Errorf("failed to discover Kratos service: %w", err)
	}

	address := instance.Address()
	target := fmt.Sprintf("%s:%d", address, 9092) // Assuming gRPC port 9092
	log.Info("Discovered Kratos gRPC service", "address", target, "instance_id", instance.ID)

	return NewOrderGRPCClient(target, config)
}

// CreateOrder creates a new order via gRPC
func (c *OrderGRPCClient) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.Order, error) {
	c.logger.Debug("Creating order via gRPC", "order_no", req.OrderNo, "customer_id", req.CustomerId)

	resp, err := c.client.CreateOrder(ctx, req)
	if err != nil {
		c.logger.Error("Failed to create order via gRPC", "error", err.Error())
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	c.logger.Info("Order created successfully via gRPC", "order_id", resp.Order.Id, "order_no", resp.Order.OrderNo)
	return resp.Order, nil
}

// GetOrder retrieves an order by ID via gRPC
func (c *OrderGRPCClient) GetOrder(ctx context.Context, orderID uint64) (*orderv1.Order, error) {
	c.logger.Debug("Getting order via gRPC", "order_id", orderID)

	req := &orderv1.GetOrderRequest{
		Id: orderID,
	}

	resp, err := c.client.GetOrder(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get order via gRPC", "order_id", orderID, "error", err.Error())
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	c.logger.Debug("Order retrieved successfully via gRPC", "order_id", orderID)
	return resp.Order, nil
}

// GetOrderWithItems retrieves an order with items via gRPC
func (c *OrderGRPCClient) GetOrderWithItems(ctx context.Context, orderID uint64) (*orderv1.Order, []*orderv1.OrderItem, error) {
	c.logger.Debug("Getting order with items via gRPC", "order_id", orderID)

	req := &orderv1.GetOrderRequest{
		Id: orderID,
	}

	resp, err := c.client.GetOrderWithItems(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get order with items via gRPC", "order_id", orderID, "error", err.Error())
		return nil, nil, fmt.Errorf("failed to get order with items: %w", err)
	}

	c.logger.Debug("Order with items retrieved successfully via gRPC", "order_id", orderID, "items_count", len(resp.Items))
	return resp.Order, resp.Items, nil
}

// ListOrders retrieves a paginated list of orders via gRPC
func (c *OrderGRPCClient) ListOrders(ctx context.Context, page, pageSize int32, customerID *uint64, status *orderv1.OrderStatus) (*orderv1.ListOrdersResponse, error) {
	c.logger.Debug("Listing orders via gRPC", "page", page, "page_size", pageSize)

	req := &orderv1.ListOrdersRequest{
		Page:     page,
		PageSize: pageSize,
	}

	if customerID != nil {
		req.CustomerId = customerID
	}
	if status != nil {
		req.Status = status
	}

	resp, err := c.client.ListOrders(ctx, req)
	if err != nil {
		c.logger.Error("Failed to list orders via gRPC", "error", err.Error())
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	c.logger.Debug("Orders listed successfully via gRPC", "total", resp.Total, "page", resp.Page)
	return resp, nil
}

// UpdateOrderStatus updates the status of an order via gRPC
func (c *OrderGRPCClient) UpdateOrderStatus(ctx context.Context, orderID uint64, status orderv1.OrderStatus, reason string, operatorID *uint64) (*orderv1.Order, error) {
	c.logger.Debug("Updating order status via gRPC", "order_id", orderID, "status", status.String())

	req := &orderv1.UpdateOrderStatusRequest{
		Id:     orderID,
		Status: status,
		Reason: reason,
	}

	if operatorID != nil {
		req.OperatorId = operatorID
	}

	resp, err := c.client.UpdateOrderStatus(ctx, req)
	if err != nil {
		c.logger.Error("Failed to update order status via gRPC", "order_id", orderID, "error", err.Error())
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	c.logger.Info("Order status updated successfully via gRPC", "order_id", orderID, "new_status", status.String())
	return resp.Order, nil
}

// DeleteOrder deletes an order via gRPC
func (c *OrderGRPCClient) DeleteOrder(ctx context.Context, orderID uint64) error {
	c.logger.Debug("Deleting order via gRPC", "order_id", orderID)

	req := &orderv1.DeleteOrderRequest{
		Id: orderID,
	}

	_, err := c.client.DeleteOrder(ctx, req)
	if err != nil {
		c.logger.Error("Failed to delete order via gRPC", "order_id", orderID, "error", err.Error())
		return fmt.Errorf("failed to delete order: %w", err)
	}

	c.logger.Info("Order deleted successfully via gRPC", "order_id", orderID)
	return nil
}

// GetOrderLogs retrieves status change logs for an order via gRPC
func (c *OrderGRPCClient) GetOrderLogs(ctx context.Context, orderID uint64) ([]*orderv1.OrderStatusLog, error) {
	c.logger.Debug("Getting order logs via gRPC", "order_id", orderID)

	req := &orderv1.GetOrderLogsRequest{
		OrderId: orderID,
	}

	resp, err := c.client.GetOrderLogs(ctx, req)
	if err != nil {
		c.logger.Error("Failed to get order logs via gRPC", "order_id", orderID, "error", err.Error())
		return nil, fmt.Errorf("failed to get order logs: %w", err)
	}

	c.logger.Debug("Order logs retrieved successfully via gRPC", "order_id", orderID, "logs_count", len(resp.Logs))
	return resp.Logs, nil
}

// Close closes the gRPC connection
func (c *OrderGRPCClient) Close() error {
	if c.conn != nil {
		c.logger.Info("Closing Order gRPC client connection")
		return c.conn.Close()
	}
	return nil
}

// GetStats returns connection statistics for monitoring
func (c *OrderGRPCClient) GetStats() map[string]interface{} {
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