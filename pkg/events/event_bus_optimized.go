package events

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// EventBatchProcessor 批量事件处理器
type EventBatchProcessor struct {
	batchSize   int
	flushDelay  time.Duration
	eventBuffer []*Event
	mu          sync.Mutex
	handler     func([]*Event) error
	logger      *logger.Logger
}

// NewEventBatchProcessor 创建批量事件处理器
func NewEventBatchProcessor(batchSize int, flushDelay time.Duration, handler func([]*Event) error, log *logger.Logger) *EventBatchProcessor {
	return &EventBatchProcessor{
		batchSize:   batchSize,
		flushDelay:  flushDelay,
		eventBuffer: make([]*Event, 0, batchSize),
		handler:     handler,
		logger:      log,
	}
}

// AddEvent 添加事件到批量处理器
func (p *EventBatchProcessor) AddEvent(event *Event) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.eventBuffer = append(p.eventBuffer, event)

	if len(p.eventBuffer) >= p.batchSize {
		p.flush()
	}
}

// Flush 刷新缓冲区
func (p *EventBatchProcessor) Flush() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.flush()
}

// flush 内部刷新方法
func (p *EventBatchProcessor) flush() {
	if len(p.eventBuffer) == 0 {
		return
	}

	batch := make([]*Event, len(p.eventBuffer))
	copy(batch, p.eventBuffer)
	p.eventBuffer = p.eventBuffer[:0]

	// 异步处理批量事件
	go func() {
		if err := p.handler(batch); err != nil {
			p.logger.Error("批量事件处理失败", "batch_size", len(batch), "error", err)
		}
	}()
}

// EventWorkerPool 事件处理工作池
type EventWorkerPool struct {
	workers   int
	jobQueue  chan *EventJob
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *logger.Logger
}

// EventJob 事件处理任务
type EventJob struct {
	Event    *Event
	Handler EventHandler
	Done     chan error
}

// NewEventWorkerPool 创建事件处理工作池
func NewEventWorkerPool(workers int, ctx context.Context, log *logger.Logger) *EventWorkerPool {
	ctx, cancel := context.WithCancel(ctx)

	pool := &EventWorkerPool{
		workers:  workers,
		jobQueue: make(chan *EventJob, workers*10), // 缓冲队列
		ctx:      ctx,
		cancel:   cancel,
		logger:   log,
	}

	// 启动工作池
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	return pool
}

// worker 工作协程
func (p *EventWorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case job := <-p.jobQueue:
			if job != nil {
				err := job.Handler(p.ctx, job.Event)
				if job.Done != nil {
					select {
					case job.Done <- err:
					case <-p.ctx.Done():
						return
					}
				}
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// Submit 提交事件处理任务
func (p *EventWorkerPool) Submit(event *Event, handler EventHandler) <-chan error {
	done := make(chan error, 1)
	job := &EventJob{
		Event:    event,
		Handler: handler,
		Done:     done,
	}

	select {
	case p.jobQueue <- job:
		return done
	case <-p.ctx.Done():
		close(done)
		return nil
	}
}

// Close 关闭工作池
func (p *EventWorkerPool) Close() {
	p.cancel()
	close(p.jobQueue)
	p.wg.Wait()
}

// OptimizedEventBus 优化的事件总线实现
type OptimizedEventBus struct {
	// 事件处理
	workerPool      *EventWorkerPool
	batchProcessor  *EventBatchProcessor

	// 订阅管理
	handlers       map[EventType][]EventHandler
	handlersMu     sync.RWMutex

	// 配置
	config         *EventBusConfig

	// 统计信息
	stats          *EventBusStats

	// 生命周期
	ctx            context.Context
	cancel         context.CancelFunc

	// 日志
	logger         *logger.Logger
}

// EventBusConfig 事件总线配置
type EventBusConfig struct {
	BufferSize       int           `yaml:"buffer_size"`
	MaxMemoryEvents  int           `yaml:"max_memory_events"`
	WorkerPoolSize   int           `yaml:"worker_pool_size"`
	BatchSize        int           `yaml:"batch_size"`
	BatchFlushDelay  time.Duration `yaml:"batch_flush_delay"`
	CleanupInterval  time.Duration `yaml:"cleanup_interval"`
	EnableBatching   bool          `yaml:"enable_batching"`
	EnableWorkerPool bool          `yaml:"enable_worker_pool"`
}

// DefaultEventBusConfig 默认配置
func DefaultEventBusConfig() *EventBusConfig {
	return &EventBusConfig{
		BufferSize:       1000,
	MaxMemoryEvents:  10000,
		WorkerPoolSize:   4,
		BatchSize:        10,
		BatchFlushDelay:  100 * time.Millisecond,
		CleanupInterval:  5 * time.Minute,
		EnableBatching:   true,
		EnableWorkerPool: true,
	}
}

// NewOptimizedEventBus 创建优化的事件总线
func NewOptimizedEventBus(ctx context.Context, config *EventBusConfig, log *logger.Logger) *OptimizedEventBus {
	if config == nil {
		config = DefaultEventBusConfig()
	}

	ctx, cancel := context.WithCancel(ctx)

	bus := &OptimizedEventBus{
		handlers:   make(map[EventType][]EventHandler),
		config:     config,
		stats:      &EventBusStats{},
		ctx:        ctx,
		cancel:     cancel,
		logger:     log,
	}

	// 初始化工作池
	if config.EnableWorkerPool {
		bus.workerPool = NewEventWorkerPool(config.WorkerPoolSize, ctx, log)
	}

	// 初始化批量处理器
	if config.EnableBatching {
		bus.batchProcessor = NewEventBatchProcessor(
			config.BatchSize,
			config.BatchFlushDelay,
			bus.processBatchEvents,
			log,
		)
	}

	// 启动清理协程
	go bus.cleanupRoutine()

	return bus
}

// processBatchEvents 批量处理事件
func (bus *OptimizedEventBus) processBatchEvents(events []*Event) error {
	bus.handlersMu.RLock()
	defer bus.handlersMu.RUnlock()

	var processedCount int64
	var failedCount int64

	for _, event := range events {
		handlers := bus.handlers[event.Type]
		if len(handlers) == 0 {
			continue
		}

		for _, handler := range handlers {
			if err := handler(bus.ctx, event); err != nil {
				bus.logger.Error("事件处理失败",
					"event_id", event.ID,
					"event_type", event.Type,
					"error", err)
				failedCount++
			} else {
				processedCount++
			}
		}
	}

	// 更新统计信息
	atomic.AddInt64(&bus.stats.ProcessedEvents, processedCount)
	atomic.AddInt64(&bus.stats.FailedEvents, failedCount)

	return nil
}

// cleanupRoutine 清理协程
func (bus *OptimizedEventBus) cleanupRoutine() {
	ticker := time.NewTicker(bus.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bus.performCleanup()
		case <-bus.ctx.Done():
			return
		}
	}
}

// performCleanup 执行清理操作
func (bus *OptimizedEventBus) performCleanup() {
	bus.logger.Debug("开始事件总线清理操作")

	// 检查是否需要清理订阅者
	bus.handlersMu.Lock()
	totalHandlers := 0
	for _, handlers := range bus.handlers {
		totalHandlers += len(handlers)
	}
	bus.handlersMu.Unlock()

	// 记录统计信息
	bus.logger.Debug("事件总线统计",
		"total_events", bus.stats.TotalEvents,
		"processed_events", bus.stats.ProcessedEvents,
		"failed_events", bus.stats.FailedEvents,
	"pending_events", bus.stats.PendingEvents,
	"total_handlers", totalHandlers)

	// 可以添加更多清理逻辑，如清理过期事件等
}

// Publish 发布事件（优化版本）
func (bus *OptimizedEventBus) Publish(ctx context.Context, event *Event) error {
	atomic.AddInt64(&bus.stats.TotalEvents, 1)

	// 设置事件元数据
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 快速验证事件
	if err := bus.validateEvent(event); err != nil {
		atomic.AddInt64(&bus.stats.FailedEvents, 1)
		bus.logger.Error("事件验证失败", "event_id", event.ID, "error", err)
		return fmt.Errorf("事件验证失败: %w", err)
	}

	// 使用批量处理器或工作池处理事件
	if bus.config.EnableBatching && bus.batchProcessor != nil {
		bus.batchProcessor.AddEvent(event)
		atomic.AddInt64(&bus.stats.PendingEvents, 1)
	} else if bus.config.EnableWorkerPool && bus.workerPool != nil {
		atomic.AddInt64(&bus.stats.PendingEvents, 1)
		done := bus.workerPool.Submit(event, func(ctx context.Context, e *Event) error {
			return bus.processEvent(ctx, e)
		})

		select {
		case err := <-done:
			if err != nil {
				atomic.AddInt64(&bus.stats.FailedEvents, 1)
				return err
			}
			atomic.AddInt64(&bus.stats.ProcessedEvents, 1)
			atomic.AddInt64(&bus.stats.PendingEvents, -1)
			return nil
		case <-ctx.Done():
			atomic.AddInt64(&bus.stats.PendingEvents, -1)
			return ctx.Err()
		case <-time.After(5 * time.Second):
			atomic.AddInt64(&bus.stats.FailedEvents, 1)
			atomic.AddInt64(&bus.PendingEvents, -1)
			return fmt.Errorf("事件处理超时")
		}
	}

	return nil
}

// processEvent 处理单个事件
func (bus *OptimizedEventBus) processEvent(ctx context.Context, event *Event) error {
	bus.handlersMu.RLock()
	handlers := bus.handlers[event.Type]
	bus.handlersMu.RUnlock()

	if len(handlers) == 0 {
		bus.logger.Debug("没有处理器处理事件", "event_type", event.Type, "event_id", event.ID)
		return nil
	}

	// 并发执行所有处理器
	var wg sync.WaitGroup
	errorChan := make(chan error, len(handlers))

	for _, handler := range handlers {
		wg.Add(1)
		go func(h EventHandler) {
			defer wg.Done()
			if err := h(ctx, event); err != nil {
				errorChan <- err
			}
		}(handler)
	}

	// 等待所有处理器完成
	wg.Wait()
	close(errorChan)

	// 处理错误
	var hasErrors bool
	for err := range errorChan {
		bus.logger.Error("事件处理器执行失败", "event_id", event.ID, "event_type", event.Type, "error", err)
		hasErrors = true
	}

	if hasErrors {
		return fmt.Errorf("部分事件处理器失败")
	}

	return nil
}

// validateEvent 验证事件
func (bus *OptimizedEventBus) validateEvent(event *Event) error {
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

// Subscribe 订阅事件
func (bus *OptimizedEventBus) Subscribe(eventType EventType, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("事件处理器不能为空")
	}

	bus.handlersMu.Lock()
	defer bus.handlersMu.Unlock()

	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	bus.stats.SubscribersCount++

	bus.logger.Info("事件订阅成功", "event_type", eventType)
	return nil
}

// Unsubscribe 取消订阅事件
func (bus *OptimizedEventBus) Unsubscribe(eventType EventType, handler EventHandler) error {
	bus.handlersMu.Lock()
	defer bus.handlersMu.Unlock()

	handlers := bus.handlers[eventType]
	if handlers == nil {
		return fmt.Errorf("没有订阅者: %s", eventType)
	}

	// 查找并移除处理器（简化实现）
	for i, h := range handlers {
		if &h == &handler {
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			bus.stats.SubscribersCount--
			bus.logger.Info("取消事件订阅成功", "event_type", eventType)
			return nil
		}
	}

	return fmt.Errorf("未找到要取消的处理器: %s", eventType)
}

// GetStats 获取统计信息
func (bus *OptimizedEventBus) GetStats() *EventBusStats {
	// 创建统计信息的副本
	return &EventBusStats{
		TotalEvents:      bus.stats.TotalEvents,
		ProcessedEvents:  bus.stats.ProcessedEvents,
		FailedEvents:     bus.stats.FailedEvents,
		PendingEvents:    bus.stats.PendingEvents,
		SubscribersCount: bus.stats.SubscribersCount,
	}
}

// Close 关闭事件总线
func (bus *OptimizedEventBus) Close() error {
	bus.cancel()

	if bus.workerPool != nil {
		bus.workerPool.Close()
	}

	bus.logger.Info("优化事件总线已关闭")
	return nil
}

// generateEventID 生成事件ID（优化版本）
func generateEventID() string {
	// 使用时间戳和原子计数器生成唯一ID
	timestamp := time.Now().UnixNano()
	var counter int64
	counter = atomic.AddInt64(&counter, 1)

	return fmt.Sprintf("EVT_%d_%d", timestamp, counter)
}