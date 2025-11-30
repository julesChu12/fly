package handler

import (
	"github.com/gin-gonic/gin"
)

// SearchHandler 搜索处理器
type SearchHandler struct {
	itemHandler *ItemHandler
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler(itemHandler *ItemHandler) *SearchHandler {
	return &SearchHandler{
		itemHandler: itemHandler,
	}
}

// SearchItems 搜索商品
// @Summary Search items
// @Description Search items by query string
// @Tags Search
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param type query string false "Item type (SERVICE|PRODUCT)"
// @Param status query string false "Item status (ACTIVE|INACTIVE|DRAFT|ARCHIVED)"
// @Param limit query int false "Result limit" default(20)
// @Success 200 {object} Response
// @Router /search/items [get]
func (h *SearchHandler) SearchItems(c *gin.Context) {
	// 委托给ItemHandler的SearchItems方法
	h.itemHandler.SearchItems(c)
}