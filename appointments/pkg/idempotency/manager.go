package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// IdempotencyManager 幂等性管理器接口
type IdempotencyManager interface {
	// CheckAndRecord 检查并记录幂等性键
	CheckAndRecord(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// GetResult 获取幂等性键的结果
	GetResult(ctx context.Context, key string) (interface{}, error)

	// SaveResult 保存幂等性键的结果
	SaveResult(ctx context.Context, key string, result interface{}, ttl time.Duration) error

	// Delete 删除幂等性键
	Delete(ctx context.Context, key string) error

	// DeletePattern 删除匹配模式的幂等性键
	DeletePattern(ctx context.Context, pattern string) error

	// GetStats 获取幂等性统计信息
	GetStats(ctx context.Context) (*IdempotencyStats, error)
}

// IdempotencyStats 幂等性统计信息
type IdempotencyStats struct {
	TotalKeys  int64   `json:"total_keys"`
	ActiveKeys int64   `json:"active_keys"`
	HitCount   int64   `json:"hit_count"`
	MissCount  int64   `json:"miss_count"`
	HitRate    float64 `json:"hit_rate"`
	AverageTTL int64   `json:"average_ttl"`
}

// RedisIdempotencyManager Redis实现的幂等性管理器
type RedisIdempotencyManager struct {
	redis  *redis.Client
	prefix string
	logger *logger.Logger
	config *IdempotencyConfig
}

// IdempotencyConfig 幂等性配置
type IdempotencyConfig struct {
	Prefix          string        `yaml:"prefix"`           // 键前缀
	DefaultTTL      time.Duration `yaml:"default_ttl"`      // 默认TTL
	MaxTTL          time.Duration `yaml:"max_ttl"`          // 最大TTL
	CleanupInterval time.Duration `yaml:"cleanup_interval"` // 清理间隔
	EnableStats     bool          `yaml:"enable_stats"`     // 是否启用统计
}

// DefaultIdempotencyConfig 默认幂等性配置
func DefaultIdempotencyConfig() *IdempotencyConfig {
	return &IdempotencyConfig{
		Prefix:          "idempotency:",
		DefaultTTL:      24 * time.Hour,
		MaxTTL:          7 * 24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
		EnableStats:     true,
	}
}

// NewRedisIdempotencyManager 创建Redis幂等性管理器
func NewRedisIdempotencyManager(redis *redis.Client, config *IdempotencyConfig, logger *logger.Logger) *RedisIdempotencyManager {
	if config == nil {
		config = DefaultIdempotencyConfig()
	}

	manager := &RedisIdempotencyManager{
		redis:  redis,
		prefix: config.Prefix,
		logger: logger,
		config: config,
	}

	// 启动清理协程
	go manager.startCleanupRoutine()

	return manager
}

// CheckAndRecord 检查并记录幂等性键
func (m *RedisIdempotencyManager) CheckAndRecord(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	fullKey := m.buildKey(key)

	// 验证TTL
	if ttl <= 0 {
		ttl = m.config.DefaultTTL
	}
	if ttl > m.config.MaxTTL {
		ttl = m.config.MaxTTL
	}

	// 使用SETNX实现原子性检查和设置
	result, err := m.redis.SetNX(ctx, fullKey, "processing", ttl).Result()
	if err != nil {
		m.logger.Error("幂等性键检查失败",
			map[string]interface{}{
				"key":   key,
				"error": err,
			})
		return false, fmt.Errorf("幂等性键检查失败: %w", err)
	}

	// 更新统计信息
	if m.config.EnableStats {
		if result {
			m.incrementMissCount(ctx)
		} else {
			m.incrementHitCount(ctx)
		}
	}

	m.logger.Debug("幂等性键检查完成",
		map[string]interface{}{
			"key":     key,
			"success": result,
			"ttl":     ttl,
		})

	return result, nil
}

// GetResult 获取幂等性键的结果
func (m *RedisIdempotencyManager) GetResult(ctx context.Context, key string) (interface{}, error) {
	fullKey := m.buildKey(key)

	// 获取结果数据
	result, err := m.redis.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 键不存在
		}
		m.logger.Error("获取幂等性结果失败",
			map[string]interface{}{
				"key":   key,
				"error": err,
			})
		return nil, fmt.Errorf("获取幂等性结果失败: %w", err)
	}

	// 如果是处理中的状态，返回nil
	if result == "processing" {
		return nil, nil
	}

	// 尝试解析JSON结果
	var parsedResult interface{}
	if err := json.Unmarshal([]byte(result), &parsedResult); err != nil {
		// 如果不是JSON，直接返回字符串
		return result, nil
	}

	return parsedResult, nil
}

// SaveResult 保存幂等性键的结果
func (m *RedisIdempotencyManager) SaveResult(ctx context.Context, key string, result interface{}, ttl time.Duration) error {
	fullKey := m.buildKey(key)

	// 验证TTL
	if ttl <= 0 {
		ttl = m.config.DefaultTTL
	}
	if ttl > m.config.MaxTTL {
		ttl = m.config.MaxTTL
	}

	// 序列化结果
	resultJSON, err := json.Marshal(result)
	if err != nil {
		m.logger.Error("序列化幂等性结果失败",
			map[string]interface{}{
				"key":   key,
				"error": err,
			})
		return fmt.Errorf("序列化幂等性结果失败: %w", err)
	}

	// 保存结果
	err = m.redis.Set(ctx, fullKey, string(resultJSON), ttl).Err()
	if err != nil {
		m.logger.Error("保存幂等性结果失败",
			map[string]interface{}{
				"key":   key,
				"error": err,
			})
		return fmt.Errorf("保存幂等性结果失败: %w", err)
	}

	m.logger.Debug("幂等性结果保存成功",
		map[string]interface{}{
			"key": key,
			"ttl": ttl,
		})

	return nil
}

// Delete 删除幂等性键
func (m *RedisIdempotencyManager) Delete(ctx context.Context, key string) error {
	fullKey := m.buildKey(key)

	err := m.redis.Del(ctx, fullKey).Err()
	if err != nil {
		m.logger.Error("删除幂等性键失败",
			map[string]interface{}{
				"key":   key,
				"error": err,
			})
		return fmt.Errorf("删除幂等性键失败: %w", err)
	}

	m.logger.Debug("幂等性键删除成功", "key", key)
	return nil
}

// DeletePattern 删除匹配模式的幂等性键
func (m *RedisIdempotencyManager) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := m.prefix + pattern

	keys, err := m.redis.Keys(ctx, fullPattern).Result()
	if err != nil {
		m.logger.Error("获取幂等性键模式失败",
			map[string]interface{}{
				"pattern": pattern,
				"error":   err,
			})
		return fmt.Errorf("获取幂等性键模式失败: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// 批量删除
	err = m.redis.Del(ctx, keys...).Err()
	if err != nil {
		m.logger.Error("批量删除幂等性键失败",
			map[string]interface{}{
				"pattern":   pattern,
				"key_count": len(keys),
				"error":     err,
			})
		return fmt.Errorf("批量删除幂等性键失败: %w", err)
	}

	m.logger.Debug("批量删除幂等性键成功",
		map[string]interface{}{
			"pattern":   pattern,
			"key_count": len(keys),
		})

	return nil
}

// GetStats 获取幂等性统计信息
func (m *RedisIdempotencyManager) GetStats(ctx context.Context) (*IdempotencyStats, error) {
	if !m.config.EnableStats {
		return &IdempotencyStats{}, nil
	}

	// 获取总键数
	totalKeys, err := m.redis.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("获取总键数失败: %w", err)
	}

	// 获取活跃的幂等性键
	activeKeys, err := m.redis.Keys(ctx, m.prefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("获取活跃键数失败: %w", err)
	}

	// 获取命中和未命中次数
	hitCount, _ := m.redis.Get(ctx, m.prefix+"stats:hit_count").Int64()
	missCount, _ := m.redis.Get(ctx, m.prefix+"stats:miss_count").Int64()

	// 计算命中率
	total := hitCount + missCount
	var hitRate float64
	if total > 0 {
		hitRate = float64(hitCount) / float64(total) * 100
	}

	stats := &IdempotencyStats{
		TotalKeys:  totalKeys,
		ActiveKeys: int64(len(activeKeys)),
		HitCount:   hitCount,
		MissCount:  missCount,
		HitRate:    hitRate,
		AverageTTL: m.config.DefaultTTL.Milliseconds(),
	}

	return stats, nil
}

// buildKey 构建完整的键名
func (m *RedisIdempotencyManager) buildKey(key string) string {
	return m.prefix + key
}

// incrementHitCount 增加命中计数
func (m *RedisIdempotencyManager) incrementHitCount(ctx context.Context) {
	if !m.config.EnableStats {
		return
	}
	m.redis.Incr(ctx, m.prefix+"stats:hit_count")
}

// incrementMissCount 增加未命中计数
func (m *RedisIdempotencyManager) incrementMissCount(ctx context.Context) {
	if !m.config.EnableStats {
		return
	}
	m.redis.Incr(ctx, m.prefix+"stats:miss_count")
}

// startCleanupRoutine 启动清理协程
func (m *RedisIdempotencyManager) startCleanupRoutine() {
	if m.config.CleanupInterval <= 0 {
		return
	}

	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		}
	}
}

// cleanup 清理过期的幂等性键
func (m *RedisIdempotencyManager) cleanup() {
	ctx := context.Background()
	keys, err := m.redis.Keys(ctx, m.prefix+"*").Result()
	if err != nil {
		m.logger.Error("获取清理键列表失败", "error", err)
		return
	}

	expiredKeys := make([]string, 0)

	for _, key := range keys {
		// 获取键的TTL
		ttl, err := m.redis.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		// 如果键已过期，加入删除列表
		if ttl == -1 { // 没有设置过期时间
			expiredKeys = append(expiredKeys, key)
		}
	}

	// 批量删除过期键
	if len(expiredKeys) > 0 {
		err = m.redis.Del(ctx, expiredKeys...).Err()
		if err != nil {
			m.logger.Error("批量删除过期键失败",
				map[string]interface{}{
					"expired_count": len(expiredKeys),
					"error":         err,
				})
		} else {
			m.logger.Debug("清理过期键成功",
				map[string]interface{}{
					"expired_count": len(expiredKeys),
				})
		}
	}
}

// GenerateIdempotencyKey 生成幂等性键
func GenerateIdempotencyKey(userID, resource, operation string, params map[string]string) string {
	timestamp := time.Now().Format("20060102")
	return fmt.Sprintf("%s:%s:%s:%s", userID, resource, operation, timestamp)
}
