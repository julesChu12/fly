package grpc

import (
	"context"
	"testing"
	"time"

	orderv1 "github.com/julesChu12/fly/kratos/api/proto/order/v1"
	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/pkg/constants"
	"github.com/julesChu12/fly/kratos/pkg/errors"
	"github.com/julesChu12/fly/kratos/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockOrderService implements service.OrderService for testing
type MockOrderService struct {
	CreateOrderFunc       func(ctx context.Context, req *types.CreateOrderRequest) (*types.OrderResponse, error)
	GetOrderFunc          func(ctx context.Context, id uint) (*types.OrderResponse, error)
	GetOrderWithItemsFunc func(ctx context.Context, id uint) (*types.OrderResponse, error)
	UpdateOrderStatusFunc func(ctx context.Context, id uint, req *types.UpdateOrderStatusRequest) (*types.OrderResponse, error)
	DeleteOrderFunc       func(ctx context.Context, id uint) error
	ListOrdersFunc        func(ctx context.Context, req *types.ListOrdersRequest) (*types.ListResponse, error)
	GetOrderLogsFunc      func(ctx context.Context, orderID uint) ([]types.OrderStatusLogResponse, error)
}

func (m *MockOrderService) CreateOrder(ctx context.Context, req *types.CreateOrderRequest) (*types.OrderResponse, error) {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, req)
	}
	return &types.OrderResponse{
		ID:          1,
		TenantID:    1,
		OrderNo:     req.OrderNo,
		CustomerID:  req.CustomerID,
		TotalAmount: req.TotalAmount,
		Currency:    req.Currency,
		Status:      entity.OrderStatusPending,
		Remark:      req.Remark,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (m *MockOrderService) GetOrder(ctx context.Context, id uint) (*types.OrderResponse, error) {
	if m.GetOrderFunc != nil {
		return m.GetOrderFunc(ctx, id)
	}
	return &types.OrderResponse{
		ID:          id,
		TenantID:    1,
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Currency:    "CNY",
		Status:      entity.OrderStatusPending,
		Remark:      "Test order",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (m *MockOrderService) GetOrderWithItems(ctx context.Context, id uint) (*types.OrderResponse, error) {
	if m.GetOrderWithItemsFunc != nil {
		return m.GetOrderWithItemsFunc(ctx, id)
	}
	productID := uint(100)
	return &types.OrderResponse{
		ID:          id,
		TenantID:    1,
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Currency:    "CNY",
		Status:      entity.OrderStatusPending,
		Remark:      "Test order",
		Items: []types.OrderItemResponse{
			{
				ID:          1,
				TenantID:    1,
				OrderID:     id,
				ProductID:   &productID,
				ProductName: "Test Product",
				SKU:         "SKU001",
				Quantity:    2,
				UnitPrice:   50.0,
				TotalPrice:  100.0,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, id uint, req *types.UpdateOrderStatusRequest) (*types.OrderResponse, error) {
	if m.UpdateOrderStatusFunc != nil {
		return m.UpdateOrderStatusFunc(ctx, id, req)
	}
	return &types.OrderResponse{
		ID:          id,
		TenantID:    1,
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Currency:    "CNY",
		Status:      req.Status,
		Remark:      "Test order",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (m *MockOrderService) DeleteOrder(ctx context.Context, id uint) error {
	if m.DeleteOrderFunc != nil {
		return m.DeleteOrderFunc(ctx, id)
	}
	return nil
}

func (m *MockOrderService) ListOrders(ctx context.Context, req *types.ListOrdersRequest) (*types.ListResponse, error) {
	if m.ListOrdersFunc != nil {
		return m.ListOrdersFunc(ctx, req)
	}
	return &types.ListResponse{
		Data: []types.OrderResponse{
			{
				ID:          1,
				TenantID:    1,
				OrderNo:     "ORD001",
				CustomerID:  1,
				TotalAmount: 100.0,
				Currency:    "CNY",
				Status:      entity.OrderStatusPending,
				Remark:      "Test order",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		Total: 1,
		Page:  req.Page,
		Size:  1,
	}, nil
}

func (m *MockOrderService) GetOrderLogs(ctx context.Context, orderID uint) ([]types.OrderStatusLogResponse, error) {
	if m.GetOrderLogsFunc != nil {
		return m.GetOrderLogsFunc(ctx, orderID)
	}
	fromStatus := entity.OrderStatusPending
	operatorID := uint(10)
	return []types.OrderStatusLogResponse{
		{
			ID:         1,
			TenantID:   1,
			OrderID:    orderID,
			FromStatus: nil,
			ToStatus:   entity.OrderStatusPending,
			Reason:     "Order created",
			OperatorID: nil,
			CreatedAt:  time.Now(),
		},
		{
			ID:         2,
			TenantID:   1,
			OrderID:    orderID,
			FromStatus: &fromStatus,
			ToStatus:   entity.OrderStatusPaid,
			Reason:     "Payment received",
			OperatorID: &operatorID,
			CreatedAt:  time.Now(),
		},
	}, nil
}

func TestOrderServiceServer_CreateOrder(t *testing.T) {
	mockService := &MockOrderService{}
	server := NewOrderServiceServer(mockService)
	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	productID := uint64(100)
	req := &orderv1.CreateOrderRequest{
		OrderNo:     "ORD001",
		CustomerId:  1,
		TotalAmount: 100.0,
		Currency:    "CNY",
		Remark:      "Test order",
		Items: []*orderv1.CreateOrderItemRequest{
			{
				ProductId:   &productID,
				ProductName: "Test Product",
				Sku:         "SKU001",
				Quantity:    2,
				UnitPrice:   50.0,
			},
		},
	}

	resp, err := server.CreateOrder(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.Order.Id)
	assert.Equal(t, "ORD001", resp.Order.OrderNo)
	assert.Equal(t, orderv1.OrderStatus_ORDER_STATUS_PENDING, resp.Order.Status)
}

func TestOrderServiceServer_CreateOrder_Error(t *testing.T) {
	mockService := &MockOrderService{
		CreateOrderFunc: func(ctx context.Context, req *types.CreateOrderRequest) (*types.OrderResponse, error) {
			return nil, errors.ErrDuplicateOrderNo
		},
	}
	server := NewOrderServiceServer(mockService)
	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	req := &orderv1.CreateOrderRequest{
		OrderNo:     "ORD001",
		CustomerId:  1,
		TotalAmount: 100.0,
		Currency:    "CNY",
	}

	resp, err := server.CreateOrder(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestOrderServiceServer_GetOrder(t *testing.T) {
	mockService := &MockOrderService{}
	server := NewOrderServiceServer(mockService)
	ctx := context.Background()

	req := &orderv1.GetOrderRequest{Id: 1}

	resp, err := server.GetOrder(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.Order.Id)
	assert.Equal(t, "ORD001", resp.Order.OrderNo)
}

func TestOrderServiceServer_GetOrder_NotFound(t *testing.T) {
	mockService := &MockOrderService{
		GetOrderFunc: func(ctx context.Context, id uint) (*types.OrderResponse, error) {
			return nil, errors.ErrOrderNotFound
		},
	}
	server := NewOrderServiceServer(mockService)
	ctx := context.Background()

	req := &orderv1.GetOrderRequest{Id: 99999}

	resp, err := server.GetOrder(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestOrderServiceServer_GetOrderWithItems(t *testing.T) {
	mockService := &MockOrderService{}
	server := NewOrderServiceServer(mockService)
	ctx := context.Background()

	req := &orderv1.GetOrderRequest{Id: 1}

	resp, err := server.GetOrderWithItems(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.Order.Id)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "Test Product", resp.Items[0].ProductName)
	assert.Equal(t, int32(2), resp.Items[0].Quantity)
}

func TestOrderServiceServer_ListOrders(t *testing.T) {
	mockService := &MockOrderService{}
	server := NewOrderServiceServer(mockService)
	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	req := &orderv1.ListOrdersRequest{
		Page:     1,
		PageSize: 10,
	}

	resp, err := server.ListOrders(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Orders, 1)
	assert.Equal(t, int64(1), resp.Total)
	assert.Equal(t, int32(1), resp.Page)
}

func TestOrderServiceServer_ListOrders_WithFilters(t *testing.T) {
	mockService := &MockOrderService{
		ListOrdersFunc: func(ctx context.Context, req *types.ListOrdersRequest) (*types.ListResponse, error) {
			// Verify filters are passed correctly
			assert.NotNil(t, req.CustomerID)
			assert.Equal(t, uint(123), *req.CustomerID)
			assert.NotNil(t, req.Status)
			assert.Equal(t, entity.OrderStatusPaid, *req.Status)

			return &types.ListResponse{
				Data: []types.OrderResponse{
					{
						ID:          1,
						CustomerID:  123,
						Status:      entity.OrderStatusPaid,
						TotalAmount: 100.0,
					},
				},
				Total: 1,
				Page:  req.Page,
				Size:  1,
			}, nil
		},
	}
	server := NewOrderServiceServer(mockService)
	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	customerID := uint64(123)
	statusPaid := orderv1.OrderStatus_ORDER_STATUS_PAID
	req := &orderv1.ListOrdersRequest{
		Page:       1,
		PageSize:   10,
		CustomerId: &customerID,
		Status:     &statusPaid,
	}

	resp, err := server.ListOrders(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Orders, 1)
	assert.Equal(t, orderv1.OrderStatus_ORDER_STATUS_PAID, resp.Orders[0].Status)
}

func TestOrderServiceServer_UpdateOrderStatus(t *testing.T) {
	mockService := &MockOrderService{}
	server := NewOrderServiceServer(mockService)
	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	operatorID := uint64(10)
	req := &orderv1.UpdateOrderStatusRequest{
		Id:         1,
		Status:     orderv1.OrderStatus_ORDER_STATUS_PAID,
		Reason:     "Payment received",
		OperatorId: &operatorID,
	}

	resp, err := server.UpdateOrderStatus(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.Order.Id)
	assert.Equal(t, orderv1.OrderStatus_ORDER_STATUS_PAID, resp.Order.Status)
}

func TestOrderServiceServer_UpdateOrderStatus_InvalidTransition(t *testing.T) {
	mockService := &MockOrderService{
		UpdateOrderStatusFunc: func(ctx context.Context, id uint, req *types.UpdateOrderStatusRequest) (*types.OrderResponse, error) {
			return nil, errors.ErrInvalidStatusTransition
		},
	}
	server := NewOrderServiceServer(mockService)
	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	req := &orderv1.UpdateOrderStatusRequest{
		Id:     1,
		Status: orderv1.OrderStatus_ORDER_STATUS_FULFILLED,
		Reason: "Invalid transition",
	}

	resp, err := server.UpdateOrderStatus(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestOrderServiceServer_DeleteOrder(t *testing.T) {
	mockService := &MockOrderService{}
	server := NewOrderServiceServer(mockService)
	ctx := context.Background()

	req := &orderv1.DeleteOrderRequest{Id: 1}

	resp, err := server.DeleteOrder(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Order deleted successfully", resp.Message)
}

func TestOrderServiceServer_DeleteOrder_Error(t *testing.T) {
	mockService := &MockOrderService{
		DeleteOrderFunc: func(ctx context.Context, id uint) error {
			return errors.ErrOrderCannotBeModified
		},
	}
	server := NewOrderServiceServer(mockService)
	ctx := context.Background()

	req := &orderv1.DeleteOrderRequest{Id: 1}

	resp, err := server.DeleteOrder(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestOrderServiceServer_GetOrderLogs(t *testing.T) {
	mockService := &MockOrderService{}
	server := NewOrderServiceServer(mockService)
	ctx := context.Background()

	req := &orderv1.GetOrderLogsRequest{OrderId: 1}

	resp, err := server.GetOrderLogs(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Logs, 2)

	// First log (order created)
	assert.Equal(t, uint64(1), resp.Logs[0].Id)
	assert.Nil(t, resp.Logs[0].FromStatus)
	assert.Equal(t, orderv1.OrderStatus_ORDER_STATUS_PENDING, resp.Logs[0].ToStatus)
	assert.Equal(t, "Order created", resp.Logs[0].Reason)
	assert.Nil(t, resp.Logs[0].OperatorId)

	// Second log (status transition)
	assert.Equal(t, uint64(2), resp.Logs[1].Id)
	assert.NotNil(t, resp.Logs[1].FromStatus)
	assert.Equal(t, orderv1.OrderStatus_ORDER_STATUS_PENDING, *resp.Logs[1].FromStatus)
	assert.Equal(t, orderv1.OrderStatus_ORDER_STATUS_PAID, resp.Logs[1].ToStatus)
	assert.Equal(t, "Payment received", resp.Logs[1].Reason)
	assert.NotNil(t, resp.Logs[1].OperatorId)
	assert.Equal(t, uint64(10), *resp.Logs[1].OperatorId)
}

func TestOrderServiceServer_GetOrderLogs_OrderNotFound(t *testing.T) {
	mockService := &MockOrderService{
		GetOrderLogsFunc: func(ctx context.Context, orderID uint) ([]types.OrderStatusLogResponse, error) {
			return nil, errors.ErrOrderNotFound
		},
	}
	server := NewOrderServiceServer(mockService)
	ctx := context.Background()

	req := &orderv1.GetOrderLogsRequest{OrderId: 99999}

	resp, err := server.GetOrderLogs(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestOrderServiceServer_GetOrderLogs_EmptyLogs(t *testing.T) {
	mockService := &MockOrderService{
		GetOrderLogsFunc: func(ctx context.Context, orderID uint) ([]types.OrderStatusLogResponse, error) {
			return []types.OrderStatusLogResponse{}, nil
		},
	}
	server := NewOrderServiceServer(mockService)
	ctx := context.Background()

	req := &orderv1.GetOrderLogsRequest{OrderId: 1}

	resp, err := server.GetOrderLogs(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Logs)
}

func TestConvertError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode codes.Code
	}{
		{"Unauthorized", errors.ErrUnauthorized, codes.Unauthenticated},
		{"OrderNotFound", errors.ErrOrderNotFound, codes.NotFound},
		{"DuplicateOrderNo", errors.ErrDuplicateOrderNo, codes.AlreadyExists},
		{"InvalidAmount", errors.ErrInvalidAmount, codes.InvalidArgument},
		{"InvalidQuantity", errors.ErrInvalidQuantity, codes.InvalidArgument},
		{"EmptyOrderItems", errors.ErrEmptyOrderItems, codes.InvalidArgument},
		{"InvalidOrderStatus", errors.ErrInvalidOrderStatus, codes.InvalidArgument},
		{"InvalidStatusTransition", errors.ErrInvalidStatusTransition, codes.FailedPrecondition},
		{"OrderCannotBeModified", errors.ErrOrderCannotBeModified, codes.FailedPrecondition},
		{"InvalidRequest", errors.ErrInvalidRequest, codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcErr := convertError(tt.err)
			assert.Equal(t, tt.expectedCode, status.Code(grpcErr))
		})
	}
}

func TestToProtoOrderStatus(t *testing.T) {
	tests := []struct {
		entity entity.OrderStatus
		proto  orderv1.OrderStatus
	}{
		{entity.OrderStatusPending, orderv1.OrderStatus_ORDER_STATUS_PENDING},
		{entity.OrderStatusPaid, orderv1.OrderStatus_ORDER_STATUS_PAID},
		{entity.OrderStatusFulfilled, orderv1.OrderStatus_ORDER_STATUS_FULFILLED},
		{entity.OrderStatusCanceled, orderv1.OrderStatus_ORDER_STATUS_CANCELED},
		{"unknown", orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(string(tt.entity), func(t *testing.T) {
			result := toProtoOrderStatus(tt.entity)
			assert.Equal(t, tt.proto, result)
		})
	}
}

func TestToEntityOrderStatus(t *testing.T) {
	tests := []struct {
		proto  orderv1.OrderStatus
		entity entity.OrderStatus
	}{
		{orderv1.OrderStatus_ORDER_STATUS_PENDING, entity.OrderStatusPending},
		{orderv1.OrderStatus_ORDER_STATUS_PAID, entity.OrderStatusPaid},
		{orderv1.OrderStatus_ORDER_STATUS_FULFILLED, entity.OrderStatusFulfilled},
		{orderv1.OrderStatus_ORDER_STATUS_CANCELED, entity.OrderStatusCanceled},
		{orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED, entity.OrderStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.proto.String(), func(t *testing.T) {
			result := toEntityOrderStatus(tt.proto)
			assert.Equal(t, tt.entity, result)
		})
	}
}

func TestToProtoOrder(t *testing.T) {
	now := time.Now()
	order := &types.OrderResponse{
		ID:          1,
		TenantID:    10,
		OrderNo:     "ORD001",
		CustomerID:  100,
		TotalAmount: 299.99,
		Currency:    "USD",
		Status:      entity.OrderStatusPaid,
		Remark:      "Test order",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	protoOrder := toProtoOrder(order)

	assert.Equal(t, uint64(1), protoOrder.Id)
	assert.Equal(t, uint64(10), protoOrder.TenantId)
	assert.Equal(t, "ORD001", protoOrder.OrderNo)
	assert.Equal(t, uint64(100), protoOrder.CustomerId)
	assert.Equal(t, 299.99, protoOrder.TotalAmount)
	assert.Equal(t, "USD", protoOrder.Currency)
	assert.Equal(t, orderv1.OrderStatus_ORDER_STATUS_PAID, protoOrder.Status)
	assert.Equal(t, "Test order", protoOrder.Remark)
	assert.NotNil(t, protoOrder.CreatedAt)
	assert.NotNil(t, protoOrder.UpdatedAt)
}

func TestToProtoOrderItem(t *testing.T) {
	now := time.Now()
	productID := uint(100)
	item := &types.OrderItemResponse{
		ID:          1,
		TenantID:    10,
		OrderID:     5,
		ProductID:   &productID,
		ProductName: "Test Product",
		SKU:         "SKU001",
		Quantity:    3,
		UnitPrice:   99.99,
		TotalPrice:  299.97,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	protoItem := toProtoOrderItem(item)

	assert.Equal(t, uint64(1), protoItem.Id)
	assert.Equal(t, uint64(10), protoItem.TenantId)
	assert.Equal(t, uint64(5), protoItem.OrderId)
	assert.NotNil(t, protoItem.ProductId)
	assert.Equal(t, uint64(100), *protoItem.ProductId)
	assert.Equal(t, "Test Product", protoItem.ProductName)
	assert.Equal(t, "SKU001", protoItem.Sku)
	assert.Equal(t, int32(3), protoItem.Quantity)
	assert.Equal(t, 99.99, protoItem.UnitPrice)
	assert.Equal(t, 299.97, protoItem.TotalPrice)
	assert.NotNil(t, protoItem.CreatedAt)
	assert.NotNil(t, protoItem.UpdatedAt)
}

func TestToProtoOrderItem_NilProductID(t *testing.T) {
	now := time.Now()
	item := &types.OrderItemResponse{
		ID:          1,
		TenantID:    10,
		OrderID:     5,
		ProductID:   nil, // Test nil product ID
		ProductName: "Test Product",
		SKU:         "SKU001",
		Quantity:    3,
		UnitPrice:   99.99,
		TotalPrice:  299.97,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	protoItem := toProtoOrderItem(item)

	assert.Nil(t, protoItem.ProductId)
	assert.Equal(t, "Test Product", protoItem.ProductName)
}
