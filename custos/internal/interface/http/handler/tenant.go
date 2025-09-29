package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/domain/entity"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/interface/http/middleware"
)

type TenantHandler struct {
	tenantRepo repository.TenantRepository
	userRepo   repository.UserRepository
}

func NewTenantHandler(tenantRepo repository.TenantRepository, userRepo repository.UserRepository) *TenantHandler {
	return &TenantHandler{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
	}
}

type CreateTenantRequest struct {
	Name     string `json:"name" binding:"required"`
	Slug     string `json:"slug" binding:"required"`
	Domain   string `json:"domain"`
	Plan     string `json:"plan"`
	MaxUsers int    `json:"max_users"`
}

type TenantResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Domain    string `json:"domain"`
	Status    string `json:"status"`
	Plan      string `json:"plan"`
	MaxUsers  int    `json:"max_users"`
	UserCount int64  `json:"user_count"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateTenant creates a new tenant
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if slug already exists
	exists, err := h.tenantRepo.ExistsBySlug(c.Request.Context(), req.Slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check slug availability"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Tenant slug already exists"})
		return
	}

	// Check domain if provided
	if req.Domain != "" {
		exists, err := h.tenantRepo.ExistsByDomain(c.Request.Context(), req.Domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check domain availability"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Domain already exists"})
			return
		}
	}

	// Create tenant
	tenant := entity.NewTenant(req.Name, req.Slug)
	if req.Domain != "" {
		tenant.Domain = req.Domain
	}
	if req.Plan != "" {
		tenant.Plan = req.Plan
	}
	if req.MaxUsers > 0 {
		tenant.MaxUsers = req.MaxUsers
	}

	if err := h.tenantRepo.Create(c.Request.Context(), tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	// Get user count
	userCount, _ := h.tenantRepo.GetUserCount(c.Request.Context(), tenant.ID)

	response := &TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Domain:    tenant.Domain,
		Status:    tenant.Status,
		Plan:      tenant.Plan,
		MaxUsers:  tenant.MaxUsers,
		UserCount: userCount,
		CreatedAt: tenant.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: tenant.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusCreated, response)
}

// GetTenant retrieves tenant information
func (h *TenantHandler) GetTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := h.tenantRepo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get user count
	userCount, _ := h.tenantRepo.GetUserCount(c.Request.Context(), tenant.ID)

	response := &TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Domain:    tenant.Domain,
		Status:    tenant.Status,
		Plan:      tenant.Plan,
		MaxUsers:  tenant.MaxUsers,
		UserCount: userCount,
		CreatedAt: tenant.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: tenant.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// GetCurrentTenant retrieves current tenant from context
func (h *TenantHandler) GetCurrentTenant(c *gin.Context) {
	tenantID, exists := middleware.GetTenantIDFromGin(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No tenant context"})
		return
	}

	tenant, err := h.tenantRepo.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get user count
	userCount, _ := h.tenantRepo.GetUserCount(c.Request.Context(), tenant.ID)

	response := &TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Domain:    tenant.Domain,
		Status:    tenant.Status,
		Plan:      tenant.Plan,
		MaxUsers:  tenant.MaxUsers,
		UserCount: userCount,
		CreatedAt: tenant.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: tenant.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// ListTenants lists all tenants (admin only)
func (h *TenantHandler) ListTenants(c *gin.Context) {
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	tenants, err := h.tenantRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenants"})
		return
	}

	var responses []TenantResponse
	for _, tenant := range tenants {
		userCount, _ := h.tenantRepo.GetUserCount(c.Request.Context(), tenant.ID)
		responses = append(responses, TenantResponse{
			ID:        tenant.ID,
			Name:      tenant.Name,
			Slug:      tenant.Slug,
			Domain:    tenant.Domain,
			Status:    tenant.Status,
			Plan:      tenant.Plan,
			MaxUsers:  tenant.MaxUsers,
			UserCount: userCount,
			CreatedAt: tenant.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: tenant.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": responses,
		"limit":   limit,
		"offset":  offset,
	})
}