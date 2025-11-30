package repository

import (
	"fmt"
	"time"

	"github.com/julesChu12/fly/appointments/internal/domain/appointment"
	"gorm.io/gorm"
)

// appointmentRepository 预约仓储实现
type appointmentRepository struct {
	db *gorm.DB
}

// NewAppointmentRepository 创建预约仓储实例
func NewAppointmentRepository(db *gorm.DB) appointment.Repository {
	return &appointmentRepository{db: db}
}

// Create 创建预约
func (r *appointmentRepository) Create(appointment *appointment.Appointment) error {
	if err := r.db.Create(appointment).Error; err != nil {
		return fmt.Errorf("创建预约失败: %w", err)
	}
	return nil
}

// GetByID 根据ID获取预约
func (r *appointmentRepository) GetByID(id string) (*appointment.Appointment, error) {
	var appointment appointment.Appointment
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&appointment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("预约不存在")
		}
		return nil, fmt.Errorf("获取预约失败: %w", err)
	}
	return &appointment, nil
}

// Update 更新预约
func (r *appointmentRepository) Update(appointment *appointment.Appointment) error {
	if err := r.db.Save(appointment).Error; err != nil {
		return fmt.Errorf("更新预约失败: %w", err)
	}
	return nil
}

// Delete 删除预约
func (r *appointmentRepository) Delete(id string) error {
	if err := r.db.Unscoped().Where("id = ?", id).Delete(&appointment.Appointment{}).Error; err != nil {
		return fmt.Errorf("删除预约失败: %w", err)
	}
	return nil
}

// SoftDelete 软删除预约
func (r *appointmentRepository) SoftDelete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&appointment.Appointment{}).Error; err != nil {
		return fmt.Errorf("软删除预约失败: %w", err)
	}
	return nil
}

// List 获取预约列表
func (r *appointmentRepository) List(filter *appointment.Filter) ([]*appointment.Appointment, error) {
	var appointments []*appointment.Appointment
	query := r.buildQuery(filter)

	if err := query.Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("获取预约列表失败: %w", err)
	}

	return appointments, nil
}

// Count 统计预约数量
func (r *appointmentRepository) Count(filter *appointment.Filter) (int64, error) {
	var count int64
	query := r.buildQuery(filter)

	if err := query.Model(&appointment.Appointment{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计预约数量失败: %w", err)
	}

	return count, nil
}

// GetByCustomerID 根据客户ID获取预约
func (r *appointmentRepository) GetByCustomerID(customerID string, filter *appointment.Filter) ([]*appointment.Appointment, error) {
	var appointments []*appointment.Appointment
	query := r.buildQuery(filter)
	query = query.Where("customer_id = ?", customerID)

	if err := query.Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("获取客户预约失败: %w", err)
	}

	return appointments, nil
}

// GetByStaffID 根据员工ID获取预约
func (r *appointmentRepository) GetByStaffID(staffID string, filter *appointment.Filter) ([]*appointment.Appointment, error) {
	var appointments []*appointment.Appointment
	query := r.buildQuery(filter)
	query = query.Where("staff_id = ?", staffID)

	if err := query.Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("获取员工预约失败: %w", err)
	}

	return appointments, nil
}

// GetByDateRange 根据日期范围获取预约
func (r *appointmentRepository) GetByDateRange(startDate, endDate time.Time, filter *appointment.Filter) ([]*appointment.Appointment, error) {
	var appointments []*appointment.Appointment
	query := r.buildQuery(filter)
	query = query.Where("start_time >= ? AND end_time <= ?", startDate, endDate)

	if err := query.Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("获取日期范围预约失败: %w", err)
	}

	return appointments, nil
}

// CheckConflict 检查时间冲突
func (r *appointmentRepository) CheckConflict(staffID string, startTime, endTime time.Time, excludeID *string) ([]*appointment.Appointment, error) {
	var conflicts []*appointment.Appointment
	query := r.db.Where("staff_id = ? AND deleted_at IS NULL", staffID).
		Where("((start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?))",
			startTime, startTime, endTime, endTime, startTime, endTime)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	if err := query.Find(&conflicts).Error; err != nil {
		return nil, fmt.Errorf("检查时间冲突失败: %w", err)
	}

	return conflicts, nil
}

// GetAvailableSlots 获取可用时间段
func (r *appointmentRepository) GetAvailableSlots(staffID string, date time.Time, serviceDuration time.Duration) ([]*time.Time, error) {
	// 获取当天所有预约
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var appointments []*appointment.Appointment
	if err := r.db.Where("staff_id = ? AND start_time >= ? AND end_time <= ? AND deleted_at IS NULL",
		staffID, startOfDay, endOfDay).
		Order("start_time ASC").
		Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("获取预约信息失败: %w", err)
	}

	// 生成可用时间段（简化实现）
	var slots []*time.Time
	businessStart := startOfDay.Add(9 * time.Hour) // 9:00
	businessEnd := startOfDay.Add(18 * time.Hour)  // 18:00

	current := businessStart
	for current.Add(serviceDuration).Before(businessEnd) || current.Add(serviceDuration).Equal(businessEnd) {
		available := true
		for _, apt := range appointments {
			if apt.Status == appointment.AppointmentStatusCancelled {
				continue
			}
			// 检查时间重叠
			slotEnd := current.Add(serviceDuration)
			if current.Before(apt.EndTime) && slotEnd.After(apt.StartTime) {
				available = false
				break
			}
		}

		if available {
			slotCopy := current
			slots = append(slots, &slotCopy)
		}

		current = current.Add(30 * time.Minute) // 30分钟间隔
	}

	return slots, nil
}

// GetByStatus 根据状态获取预约
func (r *appointmentRepository) GetByStatus(status appointment.AppointmentStatus, filter *appointment.Filter) ([]*appointment.Appointment, error) {
	var appointments []*appointment.Appointment
	query := r.buildQuery(filter)
	query = query.Where("status = ?", status)

	if err := query.Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("获取状态预约失败: %w", err)
	}

	return appointments, nil
}

// UpdateStatus 更新预约状态
func (r *appointmentRepository) UpdateStatus(id string, status appointment.AppointmentStatus) error {
	if err := r.db.Model(&appointment.Appointment{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("更新预约状态失败: %w", err)
	}
	return nil
}

// GetPendingReminders 获取待处理的提醒
func (r *appointmentRepository) GetPendingReminders(beforeTime time.Time) ([]*appointment.Appointment, error) {
	var appointments []*appointment.Appointment
	if err := r.db.Where("reminder = ? AND reminder_time <= ? AND start_time > ? AND status IN ? AND deleted_at IS NULL",
		true, beforeTime, time.Now(), []appointment.AppointmentStatus{
			appointment.AppointmentStatusPending,
			appointment.AppointmentStatusConfirmed,
		}).
		Find(&appointments).Error; err != nil {
		return nil, fmt.Errorf("获取待处理提醒失败: %w", err)
	}

	return appointments, nil
}

// buildQuery 构建查询条件
func (r *appointmentRepository) buildQuery(filter *appointment.Filter) *gorm.DB {
	query := r.db.Model(&appointment.Appointment{}).Where("deleted_at IS NULL")

	if filter == nil {
		return query
	}

	if filter.CustomerID != nil {
		query = query.Where("customer_id = ?", *filter.CustomerID)
	}
	if filter.StaffID != nil {
		query = query.Where("staff_id = ?", *filter.StaffID)
	}
	if filter.ServiceID != nil {
		query = query.Where("service_id = ?", *filter.ServiceID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.StartDate != nil {
		query = query.Where("start_time >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("end_time <= ?", *filter.EndDate)
	}
	if filter.Date != nil {
		startOfDay := time.Date(filter.Date.Year(), filter.Date.Month(), filter.Date.Day(), 0, 0, 0, 0, filter.Date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		query = query.Where("start_time >= ? AND start_time < ?", startOfDay, endOfDay)
	}
	if filter.MinDuration != nil {
		query = query.Where("TIMESTAMPDIFF(SECOND, start_time, end_time) >= ?", int(filter.MinDuration.Seconds()))
	}
	if filter.MaxDuration != nil {
		query = query.Where("TIMESTAMPDIFF(SECOND, start_time, end_time) <= ?", int(filter.MaxDuration.Seconds()))
	}
	if filter.Reminder != nil {
		query = query.Where("reminder = ?", *filter.Reminder)
	}

	// 排序
	if filter.Sort != "" {
		order := "ASC"
		if filter.Order == "desc" {
			order = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", filter.Sort, order))
	}

	// 分页
	if filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	return query
}
