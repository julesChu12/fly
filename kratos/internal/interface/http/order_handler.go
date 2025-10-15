package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/julesChu12/fly/kratos/internal/application/service"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/kratos/pkg/constants"
	"github.com/julesChu12/fly/kratos/pkg/errors"
	"github.com/julesChu12/fly/kratos/pkg/types"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// @Summary Create order
// @Description Create a new order
// @Tags orders
// @Accept json
// @Produce json
// @Param order body types.CreateOrderRequest true "Order data"
// @Success 201 {object} types.Response{data=types.OrderResponse}
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 409 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req types.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	// Add tenant and user context
	ctx := h.addContextFromHeaders(c)

	order, err := h.orderService.CreateOrder(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Order created successfully",
		Data:    order,
	})
}

// @Summary Get order
// @Description Get order by ID
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} types.Response{data=types.OrderResponse}
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	// Add tenant and user context
	ctx := h.addContextFromHeaders(c)

	order, err := h.orderService.GetOrder(ctx, uint(id))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Order retrieved successfully",
		Data:    order,
	})
}

// @Summary Get order with items
// @Description Get order by ID with all items
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} types.Response{data=types.OrderResponse}
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/orders/{id}/items [get]
func (h *OrderHandler) GetOrderWithItems(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	// Add tenant and user context
	ctx := h.addContextFromHeaders(c)

	order, err := h.orderService.GetOrderWithItems(ctx, uint(id))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Order with items retrieved successfully",
		Data:    order,
	})
}

// @Summary Get order logs
// @Description Get status change logs for an order
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} types.Response{data=[]types.OrderStatusLogResponse}
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 404 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/orders/{id}/logs [get]
func (h *OrderHandler) GetOrderLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	ctx := h.addContextFromHeaders(c)

	logs, err := h.orderService.GetOrderLogs(ctx, uint(id))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Order logs retrieved successfully",
		Data:    logs,
	})
}

// @Summary Update order status
// @Description Update order status
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param status body types.UpdateOrderStatusRequest true "Status update data"
// @Success 200 {object} types.Response{data=types.OrderResponse}
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Failure 422 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	var req types.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	// Add tenant and user context
	ctx := h.addContextFromHeaders(c)

	order, err := h.orderService.UpdateOrderStatus(ctx, uint(id), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Order status updated successfully",
		Data:    order,
	})
}

// @Summary Delete order
// @Description Delete order by ID
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} types.Response
// @Failure 400 {object} types.Response
// @Failure 404 {object} types.Response
// @Failure 422 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/orders/{id} [delete]
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	// Add tenant and user context
	ctx := h.addContextFromHeaders(c)

	err = h.orderService.DeleteOrder(ctx, uint(id))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Order deleted successfully",
	})
}

// @Summary List orders
// @Description List orders with pagination and filters
// @Tags orders
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param customer_id query int false "Customer ID filter"
// @Param status query string false "Order status filter"
// @Success 200 {object} types.Response{data=types.ListResponse}
// @Failure 400 {object} types.Response
// @Failure 401 {object} types.Response
// @Failure 500 {object} types.Response
// @Router /api/orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	var req types.ListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest)
		return
	}

	// Add tenant and user context
	ctx := h.addContextFromHeaders(c)

	result, err := h.orderService.ListOrders(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, types.Response{
		Code:    errors.CodeSuccess,
		Message: "Orders retrieved successfully",
		Data:    result,
	})
}

// Helper methods
func (h *OrderHandler) addContextFromHeaders(c *gin.Context) context.Context {
	ctx := c.Request.Context()

	// Add tenant ID
	if value, exists := c.Get(constants.ContextKeyTenantID); exists {
		switch v := value.(type) {
		case uint:
			ctx = context.WithValue(ctx, constants.ContextKeyTenantID, v)
		case int:
			ctx = context.WithValue(ctx, constants.ContextKeyTenantID, uint(v))
		case string:
			if tenantID, err := strconv.ParseUint(v, 10, 32); err == nil {
				ctx = context.WithValue(ctx, constants.ContextKeyTenantID, uint(tenantID))
			}
		}
	}

	if tenantIDStr := c.GetHeader(constants.HeaderTenantID); tenantIDStr != "" {
		if tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32); err == nil {
			c.Set(constants.ContextKeyTenantID, uint(tenantID))
			ctx = context.WithValue(ctx, constants.ContextKeyTenantID, uint(tenantID))
		}
	}

	// Add user ID
	if value, exists := c.Get(constants.ContextKeyUserID); exists {
		switch v := value.(type) {
		case uint:
			ctx = context.WithValue(ctx, constants.ContextKeyUserID, v)
		case int:
			ctx = context.WithValue(ctx, constants.ContextKeyUserID, uint(v))
		case string:
			if userID, err := strconv.ParseUint(v, 10, 32); err == nil {
				ctx = context.WithValue(ctx, constants.ContextKeyUserID, uint(userID))
			}
		}
	}

	if userIDStr := c.GetHeader(constants.HeaderUserID); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			c.Set(constants.ContextKeyUserID, uint(userID))
			ctx = context.WithValue(ctx, constants.ContextKeyUserID, uint(userID))
		}
	}

	// Add trace ID
	if traceID := c.GetHeader(constants.HeaderTraceID); traceID != "" {
		c.Set(constants.ContextKeyTraceID, traceID)
		ctx = context.WithValue(ctx, constants.ContextKeyTraceID, traceID)
	}

	return ctx
}

func (h *OrderHandler) handleError(c *gin.Context, err error) {
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
	case errors.CodeInternal:
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, types.Response{
		Code:    code,
		Message: err.Error(),
	})
}
