package repository

import (
	"time"

	"github.com/julesChu12/fly/staff/internal/domain/entity"
	"github.com/julesChu12/fly/staff/internal/domain/dto"
)

// StaffRepository 员工仓储接口
type StaffRepository interface {
	// 基础CRUD操作
	Create(staff *entity.Staff) error
	GetByID(id string) (*entity.Staff, error)
	Update(staff *entity.Staff) error
	Delete(id string) error
	SoftDelete(id string) error

	// 查询操作
	List(filter *dto.StaffFilter) ([]*entity.Staff, error)
	Count(filter *dto.StaffFilter) (int64, error)

	// 业务相关查询
	GetByEmail(email string) (*entity.Staff, error)
	GetByPhone(phone string) (*entity.Staff, error)
	GetByUserID(userID string) (*entity.Staff, error)
	GetByRoleID(roleID string, filter *dto.StaffFilter) ([]*entity.Staff, error)
	GetByDepartment(department string, filter *dto.StaffFilter) ([]*entity.Staff, error)
	GetAvailableStaff(filter *dto.StaffFilter) ([]*entity.Staff, error)

	// 状态相关
	UpdateStatus(id string, status entity.StaffStatus) error
	GetByStatus(status entity.StaffStatus, filter *dto.StaffFilter) ([]*entity.Staff, error)

	// 统计相关
	GetStats() (*dto.StaffStats, error)
	GetDepartmentStats() ([]*dto.DepartmentStats, error)
	GetRoleStats() ([]*dto.RoleStats, error)

	// 业务相关
	ExistByEmail(email string) (bool, error)
}

// StaffRoleRepository 员工角色仓储接口
type StaffRoleRepository interface {
	// 基础CRUD操作
	Create(role *entity.StaffRole) error
	GetByID(id string) (*entity.StaffRole, error)
	Update(role *entity.StaffRole) error
	Delete(id string) error
	SoftDelete(id string) error

	// 查询操作
	List(filter *dto.RoleFilter) ([]*entity.StaffRole, error)
	Count(filter *dto.RoleFilter) (int64, error)
	GetByCode(code string) (*entity.StaffRole, error)
	GetByName(name string) (*entity.StaffRole, error)
	GetDefaultRole() (*entity.StaffRole, error)
	GetActiveRoles() ([]*entity.StaffRole, error)

	// 状态相关
	UpdateStatus(id string, status entity.StaffStatus) error

	// 业务相关
	ExistByName(name string) (bool, error)
}

// StaffAvailabilityRepository 员工可用性仓储接口
type StaffAvailabilityRepository interface {
	// 基础CRUD操作
	Create(availability *entity.StaffAvailability) error
	GetByID(id string) (*entity.StaffAvailability, error)
	Update(availability *entity.StaffAvailability) error
	Delete(id string) error

	// 查询操作
	GetByStaffID(staffID string) ([]*entity.StaffAvailability, error)
	GetByStaffIDAndDay(staffID string, dayOfWeek int) (*entity.StaffAvailability, error)
	GetAvailableStaff(dateTime time.Time) ([]*entity.Staff, error)

	// 批量操作
	BatchCreate(availabilities []*entity.StaffAvailability) error
	BatchUpdate(availabilities []*entity.StaffAvailability) error
	DeleteByStaffID(staffID string) error
}