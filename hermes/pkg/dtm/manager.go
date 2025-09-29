package dtm

import (
	"context"
	"fmt"

	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DTMManager manages distributed transactions using DTM
// DTM分布式事务管理器，提供跨服务的事务一致性保证
type DTMManager struct {
	dtmServer string
	conn      *grpc.ClientConn
}

// NewDTMManager creates a new DTM manager instance
// 创建新的DTM管理器实例
func NewDTMManager(dtmServer string) (*DTMManager, error) {
	conn, err := grpc.NewClient(dtmServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DTM server: %w", err)
	}

	return &DTMManager{
		dtmServer: dtmServer,
		conn:      conn,
	}, nil
}

// CreateSagaTransaction creates a SAGA distributed transaction
// 创建SAGA分布式事务，适用于长流程的业务场景
func (m *DTMManager) CreateSagaTransaction(ctx context.Context, gid string) *dtmgrpc.SagaGrpc {
	return dtmgrpc.NewSagaGrpc(m.dtmServer, gid)
}

// CreateTCCTransaction creates a TCC distributed transaction
// 创建TCC分布式事务，适用于短流程高一致性要求的场景
func (m *DTMManager) CreateTCCTransaction(ctx context.Context, gid string) *dtmgrpc.SagaGrpc {
	return dtmgrpc.NewSagaGrpc(m.dtmServer, gid).EnableWaitResult()
}

// CreateXATransaction creates an XA distributed transaction
// 创建XA分布式事务，适用于数据库事务的强一致性场景
func (m *DTMManager) CreateXATransaction(ctx context.Context, gid string) *dtmgrpc.SagaGrpc {
	return dtmgrpc.NewSagaGrpc(m.dtmServer, gid)
}

// Close closes the DTM connection
// 关闭DTM连接
func (m *DTMManager) Close() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

// BusinessTransaction represents a complete business transaction
// 表示一个完整的业务事务，包含多个服务的操作
type BusinessTransaction struct {
	CustomerID uint
	OrderItems []OrderItem
	PaymentAmount float64
}

type OrderItem struct {
	ProductID uint
	Quantity  int
	Price     float64
}

// CreateCustomerRequest DTM transaction request types
type CreateCustomerRequest struct {
	Name  string
	Phone string
	Email string
	Tags  string
}

type CreateOrderRequest struct {
	CustomerID uint
	OrderItems []OrderItem
}

type ConsumeWalletRequest struct {
	CustomerID uint
	OrderID    uint
	Amount     float64
}

// ProcessOrderTransaction 处理完整的订单业务流程
// 包括：1.创建订单 2.扣减钱包余额 3.创建支付记录
// 使用DTM SAGA模式确保分布式事务一致性
func (m *DTMManager) ProcessOrderTransaction(ctx context.Context, businessTx *BusinessTransaction) error {
	// 生成全局事务ID
	gid := dtmcli.MustGenGid(m.dtmServer)

	// 创建SAGA事务
	saga := m.CreateSagaTransaction(ctx, gid)

	// 添加创建订单步骤
	saga.Add("kratos.OrderService/CreateOrder", "kratos.OrderService/CancelOrder", nil)

	// 添加钱包扣款步骤
	saga.Add("plutus.PaymentService/ConsumeWallet", "plutus.PaymentService/RefundWallet", nil)

	// 提交SAGA事务
	return saga.Submit()
}

// ProcessRefundTransaction 处理退款业务流程
// 包括：1.取消订单 2.退款到钱包
// 使用DTM确保退款操作的一致性
func (m *DTMManager) ProcessRefundTransaction(ctx context.Context, orderID uint, refundAmount float64) error {
	gid := dtmcli.MustGenGid(m.dtmServer)

	saga := m.CreateSagaTransaction(ctx, gid)

	// 添加取消订单步骤
	saga.Add("kratos.OrderService/CancelOrder", "kratos.OrderService/RestoreOrder", nil)

	// 添加退款步骤
	saga.Add("plutus.PaymentService/RefundWallet", "plutus.PaymentService/CancelRefund", nil)

	return saga.Submit()
}