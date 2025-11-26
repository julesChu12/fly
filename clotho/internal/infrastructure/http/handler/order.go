package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
)

// OrderHandler handles HTTP requests for order management
type OrderHandler struct {
	orderProxy *usecase.OrderProxy
}

// NewOrderHandler creates a new OrderHandler instance
func NewOrderHandler(orderProxy *usecase.OrderProxy) *OrderHandler {
	return &OrderHandler{
		orderProxy: orderProxy,
	}
}

// RegisterRoutes registers order-related routes
func (h *OrderHandler) RegisterRoutes(router *gin.RouterGroup) {
	orders := router.Group("/orders")
	{
		orders.GET("", h.ListOrders)                    // 获取订单列表
		orders.POST("", h.CreateOrder)                  // 创建订单
		orders.GET("/:id", h.GetOrder)                   // 获取订单详情
		orders.GET("/:id/items", h.GetOrderWithItems)    // 获取订单详情（含商品）
		orders.GET("/:id/logs", h.GetOrderLogs)          // 获取订单状态变更日志
		orders.PATCH("/:id/status", h.UpdateOrderStatus) // 更新订单状态
		orders.DELETE("/:id", h.DeleteOrder)             // 删除订单

		// 按客户和状态查询
		orders.GET("/customer/:customerId", h.GetOrdersByCustomer) // 获取客户订单列表
		orders.GET("/status/:status", h.GetOrdersByStatus)         // 按状态获取订单列表
	}
}

// ListOrders godoc
// @Summary 获取订单列表
// @Description 获取订单列表，支持分页和过滤
// @Tags orders
// @Accept json
// @Produce json
// @Param customer_id query int false "客户ID"
// @Param status query string false "订单状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// 构建过滤条件
	filter := &client.OrderFilter{
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

	// 解析状态参数
	if statusStr := c.Query("status"); statusStr != "" {
		status := client.OrderStatus(statusStr)
		filter.Status = &status
	}

	orders, err := h.orderProxy.ListOrders(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取订单列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    orders,
	})
}

// CreateOrder godoc
// @Summary 创建订单
// @Description 创建新的订单
// @Tags orders
// @Accept json
// @Produce json
// @Param order body client.CreateOrderRequestHTTP true "订单信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req client.CreateOrderRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	order, err := h.orderProxy.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建订单失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data":    order,
	})
}

// GetOrder godoc
// @Summary 获取订单详情
// @Description 根据ID获取订单详情
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "订单ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的订单ID",
		})
		return
	}

	order, err := h.orderProxy.GetOrder(c.Request.Context(), uint(orderID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "订单不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    order,
	})
}

// GetOrderWithItems godoc
// @Summary 获取订单详情（含商品）
// @Description 根据ID获取订单详情，包含所有商品信息
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "订单ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /orders/{id}/items [get]
func (h *OrderHandler) GetOrderWithItems(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的订单ID",
		})
		return
	}

	order, err := h.orderProxy.GetOrderWithItems(c.Request.Context(), uint(orderID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "订单不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    order,
	})
}

// GetOrderLogs godoc
// @Summary 获取订单状态变更日志
// @Description 根据订单ID获取状态变更日志
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "订单ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /orders/{id}/logs [get]
func (h *OrderHandler) GetOrderLogs(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的订单ID",
		})
		return
	}

	logs, err := h.orderProxy.GetOrderLogs(c.Request.Context(), uint(orderID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取订单日志失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    logs,
	})
}

// UpdateOrderStatus godoc
// @Summary 更新订单状态
// @Description 更新订单状态
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "订单ID"
// @Param status body client.UpdateOrderStatusRequestHTTP true "状态信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的订单ID",
		})
		return
	}

	var req client.UpdateOrderStatusRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	order, err := h.orderProxy.UpdateOrderStatus(c.Request.Context(), uint(orderID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "更新订单状态失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新状态成功",
		"data":    order,
	})
}

// DeleteOrder godoc
// @Summary 删除订单
// @Description 删除指定订单
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "订单ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /orders/{id} [delete]
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的订单ID",
		})
		return
	}

	err = h.orderProxy.DeleteOrder(c.Request.Context(), uint(orderID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "删除订单失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetOrdersByCustomer godoc
// @Summary 获取客户订单列表
// @Description 根据客户ID获取订单列表
// @Tags orders
// @Accept json
// @Produce json
// @Param customerId path int true "客户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /orders/customer/{customerId} [get]
func (h *OrderHandler) GetOrdersByCustomer(c *gin.Context) {
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

	page := parseIntQueryParam(c.Query("page"), 1)
	pageSize := parseIntQueryParam(c.Query("page_size"), 20)

	orders, err := h.orderProxy.GetOrdersByCustomer(c.Request.Context(), uint(customerID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取客户订单失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    orders,
	})
}

// GetOrdersByStatus godoc
// @Summary 按状态获取订单列表
// @Description 根据状态获取订单列表
// @Tags orders
// @Accept json
// @Produce json
// @Param status path string true "订单状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /orders/status/{status} [get]
func (h *OrderHandler) GetOrdersByStatus(c *gin.Context) {
	statusStr := c.Param("status")
	if statusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "订单状态不能为空",
		})
		return
	}

	status := client.OrderStatus(statusStr)

	page := parseIntQueryParam(c.Query("page"), 1)
	pageSize := parseIntQueryParam(c.Query("page_size"), 20)

	orders, err := h.orderProxy.GetOrdersByStatus(c.Request.Context(), status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "按状态获取订单失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    orders,
	})
}