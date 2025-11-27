package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/julesChu12/fly/appointments/pkg/idempotency"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// CacheManager 缓存管理器接口
type CacheManager interface {
	// 基本操作
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	MGet(ctx context.Context, keys []string) ([]interface{}, error)
	MSet(ctx context.Context, pairs map[string]interface{}, ttl time.Duration) error

	// 业务特定方法
	GetAppointment(ctx context.Context, id string) (*AppointmentCache, error)
	SetAppointment(ctx context.Context, id string, appointment *AppointmentCache, ttl time.Duration) error
	GetAppointmentList(ctx context.Context, filterHash string) (*AppointmentListCache, error)
	SetAppointmentList(ctx context.Context, filterHash string, list *AppointmentListCache, ttl time.Duration) error
	GetAvailability(ctx context.Context, staffID string, date string) (*AvailabilityCache, error)
	SetAvailability(ctx context.Context, staffID string, date string, availability *AvailabilityCache, ttl time.Duration) error

	// 缓存失效
	InvalidateAppointment(ctx context.Context, id string) error
	InvalidateAppointmentList(ctx context.Context) error
	InvalidateStaffAvailability(ctx context.Context, staffID string, date string) error

	// 统计和监控
	GetStats(ctx context.Context) (*CacheStats, error)
	ResetStats(ctx context.Context) error

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// DefaultCacheManager 默认缓存管理器实现
type DefaultCacheManager struct {
	redis          *redis.Client
	idempotencyMgr idempotency.IdempotencyManager
	config         *CacheConfig
	logger         *logger.Logger
	stats          *CacheStats
	mu             sync.RWMutex

	// 内存缓存（可选）
	memoryCache    map[string]*memoryCacheItem
	memoryCacheMu  sync.RWMutex
	memoryEnabled  bool
}

// memoryCacheItem 内存缓存项
type memoryCacheItem struct {
	Value      interface{}
	ExpiresAt  time.Time
	CreatedAt  time.Time
	AccessCount int64
}

// AppointmentCache 预约缓存
type AppointmentCache struct {
	ID            string                 `json:"id"`
	CustomerID    string                 `json:"customer_id"`
	StaffID       string                 `json:"staff_id"`
	ServiceID     string                 `json:"service_id"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Status        string                 `json:"status"`
	Notes         *string                `json:"notes"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CachedAt      time.Time              `json:"cached_at"`
	Version       int64                  `json:"version"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// AppointmentListCache 预约列表缓存
type AppointmentListCache struct {
	Appointments []*AppointmentCache   `json:"appointments"`
	Total        int64                  `json:"total"`
	Page         int                    `json:"page"`
	PageSize     int                    `json:"page_size"`
	FilterHash   string                 `json:"filter_hash"`
	CachedAt     time.Time              `json:"cached_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
}

// AvailabilityCache 可用性缓存
type AvailabilityCache struct {
	StaffID      string                 `json:"staff_id"`
	Date         string                 `json:"date"`
	IsAvailable  bool                   `json:"is_available"`
	BusySlots    []TimeSlot             `json:"busy_slots"`
	TotalSlots   int                    `json:"total_slots"`
	CachedAt     time.Time              `json:"cached_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
}

// TimeSlot 时间段
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalRequests    int64     `json:"total_requests"`
	CacheHits        int64     `json:"cache_hits"`
	CacheMisses      int64     `json:"cache_misses"`
	HitRate          float64   `json:"hit_rate"`
	AverageAccessTime int64    `json:"avg_access_time_ms"`
	TotalKeys        int64     `json:"total_keys"`
	MemoryKeys       int64     `json:"memory_keys"`
	RedisKeys        int64     `json:"redis_keys"`
	LastCleanupTime  time.Time `json:"last_cleanup_time"`
	ErrorCount       int64     `json:"error_count"`
	mu               sync.RWMutex
}

// NewDefaultCacheManager 创建默认缓存管理器
func NewDefaultCacheManager(
	config *CacheConfig,
	idempotencyMgr idempotency.IdempotencyManager,
	logger *logger.Logger,
) *DefaultCacheManager {
	if config == nil {
		config = DefaultCacheConfig()
	}

	// 基于幂等性管理器的Redis客户端
	var redisClient *redis.Client
	if idempotencyMgr != nil {
		// 这里需要从幂等性管理器获取Redis客户端
		// 由于接口限制，我们创建新的客户端
		redisClient = redis.NewClient(&redis.Options{
			Addr:     config.Redis.Address,
			Password: config.Redis.Password,
			DB:       config.Redis.DB,
			PoolSize: config.Redis.PoolSize,
		})
	}

	manager := &DefaultCacheManager{
		redis:          redisClient,
		idempotencyMgr: idempotencyMgr,
		config:         config,
		logger:         logger,
		stats:          &CacheStats{LastCleanupTime: time.Now()},
		memoryCache:    make(map[string]*memoryCacheItem),
		memoryEnabled:  config.Memory.Enabled,
	}

	// 启动后台清理协程
	go manager.startCleanupRoutine()

	// 启动统计协程
	go manager.startStatsRoutine()

	logger.Info("缓存管理器初始化完成",
		map[string]interface{}{
			"redis_enabled": redisClient != nil,
			"memory_enabled": config.Memory.Enabled,
			"memory_size": config.Memory.MaxSize,
		})

	return manager
}

// Get 获取缓存值
func (c *DefaultCacheManager) Get(ctx context.Context, key string) (interface{}, error) {
	start := time.Now()
	defer func() {
		c.updateStats(time.Since(start), false, false)
	}()

	// 1. 先检查内存缓存
	if c.memoryEnabled {
		if value := c.getFromMemory(key); value != nil {
			c.updateStats(time.Since(start), true, false)
			return value, nil
		}
	}

	// 2. 检查Redis缓存
	if c.redis != nil {
		value, err := c.redis.Get(ctx, key).Result()
		if err == nil {
			// 反序列化
			var result interface{}
			if err := json.Unmarshal([]byte(value), &result); err != nil {
				c.logger.Error("反序列化缓存失败",
					map[string]interface{}{
						"key":   key,
						"error": err,
					})
				return nil, fmt.Errorf("反序列化缓存失败: %w", err)
			}

			// 写入内存缓存
			if c.memoryEnabled {
				c.setToMemory(key, result, c.config.Memory.DefaultTTL)
			}

			c.updateStats(time.Since(start), true, false)
			return result, nil
		}

		if err != redis.Nil {
			c.logger.Error("Redis缓存获取失败",
				map[string]interface{}{
					"key":   key,
					"error": err,
				})
			c.updateStats(time.Since(start), false, true)
			return nil, fmt.Errorf("Redis缓存获取失败: %w", err)
		}
	}

	c.updateStats(time.Since(start), false, false)
	return nil, fmt.Errorf("缓存未命中")
}

// Set 设置缓存值
func (c *DefaultCacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 1. 设置内存缓存
	if c.memoryEnabled {
		c.setToMemory(key, value, ttl)
	}

	// 2. 设置Redis缓存
	if c.redis != nil {
		// 序列化
		data, err := json.Marshal(value)
		if err != nil {
			c.logger.Error("序列化缓存失败",
				map[string]interface{}{
					"key":   key,
					"error": err,
				})
			return fmt.Errorf("序列化缓存失败: %w", err)
		}

		// 设置到Redis
		if err := c.redis.Set(ctx, key, data, ttl).Err(); err != nil {
			c.logger.Error("Redis缓存设置失败",
				map[string]interface{}{
					"key":   key,
					"ttl":   ttl,
					"error": err,
				})
			return fmt.Errorf("Redis缓存设置失败: %w", err)
		}
	}

	return nil
}

// Delete 删除缓存
func (c *DefaultCacheManager) Delete(ctx context.Context, key string) error {
	// 1. 删除内存缓存
	if c.memoryEnabled {
		c.deleteFromMemory(key)
	}

	// 2. 删除Redis缓存
	if c.redis != nil {
		if err := c.redis.Del(ctx, key).Err(); err != nil {
			c.logger.Error("Redis缓存删除失败",
				map[string]interface{}{
					"key":   key,
					"error": err,
				})
			return fmt.Errorf("Redis缓存删除失败: %w", err)
		}
	}

	return nil
}

// DeletePattern 删除匹配模式的缓存
func (c *DefaultCacheManager) DeletePattern(ctx context.Context, pattern string) error {
	// 1. 删除匹配的内存缓存
	if c.memoryEnabled {
		c.deleteFromMemoryByPattern(pattern)
	}

	// 2. 删除匹配的Redis缓存
	if c.redis != nil {
		keys, err := c.redis.Keys(ctx, pattern).Result()
		if err != nil {
			return fmt.Errorf("获取匹配键失败: %w", err)
		}

		if len(keys) > 0 {
			if err := c.redis.Del(ctx, keys...).Err(); err != nil {
				c.logger.Error("批量删除Redis缓存失败",
					map[string]interface{}{
						"pattern": pattern,
						"keys":    len(keys),
						"error":   err,
					})
				return fmt.Errorf("批量删除Redis缓存失败: %w", err)
			}
		}
	}

	return nil
}

// Exists 检查缓存是否存在
func (c *DefaultCacheManager) Exists(ctx context.Context, key string) (bool, error) {
	// 1. 检查内存缓存
	if c.memoryEnabled {
		if c.existsInMemory(key) {
			return true, nil
		}
	}

	// 2. 检查Redis缓存
	if c.redis != nil {
		exists, err := c.redis.Exists(ctx, key).Result()
		if err != nil {
			return false, fmt.Errorf("检查Redis缓存存在性失败: %w", err)
		}
		return exists > 0, nil
	}

	return false, nil
}

// MGet 批量获取
func (c *DefaultCacheManager) MGet(ctx context.Context, keys []string) ([]interface{}, error) {
	if c.redis == nil {
		return nil, fmt.Errorf("Redis未启用")
	}

	results := make([]interface{}, len(keys))
	for i, key := range keys {
		value, err := c.redis.Get(ctx, key).Result()
		if err == nil {
			var result interface{}
			if err := json.Unmarshal([]byte(value), &result); err == nil {
				results[i] = result
			}
		}
		// 如果获取失败，保持nil
	}

	return results, nil
}

// MSet 批量设置
func (c *DefaultCacheManager) MSet(ctx context.Context, pairs map[string]interface{}, ttl time.Duration) error {
	if c.redis == nil {
		return fmt.Errorf("Redis未启用")
	}

	// 序列化所有值
	serializedPairs := make(map[string]string)
	for key, value := range pairs {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("序列化键 %s 失败: %w", key, err)
		}
		serializedPairs[key] = string(data)
	}

	// 使用管道批量设置
	pipe := c.redis.Pipeline()
	for key, value := range serializedPairs {
		pipe.Set(ctx, key, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("批量设置缓存失败: %w", err)
	}

	return nil
}