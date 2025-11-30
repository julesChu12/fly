package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/spf13/viper"

	"github.com/julesChu12/fly/items/internal/application/service"
	"github.com/julesChu12/fly/items/internal/infrastructure/http/router"
	"github.com/julesChu12/fly/items/internal/infrastructure/persistence/model"
	"github.com/julesChu12/fly/items/internal/infrastructure/persistence/repository"
)

// setupTestServer 设置测试HTTP服务器
func setupTestServer(t *testing.T) (*gin.Engine, *TestContainer, service.ItemService) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 设置测试容器
	ctx := context.Background()
	container, err := SetupTestContainers(ctx)
	require.NoError(t, err, "Failed to setup test containers")

	// 创建表结构
	err = container.Database.AutoMigrate(&model.Item{}, &model.Category{})
	require.NoError(t, err, "Failed to auto-migrate tables")

	// 创建仓储和服务实例
	itemRepo := repository.NewItemRepository(container.Database)
	itemService := service.NewItemService(itemRepo)

	// 设置配置
	cfg := viper.New()
	cfg.Set("app.mode", "test")
	cfg.Set("server.port", "0") // 随机端口
	cfg.Set("database.driver", "mysql")
	cfg.Set("database.host", container.MySQLHost)
	cfg.Set("database.port", container.MySQLPort)
	cfg.Set("database.username", "testuser")
	cfg.Set("database.password", "testpass")
	cfg.Set("database.database", "items_test")

	// 设置路由
	ginRouter, _ := router.SetupRouter(cfg)

	return ginRouter, container, itemService
}

// TestItemAPIEndpoints 测试商品API端点
func TestItemAPIEndpoints(t *testing.T) {
	router, container, _ := setupTestServer(t)
	defer container.Cleanup(context.Background())

	t.Run("CreateItem", func(t *testing.T) {
		// 首先创建一个分类
		categoryID := uuid.New()
		err := container.Database.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "电子产品分类", "电子设备分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		// 准备请求数据
		requestBody := map[string]interface{}{
			"name":         "测试商品",
			"description":  "这是一个测试商品",
			"type":         "PRODUCT",
			"price":        99.99,
			"category_id":  categoryID.String(),
			"stock":        100,
			"cost_price":   80.00,
			"sku":          "TEST-001",
		}

		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)

		// 创建HTTP请求
		req, err := http.NewRequest("POST", "/api/v1/items", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// 使用httptest记录响应
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "CREATED", response["code"])
		assert.Equal(t, "Item created successfully", response["message"])
	})

	t.Run("GetItems", func(t *testing.T) {
		// 创建HTTP请求
		req, err := http.NewRequest("GET", "/api/v1/items", nil)
		require.NoError(t, err)

		// 记录响应
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "SUCCESS", response["code"])
		assert.Contains(t, response, "data")
	})

	t.Run("GetItemsWithPagination", func(t *testing.T) {
		// 创建带分页参数的请求
		req, err := http.NewRequest("GET", "/api/v1/items?page=1&page_size=10", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response["data"].(map[string]interface{})
		pagination := data["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(10), pagination["page_size"])
	})

	t.Run("GetItemsWithFilters", func(t *testing.T) {
		// 创建带过滤参数的请求
		req, err := http.NewRequest("GET", "/api/v1/items?type=PRODUCT", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "SUCCESS", response["code"])
	})

	t.Run("GetItemByID", func(t *testing.T) {
		// 首先创建一个商品以便获取ID
		categoryID := uuid.New()
		err := container.Database.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "测试分类", "测试分类描述", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		// 创建测试商品
		itemID := uuid.New()
		err = container.Database.Exec("INSERT INTO items (id, name, description, type, price, category_id, status, sku) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			itemID.String(), "测试商品", "这是一个测试商品", "PRODUCT", 99.99, categoryID.String(), "ACTIVE", "TEST-002").Error
		require.NoError(t, err, "Failed to create item")

		// 创建HTTP请求
		req, err := http.NewRequest("GET", "/api/v1/items/"+itemID.String(), nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "SUCCESS", response["code"])
		data := response["data"].(map[string]interface{})
		assert.Equal(t, itemID.String(), data["id"])
	})

	t.Run("UpdateItem", func(t *testing.T) {
		// 首先创建一个分类和商品
		categoryID := uuid.New()
		err := container.Database.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "更新测试分类", "用于更新测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		itemID := uuid.New()
		err = container.Database.Exec("INSERT INTO items (id, name, description, type, price, category_id, status, sku) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			itemID.String(), "原始商品", "原始描述", "PRODUCT", 99.99, categoryID.String(), "ACTIVE", "TEST-003").Error
		require.NoError(t, err, "Failed to create item")

		// 准备更新数据
		requestBody := map[string]interface{}{
			"name":        "更新后的商品名称",
			"description": "更新后的商品描述",
			"price":       149.99,
		}

		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)

		// 创建HTTP请求
		req, err := http.NewRequest("PUT", "/api/v1/items/"+itemID.String(), bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "SUCCESS", response["code"])
		assert.Equal(t, "Item updated successfully", response["message"])
	})

	t.Run("UpdateItemStatus", func(t *testing.T) {
		// 首先创建一个分类和商品
		categoryID := uuid.New()
		err := container.Database.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "状态测试分类", "用于状态测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		itemID := uuid.New()
		err = container.Database.Exec("INSERT INTO items (id, name, description, type, price, category_id, status, sku) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			itemID.String(), "状态测试商品", "用于状态测试的商品", "PRODUCT", 99.99, categoryID.String(), "DRAFT", "TEST-004").Error
		require.NoError(t, err, "Failed to create item")

		// 准备状态更新数据
		requestBody := map[string]interface{}{
			"status": "ACTIVE",
		}

		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)

		// 创建HTTP请求
		req, err := http.NewRequest("PATCH", "/api/v1/items/"+itemID.String()+"/status", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "SUCCESS", response["code"])
		assert.Equal(t, "Item status updated successfully", response["message"])
	})

	t.Run("DeleteItem", func(t *testing.T) {
		// 首先创建一个分类和商品
		categoryID := uuid.New()
		err := container.Database.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "删除测试分类", "用于删除测试的分类", "ACTIVE").Error
		require.NoError(t, err, "Failed to create category")

		itemID := uuid.New()
		err = container.Database.Exec("INSERT INTO items (id, name, description, type, price, category_id, status, sku) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			itemID.String(), "待删除商品", "用于删除测试的商品", "PRODUCT", 99.99, categoryID.String(), "ACTIVE", "TEST-005").Error
		require.NoError(t, err, "Failed to create item")

		// 创建HTTP请求
		req, err := http.NewRequest("DELETE", "/api/v1/items/"+itemID.String(), nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "SUCCESS", response["code"])
		assert.Equal(t, "Item deleted successfully", response["message"])
	})
}

// TestAPIErrorHandling 测试API错误处理
func TestAPIErrorHandling(t *testing.T) {
	router, container, _ := setupTestServer(t)
	defer container.Cleanup(context.Background())

	t.Run("InvalidJSON", func(t *testing.T) {
		// 发送无效的JSON数据
		invalidJSON := []byte(`{"name": "test", "price": invalid_number}`)

		req, err := http.NewRequest("POST", "/api/v1/items", bytes.NewBuffer(invalidJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 应该返回400错误
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/items/invalid-uuid", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 应该返回400或404错误
		assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound)
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		// 发送缺少必填字段的数据
		requestBody := map[string]interface{}{
			"description": "缺少name和price字段",
		}

		jsonData, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/items", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 应该返回400错误
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidPagination", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/v1/items?page=invalid&page_size=0", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 应该返回400错误或使用默认值
		// 具体响应码取决于实现方式
		assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK)
	})
}

// TestAPIResponseFormat 测试API响应格式
func TestAPIResponseFormat(t *testing.T) {
	router, container, _ := setupTestServer(t)
	defer container.Cleanup(context.Background())

	t.Run("ConsistentResponseFormat", func(t *testing.T) {
		// 测试健康检查端点
		req, err := http.NewRequest("GET", "/health", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// 验证响应包含必要字段
		assert.Contains(t, response, "status")
		assert.Contains(t, response, "timestamp")
		assert.Contains(t, response, "service")
	})

	t.Run("ErrorResponseFormat", func(t *testing.T) {
		// 测试不存在的端点
		req, err := http.NewRequest("GET", "/api/v1/nonexistent", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 404错误
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestAPIPerformance 测试API性能
func TestAPIPerformance(t *testing.T) {
	router, container, _ := setupTestServer(t)
	defer container.Cleanup(context.Background())

	t.Run("ResponseTime", func(t *testing.T) {
		// 创建一些测试数据
		categoryID := uuid.New()
		container.Database.Exec("INSERT INTO categories (id, name, description, status) VALUES (?, ?, ?, ?)",
			categoryID.String(), "性能测试分类", "用于性能测试的分类", "ACTIVE")

		// 测试列表接口的响应时间
		req, err := http.NewRequest("GET", "/api/v1/items", nil)
		require.NoError(t, err)

		start := time.Now()
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		duration := time.Since(start)

		// 验证响应时间应该在合理范围内（比如小于100ms）
		assert.Less(t, duration, 100*time.Millisecond, "API response time should be under 100ms")
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestAPICorsHeaders 测试CORS头
func TestAPICorsHeaders(t *testing.T) {
	router, container, _ := setupTestServer(t)
	defer container.Cleanup(context.Background())

	req, err := http.NewRequest("OPTIONS", "/api/v1/items", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://localhost:3000")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证CORS头
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "*")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "PUT")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "DELETE")
}

// TestAPISwaggerDocumentation 测试Swagger文档端点
func TestAPISwaggerDocumentation(t *testing.T) {
	router, container, _ := setupTestServer(t)
	defer container.Cleanup(context.Background())

	t.Run("SwaggerJSON", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/swagger/doc.json", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var swaggerDoc map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &swaggerDoc)
		require.NoError(t, err)

		assert.Contains(t, swaggerDoc, "swagger")
		assert.Contains(t, swaggerDoc, "info")
		assert.Contains(t, swaggerDoc, "paths")
	})

	t.Run("SwaggerUI", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/swagger/index.html", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Swagger UI通常返回200或重定向
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusMovedPermanently)
	})
}