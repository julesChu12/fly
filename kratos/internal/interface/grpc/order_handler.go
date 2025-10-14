package grpc

import (
	"context"

	"github.com/julesChu12/fly/kratos/api/proto/order/v1"
	"github.com/julesChu12/fly/kratos/internal/application/service"
	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/pkg/errors"
	"github.com/julesChu12/fly/kratos/pkg/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrderServiceServer implements the gRPC OrderService
type OrderServiceServer struct {
	orderv1.UnimplementedOrderServiceServer
	orderService service.OrderService
}

// NewOrderServiceServer creates a new OrderServiceServer
func NewOrderServiceServer(orderService service.OrderService) *OrderServiceServer {
	return &OrderServiceServer{
		orderService: orderService,
	}
}

// CreateOrder creates a new order
func (s *OrderServiceServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.OrderResponse, error) {
	// Convert protobuf request to internal type
	items := make([]types.CreateOrderItemRequest, len(req.Items))
	for i, item := range req.Items {
		var productID *uint
		if item.ProductId != nil {
			pid := uint(*item.ProductId)
			productID = &pid
		}
		items[i] = types.CreateOrderItemRequest{
			ProductID:   productID,
			ProductName: item.ProductName,
			SKU:         item.Sku,
			Quantity:    int(item.Quantity),
			UnitPrice:   item.UnitPrice,
		}
	}

	createReq := &types.CreateOrderRequest{
		OrderNo:     req.OrderNo,
		CustomerID:  uint(req.CustomerId),
		TotalAmount: req.TotalAmount,
		Currency:    req.Currency,
		Remark:      req.Remark,
		Items:       items,
	}

	// Call service
	result, err := s.orderService.CreateOrder(ctx, createReq)
	if err != nil {
		return nil, convertError(err)
	}

	// Convert response to protobuf
	return &orderv1.OrderResponse{
		Order: toProtoOrder(result),
	}, nil
}

// GetOrder retrieves an order by ID
func (s *OrderServiceServer) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.OrderResponse, error) {
	result, err := s.orderService.GetOrder(ctx, uint(req.Id))
	if err != nil {
		return nil, convertError(err)
	}

	return &orderv1.OrderResponse{
		Order: toProtoOrder(result),
	}, nil
}

// GetOrderWithItems retrieves an order with its items
func (s *OrderServiceServer) GetOrderWithItems(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.OrderWithItemsResponse, error) {
	result, err := s.orderService.GetOrderWithItems(ctx, uint(req.Id))
	if err != nil {
		return nil, convertError(err)
	}

	items := make([]*orderv1.OrderItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = toProtoOrderItem(&item)
	}

	return &orderv1.OrderWithItemsResponse{
		Order: toProtoOrder(result),
		Items: items,
	}, nil
}

// ListOrders retrieves a paginated list of orders
func (s *OrderServiceServer) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	// Convert request
	listReq := &types.ListOrdersRequest{}
	listReq.Page = int(req.Page)
	listReq.PageSize = int(req.PageSize)

	if req.CustomerId != nil {
		cid := uint(*req.CustomerId)
		listReq.CustomerID = &cid
	}

	if req.Status != nil {
		status := toEntityOrderStatus(*req.Status)
		listReq.Status = &status
	}

	// Call service
	result, err := s.orderService.ListOrders(ctx, listReq)
	if err != nil {
		return nil, convertError(err)
	}

	// Convert response
	orders := make([]*orderv1.Order, len(result.Data.([]types.OrderResponse)))
	for i, order := range result.Data.([]types.OrderResponse) {
		orders[i] = &orderv1.Order{
			Id:          uint64(order.ID),
			TenantId:    uint64(order.TenantID),
			OrderNo:     order.OrderNo,
			CustomerId:  uint64(order.CustomerID),
			TotalAmount: order.TotalAmount,
			Currency:    order.Currency,
			Status:      toProtoOrderStatus(order.Status),
			Remark:      order.Remark,
			CreatedAt:   timestamppb.New(order.CreatedAt),
			UpdatedAt:   timestamppb.New(order.UpdatedAt),
		}
	}

	return &orderv1.ListOrdersResponse{
		Orders: orders,
		Total:  result.Total,
		Page:   int32(result.Page),
		Size:   int32(result.Size),
	}, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderServiceServer) UpdateOrderStatus(ctx context.Context, req *orderv1.UpdateOrderStatusRequest) (*orderv1.OrderResponse, error) {
	// Convert request
	updateReq := &types.UpdateOrderStatusRequest{
		Status: toEntityOrderStatus(req.Status),
		Reason: req.Reason,
	}

	if req.OperatorId != nil {
		oid := uint(*req.OperatorId)
		updateReq.OperatorID = &oid
	}

	// Call service
	result, err := s.orderService.UpdateOrderStatus(ctx, uint(req.Id), updateReq)
	if err != nil {
		return nil, convertError(err)
	}

	return &orderv1.OrderResponse{
		Order: toProtoOrder(result),
	}, nil
}

// DeleteOrder deletes an order (soft delete)
func (s *OrderServiceServer) DeleteOrder(ctx context.Context, req *orderv1.DeleteOrderRequest) (*orderv1.DeleteOrderResponse, error) {
	err := s.orderService.DeleteOrder(ctx, uint(req.Id))
	if err != nil {
		return nil, convertError(err)
	}

	return &orderv1.DeleteOrderResponse{
		Success: true,
		Message: "Order deleted successfully",
	}, nil
}

// GetOrderLogs retrieves status change logs for an order
func (s *OrderServiceServer) GetOrderLogs(ctx context.Context, req *orderv1.GetOrderLogsRequest) (*orderv1.GetOrderLogsResponse, error) {
	logs, err := s.orderService.GetOrderLogs(ctx, uint(req.OrderId))
	if err != nil {
		return nil, convertError(err)
	}

	// Convert to protobuf
	protoLogs := make([]*orderv1.OrderStatusLog, len(logs))
	for i, log := range logs {
		protoLog := &orderv1.OrderStatusLog{
			Id:        uint64(log.ID),
			TenantId:  uint64(log.TenantID),
			OrderId:   uint64(log.OrderID),
			ToStatus:  toProtoOrderStatus(log.ToStatus),
			Reason:    log.Reason,
			CreatedAt: timestamppb.New(log.CreatedAt),
		}

		if log.FromStatus != nil {
			fromStatus := toProtoOrderStatus(*log.FromStatus)
			protoLog.FromStatus = &fromStatus
		}

		if log.OperatorID != nil {
			operatorID := uint64(*log.OperatorID)
			protoLog.OperatorId = &operatorID
		}

		protoLogs[i] = protoLog
	}

	return &orderv1.GetOrderLogsResponse{
		Logs: protoLogs,
	}, nil
}

// Helper functions to convert between protobuf and internal types

func toProtoOrder(order *types.OrderResponse) *orderv1.Order {
	return &orderv1.Order{
		Id:          uint64(order.ID),
		TenantId:    uint64(order.TenantID),
		OrderNo:     order.OrderNo,
		CustomerId:  uint64(order.CustomerID),
		TotalAmount: order.TotalAmount,
		Currency:    order.Currency,
		Status:      toProtoOrderStatus(order.Status),
		Remark:      order.Remark,
		CreatedAt:   timestamppb.New(order.CreatedAt),
		UpdatedAt:   timestamppb.New(order.UpdatedAt),
	}
}

func toProtoOrderItem(item *types.OrderItemResponse) *orderv1.OrderItem {
	var productID *uint64
	if item.ProductID != nil {
		pid := uint64(*item.ProductID)
		productID = &pid
	}

	return &orderv1.OrderItem{
		Id:          uint64(item.ID),
		TenantId:    uint64(item.TenantID),
		OrderId:     uint64(item.OrderID),
		ProductId:   productID,
		ProductName: item.ProductName,
		Sku:         item.SKU,
		Quantity:    int32(item.Quantity),
		UnitPrice:   item.UnitPrice,
		TotalPrice:  item.TotalPrice,
		CreatedAt:   timestamppb.New(item.CreatedAt),
		UpdatedAt:   timestamppb.New(item.UpdatedAt),
	}
}

func toProtoOrderStatus(status entity.OrderStatus) orderv1.OrderStatus {
	switch status {
	case entity.OrderStatusPending:
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	case entity.OrderStatusPaid:
		return orderv1.OrderStatus_ORDER_STATUS_PAID
	case entity.OrderStatusFulfilled:
		return orderv1.OrderStatus_ORDER_STATUS_FULFILLED
	case entity.OrderStatusCanceled:
		return orderv1.OrderStatus_ORDER_STATUS_CANCELED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func toEntityOrderStatus(status orderv1.OrderStatus) entity.OrderStatus {
	switch status {
	case orderv1.OrderStatus_ORDER_STATUS_PENDING:
		return entity.OrderStatusPending
	case orderv1.OrderStatus_ORDER_STATUS_PAID:
		return entity.OrderStatusPaid
	case orderv1.OrderStatus_ORDER_STATUS_FULFILLED:
		return entity.OrderStatusFulfilled
	case orderv1.OrderStatus_ORDER_STATUS_CANCELED:
		return entity.OrderStatusCanceled
	default:
		return entity.OrderStatusPending
	}
}

func convertError(err error) error {
	switch {
	case err == errors.ErrUnauthorized:
		return status.Error(codes.Unauthenticated, err.Error())
	case err == errors.ErrOrderNotFound:
		return status.Error(codes.NotFound, err.Error())
	case err == errors.ErrDuplicateOrderNo:
		return status.Error(codes.AlreadyExists, err.Error())
	case err == errors.ErrInvalidAmount:
		return status.Error(codes.InvalidArgument, "Invalid amount")
	case err == errors.ErrInvalidQuantity:
		return status.Error(codes.InvalidArgument, "Invalid quantity")
	case err == errors.ErrEmptyOrderItems:
		return status.Error(codes.InvalidArgument, "Order items cannot be empty")
	case err == errors.ErrInvalidOrderStatus:
		return status.Error(codes.InvalidArgument, "Invalid order status")
	case err == errors.ErrInvalidStatusTransition:
		return status.Error(codes.FailedPrecondition, "Invalid status transition")
	case err == errors.ErrOrderCannotBeModified:
		return status.Error(codes.FailedPrecondition, "Order cannot be modified")
	case err == errors.ErrInvalidRequest:
		return status.Error(codes.InvalidArgument, "Invalid request")
	default:
		return status.Error(codes.Internal, "Internal server error")
	}
}
