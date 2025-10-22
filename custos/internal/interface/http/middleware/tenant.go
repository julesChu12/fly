package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// TenantContextKey is the key for storing tenant ID in context.Context
	TenantContextKey contextKey = "tenant_id"
	// tenantGinKey is the key for storing tenant ID in gin.Context
	tenantGinKey = "tenant_id"
	// TenantSlugHeader is the HTTP header name for tenant slug
	TenantSlugHeader = "X-Tenant-Slug"
	// TenantIDHeader is the HTTP header name for tenant ID
	TenantIDHeader   = "X-Tenant-ID"
)

type TenantMiddleware struct {
	tenantRepo repository.TenantRepository
}

func NewTenantMiddleware(tenantRepo repository.TenantRepository) *TenantMiddleware {
	return &TenantMiddleware{
		tenantRepo: tenantRepo,
	}
}

// ResolveTenant middleware resolves tenant from various sources
func (m *TenantMiddleware) ResolveTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenantID uint

		// Method 1: Check X-Tenant-ID header
		if tenantIDStr := c.GetHeader(TenantIDHeader); tenantIDStr != "" {
			if id, parseErr := strconv.ParseUint(tenantIDStr, 10, 32); parseErr == nil {
				tenantID = uint(id)
			}
		}

		// Method 2: Check X-Tenant-Slug header
		if tenantID == 0 {
			if tenantSlug := c.GetHeader(TenantSlugHeader); tenantSlug != "" {
				tenant, err := m.tenantRepo.GetBySlug(c.Request.Context(), tenantSlug)
				if err == nil {
					tenantID = tenant.ID
				}
			}
		}

		// Method 3: Extract from subdomain
		if tenantID == 0 {
			host := c.Request.Host
			if subdomain := extractSubdomain(host); subdomain != "" {
				tenant, subErr := m.tenantRepo.GetBySlug(c.Request.Context(), subdomain)
				if subErr == nil {
					tenantID = tenant.ID
				}
			}
		}

		// Method 4: Extract from custom domain
		if tenantID == 0 {
			host := c.Request.Host
			tenant, domainErr := m.tenantRepo.GetByDomain(c.Request.Context(), host)
			if domainErr == nil {
				tenantID = tenant.ID
			}
		}

		// Set tenant ID in context if found
		if tenantID > 0 {
			// Verify tenant is active
			tenant, err := m.tenantRepo.GetByID(c.Request.Context(), tenantID)
			if err != nil || !tenant.IsActive() {
				c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or inactive tenant"})
				c.Abort()
				return
			}

			// Set tenant in context
			ctx := context.WithValue(c.Request.Context(), TenantContextKey, tenantID)
			c.Request = c.Request.WithContext(ctx)
			c.Set(tenantGinKey, tenantID)
		}

		c.Next()
	}
}

// RequireTenant middleware ensures a tenant is present
func (m *TenantMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get(tenantGinKey)
		if !exists || tenantID.(uint) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (uint, bool) {
	tenantID, ok := ctx.Value(TenantContextKey).(uint)
	return tenantID, ok
}

// GetTenantIDFromGin extracts tenant ID from Gin context
func GetTenantIDFromGin(c *gin.Context) (uint, bool) {
	if tenantID, exists := c.Get(tenantGinKey); exists {
		return tenantID.(uint), true
	}
	return 0, false
}

// extractSubdomain extracts subdomain from host
// Example: "salon1.fly.local" -> "salon1"
func extractSubdomain(host string) string {
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	parts := strings.Split(host, ".")
	if len(parts) >= 3 {
		// Assume format: subdomain.domain.tld
		return parts[0]
	}
	return ""
}

// SetTenantInContext sets tenant ID in request context (for testing)
func SetTenantInContext(c *gin.Context, tenantID uint) {
	ctx := context.WithValue(c.Request.Context(), TenantContextKey, tenantID)
	c.Request = c.Request.WithContext(ctx)
	c.Set(tenantGinKey, tenantID)
}
