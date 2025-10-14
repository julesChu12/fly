package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/plutus/internal/domain/entity"
	"github.com/julesChu12/fly/plutus/pkg/constants"
	"github.com/julesChu12/fly/plutus/pkg/errors"
	"github.com/julesChu12/fly/plutus/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWalletService is a mock implementation of WalletService
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

func setupTestRouter(handler *WalletHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/api")
	{
		wallets := api.Group("/wallets")
		{
			wallets.POST("", handler.CreateWallet)
			wallets.GET("", handler.ListWallets)
			wallets.GET("/:id", handler.GetWallet)
			wallets.GET("/customer/:customer_id", handler.GetWalletByCustomerID)
		}

		transactions := api.Group("/transactions")
		{
			transactions.POST("/recharge", handler.Recharge)
			transactions.POST("/consume", handler.Consume)
			transactions.POST("/refund", handler.Refund)
			transactions.GET("", handler.ListTransactions)
			transactions.GET("/:id", handler.GetTransaction)
		}
	}

	return router
}

func TestCreateWallet(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)
	router := setupTestRouter(handler)

	t.Run("success", func(t *testing.T) {
		req := &types.CreateWalletRequest{
			CustomerID: 1,
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

		mockService.On("CreateWallet", mock.Anything, req).Return(expected, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/wallets", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/wallets", bytes.NewBuffer([]byte("invalid json")))
		r.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetWallet(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)
	router := setupTestRouter(handler)

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

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/wallets/1", nil)
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockService.On("GetWallet", mock.Anything, uint(999)).Return(nil, errors.ErrWalletNotFound)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/wallets/999", nil)
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/wallets/invalid", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetWalletByCustomerID(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)
	router := setupTestRouter(handler)

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

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/wallets/customer/1", nil)
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestListWallets(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)
	router := setupTestRouter(handler)

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

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/wallets?page=1&page_size=20", nil)
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestRecharge(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)
	router := setupTestRouter(handler)

	t.Run("success", func(t *testing.T) {
		req := &types.RechargeRequest{
			CustomerID: 1,
			Amount:     100.0,
			Currency:   "CNY",
			Channel:    entity.ChannelAlipay,
		}

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

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/transactions/recharge", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/transactions/recharge", bytes.NewBuffer([]byte("{}")))
		r.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestConsume(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)
	router := setupTestRouter(handler)

	t.Run("success", func(t *testing.T) {
		orderID := uint(1)
		req := &types.ConsumeRequest{
			CustomerID: 1,
			OrderID:    &orderID,
			Amount:     50.0,
			Currency:   "CNY",
		}

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

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/transactions/consume", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestRefund(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)
	router := setupTestRouter(handler)

	t.Run("success", func(t *testing.T) {
		orderID := uint(1)
		req := &types.RefundRequest{
			CustomerID: 1,
			OrderID:    &orderID,
			Amount:     50.0,
			Currency:   "CNY",
		}

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

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/transactions/refund", bytes.NewBuffer(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestGetTransaction(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)
	router := setupTestRouter(handler)

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

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/transactions/1", nil)
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestListTransactions(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)
	router := setupTestRouter(handler)

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

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/transactions?page=1&page_size=20", nil)
		r.Header.Set(constants.HeaderTenantID, "1")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
