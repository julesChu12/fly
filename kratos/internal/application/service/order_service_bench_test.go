package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/pkg/constants"
	"github.com/julesChu12/fly/kratos/pkg/types"
)

func BenchmarkOrderServiceCreateOrder(b *testing.B) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()
	cache := newMockOrderCache()

	svc := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo, WithOrderCache(cache, time.Minute))

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	prototype := &types.CreateOrderRequest{
		CustomerID:  1,
		TotalAmount: 100,
		Currency:    "USD",
		Items: []types.CreateOrderItemRequest{
			{
				ProductName: "Benchmark Item",
				Quantity:    2,
				UnitPrice:   50,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := *prototype
		req.OrderNo = fmt.Sprintf("BM-%06d", i)
		if _, err := svc.CreateOrder(ctx, &req); err != nil {
			b.Fatalf("create order failed: %v", err)
		}
	}

	b.ReportMetric(float64(orderRepo.CallsToGetByID()), "order_repo_get_by_id_calls")
}

func BenchmarkOrderServiceUpdateOrderStatus(b *testing.B) {
	orderRepo := NewMockOrderRepository()
	itemRepo := NewMockOrderItemRepository()
	statusRepo := NewMockOrderStatusLogRepository()
	auditRepo := NewMockOrderAuditRepository()

	svc := NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(1))

	order, err := svc.CreateOrder(ctx, &types.CreateOrderRequest{
		OrderNo:     "BM-STATUS",
		CustomerID:  1,
		TotalAmount: 100,
		Items: []types.CreateOrderItemRequest{{
			ProductName: "Benchmark Item",
			Quantity:    1,
			UnitPrice:   100,
		}},
	})
	if err != nil {
		b.Fatalf("failed to seed order: %v", err)
	}

	req := &types.UpdateOrderStatusRequest{Status: entity.OrderStatusPaid}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.UpdateOrderStatus(ctx, order.ID, req); err != nil {
			b.Fatalf("update status failed: %v", err)
		}
	}
}
