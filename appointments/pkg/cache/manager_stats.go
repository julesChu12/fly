package cache

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// GetStats 获取缓存统计信息
func (c *DefaultCacheManager) GetStats(ctx context.Context) (*CacheStats, error) {
	c.stats.mu.Lock()
	defer c.stats.mu.Unlock()

	// 更新当前缓存数量
	c.updateCacheCount()

	// 创建统计副本
	stats := &CacheStats{
		TotalRequests:     c.stats.TotalRequests,
		CacheHits:         c.stats.CacheHits,
		CacheMisses:       c.stats.CacheMisses,
		HitRate:           c.stats.HitRate,
		AverageAccessTime: c.stats.AverageAccessTime,
		TotalKeys:         c.stats.TotalKeys,
		MemoryKeys:        c.stats.MemoryKeys,
		RedisKeys:         c.stats.RedisKeys,
		LastCleanupTime:   c.stats.LastCleanupTime,
		ErrorCount:        c.stats.ErrorCount,
	}

	return stats, nil
}

// ResetStats 重置统计信息
func (c *DefaultCacheManager) ResetStats(ctx context.Context) error {
	c.stats.mu.Lock()
	defer c.stats.mu.Unlock()

	c.stats = &CacheStats{
		LastCleanupTime: time.Now(),
	}

	c.logger.Info("缓存统计信息已重置")
	return nil
}

// GetDetailedStats 获取详细统计信息
func (c *DefaultCacheManager) GetDetailedStats(ctx context.Context) (*DetailedStats, error) {
	basicStats, err := c.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取基础统计失败: %w", err)
	}

	detailedStats := &DetailedStats{
		CacheStats: *basicStats,
	}

	// 获取内存缓存详情
	if c.memoryEnabled {
		c.memoryCacheMu.RLock()
		detailedStats.MemoryStats = &MemoryStats{
			TotalItems:    int64(len(c.memoryCache)),
			ExpiredItems:  c.countExpiredItems(),
			MemoryUsage:   c.estimateMemoryUsage(),
		}
		c.memoryCacheMu.RUnlock()
	}

	// 获取Redis详情
	if c.redis != nil {
		redisStats, err := c.getRedisStats(ctx)
		if err != nil {
			c.logger.Error("获取Redis统计失败", "error", err)
		} else {
			detailedStats.RedisStats = redisStats
		}
	}

	// 获取性能指标
	detailedStats.PerformanceStats = c.getPerformanceStats()

	return detailedStats, nil
}

// DetailedStats 详细统计信息
type DetailedStats struct {
	CacheStats
	MemoryStats   *MemoryStats     `json:"memory_stats,omitempty"`
	RedisStats    *RedisStats      `json:"redis_stats,omitempty"`
	PerformanceStats *PerformanceStats `json:"performance_stats,omitempty"`
}

// MemoryStats 内存统计
type MemoryStats struct {
	TotalItems   int64 `json:"total_items"`    // 总项数
	ExpiredItems int64 `json:"expired_items"`  // 过期项数
	MemoryUsage  int64 `json:"memory_usage"`   // 内存使用量(字节)
}

// RedisStats Redis统计
type RedisStats struct {
	ConnectedClients int64   `json:"connected_clients"` // 连接的客户端数
	UsedMemory       int64   `json:"used_memory"`        // 已使用内存
	TotalKeys        int64   `json:"total_keys"`         // 总键数
	ExpiredKeys      int64   `json:"expired_keys"`       // 过期键数
	Hits             int64   `json:"hits"`              // 命中数
	Misses           int64   `json:"misses"`            // 未命中数
	HitRate          float64 `json:"hit_rate"`          // 命中率
	Commands         int64   `json:"commands"`          // 执行的命令数
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	AverageResponseTime int64   `json:"average_response_time_ms"` // 平均响应时间
	P95ResponseTime     int64   `json:"p95_response_time_ms"`     // 95%响应时间
	P99ResponseTime     int64   `json:"p99_response_time_ms"`     // 99%响应时间
	ThroughputPerSecond  float64 `json:"throughput_per_second"`   // 每秒吞吐量
	ErrorRate           float64 `json:"error_rate"`             // 错误率
}

// countExpiredItems 统计过期项数
func (c *DefaultCacheManager) countExpiredItems() int64 {
	count := int64(0)
	now := time.Now()

	for _, item := range c.memoryCache {
		if now.After(item.ExpiresAt) {
			count++
		}
	}

	return count
}

// estimateMemoryUsage 估算内存使用量
func (c *DefaultCacheManager) estimateMemoryUsage() int64 {
	totalSize := int64(0)

	for range c.memoryCache {
		// 简单估算：每个缓存项约占用指针开销 + 数据大小
		// 这里使用固定估算值，实际应用中可以使用更精确的方法
		totalSize += 64 // 估算每个缓存项64字节
	}

	return totalSize
}

// getRedisStats 获取Redis统计信息
func (c *DefaultCacheManager) getRedisStats(ctx context.Context) (*RedisStats, error) {
	if c.redis == nil {
		return nil, fmt.Errorf("Redis未启用")
	}

	// 获取Redis INFO命令的输出
	info, err := c.redis.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("获取Redis INFO失败: %w", err)
	}

	// 解析INFO输出
	stats := &RedisStats{}
	c.parseRedisInfo(info, stats)

	// 获取键的统计信息
	keyPattern := c.buildKey("*")
	keys, err := c.redis.Keys(ctx, keyPattern).Result()
	if err == nil {
		stats.TotalKeys = int64(len(keys))
	}

	return stats, nil
}

// parseRedisInfo 解析Redis INFO输出
func (c *DefaultCacheManager) parseRedisInfo(info string, stats *RedisStats) {
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "connected_clients":
			if val, err := parseInt64(value); err == nil {
				stats.ConnectedClients = val
			}
		case "used_memory":
			if val, err := parseInt64(value); err == nil {
				stats.UsedMemory = val
			}
		case "expired_keys":
			if val, err := parseInt64(value); err == nil {
				stats.ExpiredKeys = val
			}
		case "keyspace_hits":
			if val, err := parseInt64(value); err == nil {
				stats.Hits = val
			}
		case "keyspace_misses":
			if val, err := parseInt64(value); err == nil {
				stats.Misses = val
			}
		case "total_commands_processed":
			if val, err := parseInt64(value); err == nil {
				stats.Commands = val
			}
		}
	}

	// 计算命中率
	if stats.Hits+stats.Misses > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
	}
}

// getPerformanceStats 获取性能统计
func (c *DefaultCacheManager) getPerformanceStats() *PerformanceStats {
	stats := &PerformanceStats{}

	// 从基本统计中获取信息
	c.stats.mu.RLock()
	stats.AverageResponseTime = c.stats.AverageAccessTime
	stats.ErrorRate = 0.0
	if c.stats.TotalRequests > 0 {
		stats.ErrorRate = float64(c.stats.ErrorCount) / float64(c.stats.TotalRequests) * 100
	}
	c.stats.mu.RUnlock()

	// 计算吞吐量（基于最近的请求数）
	// 这里简化处理，实际应用中应该基于时间窗口
	stats.ThroughputPerSecond = float64(c.stats.TotalRequests) / 60.0 // 假设运行了1分钟

	// P95和P99响应时间需要更复杂的统计收集，这里使用简化估算
	stats.P95ResponseTime = stats.AverageResponseTime * 2
	stats.P99ResponseTime = stats.AverageResponseTime * 3

	return stats
}

// GetCacheHealth 获取缓存健康状态
func (c *DefaultCacheManager) GetCacheHealth(ctx context.Context) (*CacheHealth, error) {
	health := &CacheHealth{
		Status: "healthy",
		Issues: []string{},
	}

	// 检查Redis连接
	if c.redis != nil {
		if err := c.redis.Ping(ctx).Err(); err != nil {
			health.Status = "degraded"
			health.Issues = append(health.Issues, fmt.Sprintf("Redis连接失败: %v", err))
		}
	}

	// 检查内存缓存
	if c.memoryEnabled {
		c.memoryCacheMu.RLock()
		memoryUsage := len(c.memoryCache)
		c.memoryCacheMu.RUnlock()

		if memoryUsage > c.config.Memory.MaxSize {
			health.Status = "warning"
			health.Issues = append(health.Issues,
				fmt.Sprintf("内存缓存使用过高: %d/%d", memoryUsage, c.config.Memory.MaxSize))
		}
	}

	// 检查命中率
	c.stats.mu.RLock()
	hitRate := c.stats.HitRate
	errorRate := 0.0
	if c.stats.TotalRequests > 0 {
		errorRate = float64(c.stats.ErrorCount) / float64(c.stats.TotalRequests) * 100
	}
	c.stats.mu.RUnlock()

	if hitRate < 50.0 && c.stats.TotalRequests > 100 {
		health.Status = "warning"
		health.Issues = append(health.Issues,
			fmt.Sprintf("缓存命中率过低: %.2f%%", hitRate))
	}

	if errorRate > 5.0 {
		health.Status = "degraded"
		health.Issues = append(health.Issues,
			fmt.Sprintf("错误率过高: %.2f%%", errorRate))
	}

	return health, nil
}

// CacheHealth 缓存健康状态
type CacheHealth struct {
	Status string   `json:"status"` // healthy, warning, degraded, error
	Issues []string `json:"issues"` // 健康问题列表
}

// ExportMetrics 导出指标（用于监控系统）
func (c *DefaultCacheManager) ExportMetrics(ctx context.Context) (map[string]interface{}, error) {
	stats, err := c.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	metrics := map[string]interface{}{
		"cache_total_requests":     stats.TotalRequests,
		"cache_hits":               stats.CacheHits,
		"cache_misses":             stats.CacheMisses,
		"cache_hit_rate":           stats.HitRate,
		"cache_avg_access_time":    stats.AverageAccessTime,
		"cache_total_keys":         stats.TotalKeys,
		"cache_memory_keys":        stats.MemoryKeys,
		"cache_redis_keys":         stats.RedisKeys,
		"cache_error_count":        stats.ErrorCount,
		"cache_last_cleanup_time":  stats.LastCleanupTime,
	}

	return metrics, nil
}

// 辅助函数
func parseInt64(s string) (int64, error) {
	// 简化实现，实际应用中应该使用strconv.ParseInt64
	var result int64 = 0
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int64(ch-'0')
		} else {
			return 0, fmt.Errorf("invalid number")
		}
	}
	return result, nil
}