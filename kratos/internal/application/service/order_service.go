package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/internal/domain/repository"
	"github.com/julesChu12/fly/kratos/pkg/constants"
	"github.com/julesChu12/fly/kratos/pkg/errors"
	"github.com/julesChu12/fly/kratos/pkg/types"
)

type OrderService interface {
	CreateOrder(ctx context.Context, req *types.CreateOrderRequest) (*types.OrderResponse, error)
	GetOrder(ctx context.Context, id uint) (*types.OrderResponse, error)
	GetOrderWithItems(ctx context.Context, id uint) (*types.OrderResponse, error)
	UpdateOrderStatus(ctx context.Context, id uint, req *types.UpdateOrderStatusRequest) (*types.OrderResponse, error)
	DeleteOrder(ctx context.Context, id uint) error
	ListOrders(ctx context.Context, req *types.ListOrdersRequest) (*types.ListResponse, error)
	GetOrderLogs(ctx context.Context, orderID uint) ([]types.OrderStatusLogResponse, error)
}

type orderService struct {
	orderRepo     repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
	statusLogRepo repository.OrderStatusLogRepository
	auditRepo     repository.OrderAuditRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	statusLogRepo repository.OrderStatusLogRepository,
	auditRepo repository.OrderAuditRepository,
) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		statusLogRepo: statusLogRepo,
		auditRepo:     auditRepo,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, req *types.CreateOrderRequest) (*types.OrderResponse, error) {
	// Get tenant ID from context
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	// Validate request
	if err := s.validateCreateOrderRequest(req); err != nil {
		return nil, err
	}

	// Check if order number already exists
	_, err := s.orderRepo.GetByOrderNo(ctx, req.OrderNo)
	if err == nil {
		return nil, errors.ErrDuplicateOrderNo
	}
	if err != errors.ErrOrderNotFound {
		if _, ok := err.(*errors.NotFoundError); !ok {
			return nil, err
		}
	}

	// Calculate total amount from items
	totalAmount := s.calculateTotalAmount(req.Items)
	if totalAmount != req.TotalAmount {
		return nil, errors.ErrInvalidAmount
	}

	// Create order entity
	order := &entity.Order{
		TenantID:    tenantID,
		OrderNo:     req.OrderNo,
		CustomerID:  req.CustomerID,
		TotalAmount: req.TotalAmount,
		Currency:    req.Currency,
		Status:      entity.OrderStatusPending,
		Remark:      req.Remark,
	}

	if order.Currency == "" {
		order.Currency = constants.DefaultCurrency
	}

	// Create order
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Create order items
	for _, itemReq := range req.Items {
		item := &entity.OrderItem{
			TenantID:    tenantID,
			OrderID:     order.ID,
			ProductID:   itemReq.ProductID,
			ProductName: itemReq.ProductName,
			SKU:         itemReq.SKU,
			Quantity:    itemReq.Quantity,
			UnitPrice:   itemReq.UnitPrice,
		}
		if err := s.orderItemRepo.Create(ctx, item); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, *item)
	}

	// Create status log
	statusLog := &entity.OrderStatusLog{
		TenantID:   tenantID,
		OrderID:    order.ID,
		FromStatus: nil,
		ToStatus:   entity.OrderStatusPending,
		Reason:     "Order created",
	}
	if err := s.statusLogRepo.Create(ctx, statusLog); err != nil {
		return nil, err
	}

	// Create audit log
	auditPayload, _ := json.Marshal(req)
	audit := &entity.OrderAudit{
		TenantID: tenantID,
		OrderID:  order.ID,
		Action:   "CREATE",
		Actor:    s.getActorFromContext(ctx),
		Payload:  string(auditPayload),
	}
	if err := s.auditRepo.Create(ctx, audit); err != nil {
		// Don't fail the request if audit fails
	}

	return s.toOrderResponse(order), nil
}

func (s *orderService) GetOrder(ctx context.Context, id uint) (*types.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toOrderResponse(order), nil
}

func (s *orderService) GetOrderWithItems(ctx context.Context, id uint) (*types.OrderResponse, error) {
	order, err := s.orderRepo.GetWithItems(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := s.toOrderResponse(order)
	resp.Items = make([]types.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		resp.Items[i] = s.toOrderItemResponse(&item)
	}

	return resp, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id uint, req *types.UpdateOrderStatusRequest) (*types.OrderResponse, error) {
	// Get tenant ID from context
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	// Get current order
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if err := s.validateStatusTransition(order.Status, req.Status); err != nil {
		return nil, err
	}

	// Update order status
	if err := s.orderRepo.UpdateStatus(ctx, id, req.Status, req.Reason, req.OperatorID); err != nil {
		return nil, err
	}

	// Create status log
	statusLog := &entity.OrderStatusLog{
		TenantID:   tenantID,
		OrderID:    order.ID,
		FromStatus: &order.Status,
		ToStatus:   req.Status,
		Reason:     req.Reason,
		OperatorID: req.OperatorID,
	}
	if err := s.statusLogRepo.Create(ctx, statusLog); err != nil {
		return nil, err
	}

	// Create audit log
	auditPayload, _ := json.Marshal(req)
	audit := &entity.OrderAudit{
		TenantID: tenantID,
		OrderID:  order.ID,
		Action:   "UPDATE_STATUS",
		Actor:    s.getActorFromContext(ctx),
		Payload:  string(auditPayload),
	}
	if err := s.auditRepo.Create(ctx, audit); err != nil {
		// Don't fail the request if audit fails
	}

	// Get updated order
	updatedOrder, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toOrderResponse(updatedOrder), nil
}

func (s *orderService) DeleteOrder(ctx context.Context, id uint) error {
	// Check if order exists
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if order can be deleted
	if order.Status != entity.OrderStatusPending {
		return errors.ErrOrderCannotBeModified
	}

	return s.orderRepo.Delete(ctx, id)
}

func (s *orderService) ListOrders(ctx context.Context, req *types.ListOrdersRequest) (*types.ListResponse, error) {
	// Get tenant ID from context
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	// Set default pagination
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

	var orders []*entity.Order
	var total int64
	var err error

	// Query based on filters
	if req.CustomerID != nil {
		orders, total, err = s.orderRepo.ListByCustomer(ctx, tenantID, *req.CustomerID, offset, req.PageSize)
	} else if req.Status != nil {
		orders, total, err = s.orderRepo.ListByStatus(ctx, tenantID, *req.Status, offset, req.PageSize)
	} else {
		orders, total, err = s.orderRepo.List(ctx, tenantID, offset, req.PageSize)
	}

	if err != nil {
		return nil, err
	}

	responses := make([]types.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = *s.toOrderResponse(order)
	}

	return &types.ListResponse{
		Data:  responses,
		Total: total,
		Page:  req.Page,
		Size:  len(responses),
	}, nil
}

func (s *orderService) GetOrderLogs(ctx context.Context, orderID uint) ([]types.OrderStatusLogResponse, error) {
	// Verify order exists and user has access
	_, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Get status logs
	logs, err := s.statusLogRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Convert to response type
	responses := make([]types.OrderStatusLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = types.OrderStatusLogResponse{
			ID:         log.ID,
			TenantID:   log.TenantID,
			OrderID:    log.OrderID,
			FromStatus: log.FromStatus,
			ToStatus:   log.ToStatus,
			Reason:     log.Reason,
			OperatorID: log.OperatorID,
			CreatedAt:  log.CreatedAt,
		}
	}

	return responses, nil
}

// Helper methods
func (s *orderService) validateCreateOrderRequest(req *types.CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return errors.ErrEmptyOrderItems
	}

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return errors.ErrInvalidQuantity
		}
		if item.UnitPrice < 0 {
			return errors.ErrInvalidAmount
		}
		if item.ProductName == "" {
			return errors.ErrInvalidRequest
		}
	}

	if req.TotalAmount < 0 {
		return errors.ErrInvalidAmount
	}

	return nil
}

func (s *orderService) calculateTotalAmount(items []types.CreateOrderItemRequest) float64 {
	total := 0.0
	for _, item := range items {
		total += float64(item.Quantity) * item.UnitPrice
	}
	return total
}

func (s *orderService) validateStatusTransition(from, to entity.OrderStatus) error {
	// Define valid transitions
	validTransitions := map[entity.OrderStatus][]entity.OrderStatus{
		entity.OrderStatusPending:   {entity.OrderStatusPaid, entity.OrderStatusCanceled},
		entity.OrderStatusPaid:      {entity.OrderStatusFulfilled, entity.OrderStatusCanceled},
		entity.OrderStatusFulfilled: {}, // Final state
		entity.OrderStatusCanceled:  {}, // Final state
	}

	allowed, exists := validTransitions[from]
	if !exists {
		return errors.ErrInvalidOrderStatus
	}

	for _, allowedStatus := range allowed {
		if allowedStatus == to {
			return nil
		}
	}

	return errors.ErrInvalidStatusTransition
}

func (s *orderService) getActorFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(constants.ContextKeyUserID).(uint); ok {
		return fmt.Sprintf("user:%d", userID)
	}
	return "system"
}

func (s *orderService) toOrderResponse(order *entity.Order) *types.OrderResponse {
	return &types.OrderResponse{
		ID:          order.ID,
		TenantID:    order.TenantID,
		OrderNo:     order.OrderNo,
		CustomerID:  order.CustomerID,
		TotalAmount: order.TotalAmount,
		Currency:    order.Currency,
		Status:      order.Status,
		Remark:      order.Remark,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

func (s *orderService) toOrderItemResponse(item *entity.OrderItem) types.OrderItemResponse {
	return types.OrderItemResponse{
		ID:          item.ID,
		TenantID:    item.TenantID,
		OrderID:     item.OrderID,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		SKU:         item.SKU,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		TotalPrice:  item.TotalPrice,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
