package database

import (
	"context"
	"testing"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderStatusLogRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderStatusLogRepository(db)
	ctx := context.Background()

	// Create order first
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-LOG-001",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create status log
	fromStatus := entity.OrderStatusPending
	operatorID := uint(10)
	log := &entity.OrderStatusLog{
		TenantID:   1,
		OrderID:    order.ID,
		FromStatus: &fromStatus,
		ToStatus:   entity.OrderStatusPaid,
		Reason:     "Payment received",
		OperatorID: &operatorID,
	}

	err := repo.Create(ctx, log)
	assert.NoError(t, err)
	assert.NotZero(t, log.ID)
	assert.NotZero(t, log.CreatedAt)
}

func TestOrderStatusLogRepository_GetByOrderID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderStatusLogRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-LOG-002",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create multiple status logs
	logs := []*entity.OrderStatusLog{
		{
			TenantID:   1,
			OrderID:    order.ID,
			FromStatus: nil,
			ToStatus:   entity.OrderStatusPending,
			Reason:     "Order created",
		},
		{
			TenantID: 1,
			OrderID:  order.ID,
			FromStatus: func() *entity.OrderStatus {
				s := entity.OrderStatusPending
				return &s
			}(),
			ToStatus: entity.OrderStatusPaid,
			Reason:   "Payment completed",
		},
		{
			TenantID: 1,
			OrderID:  order.ID,
			FromStatus: func() *entity.OrderStatus {
				s := entity.OrderStatusPaid
				return &s
			}(),
			ToStatus: entity.OrderStatusFulfilled,
			Reason:   "Order fulfilled",
		},
	}

	for _, log := range logs {
		require.NoError(t, repo.Create(ctx, log))
	}

	// Get logs by order ID
	foundLogs, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Len(t, foundLogs, 3)

	// Verify order (should be ASC by created_at)
	assert.Equal(t, entity.OrderStatusPending, foundLogs[0].ToStatus)
	assert.Equal(t, entity.OrderStatusPaid, foundLogs[1].ToStatus)
	assert.Equal(t, entity.OrderStatusFulfilled, foundLogs[2].ToStatus)
}

func TestOrderStatusLogRepository_GetByOrderID_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderStatusLogRepository(db)
	ctx := context.Background()

	// Create order without status logs
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-LOG-003",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Get logs (should be empty)
	logs, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Empty(t, logs)
}

func TestOrderStatusLogRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderStatusLogRepository(db)
	ctx := context.Background()

	// Create multiple orders and logs
	for i := 1; i <= 5; i++ {
		order := &entity.Order{
			TenantID:    1,
			OrderNo:     "ORD-LOG-LIST-" + string(rune(i)),
			CustomerID:  1,
			TotalAmount: 100.00,
			Status:      entity.OrderStatusPending,
		}
		require.NoError(t, db.Create(order).Error)

		log := &entity.OrderStatusLog{
			TenantID: 1,
			OrderID:  order.ID,
			ToStatus: entity.OrderStatusPending,
			Reason:   "Order created",
		}
		require.NoError(t, repo.Create(ctx, log))
	}

	// Test pagination - first page
	logs, err := repo.List(ctx, 0, 3)
	assert.NoError(t, err)
	assert.Len(t, logs, 3)

	// Test pagination - second page
	logs, err = repo.List(ctx, 3, 3)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
}

func TestOrderStatusLogRepository_StatusTransition(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderStatusLogRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-LOG-004",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create status log with nil FromStatus (initial creation)
	log1 := &entity.OrderStatusLog{
		TenantID:   1,
		OrderID:    order.ID,
		FromStatus: nil,
		ToStatus:   entity.OrderStatusPending,
		Reason:     "Order created",
	}
	require.NoError(t, repo.Create(ctx, log1))

	// Create status log with transition
	fromStatus := entity.OrderStatusPending
	log2 := &entity.OrderStatusLog{
		TenantID:   1,
		OrderID:    order.ID,
		FromStatus: &fromStatus,
		ToStatus:   entity.OrderStatusPaid,
		Reason:     "Payment received",
	}
	require.NoError(t, repo.Create(ctx, log2))

	// Verify both logs
	logs, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)

	// First log should have nil FromStatus
	assert.Nil(t, logs[0].FromStatus)
	assert.Equal(t, entity.OrderStatusPending, logs[0].ToStatus)

	// Second log should have FromStatus
	assert.NotNil(t, logs[1].FromStatus)
	assert.Equal(t, entity.OrderStatusPending, *logs[1].FromStatus)
	assert.Equal(t, entity.OrderStatusPaid, logs[1].ToStatus)
}

func TestOrderStatusLogRepository_WithOperator(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderStatusLogRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-LOG-005",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create status log with operator
	operatorID := uint(100)
	log := &entity.OrderStatusLog{
		TenantID:   1,
		OrderID:    order.ID,
		ToStatus:   entity.OrderStatusCanceled,
		Reason:     "Canceled by admin",
		OperatorID: &operatorID,
	}
	require.NoError(t, repo.Create(ctx, log))

	// Verify operator ID is saved
	logs, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.NotNil(t, logs[0].OperatorID)
	assert.Equal(t, uint(100), *logs[0].OperatorID)
}

func TestOrderStatusLogRepository_MultipleOrders(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderStatusLogRepository(db)
	ctx := context.Background()

	// Create two orders
	order1 := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-LOG-006",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	order2 := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-LOG-007",
		CustomerID:  1,
		TotalAmount: 200.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order1).Error)
	require.NoError(t, db.Create(order2).Error)

	// Create logs for both orders
	log1 := &entity.OrderStatusLog{
		TenantID: 1,
		OrderID:  order1.ID,
		ToStatus: entity.OrderStatusPending,
		Reason:   "Order 1 created",
	}
	log2 := &entity.OrderStatusLog{
		TenantID: 1,
		OrderID:  order2.ID,
		ToStatus: entity.OrderStatusPending,
		Reason:   "Order 2 created",
	}
	require.NoError(t, repo.Create(ctx, log1))
	require.NoError(t, repo.Create(ctx, log2))

	// Verify logs are isolated per order
	logs1, err := repo.GetByOrderID(ctx, order1.ID)
	assert.NoError(t, err)
	assert.Len(t, logs1, 1)
	assert.Equal(t, "Order 1 created", logs1[0].Reason)

	logs2, err := repo.GetByOrderID(ctx, order2.ID)
	assert.NoError(t, err)
	assert.Len(t, logs2, 1)
	assert.Equal(t, "Order 2 created", logs2[0].Reason)
}
