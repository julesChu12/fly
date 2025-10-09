package service

import (
	"context"
	"encoding/json"

	"github.com/julesChu12/fly/plutus/internal/domain/entity"
	"github.com/julesChu12/fly/plutus/internal/domain/repository"
	"github.com/julesChu12/fly/plutus/pkg/constants"
	"github.com/julesChu12/fly/plutus/pkg/database"
	"github.com/julesChu12/fly/plutus/pkg/errors"
	"github.com/julesChu12/fly/plutus/pkg/types"
	"gorm.io/gorm"
)

type WalletService interface {
	CreateWallet(ctx context.Context, req *types.CreateWalletRequest) (*types.WalletResponse, error)
	GetWallet(ctx context.Context, id uint) (*types.WalletResponse, error)
	GetWalletByCustomerID(ctx context.Context, customerID uint) (*types.WalletResponse, error)
	ListWallets(ctx context.Context, req *types.ListWalletsRequest) (*types.ListResponse, error)
	Recharge(ctx context.Context, req *types.RechargeRequest) (*types.TransactionResponse, error)
	Consume(ctx context.Context, req *types.ConsumeRequest) (*types.TransactionResponse, error)
	Refund(ctx context.Context, req *types.RefundRequest) (*types.TransactionResponse, error)
	GetTransaction(ctx context.Context, id uint) (*types.TransactionResponse, error)
	ListTransactions(ctx context.Context, req *types.ListTransactionsRequest) (*types.ListResponse, error)
}

type walletService struct {
	db              *gorm.DB
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	channelRepo     repository.PaymentChannelRepository
}

func NewWalletService(
	db *gorm.DB,
	walletRepo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	channelRepo repository.PaymentChannelRepository,
) WalletService {
	return &walletService{
		db:              db,
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		channelRepo:     channelRepo,
	}
}

func (s *walletService) CreateWallet(ctx context.Context, req *types.CreateWalletRequest) (*types.WalletResponse, error) {
	// Get tenant ID from context
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	// Check if wallet already exists for this customer
	_, err := s.walletRepo.GetByCustomerID(ctx, tenantID, req.CustomerID)
	if err == nil {
		return nil, errors.ErrWalletAlreadyExists
	}
	if err != errors.ErrWalletNotFound {
		if _, ok := err.(*errors.NotFoundError); !ok {
			return nil, err
		}
	}

	// Create wallet
	wallet := &entity.Wallet{
		TenantID:   tenantID,
		CustomerID: req.CustomerID,
		Balance:    0.0,
		Currency:   req.Currency,
		Status:     entity.WalletStatusActive,
	}

	if wallet.Currency == "" {
		wallet.Currency = constants.DefaultCurrency
	}

	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}

	return s.toWalletResponse(wallet), nil
}

func (s *walletService) GetWallet(ctx context.Context, id uint) (*types.WalletResponse, error) {
	wallet, err := s.walletRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toWalletResponse(wallet), nil
}

func (s *walletService) GetWalletByCustomerID(ctx context.Context, customerID uint) (*types.WalletResponse, error) {
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	wallet, err := s.walletRepo.GetByCustomerID(ctx, tenantID, customerID)
	if err != nil {
		return nil, err
	}

	return s.toWalletResponse(wallet), nil
}

func (s *walletService) ListWallets(ctx context.Context, req *types.ListWalletsRequest) (*types.ListResponse, error) {
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	// Set default pagination
	if req.PageSize == 0 {
		req.PageSize = constants.DefaultPageSize
	}
	if req.PageSize > constants.MaxPageSize {
		req.PageSize = constants.MaxPageSize
	}
	if req.Page < 1 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.PageSize

	wallets, total, err := s.walletRepo.List(ctx, tenantID, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]types.WalletResponse, len(wallets))
	for i, wallet := range wallets {
		responses[i] = *s.toWalletResponse(wallet)
	}

	return &types.ListResponse{
		Data:  responses,
		Total: total,
		Page:  req.Page,
		Size:  len(responses),
	}, nil
}

func (s *walletService) Recharge(ctx context.Context, req *types.RechargeRequest) (*types.TransactionResponse, error) {
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	// Validate amount
	if req.Amount < constants.MinAmount || req.Amount > constants.MaxAmount {
		return nil, errors.ErrInvalidAmount
	}

	// Check for duplicate transaction if idempotency key is provided
	if req.IdempotencyKey != nil {
		existing, err := s.transactionRepo.GetByIdempotencyKey(ctx, tenantID, *req.IdempotencyKey)
		if err == nil {
			return s.toTransactionResponse(existing), nil
		}
		if err != errors.ErrTransactionNotFound {
			if _, ok := err.(*errors.NotFoundError); !ok {
				return nil, err
			}
		}
	}

	// Get or create wallet
	wallet, err := s.walletRepo.GetByCustomerID(ctx, tenantID, req.CustomerID)
	if err != nil {
		if _, ok := err.(*errors.NotFoundError); ok {
			// Auto-create wallet
			createReq := &types.CreateWalletRequest{
				CustomerID: req.CustomerID,
				Currency:   req.Currency,
			}
			walletResp, err := s.CreateWallet(ctx, createReq)
			if err != nil {
				return nil, err
			}
			wallet = &entity.Wallet{
				ID:         walletResp.ID,
				TenantID:   walletResp.TenantID,
				CustomerID: walletResp.CustomerID,
				Balance:    walletResp.Balance,
				Currency:   walletResp.Currency,
				Status:     walletResp.Status,
			}
		} else {
			return nil, err
		}
	}

	// Check wallet status
	if wallet.Status != entity.WalletStatusActive {
		return nil, errors.ErrWalletFrozen
	}

	var transaction *entity.Transaction

	// Execute transaction and balance update atomically within a database transaction
	err = database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		// Create transaction record
		metaJson := ""
		if req.Meta != nil {
			metaBytes, _ := json.Marshal(req.Meta)
			metaJson = string(metaBytes)
		}

		transaction = &entity.Transaction{
			TenantID:       tenantID,
			WalletID:       wallet.ID,
			Type:           entity.TransactionTypeRecharge,
			Amount:         req.Amount,
			Currency:       req.Currency,
			Channel:        req.Channel,
			Status:         entity.TransactionStatusSuccess,
			IdempotencyKey: req.IdempotencyKey,
			ReferenceNo:    req.ReferenceNo,
			Meta:           metaJson,
		}

		if transaction.Currency == "" {
			transaction.Currency = constants.DefaultCurrency
		}

		// Create transaction record
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Update wallet balance
		newBalance := wallet.Balance + req.Amount
		if err := tx.Model(&entity.Wallet{}).Where("id = ?", wallet.ID).Update("balance", newBalance).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.toTransactionResponse(transaction), nil
}

func (s *walletService) Consume(ctx context.Context, req *types.ConsumeRequest) (*types.TransactionResponse, error) {
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	// Validate amount
	if req.Amount < constants.MinAmount || req.Amount > constants.MaxAmount {
		return nil, errors.ErrInvalidAmount
	}

	// Check for duplicate transaction if idempotency key is provided
	if req.IdempotencyKey != nil {
		existing, err := s.transactionRepo.GetByIdempotencyKey(ctx, tenantID, *req.IdempotencyKey)
		if err == nil {
			return s.toTransactionResponse(existing), nil
		}
		if err != errors.ErrTransactionNotFound {
			if _, ok := err.(*errors.NotFoundError); !ok {
				return nil, err
			}
		}
	}

	// Get wallet first (not locked yet) to get wallet ID
	wallet, err := s.walletRepo.GetByCustomerID(ctx, tenantID, req.CustomerID)
	if err != nil {
		return nil, err
	}

	// Check wallet status before locking
	if wallet.Status != entity.WalletStatusActive {
		return nil, errors.ErrWalletFrozen
	}

	var transaction *entity.Transaction

	// Execute in database transaction with row lock
	err = database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		// Lock wallet row for update
		var lockedWallet entity.Wallet
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&lockedWallet, wallet.ID).Error; err != nil {
			return err
		}

		// Re-check wallet status after lock
		if lockedWallet.Status != entity.WalletStatusActive {
			return errors.ErrWalletFrozen
		}

		// Check balance
		if lockedWallet.Balance < req.Amount {
			return &errors.InsufficientBalanceError{
				Message:  "insufficient balance",
				Current:  lockedWallet.Balance,
				Required: req.Amount,
			}
		}

		// Create transaction record
		metaJson := ""
		if req.Meta != nil {
			metaBytes, _ := json.Marshal(req.Meta)
			metaJson = string(metaBytes)
		}

		transaction = &entity.Transaction{
			TenantID:       tenantID,
			WalletID:       lockedWallet.ID,
			OrderID:        req.OrderID,
			Type:           entity.TransactionTypeConsume,
			Amount:         req.Amount,
			Currency:       req.Currency,
			Channel:        entity.ChannelWallet,
			Status:         entity.TransactionStatusSuccess,
			IdempotencyKey: req.IdempotencyKey,
			Meta:           metaJson,
		}

		if transaction.Currency == "" {
			transaction.Currency = constants.DefaultCurrency
		}

		// Create transaction record
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Update wallet balance
		newBalance := lockedWallet.Balance - req.Amount
		if err := tx.Model(&entity.Wallet{}).Where("id = ?", lockedWallet.ID).Update("balance", newBalance).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.toTransactionResponse(transaction), nil
}

func (s *walletService) Refund(ctx context.Context, req *types.RefundRequest) (*types.TransactionResponse, error) {
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	// Validate amount
	if req.Amount < constants.MinAmount || req.Amount > constants.MaxAmount {
		return nil, errors.ErrInvalidAmount
	}

	// Check for duplicate transaction if idempotency key is provided
	if req.IdempotencyKey != nil {
		existing, err := s.transactionRepo.GetByIdempotencyKey(ctx, tenantID, *req.IdempotencyKey)
		if err == nil {
			return s.toTransactionResponse(existing), nil
		}
		if err != errors.ErrTransactionNotFound {
			if _, ok := err.(*errors.NotFoundError); !ok {
				return nil, err
			}
		}
	}

	// Get wallet
	wallet, err := s.walletRepo.GetByCustomerID(ctx, tenantID, req.CustomerID)
	if err != nil {
		return nil, err
	}

	// Check wallet status
	if wallet.Status != entity.WalletStatusActive {
		return nil, errors.ErrWalletFrozen
	}

	var transaction *entity.Transaction

	// Execute transaction and balance update atomically within a database transaction
	err = database.WithTransaction(ctx, s.db, func(tx *gorm.DB) error {
		// Create transaction record
		metaJson := ""
		if req.Meta != nil {
			metaBytes, _ := json.Marshal(req.Meta)
			metaJson = string(metaBytes)
		}

		transaction = &entity.Transaction{
			TenantID:       tenantID,
			WalletID:       wallet.ID,
			OrderID:        req.OrderID,
			Type:           entity.TransactionTypeRefund,
			Amount:         req.Amount,
			Currency:       req.Currency,
			Channel:        entity.ChannelWallet,
			Status:         entity.TransactionStatusSuccess,
			IdempotencyKey: req.IdempotencyKey,
			ReferenceNo:    req.ReferenceNo,
			Meta:           metaJson,
		}

		if transaction.Currency == "" {
			transaction.Currency = constants.DefaultCurrency
		}

		// Create transaction record
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		// Update wallet balance
		newBalance := wallet.Balance + req.Amount
		if err := tx.Model(&entity.Wallet{}).Where("id = ?", wallet.ID).Update("balance", newBalance).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.toTransactionResponse(transaction), nil
}

func (s *walletService) GetTransaction(ctx context.Context, id uint) (*types.TransactionResponse, error) {
	transaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toTransactionResponse(transaction), nil
}

func (s *walletService) ListTransactions(ctx context.Context, req *types.ListTransactionsRequest) (*types.ListResponse, error) {
	tenantID, ok := ctx.Value(constants.ContextKeyTenantID).(uint)
	if !ok {
		return nil, errors.ErrUnauthorized
	}

	// Set default pagination
	if req.PageSize == 0 {
		req.PageSize = constants.DefaultPageSize
	}
	if req.PageSize > constants.MaxPageSize {
		req.PageSize = constants.MaxPageSize
	}
	if req.Page < 1 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.PageSize

	transactions, total, err := s.transactionRepo.List(ctx, tenantID, req.WalletID, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]types.TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		responses[i] = *s.toTransactionResponse(transaction)
	}

	return &types.ListResponse{
		Data:  responses,
		Total: total,
		Page:  req.Page,
		Size:  len(responses),
	}, nil
}

// Helper methods
func (s *walletService) toWalletResponse(wallet *entity.Wallet) *types.WalletResponse {
	return &types.WalletResponse{
		ID:         wallet.ID,
		TenantID:   wallet.TenantID,
		CustomerID: wallet.CustomerID,
		Balance:    wallet.Balance,
		Currency:   wallet.Currency,
		Status:     wallet.Status,
		CreatedAt:  wallet.CreatedAt,
		UpdatedAt:  wallet.UpdatedAt,
	}
}

func (s *walletService) toTransactionResponse(transaction *entity.Transaction) *types.TransactionResponse {
	return &types.TransactionResponse{
		ID:             transaction.ID,
		TenantID:       transaction.TenantID,
		WalletID:       transaction.WalletID,
		OrderID:        transaction.OrderID,
		Type:           transaction.Type,
		Amount:         transaction.Amount,
		Currency:       transaction.Currency,
		Channel:        transaction.Channel,
		Status:         transaction.Status,
		IdempotencyKey: transaction.IdempotencyKey,
		ReferenceNo:    transaction.ReferenceNo,
		Meta:           transaction.Meta,
		CreatedAt:      transaction.CreatedAt,
		UpdatedAt:      transaction.UpdatedAt,
	}
}
