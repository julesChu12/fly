package handler

import (
	"github.com/gin-gonic/gin"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	itemHandler *ItemHandler
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(itemHandler *ItemHandler) *StatsHandler {
	return &StatsHandler{
		itemHandler: itemHandler,
	}
}

// GetOverviewStats 获取总览统计
// @Summary Get overview statistics
// @Description Get comprehensive overview statistics
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /stats/overview [get]
func (h *StatsHandler) GetOverviewStats(c *gin.Context) {
	// 调用ItemHandler的GetItemsStats方法来获取统计数据
	h.itemHandler.GetItemsStats(c)
}

// GetItemStats 获取商品统计
// @Summary Get items statistics
// @Description Get detailed items statistics
// @Tags Statistics
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /stats/items [get]
func (h *StatsHandler) GetItemStats(c *gin.Context) {
	// 调用ItemHandler的GetItemsStats方法
	h.itemHandler.GetItemsStats(c)
}