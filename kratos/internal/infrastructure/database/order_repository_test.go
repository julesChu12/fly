package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Enable foreign key constraints
	db.Exec("PRAGMA foreign_keys = ON")

	// Auto-migrate
	err = db.AutoMigrate(&entity.Order{}, &entity.OrderItem{}, &entity.OrderStatusLog{}, &entity.OrderAudit{})
	require.NoError(t, err)

	return db
}

func TestOrderRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD20231001001",
		CustomerID:  1,
		TotalAmount: 99.99,
		Currency:    "CNY",
		Status:      entity.OrderStatusPending,
		Remark:      "Test order",
	}

	err := repo.Create(ctx, order)
	assert.NoError(t, err)
	assert.NotZero(t, order.ID)
	assert.NotZero(t, order.CreatedAt)
}

func TestOrderRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// Create test data
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD20231001002",
		CustomerID:  1,
		TotalAmount: 199.99,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, repo.Create(ctx, order))

	// Test get
	found, err := repo.GetByID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, found.ID)
	assert.Equal(t, order.OrderNo, found.OrderNo)
	assert.Equal(t, order.TotalAmount, found.TotalAmount)
}

func TestOrderRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 99999)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrOrderNotFound, err)
}

func TestOrderRepository_GetByOrderNo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD20231001003",
		CustomerID:  1,
		TotalAmount: 299.99,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, repo.Create(ctx, order))

	// Test get by order number
	found, err := repo.GetByOrderNo(ctx, "ORD20231001003")
	assert.NoError(t, err)
	assert.Equal(t, order.ID, found.ID)
	assert.Equal(t, "ORD20231001003", found.OrderNo)
}

func TestOrderRepository_GetByOrderNo_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	_, err := repo.GetByOrderNo(ctx, "NONEXISTENT")
	assert.Error(t, err)
	assert.Equal(t, errors.ErrOrderNotFound, err)
}

func TestOrderRepository_GetWithItems(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD20231001004",
		CustomerID:  1,
		TotalAmount: 399.99,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, repo.Create(ctx, order))

	// Add order items
	items := []entity.OrderItem{
		{TenantID: 1, OrderID: order.ID, ProductName: "Product A", Quantity: 2, UnitPrice: 100.00, TotalPrice: 200.00},
		{TenantID: 1, OrderID: order.ID, ProductName: "Product B", Quantity: 1, UnitPrice: 199.99, TotalPrice: 199.99},
	}
	for _, item := range items {
		require.NoError(t, db.Create(&item).Error)
	}

	// Get order with items
	found, err := repo.GetWithItems(ctx, order.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.ID, found.ID)
	assert.Len(t, found.Items, 2)
	assert.Equal(t, "Product A", found.Items[0].ProductName)
	assert.Equal(t, "Product B", found.Items[1].ProductName)
}

func TestOrderRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD20231001005",
		CustomerID:  1,
		TotalAmount: 99.99,
		Status:      entity.OrderStatusPending,
		Remark:      "Original remark",
	}
	require.NoError(t, repo.Create(ctx, order))

	// Update
	order.Remark = "Updated remark"
	order.TotalAmount = 149.99
	err := repo.Update(ctx, order)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated remark", found.Remark)
	assert.Equal(t, 149.99, found.TotalAmount)
}

func TestOrderRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD20231001006",
		CustomerID:  1,
		TotalAmount: 199.99,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, repo.Create(ctx, order))

	// Update status
	operatorID := uint(10)
	err := repo.UpdateStatus(ctx, order.ID, entity.OrderStatusPaid, "Payment received", &operatorID)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.OrderStatusPaid, found.Status)
}

func TestOrderRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD20231001007",
		CustomerID:  1,
		TotalAmount: 99.99,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, repo.Create(ctx, order))

	// Delete
	err := repo.Delete(ctx, order.ID)
	assert.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(ctx, order.ID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrOrderNotFound, err)
}

func TestOrderRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// Create multiple orders
	for i := 1; i <= 15; i++ {
		order := &entity.Order{
			TenantID:    1,
			OrderNo:     fmt.Sprintf("ORD2023100%03d", i),
			CustomerID:  1,
			TotalAmount: float64(i * 100),
			Status:      entity.OrderStatusPending,
		}
		require.NoError(t, repo.Create(ctx, order))
	}

	// Test pagination
	orders, total, err := repo.List(ctx, 1, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, orders, 10)

	// Second page
	orders, total, err = repo.List(ctx, 1, 10, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, orders, 5)
}

func TestOrderRepository_List_MultiTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// Create orders for different tenants
	for i := 1; i <= 5; i++ {
		order1 := &entity.Order{
			TenantID:    1,
			OrderNo:     fmt.Sprintf("T1-ORD%03d", i),
			CustomerID:  1,
			TotalAmount: 100.00,
			Status:      entity.OrderStatusPending,
		}
		order2 := &entity.Order{
			TenantID:    2,
			OrderNo:     fmt.Sprintf("T2-ORD%03d", i),
			CustomerID:  2,
			TotalAmount: 200.00,
			Status:      entity.OrderStatusPending,
		}
		require.NoError(t, repo.Create(ctx, order1))
		require.NoError(t, repo.Create(ctx, order2))
	}

	// Query tenant 1 orders
	orders, total, err := repo.List(ctx, 1, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, orders, 5)

	// Verify all are tenant 1 data
	for _, order := range orders {
		assert.Equal(t, uint(1), order.TenantID)
	}
}

func TestOrderRepository_ListByCustomer(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// Create orders for different customers
	for i := 1; i <= 3; i++ {
		order1 := &entity.Order{
			TenantID:    1,
			OrderNo:     fmt.Sprintf("C1-ORD%03d", i),
			CustomerID:  1,
			TotalAmount: 100.00,
			Status:      entity.OrderStatusPending,
		}
		order2 := &entity.Order{
			TenantID:    1,
			OrderNo:     fmt.Sprintf("C2-ORD%03d", i),
			CustomerID:  2,
			TotalAmount: 200.00,
			Status:      entity.OrderStatusPending,
		}
		require.NoError(t, repo.Create(ctx, order1))
		require.NoError(t, repo.Create(ctx, order2))
	}

	// Query customer 1 orders
	orders, total, err := repo.ListByCustomer(ctx, 1, 1, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, orders, 3)

	// Verify all are customer 1 orders
	for _, order := range orders {
		assert.Equal(t, uint(1), order.CustomerID)
	}
}

func TestOrderRepository_ListByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	// Create orders with different statuses
	statuses := []entity.OrderStatus{
		entity.OrderStatusPending,
		entity.OrderStatusPending,
		entity.OrderStatusPaid,
		entity.OrderStatusPaid,
		entity.OrderStatusFulfilled,
	}

	for i, status := range statuses {
		order := &entity.Order{
			TenantID:    1,
			OrderNo:     fmt.Sprintf("ORD-STATUS-%03d", i),
			CustomerID:  1,
			TotalAmount: 100.00,
			Status:      status,
		}
		require.NoError(t, repo.Create(ctx, order))
	}

	// Query pending orders
	orders, total, err := repo.ListByStatus(ctx, 1, entity.OrderStatusPending, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, orders, 2)

	// Verify all are pending status
	for _, order := range orders {
		assert.Equal(t, entity.OrderStatusPending, order.Status)
	}

	// Query paid orders
	orders, total, err = repo.ListByStatus(ctx, 1, entity.OrderStatusPaid, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, orders, 2)
}

func TestOrderRepository_UniqueOrderNo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	order1 := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-UNIQUE-001",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, repo.Create(ctx, order1))

	// Try to create order with same order number (should fail)
	order2 := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-UNIQUE-001", // Same order number
		CustomerID:  2,
		TotalAmount: 200.00,
		Status:      entity.OrderStatusPending,
	}
	err := repo.Create(ctx, order2)
	assert.Error(t, err) // Should fail due to duplicate order number
}
