package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// ItemsHandler 商品代理处理器
type ItemsHandler struct {
	itemsClient *client.ItemsHTTPClient
	logger      *logger.Logger
}

// NewItemsHandler 创建新的商品处理器
func NewItemsHandler(itemsClient *client.ItemsHTTPClient, logger *logger.Logger) *ItemsHandler {
	return &ItemsHandler{
		itemsClient: itemsClient,
		logger:      logger,
	}
}

// RegisterRoutes 注册商品相关路由
func (h *ItemsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	items := rg.Group("/items")
	{
		items.POST("", h.CreateItem)
		items.GET("", h.GetItems)
		items.GET("/:id", h.GetItemByID)
		items.PUT("/:id", h.UpdateItem)
		items.DELETE("/:id", h.DeleteItem)
		items.DELETE("/batch", h.BatchDeleteItems)     // 批量删除商品
		items.PUT("/batch", h.BatchUpdateItems)       // 批量更新商品
	}

	// 搜索路由
	search := rg.Group("/search")
	{
		search.GET("/items", h.SearchItems)
	}
}

// === 商品管理相关处理器 ===

// CreateItem 创建商品
// @Summary 创建商品
// @Description 创建新的商品或服务
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param request body client.CreateItemRequest true "商品信息"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Failure 500 {object} object
// @Router /api/v1/items [post]
func (h *ItemsHandler) CreateItem(c *gin.Context) {
	var req client.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create item request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	resp, err := h.itemsClient.CreateItem(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create item", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "CREATE_FAILED",
			"message": "Failed to create item",
		})
		return
	}

	h.logger.Info("Item created successfully", "itemID", resp.ID)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Item created successfully",
		"data":    resp,
	})
}

// GetItems 获取商品列表
// @Summary 获取商品列表
// @Description 获取商品列表，支持分页和过滤
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param type query string false "商品类型 (SERVICE|PRODUCT)"
// @Param status query string false "商品状态"
// @Param category_id query string false "分类ID"
// @Param min_price query number false "最低价格"
// @Param max_price query number false "最高价格"
// @Param search query string false "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} object
// @Failure 500 {object} object
// @Router /api/v1/items [get]
func (h *ItemsHandler) GetItems(c *gin.Context) {
	req := h.buildGetItemsRequest(c)

	resp, err := h.itemsClient.GetItems(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get items", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "GET_FAILED",
			"message": "Failed to get items",
		})
		return
	}

	h.logger.Info("Items retrieved successfully", "count", len(resp.Items))
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Items retrieved successfully",
		"data":    resp,
	})
}

// GetItemByID 根据ID获取商品
// @Summary 获取商品详情
// @Description 根据ID获取单个商品的详细信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path string true "商品ID"
// @Success 200 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/items/{id} [get]
func (h *ItemsHandler) GetItemByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_ID",
			"message": "Item ID is required",
		})
		return
	}

	resp, err := h.itemsClient.GetItemByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get item", "error", err, "itemID", id)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Item not found",
		})
		return
	}

	h.logger.Info("Item retrieved successfully", "itemID", id)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Item retrieved successfully",
		"data":    resp,
	})
}

// UpdateItem 更新商品
// @Summary 更新商品
// @Description 更新现有商品的信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path string true "商品ID"
// @Param request body client.CreateItemRequest true "商品信息"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/items/{id} [put]
func (h *ItemsHandler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_ID",
			"message": "Item ID is required",
		})
		return
	}

	var req client.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update item request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	resp, err := h.itemsClient.UpdateItem(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("Failed to update item", "error", err, "itemID", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "UPDATE_FAILED",
			"message": "Failed to update item",
		})
		return
	}

	h.logger.Info("Item updated successfully", "itemID", id)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Item updated successfully",
		"data":    resp,
	})
}

// DeleteItem 删除商品
// @Summary 删除商品
// @Description 删除指定的商品
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path string true "商品ID"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/items/{id} [delete]
func (h *ItemsHandler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_ID",
			"message": "Item ID is required",
		})
		return
	}

	err := h.itemsClient.DeleteItem(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete item", "error", err, "itemID", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DELETE_FAILED",
			"message": "Failed to delete item",
		})
		return
	}

	h.logger.Info("Item deleted successfully", "itemID", id)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Item deleted successfully",
	})
}

// SearchItems 搜索商品
// @Summary 搜索商品
// @Description 根据关键词搜索商品
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词"
// @Param type query string false "商品类型过滤"
// @Param category_id query string false "分类ID过滤"
// @Param min_price query number false "最低价格"
// @Param max_price query number false "最高价格"
// @Param limit query int false "返回结果数量限制" default(20)
// @Success 200 {object} object
// @Failure 500 {object} object
// @Router /api/v1/search/items [get]
func (h *ItemsHandler) SearchItems(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_QUERY",
			"message": "Search query is required",
		})
		return
	}

	req := &client.SearchItemsRequest{
		Query: query,
	}

	// 设置可选参数
	if itemType := c.Query("type"); itemType != "" {
		req.Type = &itemType
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		req.CategoryID = &categoryID
	}
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			req.MinPrice = &minPrice
		}
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			req.MaxPrice = &maxPrice
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	items, err := h.itemsClient.SearchItems(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to search items", "error", err, "query", query)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "SEARCH_FAILED",
			"message": "Failed to search items",
		})
		return
	}

	h.logger.Info("Items searched successfully", "query", query, "count", len(items))
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Items searched successfully",
		"data": gin.H{
			"items": items,
			"total": len(items),
		},
	})
}

// buildGetItemsRequest 构建获取商品列表的请求
func (h *ItemsHandler) buildGetItemsRequest(c *gin.Context) *client.GetItemsRequest {
	req := &client.GetItemsRequest{
		Page:     1,
		PageSize: 20,
	}

	// 解析查询参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			req.PageSize = pageSize
		}
	}

	if itemType := c.Query("type"); itemType != "" {
		req.Type = &itemType
	}

	if status := c.Query("status"); status != "" {
		req.Status = &status
	}

	if categoryID := c.Query("category_id"); categoryID != "" {
		req.CategoryID = &categoryID
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			req.MinPrice = &minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			req.MaxPrice = &maxPrice
		}
	}

	if search := c.Query("search"); search != "" {
		req.Search = &search
	}

	return req
}

// BatchDeleteItems 批量删除商品
// @Summary 批量删除商品
// @Description 批量删除指定的商品
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param request body map[string][]string true "商品ID列表"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /items/batch [delete]
func (h *ItemsHandler) BatchDeleteItems(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid batch delete items request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_IDS",
			"message": "Item IDs list cannot be empty",
		})
		return
	}

	// 批量删除商品
	err := h.itemsClient.BatchDeleteItems(c.Request.Context(), req.IDs)
	if err != nil {
		h.logger.Error("Failed to batch delete items", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DELETE_FAILED",
			"message": "Failed to batch delete items",
		})
		return
	}

	h.logger.Info("Items batch deleted successfully", "count", len(req.IDs))
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Items batch deleted successfully",
		"data": gin.H{
			"deleted_count": len(req.IDs),
		},
	})
}

// BatchUpdateItems 批量更新商品
// @Summary 批量更新商品
// @Description 批量更新商品信息（如状态、价格等）
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param request body client.BatchUpdateItemsRequest true "批量更新请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /items/batch [put]
func (h *ItemsHandler) BatchUpdateItems(c *gin.Context) {
	var req client.BatchUpdateItemsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid batch update items request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_IDS",
			"message": "Item IDs list cannot be empty",
		})
		return
	}

	// 批量更新商品
	items, err := h.itemsClient.BatchUpdateItems(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to batch update items", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "UPDATE_FAILED",
			"message": "Failed to batch update items",
		})
		return
	}

	h.logger.Info("Items batch updated successfully", "count", len(req.IDs))
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Items batch updated successfully",
		"data": gin.H{
			"updated_count": len(items),
			"items":         items,
		},
	})
}