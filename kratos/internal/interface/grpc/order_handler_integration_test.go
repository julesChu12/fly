package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/julesChu12/fly/kratos/api/proto/order/v1"
	"github.com/julesChu12/fly/kratos/internal/application/service"
	"github.com/julesChu12/fly/kratos/internal/domain/entity"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/database"
	"github.com/julesChu12/fly/kratos/pkg/constants"
	"github.com/julesChu12/fly/kratos/pkg/types"
	"github.com/julesChu12/fly/mora/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGRPCIntegration_GetOrderLogs(t *testing.T) {
	db := newIntegrationDB(t)

	orderRepo := database.NewOrderRepository(db)
	itemRepo := database.NewOrderItemRepository(db)
	statusRepo := database.NewOrderStatusLogRepository(db)
	auditRepo := database.NewOrderAuditRepository(db)

	orderService := service.NewOrderService(orderRepo, itemRepo, statusRepo, auditRepo)

	ctx := context.WithValue(context.Background(), constants.ContextKeyTenantID, uint(7))
	ctx = context.WithValue(ctx, constants.ContextKeyUserID, uint(1))

	order, err := orderService.CreateOrder(ctx, &types.CreateOrderRequest{
		OrderNo:     "GRPC-001",
		CustomerID:  11,
		TotalAmount: 100,
		Currency:    "USD",
		Items: []types.CreateOrderItemRequest{
			{
				ProductName: "GRPC Item",
				Quantity:    1,
				UnitPrice:   100,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	secret := "grpc-integration-secret"
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			UnaryServerInterceptor(),
			RateLimitInterceptor(nil),
			AuthInterceptor(secret),
			ContextInjectorInterceptor(),
		),
	)

	orderv1.RegisterOrderServiceServer(server, NewOrderServiceServer(orderService))

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("grpc server exited: %v", err)
		}
	}()
	t.Cleanup(func() {
		server.Stop()
		lis.Close()
	})

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	client := orderv1.NewOrderServiceClient(conn)

	token, err := auth.GenerateToken("1", "tester", secret, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
		"tenant_id":     "7",
	})

	resp, err := client.GetOrderLogs(metadata.NewOutgoingContext(context.Background(), md), &orderv1.GetOrderLogsRequest{
		OrderId: uint64(order.ID),
	})
	if err != nil {
		t.Fatalf("gRPC GetOrderLogs failed: %v", err)
	}

	if len(resp.Logs) == 0 {
		t.Fatalf("expected at least one status log")
	}
}

func newIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	// SQLite doesn't support ENUM or GENERATED ALWAYS AS columns
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
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}
