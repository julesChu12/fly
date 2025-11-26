package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
)

// PaymentHandler handles HTTP requests for payment and wallet management
type PaymentHandler struct {
	paymentProxy *usecase.PaymentProxy
}

// NewPaymentHandler creates a new PaymentHandler instance
func NewPaymentHandler(paymentProxy *usecase.PaymentProxy) *PaymentHandler {
	return &PaymentHandler{
		paymentProxy: paymentProxy,
	}
}

// RegisterRoutes registers payment-related routes
func (h *PaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		// Wallet management routes
		wallets := payments.Group("/wallets")
		{
			wallets.GET("", h.ListWallets)                              // 获取钱包列表
			wallets.POST("", h.CreateWallet)                          // 创建钱包
			wallets.GET("/:id", h.GetWallet)                           // 获取钱包详情
			wallets.GET("/customer/:customerId", h.GetWalletByCustomerID) // 根据客户ID获取钱包
			wallets.GET("/:id/balance", h.GetWalletBalance)             // 获取钱包余额
		}

		// Transaction routes
		transactions := payments.Group("/transactions")
		{
			transactions.GET("", h.ListTransactions)                        // 获取交易列表
			transactions.POST("/recharge", h.Recharge)                    // 充值
			transactions.POST("/consume", h.Consume)                      // 消费
			transactions.POST("/refund", h.Refund)                        // 退款
			transactions.GET("/order/:orderId", h.GetTransactionsByOrderID) // 根据订单ID获取交易记录
		}

		// Order payment processing routes
		payments.POST("/orders/:orderId/pay", h.ProcessPaymentForOrder)      // 处理订单支付
		payments.POST("/orders/:orderId/refund", h.ProcessRefundForOrder)  // 处理订单退款
	}
}

// ListWallets godoc
// @Summary 获取钱包列表
// @Description 获取钱包列表，支持分页和过滤
// @Tags wallets
// @Accept json
// @Produce json
// @Param customer_id query int false "客户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/wallets [get]
func (h *PaymentHandler) ListWallets(c *gin.Context) {
	// 构建过滤条件
	filter := &client.WalletFilter{
		Page:     parseIntQueryParam(c.Query("page"), 1),
		PageSize: parseIntQueryParam(c.Query("page_size"), 20),
	}

	// 解析客户ID参数
	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		if customerID, err := strconv.ParseUint(customerIDStr, 10, 32); err == nil {
			customerIDUint := uint(customerID)
			filter.CustomerID = &customerIDUint
		}
	}

	wallets, err := h.paymentProxy.ListWallets(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取钱包列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    wallets,
	})
}

// CreateWallet godoc
// @Summary 创建钱包
// @Description 创建新的钱包
// @Tags wallets
// @Accept json
// @Produce json
// @Param wallet body client.CreateWalletRequestHTTP true "钱包信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/wallets [post]
func (h *PaymentHandler) CreateWallet(c *gin.Context) {
	var req client.CreateWalletRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	wallet, err := h.paymentProxy.CreateWallet(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建钱包失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data":    wallet,
	})
}

// GetWallet godoc
// @Summary 获取钱包详情
// @Description 根据ID获取钱包详情
// @Tags wallets
// @Accept json
// @Produce json
// @Param id path int true "钱包ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/wallets/{id} [get]
func (h *PaymentHandler) GetWallet(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "钱包ID不能为空",
		})
		return
	}

	walletID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的钱包ID",
		})
		return
	}

	wallet, err := h.paymentProxy.GetWallet(c.Request.Context(), uint(walletID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "钱包不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    wallet,
	})
}

// GetWalletByCustomerID godoc
// @Summary 根据客户ID获取钱包
// @Description 根据客户ID获取钱包
// @Tags wallets
// @Accept json
// @Produce json
// @Param customerId path int true "客户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/wallets/customer/{customerId} [get]
func (h *PaymentHandler) GetWalletByCustomerID(c *gin.Context) {
	customerIDStr := c.Param("customerId")
	if customerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "客户ID不能为空",
		})
		return
	}

	customerID, err := strconv.ParseUint(customerIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的客户ID",
		})
		return
	}

	wallet, err := h.paymentProxy.GetWalletByCustomerID(c.Request.Context(), uint(customerID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "钱包不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    wallet,
	})
}

// GetWalletBalance godoc
// @Summary 获取钱包余额
// @Description 根据钱包ID获取当前余额
// @Tags wallets
// @Accept json
// @Produce json
// @Param id path int true "钱包ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/wallets/{id}/balance [get]
func (h *PaymentHandler) GetWalletBalance(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "钱包ID不能为空",
		})
		return
	}

	walletID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的钱包ID",
		})
		return
	}

	// 先获取钱包信息以获取客户ID
	wallet, err := h.paymentProxy.GetWallet(c.Request.Context(), uint(walletID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "钱包不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"wallet_id": wallet.ID,
			"customer_id": wallet.CustomerID,
			"balance":    wallet.Balance,
			"currency":   wallet.Currency,
		},
	})
}

// Recharge godoc
// @Summary 充值
// @Description 向钱包充值
// @Tags transactions
// @Accept json
// @Produce json
// @Param recharge body client.RechargeRequestHTTP true "充值信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/transactions/recharge [post]
func (h *PaymentHandler) Recharge(c *gin.Context) {
	var req client.RechargeRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	transaction, err := h.paymentProxy.Recharge(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "充值失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "充值成功",
		"data":    transaction,
	})
}

// Consume godoc
// @Summary 消费
// @Description 从钱包消费
// @Tags transactions
// @Accept json
// @Produce json
// @Param consume body client.ConsumeRequestHTTP true "消费信��"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/transactions/consume [post]
func (h *PaymentHandler) Consume(c *gin.Context) {
	var req client.ConsumeRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	transaction, err := h.paymentProxy.Consume(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "消费失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "消费成功",
		"data":    transaction,
	})
}

// Refund godoc
// @Summary 退款
// @Description 向钱包退款
// @Tags transactions
// @Accept json
// @Produce json
// @Param refund body client.RefundRequestHTTP true "退款信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/transactions/refund [post]
func (h *PaymentHandler) Refund(c *gin.Context) {
	var req client.RefundRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	transaction, err := h.paymentProxy.Refund(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "退款失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "退款成功",
		"data":    transaction,
	})
}

// ListTransactions godoc
// @Summary 获取交易列表
// @Description 获取交易列表，支持分页和过滤
// @Tags transactions
// @Accept json
// @Produce json
// @Param wallet_id query int false "钱包ID"
// @Param order_id query int false "订单ID"
// @Param type query string false "交易类型"
// @Param status query string false "交易状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/transactions [get]
func (h *PaymentHandler) ListTransactions(c *gin.Context) {
	// 构建过滤条件
	filter := &client.TransactionFilter{
		Page:     parseIntQueryParam(c.Query("page"), 1),
		PageSize: parseIntQueryParam(c.Query("page_size"), 20),
	}

	// 解析钱包ID参数
	if walletIDStr := c.Query("wallet_id"); walletIDStr != "" {
		if walletID, err := strconv.ParseUint(walletIDStr, 10, 32); err == nil {
			walletIDUint := uint(walletID)
			filter.WalletID = &walletIDUint
		}
	}

	// 解析订单ID参数
	if orderIDStr := c.Query("order_id"); orderIDStr != "" {
		if orderID, err := strconv.ParseUint(orderIDStr, 10, 32); err == nil {
			orderIDUint := uint(orderID)
			filter.OrderID = &orderIDUint
		}
	}

	// 解析类型参数
	if typeStr := c.Query("type"); typeStr != "" {
		transactionType := client.TransactionType(typeStr)
		filter.Type = &transactionType
	}

	// 解析状态参数
	if statusStr := c.Query("status"); statusStr != "" {
		transactionStatus := client.TransactionStatus(statusStr)
		filter.Status = &transactionStatus
	}

	transactions, err := h.paymentProxy.ListTransactions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取交易列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    transactions,
	})
}

// GetTransactionsByOrderID godoc
// @Summary 根据订单ID获取交易记录
// @Description 根据订单ID获取交易记录
// @Tags transactions
// @Accept json
// @Produce json
// @Param orderId path int true "订单ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/transactions/order/{orderId} [get]
func (h *PaymentHandler) GetTransactionsByOrderID(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的订单ID",
		})
		return
	}

	transactions, err := h.paymentProxy.GetTransactionsByOrderID(c.Request.Context(), uint(orderID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取订单交易记录失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    transactions,
	})
}

// ProcessPaymentForOrder godoc
// @Summary 处理订单支付
// @Description 处理订单支付（从客户钱包扣款）
// @Tags payments
// @Accept json
// @Produce json
// @Param orderId path int true "订单ID"
// @Param payment body map[string]interface{} true "支付信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/orders/{orderId}/pay [post]
func (h *PaymentHandler) ProcessPaymentForOrder(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的订单ID",
		})
		return
	}

	var paymentData struct {
		CustomerID uint    `json:"customer_id" binding:"required"`
		Amount     float64 `json:"amount" binding:"required,min=0.01"`
		Currency   string  `json:"currency"`
	}

	if err := c.ShouldBindJSON(&paymentData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	transaction, err := h.paymentProxy.ProcessPaymentForOrder(
		c.Request.Context(),
		paymentData.CustomerID,
		uint(orderID),
		paymentData.Amount,
		paymentData.Currency,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "处理订单支付失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "支付处理成功",
		"data":    transaction,
	})
}

// ProcessRefundForOrder godoc
// @Summary 处理订单退款
// @Description 处理订单退款（向客户钱包退款）
// @Tags payments
// @Accept json
// @Produce json
// @Param orderId path int true "订单ID"
// @Param refund body map[string]interface{} true "退款信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/orders/{orderId}/refund [post]
func (h *PaymentHandler) ProcessRefundForOrder(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的订单ID",
		})
		return
	}

	var refundData struct {
		CustomerID uint    `json:"customer_id" binding:"required"`
		Amount     float64 `json:"amount" binding:"required,min=0.01"`
		Currency   string  `json:"currency"`
		Reason     string  `json:"reason"`
	}

	if err := c.ShouldBindJSON(&refundData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	transaction, err := h.paymentProxy.ProcessRefundForOrder(
		c.Request.Context(),
		refundData.CustomerID,
		uint(orderID),
		refundData.Amount,
		refundData.Currency,
		refundData.Reason,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "处理订单退款失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "退款处理成功",
		"data":    transaction,
	})
}