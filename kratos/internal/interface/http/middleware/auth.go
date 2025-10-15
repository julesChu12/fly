package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/custos"
)

// Error definitions
var (
	ErrUnauthorized = errors.New("unauthorized")
)

const (
	// AuthorizationHeader is the header key for authorization token
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer token
	BearerPrefix = "Bearer "

	// Context keys
	UserIDKey   = "user_id"
	UsernameKey = "username"
	TenantIDKey = "tenant_id"
	UserTypeKey = "user_type"
	EmailKey    = "email"
)

// AuthMiddleware handles JWT authentication via Custos
type AuthMiddleware struct {
	custosClient *custos.Client
	skipPaths    []string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(custosClient *custos.Client, skipPaths []string) *AuthMiddleware {
	return &AuthMiddleware{
		custosClient: custosClient,
		skipPaths:    skipPaths,
	}
}

// RequireAuth validates JWT token and injects user context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path should be skipped
		if m.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract token from Authorization header
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Missing authorization token",
			})
			c.Abort()
			return
		}

		// Validate token via Custos
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		result, err := m.custosClient.ValidateToken(ctx, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Failed to validate token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		if !result.IsValid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": result.ErrorMessage,
			})
			c.Abort()
			return
		}

		// Inject user context
		c.Set(UserIDKey, result.UserID)
		c.Set(UsernameKey, result.Username)
		c.Set(TenantIDKey, result.TenantID)
		c.Set(UserTypeKey, result.UserType)
		c.Set(EmailKey, result.Email)

		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return ""
	}

	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return ""
	}

	return strings.TrimPrefix(authHeader, BearerPrefix)
}

// shouldSkipPath checks if the path should skip authentication
func (m *AuthMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range m.skipPaths {
		// Support wildcard matching (e.g., "/swagger/*")
		if strings.Contains(skipPath, "*") {
			prefix := strings.TrimSuffix(skipPath, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		} else if path == skipPath {
			return true
		}
	}
	return false
}

// Context helper functions

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) (uint, error) {
	value, exists := c.Get(UserIDKey)
	if !exists {
		return 0, ErrUnauthorized
	}

	userID, ok := value.(uint)
	if !ok {
		return 0, ErrUnauthorized
	}

	return userID, nil
}

// MustGetUserID extracts user ID from gin context and panics if not found
func MustGetUserID(c *gin.Context) uint {
	userID, err := GetUserID(c)
	if err != nil {
		panic("user_id not found in context")
	}
	return userID
}

// GetTenantID extracts tenant ID from gin context
func GetTenantID(c *gin.Context) (uint, error) {
	value, exists := c.Get(TenantIDKey)
	if !exists {
		return 0, nil // Tenant ID is optional
	}

	tenantID, ok := value.(uint)
	if !ok {
		return 0, nil
	}

	return tenantID, nil
}

// GetUsername extracts username from gin context
func GetUsername(c *gin.Context) string {
	value, exists := c.Get(UsernameKey)
	if !exists {
		return ""
	}

	username, ok := value.(string)
	if !ok {
		return ""
	}

	return username
}

// GetUserType extracts user type from gin context
func GetUserType(c *gin.Context) string {
	value, exists := c.Get(UserTypeKey)
	if !exists {
		return ""
	}

	userType, ok := value.(string)
	if !ok {
		return ""
	}

	return userType
}
