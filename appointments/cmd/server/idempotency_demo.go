//go:build ignore

package main

import (
	"context"

	"fmt"

	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/appointments/internal/application/service"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/appointments/pkg/idempotency"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// setupIdempotency 设置幂等性功能
func setupIdempotency(
	appointmentService *service.AppointmentService,
	appLogger *logger.Logger,
) (*service.IdempotentService, idempotency.IdempotencyManager, error) {

	appLogger.Info("开始初始化幂等性功能")

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	// 测试Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		appLogger.Error("Redis连接失败",
			map[string]interface{}{
				"error": err,
			})
		return nil, nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	appLogger.Info("Redis连接成功")

	// 创建幂等性管理器
	idempotencyManager := idempotency.NewRedisIdempotencyManager(
		redisClient,
		idempotency.DefaultIdempotencyConfig(),
		appLogger,
	)

	// 创建幂等性服务
	idempotencyConfig := service.DefaultIdempotencyServiceConfig()
	idempotentService := service.NewIdempotentService(
		*appointmentService,
		idempotencyManager,
		idempotencyConfig,
		appLogger,
	)

	appLogger.Info("幂等性功能初始化完成")
	return idempotentService, idempotencyManager, nil
}

// demoIdempotency 演示幂等性功能
func demoIdempotency(
	ctx context.Context,
	idempotentService *service.IdempotentService,
	idempotencyManager idempotency.IdempotencyManager,
	appLogger *logger.Logger,
) {

	appLogger.Info("开始演示幂等性功能")

	// 1. 演示第一次请求
	appLogger.Info("=== 演示第一次请求 ===")
	idempotencyKey := generateIdempotencyKey()

	req1 := &dto.CreateAppointmentRequest{
		CustomerID: uuid.New().String(),
		StaffID:    uuid.New().String(),
		ServiceID:  uuid.New().String(),
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
		Notes:      stringPtr("幂等性测试预约 - 第一次请求"),
	}

	response1, err := idempotentService.CreateAppointmentWithIdempotency(ctx, req1, idempotencyKey)
	if err != nil {
		appLogger.Error("第一次请求失败", "error", err)
		return
	}

	appLogger.Info("第一次请求成功",
		map[string]interface{}{
			"appointment_id":  response1.ID,
			"idempotency_key": idempotencyKey,
		})

	// 2. 演示重复请求（应该返回缓存结果）
	appLogger.Info("=== 演示重复请求 ===")
	req2 := &dto.CreateAppointmentRequest{
		CustomerID: req1.CustomerID,
		StaffID:    req1.StaffID,
		ServiceID:  req1.ServiceID,
		StartTime:  req1.StartTime,
		EndTime:    req1.EndTime,
		Notes:      stringPtr("幂等性测试预约 - 重复请求"),
	}

	response2, err := idempotentService.CreateAppointmentWithIdempotency(ctx, req2, idempotencyKey)
	if err != nil {
		appLogger.Error("重复请求失败", "error", err)
		return
	}

	appLogger.Info("重复请求成功（返回缓存结果）",
		map[string]interface{}{
			"appointment_id":  response2.ID,
			"idempotency_key": idempotencyKey,
		})

	// 验证两次请求的结果是否相同
	if response1.ID == response2.ID {
		appLogger.Info("幂等性验证成功：两次请求返回相同结果")
	} else {
		appLogger.Error("幂等性验证失败：两次请求返回不同结果",
			map[string]interface{}{
				"first_response_id":  response1.ID,
				"second_response_id": response2.ID,
			})
	}

	// 3. 演示更新操作的幂等性
	appLogger.Info("=== 演示更新操作幂等性 ===")
	updateKey := generateIdempotencyKey()

	updateReq := &dto.UpdateAppointmentRequest{
		Notes: stringPtr("幂等性更新测试"),
	}

	updateResponse1, err := idempotentService.UpdateAppointmentWithIdempotency(
		ctx,
		response1.ID,
		updateReq,
		updateKey,
	)
	if err != nil {
		appLogger.Error("第一次更新失败", "error", err)
		return
	}

	appLogger.Info("第一次更新成功",
		map[string]interface{}{
			"appointment_id": updateResponse1.ID,
			"update_key":     updateKey,
			"notes":          updateResponse1.Notes,
		})

	// 重复更新
	updateResponse2, err := idempotentService.UpdateAppointmentWithIdempotency(
		ctx,
		response1.ID,
		updateReq,
		updateKey,
	)
	if err != nil {
		appLogger.Error("重复更新失败", "error", err)
		return
	}

	appLogger.Info("重复更新成功（返回缓存结果）",
		map[string]interface{}{
			"appointment_id": updateResponse2.ID,
			"update_key":     updateKey,
			"notes":          updateResponse2.Notes,
		})

	// 4. 演示删除操作的幂等性
	appLogger.Info("=== 演示删除操作幂等性 ===")
	deleteKey := generateIdempotencyKey()

	err = idempotentService.DeleteAppointmentWithIdempotency(ctx, updateResponse1.ID, deleteKey)
	if err != nil {
		appLogger.Error("第一次删除失败", "error", err)
		return
	}

	appLogger.Info("第一次删除成功", "appointment_id", updateResponse1.ID)

	// 重复删除
	err = idempotentService.DeleteAppointmentWithIdempotency(ctx, updateResponse1.ID, deleteKey)
	if err != nil {
		appLogger.Error("重复删除失败", "error", err)
		return
	}

	appLogger.Info("重复删除成功（幂等性命中）", "appointment_id", updateResponse1.ID)

	// 5. 获取统计信息
	appLogger.Info("=== 获取幂等性统计信息 ===")
	stats, err := idempotencyManager.GetStats(ctx)
	if err != nil {
		appLogger.Error("获取统计信息失败", "error", err)
		return
	}

	appLogger.Info("幂等性统计信息",
		map[string]interface{}{
			"total_keys":  stats.TotalKeys,
			"active_keys": stats.ActiveKeys,
			"hit_count":   stats.HitCount,
			"miss_count":  stats.MissCount,
			"hit_rate":    fmt.Sprintf("%.2f%%", stats.HitRate),
			"average_ttl": stats.AverageTTL,
		})

	// 6. 演示过期键清理
	appLogger.Info("=== 演示过期键清理 ===")
	err = idempotentService.CleanupExpiredKeys(ctx)
	if err != nil {
		appLogger.Error("清理过期键失败", "error", err)
	} else {
		appLogger.Info("过期键清理成功")
	}
}

// generateIdempotencyKey 生成幂等性键
func generateIdempotencyKey() string {
	return fmt.Sprintf("test-key-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
}

// testIdempotencyPerformance 测试幂等性性能
func testIdempotencyPerformance(
	ctx context.Context,
	idempotentService *service.IdempotentService,
	appLogger *logger.Logger,
) {

	appLogger.Info("开始幂等性性能测试")

	// 测试参数
	testCount := 1000

	// 创建测试请求
	req := &dto.CreateAppointmentRequest{
		CustomerID: uuid.New().String(),
		StaffID:    uuid.New().String(),
		ServiceID:  uuid.New().String(),
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
		Notes:      stringPtr("性能测试预约"),
	}

	// 固定的幂等性键以确保重复请求
	idempotencyKey := fmt.Sprintf("perf-test-%d", time.Now().Unix())

	// 性能测试
	start := time.Now()
	successCount := 0
	failedCount := 0

	for i := 0; i < testCount; i++ {
		_, err := idempotentService.CreateAppointmentWithIdempotency(ctx, req, idempotencyKey)
		if err != nil {
			failedCount++
			appLogger.Error("测试请求失败",
				map[string]interface{}{
					"iteration": i,
					"error":     err,
				})
		} else {
			successCount++
		}

		// 每100次记录一次进度
		if i > 0 && i%100 == 0 {
			appLogger.Info("性能测试进度",
				map[string]interface{}{
					"iteration":     i,
					"success_count": successCount,
					"failed_count":  failedCount,
					"total":         testCount,
				})
		}
	}

	duration := time.Since(start)
	qps := float64(testCount) / duration.Seconds()

	appLogger.Info("幂等性性能测试完成",
		map[string]interface{}{
			"total_requests": testCount,
			"success_count":  successCount,
			"failed_count":   failedCount,
			"duration":       duration,
			"qps":            qps,
			"avg_latency_ms": float64(duration.Nanoseconds()) / float64(testCount) / 1e6,
		})

	// 输出性能建议
	if qps < 100 {
		appLogger.Warn("性能较低，建议优化Redis连接或减少日志记录")
	} else if qps > 1000 {
		appLogger.Info("性能优秀")
	} else {
		appLogger.Info("性能良好")
	}
}

// testIdempotencyConcurrency 测试幂等性并发
func testIdempotencyConcurrency(
	ctx context.Context,
	idempotentService *service.IdempotentService,
	appLogger *logger.Logger,
) {

	appLogger.Info("开始幂等性并发测试")

	// 测试参数
	concurrentGoroutines := 50
	requestsPerGoroutine := 20

	// 创建测试请求
	req := &dto.CreateAppointmentRequest{
		CustomerID: uuid.New().String(),
		StaffID:    uuid.New().String(),
		ServiceID:  uuid.New().String(),
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
		Notes:      stringPtr("并发测试预约"),
	}

	idempotencyKey := fmt.Sprintf("concurrent-test-%d", time.Now().Unix())

	// 创建结果通道
	resultChan := make(chan error, concurrentGoroutines*requestsPerGoroutine)

	// 启动并发goroutines
	start := time.Now()
	for i := 0; i < concurrentGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < requestsPerGoroutine; j++ {
				_, err := idempotentService.CreateAppointmentWithIdempotency(ctx, req, idempotencyKey)
				resultChan <- err
			}
		}(i)
	}

	// 收集结果
	successCount := 0
	failedCount := 0

	totalRequests := concurrentGoroutines * requestsPerGoroutine
	for i := 0; i < totalRequests; i++ {
		err := <-resultChan
		if err != nil {
			failedCount++
		} else {
			successCount++
		}
	}

	duration := time.Since(start)
	qps := float64(successCount) / duration.Seconds()

	appLogger.Info("幂等性并发测试完成",
		map[string]interface{}{
			"concurrent_goroutines":  concurrentGoroutines,
			"requests_per_goroutine": requestsPerGoroutine,
			"total_requests":         totalRequests,
			"success_count":          successCount,
			"failed_count":           failedCount,
			"duration":               duration,
			"qps":                    qps,
		})

	// 验证并发安全
	if failedCount == 0 {
		appLogger.Info("并发测试通过，幂等性机制工作正常")
	} else {
		appLogger.Warn("并发测试发现问题，可能存在并发安全问题",
			map[string]interface{}{
				"failed_rate": float64(failedCount) / float64(totalRequests) * 100,
			})
	}
}
