package dto

import "time"

// ListUsersRequest defines the request for listing users
type ListUsersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive frozen disabled locked deleted merged"`
	Role     string `form:"role" binding:"omitempty,oneof=admin user guest"`
	UserType string `form:"user_type" binding:"omitempty,oneof=customer staff partner"`
	Keyword  string `form:"keyword" binding:"omitempty,max=100"` // Search by username or email
	TenantID *uint  `form:"tenant_id" binding:"omitempty"`
}

// ListUsersResponse defines the response for listing users
type ListUsersResponse struct {
	Users      []*AdminUserInfo `json:"users"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// AdminUserInfo contains detailed user information for admin view
type AdminUserInfo struct {
	ID               uint       `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	Nickname         string     `json:"nickname"`
	Avatar           string     `json:"avatar"`
	Status           string     `json:"status"`
	Role             string     `json:"role"`
	UserType         string     `json:"user_type"`
	TenantID         *uint      `json:"tenant_id,omitempty"`
	TokenVersion     int        `json:"token_version"`
	MergedIntoUserID *uint      `json:"merged_into_user_id,omitempty"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// UpdateUserStatusRequest defines the request for updating user status
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive frozen disabled locked deleted"`
	Reason string `json:"reason" binding:"omitempty,max=500"` // Optional reason for audit
}

// UpdateUserStatusResponse defines the response for updating user status
type UpdateUserStatusResponse struct {
	UserID    uint      `json:"user_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	UpdatedAt time.Time `json:"updated_at"`
	Message   string    `json:"message"`
}

// ForceLogoutUserRequest defines the request for force logout
type ForceLogoutUserRequest struct {
	SessionID string `json:"session_id" binding:"omitempty"` // If empty, logout all sessions
	Reason    string `json:"reason" binding:"omitempty,max=500"`
}

// ForceLogoutUserResponse defines the response for force logout
type ForceLogoutUserResponse struct {
	UserID           uint   `json:"user_id"`
	SessionsRevoked  int    `json:"sessions_revoked"`
	TokenVersionOld  int    `json:"token_version_old"`
	TokenVersionNew  int    `json:"token_version_new"`
	Message          string `json:"message"`
}

// SystemStatsResponse defines the response for system statistics
type SystemStatsResponse struct {
	TotalUsers       int64            `json:"total_users"`
	ActiveUsers      int64            `json:"active_users"`
	InactiveUsers    int64            `json:"inactive_users"`
	FrozenUsers      int64            `json:"frozen_users"`
	DeletedUsers     int64            `json:"deleted_users"`
	TotalSessions    int64            `json:"total_sessions"`
	ActiveSessions   int64            `json:"active_sessions"`
	UsersByRole      map[string]int64 `json:"users_by_role"`
	UsersByType      map[string]int64 `json:"users_by_type"`
	NewUsersToday    int64            `json:"new_users_today"`
	NewUsersThisWeek int64            `json:"new_users_this_week"`
}
