package entity

import (
	"time"

	"github.com/google/uuid"
)

// StaffStatus 员工状态枚举
type StaffStatus string

const (
	StaffStatusActive   StaffStatus = "active"    // 在职
	StaffStatusInactive StaffStatus = "inactive"  // 离职
	StaffStatusOnLeave  StaffStatus = "on_leave"  // 休假
	StaffStatusSuspended StaffStatus = "suspended" // 停职
)

// Gender 性别枚举
type Gender string

const (
	GenderMale   Gender = "male"   // 男
	GenderFemale Gender = "female" // 女
	GenderOther  Gender = "other"  // 其他
)

// Staff 员工实体
type Staff struct {
	ID           uuid.UUID  `json:"id" gorm:"type:char(36);primaryKey;uniqueIndex"`
	UserID       *uuid.UUID `json:"user_id" gorm:"type:char(36);index"` // 关联Custos用户ID
	Name         string     `json:"name" gorm:"size:100;not null"`
	Email        string     `json:"email" gorm:"size:255;not null;uniqueIndex"`
	Phone        string     `json:"phone" gorm:"size:20"`
	Gender       Gender     `json:"gender" gorm:"type:varchar(10);default:'other'"`
	Birthday     *time.Time `json:"birthday"`
	Avatar       *string    `json:"avatar" gorm:"size:500"`
	Department   string     `json:"department" gorm:"size:100"`
	Position     string     `json:"position" gorm:"size:100"`
	RoleID       uuid.UUID  `json:"role_id" gorm:"type:char(36);not null;index"`
	Status       StaffStatus `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	HireDate     *time.Time `json:"hire_date"`
	Salary       *float64   `json:"salary" gorm:"type:decimal(10,2)"`
	Address      *string    `json:"address" gorm:"type:text"`
	EmergencyContact *string `json:"emergency_contact" gorm:"type:text"`
	Notes        *string    `json:"notes" gorm:"type:text"`
	Skills       string     `json:"skills" gorm:"type:text"` // 技能标签，JSON格式存储
	WorkingHours *string    `json:"working_hours" gorm:"type:text"` // 工作时间安排，JSON格式
	IsAvailable  bool       `json:"is_available" gorm:"default:true"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Staff) TableName() string {
	return "staff"
}

// IsValidStatus 检查状态是否有效
func (s *Staff) IsValidStatus(status StaffStatus) bool {
	switch status {
	case StaffStatusActive, StaffStatusInactive, StaffStatusOnLeave, StaffStatusSuspended:
		return true
	default:
		return false
	}
}

// CanTransitionTo 检查状态转换是否合法
func (s *Staff) CanTransitionTo(newStatus StaffStatus) bool {
	switch s.Status {
	case StaffStatusActive:
		return newStatus == StaffStatusInactive || newStatus == StaffStatusOnLeave || newStatus == StaffStatusSuspended
	case StaffStatusInactive:
		return newStatus == StaffStatusActive // 离职后可以重新激活
	case StaffStatusOnLeave:
		return newStatus == StaffStatusActive || newStatus == StaffStatusInactive
	case StaffStatusSuspended:
		return newStatus == StaffStatusActive || newStatus == StaffStatusInactive
	default:
		return false
	}
}

// IsActive 检查员工是否在职可用
func (s *Staff) IsActive() bool {
	return s.Status == StaffStatusActive && s.IsAvailable
}

// GetAge 计算年龄
func (s *Staff) GetAge() *int {
	if s.Birthday == nil {
		return nil
	}
	age := int(time.Since(*s.Birthday).Hours() / 24 / 365)
	return &age
}

// WorkingYears 计算工作年限
func (s *Staff) WorkingYears() *int {
	if s.HireDate == nil {
		return nil
	}
	years := int(time.Since(*s.HireDate).Hours() / 24 / 365)
	return &years
}

// StaffRole 员工角色实体
type StaffRole struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primaryKey;uniqueIndex"`
	Name        string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Code        string    `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Description *string   `json:"description" gorm:"type:text"`
	Permissions string    `json:"permissions" gorm:"type:text"` // 权限列表，JSON格式
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	Status      StaffStatus `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"-" gorm:"index"`
}

// TableName 指定表名
func (StaffRole) TableName() string {
	return "staff_roles"
}

// StaffAvailability 员工可用性设置
type StaffAvailability struct {
	ID         uuid.UUID `json:"id" gorm:"type:char(36);primaryKey;uniqueIndex"`
	StaffID    uuid.UUID `json:"staff_id" gorm:"type:char(36);not null;index"`
	DayOfWeek  int       `json:"day_of_week" gorm:"not null"` // 0-6 (Sunday-Saturday)
	StartTime  string    `json:"start_time" gorm:"size:5;not null"` // HH:MM
	EndTime    string    `json:"end_time" gorm:"size:5;not null"`   // HH:MM
	IsAvailable bool     `json:"is_available" gorm:"default:true"`
	Notes      *string   `json:"notes" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (StaffAvailability) TableName() string {
	return "staff_availability"
}

// StaffStats 员工统计信息
type StaffStats struct {
	TotalStaff      int64            `json:"total_staff"`
	ActiveStaff     int64            `json:"active_staff"`
	AvailableStaff  int64            `json:"available_staff"`
	ByDepartment    map[string]int64 `json:"by_department"`
	ByRole          map[string]int64 `json:"by_role"`
}