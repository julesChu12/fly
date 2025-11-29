package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// CategoriesHandler 分类代理处理器
type CategoriesHandler struct {
	itemsClient *client.ItemsHTTPClient
	logger      *logger.Logger
}

// NewCategoriesHandler 创建新的分类处理器
func NewCategoriesHandler(itemsClient *client.ItemsHTTPClient, logger *logger.Logger) *CategoriesHandler {
	return &CategoriesHandler{
		itemsClient: itemsClient,
		logger:      logger,
	}
}

// RegisterRoutes 注册分类相关路由
func (h *CategoriesHandler) RegisterRoutes(rg *gin.RouterGroup) {
	categories := rg.Group("/categories")
	{
		categories.POST("", h.CreateCategory)
		categories.GET("", h.GetCategories)
		categories.GET("/tree", h.GetCategoryTree)
		categories.GET("/:id", h.GetCategoryByID)
		categories.PUT("/:id", h.UpdateCategory)
		categories.DELETE("/:id", h.DeleteCategory)
	}
}

// === 分类管理相关处理器 ===

// CreateCategory 创建分类
// @Summary 创建分类
// @Description 创建新的商品分类
// @Tags 分类管理
// @Accept json
// @Produce json
// @Param request body client.CreateCategoryRequest true "分类信息"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Failure 500 {object} object
// @Router /api/v1/categories [post]
func (h *CategoriesHandler) CreateCategory(c *gin.Context) {
	var req client.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid create category request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	resp, err := h.itemsClient.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create category", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "CREATE_FAILED",
			"message": "Failed to create category",
		})
		return
	}

	h.logger.Info("Category created successfully", "categoryID", resp.ID)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Category created successfully",
		"data":    resp,
	})
}

// GetCategories 获取分类列表
// @Summary 获取分类列表
// @Description 获取所有商品分类的列表
// @Tags 分类管理
// @Accept json
// @Produce json
// @Success 200 {object} object
// @Failure 500 {object} object
// @Router /api/v1/categories [get]
func (h *CategoriesHandler) GetCategories(c *gin.Context) {
	categories, err := h.itemsClient.GetCategories(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get categories", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "GET_FAILED",
			"message": "Failed to get categories",
		})
		return
	}

	h.logger.Info("Categories retrieved successfully", "count", len(categories))
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Categories retrieved successfully",
		"data": gin.H{
			"categories": categories,
			"total":      len(categories),
		},
	})
}

// GetCategoryTree 获取分类树
// @Summary 获取分类树
// @Description 获取层次化的分类树结构
// @Tags 分类管理
// @Accept json
// @Produce json
// @Success 200 {object} object
// @Failure 500 {object} object
// @Router /api/v1/categories/tree [get]
func (h *CategoriesHandler) GetCategoryTree(c *gin.Context) {
	tree, err := h.itemsClient.GetCategoryTree(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get category tree", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "GET_FAILED",
			"message": "Failed to get category tree",
		})
		return
	}

	h.logger.Info("Category tree retrieved successfully", "rootCount", len(tree))
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Category tree retrieved successfully",
		"data": gin.H{
			"tree":  tree,
			"total": len(tree),
		},
	})
}

// GetCategoryByID 根据ID获取分类
// @Summary 获取分类详情
// @Description 根据ID获取单个分类的详细信息
// @Tags 分类管理
// @Accept json
// @Produce json
// @Param id path string true "分类ID"
// @Success 200 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/categories/{id} [get]
func (h *CategoriesHandler) GetCategoryByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_ID",
			"message": "Category ID is required",
		})
		return
	}

	resp, err := h.itemsClient.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get category", "error", err, "categoryID", id)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Category not found",
		})
		return
	}

	h.logger.Info("Category retrieved successfully", "categoryID", id)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Category retrieved successfully",
		"data":    resp,
	})
}

// UpdateCategory 更新分类
// @Summary 更新分类
// @Description 更新现有分类的信息
// @Tags 分类管理
// @Accept json
// @Produce json
// @Param id path string true "分类ID"
// @Param request body client.CreateCategoryRequest true "分类信息"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/categories/{id} [put]
func (h *CategoriesHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_ID",
			"message": "Category ID is required",
		})
		return
	}

	var req client.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid update category request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request parameters",
		})
		return
	}

	resp, err := h.itemsClient.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("Failed to update category", "error", err, "categoryID", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "UPDATE_FAILED",
			"message": "Failed to update category",
		})
		return
	}

	h.logger.Info("Category updated successfully", "categoryID", id)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Category updated successfully",
		"data":    resp,
	})
}

// DeleteCategory 删除分类
// @Summary 删除分类
// @Description 删除指定的分类
// @Tags 分类管理
// @Accept json
// @Produce json
// @Param id path string true "分类ID"
// @Success 200 {object} object
// @Failure 400 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/categories/{id} [delete]
func (h *CategoriesHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_ID",
			"message": "Category ID is required",
		})
		return
	}

	err := h.itemsClient.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete category", "error", err, "categoryID", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "DELETE_FAILED",
			"message": "Failed to delete category",
		})
		return
	}

	h.logger.Info("Category deleted successfully", "categoryID", id)
	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "Category deleted successfully",
	})
}