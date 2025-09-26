package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "clotho",
		Version:   "0.1.0",
		Uptime:    uptime,
	}

	c.JSON(http.StatusOK, response)
}