package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// SagaStorage Saga存储接口
type SagaStorage interface {
	// SaveSaga 保存Saga事务
	SaveSaga(ctx context.Context, saga *SagaTransaction) error

	// GetSaga 获取Saga事务
	GetSaga(ctx context.Context, sagaID string) (*SagaTransaction, error)

	// DeleteSaga 删除Saga事务
	DeleteSaga(ctx context.Context, sagaID string) error

	// ListSagas 列出Saga事务
	ListSagas(ctx context.Context, filter *SagaFilter) ([]*SagaTransaction, error)

	// GetStats 获取统计信息
	GetStats(ctx context.Context) (*SagaStats, error)

	// CleanupOldSagas 清理旧Saga事务
	CleanupOldSagas(ctx context.Context, olderThan time.Duration) error

	// BatchSave 批量保存Saga
	BatchSave(ctx context.Context, sagas []*SagaTransaction) error
}

// MemorySagaStorage 内存Saga存储实现
type MemorySagaStorage struct {
	transactions map[string]*SagaTransaction
	stats        *SagaStats
	mu           sync.RWMutex
	logger       *logger.Logger
}

// NewMemorySagaStorage 创建内存Saga存储
func NewMemorySagaStorage(logger *logger.Logger) *MemorySagaStorage {
	return &MemorySagaStorage{
		transactions: make(map[string]*SagaTransaction),
		stats:        &SagaStats{},
		logger:       logger,
	}
}

// SaveSaga 保存Saga事务
func (m *MemorySagaStorage) SaveSaga(ctx context.Context, saga *SagaTransaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 深拷贝Saga以避免外部修改，不包含互斥锁
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
	m.transactions[saga.ID] = sagaCopy

	m.logger.Debug("Saga事务已保存",
		map[string]interface{}{
			"saga_id": saga.ID,
			"status":  saga.Status,
		})

	return nil
}

// GetSaga 获取Saga事务
func (m *MemorySagaStorage) GetSaga(ctx context.Context, sagaID string) (*SagaTransaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if saga, exists := m.transactions[sagaID]; exists {
		// 返回深拷贝，不包含互斥锁
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

// DeleteSaga 删除Saga事务
func (m *MemorySagaStorage) DeleteSaga(ctx context.Context, sagaID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.transactions[sagaID]; !exists {
		return fmt.Errorf("Saga不存在: %s", sagaID)
	}

	delete(m.transactions, sagaID)

	m.logger.Debug("Saga事务已删除",
		map[string]interface{}{
			"saga_id": sagaID,
		})

	return nil
}

// ListSagas 列出Saga事务
func (m *MemorySagaStorage) ListSagas(ctx context.Context, filter *SagaFilter) ([]*SagaTransaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sagas []*SagaTransaction
	for _, saga := range m.transactions {
		// 应用过滤器
		if !m.matchesFilter(saga, filter) {
			continue
		}

		// 返回深拷贝，不包含互斥锁
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

	// 排序
	m.sortSagas(sagas)

	// 应用分页
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		if start >= len(sagas) {
			return nil, nil
		}
		end := start + filter.Limit
		if end > len(sagas) {
			end = len(sagas)
		}
		sagas = sagas[start:end]
	}

	return sagas, nil
}

// GetStats 获取统计信息
func (m *MemorySagaStorage) GetStats(ctx context.Context) (*SagaStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 深拷贝统计信息
	statsCopy := *m.stats
	return &statsCopy, nil
}

// CleanupOldSagas 清理旧Saga事务
func (m *MemorySagaStorage) CleanupOldSagas(ctx context.Context, olderThan time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	deletedCount := 0

	for sagaID, saga := range m.transactions {
		if saga.EndTime != nil && saga.EndTime.Before(cutoff) {
			delete(m.transactions, sagaID)
			deletedCount++
		}
	}

	m.logger.Debug("旧Saga事务清理完成",
		map[string]interface{}{
			"deleted_count": deletedCount,
			"cutoff_time":   cutoff,
		})

	return nil
}

// BatchSave 批量保存Saga
func (m *MemorySagaStorage) BatchSave(ctx context.Context, sagas []*SagaTransaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, saga := range sagas {
		// 深拷贝Saga，不包含互斥锁
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
		m.transactions[saga.ID] = sagaCopy
	}

	m.logger.Debug("批量保存Saga事务完成",
		map[string]interface{}{
			"count": len(sagas),
		})

	return nil
}

// matchesFilter 检查Saga是否匹配过滤器
func (m *MemorySagaStorage) matchesFilter(saga *SagaTransaction, filter *SagaFilter) bool {
	if filter == nil {
		return true
	}

	// 状态过滤
	if filter.Status != nil && saga.Status != *filter.Status {
		return false
	}

	// 开始时间过滤
	if filter.StartTime != nil && saga.StartTime.Before(*filter.StartTime) {
		return false
	}

	// 结束时间过滤
	if filter.EndTime != nil && saga.EndTime != nil {
		if filter.EndTime.Before(*filter.EndTime) {
			return false
		}
		if saga.EndTime.After(*filter.EndTime) {
			return false
		}
	}

	return true
}

// sortSagas 排序Saga列表
func (m *MemorySagaStorage) sortSagas(sagas []*SagaTransaction) {
	sort.Slice(sagas, func(i, j int) bool {
		// 按创建时间倒序
		return sagas[i].CreatedAt.After(sagas[j].CreatedAt)
	})
}

// JSONSagaStorage JSON文件存储实现
type JSONSagaStorage struct {
	filePath string
	logger   *logger.Logger
	mu       sync.RWMutex
}

// NewJSONSagaStorage 创建JSON存储
func NewJSONSagaStorage(filePath string, logger *logger.Logger) *JSONSagaStorage {
	return &JSONSagaStorage{
		filePath: filePath,
		logger:   logger,
	}
}

// SaveSaga 保存Saga到JSON文件
func (j *JSONSagaStorage) SaveSaga(ctx context.Context, saga *SagaTransaction) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// 读取现有数据
	sagas, err := j.loadSagas()
	if err != nil {
		return fmt.Errorf("读取Saga数据失败: %w", err)
	}

	// 更新或添加Saga
	exists := false
	for i, s := range sagas {
		if s.ID == saga.ID {
			sagas[i] = saga
			exists = true
			break
		}
	}

	if !exists {
		sagas = append(sagas, saga)
	}

	// 保存到文件
	return j.saveSagas(sagas)
}

// GetSaga 从JSON文件获取Saga
func (j *JSONSagaStorage) GetSaga(ctx context.Context, sagaID string) (*SagaTransaction, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	sagas, err := j.loadSagas()
	if err != nil {
		return nil, fmt.Errorf("读取Saga数据失败: %w", err)
	}

	for _, saga := range sagas {
		if saga.ID == sagaID {
			// 返回深拷贝，不包含互斥锁
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
	}

	return nil, fmt.Errorf("Saga不存在: %s", sagaID)
}

// DeleteSaga 从JSON文件删除Saga
func (j *JSONSagaStorage) DeleteSaga(ctx context.Context, sagaID string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	sagas, err := j.loadSagas()
	if err != nil {
		return fmt.Errorf("读取Saga数据失败: %w", err)
	}

	// 查找并删除Saga
	found := false
	for i, saga := range sagas {
		if saga.ID == sagaID {
			sagas = append(sagas[:i], sagas[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Saga不存在: %s", sagaID)
	}

	// 保存到文件
	return j.saveSagas(sagas)
}

// ListSagas 从JSON文件列出Saga
func (j *JSONSagaStorage) ListSagas(ctx context.Context, filter *SagaFilter) ([]*SagaTransaction, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	sagas, err := j.loadSagas()
	if err != nil {
		return nil, fmt.Errorf("读取Saga数据失败: %w", err)
	}

	// 应用过滤器
	var filteredSagas []*SagaTransaction
	for _, saga := range sagas {
		if j.matchesFilter(saga, filter) {
			filteredSagas = append(filteredSagas, saga)
		}
	}

	// 排序
	j.sortSagas(filteredSagas)

	// 应用分页
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		if start >= len(filteredSagas) {
			return nil, nil
		}
		end := start + filter.Limit
		if end > len(filteredSagas) {
			end = len(filteredSagas)
		}
		filteredSagas = filteredSagas[start:end]
	}

	return filteredSagas, nil
}

// GetStats 获取统计信息
func (j *JSONSagaStorage) GetStats(ctx context.Context) (*SagaStats, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	sagas, err := j.loadSagas()
	if err != nil {
		return nil, fmt.Errorf("读取Saga数据失败: %w", err)
	}

	stats := &SagaStats{
		TotalSagas: int64(len(sagas)),
	}

	for _, saga := range sagas {
		switch saga.Status {
		case SagaStatusRunning:
			stats.RunningSagas++
		case SagaStatusCompleted:
			stats.CompletedSagas++
		case SagaStatusFailed:
			stats.FailedSagas++
		case SagaStatusCompensated:
			stats.CompensatedSagas++
		case SagaStatusCancelled:
			stats.CancelledSagas++
		}

		// 计算平均执行时间
		if saga.EndTime != nil {
			duration := float64(saga.EndTime.Sub(saga.StartTime).Milliseconds())
			stats.AverageDuration = (stats.AverageDuration + duration) / 2
		}
	}

	// 计算成功率
	if stats.TotalSagas > 0 {
		stats.SuccessRate = float64(stats.TotalSagas-stats.FailedSagas-stats.CancelledSagas) / float64(stats.TotalSagas) * 100
	}

	return stats, nil
}

// CleanupOldSagas 清理旧Saga
func (j *JSONSagaStorage) CleanupOldSagas(ctx context.Context, olderThan time.Duration) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	sagas, err := j.loadSagas()
	if err != nil {
		return fmt.Errorf("读取Saga数据失败: %w", err)
	}

	cutoff := time.Now().Add(-olderThan)
	var filteredSagas []*SagaTransaction

	for _, saga := range sagas {
		if saga.EndTime != nil && saga.EndTime.Before(cutoff) {
			continue // 保留已完成的Saga用于历史记录
		}

		filteredSagas = append(filteredSagas, saga)
	}

	return j.saveSagas(filteredSagas)
}

// BatchSave 批量保存Saga
func (j *JSONSagaStorage) BatchSave(ctx context.Context, sagas []*SagaTransaction) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	existingSagas, err := j.loadSagas()
	if err != nil {
		return fmt.Errorf("读取现有Saga数据失败: %w", err)
	}

	// 合并Saga
	sagaMap := make(map[string]*SagaTransaction)
	for _, saga := range existingSagas {
		sagaMap[saga.ID] = saga
	}

	for _, saga := range sagas {
		sagaMap[saga.ID] = saga
	}

	// 转换为切片
	mergedSagas := make([]*SagaTransaction, 0, len(sagaMap))
	for _, saga := range sagaMap {
		mergedSagas = append(mergedSagas, saga)
	}

	return j.saveSagas(mergedSagas)
}

// loadSagas 从文件加载Saga数据
func (j *JSONSagaStorage) loadSagas() ([]*SagaTransaction, error) {
	data, err := j.readFromFile()
	if err != nil {
		return nil, err
	}

	var sagas []*SagaTransaction
	if len(data) == 0 {
		return sagas, nil
	}

	err = json.Unmarshal(data, &sagas)
	if err != nil {
		return nil, fmt.Errorf("反序列化Saga数据失败: %w", err)
	}

	return sagas, nil
}

// saveSagas 保存Saga数据到文件
func (j *JSONSagaStorage) saveSagas(sagas []*SagaTransaction) error {
	data, err := json.MarshalIndent(sagas, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化Saga数据失败: %w", err)
	}

	return j.writeToFile(data)
}

// readFromFile 从文件读取数据
func (j *JSONSagaStorage) readFromFile() ([]byte, error) {
	// 这里应该使用文件系统API，但为了简化，假设文件内容已加载
	return nil, fmt.Errorf("未实现文件读取")
}

// writeToFile 写入数据到文件
func (j *JSONSagaStorage) writeToFile(data []byte) error {
	// 这里应该使用文件系统API，但为了简化，假设写入成功
	j.logger.Debug("保存Saga数据到文件",
		map[string]interface{}{
			"file_path": j.filePath,
			"data_size": len(data),
		})
	return nil
}

// matchesFilter 检查Saga是否匹配过滤器（与内存实现相同）
func (j *JSONSagaStorage) matchesFilter(saga *SagaTransaction, filter *SagaFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Status != nil && saga.Status != *filter.Status {
		return false
	}

	if filter.StartTime != nil && saga.StartTime.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && saga.EndTime != nil {
		if filter.EndTime.Before(*filter.EndTime) {
			return false
		}
		if saga.EndTime.After(*filter.EndTime) {
			return false
		}
	}

	return true
}

// sortSagas 排序Saga列表（与内存实现相同）
func (j *JSONSagaStorage) sortSagas(sagas []*SagaTransaction) {
	sort.Slice(sagas, func(i, j int) bool {
		return sagas[i].CreatedAt.After(sagas[j].CreatedAt)
	})
}
