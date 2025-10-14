package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/hermes/internal/application/service"
	"github.com/julesChu12/fly/hermes/pkg/types"
)

// ContactHandler handles HTTP requests for contact operations
// 联系方式操作的HTTP处理器
type ContactHandler struct {
	contactService service.ContactService
}

// NewContactHandler creates a new ContactHandler
func NewContactHandler(contactService service.ContactService) *ContactHandler {
	return &ContactHandler{
		contactService: contactService,
	}
}

// CreateContact godoc
// @Summary 创建联系方式
// @Description 创建新的联系方式记录
// @Tags 联系方式管理
// @Accept json
// @Produce json
// @Param contact body types.CreateContactRequest true "联系方式信息"
// @Success 201 {object} map[string]types.ContactResponse "创建成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "内部服务器错误"
// @Router /api/contacts [post]
func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req types.CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact, err := h.contactService.CreateContact(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": contact})
}

// GetContact godoc
// @Summary 获取联系方式信息
// @Description 根据联系方式ID获取详细信息
// @Tags 联系方式管理
// @Accept json
// @Produce json
// @Param id path uint true "联系方式ID"
// @Success 200 {object} map[string]types.ContactResponse "获取成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 404 {object} map[string]string "联系方式不存在"
// @Router /api/contacts/{id} [get]
func (h *ContactHandler) GetContact(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}

	contact, err := h.contactService.GetContact(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": contact})
}

// UpdateContact godoc
// @Summary 更新联系方式信息
// @Description 根据联系方式ID更新信息
// @Tags 联系方式管理
// @Accept json
// @Produce json
// @Param id path uint true "联系方式ID"
// @Param contact body types.UpdateContactRequest true "更新的联系方式信息"
// @Success 200 {object} map[string]types.ContactResponse "更新成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "内部服务器错误"
// @Router /api/contacts/{id} [put]
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}

	var req types.UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact, err := h.contactService.UpdateContact(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": contact})
}

// DeleteContact godoc
// @Summary 删除联系方式
// @Description 根据联系方式ID删除
// @Tags 联系方式管理
// @Accept json
// @Produce json
// @Param id path uint true "联系方式ID"
// @Success 200 {object} map[string]string "删除成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "内部服务器错误"
// @Router /api/contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contact id"})
		return
	}

	err = h.contactService.DeleteContact(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "contact deleted successfully"})
}

// ListContacts godoc
// @Summary 获取联系方式列表
// @Description 分页获取联系方式列表
// @Tags 联系方式管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认为1"
// @Param page_size query int false "每页大小，默认为20"
// @Success 200 {object} types.ContactListResponse "获取成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "内部服务器错误"
// @Router /api/contacts [get]
func (h *ContactHandler) ListContacts(c *gin.Context) {
	var req types.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.contactService.ListContacts(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RegisterRoutes registers the contact routes
// 注册联系方式相关路由
func (h *ContactHandler) RegisterRoutes(r *gin.RouterGroup) {
	contacts := r.Group("/contacts")
	{
		contacts.POST("", h.CreateContact)
		contacts.GET("/:id", h.GetContact)
		contacts.PUT("/:id", h.UpdateContact)
		contacts.DELETE("/:id", h.DeleteContact)
		contacts.GET("", h.ListContacts)
	}
}
