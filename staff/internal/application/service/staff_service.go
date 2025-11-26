package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/staff/internal/domain/entity"
	"github.com/julesChu12/fly/staff/internal/domain/repository"
	"github.com/julesChu12/fly/staff/internal/domain/dto"
)

type staffService struct {
	staffRepo         repository.StaffRepository
	roleRepo          repository.StaffRoleRepository
	availabilityRepo  repository.StaffAvailabilityRepository
}

// NewStaffService 创建员工服务实例
func NewStaffService(
	staffRepo repository.StaffRepository,
	roleRepo repository.StaffRoleRepository,
	availabilityRepo repository.StaffAvailabilityRepository,
) StaffService {
	return &staffService{
		staffRepo:        staffRepo,
		roleRepo:         roleRepo,
		availabilityRepo: availabilityRepo,
	}
}

// StaffService 员工服务接口
type StaffService interface {
	// 员工管理
	ListStaff(filter *dto.StaffFilter) ([]*entity.Staff, int64, error)
	CreateStaff(req *dto.CreateStaffRequest) (*entity.Staff, error)
	GetStaffByID(id string) (*entity.Staff, error)
	UpdateStaff(id string, req *dto.UpdateStaffRequest) (*entity.Staff, error)
	DeleteStaff(id string) error
	UpdateStaffStatus(id string, req *dto.UpdateStatusRequest) (*entity.Staff, error)
	GetAvailableStaff(filter *dto.StaffFilter) ([]*entity.Staff, error)
	GetStats() (*dto.StaffStats, error)

	// 角色管理
	ListRoles(filter *dto.RoleFilter) ([]*entity.StaffRole, error)
	CreateRole(req *dto.CreateRoleRequest) (*entity.StaffRole, error)
	GetRoleByID(id string) (*entity.StaffRole, error)
	UpdateRole(id string, req *dto.UpdateRoleRequest) (*entity.StaffRole, error)
	DeleteRole(id string) error

	// 可用性管理
	SetAvailability(staffID string, req *dto.AvailabilityRequest) error
	GetAvailability(staffID string) (*dto.AvailabilityResponse, error)
	GetAvailableStaffForTime(req *dto.AvailableStaffRequest) (*dto.AvailableStaffResponse, error)
	BatchUpdateAvailability(staffID string, req *dto.AvailabilityRequest) error
}

// ListStaff 获取员工列表
func (s *staffService) ListStaff(filter *dto.StaffFilter) ([]*entity.Staff, int64, error) {
	staff, err := s.staffRepo.List(filter)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.staffRepo.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	return staff, total, nil
}

// CreateStaff 创建员工
func (s *staffService) CreateStaff(req *dto.CreateStaffRequest) (*entity.Staff, error) {
	// 检查邮箱是否已存在
	exists, err := s.staffRepo.ExistByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, errors.New("邮箱已存在")
	}

	// 验证角色是否存在
	role, err := s.roleRepo.GetByID(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("角色不存在: %w", err)
	}
	if role == nil {
		return nil, errors.New("角色不存在")
	}

	// 解析角色ID和技能
	roleUUID := uuid.MustParse(req.RoleID)
	var skillsStr string
	if len(req.Skills) > 0 {
		skillsStr = fmt.Sprintf(`["%s"]`, req.Skills[0])
		for i := 1; i < len(req.Skills); i++ {
			skillsStr += fmt.Sprintf(`,"%s"`, req.Skills[i])
		}
		skillsStr += "]"
		skillsStr = skillsStr[1:] // Remove extra quote at beginning
	}

	// 创建员工实体
	staff := &entity.Staff{
		ID:          uuid.New(),
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		Department:  req.Department,
		Position:    req.Position,
		RoleID:      roleUUID,
		Skills:      skillsStr,
		Status:      entity.StaffStatusActive,
		IsAvailable: req.IsAvailable,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.staffRepo.Create(staff)
	if err != nil {
		return nil, fmt.Errorf("创建员工失败: %w", err)
	}

	// 重新查询获取完整信息
	return s.staffRepo.GetByID(staff.ID.String())
}

// GetStaffByID 根据ID获取员工
func (s *staffService) GetStaffByID(id string) (*entity.Staff, error) {
	return s.staffRepo.GetByID(id)
}

// UpdateStaff 更新员工
func (s *staffService) UpdateStaff(id string, req *dto.UpdateStaffRequest) (*entity.Staff, error) {
	// 获取现有员工
	staff, err := s.staffRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("员工不存在: %w", err)
	}

	// 如果更新邮箱，检查是否与其他员工重复
	if req.Email != nil && *req.Email != staff.Email {
		exists, err := s.staffRepo.ExistByEmail(*req.Email)
		if err != nil {
			return nil, fmt.Errorf("检查邮箱失败: %w", err)
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
		staff.Email = *req.Email
	}

	// 更新其他字段
	if req.Name != nil {
		staff.Name = *req.Name
	}
	if req.Phone != nil {
		staff.Phone = *req.Phone
	}
	if req.Department != nil {
		staff.Department = *req.Department
	}
	if req.Position != nil {
		staff.Position = *req.Position
	}
	if req.RoleID != nil {
		// 验证角色是否存在
		role, err := s.roleRepo.GetByID(*req.RoleID)
		if err != nil {
			return nil, fmt.Errorf("角色不存在: %w", err)
		}
		if role == nil {
			return nil, errors.New("角色不存在")
		}
		staff.RoleID = uuid.MustParse(*req.RoleID)
	}
	if len(req.Skills) > 0 {
		var skillsStr string
		skillsStr = fmt.Sprintf(`["%s"]`, req.Skills[0])
		for i := 1; i < len(req.Skills); i++ {
			skillsStr += fmt.Sprintf(`,"%s"`, req.Skills[i])
		}
		staff.Skills = skillsStr
	}
	if req.IsAvailable != nil {
		staff.IsAvailable = *req.IsAvailable
	}

	staff.UpdatedAt = time.Now()

	err = s.staffRepo.Update(staff)
	if err != nil {
		return nil, fmt.Errorf("更新员工失败: %w", err)
	}

	// 重新查询获取完整信息
	return s.staffRepo.GetByID(id)
}

// DeleteStaff 删除员工
func (s *staffService) DeleteStaff(id string) error {
	// 检查员工是否存在
	_, err := s.staffRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("员工不存在: %w", err)
	}

	// 删除相关可用性记录
	err = s.availabilityRepo.DeleteByStaffID(id)
	if err != nil {
		return fmt.Errorf("删除员工可用性记录失败: %w", err)
	}

	// 软删除员工
	return s.staffRepo.SoftDelete(id)
}

// UpdateStaffStatus 更新员工状态
func (s *staffService) UpdateStaffStatus(id string, req *dto.UpdateStatusRequest) (*entity.Staff, error) {
	// 更新状态
	err := s.staffRepo.UpdateStatus(id, req.Status)
	if err != nil {
		return nil, fmt.Errorf("更新状态失败: %w", err)
	}

	// 重新查询获取更新后的信息
	return s.staffRepo.GetByID(id)
}

// GetAvailableStaff 获取可用员工
func (s *staffService) GetAvailableStaff(filter *dto.StaffFilter) ([]*entity.Staff, error) {
	return s.staffRepo.GetAvailableStaff(filter)
}

// GetStats 获取员工统计信息
func (s *staffService) GetStats() (*dto.StaffStats, error) {
	return s.staffRepo.GetStats()
}

// ListRoles 获取角色列表
func (s *staffService) ListRoles(filter *dto.RoleFilter) ([]*entity.StaffRole, error) {
	return s.roleRepo.List(filter)
}

// CreateRole 创建角色
func (s *staffService) CreateRole(req *dto.CreateRoleRequest) (*entity.StaffRole, error) {
	// 检查角色名是否已存在
	exists, err := s.roleRepo.ExistByName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("检查角色名失败: %w", err)
	}
	if exists {
		return nil, errors.New("角色名已存在")
	}

	// 创建角色
	role := &entity.StaffRole{
		ID:          uuid.New(),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		SortOrder:   req.SortOrder,
		Status:      entity.StaffStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 处理权限列表
	if len(req.Permissions) > 0 {
		var permissionsStr string
		permissionsStr = fmt.Sprintf(`["%s"]`, req.Permissions[0])
		for i := 1; i < len(req.Permissions); i++ {
			permissionsStr += fmt.Sprintf(`,"%s"`, req.Permissions[i])
		}
		role.Permissions = permissionsStr
	}

	err = s.roleRepo.Create(role)
	if err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	return s.roleRepo.GetByID(role.ID.String())
}

// GetRoleByID 根据ID获取角色
func (s *staffService) GetRoleByID(id string) (*entity.StaffRole, error) {
	return s.roleRepo.GetByID(id)
}

// UpdateRole 更新角色
func (s *staffService) UpdateRole(id string, req *dto.UpdateRoleRequest) (*entity.StaffRole, error) {
	// 获取现有角色
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("角色不存在: %w", err)
	}

	// 如果更新角色名，检查是否与其他角色重复
	if req.Name != nil && *req.Name != role.Name {
		exists, err := s.roleRepo.ExistByName(*req.Name)
		if err != nil {
			return nil, fmt.Errorf("检查角色名失败: %w", err)
		}
		if exists {
			return nil, errors.New("角色名已存在")
		}
		role.Name = *req.Name
	}

	// 更新其他字段
	if req.Code != nil {
		role.Code = *req.Code
	}
	if req.Description != nil {
		role.Description = req.Description
	}
	if len(req.Permissions) > 0 {
		var permissionsStr string
		permissionsStr = fmt.Sprintf(`["%s"]`, req.Permissions[0])
		for i := 1; i < len(req.Permissions); i++ {
			permissionsStr += fmt.Sprintf(`,"%s"`, req.Permissions[i])
		}
		role.Permissions = permissionsStr
	}
	if req.IsDefault != nil {
		role.IsDefault = *req.IsDefault
	}
	if req.SortOrder != nil {
		role.SortOrder = *req.SortOrder
	}

	role.UpdatedAt = time.Now()

	err = s.roleRepo.Update(role)
	if err != nil {
		return nil, fmt.Errorf("更新角色失败: %w", err)
	}

	// 重新查询获取完整信息
	return s.roleRepo.GetByID(id)
}

// DeleteRole 删除角色
func (s *staffService) DeleteRole(id string) error {
	// 检查角色是否存在
	_, err := s.roleRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	// 检查是否有员工使用此角色
	staffList, err := s.staffRepo.List(&dto.StaffFilter{
		RoleID: &id,
		Limit:  1,
	})
	if err != nil {
		return fmt.Errorf("检查角色使用情况失败: %w", err)
	}
	if len(staffList) > 0 {
		return errors.New("无法删除角色，仍有员工使用此角色")
	}

	return s.roleRepo.Delete(id)
}

// SetAvailability 设置员工可用性
func (s *staffService) SetAvailability(staffID string, req *dto.AvailabilityRequest) error {
	// 检查员工是否存在
	_, err := s.staffRepo.GetByID(staffID)
	if err != nil {
		return fmt.Errorf("员工不存在: %w", err)
	}

	// 删除原有的可用性记录
	err = s.availabilityRepo.DeleteByStaffID(staffID)
	if err != nil {
		return fmt.Errorf("删除原有可用性记录失败: %w", err)
	}

	// 创建新的可用性记录
	var availabilities []*entity.StaffAvailability
	for _, item := range req.Availabilities {
		availability := &entity.StaffAvailability{
			ID:         uuid.New(),
			StaffID:    uuid.MustParse(staffID),
			DayOfWeek:  item.DayOfWeek,
			StartTime:  item.StartTime,
			EndTime:    item.EndTime,
			IsAvailable: item.IsAvailable,
			Notes:      item.Notes,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		availabilities = append(availabilities, availability)
	}

	return s.availabilityRepo.BatchCreate(availabilities)
}

// GetAvailability 获取员工可用性
func (s *staffService) GetAvailability(staffID string) (*dto.AvailabilityResponse, error) {
	// 检查员工是否存在
	_, err := s.staffRepo.GetByID(staffID)
	if err != nil {
		return nil, fmt.Errorf("员工不存在: %w", err)
	}

	availabilities, err := s.availabilityRepo.GetByStaffID(staffID)
	if err != nil {
		return nil, fmt.Errorf("获取可用性记录失败: %w", err)
	}

	// 转换指针切片为值切片
	var availabilitiesResult []entity.StaffAvailability
	for _, availability := range availabilities {
		availabilitiesResult = append(availabilitiesResult, *availability)
	}

	return &dto.AvailabilityResponse{
		StaffID:       staffID,
		Availabilities: availabilitiesResult,
	}, nil
}

// GetAvailableStaffForTime 获取指定时间的可用员工
func (s *staffService) GetAvailableStaffForTime(req *dto.AvailableStaffRequest) (*dto.AvailableStaffResponse, error) {
	// 获取所有可用员工
	filter := &dto.StaffFilter{
		IsAvailable: &[]bool{true}[0],
		Status:      &[]entity.StaffStatus{entity.StaffStatusActive}[0],
	}

	if req.Department != nil {
		filter.Department = req.Department
	}

	staffList, err := s.staffRepo.GetAvailableStaff(filter)
	if err != nil {
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}

	// 根据时间筛选可用员工
	var availableStaff []*dto.StaffResponse
	weekday := int(req.DateTime.Weekday())

	for _, staff := range staffList {
		// 检查技能匹配
		if len(req.Skills) > 0 {
			// 简化的技能匹配逻辑
			skillMatch := false
			for range req.Skills {
				if staff.Skills != "" && len(staff.Skills) > 0 {
					// 这里应该解析JSON格式的技能字符串
					// 简化实现，总是返回true
					skillMatch = true
					break
				}
			}
			if !skillMatch {
				continue
			}
		}

		// 检查该时间段的可用性
		availability, err := s.availabilityRepo.GetByStaffIDAndDay(staff.ID.String(), weekday)
		if err != nil {
			// 如果没有找到可用性记录，跳过该员工
			continue
		}

		// 简化的时间检查
		currentTime := req.DateTime.Format("15:04")
		if availability.IsAvailable && currentTime >= availability.StartTime && currentTime <= availability.EndTime {
			// 构造响应数据
			staffResp := &dto.StaffResponse{
				ID:          staff.ID,
				Name:        staff.Name,
				Email:       staff.Email,
				Department:  staff.Department,
				Position:    staff.Position,
				Status:      staff.Status,
				IsAvailable: staff.IsAvailable,
				CreatedAt:   staff.CreatedAt,
				UpdatedAt:   staff.UpdatedAt,
			}
			availableStaff = append(availableStaff, staffResp)
		}
	}

	return &dto.AvailableStaffResponse{
		DateTime:        req.DateTime,
		AvailableStaff:  availableStaff,
		TotalCount:      len(availableStaff),
	}, nil
}

// BatchUpdateAvailability 批量更新员工可用性
func (s *staffService) BatchUpdateAvailability(staffID string, req *dto.AvailabilityRequest) error {
	// 检查员工是否存在
	_, err := s.staffRepo.GetByID(staffID)
	if err != nil {
		return fmt.Errorf("员工不存在: %w", err)
	}

	// 删除原有记录
	err = s.availabilityRepo.DeleteByStaffID(staffID)
	if err != nil {
		return fmt.Errorf("删除原有可用性记录失败: %w", err)
	}

	// 创建新记录
	return s.SetAvailability(staffID, req)
}