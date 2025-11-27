package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetAppointment 获取预约缓存
func (c *DefaultCacheManager) GetAppointment(ctx context.Context, id string) (*AppointmentCache, error) {
	key := c.buildKey("appointments:single", id)

	var result *AppointmentCache
	if err := c.getAndUnmarshal(ctx, key, &result); err != nil {
		return nil, fmt.Errorf("获取预约缓存失败: %w", err)
	}

	return result, nil
}

// SetAppointment 设置预约缓存
func (c *DefaultCacheManager) SetAppointment(ctx context.Context, id string, appointment *AppointmentCache, ttl time.Duration) error {
	if appointment == nil {
		return fmt.Errorf("预约数据不能为空")
	}

	// 更新缓存时间戳
	appointment.CachedAt = time.Now()
	if ttl == 0 {
		ttl = c.config.TTL.AppointmentSingle
	}

	key := c.buildKey("appointments:single", id)
	return c.Set(ctx, key, appointment, ttl)
}

// GetAppointmentList 获取预约列表缓存
func (c *DefaultCacheManager) GetAppointmentList(ctx context.Context, filterHash string) (*AppointmentListCache, error) {
	key := c.buildKey("appointments:list", filterHash)

	var result *AppointmentListCache
	if err := c.getAndUnmarshal(ctx, key, &result); err != nil {
		return nil, fmt.Errorf("获取预约列表缓存失败: %w", err)
	}

	// 检查缓存是否过期
	if result != nil && time.Now().After(result.ExpiresAt) {
		c.Delete(ctx, key)
		return nil, fmt.Errorf("缓存已过期")
	}

	return result, nil
}

// SetAppointmentList 设置预约列表缓存
func (c *DefaultCacheManager) SetAppointmentList(ctx context.Context, filterHash string, list *AppointmentListCache, ttl time.Duration) error {
	if list == nil {
		return fmt.Errorf("预约列表数据不能为空")
	}

	// 设置缓存时间戳
	now := time.Now()
	list.CachedAt = now
	list.FilterHash = filterHash
	if ttl == 0 {
		ttl = c.config.TTL.AppointmentList
	}
	list.ExpiresAt = now.Add(ttl)

	key := c.buildKey("appointments:list", filterHash)
	return c.Set(ctx, key, list, ttl)
}

// GetAvailability 获取员工可用性缓存
func (c *DefaultCacheManager) GetAvailability(ctx context.Context, staffID string, date string) (*AvailabilityCache, error) {
	key := c.buildKey("appointments:availability", staffID, date)

	var result *AvailabilityCache
	if err := c.getAndUnmarshal(ctx, key, &result); err != nil {
		return nil, fmt.Errorf("获取可用性缓存失败: %w", err)
	}

	// 检查缓存是否过期
	if result != nil && time.Now().After(result.ExpiresAt) {
		c.Delete(ctx, key)
		return nil, fmt.Errorf("缓存已过期")
	}

	return result, nil
}

// SetAvailability 设置员工可用性缓存
func (c *DefaultCacheManager) SetAvailability(ctx context.Context, staffID string, date string, availability *AvailabilityCache, ttl time.Duration) error {
	if availability == nil {
		return fmt.Errorf("可用性数据不能为空")
	}

	// 设置缓存时间戳
	now := time.Now()
	availability.CachedAt = now
	availability.StaffID = staffID
	availability.Date = date
	if ttl == 0 {
		ttl = c.config.TTL.AvailabilityCheck
	}
	availability.ExpiresAt = now.Add(ttl)

	key := c.buildKey("appointments:availability", staffID, date)
	return c.Set(ctx, key, availability, ttl)
}

// InvalidateAppointment 使预约缓存失效
func (c *DefaultCacheManager) InvalidateAppointment(ctx context.Context, id string) error {
	// 删除单个预约缓存
	singleKey := c.buildKey("appointments:single", id)
	if err := c.Delete(ctx, singleKey); err != nil {
		c.logger.Error("删除预约缓存失败",
			map[string]interface{}{
				"appointment_id": id,
				"error":          err,
			})
	}

	// 删除所有预约列表缓存（因为可能包含该预约）
	listPattern := c.buildKey("appointments:list", "*")
	if err := c.DeletePattern(ctx, listPattern); err != nil {
		c.logger.Error("删除预约列表缓存失败",
			map[string]interface{}{
				"appointment_id": id,
				"pattern":        listPattern,
				"error":          err,
			})
	}

	return nil
}

// InvalidateAppointmentList 使所有预约列表缓存失效
func (c *DefaultCacheManager) InvalidateAppointmentList(ctx context.Context) error {
	pattern := c.buildKey("appointments:list", "*")
	return c.DeletePattern(ctx, pattern)
}

// InvalidateStaffAvailability 使员工可用性缓存失效
func (c *DefaultCacheManager) InvalidateStaffAvailability(ctx context.Context, staffID string, date string) error {
	if date == "" {
		// 删除指定员工所有日期的可用性缓存
		pattern := c.buildKey("appointments:availability", staffID, "*")
		return c.DeletePattern(ctx, pattern)
	}

	// 删除指定员工指定日期的可用性缓存
	key := c.buildKey("appointments:availability", staffID, date)
	return c.Delete(ctx, key)
}

// InvalidateStaffRangeAvailability 使员工日期范围内的可用性缓存失效
func (c *DefaultCacheManager) InvalidateStaffRangeAvailability(ctx context.Context, staffID string, startDate, endDate time.Time) error {
	// 对于日期范围，我们需要删除该员工的所有可用性缓存
	// 因为可能的时间段太多，直接删除所有
	return c.InvalidateStaffAvailability(ctx, staffID, "")
}

// getAndUnmarshal 通用的获取并反序列化方法
func (c *DefaultCacheManager) getAndUnmarshal(ctx context.Context, key string, target interface{}) error {
	// 直接从Redis获取并反序列化
	if c.redis != nil {
		value, err := c.redis.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				return fmt.Errorf("缓存未命中")
			}
			return fmt.Errorf("Redis获取失败: %w", err)
		}

		// 反序列化
		if err := json.Unmarshal([]byte(value), target); err != nil {
			return fmt.Errorf("反序列化失败: %w", err)
		}

		// 同时更新内存缓存
		if c.memoryEnabled {
			c.setToMemory(key, target, c.config.Memory.DefaultTTL)
		}

		return nil
	}

	return fmt.Errorf("Redis未启用")
}

// 内存缓存相关方法

// getFromMemory 从内存缓存获取
func (c *DefaultCacheManager) getFromMemory(key string) interface{} {
	c.memoryCacheMu.RLock()
	defer c.memoryCacheMu.RUnlock()

	item, exists := c.memoryCache[key]
	if !exists || time.Now().After(item.ExpiresAt) {
		return nil
	}

	// 更新访问计数
	c.memoryCacheMu.RUnlock()
	c.memoryCacheMu.Lock()
	item.AccessCount++
	c.memoryCacheMu.RLock()

	return item.Value
}

// setToMemory 设置到内存缓存
func (c *DefaultCacheManager) setToMemory(key string, value interface{}, ttl time.Duration) {
	c.memoryCacheMu.Lock()
	defer c.memoryCacheMu.Unlock()

	// 检查缓存大小限制
	if len(c.memoryCache) >= c.config.Memory.MaxSize {
		c.evictLRU()
	}

	c.memoryCache[key] = &memoryCacheItem{
		Value:       value,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 1,
	}
}

// deleteFromMemory 从内存缓存删除
func (c *DefaultCacheManager) deleteFromMemory(key string) {
	c.memoryCacheMu.Lock()
	defer c.memoryCacheMu.Unlock()

	delete(c.memoryCache, key)
}

// deleteFromMemoryByPattern 从内存缓存删除匹配模式的键
func (c *DefaultCacheManager) deleteFromMemoryByPattern(pattern string) {
	c.memoryCacheMu.Lock()
	defer c.memoryCacheMu.Unlock()

	for key := range c.memoryCache {
		if strings.Contains(key, strings.TrimSuffix(pattern, "*")) {
			delete(c.memoryCache, key)
		}
	}
}

// existsInMemory 检查内存缓存中是否存在
func (c *DefaultCacheManager) existsInMemory(key string) bool {
	c.memoryCacheMu.RLock()
	defer c.memoryCacheMu.RUnlock()

	item, exists := c.memoryCache[key]
	return exists && !time.Now().After(item.ExpiresAt)
}

// evictLRU 使用LRU策略驱逐内存缓存
func (c *DefaultCacheManager) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	var lowestAccessCount int64 = -1

	for key, item := range c.memoryCache {
		// 如果已过期，直接删除
		if time.Now().After(item.ExpiresAt) {
			delete(c.memoryCache, key)
			continue
		}

		// 找到访问次数最少且最旧的项
		if lowestAccessCount == -1 || item.AccessCount < lowestAccessCount ||
		   (item.AccessCount == lowestAccessCount && item.CreatedAt.Before(oldestTime)) {
			oldestKey = key
			oldestTime = item.CreatedAt
			lowestAccessCount = item.AccessCount
		}
	}

	if oldestKey != "" {
		delete(c.memoryCache, oldestKey)
	}
}

// 统计相关方法

// updateStats 更新统计信息
func (c *DefaultCacheManager) updateStats(accessTime time.Duration, isHit bool, isError bool) {
	c.stats.mu.Lock()
	defer c.stats.mu.Unlock()

	c.stats.TotalRequests++
	if isHit {
		c.stats.CacheHits++
	} else {
		c.stats.CacheMisses++
	}

	if isError {
		c.stats.ErrorCount++
	}

	// 更新平均访问时间
	if c.stats.TotalRequests == 1 {
		c.stats.AverageAccessTime = accessTime.Milliseconds()
	} else {
		c.stats.AverageAccessTime = (c.stats.AverageAccessTime + accessTime.Milliseconds()) / 2
	}

	// 计算命中率
	c.stats.HitRate = float64(c.stats.CacheHits) / float64(c.stats.TotalRequests) * 100
}

// startCleanupRoutine 启动清理协程
func (c *DefaultCacheManager) startCleanupRoutine() {
	ticker := time.NewTicker(c.config.Cleanup.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		}
	}
}

// startStatsRoutine 启动统计协程
func (c *DefaultCacheManager) startStatsRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.updateCacheCount()
		}
	}
}

// cleanup 清理过期缓存
func (c *DefaultCacheManager) cleanup() {
	// 清理内存缓存
	if c.memoryEnabled {
		c.memoryCacheMu.Lock()
		for key, item := range c.memoryCache {
			if time.Now().After(item.ExpiresAt) {
				delete(c.memoryCache, key)
			}
		}
		c.memoryCacheMu.Unlock()
	}

	// 记录清理时间
	c.stats.mu.Lock()
	c.stats.LastCleanupTime = time.Now()
	c.stats.mu.Unlock()

	c.logger.Debug("缓存清理完成")
}

// updateCacheCount 更新缓存数量统计
func (c *DefaultCacheManager) updateCacheCount() {
	c.stats.mu.Lock()
	defer c.stats.mu.Unlock()

	// 内存缓存数量
	if c.memoryEnabled {
		c.stats.MemoryKeys = int64(len(c.memoryCache))
	}

	// Redis缓存数量（这里简化处理，实际可以通过SCAN命令获取）
	c.stats.TotalKeys = c.stats.MemoryKeys + c.stats.RedisKeys
}

// HealthCheck 健康检查
func (c *DefaultCacheManager) HealthCheck(ctx context.Context) error {
	// 检查Redis连接
	if c.redis != nil {
		if err := c.redis.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("Redis连接检查失败: %w", err)
		}
	}

	// 检查内存缓存状态
	if c.memoryEnabled && c.config.Memory.MaxSize > 0 {
		c.memoryCacheMu.RLock()
		memoryUsage := len(c.memoryCache)
		c.memoryCacheMu.RUnlock()

		if memoryUsage > c.config.Memory.MaxSize*2 { // 超过限制的2倍
			return fmt.Errorf("内存缓存使用异常: 当前%d，最大%d", memoryUsage, c.config.Memory.MaxSize)
		}
	}

	return nil
}

// 辅助方法

// buildKey 构建缓存键
func (c *DefaultCacheManager) buildKey(parts ...string) string {
	if c.config.Redis.KeyPrefix != "" {
		parts = append([]string{c.config.Redis.KeyPrefix}, parts...)
	}
	return strings.Join(parts, ":")
}

// jsonMarshal JSON序列化
func (c *DefaultCacheManager) jsonMarshal(value interface{}) (string, error) {
	data, err := c.jsonMarshalToBytes(value)
	return string(data), err
}

// jsonMarshalToBytes JSON序列化到字节
func (c *DefaultCacheManager) jsonMarshalToBytes(value interface{}) ([]byte, error) {
	if marshaler, ok := c.idempotencyMgr.(interface {
		MarshalJSON(v interface{}) ([]byte, error)
	}); ok {
		return marshaler.MarshalJSON(value)
	}
	// 使用标准json包
	return nil, fmt.Errorf("JSON序列化器不可用")
}

// jsonUnmarshal JSON反序列化
func (c *DefaultCacheManager) jsonUnmarshal(data []byte, target interface{}) error {
	if unmarshaler, ok := c.idempotencyMgr.(interface {
		UnmarshalJSON(data []byte, v interface{}) error
	}); ok {
		return unmarshaler.UnmarshalJSON(data, target)
	}
	// 使用标准json包
	return fmt.Errorf("JSON反序列化器不可用")
}