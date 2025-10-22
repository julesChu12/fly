package handler

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/dto"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	"github.com/julesChu12/fly/custos/internal/domain/service/rbac"
	"github.com/julesChu12/fly/custos/pkg/types"
	"gorm.io/gorm"
)

type AdminHandler struct {
	userRepo        repository.UserRepository
	userProfileRepo repository.UserProfileRepository
	sessionRepo     repository.SessionRepository
	rbacSvc         *rbac.RBACService
}

func NewAdminHandler(userRepo repository.UserRepository, userProfileRepo repository.UserProfileRepository, sessionRepo repository.SessionRepository, rbacSvc *rbac.RBACService) *AdminHandler {
	return &AdminHandler{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		sessionRepo:     sessionRepo,
		rbacSvc:         rbacSvc,
	}
}

// AssignRole assigns a role to a user
// @Summary 分配角色
// @Description 为指定用户分配角色（需要管理员权限）
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body object{role=string} true "角色信息"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users/{id}/roles [post]
func (h *AdminHandler) AssignRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	// Validate role
	validRoles := []string{"admin", "user", "guest"}
	isValidRole := false
	for _, role := range validRoles {
		if req.Role == role {
			isValidRole = true
			break
		}
	}

	if !isValidRole {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	// Check if user exists
	_, err = h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Assign role
	if err := h.rbacSvc.AssignRole(c.Request.Context(), uint(userID), req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role assigned successfully"})
}

// GetUserRoles gets all roles for a user
// @Summary 获取用户角色
// @Description 获取指定用户的所有角色和权限（需要管理员权限）
// @Tags 管理员
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} object{user_id=int,roles=[]string,permissions=[]string}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users/{id}/roles [get]
func (h *AdminHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Check if user exists
	user, err := h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get roles
	roles, err := h.rbacSvc.GetUserRoles(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user roles"})
		return
	}

	// Get permissions
	permissions := h.rbacSvc.GetUserPermissions(c.Request.Context(), user)

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"roles":       roles,
		"permissions": permissions,
	})
}

// AddPolicy adds a new policy rule
// @Summary 添加策略规则
// @Description 添加新的RBAC策略规则（需要管理员权限）
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{subject=string,object=string,action=string} true "策略规则"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/policies [post]
func (h *AdminHandler) AddPolicy(c *gin.Context) {
	var req struct {
		Subject string `json:"subject" binding:"required"`
		Object  string `json:"object" binding:"required"`
		Action  string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	// Add policy
	if err := h.rbacSvc.AddPolicy(c.Request.Context(), req.Subject, req.Object, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "policy added successfully"})
}

// RemovePolicy removes a policy rule
// @Summary 删除策略规则
// @Description 删除指定的RBAC策略规则（需要管理员权限）
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{subject=string,object=string,action=string} true "策略规则"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/policies [delete]
func (h *AdminHandler) RemovePolicy(c *gin.Context) {
	var req struct {
		Subject string `json:"subject" binding:"required"`
		Object  string `json:"object" binding:"required"`
		Action  string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	// Remove policy
	if err := h.rbacSvc.RemovePolicy(c.Request.Context(), req.Subject, req.Object, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "policy removed successfully"})
}

// ListUsers lists all users with pagination and filtering (admin only)
// @Summary 用户列表
// @Description 获取用户列表，支持分页和筛选（需要管理员权限）
// @Tags 管理员
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "用户状态"
// @Param role query string false "用户角色"
// @Param user_type query string false "用户类型"
// @Param tenant_id query int false "租户ID"
// @Param keyword query string false "关键词搜索（用户名、邮箱）"
// @Success 200 {object} dto.ListUsersResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var req dto.ListUsersRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// Build filter
	filter := &repository.UserListFilter{}
	if req.Status != "" {
		filter.Status = &req.Status
	}
	if req.Role != "" {
		filter.Role = &req.Role
	}
	if req.UserType != "" {
		filter.UserType = &req.UserType
	}
	if req.TenantID != nil {
		filter.TenantID = req.TenantID
	}
	if req.Keyword != "" {
		filter.Keyword = &req.Keyword
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Query users
	users, total, err := h.userRepo.ListWithFilter(c.Request.Context(), filter, req.PageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}

	// Convert to DTO
	userInfos := make([]*dto.AdminUserInfo, len(users))
	for i, user := range users {
		userInfo := &dto.AdminUserInfo{
			ID:               user.ID,
			Username:         user.Username,
			Email:            user.Email,
			Status:           string(user.Status),
			Role:             string(user.Role),
			UserType:         string(user.UserType),
			TenantID:         user.TenantID,
			TokenVersion:     user.TokenVersion,
			MergedIntoUserID: user.MergedIntoUserID,
			LastLoginAt:      user.LastLoginAt,
			CreatedAt:        user.CreatedAt,
			UpdatedAt:        user.UpdatedAt,
		}

		// Fetch profile data
		profile, err := h.userProfileRepo.GetByUserID(c.Request.Context(), user.ID)
		if err == nil {
			userInfo.Nickname = profile.Nickname
			userInfo.Avatar = profile.Avatar
		}

		userInfos[i] = userInfo
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	response := dto.ListUsersResponse{
		Users:      userInfos,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetUser gets a single user by ID (admin only)
// @Summary 获取用户详情
// @Description 获取指定用户的详细信息（需要管理员权限）
// @Tags 管理员
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} object{user=dto.AdminUserInfo,roles=[]string,active_sessions=int}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users/{id} [get]
func (h *AdminHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// Get user roles
	roles, _ := h.rbacSvc.GetUserRoles(c.Request.Context(), user.ID)

	// Get active sessions
	sessions, _ := h.sessionRepo.ListActiveByUser(c.Request.Context(), user.ID, time.Now())

	userInfo := &dto.AdminUserInfo{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		Status:           string(user.Status),
		Role:             string(user.Role),
		UserType:         string(user.UserType),
		TenantID:         user.TenantID,
		TokenVersion:     user.TokenVersion,
		MergedIntoUserID: user.MergedIntoUserID,
		LastLoginAt:      user.LastLoginAt,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	// Fetch profile data
	profile, err := h.userProfileRepo.GetByUserID(c.Request.Context(), user.ID)
	if err == nil {
		userInfo.Nickname = profile.Nickname
		userInfo.Avatar = profile.Avatar
	}

	c.JSON(http.StatusOK, gin.H{
		"user":           userInfo,
		"roles":          roles,
		"active_sessions": len(sessions),
	})
}

// UpdateUserStatus updates user status (admin only)
// @Summary 更新用户状态
// @Description 更新用户状态（active/inactive/locked/deleted）（需要管理员权限）
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body dto.UpdateUserStatusRequest true "状态信息"
// @Success 200 {object} dto.UpdateUserStatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users/{id}/status [patch]
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	// Get user
	user, err := h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	oldStatus := string(user.Status)

	// Update status
	user.Status = types.UserStatus(req.Status)

	// If status is disabled/locked/deleted, revoke all sessions
	if req.Status == "disabled" || req.Status == "locked" || req.Status == "deleted" {
		_ = h.sessionRepo.RevokeByUser(c.Request.Context(), user.ID, time.Now())
		user.IncrementTokenVersion() // Invalidate all tokens
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user status"})
		return
	}

	response := dto.UpdateUserStatusResponse{
		UserID:    user.ID,
		OldStatus: oldStatus,
		NewStatus: req.Status,
		UpdatedAt: time.Now(),
		Message:   "user status updated successfully",
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUserRole is handled by AssignRole (admin only)
// This endpoint is kept for backward compatibility
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	h.AssignRole(c)
}

// ForceLogoutUser force logout a user (admin only)
// @Summary 强制用户登出
// @Description 强制指定用户登出（单个会话或全部会话）（需要管理员权限）
// @Tags 管理员
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body dto.ForceLogoutUserRequest false "会话信息（不传则登出所有会话）"
// @Success 200 {object} dto.ForceLogoutUserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users/{id}/force-logout [post]
func (h *AdminHandler) ForceLogoutUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req dto.ForceLogoutUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, logout all sessions
		req.SessionID = ""
	}

	// Get user
	user, err := h.userRepo.GetByID(c.Request.Context(), uint(userID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	oldTokenVersion := user.TokenVersion
	sessionsRevoked := 0

	if req.SessionID != "" {
		// Revoke specific session
		err = h.sessionRepo.Revoke(c.Request.Context(), req.SessionID, time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke session"})
			return
		}
		sessionsRevoked = 1
	} else {
		// Revoke all sessions and increment token version
		sessions, _ := h.sessionRepo.ListActiveByUser(c.Request.Context(), user.ID, time.Now())
		sessionsRevoked = len(sessions)

		err = h.sessionRepo.RevokeByUser(c.Request.Context(), user.ID, time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke sessions"})
			return
		}

		// Increment token version to invalidate all tokens
		user.IncrementTokenVersion()
		if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}
	}

	response := dto.ForceLogoutUserResponse{
		UserID:          user.ID,
		SessionsRevoked: sessionsRevoked,
		TokenVersionOld: oldTokenVersion,
		TokenVersionNew: user.TokenVersion,
		Message:         "user forcefully logged out",
	}

	c.JSON(http.StatusOK, response)
}

// GetSystemStats gets system statistics (admin only)
// @Summary 获取系统统计
// @Description 获取系统统计信息（用户数、会话数等）（需要管理员权限）
// @Tags 管理员
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SystemStatsResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/stats [get]
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user counts by status
	totalUsers, _ := h.userRepo.CountTotal(ctx)
	activeUsers, _ := h.userRepo.CountByStatus(ctx, "active")
	inactiveUsers, _ := h.userRepo.CountByStatus(ctx, "inactive")
	frozenUsers, _ := h.userRepo.CountByStatus(ctx, "frozen")
	deletedUsers, _ := h.userRepo.CountByStatus(ctx, "deleted")

	// Get session counts
	totalSessions, _ := h.sessionRepo.CountTotal(ctx)
	activeSessions, _ := h.sessionRepo.CountActive(ctx)

	// Get users by role
	usersByRole, _ := h.userRepo.CountByRole(ctx)

	// Get users by type
	usersByType, _ := h.userRepo.CountByType(ctx)

	// Get new users
	today := time.Now().Format("2006-01-02")
	newUsersToday, _ := h.userRepo.CountNewUsers(ctx, today)

	oneWeekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	newUsersThisWeek, _ := h.userRepo.CountNewUsers(ctx, oneWeekAgo)

	response := dto.SystemStatsResponse{
		TotalUsers:       totalUsers,
		ActiveUsers:      activeUsers,
		InactiveUsers:    inactiveUsers,
		FrozenUsers:      frozenUsers,
		DeletedUsers:     deletedUsers,
		TotalSessions:    totalSessions,
		ActiveSessions:   activeSessions,
		UsersByRole:      usersByRole,
		UsersByType:      usersByType,
		NewUsersToday:    newUsersToday,
		NewUsersThisWeek: newUsersThisWeek,
	}

	c.JSON(http.StatusOK, response)
}