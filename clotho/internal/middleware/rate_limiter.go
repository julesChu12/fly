package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"golang.org/x/time/rate"
)

// RateLimiter represents a rate limiter with different strategies
type RateLimiter struct {
	// Global rate limiter
	globalLimiter *rate.Limiter

	// Per-IP rate limiters
	ipLimiters map[string]*rate.Limiter
	mu         sync.RWMutex

	// Per-User rate limiters (based on user_id from token)
	userLimiters map[string]*rate.Limiter
	userMu       sync.RWMutex

	// Configuration
	globalLimit    rate.Limit // requests per second globally
	globalBurst    int        // burst capacity globally
	perIPLimit     rate.Limit // requests per second per IP
	perIPBurst     int        // burst capacity per IP
	perUserLimit   rate.Limit // requests per second per user
	perUserBurst   int        // burst capacity per user

	// Cleanup settings
	cleanupInterval time.Duration
	maxIdleTime     time.Duration
}

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	GlobalRPS     float64       `mapstructure:"global_rps" default:"1000"`
	GlobalBurst   int           `mapstructure:"global_burst" default:"2000"`
	PerIPRPS      float64       `mapstructure:"per_ip_rps" default:"10"`
	PerIPBurst    int           `mapstructure:"per_ip_burst" default:"20"`
	PerUserRPS    float64       `mapstructure:"per_user_rps" default:"100"`
	PerUserBurst  int           `mapstructure:"per_user_burst" default:"200"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval" default:"5m"`
	MaxIdleTime   time.Duration `mapstructure:"max_idle_time" default:"10m"`
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		globalLimiter:   rate.NewLimiter(rate.Limit(config.GlobalRPS), config.GlobalBurst),
		ipLimiters:      make(map[string]*rate.Limiter),
		userLimiters:    make(map[string]*rate.Limiter),
		globalLimit:     rate.Limit(config.GlobalRPS),
		globalBurst:     config.GlobalBurst,
		perIPLimit:      rate.Limit(config.PerIPRPS),
		perIPBurst:      config.PerIPBurst,
		perUserLimit:    rate.Limit(config.PerUserRPS),
		perUserBurst:    config.PerUserBurst,
		cleanupInterval: config.CleanupInterval,
		maxIdleTime:     config.MaxIdleTime,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// getIPLimiter gets or creates a rate limiter for the given IP
func (rl *RateLimiter) getIPLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.ipLimiters[ip]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := rl.ipLimiters[ip]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rl.perIPLimit, rl.perIPBurst)
	rl.ipLimiters[ip] = limiter
	return limiter
}

// getUserLimiter gets or creates a rate limiter for the given user
func (rl *RateLimiter) getUserLimiter(userID string) *rate.Limiter {
	rl.userMu.RLock()
	limiter, exists := rl.userLimiters[userID]
	rl.userMu.RUnlock()

	if exists {
		return limiter
	}

	rl.userMu.Lock()
	defer rl.userMu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := rl.userLimiters[userID]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rl.perUserLimit, rl.perUserBurst)
	rl.userLimiters[userID] = limiter
	return limiter
}

// cleanup removes idle rate limiters to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		// Clean up IP limiters
		rl.mu.Lock()
		for ip, limiter := range rl.ipLimiters {
			// If limiter hasn't been used for a while and has no tokens reserved
			if limiter.Tokens() == float64(rl.perIPBurst) {
				delete(rl.ipLimiters, ip)
			}
		}
		rl.mu.Unlock()

		// Clean up user limiters
		rl.userMu.Lock()
		for userID, limiter := range rl.userLimiters {
			if limiter.Tokens() == float64(rl.perUserBurst) {
				delete(rl.userLimiters, userID)
			}
		}
		rl.userMu.Unlock()

		_ = now // Suppress unused variable warning
	}
}

// Middleware returns a Gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.NewDefault().WithContext(c.Request.Context())

		// Check global rate limit first
		if !rl.globalLimiter.Allow() {
			log.Warn("Global rate limit exceeded")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Global rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		// Check per-IP rate limit
		clientIP := c.ClientIP()
		ipLimiter := rl.getIPLimiter(clientIP)
		if !ipLimiter.Allow() {
			log.Warn("IP rate limit exceeded", "client_ip", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests from this IP. Please try again later.",
			})
			c.Abort()
			return
		}

		// Check per-user rate limit (if user is authenticated)
		if userID, exists := c.Get("user_id"); exists {
			userIDStr := strconv.FormatInt(userID.(int64), 10)
			userLimiter := rl.getUserLimiter(userIDStr)
			if !userLimiter.Allow() {
				log.Warn("User rate limit exceeded", "user_id", userIDStr)
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "rate_limit_exceeded",
					"message": "Too many requests for this user. Please try again later.",
				})
				c.Abort()
				return
			}
		}

		// Continue to next middleware/handler
		c.Next()
	}
}

// GetStats returns current statistics about the rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	ipCount := len(rl.ipLimiters)
	rl.mu.RUnlock()

	rl.userMu.RLock()
	userCount := len(rl.userLimiters)
	rl.userMu.RUnlock()

	return map[string]interface{}{
		"global_tokens":    rl.globalLimiter.Tokens(),
		"global_limit":     float64(rl.globalLimit),
		"global_burst":     rl.globalBurst,
		"ip_limiters":      ipCount,
		"user_limiters":    userCount,
		"per_ip_limit":     float64(rl.perIPLimit),
		"per_ip_burst":     rl.perIPBurst,
		"per_user_limit":   float64(rl.perUserLimit),
		"per_user_burst":   rl.perUserBurst,
	}
}