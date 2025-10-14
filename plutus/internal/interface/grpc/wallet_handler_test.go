package grpc

import (
	"context"
	"testing"
	"time"

	pb "github.com/julesChu12/fly/plutus/api/proto"
	"github.com/julesChu12/fly/plutus/internal/domain/entity"
	"github.com/julesChu12/fly/plutus/pkg/constants"
	"github.com/julesChu12/fly/plutus/pkg/errors"
	"github.com/julesChu12/fly/plutus/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWalletService is a mock implementation of WalletService for gRPC tests
type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) CreateWallet(ctx context.Context, req *types.CreateWalletRequest) (*types.WalletResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWallet(ctx context.Context, id uint) (*types.WalletResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWalletByCustomerID(ctx context.Context, customerID uint) (*types.WalletResponse, error) {
	args := m.Called(ctx, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.WalletResponse), args.Error(1)
}

func (m *MockWalletService) ListWallets(ctx context.Context, req *types.ListWalletsRequest) (*types.ListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ListResponse), args.Error(1)
}

func (m *MockWalletService) Recharge(ctx context.Context, req *types.RechargeRequest) (*types.TransactionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TransactionResponse), args.Error(1)
}

func (m *MockWalletService) Consume(ctx context.Context, req *types.ConsumeRequest) (*types.TransactionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TransactionResponse), args.Error(1)
}

func (m *MockWalletService) Refund(ctx context.Context, req *types.RefundRequest) (*types.TransactionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TransactionResponse), args.Error(1)
}

func (m *MockWalletService) GetTransaction(ctx context.Context, id uint) (*types.TransactionResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TransactionResponse), args.Error(1)
}

func (m *MockWalletService) ListTransactions(ctx context.Context, req *types.ListTransactionsRequest) (*types.ListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ListResponse), args.Error(1)
}

func getTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, constants.ContextKeyTenantID, uint(1))
	ctx = context.WithValue(ctx, constants.ContextKeyUserID, uint(1))
	return ctx
}

func TestGRPCCreateWallet(t *testing.T) {
	ctx := getTestContext()

	t.Run("success", func(t *testing.T) {
		mockService := new(MockWalletService)
		handler := NewWalletGRPCHandler(mockService)

		req := &pb.CreateWalletRequest{
			CustomerId: 1,
			Currency:   "CNY",
		}

		expected := &types.WalletResponse{
			ID:         1,
			TenantID:   1,
			CustomerID: 1,
			Balance:    0,
			Currency:   "CNY",
			Status:     entity.WalletStatusActive,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockService.On("CreateWallet", mock.Anything, mock.MatchedBy(func(r *types.CreateWalletRequest) bool {
			return r.CustomerID == 1 && r.Currency == "CNY"
		})).Return(expected, nil)

		resp, err := handler.CreateWallet(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint32(1), resp.Wallet.Id)
		assert.Equal(t, "CNY", resp.Wallet.Currency)
		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(MockWalletService)
		handler := NewWalletGRPCHandler(mockService)

		req := &pb.CreateWalletRequest{
			CustomerId: 1,
			Currency:   "CNY",
		}

		mockService.On("CreateWallet", mock.Anything, mock.Anything).Return(nil, errors.ErrWalletAlreadyExists)

		resp, err := handler.CreateWallet(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}

func TestGRPCGetWallet(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletGRPCHandler(mockService)
	ctx := getTestContext()

	t.Run("success", func(t *testing.T) {
		expected := &types.WalletResponse{
			ID:         1,
			TenantID:   1,
			CustomerID: 1,
			Balance:    100.0,
			Currency:   "CNY",
			Status:     entity.WalletStatusActive,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockService.On("GetWallet", mock.Anything, uint(1)).Return(expected, nil)

		resp, err := handler.GetWallet(ctx, &pb.GetWalletRequest{Id: 1})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint32(1), resp.Wallet.Id)
		assert.Equal(t, 100.0, resp.Wallet.Balance)
		mockService.AssertExpectations(t)
	})
}

func TestGRPCGetWalletByCustomerID(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletGRPCHandler(mockService)
	ctx := getTestContext()

	t.Run("success", func(t *testing.T) {
		expected := &types.WalletResponse{
			ID:         1,
			TenantID:   1,
			CustomerID: 1,
			Balance:    100.0,
			Currency:   "CNY",
			Status:     entity.WalletStatusActive,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockService.On("GetWalletByCustomerID", mock.Anything, uint(1)).Return(expected, nil)

		resp, err := handler.GetWalletByCustomerID(ctx, &pb.GetWalletByCustomerIDRequest{CustomerId: 1})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint32(1), resp.Wallet.CustomerId)
		mockService.AssertExpectations(t)
	})
}

func TestGRPCListWallets(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletGRPCHandler(mockService)
	ctx := getTestContext()

	t.Run("success", func(t *testing.T) {
		expected := &types.ListResponse{
			Data: []types.WalletResponse{
				{
					ID:         1,
					TenantID:   1,
					CustomerID: 1,
					Balance:    100.0,
					Currency:   "CNY",
					Status:     entity.WalletStatusActive,
				},
			},
			Total: 1,
			Page:  1,
			Size:  1,
		}

		mockService.On("ListWallets", mock.Anything, mock.Anything).Return(expected, nil)

		resp, err := handler.ListWallets(ctx, &pb.ListWalletsRequest{Page: 1, PageSize: 20})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Total)
		assert.Len(t, resp.Wallets, 1)
		mockService.AssertExpectations(t)
	})
}

func TestGRPCRecharge(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletGRPCHandler(mockService)
	ctx := getTestContext()

	t.Run("success", func(t *testing.T) {
		expected := &types.TransactionResponse{
			ID:         1,
			TenantID:   1,
			WalletID:   1,
			Type:       entity.TransactionTypeRecharge,
			Amount:     100.0,
			Currency:   "CNY",
			Channel:    entity.ChannelAlipay,
			Status:     entity.TransactionStatusSuccess,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockService.On("Recharge", mock.Anything, mock.MatchedBy(func(r *types.RechargeRequest) bool {
			return r.CustomerID == 1 && r.Amount == 100.0
		})).Return(expected, nil)

		req := &pb.RechargeRequest{
			CustomerId: 1,
			Amount:     100.0,
			Currency:   "CNY",
			Channel:    "alipay",
		}

		resp, err := handler.Recharge(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 100.0, resp.Transaction.Amount)
		mockService.AssertExpectations(t)
	})
}

func TestGRPCConsume(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletGRPCHandler(mockService)
	ctx := getTestContext()

	t.Run("success", func(t *testing.T) {
		orderID := uint(1)
		expected := &types.TransactionResponse{
			ID:         2,
			TenantID:   1,
			WalletID:   1,
			OrderID:    &orderID,
			Type:       entity.TransactionTypeConsume,
			Amount:     50.0,
			Currency:   "CNY",
			Channel:    entity.ChannelWallet,
			Status:     entity.TransactionStatusSuccess,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockService.On("Consume", mock.Anything, mock.MatchedBy(func(r *types.ConsumeRequest) bool {
			return r.CustomerID == 1 && r.Amount == 50.0
		})).Return(expected, nil)

		req := &pb.ConsumeRequest{
			CustomerId: 1,
			OrderId:    1,
			Amount:     50.0,
			Currency:   "CNY",
		}

		resp, err := handler.Consume(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 50.0, resp.Transaction.Amount)
		mockService.AssertExpectations(t)
	})
}

func TestGRPCRefund(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletGRPCHandler(mockService)
	ctx := getTestContext()

	t.Run("success", func(t *testing.T) {
		orderID := uint(1)
		expected := &types.TransactionResponse{
			ID:         3,
			TenantID:   1,
			WalletID:   1,
			OrderID:    &orderID,
			Type:       entity.TransactionTypeRefund,
			Amount:     50.0,
			Currency:   "CNY",
			Channel:    entity.ChannelWallet,
			Status:     entity.TransactionStatusSuccess,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockService.On("Refund", mock.Anything, mock.MatchedBy(func(r *types.RefundRequest) bool {
			return r.CustomerID == 1 && r.Amount == 50.0
		})).Return(expected, nil)

		req := &pb.RefundRequest{
			CustomerId: 1,
			OrderId:    1,
			Amount:     50.0,
			Currency:   "CNY",
		}

		resp, err := handler.Refund(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 50.0, resp.Transaction.Amount)
		mockService.AssertExpectations(t)
	})
}

func TestGRPCGetTransaction(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletGRPCHandler(mockService)
	ctx := getTestContext()

	t.Run("success", func(t *testing.T) {
		expected := &types.TransactionResponse{
			ID:         1,
			TenantID:   1,
			WalletID:   1,
			Type:       entity.TransactionTypeRecharge,
			Amount:     100.0,
			Currency:   "CNY",
			Channel:    entity.ChannelAlipay,
			Status:     entity.TransactionStatusSuccess,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockService.On("GetTransaction", mock.Anything, uint(1)).Return(expected, nil)

		resp, err := handler.GetTransaction(ctx, &pb.GetTransactionRequest{Id: 1})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint32(1), resp.Transaction.Id)
		mockService.AssertExpectations(t)
	})
}

func TestGRPCListTransactions(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletGRPCHandler(mockService)
	ctx := getTestContext()

	t.Run("success", func(t *testing.T) {
		expected := &types.ListResponse{
			Data: []types.TransactionResponse{
				{
					ID:         1,
					TenantID:   1,
					WalletID:   1,
					Type:       entity.TransactionTypeRecharge,
					Amount:     100.0,
					Currency:   "CNY",
					Channel:    entity.ChannelAlipay,
					Status:     entity.TransactionStatusSuccess,
				},
			},
			Total: 1,
			Page:  1,
			Size:  1,
		}

		mockService.On("ListTransactions", mock.Anything, mock.Anything).Return(expected, nil)

		resp, err := handler.ListTransactions(ctx, &pb.ListTransactionsRequest{WalletId: 1, Page: 1, PageSize: 20})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Total)
		assert.Len(t, resp.Transactions, 1)
		mockService.AssertExpectations(t)
	})
}
