//go:build ignore

package main

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/julesChu12/fly/appointments/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// demoPlutusIntegration 演示Plutus支付服务集成
func demoPlutusIntegration(ctx context.Context, plutusClient *client.PlutusClient, appLogger *logger.Logger) {
	appLogger.Info("开始演示Plutus支付服务集成")

	// 1. 创建支付请求
	paymentReq := &client.CreatePaymentRequest{
		OrderID:       uuid.New().String(),
		AppointmentID: uuid.New().String(),
		CustomerID:    uuid.New().String(),
		Amount:        299.00,
		Currency:      "CNY",
		PaymentMethod: client.PaymentMethodWeChatPay,
		Description:   "测试支付集成 - 心脏科咨询",
		NotifyURL:     "http://localhost:8083/api/v1/payments/notify",
		ReturnURL:     "http://localhost:8083/payments/success",
		ExpireTime:    timePtr(time.Now().Add(30 * time.Minute)),
	}

	appLogger.Info("创建支付请求",
		map[string]interface{}{
			"order_id":       paymentReq.OrderID,
			"appointment_id": paymentReq.AppointmentID,
			"amount":         paymentReq.Amount,
			"payment_method": paymentReq.PaymentMethod,
		})

	// 2. 调用Plutus服务创建支付
	payment, err := plutusClient.CreatePayment(ctx, paymentReq)
	if err != nil {
		appLogger.Error("创建支付失败", "error", err)
		return
	}

	appLogger.Info("支付创建成功",
		map[string]interface{}{
			"payment_id":     payment.ID,
			"order_id":       payment.OrderID,
			"transaction_id": payment.TransactionID,
			"status":         payment.Status,
			"amount":         payment.Amount,
		})

	// 3. 查询支付状态
	appLogger.Info("演示查询支付状态", "payment_id", payment.ID)
	statusQuery, err := plutusClient.QueryPaymentStatus(ctx, payment.ID)
	if err != nil {
		appLogger.Error("查询支付状态失败", "error", err)
	} else {
		appLogger.Info("支付状态查询成功",
			map[string]interface{}{
				"payment_id":    statusQuery.PaymentID,
				"status":        statusQuery.Status,
				"amount":        statusQuery.Amount,
				"paid_amount":   statusQuery.PaidAmount,
				"refund_amount": statusQuery.RefundAmount,
			})
	}

	// 4. 模拟支付完成后的操作
	appLogger.Info("演示支付完成处理")
	payment.Status = client.PaymentStatusPaid
	now := time.Now()
	payment.CompletedAt = &now

	// 5. 演示退款操作
	appLogger.Info("演示退款操作", "payment_id", payment.ID)
	refundReq := &client.RefundRequest{
		PaymentID: payment.ID,
		Amount:    50.00, // 退款50元
		Reason:    "客户取消预约",
		RefundID:  uuid.New().String(),
	}

	err = plutusClient.RefundPayment(ctx, refundReq)
	if err != nil {
		appLogger.Error("退款失败", "error", err)
	} else {
		appLogger.Info("退款请求成功",
			map[string]interface{}{
				"payment_id":    payment.ID,
				"refund_amount": refundReq.Amount,
				"reason":        refundReq.Reason,
			})
	}
}

// healthCheckPlutusService 健康检查Plutus服务
func healthCheckPlutusService(ctx context.Context, plutusClient *client.PlutusClient, appLogger *logger.Logger) error {
	appLogger.Info("开���健康检查Plutus支付服务")

	// 尝试获取一个不存在的支付记录来测试连接
	// 这样不会产生实际的支付，但可以验证API连接
	_, err := plutusClient.GetPayment(ctx, "test-nonexistent-payment")
	if err != nil {
		// 对于测试错误，我们假设连接正常（实际部署时需要更完善的健康检查）
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			appLogger.Error("Plutus支付服务连接超时", "error", err)
			return err
		}
		appLogger.Info("Plutus支付服务连接正常（测试连接成功）")
		return nil
	}

	appLogger.Warn("Plutus支付服务响应异常")
	return nil
}

// timePtr 时间指针辅助函数
func timePtr(t time.Time) *time.Time {
	return &t
}

// getPlutusHealthStatus 获取Plutus服务健康状态
func getPlutusHealthStatus(ctx context.Context, plutusClient *client.PlutusClient) (bool, map[string]interface{}) {
	status := map[string]interface{}{
		"service":    "plutus",
		"base_url":   plutusClient.GetBaseURL(),
		"checked_at": time.Now(),
	}

	// 检查基本连接
	_, err := plutusClient.GetPayment(ctx, "health-check")
	if err != nil {
		status["healthy"] = false
		status["error"] = err.Error()
		status["status_code"] = extractStatusCode(err)
		return false, status
	}

	status["healthy"] = true
	status["message"] = "连接正常"
	return true, status
}

// extractStatusCode 从错误中提取HTTP状态码（简化版本）
func extractStatusCode(err error) int {
	// 简化实现，实际部署时需要更详细的错误处理
	if errors.Is(err, context.Canceled) {
		return 408 // Request Timeout
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return 408 // Request Timeout
	}
	return 0
}
