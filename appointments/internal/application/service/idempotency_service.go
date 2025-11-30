package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/appointments/pkg/idempotency"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// IdempotentService 带幂等性保证的服务
type IdempotentService struct {
	baseService        AppointmentService
	idempotencyManager idempotency.IdempotencyManager
	logger             *logger.Logger
	config             *IdempotencyServiceConfig
}

// IdempotencyServiceConfig 幂等性服务配置
type IdempotencyServiceConfig struct {
	DefaultTTL    time.Duration `yaml:"default_ttl"`
	EnableCache   bool          `yaml:"enable_cache"`
	EnableRetry   bool          `yaml:"enable_retry"`
	MaxRetryCount int           `yaml:"max_retry_count"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
}

// DefaultIdempotencyServiceConfig 默认配置
func DefaultIdempotencyServiceConfig() *IdempotencyServiceConfig {
	return &IdempotencyServiceConfig{
		DefaultTTL:    24 * time.Hour,
		EnableCache:   true,
		EnableRetry:   true,
		MaxRetryCount: 3,
		RetryDelay:    1 * time.Second,
	}
}

// NewIdempotentService 创建带幂等性保证的服务
func NewIdempotentService(
	baseService AppointmentService,
	idempotencyManager idempotency.IdempotencyManager,
	config *IdempotencyServiceConfig,
	logger *logger.Logger,
) *IdempotentService {
	if config == nil {
		config = DefaultIdempotencyServiceConfig()
	}

	return &IdempotentService{
		baseService:        baseService,
		idempotencyManager: idempotencyManager,
		logger:             logger,
		config:             config,
	}
}

// CreateAppointmentWithIdempotency 创建带幂等性保证的预约
func (s *IdempotentService) CreateAppointmentWithIdempotency(
	ctx context.Context,
	req *dto.CreateAppointmentRequest,
	idempotencyKey string,
) (*dto.AppointmentResponse, error) {

	// 生成完整的幂等性键
	fullKey := s.generateKey("appointment", "create", idempotencyKey)

	s.logger.Debug("开始创建幂等性预约",
		map[string]interface{}{
			"idempotency_key": idempotencyKey,
			"full_key":        fullKey,
			"customer_id":     req.CustomerID,
		})

	// 检查幂等性
	isFirst, err := s.idempotencyManager.CheckAndRecord(ctx, fullKey, s.config.DefaultTTL)
	if err != nil {
		s.logger.Error("幂等性检查失败",
			map[string]interface{}{
				"full_key": fullKey,
				"error":    err,
			})
		return nil, fmt.Errorf("幂等性检查失败: %w", err)
	}

	// 如果不是第一次请求，返回缓存的结果
	if !isFirst {
		return s.getCachedResult(ctx, fullKey)
	}

	// 执行实际的业务逻辑
	response, err := s.executeCreateAppointment(ctx, req)
	if err != nil {
		// 如果业务逻辑失败，删除幂等性键以便重试
		if s.isRetryableError(err) {
			s.logger.Warn("业务逻辑失败但可重试，删除幂等性键",
				map[string]interface{}{
					"full_key": fullKey,
					"error":    err,
				})
			_ = s.idempotencyManager.Delete(ctx, fullKey)
		}
		return nil, err
	}

	// 保存结果到缓存
	if err := s.saveResult(ctx, fullKey, response); err != nil {
		s.logger.Error("保存幂等性结果失败",
			map[string]interface{}{
				"full_key": fullKey,
				"error":    err,
			})
		// 不返回错误，因为业务逻辑已成功
	}

	return response, nil
}

// executeCreateAppointment 执行实际的预约创建逻辑
func (s *IdempotentService) executeCreateAppointment(ctx context.Context, req *dto.CreateAppointmentRequest) (*dto.AppointmentResponse, error) {
	s.logger.Debug("执行预约创建逻辑",
		map[string]interface{}{
			"customer_id": req.CustomerID,
			"staff_id":    req.StaffID,
			"service_id":  req.ServiceID,
		})

	// 调用基础服务创建预约
	appointment, err := s.baseService.CreateAppointment(req)
	if err != nil {
		s.logger.Error("预约创建失败",
			map[string]interface{}{
				"customer_id": req.CustomerID,
				"error":       err,
			})
		return nil, err
	}

	s.logger.Info("预约创建成功",
		map[string]interface{}{
			"appointment_id": appointment.ID,
			"status":         appointment.Status,
		})

	// 转换为响应格式
	response := &dto.AppointmentResponse{
		ID:     appointment.ID.String(),
		Status: appointment.Status,
	}
	return response, nil
}

// getCachedResult 获取缓存结果
func (s *IdempotentService) getCachedResult(ctx context.Context, key string) (*dto.AppointmentResponse, error) {
	result, err := s.idempotencyManager.GetResult(ctx, key)
	if err != nil {
		s.logger.Error("获取缓存结果失败",
			map[string]interface{}{
				"key":   key,
				"error": err,
			})
		return nil, fmt.Errorf("获取缓存结果失败: %w", err)
	}

	if result == nil {
		s.logger.Info("请求正在处理中", "key", key)
		return nil, fmt.Errorf("请求正在处理中")
	}

	// 尝试将结果转换为AppointmentResponse
	response, ok := result.(*dto.AppointmentResponse)
	if !ok {
		// 尝试从JSON转换
		return s.convertResult(result)
	}

	s.logger.Info("返回缓存结果",
		map[string]interface{}{
			"key":            key,
			"appointment_id": response.ID,
			"status":         response.Status,
		})

	return response, nil
}

// saveResult 保存结果到缓存
func (s *IdempotentService) saveResult(ctx context.Context, key string, response *dto.AppointmentResponse) error {
	if !s.config.EnableCache {
		return nil
	}

	err := s.idempotencyManager.SaveResult(ctx, key, response, s.config.DefaultTTL)
	if err != nil {
		return fmt.Errorf("保存结果失败: %w", err)
	}

	s.logger.Debug("结果保存成功",
		map[string]interface{}{
			"key":            key,
			"appointment_id": response.ID,
			"ttl":            s.config.DefaultTTL,
		})

	return nil
}

// convertResult 转换结果格式
func (s *IdempotentService) convertResult(result interface{}) (*dto.AppointmentResponse, error) {
	// 将结果转换为JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %w", err)
	}

	// 从JSON转换为AppointmentResponse
	var response dto.AppointmentResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("反序列化结果失败: %w", err)
	}

	return &response, nil
}

// isRetryableError 判断错误是否可重试
func (s *IdempotentService) isRetryableError(err error) bool {
	if !s.config.EnableRetry {
		return false
	}

	// 这里可以根据具体的错误类型判断
	// 例如：网络错误、数据库连接错误等可以重试
	// 业务逻辑错误（如参数验证失败）不应该重试

	errorStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"timeout",
		"network unreachable",
		"database connection",
	}

	for _, retryableError := range retryableErrors {
		if contains(errorStr, retryableError) {
			return true
		}
	}

	return false
}

// generateKey 生成幂等性键
func (s *IdempotentService) generateKey(resource, operation, key string) string {
	timestamp := time.Now().Format("20060102")
	return fmt.Sprintf("appointment:%s:%s:%s:%s", resource, operation, timestamp, key)
}

// contains 检查字符串包含
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)
}

// UpdateAppointmentWithIdempotency 更新带幂等性保证的预约
func (s *IdempotentService) UpdateAppointmentWithIdempotency(
	ctx context.Context,
	appointmentID string,
	req *dto.UpdateAppointmentRequest,
	idempotencyKey string,
) (*dto.AppointmentResponse, error) {

	fullKey := s.generateKey("appointment", "update", idempotencyKey)

	// 检查幂等性
	isFirst, err := s.idempotencyManager.CheckAndRecord(ctx, fullKey, s.config.DefaultTTL)
	if err != nil {
		return nil, fmt.Errorf("幂等性检查失败: %w", err)
	}

	if !isFirst {
		return s.getCachedResult(ctx, fullKey)
	}

	// 执行更新逻辑
	appointment, err := s.baseService.UpdateAppointment(appointmentID, req)
	if err != nil {
		if s.isRetryableError(err) {
			_ = s.idempotencyManager.Delete(ctx, fullKey)
		}
		return nil, err
	}

	// 转换为响应格式并保存结果
	response := &dto.AppointmentResponse{
		ID:     appointment.ID.String(),
		Status: appointment.Status,
	}
	if err := s.saveResult(ctx, fullKey, response); err != nil {
		s.logger.Error("保存更新结果失败", "error", err)
	}

	return response, nil
}

// DeleteAppointmentWithIdempotency 删除带幂等性保证的预约
func (s *IdempotentService) DeleteAppointmentWithIdempotency(
	ctx context.Context,
	appointmentID string,
	idempotencyKey string,
) error {

	fullKey := s.generateKey("appointment", "delete", idempotencyKey)

	// 检查幂等性
	isFirst, err := s.idempotencyManager.CheckAndRecord(ctx, fullKey, s.config.DefaultTTL)
	if err != nil {
		return fmt.Errorf("幂等性检查失败: %w", err)
	}

	if !isFirst {
		// 对于删除操作，如果已存在则认为成功
		s.logger.Info("删除操作幂等性命中", "appointment_id", appointmentID)
		return nil
	}

	// 执行删除逻辑
	err = s.baseService.DeleteAppointment(appointmentID)
	if err != nil {
		if s.isRetryableError(err) {
			_ = s.idempotencyManager.Delete(ctx, fullKey)
		}
		return err
	}

	// 保存成功状态 - 删除操作使用简单的响应
	response := &dto.AppointmentResponse{
		ID:        appointmentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = s.saveResult(ctx, fullKey, response)
	if err != nil {
		s.logger.Error("保存删除结果失败", "error", err)
	}

	return nil
}

// GetIdempotencyStats 获取幂等性统计信息
func (s *IdempotentService) GetIdempotencyStats(ctx context.Context) (*idempotency.IdempotencyStats, error) {
	return s.idempotencyManager.GetStats(ctx)
}

// CleanupExpiredKeys 清理过期的幂等性键
func (s *IdempotentService) CleanupExpiredKeys(ctx context.Context) error {
	// 清理一周前的键
	pattern := "appointment:*:202[0-9]*:" // 匹配旧日期格式的键
	return s.idempotencyManager.DeletePattern(ctx, pattern)
}
