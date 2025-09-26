package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
	"github.com/julesChu12/fly/custos/pkg/errors"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
	UsernameKey         = "username"
	UserRoleKey         = "user_role"
	SessionIDKey        = "session_id"
)

type AuthMiddleware struct {
	tokenService *token.TokenService
	sessionRepo  repository.SessionRepository
}

func NewAuthMiddleware(tokenService *token.TokenService, sessionRepo repository.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		sessionRepo:  sessionRepo,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "MISSING_AUTHORIZATION",
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_AUTHORIZATION_FORMAT",
				"message": "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := m.tokenService.ValidateToken(token)
		if err != nil {
			var code, message string
			if domainErr, ok := err.(*errors.DomainError); ok {
				code = domainErr.Code
				message = domainErr.Message
			} else {
				code = "TOKEN_VALIDATION_FAILED"
				message = "Token validation failed"
			}

			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    code,
				"message": message,
			})
			c.Abort()
			return
		}

		if err := m.ensureSessionActive(c, claims); err != nil {
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(UserRoleKey, claims.Role)
		c.Set(SessionIDKey, claims.SessionID)
		c.Next()
	}
}

func (m *AuthMiddleware) ensureSessionActive(c *gin.Context, claims *token.TokenClaims) error {
	if m.sessionRepo == nil || claims.SessionID == "" {
		return nil
	}
	session, err := m.sessionRepo.GetByID(c.Request.Context(), claims.SessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "SESSION_NOT_FOUND",
			"message": "Session not found or revoked",
		})
		return err
	}
	if !session.IsValid() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "SESSION_REVOKED",
			"message": "Session is no longer valid",
		})
		return errors.NewSessionNotFoundError()
	}
	return nil
}

func (m *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(UserRoleKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "MISSING_USER_ROLE",
				"message": "User role not found in context",
			})
			c.Abort()
			return
		}

		if userRole != role {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "INSUFFICIENT_PERMISSIONS",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get(UserIDKey); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

func GetUsername(c *gin.Context) string {
	if username, exists := c.Get(UsernameKey); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

func GetUserRole(c *gin.Context) string {
	if userRole, exists := c.Get(UserRoleKey); exists {
		if role, ok := userRole.(string); ok {
			return role
		}
	}
	return ""
}

func GetSessionID(c *gin.Context) string {
	if sessionID, exists := c.Get(SessionIDKey); exists {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}
	return ""
}
