package database

import (
	"github.com/julesChu12/fly/staff/internal/domain/entity"
	"github.com/julesChu12/fly/staff/internal/domain/repository"
	"github.com/julesChu12/fly/staff/internal/domain/dto"
	"gorm.io/gorm"
)

type staffRoleRepository struct {
	db *gorm.DB
}

// NewStaffRoleRepository 创建角色仓储实例
func NewStaffRoleRepository(db *gorm.DB) repository.StaffRoleRepository {
	return &staffRoleRepository{
		db: db,
	}
}

// Create 创建角色
func (r *staffRoleRepository) Create(role *entity.StaffRole) error {
	return r.db.Create(role).Error
}

// GetByID 根据ID获取角色
func (r *staffRoleRepository) GetByID(id string) (*entity.StaffRole, error) {
	var role entity.StaffRole
	err := r.db.First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// Update 更新角色
func (r *staffRoleRepository) Update(role *entity.StaffRole) error {
	return r.db.Save(role).Error
}

// Delete 删除角色
func (r *staffRoleRepository) Delete(id string) error {
	return r.db.Delete(&entity.StaffRole{}, "id = ?", id).Error
}

// SoftDelete 软删除角色
func (r *staffRoleRepository) SoftDelete(id string) error {
	return r.db.Delete(&entity.StaffRole{}, "id = ?", id).Error
}

// Count 获取角色总数
func (r *staffRoleRepository) Count(filter *dto.RoleFilter) (int64, error) {
	query := r.db.Model(&entity.StaffRole{})

	if filter.Search != nil && *filter.Search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?",
			"%"+*filter.Search+"%", "%"+*filter.Search+"%")
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// List 获取角色列表
func (r *staffRoleRepository) List(filter *dto.RoleFilter) ([]*entity.StaffRole, error) {
	var roles []*entity.StaffRole

	query := r.db.Model(&entity.StaffRole{})

	// 应用过滤条件
	if filter.Search != nil && *filter.Search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?",
			"%"+*filter.Search+"%", "%"+*filter.Search+"%")
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// 排序
	query = query.Order("created_at desc")

	// 分页
	if filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	err := query.Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// GetByCode 根据代码获取角色
func (r *staffRoleRepository) GetByCode(code string) (*entity.StaffRole, error) {
	var role entity.StaffRole
	err := r.db.First(&role, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByName 根据名称获取角色
func (r *staffRoleRepository) GetByName(name string) (*entity.StaffRole, error) {
	var role entity.StaffRole
	err := r.db.First(&role, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetDefaultRole 获取默认角色
func (r *staffRoleRepository) GetDefaultRole() (*entity.StaffRole, error) {
	var role entity.StaffRole
	err := r.db.First(&role, "is_default = ?", true).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetActiveRoles 获取活跃角色
func (r *staffRoleRepository) GetActiveRoles() ([]*entity.StaffRole, error) {
	var roles []*entity.StaffRole
	err := r.db.Where("status = ?", entity.StaffStatusActive).
		Order("sort_order asc, name asc").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdateStatus 更新角色状态
func (r *staffRoleRepository) UpdateStatus(id string, status entity.StaffStatus) error {
	return r.db.Model(&entity.StaffRole{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// ExistByName 检查角色名称是否存在
func (r *staffRoleRepository) ExistByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.StaffRole{}).
		Where("name = ?", name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}