package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/kratos/internal/application/service"
	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/database"
	"github.com/julesChu12/fly/kratos/pkg/constants"
	"github.com/julesChu12/fly/kratos/pkg/types"
	"github.com/julesChu12/fly/mora/pkg/auth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHTTPIntegration_OrderLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newTestDatabase(t)
	orderRepo := database.NewOrderRepository(db)
	itemRepo := database.NewOrderItemRepository(db)
	statusRepo := database.NewOrderStatusLogRepository(db)
	auditRepo := database.NewOrderAuditRepository(db)

	orderService := service.NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	// No authentication in integration tests
	router := NewRouter(orderService, RouterConfig{
		CustosClient:     nil, // Skip auth for tests
		SkipAuthPaths:    []string{},
		RateLimitEnabled: false,
	})
	engine := router.SetupRoutes()

	// Generate a test token (will be ignored since CustosClient is nil)
	secret := "integration-secret"
	token, err := auth.GenerateToken("1", "tester", secret, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	orderPayload := types.CreateOrderRequest{
		OrderNo:     "INT-001",
		CustomerID:  99,
		TotalAmount: 42,
		Currency:    "CNY",
		Items: []types.CreateOrderItemRequest{
			{
				ProductName: "Integration Widget",
				Quantity:    1,
				UnitPrice:   42,
			},
		},
	}

	body, _ := json.Marshal(orderPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set(constants.HeaderTenantID, "1")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var createResp types.Response
	if err := json.Unmarshal(w.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := createResp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload, got %T", createResp.Data)
	}

	orderID := int(data["id"].(float64))

	getReq := httptest.NewRequest(http.MethodGet, "/api/orders/"+intToString(orderID)+"/logs", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getReq.Header.Set(constants.HeaderTenantID, "1")

	logsRecorder := httptest.NewRecorder()
	engine.ServeHTTP(logsRecorder, getReq)

	if logsRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for logs, got %d", logsRecorder.Code)
	}

	var logsResp struct {
		Code int                            `json:"code"`
		Data []types.OrderStatusLogResponse `json:"data"`
	}
	if err := json.Unmarshal(logsRecorder.Body.Bytes(), &logsResp); err != nil {
		t.Fatalf("failed to parse logs response: %v", err)
	}

	if len(logsResp.Data) == 0 {
		t.Fatalf("expected logs to be present")
	}
}

func newTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		// Disable foreign keys for SQLite testing
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	// SQLite doesn't support ENUM or GENERATED ALWAYS AS, so we need to modify the schema
	// Use text columns instead of enum for SQLite
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER,
			order_no TEXT UNIQUE NOT NULL,
			customer_id INTEGER NOT NULL,
			total_amount DECIMAL(12,2) NOT NULL,
			currency TEXT DEFAULT 'CNY',
			status TEXT DEFAULT 'pending',
			remark TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);

		CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			order_id INTEGER NOT NULL,
			product_id INTEGER,
			product_name TEXT NOT NULL,
			sku TEXT,
			quantity INTEGER DEFAULT 1 NOT NULL,
			unit_price DECIMAL(12,2) NOT NULL,
			total_price DECIMAL(12,2) NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		);

		CREATE TABLE IF NOT EXISTS order_status_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			order_id INTEGER NOT NULL,
			from_status TEXT,
			to_status TEXT NOT NULL,
			reason TEXT,
			operator_id INTEGER,
			created_at DATETIME
		);
	`).Error; err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	if err := db.AutoMigrate(&entity.OrderAudit{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return db
}

func intToString(v int) string {
	return strconv.Itoa(v)
}
