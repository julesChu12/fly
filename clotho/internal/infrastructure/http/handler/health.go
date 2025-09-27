package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Uptime    string    `json:"uptime"`
}

var startTime = time.Now()

func HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime).String()

	// Log health check with trace context
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Health check requested")

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "clotho",
		Version:   "0.1.0",
		Uptime:    uptime,
	}

	c.JSON(http.StatusOK, response)
}