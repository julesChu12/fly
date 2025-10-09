package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/hermes/internal/application/service"
	"github.com/julesChu12/fly/hermes/pkg/types"
)

// CustomerHandler handles HTTP requests for customer operations
// 客户操作的HTTP处理器
type CustomerHandler struct {
	customerService service.CustomerService
}

// NewCustomerHandler creates a new CustomerHandler
func NewCustomerHandler(customerService service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// CreateCustomer godoc
// @Summary 创建客户
// @Description 创建新的客户记录
// @Tags 客户管理
// @Accept json
// @Produce json
// @Param customer body types.CreateCustomerRequest true "客户信息"
// @Success 201 {object} map[string]types.CustomerResponse "创建成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "内部服务器错误"
// @Router /api/customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req types.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.customerService.CreateCustomer(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": customer})
}

// GetCustomer godoc
// @Summary 获取客户信息
// @Description 根据客户ID获取客户详细信息
// @Tags 客户管理
// @Accept json
// @Produce json
// @Param id path uint true "客户ID"
// @Success 200 {object} map[string]types.CustomerResponse "获取成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 404 {object} map[string]string "客户不存在"
// @Router /api/customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	customer, err := h.customerService.GetCustomer(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customer})
}

// GetCustomerWithContacts godoc
// @Summary 获取客户及联系方式
// @Description 根据客户ID获取客户详细信息及其所有联系方式
// @Tags 客户管理
// @Accept json
// @Produce json
// @Param id path uint true "客户ID"
// @Success 200 {object} map[string]types.CustomerResponse "获取成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 404 {object} map[string]string "客户不存在"
// @Router /api/customers/{id}/contacts [get]
func (h *CustomerHandler) GetCustomerWithContacts(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	customer, err := h.customerService.GetCustomerWithContacts(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customer})
}

// UpdateCustomer godoc
// @Summary 更新客户信息
// @Description 根据客户ID更新客户信息
// @Tags 客户管理
// @Accept json
// @Produce json
// @Param id path uint true "客户ID"
// @Param customer body types.UpdateCustomerRequest true "更新的客户信息"
// @Success 200 {object} map[string]types.CustomerResponse "更新成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "内部服务器错误"
// @Router /api/customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	var req types.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.customerService.UpdateCustomer(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customer})
}

// DeleteCustomer godoc
// @Summary 删除客户
// @Description 根据客户ID删除客户
// @Tags 客户管理
// @Accept json
// @Produce json
// @Param id path uint true "客户ID"
// @Success 200 {object} map[string]string "删除成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "内部服务器错误"
// @Router /api/customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	err = h.customerService.DeleteCustomer(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted successfully"})
}

// ListCustomers godoc
// @Summary 获取客户列表
// @Description 分页获取客户列表
// @Tags 客户管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页大小，默认为20"
// @Success 200 {object} types.ListResponse "获取成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "内部服务器错误"
// @Router /api/customers [get]
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	var req types.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.customerService.ListCustomers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RegisterRoutes registers the customer routes
// 注册客户相关路由
func (h *CustomerHandler) RegisterRoutes(r *gin.RouterGroup) {
	customers := r.Group("/customers")
	{
		customers.POST("", h.CreateCustomer)
		customers.GET("/:id", h.GetCustomer)
		customers.GET("/:id/contacts", h.GetCustomerWithContacts)
		customers.PUT("/:id", h.UpdateCustomer)
		customers.DELETE("/:id", h.DeleteCustomer)
		customers.GET("", h.ListCustomers)
	}
}
