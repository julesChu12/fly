package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// EventType 事件类型
type EventType string

const (
	// 预约相关事件
	EventAppointmentCreated     EventType = "appointment.created"
	EventAppointmentConfirmed   EventType = "appointment.confirmed"
	EventAppointmentCancelled   EventType = "appointment.cancelled"
	EventAppointmentCompleted   EventType = "appointment.completed"
	EventAppointmentReminder    EventType = "appointment.reminder"

	// 订单相关事件
	EventOrderCreated           EventType = "order.created"
	EventOrderPaid              EventType = "order.paid"
	EventOrderCancelled         EventType = "order.cancelled"
	EventOrderExpired           EventType = "order.expired"

	// 支付相关事件
	EventPaymentCreated         EventType = "payment.created"
	EventPaymentSucceeded       EventType = "payment.succeeded"
	EventPaymentFailed          EventType = "payment.failed"
	EventPaymentRefunded        EventType = "payment.refunded"

	// 员工相关事件
	EventStaffCreated           EventType = "staff.created"
	EventStaffUpdated           EventType = "staff.updated"
	EventStaffDeleted           EventType = "staff.deleted"
	EventStaffAvailabilityChanged EventType = "staff.availability_changed"

	// 系统相关事件
	EventSystemError            EventType = "system.error"
	EventSystemMaintenance      EventType = "system.maintenance"
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

// EventBus 事件总线接口
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

// InMemoryEventBus 内存事件总线实现
type InMemoryEventBus struct {
	handlers       map[EventType][]EventHandler
	eventQueue     chan *Event
	bufferSize      int
	maxMemoryEvents int
	subscribers    int
	stats           *EventBusStats
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	logger          *logger.Logger

	// 性能优化
	workerPool      int
	batchProcessor  *EventBatchProcessor

	// 内存管理
	eventCount      int64
	lastCleanup     time.Time
	cleanupInterval time.Duration
}

// NewInMemoryEventBus 创建内存事件总线
func NewInMemoryEventBus(ctx context.Context, bufferSize int, log *logger.Logger) *InMemoryEventBus {
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	ctx, cancel := context.WithCancel(ctx)

	bus := &InMemoryEventBus{
		handlers:    make(map[EventType][]EventHandler),
		eventQueue:  make(chan *Event, bufferSize),
		eventBuffer: make([]*Event, 0, bufferSize),
		bufferSize:   bufferSize,
		stats: &EventBusStats{},
		ctx:         ctx,
		cancel:      cancel,
		logger:      log,
	}

	// 启动事件处理循环
	go bus.processEvents()

	return bus
}

// Publish 发布事件
func (bus *InMemoryEventBus) Publish(ctx context.Context, event *Event) error {
	bus.stats.TotalEvents++

	// 设置事件ID和时间戳
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 验证事件
	if err := bus.validateEvent(event); err != nil {
		bus.stats.FailedEvents++
		bus.logger.Error("事件验证失败", "event_id", event.ID, "event_type", event.Type, "error", err)
		return fmt.Errorf("事件验证失败: %w", err)
	}

	select {
	case bus.eventQueue <- event:
		bus.stats.PendingEvents++
		bus.logger.Debug("事件已发布到队列", "event_id", event.ID, "event_type", event.Type)
		return nil
	case <-ctx.Done():
		bus.stats.FailedEvents++
		return fmt.Errorf("发布事件超时: %w", ctx.Err())
	case <-time.After(5 * time.Second):
		bus.stats.FailedEvents++
		return fmt.Errorf("发布事件超时")
	}
}

// PublishAsync 异步发布事件
func (bus *InMemoryEventBus) PublishAsync(ctx context.Context, event *Event) {
	go func() {
		if err := bus.Publish(ctx, event); err != nil {
			bus.logger.Error("异步发布事件失败", "event_type", event.Type, "error", err)
		}
	}()
}

// Subscribe 订阅事件
func (bus *InMemoryEventBus) Subscribe(eventType EventType, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("事件处理器不能为空")
	}

	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	bus.subscribers++
	bus.stats.SubscribersCount = int64(bus.subscribers)

	bus.logger.Info("事件订阅成功", "event_type", eventType, "subscribers_count", bus.subscribers)
	return nil
}

// Unsubscribe 取消订阅事件
func (bus *InMemoryEventBus) Unsubscribe(eventType EventType, handler EventHandler) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	handlers := bus.handlers[eventType]
	if handlers == nil {
		return fmt.Errorf("没有订阅者: %s", eventType)
	}

	// 查找并移除处理器
	for i, h := range handlers {
		// 简单的指针比较（在实际应用中可能需要更复杂的识别机制）
		if &h == &handler {
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			bus.subscribers--
			bus.stats.SubscribersCount = int64(bus.subscribers)
			bus.logger.Info("取消事件订阅成功", "event_type", eventType)
			return nil
		}
	}

	return fmt.Errorf("未找到要取消的处理器: %s", eventType)
}

// GetStats 获取统计信息
func (bus *InMemoryEventBus) GetStats() *EventBusStats {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	// 复制统计信息避免并发问题
	stats := *bus.stats
	return &stats
}

// Close 关闭事件总线
func (bus *InMemoryEventBus) Close() error {
	bus.cancel()
	close(bus.eventQueue)
	bus.logger.Info("事件总线已关闭")
	return nil
}

// processEvents 事件处理循环
func (bus *InMemoryEventBus) processEvents() {
	for {
		select {
		case event := <-bus.eventQueue:
			bus.handleEvent(event)
		case <-bus.ctx.Done():
			bus.logger.Info("事件处理循环停止")
			return
		}
	}
}

// handleEvent 处理单个事件
func (bus *InMemoryEventBus) handleEvent(event *Event) {
	bus.mu.RLock()
	handlers := bus.handlers[event.Type]
	bus.mu.RUnlock()

	if len(handlers) == 0 {
		bus.logger.Debug("没有处理器处理事件", "event_type", event.Type, "event_id", event.ID)
		bus.stats.PendingEvents--
		return
	}

	// 并发执行所有处理器
	var wg sync.WaitGroup
	errorChan := make(chan error, len(handlers))

	for _, handler := range handlers {
		wg.Add(1)
		go func(h EventHandler) {
			defer wg.Done()
			if err := h(bus.ctx, event); err != nil {
				errorChan <- err
			}
		}(handler)
	}

	// 等待所有处理器完成
	wg.Wait()
	close(errorChan)

	// 处理错误
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
		bus.stats.FailedEvents++
		bus.logger.Error("事件处理器执行失败", "event_id", event.ID, "event_type", event.Type, "error", err)
	}

	if len(errors) == 0 {
		bus.stats.ProcessedEvents++
		bus.logger.Debug("事件处理成功", "event_id", event.ID, "event_type", event.Type)
	} else {
		bus.logger.Error("部分事件处理器失败", "event_id", event.ID, "event_type", event.Type, "error_count", len(errors))
	}

	bus.stats.PendingEvents--
}

// validateEvent 验证事件
func (bus *InMemoryEventBus) validateEvent(event *Event) error {
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
		Type:     eventType,
		Source:   source,
		Data:     data,
		Metadata: map[string]string{
			"service": "appointments",
			"version": "1.0.0",
		},
	}
}

// NewOrderEvent 创建订单事件
func NewOrderEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		Type:     eventType,
		Source:   source,
		Data:     data,
		Metadata: map[string]string{
			"service": "kratos",
			"version": "1.0.0",
		},
	}
}

// NewPaymentEvent 创建支付事件
func NewPaymentEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		Type:     eventType,
		Source:   source,
		Data:     data,
		Metadata: map[string]string{
			"service": "plutus",
			"version": "1.0.0",
		},
	}
}

// NewStaffEvent 创建员工事件
func NewStaffEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		Type:     eventType,
		Source:   source,
		Data:     data,
		Metadata: map[string]string{
			"service": "staff",
			"version": "1.0.0",
		},
	}
}

// 全局事件总线实例（在实际应用中应该通过依赖注入管理）
var GlobalEventBus EventBus

// InitializeGlobalEventBus 初始化全局事件总线
func InitializeGlobalEventBus(ctx context.Context, bufferSize int, log *logger.Logger) {
	GlobalEventBus = NewInMemoryEventBus(ctx, bufferSize, log)
	log.Info("全局事件总线初始化成功")
}

// PublishGlobalEvent 发布全局事件
func PublishGlobalEvent(ctx context.Context, event *Event) error {
	if GlobalEventBus == nil {
		return fmt.Errorf("全局事件总线未初始化")
	}
	return GlobalEventBus.Publish(ctx, event)
}

// PublishGlobalEventAsync 异步发布全局事件
func PublishGlobalEventAsync(ctx context.Context, event *Event) {
	if GlobalEventBus == nil {
		return
	}
	GlobalEventBus.PublishAsync(ctx, event)
}

// SubscribeGlobalEvent 订阅全局事件
func SubscribeGlobalEvent(eventType EventType, handler EventHandler) error {
	if GlobalEventBus == nil {
		return fmt.Errorf("全局事件总线未初始化")
	}
	return GlobalEventBus.Subscribe(eventType, handler)
}

// GetGlobalEventBusStats 获取全局事件总线统计
func GetGlobalEventBusStats() *EventBusStats {
	if GlobalEventBus == nil {
		return nil
	}
	return GlobalEventBus.GetStats()
}