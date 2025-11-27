package service

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/appointments/pkg/events"
)

// EventService 事件服务
type EventService struct {
	eventBus events.EventBus
	logger    *logger.Logger
}

// NewEventService 创建事件服务
func NewEventService(eventBus events.EventBus, log *logger.Logger) *EventService {
	return &EventService{
		eventBus: eventBus,
		logger:    log,
	}
}

// InitializeEventHandlers 初始化事件处理器
func (s *EventService) InitializeEventHandlers(ctx context.Context) error {
	// 订阅外部事件
	if err := s.subscribeToExternalEvents(ctx); err != nil {
		return fmt.Errorf("订阅外部事件失败: %w", err)
	}

	s.logger.Info("事件服务初始化完成")
	return nil
}

// PublishAppointmentCreated 发布预约创建事件
func (s *EventService) PublishAppointmentCreated(ctx context.Context, appointment *entity.Appointment, orderInfo interface{}) error {
	eventData := map[string]interface{}{
		"appointment_id": appointment.ID.String(),
		"customer_id":    appointment.CustomerID.String(),
		"staff_id":       appointment.StaffID.String(),
		"service_id":     appointment.ServiceID.String(),
		"start_time":     appointment.StartTime,
		"end_time":       appointment.EndTime,
		"status":         string(appointment.Status),
		"reminder":       appointment.Reminder,
		"order_info":     orderInfo,
	}

	event := events.NewAppointmentEvent(events.EventAppointmentCreated, "appointments-service", eventData)

	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("发布预约创建事件失败", "appointment_id", appointment.ID, "error", err)
		return err
	}

	s.logger.Info("预约创建事件发布成功", "appointment_id", appointment.ID)
	return nil
}

// PublishAppointmentConfirmed 发布预约确认事件
func (s *EventService) PublishAppointmentConfirmed(ctx context.Context, appointment *entity.Appointment, paymentInfo interface{}) error {
	eventData := map[string]interface{}{
		"appointment_id": appointment.ID.String(),
		"customer_id":    appointment.CustomerID.String(),
		"staff_id":       appointment.StaffID.String(),
		"service_id":     appointment.ServiceID.String(),
		"start_time":     appointment.StartTime,
		"end_time":       appointment.EndTime,
		"status":         string(appointment.Status),
		"payment_info":   paymentInfo,
	}

	event := events.NewAppointmentEvent(events.EventAppointmentConfirmed, "appointments-service", eventData)

	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("发布预约确认事件失败", "appointment_id", appointment.ID, "error", err)
		return err
	}

	s.logger.Info("预约确认事件发布成功", "appointment_id", appointment.ID)
	return nil
}

// PublishAppointmentCancelled 发布预约取消事件
func (s *EventService) PublishAppointmentCancelled(ctx context.Context, appointment *entity.Appointment, reason string, refundInfo interface{}) error {
	eventData := map[string]interface{}{
		"appointment_id": appointment.ID.String(),
		"customer_id":    appointment.CustomerID.String(),
		"staff_id":       appointment.StaffID.String(),
		"service_id":     appointment.ServiceID.String(),
		"status":         string(appointment.Status),
		"cancel_reason":  reason,
		"refund_info":    refundInfo,
	}

	event := events.NewAppointmentEvent(events.EventAppointmentCancelled, "appointments-service", eventData)

	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("发布预约取消事件失败", "appointment_id", appointment.ID, "error", err)
		return err
	}

	s.logger.Info("预约取消事件发布成功", "appointment_id", appointment.ID, "reason", reason)
	return nil
}

// PublishAppointmentCompleted 发布预约完成事件
func (s *EventService) PublishAppointmentCompleted(ctx context.Context, appointment *entity.Appointment, completionNotes *string) error {
	eventData := map[string]interface{}{
		"appointment_id":   appointment.ID.String(),
		"customer_id":     appointment.CustomerID.String(),
		"staff_id":        appointment.StaffID.String(),
		"service_id":      appointment.ServiceID.String(),
		"start_time":      appointment.StartTime,
		"end_time":        appointment.EndTime,
		"status":          string(appointment.Status),
		"completion_notes": completionNotes,
		"duration_minutes": appointment.Duration().Minutes(),
	}

	event := events.NewAppointmentEvent(events.EventAppointmentCompleted, "appointments-service", eventData)

	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("发布预约完成事件失败", "appointment_id", appointment.ID, "error", err)
		return err
	}

	s.logger.Info("预约完成事件发布成功", "appointment_id", appointment.ID)
	return nil
}

// PublishAppointmentReminder 发布预约提醒事件
func (s *EventService) PublishAppointmentReminder(ctx context.Context, appointment *entity.Appointment, reminderType string) error {
	eventData := map[string]interface{}{
		"appointment_id": appointment.ID.String(),
		"customer_id":    appointment.CustomerID.String(),
		"staff_id":       appointment.StaffID.String(),
		"service_id":     appointment.ServiceID.String(),
		"start_time":     appointment.StartTime,
		"end_time":       appointment.EndTime,
		"reminder_type":  reminderType,
		"notes":          appointment.Notes,
	}

	event := events.NewAppointmentEvent(events.EventAppointmentReminder, "appointments-service", eventData)

	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("发布预约提醒事件失败", "appointment_id", appointment.ID, "error", err)
		return err
	}

	s.logger.Info("预约提醒事件发布成功", "appointment_id", appointment.ID, "reminder_type", reminderType)
	return nil
}

// subscribeToExternalEvents 订阅外部事件
func (s *EventService) subscribeToExternalEvents(ctx context.Context) error {
	// 订阅支付成功事件
	if err := s.eventBus.Subscribe(events.EventPaymentSucceeded, s.handlePaymentSucceeded); err != nil {
		return fmt.Errorf("订阅支付成功事件失败: %w", err)
	}

	// 订阅支付失败事件
	if err := s.eventBus.Subscribe(events.EventPaymentFailed, s.handlePaymentFailed); err != nil {
		return fmt.Errorf("订阅支付失败事件失败: %w", err)
	}

	// 订阅员工可用性变更事件
	if err := s.eventBus.Subscribe(events.EventStaffAvailabilityChanged, s.handleStaffAvailabilityChanged); err != nil {
		return fmt.Errorf("订阅员工可用性变更事件失败: %w", err)
	}

	// 订阅订单过期事件
	if err := s.eventBus.Subscribe(events.EventOrderExpired, s.handleOrderExpired); err != nil {
		return fmt.Errorf("订阅订单过期事件失败: %w", err)
	}

	s.logger.Info("外部事件订阅完成")
	return nil
}

// handlePaymentSucceeded 处理支付成功事件
func (s *EventService) handlePaymentSucceeded(ctx context.Context, event *events.Event) error {
	s.logger.Info("收到支付成功事件", "event_id", event.ID, "order_id", event.Data["order_id"])

	// 从事件数据中提取订单和预约信息
	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		return fmt.Errorf("无效的订单ID")
	}

	appointmentID, ok := event.Data["appointment_id"].(string)
	if !ok {
		return fmt.Errorf("无效的预约ID")
	}

	// 在实际应用中，这里应该更新预约状态为已确认
	s.logger.Info("支付成功，预约状态已更新", "order_id", orderID, "appointment_id", appointmentID)

	return nil
}

// handlePaymentFailed 处理支付失败事件
func (s *EventService) handlePaymentFailed(ctx context.Context, event *events.Event) error {
	s.logger.Info("收到支付失败事件", "event_id", event.ID, "order_id", event.Data["order_id"])

	// 从事件数据中提取订单和预约信息
	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		return fmt.Errorf("无效的订单ID")
	}

	appointmentID, ok := event.Data["appointment_id"].(string)
	if !ok {
		return fmt.Errorf("无效的预约ID")
	}

	// 在实际应用中，这里应该处理支付失败逻辑
	s.logger.Info("支付失败，预约状态待处理", "order_id", orderID, "appointment_id", appointmentID)

	return nil
}

// handleStaffAvailabilityChanged 处理员工可用性变更事件
func (s *EventService) handleStaffAvailabilityChanged(ctx context.Context, event *events.Event) error {
	s.logger.Info("收到员工可用性变更事件", "event_id", event.ID, "staff_id", event.Data["staff_id"])

	// 从事件数据中提取员工信息
	staffID, ok := event.Data["staff_id"].(string)
	if !ok {
		return fmt.Errorf("无效的员工ID")
	}

	// 在实际应用中，这里应该更新员工的可用性信息
	s.logger.Info("员工可用性已更新", "staff_id", staffID)

	return nil
}

// handleOrderExpired 处理订单过期事件
func (s *EventService) handleOrderExpired(ctx context.Context, event *events.Event) error {
	s.logger.Info("收到订单过期事件", "event_id", event.ID, "order_id", event.Data["order_id"])

	// 从事件数据中提取订单和预约信息
	orderID, ok := event.Data["order_id"].(string)
	if !ok {
		return fmt.Errorf("无效的订单ID")
	}

	appointmentID, ok := event.Data["appointment_id"].(string)
	if !ok {
		return fmt.Errorf("无效的预约ID")
	}

	// 在实际应用中，这里应该处理订单过期逻辑
	s.logger.Info("订单已过期，预约状态待处理", "order_id", orderID, "appointment_id", appointmentID)

	return nil
}

// ScheduleAppointmentReminder 安排预约提醒
func (s *EventService) ScheduleAppointmentReminder(ctx context.Context, appointment *entity.Appointment) {
	// 计算提醒时间（预约前24小时）
	if appointment.Reminder && appointment.ReminderTime != nil {
		// 在实际应用中，这里应该使用调度器安排提醒
		go func() {
			// 等待到提醒时间
			select {
			case <-ctx.Done():
				return
			case <-time.After(appointment.ReminderTime.Sub(time.Now())):
				if err := s.PublishAppointmentReminder(ctx, appointment, "pre_appointment"); err != nil {
					s.logger.Error("发布预约提醒失败", "appointment_id", appointment.ID, "error", err)
				}
			}
		}()
	}
}

// PublishSystemEvent 发布系统事件
func (s *EventService) PublishSystemEvent(ctx context.Context, eventType events.EventType, message string, data map[string]interface{}) error {
	eventData := map[string]interface{}{
		"message":     message,
		"service":     "appointments",
		"timestamp":   time.Now(),
		"details":     data,
	}

	event := &events.Event{
		Type:     eventType,
		Source:   "appointments-service",
		Data:     eventData,
		Metadata: map[string]string{
			"service":   "appointments",
			"version":   "1.0.0",
			"severity":  "info",
		},
	}

	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("发布系统事件失败", "event_type", eventType, "error", err)
		return err
	}

	s.logger.Info("系统事件发布成功", "event_type", eventType, "message", message)
	return nil
}