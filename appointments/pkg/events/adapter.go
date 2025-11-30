package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/julesChu12/fly/mora/pkg/mq"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// EventType 事件类型
type EventType string

const (
	// 预约相关事件
	EventAppointmentCreated   EventType = "appointment.created"
	EventAppointmentConfirmed EventType = "appointment.confirmed"
	EventAppointmentCancelled EventType = "appointment.cancelled"
	EventAppointmentCompleted EventType = "appointment.completed"
	EventAppointmentReminder  EventType = "appointment.reminder"

	// 订单相关事件
	EventOrderCreated   EventType = "order.created"
	EventOrderPaid      EventType = "order.paid"
	EventOrderCancelled EventType = "order.cancelled"
	EventOrderExpired   EventType = "order.expired"

	// 支付相关事件
	EventPaymentCreated   EventType = "payment.created"
	EventPaymentSucceeded EventType = "payment.succeeded"
	EventPaymentFailed    EventType = "payment.failed"
	EventPaymentRefunded  EventType = "payment.refunded"

	// 员工相关事件
	EventStaffCreated             EventType = "staff.created"
	EventStaffUpdated             EventType = "staff.updated"
	EventStaffDeleted             EventType = "staff.deleted"
	EventStaffAvailabilityChanged EventType = "staff.availability_changed"

	// 系统相关事件
	EventSystemError       EventType = "system.error"
	EventSystemMaintenance EventType = "system.maintenance"
)

// Event 事件结构
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]string      `json:"metadata"`
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event *Event) error

// EventBus 事件总线接口 (兼容原有接口)
type EventBus interface {
	// 发布事件
	Publish(ctx context.Context, event *Event) error
	PublishAsync(ctx context.Context, event *Event)

	// 订阅事件
	Subscribe(eventType EventType, handler EventHandler) error
	Unsubscribe(eventType EventType, handler EventHandler) error

	// 获取统计信息
	GetStats() *EventBusStats

	// 关闭事件总线
	Close() error
}

// EventBusStats 事件总线统计
type EventBusStats struct {
	TotalEvents      int64 `json:"total_events"`
	ProcessedEvents  int64 `json:"processed_events"`
	FailedEvents     int64 `json:"failed_events"`
	PendingEvents    int64 `json:"pending_events"`
	SubscribersCount int64 `json:"subscribers_count"`
}

// Adapter mora MQ 适配器
type Adapter struct {
	client     mq.Client
	handlers   map[EventType][]EventHandler
	logger     *logger.Logger
	stats      *EventBusStats
	topics     map[EventType]string
}

// NewAdapter 创建事件总线适配器
func NewAdapter(mqClient mq.Client, log *logger.Logger) *Adapter {
	return &Adapter{
		client:   mqClient,
		handlers: make(map[EventType][]EventHandler),
		logger:   log,
		stats: &EventBusStats{},
		topics:   getEventTopics(),
	}
}

// getEventTopics 获取事件主题映射
func getEventTopics() map[EventType]string {
	return map[EventType]string{
		// 预约事件
		EventAppointmentCreated:   "appointments.appointment.created",
		EventAppointmentConfirmed: "appointments.appointment.confirmed",
		EventAppointmentCancelled: "appointments.appointment.cancelled",
		EventAppointmentCompleted: "appointments.appointment.completed",
		EventAppointmentReminder:  "appointments.appointment.reminder",

		// 订单事件
		EventOrderCreated:   "orders.order.created",
		EventOrderPaid:      "orders.order.paid",
		EventOrderCancelled: "orders.order.cancelled",
		EventOrderExpired:   "orders.order.expired",

		// 支付事件
		EventPaymentCreated:   "payments.payment.created",
		EventPaymentSucceeded: "payments.payment.succeeded",
		EventPaymentFailed:    "payments.payment.failed",
		EventPaymentRefunded:  "payments.payment.refunded",

		// 员工事件
		EventStaffCreated:             "staff.staff.created",
		EventStaffUpdated:             "staff.staff.updated",
		EventStaffDeleted:             "staff.staff.deleted",
		EventStaffAvailabilityChanged: "staff.staff.availability_changed",

		// 系统事件
		EventSystemError:       "system.events.error",
		EventSystemMaintenance: "system.events.maintenance",
	}
}

// Publish 发布事件
func (a *Adapter) Publish(ctx context.Context, event *Event) error {
	a.stats.TotalEvents++

	// 设置事件ID和时间戳
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 验证事件
	if err := a.validateEvent(event); err != nil {
		a.stats.FailedEvents++
		return fmt.Errorf("事件验证失败: %w", err)
	}

	// 序列化事件
	payload, err := json.Marshal(event)
	if err != nil {
		a.stats.FailedEvents++
		return fmt.Errorf("事件序列化失败: %w", err)
	}

	// 获取主题
	topic, exists := a.topics[event.Type]
	if !exists {
		topic = "appointments.events" // 默认主题
	}

	// 发布消息
	headers := make(map[string]interface{})
	for k, v := range event.Metadata {
		headers[k] = v
	}
	err = a.client.Publish(ctx, topic, payload, mq.WithHeaders(headers))
	if err != nil {
		a.stats.FailedEvents++
		a.logger.Error("发布事件失败", "event_type", event.Type, "event_id", event.ID, "error", err)
		return err
	}

	a.stats.ProcessedEvents++
	a.logger.Debug("事件发布成功", "event_type", event.Type, "event_id", event.ID, "topic", topic)
	return nil
}

// PublishAsync 异步发布事件
func (a *Adapter) PublishAsync(ctx context.Context, event *Event) {
	go func() {
		if err := a.Publish(ctx, event); err != nil {
			a.logger.Error("异步发布事件失败", "event_type", event.Type, "error", err)
		}
	}()
}

// Subscribe 订阅事件
func (a *Adapter) Subscribe(eventType EventType, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("事件处理器不能为空")
	}

	a.handlers[eventType] = append(a.handlers[eventType], handler)
	a.stats.SubscribersCount++

	// 订阅 MQ 主题
	topic, exists := a.topics[eventType]
	if !exists {
		topic = "appointments.events" // 默认主题
	}

	err := a.client.Subscribe(context.Background(), topic, a.handleMessage(eventType))
	if err != nil {
		return fmt.Errorf("订阅主题失败: %w", err)
	}

	a.logger.Info("事件订阅成功", "event_type", eventType, "topic", topic)
	return nil
}

// Unsubscribe 取消订阅事件
func (a *Adapter) Unsubscribe(eventType EventType, handler EventHandler) error {
	handlers := a.handlers[eventType]
	if handlers == nil {
		return fmt.Errorf("没有订阅者: %s", eventType)
	}

	// 查找并移除处理器
	for i, h := range handlers {
		// 简单的指针比较
		if &h == &handler {
			a.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			a.stats.SubscribersCount--
			a.logger.Info("取消事件订阅成功", "event_type", eventType)
			return nil
		}
	}

	return fmt.Errorf("未找到要取消的处理器: %s", eventType)
}

// GetStats 获取统计信息
func (a *Adapter) GetStats() *EventBusStats {
	// 复制统计信息避免并发问题
	stats := *a.stats
	return &stats
}

// Close 关闭事件总线
func (a *Adapter) Close() error {
	return a.client.Close()
}

// handleMessage 处理 MQ 消息
func (a *Adapter) handleMessage(eventType EventType) mq.MessageHandler {
	return func(ctx context.Context, msg *mq.Message) error {
		// 反序列化事件
		var event Event
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			a.logger.Error("事件反序列化失败", "error", err)
			return err
		}

		// 执行处理器
		handlers := a.handlers[eventType]
		if len(handlers) == 0 {
			a.logger.Debug("没有处理器处理事件", "event_type", eventType, "event_id", event.ID)
			return nil
		}

		// 执行所有处理器
		var errors []error
		for _, handler := range handlers {
			if err := handler(ctx, &event); err != nil {
				errors = append(errors, err)
				a.logger.Error("事件处理器执行失败", "event_id", event.ID, "event_type", eventType, "error", err)
			}
		}

		if len(errors) == 0 {
			a.logger.Debug("事件处理成功", "event_id", event.ID, "event_type", eventType)
		}

		return nil
	}
}

// validateEvent 验证事件
func (a *Adapter) validateEvent(event *Event) error {
	if event.Type == "" {
		return fmt.Errorf("事件类型不能为空")
	}
	if event.Source == "" {
		return fmt.Errorf("事件源不能为空")
	}
	if event.Data == nil {
		return fmt.Errorf("事件数据不能为空")
	}
	return nil
}

// generateEventID 生成事件ID
func generateEventID() string {
	return fmt.Sprintf("EVT_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}

// NewAppointmentEvent 创建预约事件
func NewAppointmentEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		Type:   eventType,
		Source: source,
		Data:   data,
		Metadata: map[string]string{
			"service": "appointments",
			"version": "1.0.0",
		},
	}
}

// NewOrderEvent 创建订单事件
func NewOrderEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		Type:   eventType,
		Source: source,
		Data:   data,
		Metadata: map[string]string{
			"service": "kratos",
			"version": "1.0.0",
		},
	}
}

// NewPaymentEvent 创建支付事件
func NewPaymentEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		Type:   eventType,
		Source: source,
		Data:   data,
		Metadata: map[string]string{
			"service": "plutus",
			"version": "1.0.0",
		},
	}
}

// NewStaffEvent 创建员工事件
func NewStaffEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		Type:   eventType,
		Source: source,
		Data:   data,
		Metadata: map[string]string{
			"service": "staff",
			"version": "1.0.0",
		},
	}
}