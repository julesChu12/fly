package database

import (
	"time"

	"github.com/julesChu12/fly/staff/internal/domain/entity"
	"github.com/julesChu12/fly/staff/internal/domain/repository"
	"gorm.io/gorm"
)

type staffAvailabilityRepository struct {
	db *gorm.DB
}

// NewStaffAvailabilityRepository 创建员工可用性仓储实例
func NewStaffAvailabilityRepository(db *gorm.DB) repository.StaffAvailabilityRepository {
	return &staffAvailabilityRepository{
		db: db,
	}
}

// Create 创建可用性记录
func (r *staffAvailabilityRepository) Create(availability *entity.StaffAvailability) error {
	return r.db.Create(availability).Error
}

// GetByID 根据ID获取可用性记录
func (r *staffAvailabilityRepository) GetByID(id string) (*entity.StaffAvailability, error) {
	var availability entity.StaffAvailability
	err := r.db.First(&availability, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &availability, nil
}

// Update 更新可用性记录
func (r *staffAvailabilityRepository) Update(availability *entity.StaffAvailability) error {
	return r.db.Save(availability).Error
}

// Delete 删除可用性记录
func (r *staffAvailabilityRepository) Delete(id string) error {
	return r.db.Delete(&entity.StaffAvailability{}, "id = ?", id).Error
}

// GetByStaffID 根据员工ID获取可用性记录
func (r *staffAvailabilityRepository) GetByStaffID(staffID string) ([]*entity.StaffAvailability, error) {
	var availabilities []*entity.StaffAvailability
	err := r.db.Where("staff_id = ?", staffID).
		Order("day_of_week, start_time").
		Find(&availabilities).Error
	if err != nil {
		return nil, err
	}
	return availabilities, nil
}

// GetByStaffIDAndDay 根据员工ID和星期几获取可用性记录
func (r *staffAvailabilityRepository) GetByStaffIDAndDay(staffID string, dayOfWeek int) (*entity.StaffAvailability, error) {
	var availability entity.StaffAvailability
	err := r.db.Where("staff_id = ? AND day_of_week = ?", staffID, dayOfWeek).
		First(&availability).Error
	if err != nil {
		return nil, err
	}
	return &availability, nil
}

// GetAvailableStaff 获取指定时间的可用员工
func (r *staffAvailabilityRepository) GetAvailableStaff(dateTime time.Time) ([]*entity.Staff, error) {
	// 简化实现，返回空切片
	return []*entity.Staff{}, nil
}

// BatchCreate 批量创建可用性记录
func (r *staffAvailabilityRepository) BatchCreate(availabilities []*entity.StaffAvailability) error {
	return r.db.CreateInBatches(availabilities, 100).Error
}

// BatchUpdate 批量更新可用性记录
func (r *staffAvailabilityRepository) BatchUpdate(availabilities []*entity.StaffAvailability) error {
	for _, availability := range availabilities {
		err := r.db.Save(availability).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteByStaffID 删除员工的所有可用性记录
func (r *staffAvailabilityRepository) DeleteByStaffID(staffID string) error {
	return r.db.Where("staff_id = ?", staffID).Delete(&entity.StaffAvailability{}).Error
}