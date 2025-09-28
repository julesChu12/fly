package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/clotho/internal/middleware"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/mora/pkg/observability"
)

// MiddlewareStatsHandler provides endpoints for monitoring middleware statistics
type MiddlewareStatsHandler struct {
	rateLimiter    *middleware.RateLimiter
	circuitBreaker *middleware.CircuitBreakerManager
	metricsRegistry *observability.MetricsRegistry
}

// NewMiddlewareStatsHandler creates a new middleware stats handler
func NewMiddlewareStatsHandler(rateLimiter *middleware.RateLimiter, circuitBreaker *middleware.CircuitBreakerManager, metricsRegistry *observability.MetricsRegistry) *MiddlewareStatsHandler {
	return &MiddlewareStatsHandler{
		rateLimiter:    rateLimiter,
		circuitBreaker: circuitBreaker,
		metricsRegistry: metricsRegistry,
	}
}

// GetRateLimiterStats godoc
// @Summary Get rate limiter statistics
// @Description Get current statistics and configuration of the rate limiter
// @Tags monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "Rate limiter statistics"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /api/v1/monitoring/rate-limiter [get]
func (h *MiddlewareStatsHandler) GetRateLimiterStats(c *gin.Context) {
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Getting rate limiter statistics")

	stats := h.rateLimiter.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GetCircuitBreakerStats godoc
// @Summary Get circuit breaker statistics
// @Description Get current statistics and state of all circuit breakers
// @Tags monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "Circuit breaker statistics"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /api/v1/monitoring/circuit-breaker [get]
func (h *MiddlewareStatsHandler) GetCircuitBreakerStats(c *gin.Context) {
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Getting circuit breaker statistics")

	stats := h.circuitBreaker.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// ResetCircuitBreakers godoc
// @Summary Reset all circuit breakers
// @Description Reset all circuit breakers to closed state
// @Tags monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "Reset successful"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /api/v1/monitoring/circuit-breaker/reset [post]
func (h *MiddlewareStatsHandler) ResetCircuitBreakers(c *gin.Context) {
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Resetting circuit breakers")

	h.circuitBreaker.Reset()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "All circuit breakers have been reset",
	})
}

// GetAllStats godoc
// @Summary Get all middleware statistics
// @Description Get comprehensive statistics for rate limiter and circuit breakers
// @Tags monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "All middleware statistics"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /api/v1/monitoring/stats [get]
func (h *MiddlewareStatsHandler) GetAllStats(c *gin.Context) {
	log := logger.NewDefault().WithContext(c.Request.Context())
	log.Info("Getting all middleware statistics")

	rateLimiterStats := h.rateLimiter.GetStats()
	circuitBreakerStats := h.circuitBreaker.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"rate_limiter":    rateLimiterStats,
			"circuit_breaker": circuitBreakerStats,
			"metrics_info": gin.H{
				"prometheus_endpoint": "/metrics",
				"service_name":        "clotho",
			},
		},
	})
}