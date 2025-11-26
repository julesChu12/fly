package entity

import (
	"time"

	"github.com/google/uuid"
)

// AppointmentStatus 预约状态枚举
type AppointmentStatus string

const (
	AppointmentStatusPending     AppointmentStatus = "pending"      // 待确认
	AppointmentStatusConfirmed   AppointmentStatus = "confirmed"    // 已确认
	AppointmentStatusInProgress  AppointmentStatus = "in_progress"  // 进行中
	AppointmentStatusCompleted   AppointmentStatus = "completed"    // 已完成
	AppointmentStatusCancelled   AppointmentStatus = "cancelled"    // 已取消
	AppointmentStatusNoShow      AppointmentStatus = "no_show"      // 未到店
)

// Appointment 预约实体
type Appointment struct {
	ID           uuid.UUID         `json:"id" gorm:"type:char(36);primaryKey;uniqueIndex"`
	CustomerID   uuid.UUID         `json:"customer_id" gorm:"type:char(36);not null;index"`
	EmployeeID   uuid.UUID         `json:"employee_id" gorm:"type:char(36);not null;index"`
	ServiceID    uuid.UUID         `json:"service_id" gorm:"type:char(36);not null;index"`
	StartTime    time.Time         `json:"start_time" gorm:"not null;index"`
	EndTime      time.Time         `json:"end_time" gorm:"not null"`
	Notes        *string           `json:"notes" gorm:"type:text"`
	Status       AppointmentStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	Reminder     bool              `json:"reminder" gorm:"default:true"`
	ReminderTime *time.Time        `json:"reminder_time" gorm:"index"`
	CreatedAt    time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time        `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Appointment) TableName() string {
	return "appointments"
}

// IsValidStatus 检查状态是否有效
func (a *Appointment) IsValidStatus(status AppointmentStatus) bool {
	switch status {
	case AppointmentStatusPending, AppointmentStatusConfirmed, AppointmentStatusInProgress,
		AppointmentStatusCompleted, AppointmentStatusCancelled, AppointmentStatusNoShow:
		return true
	default:
		return false
	}
}

// CanTransitionTo 检查状态转换是否合法
func (a *Appointment) CanTransitionTo(newStatus AppointmentStatus) bool {
	switch a.Status {
	case AppointmentStatusPending:
		return newStatus == AppointmentStatusConfirmed || newStatus == AppointmentStatusCancelled
	case AppointmentStatusConfirmed:
		return newStatus == AppointmentStatusInProgress || newStatus == AppointmentStatusCancelled
	case AppointmentStatusInProgress:
		return newStatus == AppointmentStatusCompleted || newStatus == AppointmentStatusCancelled
	case AppointmentStatusCompleted:
		return false // 完成状态不能转换
	case AppointmentStatusCancelled:
		return false // 取消状态不能转换
	case AppointmentStatusNoShow:
		return false // 未到店状态不能转换
	default:
		return false
	}
}

// Duration 返回预约时长
func (a *Appointment) Duration() time.Duration {
	return a.EndTime.Sub(a.StartTime)
}

// IsInPast 检查预约是否已过期
func (a *Appointment) IsInPast() bool {
	return time.Now().After(a.EndTime)
}

// IsInProgress 检查预约是否正在进行中
func (a *Appointment) IsInProgress() bool {
	now := time.Now()
	return now.After(a.StartTime) && now.Before(a.EndTime)
}

// NeedsReminder 检查是否需要提醒
func (a *Appointment) NeedsReminder() bool {
	if !a.Reminder || a.ReminderTime == nil {
		return false
	}
	now := time.Now()
	return now.After(*a.ReminderTime) && now.Before(a.StartTime)
}