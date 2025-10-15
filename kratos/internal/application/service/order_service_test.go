package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/pkg/constants"
	"github.com/julesChu12/fly/kratos/pkg/errors"
	"github.com/julesChu12/fly/kratos/pkg/types"
)

// Mock repositories for testing
type MockOrderRepository struct {
	orders       map[uint]*entity.Order
	nextID       uint
	getByIDCalls int
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders: make(map[uint]*entity.Order),
		nextID: 1,
	}
}

func (m *MockOrderRepository) Create(ctx context.Context, order *entity.Order) error {
	order.ID = m.nextID
	m.nextID++
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id uint) (*entity.Order, error) {
	m.getByIDCalls++
	order, exists := m.orders[id]
	if !exists {
		return nil, errors.ErrOrderNotFound
	}
	return order, nil
}

func (m *MockOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error) {
	for _, order := range m.orders {
		if order.OrderNo == orderNo {
			return order, nil
		}
	}
	return nil, errors.ErrOrderNotFound
}

func (m *MockOrderRepository) GetWithItems(ctx context.Context, id uint) (*entity.Order, error) {
	order, exists := m.orders[id]
	if !exists {
		return nil, errors.ErrOrderNotFound
	}
	return order, nil
}

func (m *MockOrderRepository) Update(ctx context.Context, order *entity.Order) error {
	if _, exists := m.orders[order.ID]; !exists {
		return errors.ErrOrderNotFound
	}
	m.orders[order.ID] = order
	return nil
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id uint, status entity.OrderStatus, reason string, operatorID *uint) error {
	order, exists := m.orders[id]
	if !exists {
		return errors.ErrOrderNotFound
	}
	order.Status = status
	return nil
}

func (m *MockOrderRepository) Delete(ctx context.Context, id uint) error {
	if _, exists := m.orders[id]; !exists {
		return errors.ErrOrderNotFound
	}
	delete(m.orders, id)
	return nil
}

func (m *MockOrderRepository) CallsToGetByID() int {
	return m.getByIDCalls
}

func (m *MockOrderRepository) List(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	for _, order := range m.orders {
		if order.TenantID == tenantID {
			orders = append(orders, order)
		}
	}

	total := int64(len(orders))
	start := offset
	end := offset + limit
	if start > len(orders) {
		return []*entity.Order{}, total, nil
	}
	if end > len(orders) {
		end = len(orders)
	}

	return orders[start:end], total, nil
}

func (m *MockOrderRepository) ListByCustomer(ctx context.Context, tenantID uint, customerID uint, offset, limit int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	for _, order := range m.orders {
		if order.TenantID == tenantID && order.CustomerID == customerID {
			orders = append(orders, order)
		}
	}

	total := int64(len(orders))
	start := offset
	end := offset + limit
	if start > len(orders) {
		return []*entity.Order{}, total, nil
	}
	if end > len(orders) {
		end = len(orders)
	}

	return orders[start:end], total, nil
}

func (m *MockOrderRepository) ListByStatus(ctx context.Context, tenantID uint, status entity.OrderStatus, offset, limit int) ([]*entity.Order, int64, error) {
	var orders []*entity.Order
	for _, order := range m.orders {
		if order.TenantID == tenantID && order.Status == status {
			orders = append(orders, order)
		}
	}

	total := int64(len(orders))
	start := offset
	end := offset + limit
	if start > len(orders) {
		return []*entity.Order{}, total, nil
	}
	if end > len(orders) {
		end = len(orders)
	}

	return orders[start:end], total, nil
}

type MockOrderItemRepository struct {
	items  map[uint]*entity.OrderItem
	nextID uint
}

func NewMockOrderItemRepository() *MockOrderItemRepository {
	return &MockOrderItemRepository{
		items:  make(map[uint]*entity.OrderItem),
		nextID: 1,
	}
}

func (m *MockOrderItemRepository) Create(ctx context.Context, item *entity.OrderItem) error {
	item.ID = m.nextID
	m.nextID++
	m.items[item.ID] = item
	return nil
}

func (m *MockOrderItemRepository) GetByID(ctx context.Context, id uint) (*entity.OrderItem, error) {
	item, exists := m.items[id]
	if !exists {
		return nil, errors.ErrOrderItemNotFound
	}
	return item, nil
}

func (m *MockOrderItemRepository) GetByOrderID(ctx context.Context, orderID uint) ([]*entity.OrderItem, error) {
	var items []*entity.OrderItem
	for _, item := range m.items {
		if item.OrderID == orderID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (m *MockOrderItemRepository) Update(ctx context.Context, item *entity.OrderItem) error {
	if _, exists := m.items[item.ID]; !exists {
		return errors.ErrOrderItemNotFound
	}
	m.items[item.ID] = item
	return nil
}

func (m *MockOrderItemRepository) Delete(ctx context.Context, id uint) error {
	if _, exists := m.items[id]; !exists {
		return errors.ErrOrderItemNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *MockOrderItemRepository) DeleteByOrderID(ctx context.Context, orderID uint) error {
	for id, item := range m.items {
		if item.OrderID == orderID {
			delete(m.items, id)
		}
	}
	return nil
}

type MockOrderStatusLogRepository struct {
	logs   []*entity.OrderStatusLog
	nextID uint
}

func NewMockOrderStatusLogRepository() *MockOrderStatusLogRepository {
	return &MockOrderStatusLogRepository{
		logs:   make([]*entity.OrderStatusLog, 0),
		nextID: 1,
	}
}

func (m *MockOrderStatusLogRepository) Create(ctx context.Context, log *entity.OrderStatusLog) error {
	log.ID = m.nextID
	m.nextID++
	m.logs = append(m.logs, log)
	return nil
}

func (m *MockOrderStatusLogRepository) GetByOrderID(ctx context.Context, orderID uint) ([]*entity.OrderStatusLog, error) {
	var logs []*entity.OrderStatusLog
	for _, log := range m.logs {
		if log.OrderID == orderID {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func (m *MockOrderStatusLogRepository) List(ctx context.Context, offset, limit int) ([]*entity.OrderStatusLog, error) {
	start := offset
	end := offset + limit
	if start > len(m.logs) {
		return []*entity.OrderStatusLog{}, nil
	}
	if end > len(m.logs) {
		end = len(m.logs)
	}
	return m.logs[start:end], nil
}

type MockOrderAuditRepository struct {
	audits []*entity.OrderAudit
	nextID uint
}

func NewMockOrderAuditRepository() *MockOrderAuditRepository {
	return &MockOrderAuditRepository{
		audits: make([]*entity.OrderAudit, 0),
		nextID: 1,
	}
}

func (m *MockOrderAuditRepository) Create(ctx context.Context, audit *entity.OrderAudit) error {
	audit.ID = m.nextID
	m.nextID++
	m.audits = append(m.audits, audit)
	return nil
}

func (m *MockOrderAuditRepository) GetByOrderID(ctx context.Context, orderID uint) ([]*entity.OrderAudit, error) {
	var audits []*entity.OrderAudit
	for _, audit := range m.audits {
		if audit.OrderID == orderID {
			audits = append(audits, audit)
		}
	}
	return audits, nil
}

func (m *MockOrderAuditRepository) List(ctx context.Context, offset, limit int) ([]*entity.OrderAudit, error) {
	start := offset
	end := offset + limit
	if start > len(m.audits) {
		return []*entity.OrderAudit{}, nil
	}
	if end > len(m.audits) {
		end = len(m.audits)
	}
	return m.audits[start:end], nil
}

type mockOrderCache struct {
	data map[string][]byte
}

func newMockOrderCache() *mockOrderCache {
	return &mockOrderCache{
		data: make(map[string][]byte),
	}
}

func (m *mockOrderCache) Get(ctx context.Context, key string, dest interface{}) error {
	payload, ok := m.data[key]
	if !ok {
		return fmt.Errorf("cache miss")
	}
	return json.Unmarshal(payload, dest)
}

func (m *mockOrderCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.data[key] = bytes
	return nil
}

func (m *mockOrderCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// Test cases
func TestOrderService_CreateOrder(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()
	service := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	req := &types.CreateOrderRequest{
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Currency:    "CNY",
		Items: []types.CreateOrderItemRequest{
			{
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   50.0,
			},
		},
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))
	order, err := service.CreateOrder(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if order.OrderNo != req.OrderNo {
		t.Errorf("Expected order number %s, got %s", req.OrderNo, order.OrderNo)
	}

	if order.Status != entity.OrderStatusPending {
		t.Errorf("Expected status %s, got %s", entity.OrderStatusPending, order.Status)
	}
}

func TestOrderService_CreateOrder_DuplicateOrderNo(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()
	service := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	req := &types.CreateOrderRequest{
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Items: []types.CreateOrderItemRequest{
			{
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   50.0,
			},
		},
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	// Create first order
	_, err := service.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error for first order, got %v", err)
	}

	// Try to create duplicate
	_, err = service.CreateOrder(ctx, req)
	if err != errors.ErrDuplicateOrderNo {
		t.Errorf("Expected duplicate order error, got %v", err)
	}
}

func TestOrderService_GetOrder(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()
	service := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	// Create an order first
	req := &types.CreateOrderRequest{
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Items: []types.CreateOrderItemRequest{
			{
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   50.0,
			},
		},
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))
	created, err := service.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Get the order
	order, err := service.GetOrder(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get order: %v", err)
	}

	if order.ID != created.ID {
		t.Errorf("Expected order ID %d, got %d", created.ID, order.ID)
	}
}

func TestOrderService_GetOrder_UsesCache(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()
	cache := newMockOrderCache()

	order := &entity.Order{
		ID:          1,
		TenantID:    1,
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100,
		Currency:    "CNY",
		Status:      entity.OrderStatusPending,
	}
	orderRepo.orders[order.ID] = order

	service := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo, WithOrderCache(cache, time.Minute))

	ctx := context.Background()

	if _, err := service.GetOrder(ctx, order.ID); err != nil {
		t.Fatalf("expected first fetch to succeed: %v", err)
	}
	if _, err := service.GetOrder(ctx, order.ID); err != nil {
		t.Fatalf("expected cached fetch to succeed: %v", err)
	}

	if got := orderRepo.CallsToGetByID(); got != 1 {
		t.Fatalf("expected repository to be hit once, got %d", got)
	}
}

func TestOrderService_GetOrderLogs(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()
	service := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	order := &entity.Order{
		ID:          1,
		TenantID:    42,
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 123,
		Currency:    "CNY",
		Status:      entity.OrderStatusPending,
	}
	orderRepo.orders[order.ID] = order

	logEntry := &entity.OrderStatusLog{
		TenantID:  order.TenantID,
		OrderID:   order.ID,
		ToStatus:  entity.OrderStatusPending,
		Reason:    "created",
		CreatedAt: time.Now(),
	}
	if err := statusRepo.Create(context.Background(), logEntry); err != nil {
		t.Fatalf("failed to seed log: %v", err)
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, order.TenantID)
	logs, err := service.GetOrderLogs(ctx, order.ID)
	if err != nil {
		t.Fatalf("expected logs to be returned: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].Reason != "created" {
		t.Fatalf("unexpected log reason: %s", logs[0].Reason)
	}
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()
	service := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	// Create an order first
	req := &types.CreateOrderRequest{
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Items: []types.CreateOrderItemRequest{
			{
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   50.0,
			},
		},
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))
	created, err := service.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Update status
	updateReq := &types.UpdateOrderStatusRequest{
		Status: entity.OrderStatusPaid,
		Reason: "Payment received",
	}

	updated, err := service.UpdateOrderStatus(ctx, created.ID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update order status: %v", err)
	}

	if updated.Status != entity.OrderStatusPaid {
		t.Errorf("Expected status %s, got %s", entity.OrderStatusPaid, updated.Status)
	}
}

func TestOrderService_UpdateOrderStatus_InvalidTransition(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()
	service := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	// Create an order first
	req := &types.CreateOrderRequest{
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Items: []types.CreateOrderItemRequest{
			{
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   50.0,
			},
		},
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))
	created, err := service.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	// Try invalid transition (pending -> fulfilled)
	updateReq := &types.UpdateOrderStatusRequest{
		Status: entity.OrderStatusFulfilled,
		Reason: "Invalid transition",
	}

	_, err = service.UpdateOrderStatus(ctx, created.ID, updateReq)
	if err != errors.ErrInvalidStatusTransition {
		t.Errorf("Expected invalid status transition error, got %v", err)
	}
}

func TestOrderService_ListOrders(t *testing.T) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()
	service := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	// Create multiple orders
	for i := 1; i <= 5; i++ {
		req := &types.CreateOrderRequest{
			OrderNo:     fmt.Sprintf("ORD%03d", i),
			CustomerID:  uint(i),
			TotalAmount: float64(100 * i),
			Items: []types.CreateOrderItemRequest{
				{
					ProductName: fmt.Sprintf("Product %d", i),
					Quantity:    1,
					UnitPrice:   float64(100 * i),
				},
			},
		}
		_, err := service.CreateOrder(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create order %d: %v", i, err)
		}
	}

	// Test list orders
	listReq := &types.ListOrdersRequest{
		ListRequest: types.ListRequest{
			Page:     1,
			PageSize: 3,
		},
	}

	result, err := service.ListOrders(ctx, listReq)
	if err != nil {
		t.Fatalf("Failed to list orders: %v", err)
	}

	if result.Total != 5 {
		t.Errorf("Expected total 5, got %d", result.Total)
	}

	orders := result.Data.([]types.OrderResponse)
	if len(orders) != 3 {
		t.Errorf("Expected 3 orders in page, got %d", len(orders))
	}
}
