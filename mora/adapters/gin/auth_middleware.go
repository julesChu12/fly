package gin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/auth"
)

const (
	// ContextKeyUserID is the key used to store user ID in gin context
	ContextKeyUserID = "user_id"
	// ContextKeyClaims is the key used to store claims in gin context
	ContextKeyClaims = "claims"
)

// AuthMiddlewareConfig holds the configuration for auth middleware
type AuthMiddlewareConfig struct {
	Secret string
	// SkipPaths contains paths that should skip authentication
	SkipPaths []string
}

// AuthMiddleware creates a new authentication middleware for Gin
func AuthMiddleware(config AuthMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if current path should skip authentication
		currentPath := c.Request.URL.Path
		for _, path := range config.SkipPaths {
			// Support wildcard pattern matching
			if path == currentPath {
				c.Next()
				return
			}
			// Support path/* patterns
			if strings.HasSuffix(path, "/*") {
				prefix := strings.TrimSuffix(path, "/*")
				if strings.HasPrefix(currentPath, prefix) {
					c.Next()
					return
				}
			}
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, bearerPrefix)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "missing token",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(token, config.Secret)
		if err != nil {
			var message string
			switch err {
			case auth.ErrExpiredToken:
				message = "token expired"
			case auth.ErrMalformedToken:
				message = "malformed token"
			default:
				message = "invalid token"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": message,
			})
			c.Abort()
			return
		}

		// Store claims and user ID in context
		c.Set(ContextKeyClaims, claims)
		c.Set(ContextKeyUserID, claims.UserID)

		c.Next()
	}
}

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get(ContextKeyUserID); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetClaims extracts claims from gin context
func GetClaims(c *gin.Context) *auth.Claims {
	if claims, exists := c.Get(ContextKeyClaims); exists {
		if c, ok := claims.(*auth.Claims); ok {
			return c
		}
	}
	return nil
}
