package database

import (
	"context"
	"testing"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderItemRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderItemRepository(db)
	ctx := context.Background()

	// First create an order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-ITEM-001",
		CustomerID:  1,
		TotalAmount: 299.99,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create order item
	item := &entity.OrderItem{
		TenantID:    1,
		OrderID:     order.ID,
		ProductName: "Test Product",
		SKU:         "SKU-001",
		Quantity:    2,
		UnitPrice:   149.99,
		TotalPrice:  299.98,
	}

	err := repo.Create(ctx, item)
	assert.NoError(t, err)
	assert.NotZero(t, item.ID)
	assert.NotZero(t, item.CreatedAt)
}

func TestOrderItemRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderItemRepository(db)
	ctx := context.Background()

	// Create order and item
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-ITEM-002",
		CustomerID:  1,
		TotalAmount: 99.99,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	item := &entity.OrderItem{
		TenantID:    1,
		OrderID:     order.ID,
		ProductName: "Product A",
		Quantity:    1,
		UnitPrice:   99.99,
		TotalPrice:  99.99,
	}
	require.NoError(t, repo.Create(ctx, item))

	// Test get
	found, err := repo.GetByID(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, item.ID, found.ID)
	assert.Equal(t, item.ProductName, found.ProductName)
	assert.Equal(t, item.Quantity, found.Quantity)
}

func TestOrderItemRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderItemRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 99999)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrOrderItemNotFound, err)
}

func TestOrderItemRepository_GetByOrderID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderItemRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-ITEM-003",
		CustomerID:  1,
		TotalAmount: 399.98,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create multiple items
	items := []*entity.OrderItem{
		{TenantID: 1, OrderID: order.ID, ProductName: "Product A", Quantity: 1, UnitPrice: 100.00, TotalPrice: 100.00},
		{TenantID: 1, OrderID: order.ID, ProductName: "Product B", Quantity: 2, UnitPrice: 149.99, TotalPrice: 299.98},
	}
	for _, item := range items {
		require.NoError(t, repo.Create(ctx, item))
	}

	// Get items by order ID
	foundItems, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Len(t, foundItems, 2)
}

func TestOrderItemRepository_GetByOrderID_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderItemRepository(db)
	ctx := context.Background()

	// Create order without items
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-ITEM-004",
		CustomerID:  1,
		TotalAmount: 0,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Get items (should be empty)
	items, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Empty(t, items)
}

func TestOrderItemRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderItemRepository(db)
	ctx := context.Background()

	// Create order and item
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-ITEM-005",
		CustomerID:  1,
		TotalAmount: 99.99,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	item := &entity.OrderItem{
		TenantID:    1,
		OrderID:     order.ID,
		ProductName: "Original Product",
		Quantity:    1,
		UnitPrice:   99.99,
		TotalPrice:  99.99,
	}
	require.NoError(t, repo.Create(ctx, item))

	// Update item
	item.ProductName = "Updated Product"
	item.Quantity = 3
	item.TotalPrice = 299.97
	err := repo.Update(ctx, item)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(ctx, item.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Product", found.ProductName)
	assert.Equal(t, 3, found.Quantity)
	assert.Equal(t, 299.97, found.TotalPrice)
}

func TestOrderItemRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderItemRepository(db)
	ctx := context.Background()

	// Create order and item
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-ITEM-006",
		CustomerID:  1,
		TotalAmount: 99.99,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	item := &entity.OrderItem{
		TenantID:    1,
		OrderID:     order.ID,
		ProductName: "To Delete",
		Quantity:    1,
		UnitPrice:   99.99,
		TotalPrice:  99.99,
	}
	require.NoError(t, repo.Create(ctx, item))

	// Delete item
	err := repo.Delete(ctx, item.ID)
	assert.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(ctx, item.ID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrOrderItemNotFound, err)
}

func TestOrderItemRepository_DeleteByOrderID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderItemRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-ITEM-007",
		CustomerID:  1,
		TotalAmount: 399.98,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create multiple items
	items := []*entity.OrderItem{
		{TenantID: 1, OrderID: order.ID, ProductName: "Product A", Quantity: 1, UnitPrice: 100.00, TotalPrice: 100.00},
		{TenantID: 1, OrderID: order.ID, ProductName: "Product B", Quantity: 2, UnitPrice: 149.99, TotalPrice: 299.98},
	}
	for _, item := range items {
		require.NoError(t, repo.Create(ctx, item))
	}

	// Delete all items for this order
	err := repo.DeleteByOrderID(ctx, order.ID)
	assert.NoError(t, err)

	// Verify all deleted
	foundItems, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Empty(t, foundItems)
}

func TestOrderItemRepository_MultipleOrders(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderItemRepository(db)
	ctx := context.Background()

	// Create two orders
	order1 := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-ITEM-008",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	order2 := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-ITEM-009",
		CustomerID:  1,
		TotalAmount: 200.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order1).Error)
	require.NoError(t, db.Create(order2).Error)

	// Create items for both orders
	item1 := &entity.OrderItem{
		TenantID:    1,
		OrderID:     order1.ID,
		ProductName: "Order 1 Item",
		Quantity:    1,
		UnitPrice:   100.00,
		TotalPrice:  100.00,
	}
	item2 := &entity.OrderItem{
		TenantID:    1,
		OrderID:     order2.ID,
		ProductName: "Order 2 Item",
		Quantity:    2,
		UnitPrice:   100.00,
		TotalPrice:  200.00,
	}
	require.NoError(t, repo.Create(ctx, item1))
	require.NoError(t, repo.Create(ctx, item2))

	// Verify items are isolated per order
	items1, err := repo.GetByOrderID(ctx, order1.ID)
	assert.NoError(t, err)
	assert.Len(t, items1, 1)
	assert.Equal(t, "Order 1 Item", items1[0].ProductName)

	items2, err := repo.GetByOrderID(ctx, order2.ID)
	assert.NoError(t, err)
	assert.Len(t, items2, 1)
	assert.Equal(t, "Order 2 Item", items2[0].ProductName)
}
