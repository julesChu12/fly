package grpc

import (
	"context"

	pb "github.com/julesChu12/fly/plutus/api/proto"
	"github.com/julesChu12/fly/plutus/internal/application/service"
	"github.com/julesChu12/fly/plutus/internal/domain/entity"
	"github.com/julesChu12/fly/plutus/pkg/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WalletGRPCHandler implements the gRPC WalletService interface
// 钱包服务gRPC处理器，提供钱包和交易管理的所有gRPC接口
type WalletGRPCHandler struct {
	pb.UnimplementedWalletServiceServer
	walletService service.WalletService
}

// NewWalletGRPCHandler creates a new WalletGRPCHandler
// 创建新的钱包gRPC处理器
func NewWalletGRPCHandler(walletService service.WalletService) *WalletGRPCHandler {
	return &WalletGRPCHandler{
		walletService: walletService,
	}
}

// CreateWallet creates a new wallet via gRPC
// 通过gRPC创建新钱包
func (h *WalletGRPCHandler) CreateWallet(ctx context.Context, req *pb.CreateWalletRequest) (*pb.CreateWalletResponse, error) {
	createReq := &types.CreateWalletRequest{
		CustomerID: uint(req.CustomerId),
		Currency:   req.Currency,
	}

	wallet, err := h.walletService.CreateWallet(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create wallet: %v", err)
	}

	return &pb.CreateWalletResponse{
		Wallet: h.toProtoWallet(wallet),
	}, nil
}

// GetWallet retrieves a wallet by ID via gRPC
// 通过gRPC根据ID获取钱包信息
func (h *WalletGRPCHandler) GetWallet(ctx context.Context, req *pb.GetWalletRequest) (*pb.GetWalletResponse, error) {
	wallet, err := h.walletService.GetWallet(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "wallet not found: %v", err)
	}

	return &pb.GetWalletResponse{
		Wallet: h.toProtoWallet(wallet),
	}, nil
}

// GetWalletByCustomerID retrieves a wallet by customer ID via gRPC
// 通过gRPC根据客户ID获取钱包信息
func (h *WalletGRPCHandler) GetWalletByCustomerID(ctx context.Context, req *pb.GetWalletByCustomerIDRequest) (*pb.GetWalletResponse, error) {
	wallet, err := h.walletService.GetWalletByCustomerID(ctx, uint(req.CustomerId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "wallet not found: %v", err)
	}

	return &pb.GetWalletResponse{
		Wallet: h.toProtoWallet(wallet),
	}, nil
}

// ListWallets lists wallets via gRPC
// 通过gRPC获取钱包列表
func (h *WalletGRPCHandler) ListWallets(ctx context.Context, req *pb.ListWalletsRequest) (*pb.ListWalletsResponse, error) {
	listReq := &types.ListWalletsRequest{
		ListRequest: types.ListRequest{
			Page:     int(req.Page),
			PageSize: int(req.PageSize),
		},
	}

	result, err := h.walletService.ListWallets(ctx, listReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list wallets: %v", err)
	}

	wallets := make([]*pb.Wallet, 0)
	if walletList, ok := result.Data.([]types.WalletResponse); ok {
		for _, wallet := range walletList {
			wallets = append(wallets, &pb.Wallet{
				Id:         uint32(wallet.ID),
				TenantId:   uint32(wallet.TenantID),
				CustomerId: uint32(wallet.CustomerID),
				Balance:    wallet.Balance,
				Currency:   wallet.Currency,
				Status:     string(wallet.Status),
				CreatedAt:  timestamppb.New(wallet.CreatedAt),
				UpdatedAt:  timestamppb.New(wallet.UpdatedAt),
			})
		}
	}

	return &pb.ListWalletsResponse{
		Wallets: wallets,
		Total:   result.Total,
		Page:    int32(result.Page),
		Size:    int32(result.Size),
	}, nil
}

// Recharge performs wallet recharge via gRPC
// 通过gRPC执行钱包充值
func (h *WalletGRPCHandler) Recharge(ctx context.Context, req *pb.RechargeRequest) (*pb.TransactionResponse, error) {
	var idempotencyKey *string
	if req.IdempotencyKey != "" {
		idempotencyKey = &req.IdempotencyKey
	}

	var referenceNo *string
	if req.ReferenceNo != "" {
		referenceNo = &req.ReferenceNo
	}

	// Parse meta JSON string to map
	var meta map[string]interface{}
	if req.Meta != "" {
		// For simplicity, just pass empty map if meta is provided
		meta = make(map[string]interface{})
	}

	rechargeReq := &types.RechargeRequest{
		CustomerID:     uint(req.CustomerId),
		Amount:         req.Amount,
		Currency:       req.Currency,
		Channel:        entity.PaymentChannel(req.Channel),
		IdempotencyKey: idempotencyKey,
		ReferenceNo:    referenceNo,
		Meta:           meta,
	}

	transaction, err := h.walletService.Recharge(ctx, rechargeReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to recharge: %v", err)
	}

	return &pb.TransactionResponse{
		Transaction: h.toProtoTransaction(transaction),
	}, nil
}

// Consume performs wallet consume via gRPC
// 通过gRPC执行钱包消费
func (h *WalletGRPCHandler) Consume(ctx context.Context, req *pb.ConsumeRequest) (*pb.TransactionResponse, error) {
	var idempotencyKey *string
	if req.IdempotencyKey != "" {
		idempotencyKey = &req.IdempotencyKey
	}

	var orderID *uint
	if req.OrderId != 0 {
		id := uint(req.OrderId)
		orderID = &id
	}

	// Parse meta JSON string to map
	var meta map[string]interface{}
	if req.Meta != "" {
		// For simplicity, just pass empty map if meta is provided
		meta = make(map[string]interface{})
	}

	consumeReq := &types.ConsumeRequest{
		CustomerID:     uint(req.CustomerId),
		OrderID:        orderID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		IdempotencyKey: idempotencyKey,
		Meta:           meta,
	}

	transaction, err := h.walletService.Consume(ctx, consumeReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to consume: %v", err)
	}

	return &pb.TransactionResponse{
		Transaction: h.toProtoTransaction(transaction),
	}, nil
}

// Refund performs wallet refund via gRPC
// 通过gRPC执行退款
func (h *WalletGRPCHandler) Refund(ctx context.Context, req *pb.RefundRequest) (*pb.TransactionResponse, error) {
	var idempotencyKey *string
	if req.IdempotencyKey != "" {
		idempotencyKey = &req.IdempotencyKey
	}

	var referenceNo *string
	if req.ReferenceNo != "" {
		referenceNo = &req.ReferenceNo
	}

	var orderID *uint
	if req.OrderId != 0 {
		id := uint(req.OrderId)
		orderID = &id
	}

	// Parse meta JSON string to map
	var meta map[string]interface{}
	if req.Meta != "" {
		// For simplicity, just pass empty map if meta is provided
		meta = make(map[string]interface{})
	}

	refundReq := &types.RefundRequest{
		CustomerID:     uint(req.CustomerId),
		OrderID:        orderID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		IdempotencyKey: idempotencyKey,
		ReferenceNo:    referenceNo,
		Meta:           meta,
	}

	transaction, err := h.walletService.Refund(ctx, refundReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to refund: %v", err)
	}

	return &pb.TransactionResponse{
		Transaction: h.toProtoTransaction(transaction),
	}, nil
}

// GetTransaction retrieves a transaction by ID via gRPC
// 通过gRPC根据ID获取交易信息
func (h *WalletGRPCHandler) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.TransactionResponse, error) {
	transaction, err := h.walletService.GetTransaction(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "transaction not found: %v", err)
	}

	return &pb.TransactionResponse{
		Transaction: h.toProtoTransaction(transaction),
	}, nil
}

// ListTransactions lists transactions via gRPC
// 通过gRPC获取交易列表
func (h *WalletGRPCHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	var walletID *uint
	if req.WalletId != 0 {
		id := uint(req.WalletId)
		walletID = &id
	}

	listReq := &types.ListTransactionsRequest{
		ListRequest: types.ListRequest{
			Page:     int(req.Page),
			PageSize: int(req.PageSize),
		},
		WalletID: walletID,
	}

	result, err := h.walletService.ListTransactions(ctx, listReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list transactions: %v", err)
	}

	transactions := make([]*pb.Transaction, 0)
	if transactionList, ok := result.Data.([]types.TransactionResponse); ok {
		for _, transaction := range transactionList {
			transactions = append(transactions, h.toProtoTransaction(&transaction))
		}
	}

	return &pb.ListTransactionsResponse{
		Transactions: transactions,
		Total:        result.Total,
		Page:         int32(result.Page),
		Size:         int32(result.Size),
	}, nil
}

// toProtoWallet converts types.WalletResponse to proto Wallet
// 将内部类型转换为protobuf类型
func (h *WalletGRPCHandler) toProtoWallet(wallet *types.WalletResponse) *pb.Wallet {
	if wallet == nil {
		return nil
	}

	return &pb.Wallet{
		Id:         uint32(wallet.ID),
		TenantId:   uint32(wallet.TenantID),
		CustomerId: uint32(wallet.CustomerID),
		Balance:    wallet.Balance,
		Currency:   wallet.Currency,
		Status:     string(wallet.Status),
		CreatedAt:  timestamppb.New(wallet.CreatedAt),
		UpdatedAt:  timestamppb.New(wallet.UpdatedAt),
	}
}

// toProtoTransaction converts types.TransactionResponse to proto Transaction
// 将内部类型转换为protobuf类型
func (h *WalletGRPCHandler) toProtoTransaction(transaction *types.TransactionResponse) *pb.Transaction {
	if transaction == nil {
		return nil
	}

	var idempotencyKey string
	if transaction.IdempotencyKey != nil {
		idempotencyKey = *transaction.IdempotencyKey
	}

	var referenceNo string
	if transaction.ReferenceNo != nil {
		referenceNo = *transaction.ReferenceNo
	}

	var orderID uint32
	if transaction.OrderID != nil {
		orderID = uint32(*transaction.OrderID)
	}

	return &pb.Transaction{
		Id:             uint32(transaction.ID),
		TenantId:       uint32(transaction.TenantID),
		WalletId:       uint32(transaction.WalletID),
		OrderId:        orderID,
		Type:           string(transaction.Type),
		Amount:         transaction.Amount,
		Currency:       transaction.Currency,
		Channel:        string(transaction.Channel),
		Status:         string(transaction.Status),
		IdempotencyKey: idempotencyKey,
		ReferenceNo:    referenceNo,
		Meta:           transaction.Meta,
		CreatedAt:      timestamppb.New(transaction.CreatedAt),
		UpdatedAt:      timestamppb.New(transaction.UpdatedAt),
	}
}
