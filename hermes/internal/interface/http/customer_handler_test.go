package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/hermes/pkg/types"
)

// MockCustomerService 模拟客户服务
type MockCustomerService struct{}

func (m *MockCustomerService) CreateCustomer(ctx context.Context, req *types.CreateCustomerRequest) (*types.CustomerResponse, error) {
	return &types.CustomerResponse{
		ID:    1,
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}, nil
}

func (m *MockCustomerService) GetCustomer(ctx context.Context, id uint) (*types.CustomerResponse, error) {
	return &types.CustomerResponse{
		ID:    id,
		Name:  "测试客户",
		Email: "test@example.com",
		Phone: "13800138000",
	}, nil
}

func (m *MockCustomerService) GetCustomerWithContacts(ctx context.Context, id uint) (*types.CustomerResponse, error) {
	return &types.CustomerResponse{
		ID:    id,
		Name:  "测试客户",
		Email: "test@example.com",
		Phone: "13800138000",
	}, nil
}

func (m *MockCustomerService) UpdateCustomer(ctx context.Context, id uint, req *types.UpdateCustomerRequest) (*types.CustomerResponse, error) {
	return &types.CustomerResponse{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}, nil
}

func (m *MockCustomerService) DeleteCustomer(ctx context.Context, id uint) error {
	return nil
}

func (m *MockCustomerService) ListCustomers(ctx context.Context, req *types.ListRequest) (*types.ListResponse, error) {
	return &types.ListResponse{
		Data:  []interface{}{},
		Total: 0,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// TestCreateCustomer 测试创建客户接口
func TestCreateCustomer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := &MockCustomerService{}
	handler := NewCustomerHandler(mockService)

	// 准备测试数据
	req := &types.CreateCustomerRequest{
		Name:  "测试客户",
		Email: "test@example.com",
		Phone: "13800138000",
	}

	// 创建请求
	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/customers", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// 执行请求
	handler.CreateCustomer(c)

	// 验证响应
	if w.Code != http.StatusCreated {
		t.Errorf("期望状态码 %d, 但得到 %d", http.StatusCreated, w.Code)
	}
}

// TestGetCustomer 测试获取客户接口
func TestGetCustomer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := &MockCustomerService{}
	handler := NewCustomerHandler(mockService)

	// 创建请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/customers/1", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	// 执行请求
	handler.GetCustomer(c)

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 但得到 %d", http.StatusOK, w.Code)
	}
}

// TestGetCustomer_InvalidID 测试无效ID
func TestGetCustomer_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := &MockCustomerService{}
	handler := NewCustomerHandler(mockService)

	// 创建请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/customers/invalid", nil)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	// 执行请求
	handler.GetCustomer(c)

	// 验证响应
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 但得到 %d", http.StatusBadRequest, w.Code)
	}
}
