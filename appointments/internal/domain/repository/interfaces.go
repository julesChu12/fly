package repository

import (
	"time"

	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
)

// AppointmentRepository 预约仓储接口
type AppointmentRepository interface {
	// 基础CRUD操作
	Create(appointment *entity.Appointment) error
	GetByID(id string) (*entity.Appointment, error)
	Update(appointment *entity.Appointment) error
	Delete(id string) error
	SoftDelete(id string) error

	// 查询操作
	List(filter *dto.AppointmentFilter) ([]*entity.Appointment, error)
	Count(filter *dto.AppointmentFilter) (int64, error)

	// 业务相关查询
	GetByCustomerID(customerID string, filter *dto.AppointmentFilter) ([]*entity.Appointment, error)
	GetByStaffID(staffID string, filter *dto.AppointmentFilter) ([]*entity.Appointment, error)
	GetByDateRange(startDate, endDate time.Time, filter *dto.AppointmentFilter) ([]*entity.Appointment, error)

	// 冲突检测
	CheckConflict(staffID string, startTime, endTime time.Time, excludeID *string) ([]*entity.Appointment, error)
	GetAvailableSlots(staffID string, date time.Time, serviceDuration time.Duration) ([]*time.Time, error)

	// 状态相关
	GetByStatus(status entity.AppointmentStatus, filter *dto.AppointmentFilter) ([]*entity.Appointment, error)
	UpdateStatus(id string, status entity.AppointmentStatus) error

	// 提醒相关
	GetPendingReminders(beforeTime time.Time) ([]*entity.Appointment, error)
}

