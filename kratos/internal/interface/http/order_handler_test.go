package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/pkg/types"
)

// Mock order service for testing
type MockOrderService struct{}

func (m *MockOrderService) CreateOrder(ctx context.Context, req *types.CreateOrderRequest) (*types.OrderResponse, error) {
	return &types.OrderResponse{
		ID:          1,
		TenantID:    1,
		OrderNo:     req.OrderNo,
		CustomerID:  req.CustomerID,
		TotalAmount: req.TotalAmount,
		Currency:    req.Currency,
		Status:      entity.OrderStatusPending,
		Remark:      req.Remark,
	}, nil
}

func (m *MockOrderService) GetOrder(ctx context.Context, id uint) (*types.OrderResponse, error) {
	return &types.OrderResponse{
		ID:          id,
		TenantID:    1,
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Currency:    "CNY",
		Status:      entity.OrderStatusPending,
		Remark:      "Test order",
	}, nil
}

func (m *MockOrderService) GetOrderWithItems(ctx context.Context, id uint) (*types.OrderResponse, error) {
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
				OrderID:     id,
				ProductName: "Test Product",
				Quantity:    2,
				UnitPrice:   50.0,
				TotalPrice:  100.0,
			},
		},
	}, nil
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, id uint, req *types.UpdateOrderStatusRequest) (*types.OrderResponse, error) {
	return &types.OrderResponse{
		ID:          id,
		TenantID:    1,
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Currency:    "CNY",
		Status:      req.Status,
		Remark:      "Test order",
	}, nil
}

func (m *MockOrderService) DeleteOrder(ctx context.Context, id uint) error {
	return nil
}

func (m *MockOrderService) ListOrders(ctx context.Context, req *types.ListOrdersRequest) (*types.ListResponse, error) {
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
			},
		},
		Total: 1,
		Page:  req.Page,
		Size:  1,
	}, nil
}

func (m *MockOrderService) GetOrderLogs(ctx context.Context, orderID uint) ([]types.OrderStatusLogResponse, error) {
	return []types.OrderStatusLogResponse{
		{
			ID:        1,
			OrderID:   orderID,
			ToStatus:  entity.OrderStatusPending,
			CreatedAt: time.Now(),
		},
	}, nil
}

func TestCreateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockOrderService{}
	handler := NewOrderHandler(mockService)

	req := &types.CreateOrderRequest{
		OrderNo:     "ORD001",
		CustomerID:  1,
		TotalAmount: 100.0,
		Currency:    "CNY",
		Remark:      "Test order",
		Items: []types.CreateOrderItemRequest{
			{
				ProductName: "Test Product",
				Quantity:    2,
				UnitPrice:   50.0,
			},
		},
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/orders", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Tenant-ID", "1")
	c.Request.Header.Set("X-User-ID", "1")

	handler.CreateOrder(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response types.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected response code 0, got %d", response.Code)
	}
}

func TestGetOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockOrderService{}
	handler := NewOrderHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/orders/1", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Request.Header.Set("X-Tenant-ID", "1")

	handler.GetOrder(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response types.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected response code 0, got %d", response.Code)
	}
}

func TestGetOrder_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockOrderService{}
	handler := NewOrderHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/orders/invalid", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	handler.GetOrder(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetOrderWithItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockOrderService{}
	handler := NewOrderHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/orders/1/items", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Request.Header.Set("X-Tenant-ID", "1")

	handler.GetOrderWithItems(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response types.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected response code 0, got %d", response.Code)
	}

	// Check if items are included
	orderData := response.Data.(map[string]interface{})
	items := orderData["items"]
	if items == nil {
		t.Error("Expected items to be included in response")
	}
}

func TestGetOrderLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockOrderService{}
	handler := NewOrderHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/orders/1/logs", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Request.Header.Set("X-Tenant-ID", "1")

	handler.GetOrderLogs(c)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response types.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		t.Fatalf("Expected response code 0, got %d", response.Code)
	}
}

func TestUpdateOrderStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockOrderService{}
	handler := NewOrderHandler(mockService)

	req := &types.UpdateOrderStatusRequest{
		Status: entity.OrderStatusPaid,
		Reason: "Payment received",
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PATCH", "/api/orders/1/status", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Tenant-ID", "1")
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	handler.UpdateOrderStatus(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response types.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected response code 0, got %d", response.Code)
	}
}

func TestDeleteOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockOrderService{}
	handler := NewOrderHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", "/api/orders/1", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Request.Header.Set("X-Tenant-ID", "1")

	handler.DeleteOrder(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response types.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected response code 0, got %d", response.Code)
	}
}

func TestListOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockOrderService{}
	handler := NewOrderHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/orders?page=1&page_size=10", nil)
	c.Request.Header.Set("X-Tenant-ID", "1")

	handler.ListOrders(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response types.Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("Expected response code 0, got %d", response.Code)
	}

	// Check if data is a list response
	listData := response.Data.(map[string]interface{})
	if listData["data"] == nil {
		t.Error("Expected data field in list response")
	}
	if listData["total"] == nil {
		t.Error("Expected total field in list response")
	}
}
