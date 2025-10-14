package http

import (
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/plutus/internal/application/service"
	"github.com/julesChu12/fly/plutus/pkg/constants"
	"github.com/julesChu12/fly/plutus/pkg/errors"
	"github.com/julesChu12/fly/plutus/pkg/types"
)

type WalletHandler struct {
	walletService service.WalletService
}

func NewWalletHandler(walletService service.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

// @Summary Create wallet
// @Description Create a new wallet for customer
// @Tags wallets
// @Accept json
// @Produce json
// @Param wallet body types.CreateWalletRequest true "Wallet data"
// @Success 201 {object} types.Response{data=types.WalletResponse}
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 409 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/wallets [post]
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req types.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	wallet, err := h.walletService.CreateWallet(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Wallet created successfully",
		Data:    wallet,
	})
}

// @Summary Get wallet by ID
// @Description Get wallet by ID
// @Tags wallets
// @Produce json
// @Param id path int true "Wallet ID"
// @Success 200 {object} types.Response{data=types.WalletResponse}
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/wallets/{id} [get]
func (h *WalletHandler) GetWallet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	wallet, err := h.walletService.GetWallet(ctx, uint(id))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Wallet retrieved successfully",
		Data:    wallet,
	})
}

// @Summary Get wallet by customer ID
// @Description Get wallet by customer ID
// @Tags wallets
// @Produce json
// @Param customer_id path int true "Customer ID"
// @Success 200 {object} types.Response{data=types.WalletResponse}
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/wallets/customer/{customer_id} [get]
func (h *WalletHandler) GetWalletByCustomerID(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("customer_id"), 10, 32)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	wallet, err := h.walletService.GetWalletByCustomerID(ctx, uint(customerID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Wallet retrieved successfully",
		Data:    wallet,
	})
}

// @Summary List wallets
// @Description List wallets with pagination
// @Tags wallets
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param customer_id query int false "Customer ID filter"
// @Success 200 {object} types.Response{data=types.ListResponse}
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/wallets [get]
func (h *WalletHandler) ListWallets(c *gin.Context) {
	var req types.ListWalletsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	result, err := h.walletService.ListWallets(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Wallets retrieved successfully",
		Data:    result,
	})
}

// @Summary Recharge wallet
// @Description Recharge money to wallet
// @Tags transactions
// @Accept json
// @Produce json
// @Param recharge body types.RechargeRequest true "Recharge data"
// @Success 201 {object} types.Response{data=types.TransactionResponse}
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/transactions/recharge [post]
func (h *WalletHandler) Recharge(c *gin.Context) {
	var req types.RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	transaction, err := h.walletService.Recharge(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Recharge completed successfully",
		Data:    transaction,
	})
}

// @Summary Consume from wallet
// @Description Consume money from wallet
// @Tags transactions
// @Accept json
// @Produce json
// @Param consume body types.ConsumeRequest true "Consume data"
// @Success 201 {object} types.Response{data=types.TransactionResponse}
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 423 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/transactions/consume [post]
func (h *WalletHandler) Consume(c *gin.Context) {
	var req types.ConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	transaction, err := h.walletService.Consume(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Consume completed successfully",
		Data:    transaction,
	})
}

// @Summary Refund to wallet
// @Description Refund money to wallet
// @Tags transactions
// @Accept json
// @Produce json
// @Param refund body types.RefundRequest true "Refund data"
// @Success 201 {object} types.Response{data=types.TransactionResponse}
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/transactions/refund [post]
func (h *WalletHandler) Refund(c *gin.Context) {
	var req types.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	transaction, err := h.walletService.Refund(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Refund completed successfully",
		Data:    transaction,
	})
}

// @Summary Get transaction
// @Description Get transaction by ID
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} types.Response{data=types.TransactionResponse}
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/transactions/{id} [get]
func (h *WalletHandler) GetTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	transaction, err := h.walletService.GetTransaction(ctx, uint(id))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Transaction retrieved successfully",
		Data:    transaction,
	})
}

// @Summary List transactions
// @Description List transactions with pagination and filters
// @Tags transactions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param wallet_id query int false "Wallet ID filter"
// @Param order_id query int false "Order ID filter"
// @Param type query string false "Transaction type filter"
// @Param status query string false "Transaction status filter"
// @Success 200 {object} types.Response{data=types.ListResponse}
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/transactions [get]
func (h *WalletHandler) ListTransactions(c *gin.Context) {
	var req types.ListTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	result, err := h.walletService.ListTransactions(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Transactions retrieved successfully",
		Data:    result,
	})
}

// Helper methods
func (h *WalletHandler) addContextFromHeaders(c *gin.Context) *gin.Context {
	// Add tenant ID
	if tenantIDStr := c.GetHeader(constants.HeaderTenantID); tenantIDStr != "" {
		if tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32); err == nil {
			c.Set(constants.ContextKeyTenantID, uint(tenantID))
		}
	}

	// Add user ID
	if userIDStr := c.GetHeader(constants.HeaderUserID); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			c.Set(constants.ContextKeyUserID, uint(userID))
		}
	}

	// Add trace ID
	if traceID := c.GetHeader(constants.HeaderTraceID); traceID != "" {
		c.Set(constants.ContextKeyTraceID, traceID)
	}

	return c
}

func (h *WalletHandler) handleError(c *gin.Context, err error) {
	code := errors.GetErrorCode(err)
	statusCode := http.StatusInternalServerError

	switch code {
	case errors.CodeInvalidParam:
		statusCode = http.StatusBadRequest
	case errors.CodeUnauthorized:
		statusCode = http.StatusUnauthorized
	case errors.CodeForbidden:
		statusCode = http.StatusForbidden
	case errors.CodeNotFound:
		statusCode = http.StatusNotFound
	case errors.CodeDuplicate:
		statusCode = http.StatusConflict
	case errors.CodeBusiness:
		statusCode = http.StatusUnprocessableEntity
	case errors.CodeInsufficientFunds:
		statusCode = http.StatusLocked // 423
	case errors.CodeInternal:
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, types.Response{
		Code:    code,
		Message: err.Error(),
	})
}
