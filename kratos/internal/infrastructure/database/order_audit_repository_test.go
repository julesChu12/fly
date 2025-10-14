package database

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderAuditRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderAuditRepository(db)
	ctx := context.Background()

	// Create order first
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-AUDIT-001",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create audit record
	payload := map[string]interface{}{
		"field":     "total_amount",
		"old_value": 100.00,
		"new_value": 150.00,
	}
	payloadJSON, _ := json.Marshal(payload)

	audit := &entity.OrderAudit{
		TenantID: 1,
		OrderID:  order.ID,
		Action:   "update",
		Actor:    "admin@example.com",
		Payload:  string(payloadJSON),
	}

	err := repo.Create(ctx, audit)
	assert.NoError(t, err)
	assert.NotZero(t, audit.ID)
	assert.NotZero(t, audit.CreatedAt)
}

func TestOrderAuditRepository_GetByOrderID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderAuditRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-AUDIT-002",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create multiple audit records
	audits := []*entity.OrderAudit{
		{
			TenantID: 1,
			OrderID:  order.ID,
			Action:   "create",
			Actor:    "system",
			Payload:  `{"order_no":"ORD-AUDIT-002"}`,
		},
		{
			TenantID: 1,
			OrderID:  order.ID,
			Action:   "update_status",
			Actor:    "admin@example.com",
			Payload:  `{"from":"pending","to":"paid"}`,
		},
		{
			TenantID: 1,
			OrderID:  order.ID,
			Action:   "fulfill",
			Actor:    "system",
			Payload:  `{"tracking_no":"TRK123456"}`,
		},
	}

	for _, audit := range audits {
		require.NoError(t, repo.Create(ctx, audit))
	}

	// Get audits by order ID
	foundAudits, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Len(t, foundAudits, 3)

	// Verify order (should be ASC by created_at)
	assert.Equal(t, "create", foundAudits[0].Action)
	assert.Equal(t, "update_status", foundAudits[1].Action)
	assert.Equal(t, "fulfill", foundAudits[2].Action)
}

func TestOrderAuditRepository_GetByOrderID_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderAuditRepository(db)
	ctx := context.Background()

	// Create order without audit records
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-AUDIT-003",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Get audits (should be empty)
	audits, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Empty(t, audits)
}

func TestOrderAuditRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderAuditRepository(db)
	ctx := context.Background()

	// Create multiple orders and audits
	for i := 1; i <= 5; i++ {
		order := &entity.Order{
			TenantID:    1,
			OrderNo:     "ORD-AUDIT-LIST-" + string(rune(i)),
			CustomerID:  1,
			TotalAmount: 100.00,
			Status:      entity.OrderStatusPending,
		}
		require.NoError(t, db.Create(order).Error)

		audit := &entity.OrderAudit{
			TenantID: 1,
			OrderID:  order.ID,
			Action:   "create",
			Actor:    "system",
			Payload:  `{}`,
		}
		require.NoError(t, repo.Create(ctx, audit))
	}

	// Test pagination - first page
	audits, err := repo.List(ctx, 0, 3)
	assert.NoError(t, err)
	assert.Len(t, audits, 3)

	// Test pagination - second page
	audits, err = repo.List(ctx, 3, 3)
	assert.NoError(t, err)
	assert.Len(t, audits, 2)
}

func TestOrderAuditRepository_DifferentActions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderAuditRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-AUDIT-004",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create audits for different actions
	actions := []struct {
		action  string
		actor   string
		payload string
	}{
		{"create", "system", `{"order_no":"ORD-AUDIT-004"}`},
		{"update", "admin@example.com", `{"field":"total_amount","old":100,"new":150}`},
		{"status_change", "system", `{"from":"pending","to":"paid"}`},
		{"cancel", "customer", `{"reason":"Change of mind"}`},
	}

	for _, a := range actions {
		audit := &entity.OrderAudit{
			TenantID: 1,
			OrderID:  order.ID,
			Action:   a.action,
			Actor:    a.actor,
			Payload:  a.payload,
		}
		require.NoError(t, repo.Create(ctx, audit))
	}

	// Get all audits
	audits, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Len(t, audits, 4)

	// Verify different actions
	assert.Equal(t, "create", audits[0].Action)
	assert.Equal(t, "update", audits[1].Action)
	assert.Equal(t, "status_change", audits[2].Action)
	assert.Equal(t, "cancel", audits[3].Action)
}

func TestOrderAuditRepository_WithJSONPayload(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderAuditRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-AUDIT-005",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create audit with complex JSON payload
	payload := map[string]interface{}{
		"items": []map[string]interface{}{
			{"product_id": 1, "quantity": 2, "price": 50.00},
			{"product_id": 2, "quantity": 1, "price": 100.00},
		},
		"total":    200.00,
		"currency": "CNY",
		"metadata": map[string]string{
			"source":  "web",
			"version": "1.0",
		},
	}
	payloadJSON, _ := json.Marshal(payload)

	audit := &entity.OrderAudit{
		TenantID: 1,
		OrderID:  order.ID,
		Action:   "create_with_items",
		Actor:    "api",
		Payload:  string(payloadJSON),
	}
	require.NoError(t, repo.Create(ctx, audit))

	// Retrieve and verify
	audits, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Len(t, audits, 1)

	// Parse and verify JSON payload
	var retrievedPayload map[string]interface{}
	err = json.Unmarshal([]byte(audits[0].Payload), &retrievedPayload)
	assert.NoError(t, err)
	assert.Equal(t, 200.00, retrievedPayload["total"])
	assert.Equal(t, "CNY", retrievedPayload["currency"])
}

func TestOrderAuditRepository_MultipleOrders(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderAuditRepository(db)
	ctx := context.Background()

	// Create two orders
	order1 := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-AUDIT-006",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	order2 := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-AUDIT-007",
		CustomerID:  1,
		TotalAmount: 200.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order1).Error)
	require.NoError(t, db.Create(order2).Error)

	// Create audits for both orders
	audit1 := &entity.OrderAudit{
		TenantID: 1,
		OrderID:  order1.ID,
		Action:   "create",
		Actor:    "system",
		Payload:  `{"order":"1"}`,
	}
	audit2 := &entity.OrderAudit{
		TenantID: 1,
		OrderID:  order2.ID,
		Action:   "create",
		Actor:    "system",
		Payload:  `{"order":"2"}`,
	}
	require.NoError(t, repo.Create(ctx, audit1))
	require.NoError(t, repo.Create(ctx, audit2))

	// Verify audits are isolated per order
	audits1, err := repo.GetByOrderID(ctx, order1.ID)
	assert.NoError(t, err)
	assert.Len(t, audits1, 1)
	assert.Contains(t, audits1[0].Payload, `"order":"1"`)

	audits2, err := repo.GetByOrderID(ctx, order2.ID)
	assert.NoError(t, err)
	assert.Len(t, audits2, 1)
	assert.Contains(t, audits2[0].Payload, `"order":"2"`)
}

func TestOrderAuditRepository_EmptyActor(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderAuditRepository(db)
	ctx := context.Background()

	// Create order
	order := &entity.Order{
		TenantID:    1,
		OrderNo:     "ORD-AUDIT-008",
		CustomerID:  1,
		TotalAmount: 100.00,
		Status:      entity.OrderStatusPending,
	}
	require.NoError(t, db.Create(order).Error)

	// Create audit without actor (system action)
	audit := &entity.OrderAudit{
		TenantID: 1,
		OrderID:  order.ID,
		Action:   "auto_cancel",
		Actor:    "", // Empty actor
		Payload:  `{"reason":"timeout"}`,
	}
	err := repo.Create(ctx, audit)
	assert.NoError(t, err)

	// Verify it was saved
	audits, err := repo.GetByOrderID(ctx, order.ID)
	assert.NoError(t, err)
	assert.Len(t, audits, 1)
	assert.Equal(t, "", audits[0].Actor)
}
