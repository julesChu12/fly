//go:build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/appointments/internal/application/service"
	"github.com/julesChu12/fly/appointments/internal/infrastructure/client"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/google/uuid"
)

// setupRealOrderIntegration 设置真实订单集成
// 这个示例展示了如何在应用启动时配置和使用真实的订单服务
func setupRealOrderIntegration(
	appointmentService *service.AppointmentService,
	eventService *service.EventService,
	appLogger *logger.Logger,
) (*service.OrderIntegrationServiceReal, error) {

	appLogger.Info("开始初始化真实订单集成服务")

	// 1. 创建客户端工厂
	clientFactory := client.NewClientFactory(appLogger)

	// 2. 加载配置
	kratosConfig := client.DefaultKratosClientConfig()
	// 这里可以从配置文件或环境变量加载配置
	// kratosConfig.BaseURL = os.Getenv("KRATOS_BASE_URL")
	// kratosConfig.Timeout = parseDuration(os.Getenv("KRATOS_TIMEOUT"))

	plutusConfig := client.DefaultPlutusClientConfig()

	// 3. 创建客户端
	kratosClient := clientFactory.CreateKratosClient(kratosConfig)
	plutusClient := clientFactory.CreatePlutusClient(plutusConfig)

	// 4. 创建订单集成服务
	orderIntegrationConfig := service.DefaultOrderIntegrationConfig()

	orderService := service.NewOrderIntegrationServiceReal(
		*appointmentService,
		kratosClient,
		plutusClient,
		eventService,
		orderIntegrationConfig,
		appLogger,
	)

	appLogger.Info("真实订单集成服务初始化完成",
		map[string]interface{}{
			"kratos_base_url": kratosConfig.BaseURL,
			"plutus_base_url": plutusConfig.BaseURL,
		})

	return orderService, nil
}

// demoRealOrderIntegration 演示真实订单集成功能
func demoRealOrderIntegration(ctx context.Context, orderService *service.OrderIntegrationServiceReal, appLogger *logger.Logger) {
	appLogger.Info("开始演示真实订单集成功能")

	// 创建预约请求
	req := &dto.CreateAppointmentRequest{
		CustomerID: uuid.New().String(),
		StaffID:    uuid.New().String(),
		ServiceID:  "cardiology-consultation",
		StartTime:  time.Now().Add(24 * time.Hour),
		EndTime:    time.Now().Add(25 * time.Hour),
		Notes:      stringPtr("真实订单集成测试预约"),
	}

	appLogger.Info("创建带订单的预约请求",
		map[string]interface{}{
			"customer_id": req.CustomerID,
			"staff_id":    req.StaffID,
			"service_id":  req.ServiceID,
			"start_time":  req.StartTime,
			"end_time":    req.EndTime,
		})

	// 调用订单集成服务
	response, err := orderService.CreateAppointmentWithOrder(ctx, req)
	if err != nil {
		appLogger.Error("创建带订单的预约失败",
			map[string]interface{}{
				"error": err,
			})
		return
	}

	appLogger.Info("带订单的预约创建成功",
		map[string]interface{}{
			"appointment_id": response.ID,
			"status":         response.Status,
			"created_at":     response.CreatedAt,
		})

	// 演示获取订单状态
	if orderID := extractOrderIDFromResponse(response); orderID != "" {
		appLogger.Info("演示获取订单状态", "order_id", orderID)

		order, err := orderService.GetOrderStatus(ctx, orderID)
		if err != nil {
			appLogger.Error("获取订单状态失败", "error", err)
		} else {
			appLogger.Info("订单状态获取成功",
				map[string]interface{}{
					"order_id":       order.ID,
					"order_number":   order.OrderNumber,
					"status":         order.Status,
					"payment_status": order.PaymentStatus,
					"amount":         order.Amount,
				})
		}
	}
}

// healthCheckOrderServices 健康检查订单服务
func healthCheckOrderServices(
	ctx context.Context,
	kratosClient *client.KratosClient,
	plutusClient *client.PlutusClient,
	appLogger *logger.Logger,
) error {

	appLogger.Info("开始健康检查订单服务")

	// 检查Kratos服务
	kratosHealthy := checkKratosHealth(ctx, kratosClient, appLogger)
	appLogger.Info("Kratos订单服务健康检查", "healthy", kratosHealthy)

	// 检查Plutus服务（下一个任务实现）
	plutusHealthy := checkPlutusHealth(ctx, plutusClient, appLogger)
	appLogger.Info("Plutus支付服务健康检查", "healthy", plutusHealthy)

	if !kratosHealthy || !plutusHealthy {
		return fmt.Errorf("部分订单服务不健康")
	}

	return nil
}

// checkKratosHealth 检查Kratos服务健康状态
func checkKratosHealth(ctx context.Context, kratosClient *client.KratosClient, appLogger *logger.Logger) bool {
	// 这里可以调用Kratos服务的健康检查端点
	// 暂时简化实现，假设服务是健康的

	// appLogger.Debug("检查Kratos服务连接", "base_url", kratosClient.config.BaseURL) // 暂时注释，配置字段是私有的

	// TODO: 实现真实的健康检查
	// 可以通过调用GET /health 端点来检查服务状态
	// 或者尝试一个轻量级的API调用来验证连接

	return true
}

// checkPlutusHealth 检查Plutus服务健康状态
func checkPlutusHealth(ctx context.Context, plutusClient *client.PlutusClient, appLogger *logger.Logger) bool {
	appLogger.Debug("检查Plutus服务连接", "base_url", plutusClient.GetBaseURL())

	// TODO: 下一个任务实现真实的Plutus健康检查

	return true
}

// stringPtr 字符串指针辅助函数
func stringPtr(s string) *string {
	return &s
}

// extractOrderIDFromResponse 从响应中提取订单ID（示例实现）
func extractOrderIDFromResponse(response *dto.AppointmentResponse) string {
	// 这里可以根据实际的响应结构来提取订单ID
	// 暂时返回空字符串，实际实现中应该从response中获取
	return ""
}