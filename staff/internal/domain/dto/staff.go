package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/staff/internal/domain/entity"
)

// StaffFilter 员工查询过滤器
type StaffFilter struct {
	Search      *string               `json:"search"`
	Department  *string               `json:"department"`
	RoleID      *string               `json:"role_id"`
	Status      *entity.StaffStatus   `json:"status"`
	IsAvailable *bool                 `json:"is_available"`
	MinAge      *int                  `json:"min_age"`
	MaxAge      *int                  `json:"max_age"`
	Gender      *entity.Gender        `json:"gender"`
	Skills      []string              `json:"skills"`
	Page        int                   `json:"page"`
	Limit       int                   `json:"limit"`
	Sort        string                `json:"sort"`     // 排序字段
	Order       string                `json:"order"`    // 排序方向: asc/desc
}

// SetDefaults 设置默认值
func (f *StaffFilter) SetDefaults() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Sort == "" {
		f.Sort = "created_at"
	}
	if f.Order == "" {
		f.Order = "desc"
	}
}

// CreateStaffRequest 创建员工请求
type CreateStaffRequest struct {
	UserID           *string     `json:"user_id" binding:"omitempty,uuid"`
	Name             string      `json:"name" binding:"required,min=1,max=100"`
	Email            string      `json:"email" binding:"required,email,max=255"`
	Phone            string      `json:"phone" binding:"omitempty,max=20"`
	Gender           entity.Gender `json:"gender" binding:"omitempty,oneof=male female other"`
	Birthday         *time.Time  `json:"birthday"`
	Avatar           *string     `json:"avatar"`
	Department       string      `json:"department" binding:"required,max=100"`
	Position         string      `json:"position" binding:"required,max=100"`
	RoleID           string      `json:"role_id" binding:"required,uuid"`
	Status           entity.StaffStatus `json:"status" binding:"omitempty,oneof=active inactive on_leave suspended"`
	HireDate         *time.Time  `json:"hire_date"`
	Salary           *float64    `json:"salary"`
	Address          *string     `json:"address"`
	EmergencyContact *string     `json:"emergency_contact"`
	Notes            *string     `json:"notes"`
	Skills           []string    `json:"skills"`
	WorkingHours     *WorkingHours `json:"working_hours"`
	IsAvailable      bool        `json:"is_available"`
}

// UpdateStaffRequest 更新员工请求
type UpdateStaffRequest struct {
	UserID           *string     `json:"user_id" binding:"omitempty,uuid"`
	Name             *string     `json:"name" binding:"omitempty,min=1,max=100"`
	Email            *string     `json:"email" binding:"omitempty,email,max=255"`
	Phone            *string     `json:"phone" binding:"omitempty,max=20"`
	Gender           *entity.Gender `json:"gender" binding:"omitempty,oneof=male female other"`
	Birthday         *time.Time  `json:"birthday"`
	Avatar           *string     `json:"avatar"`
	Department       *string     `json:"department" binding:"omitempty,max=100"`
	Position         *string     `json:"position" binding:"omitempty,max=100"`
	RoleID           *string     `json:"role_id" binding:"omitempty,uuid"`
	Status           *entity.StaffStatus `json:"status" binding:"omitempty,oneof=active inactive on_leave suspended"`
	HireDate         *time.Time  `json:"hire_date"`
	Salary           *float64    `json:"salary"`
	Address          *string     `json:"address"`
	EmergencyContact *string     `json:"emergency_contact"`
	Notes            *string     `json:"notes"`
	Skills           []string    `json:"skills"`
	WorkingHours     *WorkingHours `json:"working_hours"`
	IsAvailable      *bool       `json:"is_available"`
}

// UpdateStatusRequest 更新状态请求
type UpdateStatusRequest struct {
	Status entity.StaffStatus `json:"status" binding:"required,oneof=active inactive on_leave suspended"`
	Reason *string            `json:"reason"`
}

// WorkingHours 工作时间安排
type WorkingHours struct {
	Monday    *DaySchedule `json:"monday"`
	Tuesday   *DaySchedule `json:"tuesday"`
	Wednesday *DaySchedule `json:"wednesday"`
	Thursday  *DaySchedule `json:"thursday"`
	Friday    *DaySchedule `json:"friday"`
	Saturday  *DaySchedule `json:"saturday"`
	Sunday    *DaySchedule `json:"sunday"`
}

// DaySchedule 日程安排
type DaySchedule struct {
	Enabled    bool   `json:"enabled"`
	StartTime  string `json:"start_time"`  // HH:MM
	EndTime    string `json:"end_time"`    // HH:MM
	BreakStart string `json:"break_start"` // HH:MM
	BreakEnd   string `json:"break_end"`   // HH:MM
	Notes      *string `json:"notes"`
}

// RoleFilter 角色查询过滤器
type RoleFilter struct {
	Search  *string             `json:"search"`
	Status  *entity.StaffStatus `json:"status"`
	Page    int                 `json:"page"`
	Limit   int                 `json:"limit"`
	Sort    string              `json:"sort"`
	Order   string              `json:"order"`
}

// SetDefaults 设置默认值
func (f *RoleFilter) SetDefaults() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Sort == "" {
		f.Sort = "sort_order"
	}
	if f.Order == "" {
		f.Order = "asc"
	}
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required,min=1,max=100"`
	Code        string   `json:"code" binding:"required,min=1,max=50"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
	IsDefault   bool     `json:"is_default"`
	SortOrder   int      `json:"sort_order"`
	Status      entity.StaffStatus `json:"status" binding:"omitempty,oneof=active inactive"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        *string  `json:"name" binding:"omitempty,min=1,max=100"`
	Code        *string  `json:"code" binding:"omitempty,min=1,max=50"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
	IsDefault   *bool    `json:"is_default"`
	SortOrder   *int     `json:"sort_order"`
	Status      *entity.StaffStatus `json:"status" binding:"omitempty,oneof=active inactive"`
}

// AvailabilityRequest 可用性请求
type AvailabilityRequest struct {
	StaffID     string                  `json:"staff_id" binding:"required,uuid"`
	Availabilities []AvailabilityItem    `json:"availabilities" binding:"required,min=1,dive"`
}

// AvailabilityItem 可用性项
type AvailabilityItem struct {
	DayOfWeek    int     `json:"day_of_week" binding:"required,min=0,max=6"` // 0-6 (Sunday-Saturday)
	StartTime    string  `json:"start_time" binding:"required"` // HH:MM
	EndTime      string  `json:"end_time" binding:"required"`   // HH:MM
	IsAvailable  bool    `json:"is_available"`
	Notes        *string `json:"notes"`
}

// AvailabilityResponse 可用性响应
type AvailabilityResponse struct {
	StaffID       string                  `json:"staff_id"`
	Availabilities []entity.StaffAvailability `json:"availabilities"`
}

// StaffStats 员工统计
type StaffStats struct {
	TotalStaff      int            `json:"total_staff"`
	ActiveStaff     int            `json:"active_staff"`
	InactiveStaff   int            `json:"inactive_staff"`
	OnLeaveStaff    int            `json:"on_leave_staff"`
	SuspendedStaff  int            `json:"suspended_staff"`
	AvailableStaff  int            `json:"available_staff"`
	NewHiresThisMonth int          `json:"new_hires_this_month"`
	AverageSalary   float64        `json:"average_salary"`
	DepartmentStats []DepartmentStats `json:"department_stats"`
	RoleStats       []RoleStats     `json:"role_stats"`
}

// DepartmentStats 部门统计
type DepartmentStats struct {
	Department  string  `json:"department"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
	AverageSalary float64 `json:"average_salary"`
}

// RoleStats 角色统计
type RoleStats struct {
	RoleID      string  `json:"role_id"`
	RoleName    string  `json:"role_name"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
}

// StaffResponse 员工响应
type StaffResponse struct {
	ID               uuid.UUID      `json:"id"`
	UserID           *uuid.UUID     `json:"user_id"`
	Name             string         `json:"name"`
	Email            string         `json:"email"`
	Phone            string         `json:"phone"`
	Gender           entity.Gender  `json:"gender"`
	Birthday         *time.Time     `json:"birthday"`
	Age              *int           `json:"age"`
	Avatar           *string        `json:"avatar"`
	Department       string         `json:"department"`
	Position         string         `json:"position"`
	RoleID           uuid.UUID      `json:"role_id"`
	RoleName         *string        `json:"role_name"` // 从角色表获取
	Status           entity.StaffStatus `json:"status"`
	HireDate         *time.Time     `json:"hire_date"`
	WorkingYears     *int           `json:"working_years"`
	Salary           *float64       `json:"salary"`
	Address          *string        `json:"address"`
	EmergencyContact *string        `json:"emergency_contact"`
	Notes            *string        `json:"notes"`
	Skills           []string       `json:"skills"`
	WorkingHours     *WorkingHours  `json:"working_hours"`
	IsAvailable      bool           `json:"is_available"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Code        string           `json:"code"`
	Description *string          `json:"description"`
	Permissions []string         `json:"permissions"`
	IsDefault   bool             `json:"is_default"`
	SortOrder   int              `json:"sort_order"`
	Status      entity.StaffStatus `json:"status"`
	StaffCount  int              `json:"staff_count"` // 该角色的员工数量
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// AvailableStaffRequest 可用员工查询请求
type AvailableStaffRequest struct {
	DateTime    time.Time `json:"date_time" binding:"required"`
	Skills      []string  `json:"skills"`
	Department  *string   `json:"department"`
}

// AvailableStaffResponse 可用员工响应
type AvailableStaffResponse struct {
	DateTime     time.Time         `json:"date_time"`
	AvailableStaff []*StaffResponse `json:"available_staff"`
	TotalCount   int               `json:"total_count"`
}

// BatchUpdateRequest 批量更新请求
type BatchUpdateRequest struct {
	StaffIDs []string              `json:"staff_ids" binding:"required,min=1"`
	Updates  *UpdateStaffRequest   `json:"updates" binding:"required"`
}

// BatchStatusUpdateRequest 批量状态更新请求
type BatchStatusUpdateRequest struct {
	StaffIDs []string            `json:"staff_ids" binding:"required,min=1"`
	Status   entity.StaffStatus  `json:"status" binding:"required,oneof=active inactive on_leave suspended"`
	Reason   *string             `json:"reason"`
}