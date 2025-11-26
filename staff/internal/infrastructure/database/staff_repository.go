package database

import (
	"fmt"

	"github.com/julesChu12/fly/staff/internal/domain/entity"
	"github.com/julesChu12/fly/staff/internal/domain/repository"
	"github.com/julesChu12/fly/staff/internal/domain/dto"
	"gorm.io/gorm"
)

type staffRepository struct {
	db *gorm.DB
}

// NewStaffRepository 创建员工仓储实例
func NewStaffRepository(db *gorm.DB) repository.StaffRepository {
	return &staffRepository{
		db: db,
	}
}

// Create 创建员工
func (r *staffRepository) Create(staff *entity.Staff) error {
	return r.db.Create(staff).Error
}

// GetByID 根据ID获取员工
func (r *staffRepository) GetByID(id string) (*entity.Staff, error) {
	var staff entity.Staff
	err := r.db.First(&staff, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// Update 更新员工
func (r *staffRepository) Update(staff *entity.Staff) error {
	return r.db.Save(staff).Error
}

// Delete 删除员工
func (r *staffRepository) Delete(id string) error {
	return r.db.Delete(&entity.Staff{}, "id = ?", id).Error
}

// SoftDelete 软删除员工
func (r *staffRepository) SoftDelete(id string) error {
	return r.db.Delete(&entity.Staff{}, "id = ?", id).Error
}

// Count 获取员工总数
func (r *staffRepository) Count(filter *dto.StaffFilter) (int64, error) {
	query := r.db.Model(&entity.Staff{})

	// 应用过滤条件
	if filter.Search != nil && *filter.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ? OR phone LIKE ?",
			"%"+*filter.Search+"%", "%"+*filter.Search+"%", "%"+*filter.Search+"%")
	}

	if filter.Department != nil && *filter.Department != "" {
		query = query.Where("department = ?", *filter.Department)
	}

	if filter.RoleID != nil && *filter.RoleID != "" {
		query = query.Where("role_id = ?", *filter.RoleID)
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.IsAvailable != nil {
		query = query.Where("is_available = ?", *filter.IsAvailable)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// List 获取员工列表
func (r *staffRepository) List(filter *dto.StaffFilter) ([]*entity.Staff, error) {
	var staff []*entity.Staff

	query := r.db.Model(&entity.Staff{})

	// 应用过滤条件
	if filter.Search != nil && *filter.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ? OR phone LIKE ?",
			"%"+*filter.Search+"%", "%"+*filter.Search+"%", "%"+*filter.Search+"%")
	}

	if filter.Department != nil && *filter.Department != "" {
		query = query.Where("department = ?", *filter.Department)
	}

	if filter.RoleID != nil && *filter.RoleID != "" {
		query = query.Where("role_id = ?", *filter.RoleID)
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.IsAvailable != nil {
		query = query.Where("is_available = ?", *filter.IsAvailable)
	}

	// 排序
	sortField := "created_at"
	if filter.Sort != "" {
		sortField = filter.Sort
	}
	sortOrder := "desc"
	if filter.Order != "" {
		sortOrder = filter.Order
	}
	query = query.Order(fmt.Sprintf("%s %s", sortField, sortOrder))

	// 分页
	if filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	err := query.Find(&staff).Error
	if err != nil {
		return nil, err
	}

	return staff, nil
}

// GetByEmail 根据邮箱获��员工
func (r *staffRepository) GetByEmail(email string) (*entity.Staff, error) {
	var staff entity.Staff
	err := r.db.First(&staff, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetByPhone 根据电话获取员工
func (r *staffRepository) GetByPhone(phone string) (*entity.Staff, error) {
	var staff entity.Staff
	err := r.db.First(&staff, "phone = ?", phone).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetByUserID 根据用户ID获取员工
func (r *staffRepository) GetByUserID(userID string) (*entity.Staff, error) {
	var staff entity.Staff
	err := r.db.First(&staff, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetByRoleID 根据角色ID获取员工
func (r *staffRepository) GetByRoleID(roleID string, filter *dto.StaffFilter) ([]*entity.Staff, error) {
	var staff []*entity.Staff

	query := r.db.Model(&entity.Staff{}).
		Where("role_id = ? AND is_available = ?", roleID, true)

	// 应用其他过滤条件
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	err := query.Find(&staff).Error
	if err != nil {
		return nil, err
	}

	return staff, nil
}

// GetByDepartment 根据部门获取员工
func (r *staffRepository) GetByDepartment(department string, filter *dto.StaffFilter) ([]*entity.Staff, error) {
	var staff []*entity.Staff

	query := r.db.Model(&entity.Staff{}).
		Where("department = ? AND is_available = ?", department, true)

	// 应用其他过滤条件
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	err := query.Find(&staff).Error
	if err != nil {
		return nil, err
	}

	return staff, nil
}

// GetAvailableStaff 获取可用员工
func (r *staffRepository) GetAvailableStaff(filter *dto.StaffFilter) ([]*entity.Staff, error) {
	var staff []*entity.Staff

	query := r.db.Model(&entity.Staff{}).
		Where("is_available = ? AND status = ?", true, entity.StaffStatusActive)

	// 应用其他过滤条件
	if filter.Department != nil && *filter.Department != "" {
		query = query.Where("department = ?", *filter.Department)
	}

	err := query.Find(&staff).Error
	if err != nil {
		return nil, err
	}

	return staff, nil
}

// UpdateStatus 更新员工状态
func (r *staffRepository) UpdateStatus(id string, status entity.StaffStatus) error {
	return r.db.Model(&entity.Staff{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// GetByStatus 根据状态获取员工
func (r *staffRepository) GetByStatus(status entity.StaffStatus, filter *dto.StaffFilter) ([]*entity.Staff, error) {
	var staff []*entity.Staff

	query := r.db.Model(&entity.Staff{}).
		Where("status = ?", status)

	// 应用其他过滤条件
	if filter.Department != nil && *filter.Department != "" {
		query = query.Where("department = ?", *filter.Department)
	}

	if filter.IsAvailable != nil {
		query = query.Where("is_available = ?", *filter.IsAvailable)
	}

	err := query.Find(&staff).Error
	if err != nil {
		return nil, err
	}

	return staff, nil
}

// GetStats 获取员工统计信息
func (r *staffRepository) GetStats() (*dto.StaffStats, error) {
	stats := &dto.StaffStats{}

	// 简化统计实现
	var totalStaff, activeStaff, availableStaff int64
	r.db.Model(&entity.Staff{}).Count(&totalStaff)
	r.db.Model(&entity.Staff{}).Where("status = ?", entity.StaffStatusActive).Count(&activeStaff)
	r.db.Model(&entity.Staff{}).Where("is_available = ?", true).Count(&availableStaff)

	stats.TotalStaff = int(totalStaff)
	stats.ActiveStaff = int(activeStaff)
	stats.AvailableStaff = int(availableStaff)

	return stats, nil
}

// GetDepartmentStats 获取部门统计
func (r *staffRepository) GetDepartmentStats() ([]*dto.DepartmentStats, error) {
	return []*dto.DepartmentStats{}, nil
}

// GetRoleStats 获取角色统计
func (r *staffRepository) GetRoleStats() ([]*dto.RoleStats, error) {
	return []*dto.RoleStats{}, nil
}

// ExistByEmail 检查邮箱是否存在 - 添加这个缺失的方法
func (r *staffRepository) ExistByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Staff{}).
		Where("email = ?", email).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}