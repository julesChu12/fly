package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/application/usecase"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
)

// CustomerHandler handles HTTP requests for customer management
type CustomerHandler struct {
	customerProxy *usecase.CustomerProxy
}

// NewCustomerHandler creates a new CustomerHandler instance
func NewCustomerHandler(customerProxy *usecase.CustomerProxy) *CustomerHandler {
	return &CustomerHandler{
		customerProxy: customerProxy,
	}
}

// RegisterRoutes registers customer-related routes
func (h *CustomerHandler) RegisterRoutes(router *gin.RouterGroup) {
	{
		// 客户管理路由（不需要 /customers 前缀，因为组已经是 /customers）
		router.GET("", h.ListCustomers)                    // 获取客户列表
		router.POST("", h.CreateCustomer)                  // 创建客户
		router.GET("/search", h.SearchCustomers)           // 搜索客户
		router.DELETE("/batch", h.BatchDeleteCustomers)    // 批量删除客户
		router.GET("/stats", h.GetCustomerStats)           // 获取客户统计信息

		// 客户详情路由 - 使用更具体的参数名避免与联系人路由冲突
		router.GET("/customer/:id", h.GetCustomer)         // 获取客户详情
		router.PUT("/customer/:id", h.UpdateCustomer)      // 更新客户
		router.DELETE("/customer/:id", h.DeleteCustomer)   // 删除客户

		// 联系人管理 - 现在不会与 /customer/:id 冲突
		contacts := router.Group("/:customerId/contacts")
		{
			contacts.GET("", h.GetContacts)               // 获取客户联系人列表
			contacts.POST("", h.CreateContact)           // 创建联系人
			contacts.PUT("/:contactId", h.UpdateContact) // 更新联系人
			contacts.DELETE("/:contactId", h.DeleteContact) // 删除联系人
		}
	}
}

// ListCustomers godoc
// @Summary 获取客户列表
// @Description 获取客户列表，支持分页和搜索
// @Tags customers
// @Accept json
// @Produce json
// @Param search query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers [get]
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	// 构建过滤条件
	filter := &client.CustomerFilter{
		Search: c.Query("search"),
		Page:   parseIntQueryParam(c.Query("page"), 1),
		Limit:  parseIntQueryParam(c.Query("limit"), 20),
	}

	customers, total, err := h.customerProxy.ListCustomers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取客户列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"customers": customers,
			"total":      total,
			"page":       filter.Page,
			"limit":      filter.Limit,
		},
	})
}

// CreateCustomer godoc
// @Summary 创建客户
// @Description 创建新的客户
// @Tags customers
// @Accept json
// @Produce json
// @Param customer body client.CreateCustomerRequestHTTP true "客户信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req client.CreateCustomerRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	customer, err := h.customerProxy.CreateCustomer(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建客户失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data":    customer,
	})
}

// GetCustomer godoc
// @Summary 获取客户详情
// @Description 根据ID获取客户详情
// @Tags customers
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "客户ID不能为空",
		})
		return
	}

	customerID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的客户ID",
		})
		return
	}

	customer, err := h.customerProxy.GetCustomer(c.Request.Context(), uint(customerID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "客户不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    customer,
	})
}

// UpdateCustomer godoc
// @Summary 更新客户
// @Description 更新现有客户信息
// @Tags customers
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Param customer body client.UpdateCustomerRequestHTTP true "客户信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "客户ID不能为空",
		})
		return
	}

	customerID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的客户ID",
		})
		return
	}

	var req client.UpdateCustomerRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	customer, err := h.customerProxy.UpdateCustomer(c.Request.Context(), uint(customerID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "更新客户失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    customer,
	})
}

// DeleteCustomer godoc
// @Summary 删除客户
// @Description 删除指定客户
// @Tags customers
// @Accept json
// @Produce json
// @Param id path int true "客户ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "客户ID不能为空",
		})
		return
	}

	customerID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的客户ID",
		})
		return
	}

	err = h.customerProxy.DeleteCustomer(c.Request.Context(), uint(customerID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "删除客户失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// SearchCustomers godoc
// @Summary 搜索客户
// @Description 根据关键词搜索客户
// @Tags customers
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers/search [get]
func (h *CustomerHandler) SearchCustomers(c *gin.Context) {
	searchTerm := c.Query("q")
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "搜索关键词不能为空",
		})
		return
	}

	page := parseIntQueryParam(c.Query("page"), 1)
	limit := parseIntQueryParam(c.Query("limit"), 20)

	customers, total, err := h.customerProxy.SearchCustomers(c.Request.Context(), searchTerm, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索客户失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "搜索成功",
		"data": gin.H{
			"customers": customers,
			"total":      total,
			"page":       page,
			"limit":      limit,
		},
	})
}

// GetContacts godoc
// @Summary 获取客户联系人列表
// @Description 根据客户ID获取联系人列表
// @Tags customers
// @Accept json
// @Produce json
// @Param customerId path int true "客户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers/{customerId}/contacts [get]
func (h *CustomerHandler) GetContacts(c *gin.Context) {
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

	contacts, err := h.customerProxy.GetContacts(c.Request.Context(), uint(customerID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取联系人失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    contacts,
	})
}

// CreateContact godoc
// @Summary 创建联系人
// @Description 为指定客户创建新的联系人
// @Tags customers
// @Accept json
// @Produce json
// @Param customerId path int true "客户ID"
// @Param contact body client.CreateContactRequestHTTP true "联系人信息"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers/{customerId}/contacts [post]
func (h *CustomerHandler) CreateContact(c *gin.Context) {
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

	var req client.CreateContactRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	// Set customer ID from path parameter
	req.CustomerID = uint(customerID)

	contact, err := h.customerProxy.CreateContact(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "创建联系人失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data":    contact,
	})
}

// UpdateContact godoc
// @Summary 更新联系人
// @Description 更新指定联系人的信息
// @Tags customers
// @Accept json
// @Produce json
// @Param customerId path int true "客户ID"
// @Param contactId path int true "联系人ID"
// @Param contact body client.UpdateContactRequestHTTP true "联系人信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers/{customerId}/contacts/{contactId} [put]
func (h *CustomerHandler) UpdateContact(c *gin.Context) {
	contactIdStr := c.Param("contactId")
	if contactIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "联系人ID不能为空",
		})
		return
	}

	contactID, err := strconv.ParseUint(contactIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的联系人ID",
		})
		return
	}

	var req client.UpdateContactRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	contact, err := h.customerProxy.UpdateContact(c.Request.Context(), uint(contactID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "更新联系人失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    contact,
	})
}

// DeleteContact godoc
// @Summary 删除联系人
// @Description 删除指定联系人
// @Tags customers
// @Accept json
// @Produce json
// @Param customerId path int true "客户ID"
// @Param contactId path int true "联系人ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers/{customerId}/contacts/{contactId} [delete]
func (h *CustomerHandler) DeleteContact(c *gin.Context) {
	contactIdStr := c.Param("contactId")
	if contactIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "联系人ID不能为空",
		})
		return
	}

	contactID, err := strconv.ParseUint(contactIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的联系人ID",
		})
		return
	}

	err = h.customerProxy.DeleteContact(c.Request.Context(), uint(contactID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "删除联系人失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// BatchDeleteCustomers godoc
// @Summary 批量删除客户
// @Description 批量删除指定的客户
// @Tags customers
// @Accept json
// @Produce json
// @Param request body map[string][]int true "客户ID列表"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /customers/batch [delete]
func (h *CustomerHandler) BatchDeleteCustomers(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "客户ID列表不能为空",
		})
		return
	}

	// 批量删除客户
	err := h.customerProxy.BatchDeleteCustomers(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "批量删除客户失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "���量删除成功",
		"data": gin.H{
			"deleted_count": len(req.IDs),
		},
	})
}


// GetCustomerStats godoc
// @Summary 获取客户统计信息
// @Description 获取客户相关的统计数据，包括总数、活跃客户、新增客户等
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/customers/stats [get]
func (h *CustomerHandler) GetCustomerStats(c *gin.Context) {
	// 调用客户代理获取统计信息
	stats, err := h.customerProxy.GetCustomerStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取客户统计失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}
