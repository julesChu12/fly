package dto

import (
	"time"

	"github.com/julesChu12/fly/appointments/internal/domain/appointment"
)

// AppointmentFilter 预约查询过滤器
type AppointmentFilter struct {
	CustomerID  *string                        `json:"customer_id"`
	StaffID     *string                        `json:"staff_id"`
	ServiceID   *string                        `json:"service_id"`
	Status      *appointment.AppointmentStatus `json:"status"`
	StartDate   *time.Time                     `json:"start_date"`
	EndDate     *time.Time                     `json:"end_date"`
	Date        *time.Time                     `json:"date"` // 特定日期查询
	MinDuration *time.Duration                 `json:"min_duration"`
	MaxDuration *time.Duration                 `json:"max_duration"`
	Reminder    *bool                          `json:"reminder"`
	Page        int                            `json:"page"`
	Limit       int                            `json:"limit"`
	Sort        string                         `json:"sort"`  // 排序字段
	Order       string                         `json:"order"` // 排序方向: asc/desc
}

// SetDefaults 设置默认值
func (f *AppointmentFilter) SetDefaults() {
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
		f.Sort = "start_time"
	}
	if f.Order == "" {
		f.Order = "asc"
	}
}

// CreateAppointmentRequest 创建预约请求
type CreateAppointmentRequest struct {
	CustomerID   string     `json:"customer_id" binding:"required,uuid"`
	StaffID      string     `json:"staff_id" binding:"required,uuid"`
	ServiceID    string     `json:"service_id" binding:"required,uuid"`
	StartTime    time.Time  `json:"start_time" binding:"required"`
	EndTime      time.Time  `json:"end_time" binding:"required,gtfield=StartTime"`
	Notes        *string    `json:"notes"`
	Reminder     bool       `json:"reminder"`
	ReminderTime *time.Time `json:"reminder_time"`
}

// UpdateAppointmentRequest 更新预约请求
type UpdateAppointmentRequest struct {
	CustomerID   *string    `json:"customer_id" binding:"omitempty,uuid"`
	StaffID      *string    `json:"staff_id" binding:"omitempty,uuid"`
	ServiceID    *string    `json:"service_id" binding:"omitempty,uuid"`
	StartTime    *time.Time `json:"start_time" binding:"omitempty"`
	EndTime      *time.Time `json:"end_time" binding:"omitempty,gtfield=StartTime"`
	Notes        *string    `json:"notes"`
	Status       *string    `json:"status" binding:"omitempty,oneof= pending confirmed in_progress completed cancelled no_show"`
	Reminder     *bool      `json:"reminder"`
	ReminderTime *time.Time `json:"reminder_time"`
}

// UpdateStatusRequest 更新状态请求
type UpdateStatusRequest struct {
	Status          string  `json:"status" binding:"required,oneof= pending confirmed in_progress completed cancelled no_show"`
	CompletionNotes *string `json:"completion_notes"`
}

// CalendarEvent 日历事件
type CalendarEvent struct {
	ID           string                        `json:"id"`
	CustomerID   string                        `json:"customer_id"`
	CustomerName string                        `json:"customer_name"` // 从客户服务获取
	StaffID      string                        `json:"staff_id"`
	StaffName    string                        `json:"staff_name"` // 从员工服务获取
	ServiceID    string                        `json:"service_id"`
	ServiceName  string                        `json:"service_name"` // 从产品服务获取
	StartTime    time.Time                     `json:"start_time"`
	EndTime      time.Time                     `json:"end_time"`
	Status       appointment.AppointmentStatus `json:"status"`
	Notes        *string                       `json:"notes"`
	Duration     time.Duration                 `json:"duration"`
}

// AvailabilityRequest 可用时间查询请求
type AvailabilityRequest struct {
	StaffID         string        `json:"staff_id" binding:"required,uuid"`
	Date            time.Time     `json:"date" binding:"required"`
	ServiceDuration time.Duration `json:"service_duration" binding:"required,gt=0"`
}

// AvailableSlot 可用时间段
type AvailableSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Available bool      `json:"available"`
}

// AvailabilityResponse 可用时间响应
type AvailabilityResponse struct {
	StaffID string          `json:"staff_id"`
	Date    time.Time       `json:"date"`
	Slots   []AvailableSlot `json:"slots"`
}

// DailySummary 每日汇总
type DailySummary struct {
	Date                  time.Time `json:"date"`
	TotalAppointments     int       `json:"total_appointments"`
	ConfirmedAppointments int       `json:"confirmed_appointments"`
	CompletedAppointments int       `json:"completed_appointments"`
	CancelledAppointments int       `json:"cancelled_appointments"`
	NoShowAppointments    int       `json:"no_show_appointments"`
	TotalRevenue          float64   `json:"total_revenue"` // 从支付服务获取
}

// AppointmentResponse 预约响应
type AppointmentResponse struct {
	ID           string                        `json:"id"`
	CustomerID   string                        `json:"customer_id"`
	StaffID      string                        `json:"staff_id"`
	ServiceID    string                        `json:"service_id"`
	StartTime    time.Time                     `json:"start_time"`
	EndTime      time.Time                     `json:"end_time"`
	Notes        *string                       `json:"notes"`
	Status       appointment.AppointmentStatus `json:"status"`
	Reminder     bool                          `json:"reminder"`
	ReminderTime *time.Time                    `json:"reminder_time"`
	CreatedAt    time.Time                     `json:"created_at"`
	UpdatedAt    time.Time                     `json:"updated_at"`
}

// CalendarViewRequest 日历视图请求
type CalendarViewRequest struct {
	StaffID   *string   `json:"staff_id" binding:"omitempty,uuid"`
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required,gtfield=StartDate"`
}

// ConflictCheckRequest 冲突检查请求
type ConflictCheckRequest struct {
	StaffID   string    `json:"staff_id" binding:"required,uuid"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
	ExcludeID *string   `json:"exclude_id" binding:"omitempty,uuid"`
}

// ConflictInfo 冲突信息
type ConflictInfo struct {
	Conflict      bool        `json:"conflict"`
	ConflictIDs   []string    `json:"conflict_ids,omitempty"`
	ConflictCount int         `json:"conflict_count"`
	Suggestions   []time.Time `json:"suggestions,omitempty"` // 建议的其他时间
}
