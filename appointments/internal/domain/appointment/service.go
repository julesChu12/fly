package appointment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Service 预约领域服务
type Service struct {
	repo Repository
}

// NewService 创建预约领域服务
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateAppointment 创建预约
func (s *Service) CreateAppointment(customerID, staffID, serviceID string, startTime, endTime time.Time, notes *string) (*Appointment, error) {
	// 验证时间
	if startTime.After(endTime) {
		return nil, errors.New("开始时间不能晚于结束时间")
	}

	if startTime.Before(time.Now()) {
		return nil, errors.New("不能创建过去的预约")
	}

	// 检查时间冲突
	conflicts, err := s.repo.CheckConflict(staffID, startTime, endTime, nil)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, errors.New("该时间段已有预约")
	}

	// 创建预约实体
	apt := &Appointment{
		ID:         uuid.New(),
		CustomerID: uuid.MustParse(customerID),
		StaffID:    uuid.MustParse(staffID),
		ServiceID:  uuid.MustParse(serviceID),
		StartTime:  startTime,
		EndTime:    endTime,
		Notes:      notes,
		Status:     AppointmentStatusPending,
		Reminder:   true,
	}

	// 设置提醒时间
	reminderTime := startTime.Add(-15 * time.Minute)
	apt.ReminderTime = &reminderTime

	// 保存预约
	if err := s.repo.Create(apt); err != nil {
		return nil, err
	}

	return apt, nil
}

// ConfirmAppointment 确认预约
func (s *Service) ConfirmAppointment(id string) error {
	apt, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !apt.CanTransitionTo(AppointmentStatusConfirmed) {
		return errors.New("当前状态不能确认")
	}

	apt.Status = AppointmentStatusConfirmed
	return s.repo.Update(apt)
}

// StartAppointment 开始预约
func (s *Service) StartAppointment(id string) error {
	apt, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !apt.CanTransitionTo(AppointmentStatusInProgress) {
		return errors.New("当前状态不能开始")
	}

	apt.Status = AppointmentStatusInProgress
	return s.repo.Update(apt)
}

// CompleteAppointment 完成预约
func (s *Service) CompleteAppointment(id string) error {
	apt, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !apt.CanTransitionTo(AppointmentStatusCompleted) {
		return errors.New("当前状态不能完成")
	}

	apt.Status = AppointmentStatusCompleted
	return s.repo.Update(apt)
}

// CancelAppointment 取消预约
func (s *Service) CancelAppointment(id string) error {
	apt, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !apt.CanTransitionTo(AppointmentStatusCancelled) {
		return errors.New("当前状态不能取消")
	}

	apt.Status = AppointmentStatusCancelled
	return s.repo.Update(apt)
}

// RescheduleAppointment 重新安排预约
func (s *Service) RescheduleAppointment(id string, newStartTime, newEndTime time.Time) error {
	apt, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 验证时间
	if newStartTime.After(newEndTime) {
		return errors.New("开始时间不能晚于结束时间")
	}

	// 检查时间冲突
	conflicts, err := s.repo.CheckConflict(apt.StaffID.String(), newStartTime, newEndTime, &id)
	if err != nil {
		return err
	}
	if len(conflicts) > 0 {
		return errors.New("该时间段已有预约")
	}

	// 更新时间
	apt.StartTime = newStartTime
	apt.EndTime = newEndTime

	// 重新设置提醒时间
	reminderTime := newStartTime.Add(-15 * time.Minute)
	apt.ReminderTime = &reminderTime

	return s.repo.Update(apt)
}
