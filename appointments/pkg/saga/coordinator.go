package saga

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// SagaStatus Saga状态
type SagaStatus string

const (
	SagaStatusPending      SagaStatus = "pending"      // 待执行
	SagaStatusRunning      SagaStatus = "running"      // 执行中
	SagaStatusCompleted    SagaStatus = "completed"    // 已完成
	SagaStatusFailed       SagaStatus = "failed"       // 失败
	SagaStatusCompensating SagaStatus = "compensating" // 补偿中
	SagaStatusCompensated  SagaStatus = "compensated"  // 已补偿
	SagaStatusCancelled    SagaStatus = "cancelled"    // 已取消
)

// StepStatus 步骤状态
type StepStatus string

const (
	StepStatusPending     StepStatus = "pending"     // 待执行
	StepStatusRunning     StepStatus = "running"     // 执行中
	StepStatusCompleted   StepStatus = "completed"   // 已完成
	StepStatusFailed      StepStatus = "failed"      // 失败
	StepStatusSkipped     StepStatus = "skipped"     // 已跳过
	StepStatusCompensated StepStatus = "compensated" // 已补偿
)

// SagaStep Saga步骤
type SagaStep struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Execute     StepHandler  `json:"-"`
	Compensate  StepHandler  `json:"-"`
	Status      StepStatus   `json:"status"`
	RetryPolicy *RetryPolicy `json:"retry_policy"`
	Result      interface{}  `json:"result,omitempty"`
	Error       error        `json:"error,omitempty"`
	StartTime   time.Time    `json:"start_time"`
	EndTime     *time.Time   `json:"end_time,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// StepHandler 步骤处理器函数类型
type StepHandler func(ctx context.Context) (interface{}, error)

// SagaTransaction Saga事务
type SagaTransaction struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Steps          []*SagaStep `json:"steps"`
	Status         SagaStatus  `json:"status"`
	CurrentStep    int         `json:"current_step"`
	Payload        interface{} `json:"payload,omitempty"`
	CompletedSteps int         `json:"completed_steps"`
	FailedSteps    int         `json:"failed_steps"`
	StartTime      time.Time   `json:"start_time"`
	EndTime        *time.Time  `json:"end_time,omitempty"`
	CompensateAll  bool        `json:"compensate_all"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`

	mu sync.RWMutex `json:"-"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	Delay       time.Duration `json:"delay"`
	Backoff     BackoffType   `json:"backoff"`
}

// BackoffType 退避类型
type BackoffType string

const (
	BackoffFixed       BackoffType = "fixed"       // 固定延迟
	BackoffLinear      BackoffType = "linear"      // 线性退避
	BackoffExponential BackoffType = "exponential" // 指数退避
)

// SagaCoordinator Saga协调器接口
type SagaCoordinator interface {
	// ExecuteSaga 执行Saga事务
	ExecuteSaga(ctx context.Context, saga *SagaTransaction) error

	// GetSaga 获取Saga事务
	GetSaga(ctx context.Context, sagaID string) (*SagaTransaction, error)

	// ListSagas 列出Saga事务
	ListSagas(ctx context.Context, filter *SagaFilter) ([]*SagaTransaction, error)

	// CancelSaga 取消Saga事务
	CancelSaga(ctx context.Context, sagaID string, reason string) error

	// GetStats 获取统计信息
	GetStats(ctx context.Context) (*SagaStats, error)

	// CleanupCompletedSagas 清理已完成的Saga
	CleanupCompletedSagas(ctx context.Context, olderThan time.Duration) error
}

// SagaFilter Saga过滤器
type SagaFilter struct {
	Status    *SagaStatus `json:"status,omitempty"`
	StartTime *time.Time  `json:"start_time,omitempty"`
	EndTime   *time.Time  `json:"end_time,omitempty"`
	Limit     int         `json:"limit,omitempty"`
	Offset    int         `json:"offset,omitempty"`
}

// SagaStats Saga统计信息
type SagaStats struct {
	TotalSagas       int64   `json:"total_sagas"`
	RunningSagas     int64   `json:"running_sagas"`
	CompletedSagas   int64   `json:"completed_sagas"`
	FailedSagas      int64   `json:"failed_sagas"`
	CompensatedSagas int64   `json:"compensated_sagas"`
	CancelledSagas   int64   `json:"cancelled_sagas"`
	AverageDuration  float64 `json:"average_duration_ms"`
	SuccessRate      float64 `json:"success_rate"`
}

// DefaultSagaCoordinator 默认Saga协调器实现
type DefaultSagaCoordinator struct {
	transactions map[string]*SagaTransaction
	mu           sync.RWMutex
	logger       *logger.Logger
	config       *SagaConfig
	storage      SagaStorage
	eventBus     EventBus
	stats        *SagaStats
}

// SagaConfig Saga配置
type SagaConfig struct {
	MaxConcurrentSagas    int           `yaml:"max_concurrent_sagas"`
	DefaultRetryAttempts  int           `yaml:"default_retry_attempts"`
	DefaultRetryDelay     time.Duration `yaml:"default_retry_delay"`
	TimeoutDuration       time.Duration `yaml:"timeout_duration"`
	CompensationTimeout   time.Duration `yaml:"compensation_timeout"`
	EnablePersistence     bool          `yaml:"enable_persistence"`
	EnableEventPublishing bool          `yaml:"enable_event_publishing"`
	CleanupInterval       time.Duration `yaml:"cleanup_interval"`
}

// DefaultSagaConfig 默认Saga配置
func DefaultSagaConfig() *SagaConfig {
	return &SagaConfig{
		MaxConcurrentSagas:    100,
		DefaultRetryAttempts:  3,
		DefaultRetryDelay:     1 * time.Second,
		TimeoutDuration:       30 * time.Minute,
		CompensationTimeout:   10 * time.Minute,
		EnablePersistence:     true,
		EnableEventPublishing: true,
		CleanupInterval:       1 * time.Hour,
	}
}

// NewDefaultSagaCoordinator 创建默认Saga协调器
func NewDefaultSagaCoordinator(config *SagaConfig, storage SagaStorage, eventBus EventBus, logger *logger.Logger) *DefaultSagaCoordinator {
	if config == nil {
		config = DefaultSagaConfig()
	}

	coordinator := &DefaultSagaCoordinator{
		transactions: make(map[string]*SagaTransaction),
		logger:       logger,
		config:       config,
		storage:      storage,
		eventBus:     eventBus,
		stats:        &SagaStats{},
	}

	// 启动清理协程
	go coordinator.startCleanupRoutine()

	return coordinator
}

// ExecuteSaga 执行Saga事务
func (c *DefaultSagaCoordinator) ExecuteSaga(ctx context.Context, saga *SagaTransaction) error {
	c.logger.Info("开始执行Saga事务",
		map[string]interface{}{
			"saga_id":     saga.ID,
			"saga_name":   saga.Name,
			"steps_count": len(saga.Steps),
		})

	// 设置Saga初始状态
	saga.Status = SagaStatusRunning
	saga.StartTime = time.Now()
	saga.CurrentStep = 0

	// 持久化Saga
	if err := c.storage.SaveSaga(ctx, saga); err != nil {
		c.logger.Error("持久化Saga失败",
			map[string]interface{}{
				"saga_id": saga.ID,
				"error":   err,
			})
		return fmt.Errorf("持久化Saga失败: %w", err)
	}

	// 发布Saga开始事件
	if c.config.EnableEventPublishing {
		c.publishSagaEvent(ctx, SagaEventStarted, saga)
	}

	// 执行所有步骤
	for i, step := range saga.Steps {
		saga.CurrentStep = i
		step.Status = StepStatusRunning
		step.StartTime = time.Now()

		c.logger.Debug("执行Saga步骤",
			map[string]interface{}{
				"saga_id":   saga.ID,
				"step_id":   step.ID,
				"step_name": step.Name,
			})

		// 持久化步骤状态
		if err := c.storage.SaveSaga(ctx, saga); err != nil {
			c.logger.Error("持久化Saga步骤状态失败",
				map[string]interface{}{
					"saga_id": saga.ID,
					"step_id": step.ID,
					"error":   err,
				})
		}

		// 发布步骤开始事件
		if c.config.EnableEventPublishing {
			c.publishStepEvent(ctx, SagaStepStarted, saga, step)
		}

		// 执行步骤
		result, err := c.executeStepWithRetry(ctx, step)
		if err != nil {
			c.logger.Error("Saga步骤执行失败",
				map[string]interface{}{
					"saga_id":   saga.ID,
					"step_id":   step.ID,
					"step_name": step.Name,
					"error":     err,
				})

			// 标记步骤失败
			step.Status = StepStatusFailed
			step.Error = err
			saga.FailedSteps++

			// 执行补偿
			compensateErr := c.compensate(ctx, saga, i)
			if compensateErr != nil {
				c.logger.Error("Saga补偿失败",
					map[string]interface{}{
						"saga_id":          saga.ID,
						"failed_step":      i,
						"compensate_error": compensateErr,
					})
			}

			saga.Status = SagaStatusFailed
			now := time.Now()
			saga.EndTime = &now

			// 持久化最终状态
			_ = c.storage.SaveSaga(ctx, saga)

			// 发布失败事件
			if c.config.EnableEventPublishing {
				c.publishSagaEvent(ctx, SagaEventFailed, saga)
			}

			return fmt.Errorf("Saga执行失败，步骤[%d] %s: %w", i, step.Name, err)
		}

		// 标记步骤成功
		step.Status = StepStatusCompleted
		step.Result = result
		saga.CompletedSteps++
		now := time.Now()
		step.EndTime = &now

		c.logger.Debug("Saga步骤执行成功",
			map[string]interface{}{
				"saga_id":   saga.ID,
				"step_id":   step.ID,
				"step_name": step.Name,
				"result":    result,
			})

		// 发布步骤完成事件
		if c.config.EnableEventPublishing {
			c.publishStepEvent(ctx, SagaStepCompleted, saga, step)
		}
	}

	// 所有步骤执行成功
	saga.Status = SagaStatusCompleted
	now := time.Now()
	saga.EndTime = &now

	c.logger.Info("Saga事务执行成功",
		map[string]interface{}{
			"saga_id":         saga.ID,
			"saga_name":       saga.Name,
			"completed_steps": saga.CompletedSteps,
			"duration":        saga.EndTime.Sub(saga.StartTime),
		})

	// 持久化最终状态
	if err := c.storage.SaveSaga(ctx, saga); err != nil {
		c.logger.Error("持久化Saga最终状态失败",
			map[string]interface{}{
				"saga_id": saga.ID,
				"error":   err,
			})
	}

	// 发布完成事件
	if c.config.EnableEventPublishing {
		c.publishSagaEvent(ctx, SagaEventCompleted, saga)
	}

	// 更新统计
	c.updateStats(saga)

	return nil
}

// executeStepWithRetry 执行步骤（带重试）
func (c *DefaultSagaCoordinator) executeStepWithRetry(ctx context.Context, step *SagaStep) (interface{}, error) {
	retryPolicy := step.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = &RetryPolicy{
			MaxAttempts: c.config.DefaultRetryAttempts,
			Delay:       c.config.DefaultRetryDelay,
			Backoff:     BackoffExponential,
		}
	}

	var lastErr error

	for attempt := 1; attempt <= retryPolicy.MaxAttempts; attempt++ {
		if attempt > 1 {
			delay := c.calculateRetryDelay(retryPolicy, attempt)
			c.logger.Debug("重试Saga步骤",
				map[string]interface{}{
					"step_id":      step.ID,
					"attempt":      attempt,
					"delay":        delay,
					"max_attempts": retryPolicy.MaxAttempts,
				})
			time.Sleep(delay)
		}

		result, err := step.Execute(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		c.logger.Debug("Saga步骤执行失败，将重试",
			map[string]interface{}{
				"step_id":   step.ID,
				"attempt":   attempt,
				"error":     err,
				"max_retry": retryPolicy.MaxAttempts,
			})
	}

	return nil, lastErr
}

// compensate 补偿已执行的步骤
func (c *DefaultSagaCoordinator) compensate(ctx context.Context, saga *SagaTransaction, failedStep int) error {
	c.logger.Info("开始Saga补偿",
		map[string]interface{}{
			"saga_id":         saga.ID,
			"failed_step":     failedStep,
			"completed_steps": saga.CompletedSteps,
		})

	saga.Status = SagaStatusCompensating

	// 发布补偿开始事件
	if c.config.EnableEventPublishing {
		c.publishSagaEvent(ctx, SagaEventCompensationStarted, saga)
	}

	// 从失败步骤开始向前补偿
	for i := failedStep; i >= 0; i-- {
		step := saga.Steps[i]
		if step.Status != StepStatusCompleted {
			continue // 只补偿已完成的步骤
		}

		c.logger.Debug("补偿Saga步骤",
			map[string]interface{}{
				"saga_id":   saga.ID,
				"step_id":   step.ID,
				"step_name": step.Name,
			})

		// 执行补偿操作
		err := c.compensateStepWithRetry(ctx, step)
		if err != nil {
			c.logger.Error("Saga步骤补偿失败",
				map[string]interface{}{
					"saga_id":   saga.ID,
					"step_id":   step.ID,
					"step_name": step.Name,
					"error":     err,
				})
			// 记录补偿失败但继续执行其他补偿
			continue
		}

		// 标记步骤已补偿
		step.Status = StepStatusCompensated
	}

	saga.Status = SagaStatusCompensated
	now := time.Now()
	saga.EndTime = &now

	// 发布补偿完成事件
	if c.config.EnableEventPublishing {
		c.publishSagaEvent(ctx, SagaEventCompensationCompleted, saga)
	}

	// 更新统计
	c.updateStats(saga)

	return nil
}

// compensateStepWithRetry 补偿步骤（带重试）
func (c *DefaultSagaCoordinator) compensateStepWithRetry(ctx context.Context, step *SagaStep) error {
	if step.Compensate == nil {
		c.logger.Warn("步骤没有补偿操作",
			map[string]interface{}{
				"step_id":   step.ID,
				"step_name": step.Name,
			})
		return nil
	}

	retryPolicy := step.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = &RetryPolicy{
			MaxAttempts: c.config.DefaultRetryAttempts,
			Delay:       c.config.DefaultRetryDelay,
			Backoff:     BackoffExponential,
		}
	}

	// 补偿的超时时间更短
	compensateCtx, cancel := context.WithTimeout(ctx, c.config.CompensationTimeout)
	defer cancel()

	var err error
	for attempt := 1; attempt <= retryPolicy.MaxAttempts; attempt++ {
		if attempt > 1 {
			delay := c.calculateRetryDelay(retryPolicy, attempt)
			c.logger.Debug("重试补偿步骤",
				map[string]interface{}{
					"step_id": step.ID,
					"attempt": attempt,
					"delay":   delay,
				})
			time.Sleep(delay)
		}

		_, compensateErr := step.Compensate(compensateCtx)
		if compensateErr == nil {
			return nil
		}

		c.logger.Debug("补偿步骤执行失败",
			map[string]interface{}{
				"step_id": step.ID,
				"attempt": attempt,
				"error":   compensateErr,
			})
		err = compensateErr
	}

	return fmt.Errorf("补偿步骤失败: %w", err)
}

// GetSaga 获取Saga事务
func (c *DefaultSagaCoordinator) GetSaga(ctx context.Context, sagaID string) (*SagaTransaction, error) {
	if c.config.EnablePersistence {
		return c.storage.GetSaga(ctx, sagaID)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if saga, exists := c.transactions[sagaID]; exists {
		// 返回副本避免外部修改，不包含互斥锁
		sagaCopy := &SagaTransaction{
			ID:             saga.ID,
			Name:           saga.Name,
			Description:    saga.Description,
			Steps:          make([]*SagaStep, len(saga.Steps)),
			Status:         saga.Status,
			CurrentStep:    saga.CurrentStep,
			Payload:        saga.Payload,
			CompletedSteps: saga.CompletedSteps,
			FailedSteps:    saga.FailedSteps,
			StartTime:      saga.StartTime,
			EndTime:        saga.EndTime,
			CompensateAll:  saga.CompensateAll,
			CreatedAt:      saga.CreatedAt,
			UpdatedAt:      saga.UpdatedAt,
		}
		// 深拷贝步骤
		for i, step := range saga.Steps {
			stepCopy := *step
			sagaCopy.Steps[i] = &stepCopy
		}
		return sagaCopy, nil
	}

	return nil, fmt.Errorf("Saga不存在: %s", sagaID)
}

// ListSagas 列出Saga事务
func (c *DefaultSagaCoordinator) ListSagas(ctx context.Context, filter *SagaFilter) ([]*SagaTransaction, error) {
	if c.config.EnablePersistence {
		return c.storage.ListSagas(ctx, filter)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	var sagas []*SagaTransaction
	for _, saga := range c.transactions {
		// 应用过滤器
		if filter != nil {
			if filter.Status != nil && saga.Status != *filter.Status {
				continue
			}
			if filter.StartTime != nil && saga.StartTime.Before(*filter.StartTime) {
				continue
			}
			if filter.EndTime != nil && saga.EndTime != nil && saga.EndTime.After(*filter.EndTime) {
				continue
			}
		}
		// 创建副本避免外部修改，不包含互斥锁
		sagaCopy := &SagaTransaction{
			ID:             saga.ID,
			Name:           saga.Name,
			Description:    saga.Description,
			Steps:          make([]*SagaStep, len(saga.Steps)),
			Status:         saga.Status,
			CurrentStep:    saga.CurrentStep,
			Payload:        saga.Payload,
			CompletedSteps: saga.CompletedSteps,
			FailedSteps:    saga.FailedSteps,
			StartTime:      saga.StartTime,
			EndTime:        saga.EndTime,
			CompensateAll:  saga.CompensateAll,
			CreatedAt:      saga.CreatedAt,
			UpdatedAt:      saga.UpdatedAt,
		}
		// 深拷贝步骤
		for i, step := range saga.Steps {
			stepCopy := *step
			sagaCopy.Steps[i] = &stepCopy
		}
		sagas = append(sagas, sagaCopy)
	}

	return sagas, nil
}

// CancelSaga 取消Saga事务
func (c *DefaultSagaCoordinator) CancelSaga(ctx context.Context, sagaID string, reason string) error {
	saga, err := c.GetSaga(ctx, sagaID)
	if err != nil {
		return fmt.Errorf("获取Saga失败: %w", err)
	}

	if saga.Status == SagaStatusCompleted || saga.Status == SagaStatusFailed || saga.Status == SagaStatusCompensated {
		return fmt.Errorf("Saga已完成，无法取消")
	}

	c.logger.Info("取消Saga事务",
		map[string]interface{}{
			"saga_id": sagaID,
			"reason":  reason,
		})

	saga.Status = SagaStatusCancelled
	now := time.Now()
	saga.EndTime = &now

	// 持久化状态
	if err := c.storage.SaveSaga(ctx, saga); err != nil {
		c.logger.Error("持久化取消Saga状态失败",
			map[string]interface{}{
				"saga_id": sagaID,
				"error":   err,
			})
	}

	// 发布取消事件
	if c.config.EnableEventPublishing {
		c.publishSagaEvent(ctx, SagaEventCancelled, saga)
	}

	return nil
}

// GetStats 获取统计信息
func (c *DefaultSagaCoordinator) GetStats(ctx context.Context) (*SagaStats, error) {
	if c.config.EnablePersistence {
		return c.storage.GetStats(ctx)
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats, nil
}

// CleanupCompletedSagas 清理已完成的Saga
func (c *DefaultSagaCoordinator) CleanupCompletedSagas(ctx context.Context, olderThan time.Duration) error {
	c.logger.Info("清理已完成的Saga",
		map[string]interface{}{
			"older_than": olderThan,
		})

	sagas, err := c.ListSagas(ctx, &SagaFilter{})
	if err != nil {
		return fmt.Errorf("获取Saga列表失败: %w", err)
	}

	cutoff := time.Now().Add(-olderThan)
	cleanedCount := 0

	for _, saga := range sagas {
		if saga.EndTime != nil && saga.EndTime.Before(cutoff) {
			// 删除内存中的Saga
			c.mu.Lock()
			delete(c.transactions, saga.ID)
			c.mu.Unlock()

			// 删除持久化的Saga
			if c.config.EnablePersistence {
				if err := c.storage.DeleteSaga(ctx, saga.ID); err != nil {
					c.logger.Error("删除持久化Saga失败",
						map[string]interface{}{
							"saga_id": saga.ID,
							"error":   err,
						})
					continue
				}
			}

			cleanedCount++
		}
	}

	c.logger.Info("Saga清理完成",
		map[string]interface{}{
			"cleaned_count": cleanedCount,
			"cutoff_time":   cutoff,
		})

	return nil
}

// publishSagaEvent 发布Saga事件
func (c *DefaultSagaCoordinator) publishSagaEvent(ctx context.Context, eventType SagaEventType, saga *SagaTransaction) {
	if c.eventBus == nil {
		return
	}

	event := &SagaEvent{
		ID:        generateSagaEventID(),
		Type:      eventType,
		SagaID:    saga.ID,
		SagaName:  saga.Name,
		Status:    saga.Status,
		Payload:   saga,
		Timestamp: time.Now(),
	}

	c.eventBus.Publish(ctx, event)
}

// publishStepEvent 发布步骤事件
func (c *DefaultSagaCoordinator) publishStepEvent(ctx context.Context, eventType SagaStepEventType, saga *SagaTransaction, step *SagaStep) {
	if c.eventBus == nil {
		return
	}

	event := &SagaStepEvent{
		ID:        generateSagaEventID(),
		Type:      eventType,
		SagaID:    saga.ID,
		SagaName:  saga.Name,
		StepID:    step.ID,
		StepName:  step.Name,
		Status:    step.Status,
		Result:    step.Result,
		Error:     step.Error,
		Payload:   step,
		Timestamp: time.Now(),
	}

	c.eventBus.Publish(ctx, event)
}

// calculateRetryDelay 计算重试延迟
func (c *DefaultSagaCoordinator) calculateRetryDelay(policy *RetryPolicy, attempt int) time.Duration {
	switch policy.Backoff {
	case BackoffFixed:
		return policy.Delay
	case BackoffLinear:
		return time.Duration(attempt) * policy.Delay
	case BackoffExponential:
		delay := policy.Delay
		for i := 1; i < attempt; i++ {
			delay *= 2
		}
		return delay
	default:
		return policy.Delay
	}
}

// updateStats 更新统计信息
func (c *DefaultSagaCoordinator) updateStats(saga *SagaTransaction) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats.TotalSagas++
	c.stats.SuccessRate = float64(c.stats.TotalSagas-c.stats.FailedSagas-c.stats.CancelledSagas) / float64(c.stats.TotalSagas) * 100

	switch saga.Status {
	case SagaStatusCompleted:
		c.stats.CompletedSagas++
	case SagaStatusFailed:
		c.stats.FailedSagas++
	case SagaStatusCompensated:
		c.stats.CompensatedSagas++
	case SagaStatusCancelled:
		c.stats.CancelledSagas++
	}

	// 计算平均执行时间
	if saga.EndTime != nil {
		duration := float64(saga.EndTime.Sub(saga.StartTime).Milliseconds())
		c.stats.AverageDuration = (c.stats.AverageDuration + duration) / 2
	}
}

// startCleanupRoutine 启动清理协程
func (c *DefaultSagaCoordinator) startCleanupRoutine() {
	if c.config.CleanupInterval <= 0 {
		return
	}

	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.CleanupCompletedSagas(context.Background(), 24*time.Hour)
		}
	}
}

// generateSagaEventID 生成Saga事件ID
func generateSagaEventID() string {
	return fmt.Sprintf("saga-event-%s", uuid.New().String())
}

// EventBus 事件总线接口
type EventBus interface {
	Publish(ctx context.Context, event interface{}) error
}
