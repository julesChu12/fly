package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/julesChu12/fly/items/internal/application/service"
	"github.com/julesChu12/fly/items/internal/domain/item"
)

// ItemHandler 商品处理器
type ItemHandler struct {
	itemService service.ItemService
}

// NewItemHandler 创建商品处理器
func NewItemHandler(itemService service.ItemService) *ItemHandler {
	return &ItemHandler{
		itemService: itemService,
	}
}

// CreateItemRequest 创建商品请求DTO
type CreateItemRequest struct {
	Name        string     `json:"name" binding:"required,min=1,max=255"`
	Description string     `json:"description"`
	Type        item.ItemType `json:"type" binding:"required"`
	Price       float64    `json:"price" binding:"required,min=0"`
	CategoryID  uuid.UUID  `json:"category_id" binding:"required"`
	ImageURL    *string    `json:"image_url"`
	Tags        *string    `json:"tags"`

	// 服务特有字段
	Duration     *int `json:"duration,omitempty"`
	StaffRequired *bool `json:"staff_required,omitempty"`
	Capacity     *int `json:"capacity,omitempty"`

	// 产品特有字段
	Stock     *float64 `json:"stock,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
	Weight    *float64 `json:"weight,omitempty"`
	SKU       *string  `json:"sku,omitempty"`
}

// UpdateItemRequest 更新商品请求DTO
type UpdateItemRequest struct {
	Name        *string     `json:"name,omitempty"`
	Description *string     `json:"description,omitempty"`
	Price       *float64    `json:"price,omitempty"`
	ImageURL    *string     `json:"image_url,omitempty"`
	Tags        *string     `json:"tags,omitempty"`

	// 服务特有字段
	Duration     *int `json:"duration,omitempty"`
	StaffRequired *bool `json:"staff_required,omitempty"`
	Capacity     *int `json:"capacity,omitempty"`

	// 产品特有字段
	Stock     *float64 `json:"stock,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
	Weight    *float64 `json:"weight,omitempty"`
	SKU       *string  `json:"sku,omitempty"`
}

// UpdateItemStatusRequest 更新商品状态请求DTO
type UpdateItemStatusRequest struct {
	Status item.ItemStatus `json:"status" binding:"required"`
}

// CreateItem 创建商品
// @Summary Create item
// @Description Create a new item
// @Tags Items
// @Accept json
// @Produce json
// @Param item body CreateItemRequest true "Item data"
// @Success 201 {object} Response
// @Failure 400 {object} ErrorResponse
// @Router /items [post]
func (h *ItemHandler) CreateItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 转换为Service层请求
	serviceReq := &service.CreateItemRequest{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		ImageURL:    req.ImageURL,
		Tags:        req.Tags,
		Duration:    req.Duration,
		StaffRequired: req.StaffRequired,
		Capacity:    req.Capacity,
		Stock:       req.Stock,
		CostPrice:   req.CostPrice,
		Weight:      req.Weight,
		SKU:         req.SKU,
	}

	// 调用Service层
	createdItem, err := h.itemService.CreateItem(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "CREATE_FAILED",
			Message: "Failed to create item: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Code:    "CREATED",
		Message: "Item created successfully",
		Data:    createdItem,
	})
}

// GetItems 获取商品列表
// @Summary Get items
// @Description Get a list of items with pagination and filters
// @Tags Items
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param type query string false "Item type (SERVICE|PRODUCT)"
// @Param status query string false "Item status (ACTIVE|INACTIVE|DRAFT|ARCHIVED)"
// @Param category_id query string false "Category ID"
// @Success 200 {object} Response
// @Router /items [get]
func (h *ItemHandler) GetItems(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var itemType *item.ItemType
	if typeStr := c.Query("type"); typeStr != "" {
		t := item.ItemType(typeStr)
		itemType = &t
	}

	var status *item.ItemStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := item.ItemStatus(statusStr)
		status = &s
	}

	var categoryID *uuid.UUID
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		id, err := uuid.Parse(categoryIDStr)
		if err == nil {
			categoryID = &id
		}
	}

	// 构建Service层请求
	serviceReq := &service.ListItemsRequest{
		Type:       itemType,
		Status:     status,
		CategoryID: categoryID,
		Limit:      pageSize,
		Offset:     (page - 1) * pageSize,
	}

	// 调用Service层
	items, total, err := h.itemService.ListItems(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "LIST_FAILED",
			Message: "Failed to list items: " + err.Error(),
		})
		return
	}

	// 构建响应
	response := gin.H{
		"items": items,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "Items retrieved successfully",
		Data:    response,
	})
}

// GetItemByID 根据ID获取商品
// @Summary Get item by ID
// @Description Get an item by its ID
// @Tags Items
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} Response
// @Failure 404 {object} ErrorResponse
// @Router /items/{id} [get]
func (h *ItemHandler) GetItemByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid item ID format",
		})
		return
	}

	// 调用Service层
	itemData, err := h.itemService.GetItemByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Item not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "Item retrieved successfully",
		Data:    itemData,
	})
}

// UpdateItem 更新商品
// @Summary Update item
// @Description Update an existing item
// @Tags Items
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Param item body UpdateItemRequest true "Updated item data"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /items/{id} [put]
func (h *ItemHandler) UpdateItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid item ID format",
		})
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 转换为Service层请求
	serviceReq := &service.UpdateItemRequest{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Tags:        req.Tags,
		Duration:    req.Duration,
		StaffRequired: req.StaffRequired,
		Capacity:    req.Capacity,
		Stock:       req.Stock,
		CostPrice:   req.CostPrice,
		Weight:      req.Weight,
		SKU:         req.SKU,
	}

	// 调用Service层
	updatedItem, err := h.itemService.UpdateItem(c.Request.Context(), id, serviceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "UPDATE_FAILED",
			Message: "Failed to update item: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "Item updated successfully",
		Data:    updatedItem,
	})
}

// DeleteItem 删除商品
// @Summary Delete item
// @Description Delete an item by its ID
// @Tags Items
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Router /items/{id} [delete]
func (h *ItemHandler) DeleteItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid item ID format",
		})
		return
	}

	// 调用Service层
	err = h.itemService.DeleteItem(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "DELETE_FAILED",
			Message: "Failed to delete item: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "Item deleted successfully",
	})
}

// UpdateItemStatus 更新商品状态
// @Summary Update item status
// @Description Update the status of an item
// @Tags Items
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Param status body UpdateItemStatusRequest true "Status data"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Router /items/{id}/status [patch]
func (h *ItemHandler) UpdateItemStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid item ID format",
		})
		return
	}

	var req UpdateItemStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// 调用Service层
	updatedItem, err := h.itemService.UpdateItemStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "STATUS_UPDATE_FAILED",
			Message: "Failed to update item status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "Item status updated successfully",
		Data:    updatedItem,
	})
}

// SearchItems 搜索商品
// @Summary Search items
// @Description Search items by query string
// @Tags Items
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param type query string false "Item type (SERVICE|PRODUCT)"
// @Param status query string false "Item status (ACTIVE|INACTIVE|DRAFT|ARCHIVED)"
// @Param limit query int false "Result limit" default(20)
// @Success 200 {object} Response
// @Router /search/items [get]
func (h *ItemHandler) SearchItems(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_QUERY",
			Message: "Search query is required",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 50 {
		limit = 20
	}

	var itemType *item.ItemType
	if typeStr := c.Query("type"); typeStr != "" {
		t := item.ItemType(typeStr)
		itemType = &t
	}

	var status *item.ItemStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := item.ItemStatus(statusStr)
		status = &s
	}

	// 调用Service层
	serviceReq := &service.SearchItemsRequest{
		Query:  query,
		Type:   itemType,
		Status: status,
		Limit:  limit,
	}

	items, err := h.itemService.SearchItems(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SEARCH_FAILED",
			Message: "Failed to search items: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "Search completed successfully",
		Data:    items,
	})
}

// GetItemsStats 获取商品统计
// @Summary Get items statistics
// @Description Get comprehensive items statistics
// @Tags Items
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /stats/items [get]
func (h *ItemHandler) GetItemsStats(c *gin.Context) {
	// 调用Service层
	stats, err := h.itemService.GetItemsStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "STATS_FAILED",
			Message: "Failed to get items stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    "SUCCESS",
		Message: "Items statistics retrieved successfully",
		Data:    stats,
	})
}

// 响应结构体
type SuccessResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}