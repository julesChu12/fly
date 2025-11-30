package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/appointments/internal/domain/appointment"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
)

// AppointmentService 预约应用服务接口
type AppointmentService interface {
	// 基础CRUD操作
	CreateAppointment(req *dto.CreateAppointmentRequest) (*appointment.Appointment, error)
	GetAppointmentByID(id string) (*appointment.Appointment, error)
	UpdateAppointment(id string, req *dto.UpdateAppointmentRequest) (*appointment.Appointment, error)
	DeleteAppointment(id string) error

	// 查询操作
	ListAppointments(filter *dto.AppointmentFilter) ([]*appointment.Appointment, int64, error)
	GetAppointmentsByCustomerID(customerID string, filter *dto.AppointmentFilter) ([]*appointment.Appointment, error)
	GetAppointmentsByStaffID(employeeID string, filter *dto.AppointmentFilter) ([]*appointment.Appointment, error)

	// 业务操作
	UpdateAppointmentStatus(id string, req *dto.UpdateStatusRequest) (*appointment.Appointment, error)
	CancelAppointment(id string, reason *string) (*appointment.Appointment, error)
	CheckAvailability(req *dto.AvailabilityRequest) (*dto.AvailabilityResponse, error)
	CheckConflict(req *dto.ConflictCheckRequest) (*dto.ConflictInfo, error)

	// 日历相关
	GetCalendarView(req *dto.CalendarViewRequest) ([]*dto.CalendarEvent, error)
	GetUpcomingAppointments(employeeID string, limit int) ([]*appointment.Appointment, error)

	// 提醒相关
	GetPendingReminders(beforeTime time.Time) ([]*appointment.Appointment, error)
}

// appointmentService 预约服务实现
type appointmentService struct {
	appointmentRepo appointment.Repository
	domainService   *appointment.Service
}

// NewAppointmentService 创建预约服务实例
func NewAppointmentService(appointmentRepo appointment.Repository) AppointmentService {
	return &appointmentService{
		appointmentRepo: appointmentRepo,
		domainService:   appointment.NewService(appointmentRepo),
	}
}

// convertFilter 转换 DTO Filter 到 Domain Filter
func (s *appointmentService) convertFilter(dtoFilter *dto.AppointmentFilter) *appointment.Filter {
	if dtoFilter == nil {
		return nil
	}

	// 调用 SetDefaults 确保有默认值
	dtoFilter.SetDefaults()

	return &appointment.Filter{
		CustomerID:  dtoFilter.CustomerID,
		StaffID:     dtoFilter.StaffID,
		ServiceID:   dtoFilter.ServiceID,
		Status:      convertStatus(dtoFilter.Status),
		StartDate:   dtoFilter.StartDate,
		EndDate:     dtoFilter.EndDate,
		Date:        dtoFilter.Date,
		MinDuration: dtoFilter.MinDuration,
		MaxDuration: dtoFilter.MaxDuration,
		Reminder:    dtoFilter.Reminder,
		Page:        dtoFilter.Page,
		Limit:       dtoFilter.Limit,
		Sort:        dtoFilter.Sort,
		Order:       dtoFilter.Order,
	}
}

// convertStatus 转换状态类型
func convertStatus(status *appointment.AppointmentStatus) *appointment.AppointmentStatus {
	return status
}

// CreateAppointment 创建预约
func (s *appointmentService) CreateAppointment(req *dto.CreateAppointmentRequest) (*appointment.Appointment, error) {
	// 验证���间
	if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("结束时间必须晚于开始时间")
	}

	// 检查冲突
	conflictReq := &dto.ConflictCheckRequest{
		StaffID:   req.StaffID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	conflictInfo, err := s.CheckConflict(conflictReq)
	if err != nil {
		return nil, fmt.Errorf("检查时间冲突失败: %w", err)
	}
	if conflictInfo.Conflict {
		return nil, fmt.Errorf("预约时间冲突，已有 %d 个预约冲突", conflictInfo.ConflictCount)
	}

	// 设置默认提醒时间
	reminderTime := req.ReminderTime
	if req.Reminder && reminderTime == nil {
		defaultReminder := req.StartTime.Add(-30 * time.Minute)
		reminderTime = &defaultReminder
	}

	// 创建预约实体
	appointment := &appointment.Appointment{
		ID:           uuid.New(),
		CustomerID:   uuid.MustParse(req.CustomerID),
		StaffID:      uuid.MustParse(req.StaffID),
		ServiceID:    uuid.MustParse(req.ServiceID),
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Notes:        req.Notes,
		Status:       appointment.AppointmentStatusPending,
		Reminder:     req.Reminder,
		ReminderTime: reminderTime,
	}

	// 保存到数据库
	if err := s.appointmentRepo.Create(appointment); err != nil {
		return nil, fmt.Errorf("创建预约失败: %w", err)
	}

	return appointment, nil
}

// GetAppointmentByID 根据ID获取预约
func (s *appointmentService) GetAppointmentByID(id string) (*appointment.Appointment, error) {
	if id == "" {
		return nil, errors.New("预约ID不能为空")
	}

	appointment, err := s.appointmentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取预约失败: %w", err)
	}

	return appointment, nil
}

// UpdateAppointment 更新预约
func (s *appointmentService) UpdateAppointment(id string, req *dto.UpdateAppointmentRequest) (*appointment.Appointment, error) {
	if id == "" {
		return nil, errors.New("预约ID不能为空")
	}

	// 获取现有预约
	apt, err := s.appointmentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取预约失败: %w", err)
	}

	// 检查预约状态是否允许修改
	if apt.Status == appointment.AppointmentStatusCompleted || apt.Status == appointment.AppointmentStatusCancelled {
		return nil, errors.New("已完成或已取消的预约不能修改")
	}

	// 更新字段
	if req.CustomerID != nil {
		apt.CustomerID = uuid.MustParse(*req.CustomerID)
	}
	if req.StaffID != nil {
		apt.StaffID = uuid.MustParse(*req.StaffID)
	}
	if req.ServiceID != nil {
		apt.ServiceID = uuid.MustParse(*req.ServiceID)
	}
	if req.StartTime != nil {
		apt.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		apt.EndTime = *req.EndTime
	}
	if req.Notes != nil {
		apt.Notes = req.Notes
	}
	if req.Reminder != nil {
		apt.Reminder = *req.Reminder
	}
	if req.ReminderTime != nil {
		apt.ReminderTime = req.ReminderTime
	}

	// 验证时间
	if apt.EndTime.Before(apt.StartTime) {
		return nil, errors.New("结束时间必须晚于开始时间")
	}

	// 检查冲突（排除当前预约）
	excludeID := &id
	conflictReq := &dto.ConflictCheckRequest{
		StaffID:   apt.StaffID.String(),
		StartTime: apt.StartTime,
		EndTime:   apt.EndTime,
		ExcludeID: excludeID,
	}
	conflictInfo, err := s.CheckConflict(conflictReq)
	if err != nil {
		return nil, fmt.Errorf("检查时间冲突失败: %w", err)
	}
	if conflictInfo.Conflict {
		return nil, fmt.Errorf("预约时间冲突，已有 %d 个预约冲突", conflictInfo.ConflictCount)
	}

	// 保存更新
	if err := s.appointmentRepo.Update(apt); err != nil {
		return nil, fmt.Errorf("更新预约失败: %w", err)
	}

	return apt, nil
}

// DeleteAppointment 删除预约
func (s *appointmentService) DeleteAppointment(id string) error {
	if id == "" {
		return errors.New("预约ID不能为空")
	}

	// 软删除
	if err := s.appointmentRepo.SoftDelete(id); err != nil {
		return fmt.Errorf("删除预约失败: %w", err)
	}

	return nil
}

// ListAppointments 获取预约列表
func (s *appointmentService) ListAppointments(filter *dto.AppointmentFilter) ([]*appointment.Appointment, int64, error) {
	if filter == nil {
		filter = &dto.AppointmentFilter{}
	}
	filter.SetDefaults()

	domainFilter := s.convertFilter(filter)

	appointments, err := s.appointmentRepo.List(domainFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("获取预约列表失败: %w", err)
	}

	total, err := s.appointmentRepo.Count(domainFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("获取预约总数失败: %w", err)
	}

	return appointments, total, nil
}

// GetAppointmentsByCustomerID 根据客户ID获取预约
func (s *appointmentService) GetAppointmentsByCustomerID(customerID string, filter *dto.AppointmentFilter) ([]*appointment.Appointment, error) {
	if customerID == "" {
		return nil, errors.New("客户ID不能为空")
	}

	if filter == nil {
		filter = &dto.AppointmentFilter{}
	}
	filter.CustomerID = &customerID
	filter.SetDefaults()

	domainFilter := s.convertFilter(filter)

	appointments, err := s.appointmentRepo.GetByCustomerID(customerID, domainFilter)
	if err != nil {
		return nil, fmt.Errorf("获取客户预约失败: %w", err)
	}

	return appointments, nil
}

// GetAppointmentsByStaffID 根据员工ID获取预约
func (s *appointmentService) GetAppointmentsByStaffID(employeeID string, filter *dto.AppointmentFilter) ([]*appointment.Appointment, error) {
	if employeeID == "" {
		return nil, errors.New("员工ID不能为空")
	}

	if filter == nil {
		filter = &dto.AppointmentFilter{}
	}
	filter.StaffID = &employeeID
	filter.SetDefaults()

	domainFilter := s.convertFilter(filter)

	appointments, err := s.appointmentRepo.GetByStaffID(employeeID, domainFilter)
	if err != nil {
		return nil, fmt.Errorf("获取员工预约失败: %w", err)
	}

	return appointments, nil
}

// UpdateAppointmentStatus 更新预约状态
func (s *appointmentService) UpdateAppointmentStatus(id string, req *dto.UpdateStatusRequest) (*appointment.Appointment, error) {
	if id == "" {
		return nil, errors.New("预约ID不能为空")
	}

	// 获取现有预约
	apt, err := s.appointmentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取预约失败: %w", err)
	}

	// 解析新状态
	newStatus := appointment.AppointmentStatus(req.Status)
	if !apt.IsValidStatus(newStatus) {
		return nil, errors.New("无效的预约状态")
	}

	// 检查状态转换是否合法
	if !apt.CanTransitionTo(newStatus) {
		return nil, fmt.Errorf("不能从 %s 状态转换为 %s 状态", apt.Status, newStatus)
	}

	// 更新状态
	apt.Status = newStatus

	// 如果是完成状态，添加完成备注
	if newStatus == appointment.AppointmentStatusCompleted && req.CompletionNotes != nil {
		apt.Notes = req.CompletionNotes
	}

	// 保存更新
	if err := s.appointmentRepo.Update(apt); err != nil {
		return nil, fmt.Errorf("更新预约状态失败: %w", err)
	}

	return apt, nil
}

// CancelAppointment 取消预约
func (s *appointmentService) CancelAppointment(id string, reason *string) (*appointment.Appointment, error) {
	req := &dto.UpdateStatusRequest{
		Status: string(appointment.AppointmentStatusCancelled),
	}
	if reason != nil {
		req.CompletionNotes = reason
	}

	return s.UpdateAppointmentStatus(id, req)
}

// CheckAvailability 检查可用时间
func (s *appointmentService) CheckAvailability(req *dto.AvailabilityRequest) (*dto.AvailabilityResponse, error) {
	// 获取员工在指定日期的所有预约
	startOfDay := time.Date(req.Date.Year(), req.Date.Month(), req.Date.Day(), 0, 0, 0, 0, req.Date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := &dto.AppointmentFilter{
		StaffID:   &req.StaffID,
		StartDate: &startOfDay,
		EndDate:   &endOfDay,
		Sort:      "start_time",
		Order:     "asc",
	}

	domainFilter := s.convertFilter(filter)
	appointments, err := s.appointmentRepo.GetByStaffID(req.StaffID, domainFilter)
	if err != nil {
		return nil, fmt.Errorf("获取预约信息失败: %w", err)
	}

	// 生成可用时间段 (这里简化处理，实际应该考虑营业时间等)
	slots := []dto.AvailableSlot{}

	// 假设营业时间 9:00-18:00
	businessStart := startOfDay.Add(9 * time.Hour)
	businessEnd := startOfDay.Add(18 * time.Hour)

	// 按服务时长划分时间段
	current := businessStart
	for current.Add(req.ServiceDuration).Before(businessEnd) || current.Add(req.ServiceDuration).Equal(businessEnd) {
		slotEnd := current.Add(req.ServiceDuration)

		// 检查该时间段是否被占用
		available := true
		for _, apt := range appointments {
			if apt.Status == appointment.AppointmentStatusCancelled {
				continue
			}
			// 检查时间重叠
			if current.Before(apt.EndTime) && slotEnd.After(apt.StartTime) {
				available = false
				break
			}
		}

		slots = append(slots, dto.AvailableSlot{
			StartTime: current,
			EndTime:   slotEnd,
			Available: available,
		})

		current = current.Add(30 * time.Minute) // 30分钟间隔
	}

	return &dto.AvailabilityResponse{
		StaffID: req.StaffID,
		Date:    req.Date,
		Slots:   slots,
	}, nil
}

// CheckConflict 检查时间冲突
func (s *appointmentService) CheckConflict(req *dto.ConflictCheckRequest) (*dto.ConflictInfo, error) {
	conflicts, err := s.appointmentRepo.CheckConflict(req.StaffID, req.StartTime, req.EndTime, req.ExcludeID)
	if err != nil {
		return nil, fmt.Errorf("检查冲突失败: %w", err)
	}

	conflictInfo := &dto.ConflictInfo{
		Conflict:      len(conflicts) > 0,
		ConflictCount: len(conflicts),
	}

	if len(conflicts) > 0 {
		conflictIDs := make([]string, len(conflicts))
		for i, conflict := range conflicts {
			conflictIDs[i] = conflict.ID.String()
		}
		conflictInfo.ConflictIDs = conflictIDs
	}

	return conflictInfo, nil
}

// GetCalendarView 获取日历视图
func (s *appointmentService) GetCalendarView(req *dto.CalendarViewRequest) ([]*dto.CalendarEvent, error) {
	// 使用appointmentRepo获取预约数据，然后转换为日历事件
	filter := &dto.AppointmentFilter{
		StartDate: &req.StartDate,
		EndDate:   &req.EndDate,
	}
	if req.StaffID != nil {
		filter.StaffID = req.StaffID
	}

	domainFilter := s.convertFilter(filter)
	appointments, err := s.appointmentRepo.GetByDateRange(req.StartDate, req.EndDate, domainFilter)
	if err != nil {
		return nil, fmt.Errorf("获取预约数据失败: %w", err)
	}

	// 转换为日历事件
	events := make([]*dto.CalendarEvent, len(appointments))
	for i, apt := range appointments {
		events[i] = &dto.CalendarEvent{
			ID:        apt.ID.String(),
			StartTime: apt.StartTime,
			EndTime:   apt.EndTime,
			Status:    apt.Status,
			Notes:     apt.Notes,
			Duration:  apt.Duration(),
		}
	}

	return events, nil
}

// GetUpcomingAppointments 获取即将到来的预约
func (s *appointmentService) GetUpcomingAppointments(employeeID string, limit int) ([]*appointment.Appointment, error) {
	if limit <= 0 {
		limit = 10
	}

	// 获取从现在开始的预约
	now := time.Now()
	status := appointment.AppointmentStatusConfirmed
	filter := &dto.AppointmentFilter{
		StaffID:   &employeeID,
		StartDate: &now,
		Status:    &status,
		Page:      1,
		Limit:     limit,
		Sort:      "start_time",
		Order:     "asc",
	}

	domainFilter := s.convertFilter(filter)
	appointments, err := s.appointmentRepo.List(domainFilter)
	if err != nil {
		return nil, fmt.Errorf("获取即将到来的预约失败: %w", err)
	}

	return appointments, nil
}

// GetPendingReminders 获取待处理的提醒
func (s *appointmentService) GetPendingReminders(beforeTime time.Time) ([]*appointment.Appointment, error) {
	appointments, err := s.appointmentRepo.GetPendingReminders(beforeTime)
	if err != nil {
		return nil, fmt.Errorf("获取待处理提醒失败: %w", err)
	}

	return appointments, nil
}
