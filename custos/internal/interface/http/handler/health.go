package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

type HealthResponse struct {
	Status  string `json:"status" example:"healthy"`
	Service string `json:"service" example:"custos-user-service"`
	Version string `json:"version" example:"1.0.0"`
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check godoc
// @Summary 健康检查
// @Description 检查服务健康状态
// @Tags 健康检查
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "custos-user-service",
		"version": "1.0.0",
	})
}