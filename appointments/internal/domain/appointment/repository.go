package appointment

import (
	"time"
)

// Filter 预约查询过滤器
type Filter struct {
	CustomerID  *string
	StaffID     *string
	ServiceID   *string
	Status      *AppointmentStatus
	StartDate   *time.Time
	EndDate     *time.Time
	Date        *time.Time // 特定日期查询
	MinDuration *time.Duration
	MaxDuration *time.Duration
	Reminder    *bool
	Page        int
	Limit       int
	Sort        string // 排序字段
	Order       string // 排序方向: asc/desc
}

// Repository 预约仓储接口
type Repository interface {
	// 基础CRUD操作
	Create(appointment *Appointment) error
	GetByID(id string) (*Appointment, error)
	Update(appointment *Appointment) error
	Delete(id string) error
	SoftDelete(id string) error

	// 查询操作
	List(filter *Filter) ([]*Appointment, error)
	Count(filter *Filter) (int64, error)

	// 业务相关查询
	GetByCustomerID(customerID string, filter *Filter) ([]*Appointment, error)
	GetByStaffID(staffID string, filter *Filter) ([]*Appointment, error)
	GetByDateRange(startDate, endDate time.Time, filter *Filter) ([]*Appointment, error)

	// 冲突检测
	CheckConflict(staffID string, startTime, endTime time.Time, excludeID *string) ([]*Appointment, error)
	GetAvailableSlots(staffID string, date time.Time, serviceDuration time.Duration) ([]*time.Time, error)

	// 状态相关
	GetByStatus(status AppointmentStatus, filter *Filter) ([]*Appointment, error)
	UpdateStatus(id string, status AppointmentStatus) error

	// 提醒相关
	GetPendingReminders(beforeTime time.Time) ([]*Appointment, error)
}
