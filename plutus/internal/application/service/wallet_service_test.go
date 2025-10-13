package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/julesChu12/fly/plutus/internal/domain/entity"
	"github.com/julesChu12/fly/plutus/pkg/constants"
	"github.com/julesChu12/fly/plutus/pkg/errors"
	"github.com/julesChu12/fly/plutus/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Test fixtures
const (
	testTenantID   uint = 1
	testCustomerID uint = 100
	testWalletID   uint = 1
)

// Mock repositories
type mockWalletRepository struct {
	wallets map[uint]*entity.Wallet
	nextID  uint
}

func newMockWalletRepository() *mockWalletRepository {
	return &mockWalletRepository{
		wallets: make(map[uint]*entity.Wallet),
		nextID:  1,
	}
}

func (r *mockWalletRepository) Create(ctx context.Context, wallet *entity.Wallet) error {
	// Check for duplicate by customer_id
	for _, w := range r.wallets {
		if w.TenantID == wallet.TenantID && w.CustomerID == wallet.CustomerID {
			return errors.ErrWalletAlreadyExists
		}
	}
	wallet.ID = r.nextID
	r.nextID++
	r.wallets[wallet.ID] = wallet
	return nil
}

func (r *mockWalletRepository) GetByID(ctx context.Context, id uint) (*entity.Wallet, error) {
	wallet, ok := r.wallets[id]
	if !ok {
		return nil, errors.ErrWalletNotFound
	}
	return wallet, nil
}

func (r *mockWalletRepository) GetByCustomerID(ctx context.Context, tenantID, customerID uint) (*entity.Wallet, error) {
	for _, wallet := range r.wallets {
		if wallet.TenantID == tenantID && wallet.CustomerID == customerID {
			return wallet, nil
		}
	}
	return nil, errors.ErrWalletNotFound
}

func (r *mockWalletRepository) List(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.Wallet, int64, error) {
	var result []*entity.Wallet
	for _, wallet := range r.wallets {
		if wallet.TenantID == tenantID {
			result = append(result, wallet)
		}
	}

	total := int64(len(result))

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(result) {
		return []*entity.Wallet{}, total, nil
	}
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], total, nil
}

func (r *mockWalletRepository) Update(ctx context.Context, wallet *entity.Wallet) error {
	if _, ok := r.wallets[wallet.ID]; !ok {
		return errors.ErrWalletNotFound
	}
	r.wallets[wallet.ID] = wallet
	return nil
}

func (r *mockWalletRepository) UpdateBalance(ctx context.Context, id uint, newBalance float64) error {
	wallet, ok := r.wallets[id]
	if !ok {
		return errors.ErrWalletNotFound
	}
	wallet.Balance = newBalance
	return nil
}

func (r *mockWalletRepository) Delete(ctx context.Context, id uint) error {
	if _, ok := r.wallets[id]; !ok {
		return errors.ErrWalletNotFound
	}
	delete(r.wallets, id)
	return nil
}

func (r *mockWalletRepository) LockForUpdate(ctx context.Context, id uint) (*entity.Wallet, error) {
	wallet, ok := r.wallets[id]
	if !ok {
		return nil, errors.ErrWalletNotFound
	}
	return wallet, nil
}

type mockTransactionRepository struct {
	transactions    map[uint]*entity.Transaction
	idempotencyKeys map[string]*entity.Transaction
	nextID          uint
}

func newMockTransactionRepository() *mockTransactionRepository {
	return &mockTransactionRepository{
		transactions:    make(map[uint]*entity.Transaction),
		idempotencyKeys: make(map[string]*entity.Transaction),
		nextID:          1,
	}
}

func (r *mockTransactionRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	// Copy transaction to avoid mutation issues
	txCopy := &entity.Transaction{
		ID:             r.nextID,
		TenantID:       tx.TenantID,
		WalletID:       tx.WalletID,
		OrderID:        tx.OrderID,
		Type:           tx.Type,
		Amount:         tx.Amount,
		Currency:       tx.Currency,
		Channel:        tx.Channel,
		Status:         tx.Status,
		IdempotencyKey: tx.IdempotencyKey,
		ReferenceNo:    tx.ReferenceNo,
		Meta:           tx.Meta,
		CreatedAt:      tx.CreatedAt,
		UpdatedAt:      tx.UpdatedAt,
	}

	// Update original transaction with generated ID
	tx.ID = r.nextID
	r.nextID++

	r.transactions[txCopy.ID] = txCopy
	if txCopy.IdempotencyKey != nil {
		r.idempotencyKeys[*txCopy.IdempotencyKey] = txCopy
	}
	return nil
}

func (r *mockTransactionRepository) GetByID(ctx context.Context, id uint) (*entity.Transaction, error) {
	tx, ok := r.transactions[id]
	if !ok {
		return nil, errors.ErrTransactionNotFound
	}
	return tx, nil
}

func (r *mockTransactionRepository) GetByIdempotencyKey(ctx context.Context, tenantID uint, key string) (*entity.Transaction, error) {
	tx, ok := r.idempotencyKeys[key]
	if !ok {
		return nil, errors.ErrTransactionNotFound
	}
	if tx.TenantID != tenantID {
		return nil, errors.ErrTransactionNotFound
	}
	return tx, nil
}

func (r *mockTransactionRepository) List(ctx context.Context, tenantID uint, walletID *uint, offset, limit int) ([]*entity.Transaction, int64, error) {
	var result []*entity.Transaction
	for _, tx := range r.transactions {
		if tx.TenantID != tenantID {
			continue
		}
		if walletID != nil && tx.WalletID != *walletID {
			continue
		}
		result = append(result, tx)
	}

	total := int64(len(result))

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(result) {
		return []*entity.Transaction{}, total, nil
	}
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], total, nil
}

func (r *mockTransactionRepository) Update(ctx context.Context, tx *entity.Transaction) error {
	if _, ok := r.transactions[tx.ID]; !ok {
		return errors.ErrTransactionNotFound
	}
	r.transactions[tx.ID] = tx
	return nil
}

func (r *mockTransactionRepository) ListByOrderID(ctx context.Context, orderID uint) ([]*entity.Transaction, error) {
	var result []*entity.Transaction
	for _, tx := range r.transactions {
		if tx.OrderID != nil && *tx.OrderID == orderID {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (r *mockTransactionRepository) ListByWalletID(ctx context.Context, walletID uint, offset, limit int) ([]*entity.Transaction, int64, error) {
	var result []*entity.Transaction
	for _, tx := range r.transactions {
		if tx.WalletID == walletID {
			result = append(result, tx)
		}
	}

	total := int64(len(result))

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(result) {
		return []*entity.Transaction{}, total, nil
	}
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], total, nil
}

type mockPaymentChannelRepository struct{}

func newMockPaymentChannelRepository() *mockPaymentChannelRepository {
	return &mockPaymentChannelRepository{}
}

func (r *mockPaymentChannelRepository) Create(ctx context.Context, channel *entity.PaymentChannelEntity) error {
	return nil
}

func (r *mockPaymentChannelRepository) GetByID(ctx context.Context, id uint) (*entity.PaymentChannelEntity, error) {
	return nil, errors.ErrPaymentChannelNotFound
}

func (r *mockPaymentChannelRepository) List(ctx context.Context, tenantID uint, offset, limit int) ([]*entity.PaymentChannelEntity, int64, error) {
	return []*entity.PaymentChannelEntity{}, 0, nil
}

func (r *mockPaymentChannelRepository) GetByName(ctx context.Context, tenantID uint, name string) (*entity.PaymentChannelEntity, error) {
	return nil, errors.ErrPaymentChannelNotFound
}

func (r *mockPaymentChannelRepository) Update(ctx context.Context, channel *entity.PaymentChannelEntity) error {
	return nil
}

func (r *mockPaymentChannelRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (r *mockPaymentChannelRepository) ListEnabled(ctx context.Context, tenantID uint) ([]*entity.PaymentChannelEntity, error) {
	return []*entity.PaymentChannelEntity{}, nil
}

// Test helper functions
func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func setupTestService(t *testing.T) (*walletService, *mockWalletRepository, *mockTransactionRepository, sqlmock.Sqlmock) {
	gormDB, mock := setupTestDB(t)
	walletRepo := newMockWalletRepository()
	transactionRepo := newMockTransactionRepository()
	channelRepo := newMockPaymentChannelRepository()

	service := &walletService{
		db:              gormDB,
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		channelRepo:     channelRepo,
	}

	return service, walletRepo, transactionRepo, mock
}

func contextWithTenant(tenantID uint) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, constants.ContextKeyTenantID, tenantID)
}

// Test cases for CreateWallet
func TestWalletService_CreateWallet_Success(t *testing.T) {
	service, walletRepo, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	req := &types.CreateWalletRequest{
		CustomerID: testCustomerID,
		Currency:   "USD",
	}

	resp, err := service.CreateWallet(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, testTenantID, resp.TenantID)
	assert.Equal(t, testCustomerID, resp.CustomerID)
	assert.Equal(t, "USD", resp.Currency)
	assert.Equal(t, 0.0, resp.Balance)
	assert.Equal(t, entity.WalletStatusActive, resp.Status)

	// Verify wallet was created in repository
	wallet, err := walletRepo.GetByCustomerID(ctx, testTenantID, testCustomerID)
	require.NoError(t, err)
	assert.Equal(t, testCustomerID, wallet.CustomerID)
}

func TestWalletService_CreateWallet_DefaultCurrency(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	req := &types.CreateWalletRequest{
		CustomerID: testCustomerID,
	}

	resp, err := service.CreateWallet(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, constants.DefaultCurrency, resp.Currency)
}

func TestWalletService_CreateWallet_NoTenantID(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := context.Background()

	req := &types.CreateWalletRequest{
		CustomerID: testCustomerID,
	}

	resp, err := service.CreateWallet(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

func TestWalletService_CreateWallet_AlreadyExists(t *testing.T) {
	service, walletRepo, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create first wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    0,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// Try to create duplicate
	req := &types.CreateWalletRequest{
		CustomerID: testCustomerID,
		Currency:   "CNY",
	}

	resp, err := service.CreateWallet(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrWalletAlreadyExists, err)
}

// Test cases for GetWallet
func TestWalletService_GetWallet_Success(t *testing.T) {
	service, walletRepo, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create a wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    100.50,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// Get wallet
	resp, err := service.GetWallet(ctx, wallet.ID)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, wallet.ID, resp.ID)
	assert.Equal(t, 100.50, resp.Balance)
}

func TestWalletService_GetWallet_NotFound(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	resp, err := service.GetWallet(ctx, 9999)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrWalletNotFound, err)
}

func TestWalletService_GetWalletByCustomerID_Success(t *testing.T) {
	service, walletRepo, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create a wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    200.00,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// Get wallet by customer ID
	resp, err := service.GetWalletByCustomerID(ctx, testCustomerID)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, testCustomerID, resp.CustomerID)
	assert.Equal(t, 200.00, resp.Balance)
}

func TestWalletService_GetWalletByCustomerID_NoTenantID(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := context.Background()

	resp, err := service.GetWalletByCustomerID(ctx, testCustomerID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// Test cases for ListWallets
func TestWalletService_ListWallets_Success(t *testing.T) {
	service, walletRepo, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create multiple wallets
	for i := uint(1); i <= 5; i++ {
		wallet := &entity.Wallet{
			TenantID:   testTenantID,
			CustomerID: 100 + i,
			Balance:    float64(i * 100),
			Currency:   "CNY",
			Status:     entity.WalletStatusActive,
		}
		err := walletRepo.Create(ctx, wallet)
		require.NoError(t, err)
	}

	// List with pagination
	req := &types.ListWalletsRequest{
		ListRequest: types.ListRequest{
			Page:     1,
			PageSize: 3,
		},
	}

	resp, err := service.ListWallets(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(5), resp.Total)
	assert.Equal(t, 3, resp.Size)
	assert.Equal(t, 1, resp.Page)
}

func TestWalletService_ListWallets_DefaultPagination(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	req := &types.ListWalletsRequest{}

	resp, err := service.ListWallets(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Page)
}

// Test cases for Recharge
func TestWalletService_Recharge_Success(t *testing.T) {
	service, walletRepo, _, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create a wallet first
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    0,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// Mock transaction
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE `wallets`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := &types.RechargeRequest{
		CustomerID: testCustomerID,
		Amount:     100.00,
		Currency:   "CNY",
		Channel:    entity.ChannelWechat,
	}

	resp, err := service.Recharge(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, entity.TransactionTypeRecharge, resp.Type)
	assert.Equal(t, 100.00, resp.Amount)
	assert.Equal(t, entity.TransactionStatusSuccess, resp.Status)
	assert.Equal(t, wallet.ID, resp.WalletID)
}

func TestWalletService_Recharge_AutoCreateWallet(t *testing.T) {
	service, walletRepo, _, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Mock transaction
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE `wallets`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := &types.RechargeRequest{
		CustomerID: testCustomerID,
		Amount:     50.00,
		Currency:   "CNY",
		Channel:    entity.ChannelAlipay,
	}

	resp, err := service.Recharge(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify wallet was auto-created
	wallet, err := walletRepo.GetByCustomerID(ctx, testTenantID, testCustomerID)
	require.NoError(t, err)
	assert.Equal(t, testCustomerID, wallet.CustomerID)
}

func TestWalletService_Recharge_InvalidAmount_TooSmall(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	req := &types.RechargeRequest{
		CustomerID: testCustomerID,
		Amount:     0.001, // Less than MinAmount
		Currency:   "CNY",
		Channel:    entity.ChannelWechat,
	}

	resp, err := service.Recharge(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrInvalidAmount, err)
}

func TestWalletService_Recharge_InvalidAmount_TooLarge(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	req := &types.RechargeRequest{
		CustomerID: testCustomerID,
		Amount:     constants.MaxAmount + 1,
		Currency:   "CNY",
		Channel:    entity.ChannelWechat,
	}

	resp, err := service.Recharge(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrInvalidAmount, err)
}

func TestWalletService_Recharge_Idempotency(t *testing.T) {
	service, walletRepo, transactionRepo, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    0,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// First request with idempotency key
	idempotencyKey := "test-key-123"
	req := &types.RechargeRequest{
		CustomerID:     testCustomerID,
		Amount:         100.00,
		Currency:       "CNY",
		Channel:        entity.ChannelWechat,
		IdempotencyKey: &idempotencyKey,
	}

	// Mock first transaction
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE `wallets`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp1, err := service.Recharge(ctx, req)
	require.NoError(t, err)

	// Manually create the transaction in mock repo for idempotency check
	tx := &entity.Transaction{
		ID:             resp1.ID,
		TenantID:       testTenantID,
		WalletID:       wallet.ID,
		Type:           entity.TransactionTypeRecharge,
		Amount:         100.00,
		Currency:       "CNY",
		Channel:        entity.ChannelWechat,
		Status:         entity.TransactionStatusSuccess,
		IdempotencyKey: &idempotencyKey,
	}
	_ = transactionRepo.Create(ctx, tx)

	// Second request with same idempotency key should return existing transaction
	resp2, err := service.Recharge(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, resp1.ID, resp2.ID)
}

func TestWalletService_Recharge_FrozenWallet(t *testing.T) {
	service, walletRepo, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create frozen wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    0,
		Currency:   "CNY",
		Status:     entity.WalletStatusFrozen,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	req := &types.RechargeRequest{
		CustomerID: testCustomerID,
		Amount:     100.00,
		Currency:   "CNY",
		Channel:    entity.ChannelWechat,
	}

	resp, err := service.Recharge(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrWalletFrozen, err)
}

// Test cases for Consume
func TestWalletService_Consume_Success(t *testing.T) {
	service, walletRepo, _, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create wallet with balance
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    200.00,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// Mock transaction with row lock
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `wallets`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "customer_id", "balance", "currency", "status"}).
			AddRow(wallet.ID, wallet.TenantID, wallet.CustomerID, wallet.Balance, wallet.Currency, wallet.Status))
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE `wallets`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	orderID := uint(12345)
	req := &types.ConsumeRequest{
		CustomerID: testCustomerID,
		OrderID:    &orderID,
		Amount:     50.00,
		Currency:   "CNY",
	}

	resp, err := service.Consume(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, entity.TransactionTypeConsume, resp.Type)
	assert.Equal(t, 50.00, resp.Amount)
	assert.Equal(t, &orderID, resp.OrderID)
	assert.Equal(t, wallet.ID, resp.WalletID)
}

func TestWalletService_Consume_InsufficientBalance(t *testing.T) {
	service, walletRepo, _, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create wallet with low balance
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    30.00,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// Mock transaction with row lock showing insufficient balance
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `wallets`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "customer_id", "balance", "currency", "status"}).
			AddRow(wallet.ID, wallet.TenantID, wallet.CustomerID, wallet.Balance, wallet.Currency, wallet.Status))
	mock.ExpectRollback()

	req := &types.ConsumeRequest{
		CustomerID: testCustomerID,
		Amount:     100.00,
		Currency:   "CNY",
	}

	resp, err := service.Consume(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	// Check error type
	insufficientErr, ok := err.(*errors.InsufficientBalanceError)
	assert.True(t, ok)
	assert.Equal(t, 30.00, insufficientErr.Current)
	assert.Equal(t, 100.00, insufficientErr.Required)
}

func TestWalletService_Consume_WalletNotFound(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	req := &types.ConsumeRequest{
		CustomerID: 99999, // Non-existent customer
		Amount:     50.00,
		Currency:   "CNY",
	}

	resp, err := service.Consume(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrWalletNotFound, err)
}

func TestWalletService_Consume_FrozenWallet(t *testing.T) {
	service, walletRepo, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create frozen wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    200.00,
		Currency:   "CNY",
		Status:     entity.WalletStatusFrozen,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	req := &types.ConsumeRequest{
		CustomerID: testCustomerID,
		Amount:     50.00,
		Currency:   "CNY",
	}

	resp, err := service.Consume(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrWalletFrozen, err)
}

func TestWalletService_Consume_Idempotency(t *testing.T) {
	service, walletRepo, transactionRepo, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    200.00,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// First request
	idempotencyKey := "consume-key-456"
	req := &types.ConsumeRequest{
		CustomerID:     testCustomerID,
		Amount:         50.00,
		Currency:       "CNY",
		IdempotencyKey: &idempotencyKey,
	}

	// Mock first transaction
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `wallets`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "customer_id", "balance", "currency", "status"}).
			AddRow(wallet.ID, wallet.TenantID, wallet.CustomerID, wallet.Balance, wallet.Currency, wallet.Status))
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE `wallets`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp1, err := service.Consume(ctx, req)
	require.NoError(t, err)

	// Manually create the transaction in mock repo for idempotency check
	tx := &entity.Transaction{
		ID:             resp1.ID,
		TenantID:       testTenantID,
		WalletID:       wallet.ID,
		Type:           entity.TransactionTypeConsume,
		Amount:         50.00,
		Currency:       "CNY",
		Channel:        entity.ChannelWallet,
		Status:         entity.TransactionStatusSuccess,
		IdempotencyKey: &idempotencyKey,
	}
	_ = transactionRepo.Create(ctx, tx)

	// Second request with same idempotency key
	resp2, err := service.Consume(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, resp1.ID, resp2.ID)
}

// Test cases for Refund
func TestWalletService_Refund_Success(t *testing.T) {
	service, walletRepo, _, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    100.00,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// Mock transaction
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE `wallets`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	orderID := uint(12345)
	refNo := "REF-12345"
	req := &types.RefundRequest{
		CustomerID:  testCustomerID,
		OrderID:     &orderID,
		Amount:      50.00,
		Currency:    "CNY",
		ReferenceNo: &refNo,
	}

	resp, err := service.Refund(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, entity.TransactionTypeRefund, resp.Type)
	assert.Equal(t, 50.00, resp.Amount)
	assert.Equal(t, &orderID, resp.OrderID)
	assert.Equal(t, &refNo, resp.ReferenceNo)
	assert.Equal(t, wallet.ID, resp.WalletID)
}

func TestWalletService_Refund_WalletNotFound(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	req := &types.RefundRequest{
		CustomerID: 99999, // Non-existent customer
		Amount:     50.00,
		Currency:   "CNY",
	}

	resp, err := service.Refund(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrWalletNotFound, err)
}

func TestWalletService_Refund_FrozenWallet(t *testing.T) {
	service, walletRepo, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create frozen wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    100.00,
		Currency:   "CNY",
		Status:     entity.WalletStatusFrozen,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	req := &types.RefundRequest{
		CustomerID: testCustomerID,
		Amount:     50.00,
		Currency:   "CNY",
	}

	resp, err := service.Refund(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrWalletFrozen, err)
}

func TestWalletService_Refund_Idempotency(t *testing.T) {
	service, walletRepo, transactionRepo, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    100.00,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// First request
	idempotencyKey := "refund-key-789"
	req := &types.RefundRequest{
		CustomerID:     testCustomerID,
		Amount:         50.00,
		Currency:       "CNY",
		IdempotencyKey: &idempotencyKey,
	}

	// Mock first transaction
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE `wallets`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp1, err := service.Refund(ctx, req)
	require.NoError(t, err)

	// Manually create the transaction in mock repo for idempotency check
	tx := &entity.Transaction{
		ID:             resp1.ID,
		TenantID:       testTenantID,
		WalletID:       wallet.ID,
		Type:           entity.TransactionTypeRefund,
		Amount:         50.00,
		Currency:       "CNY",
		Channel:        entity.ChannelWallet,
		Status:         entity.TransactionStatusSuccess,
		IdempotencyKey: &idempotencyKey,
	}
	_ = transactionRepo.Create(ctx, tx)

	// Second request with same idempotency key
	resp2, err := service.Refund(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, resp1.ID, resp2.ID)
}

// Test cases for GetTransaction
func TestWalletService_GetTransaction_Success(t *testing.T) {
	service, _, transactionRepo, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create a transaction
	tx := &entity.Transaction{
		TenantID: testTenantID,
		WalletID: testWalletID,
		Type:     entity.TransactionTypeRecharge,
		Amount:   100.00,
		Currency: "CNY",
		Channel:  entity.ChannelWechat,
		Status:   entity.TransactionStatusSuccess,
	}
	err := transactionRepo.Create(ctx, tx)
	require.NoError(t, err)

	// Get transaction
	resp, err := service.GetTransaction(ctx, tx.ID)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, tx.ID, resp.ID)
	assert.Equal(t, 100.00, resp.Amount)
}

func TestWalletService_GetTransaction_NotFound(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	resp, err := service.GetTransaction(ctx, 9999)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrTransactionNotFound, err)
}

// Test cases for ListTransactions
func TestWalletService_ListTransactions_Success(t *testing.T) {
	service, _, transactionRepo, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create multiple transactions
	for i := 1; i <= 5; i++ {
		tx := &entity.Transaction{
			TenantID: testTenantID,
			WalletID: testWalletID,
			Type:     entity.TransactionTypeRecharge,
			Amount:   float64(i * 10),
			Currency: "CNY",
			Channel:  entity.ChannelWechat,
			Status:   entity.TransactionStatusSuccess,
		}
		err := transactionRepo.Create(ctx, tx)
		require.NoError(t, err)
	}

	// List with pagination
	walletID := testWalletID
	req := &types.ListTransactionsRequest{
		ListRequest: types.ListRequest{
			Page:     1,
			PageSize: 3,
		},
		WalletID: &walletID,
	}

	resp, err := service.ListTransactions(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(5), resp.Total)
	assert.Equal(t, 3, resp.Size)
	assert.Equal(t, 1, resp.Page)
}

func TestWalletService_ListTransactions_NoTenantID(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := context.Background()

	req := &types.ListTransactionsRequest{}

	resp, err := service.ListTransactions(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, errors.ErrUnauthorized, err)
}

// Edge case tests
func TestWalletService_Recharge_WithMeta(t *testing.T) {
	service, walletRepo, _, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create wallet
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    0,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// Mock transaction
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE `wallets`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	meta := map[string]interface{}{
		"ip_address": "192.168.1.1",
		"device":     "iPhone 13",
	}

	req := &types.RechargeRequest{
		CustomerID: testCustomerID,
		Amount:     100.00,
		Currency:   "CNY",
		Channel:    entity.ChannelWechat,
		Meta:       meta,
	}

	resp, err := service.Recharge(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Meta)
}

func TestWalletService_Consume_ExactBalance(t *testing.T) {
	service, walletRepo, _, mock := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Create wallet with exact balance
	wallet := &entity.Wallet{
		TenantID:   testTenantID,
		CustomerID: testCustomerID,
		Balance:    100.00,
		Currency:   "CNY",
		Status:     entity.WalletStatusActive,
	}
	err := walletRepo.Create(ctx, wallet)
	require.NoError(t, err)

	// Mock transaction
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `wallets`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "customer_id", "balance", "currency", "status"}).
			AddRow(wallet.ID, wallet.TenantID, wallet.CustomerID, wallet.Balance, wallet.Currency, wallet.Status))
	mock.ExpectExec("INSERT INTO `transactions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE `wallets`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := &types.ConsumeRequest{
		CustomerID: testCustomerID,
		Amount:     100.00, // Exact balance
		Currency:   "CNY",
	}

	resp, err := service.Consume(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestWalletService_InvalidAmount_MinEdge(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	ctx := contextWithTenant(testTenantID)

	// Test minimum valid amount
	req := &types.RechargeRequest{
		CustomerID: testCustomerID,
		Amount:     constants.MinAmount, // 0.01
		Currency:   "CNY",
		Channel:    entity.ChannelWechat,
	}

	_, err := service.Recharge(ctx, req)

	// Should succeed with minimal setup, but might fail due to missing wallet
	// The important part is it doesn't fail on amount validation
	if err != nil {
		assert.NotEqual(t, errors.ErrInvalidAmount, err)
	}
}
