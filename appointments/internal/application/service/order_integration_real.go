package service

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/appointments/internal/domain/entity"
	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/appointments/internal/infrastructure/client"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/appointments/pkg/errors"
	"github.com/google/uuid"
)

// OrderIntegrationConfig 订单集成配置
type OrderIntegrationConfig struct {
	EnableAutoPayment    bool          `yaml:"enable_auto_payment"`     // 是否启用自动支付
	PaymentTimeout       time.Duration `yaml:"payment_timeout"`         // 支付超时时间
	OrderTimeout         time.Duration `yaml:"order_timeout"`           // 订单超时时间
	MaxRetries           int           `yaml:"max_retries"`             // 最大重试次数
	RetryDelay           time.Duration `yaml:"retry_delay"`             // 重试延迟
	EnableNotifications  bool          `yaml:"enable_notifications"`   // 是否启用通知
	CleanupInterval      time.Duration `yaml:"cleanup_interval"`        // 清理间隔
	RetentionPeriod      time.Duration `yaml:"retention_period"`        // 保留期限
}

// DefaultOrderIntegrationConfig 默认配置
func DefaultOrderIntegrationConfig() *OrderIntegrationConfig {
	return &OrderIntegrationConfig{
		EnableAutoPayment:   true,
		PaymentTimeout:      30 * time.Minute,
		OrderTimeout:        60 * time.Minute,
		MaxRetries:          3,
		RetryDelay:          5 * time.Second,
		EnableNotifications: true,
		CleanupInterval:     1 * time.Hour,
		RetentionPeriod:     7 * 24 * time.Hour,
	}
}

// OrderIntegrationServiceReal 真实订单集成服务
// 使用真实的Kratos和Plutus服务
type OrderIntegrationServiceReal struct {
	appointmentService AppointmentService
	kratosClient       *client.KratosClient  // Kratos订单服务客户端
	plutusClient       *client.PlutusClient  // Plutus支付服务客户端（下一个任务实现）
	eventService       *EventService
	logger             *logger.Logger
	config             *OrderIntegrationConfig
}

// NewOrderIntegrationServiceReal 创建真实订单集成服务
func NewOrderIntegrationServiceReal(
	appointmentService AppointmentService,
	kratosClient *client.KratosClient,
	plutusClient *client.PlutusClient, // 可选参数，暂时为nil
	eventService *EventService,
	config *OrderIntegrationConfig,
	log *logger.Logger,
) *OrderIntegrationServiceReal {
	if config == nil {
		config = DefaultOrderIntegrationConfig()
	}

	return &OrderIntegrationServiceReal{
		appointmentService: appointmentService,
		kratosClient:       kratosClient,
		plutusClient:       plutusClient,
		eventService:       eventService,
		logger:             log,
		config:             config,
	}
}

// CreateAppointmentWithOrder 创建带订单的预约（真实实现）
func (s *OrderIntegrationServiceReal) CreateAppointmentWithOrder(
	ctx context.Context,
	req *dto.CreateAppointmentRequest,
) (*dto.AppointmentResponse, error) {

	s.logger.Info("开始创建带订单的预约",
		map[string]interface{}{
			"customer_id": req.CustomerID,
			"staff_id":    req.StaffID,
			"service_id":  req.ServiceID,
		})

	// 第一步：验证请求参数
	if err := s.validateCreateRequest(req); err != nil {
		s.logger.Error("请求参数验证失败", "error", err)
		return nil, fmt.Errorf("请求参数验证失败: %w", err)
	}

	// 第二步：检查员工可用性
	if err := s.checkStaffAvailability(ctx, req.StaffID, req.StartTime, req.EndTime); err != nil {
		s.logger.Error("员工可用性检查失败", "error", err)
		return nil, fmt.Errorf("员工可用性检查失败: %w", err)
	}

	// 第三步：创建预约记录
	appointment, err := s.createAppointmentRecord(ctx, req)
	if err != nil {
		s.logger.Error("创建预约记录失败", "error", err)
		return nil, fmt.Errorf("创建预约记录失败: %w", err)
	}

	// 第四步：创建订单记录（使用真实Kratos服务）
	order, err := s.createOrderRecord(ctx, appointment)
	if err != nil {
		// 回滚：删除已创建的预约
		if rollbackErr := s.rollbackAppointmentCreation(ctx, appointment.ID.String()); rollbackErr != nil {
			s.logger.Error("回滚预约创建失败", "appointment_id", appointment.ID, "error", rollbackErr)
		}
		return nil, fmt.Errorf("创建订单记录失败: %w", err)
	}

	// 第五步：创建支付记录（暂时仍使用Mock，下一个任务实现真实集成）
	payment, err := s.createPaymentRecord(ctx, order)
	if err != nil {
		// 回滚：删除订单和预约
		if rollbackErr := s.rollbackOrderAndAppointment(ctx, appointment.ID.String(), order.ID); rollbackErr != nil {
			s.logger.Error("回滚订单和预约失败", "appointment_id", appointment.ID, "order_id", order.ID, "error", rollbackErr)
		}
		return nil, fmt.Errorf("创建支付记录失败: %w", err)
	}

	// 第六步：发布事件通知
	if s.config.EnableNotifications && s.eventService != nil {
		if err := s.publishAppointmentCreatedEvent(ctx, appointment, order, payment); err != nil {
			s.logger.Error("发布预约创建事件失败", "error", err)
			// 不返回错误，因为业务操作已成功
		}
	}

	// 第七步：构建并返回响应
	response := s.buildAppointmentResponse(appointment, order, payment)

	s.logger.Info("带订单的预约创建成功",
		map[string]interface{}{
			"appointment_id": appointment.ID,
			"order_id":        order.ID,
			"payment_id":      payment.ID,
			"amount":          order.Amount,
		})

	return response, nil
}

// validateCreateRequest 验证创建请求参数
func (s *OrderIntegrationServiceReal) validateCreateRequest(req *dto.CreateAppointmentRequest) error {
	if req.CustomerID == "" {
		return errors.ValidationError("customer_id", "", "不能为空")
	}
	if req.StaffID == "" {
		return errors.ValidationError("staff_id", "", "不能为空")
	}
	if req.ServiceID == "" {
		return errors.ValidationError("service_id", "", "不能为空")
	}
	if req.StartTime.IsZero() {
		return errors.ValidationError("start_time", "", "不能为空")
	}
	if req.EndTime.IsZero() {
		return errors.ValidationError("end_time", "", "不能为空")
	}
	if req.EndTime.Before(req.StartTime) {
		return errors.ValidationError("end_time", req.EndTime, "必须晚于开始时间")
	}
	if req.StartTime.Before(time.Now()) {
		return errors.ValidationError("start_time", req.StartTime, "不能早于当前时间")
	}

	return nil
}

// checkStaffAvailability 检查员工可用性
func (s *OrderIntegrationServiceReal) checkStaffAvailability(ctx context.Context, staffID string, startTime, endTime time.Time) error {
	// 这里应该调用Staff服务检查员工可用性
	// 暂时简化实现
	s.logger.Debug("检查员工可用性",
		map[string]interface{}{
			"staff_id":   staffID,
			"start_time": startTime,
			"end_time":   endTime,
		})
	return nil
}

// createAppointmentRecord 创建预约记录
func (s *OrderIntegrationServiceReal) createAppointmentRecord(ctx context.Context, req *dto.CreateAppointmentRequest) (*entity.Appointment, error) {
	// 验证UUID格式
	if _, err := uuid.Parse(req.CustomerID); err != nil {
		return nil, errors.ValidationError("customer_id", req.CustomerID, "无效的UUID格式")
	}

	if _, err := uuid.Parse(req.StaffID); err != nil {
		return nil, errors.ValidationError("staff_id", req.StaffID, "无效的UUID格式")
	}

	if _, err := uuid.Parse(req.ServiceID); err != nil {
		return nil, errors.ValidationError("service_id", req.ServiceID, "无效的UUID格式")
	}

	// 调用预约服务创建
	createdAppointment, err := s.appointmentService.CreateAppointment(req)
	if err != nil {
		return nil, err
	}

	return createdAppointment, nil
}

// createOrderRecord 创建订单记录（使用真实Kratos服务）
func (s *OrderIntegrationServiceReal) createOrderRecord(ctx context.Context, appointment *entity.Appointment) (*client.Order, error) {
	s.logger.Debug("开始创建订单记录",
		map[string]interface{}{
			"appointment_id": appointment.ID.String(),
			"customer_id":    appointment.CustomerID.String(),
		})

	// 计算服务价格
	amount := s.calculateServicePrice(appointment.ServiceID.String())

	// 创建订单请求
	orderReq := client.CreateOrderRequestFromAppointment(appointment, amount)

	// 调用Kratos服务创建订单
	order, err := s.kratosClient.CreateOrder(ctx, orderReq)
	if err != nil {
		s.logger.Error("调用Kratos服务创建订单失败",
			map[string]interface{}{
				"appointment_id": appointment.ID.String(),
				"error": err,
			})
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	s.logger.Info("订单记录创建成功",
		map[string]interface{}{
			"order_id":     order.ID,
			"order_number": order.OrderNumber,
			"amount":       order.Amount,
			"status":       order.Status,
		})

	return order, nil
}

// createPaymentRecord 创建支付记录（使用真实Plutus服务）
func (s *OrderIntegrationServiceReal) createPaymentRecord(ctx context.Context, order *client.Order) (*client.Payment, error) {
	s.logger.Debug("开始创建支付记录",
		map[string]interface{}{
			"order_id": order.ID,
			"amount":   order.Amount,
		})

	// 如果Plutus客户端为nil，使用Mock实现
	if s.plutusClient == nil {
		s.logger.Warn("Plutus客户端未配置，使用Mock实现")
		return s.createMockPaymentRecord(order)
	}

	// 创建支付请求
	paymentReq := client.CreatePaymentRequestFromOrder(order, client.PaymentMethodWeChatPay)

	// 调用Plutus服务创建支付
	payment, err := s.plutusClient.CreatePayment(ctx, paymentReq)
	if err != nil {
		s.logger.Error("调用Plutus服务创建支付失败",
			map[string]interface{}{
				"order_id": order.ID,
				"error":    err,
			})
		return nil, fmt.Errorf("创建支付记录失败: %w", err)
	}

	s.logger.Info("支付记录创建成功",
		map[string]interface{}{
			"payment_id":     payment.ID,
			"order_id":       payment.OrderID,
			"transaction_id": payment.TransactionID,
			"amount":         payment.Amount,
			"payment_method": payment.PaymentMethod,
		})

	return payment, nil
}

// generateTransactionID 生成交易ID
func generateTransactionID() string {
	return fmt.Sprintf("TXN%d%s", time.Now().Unix(), uuid.New().String()[:8])
}

// createMockPaymentRecord 创建Mock支付记录（备用实现）
func (s *OrderIntegrationServiceReal) createMockPaymentRecord(order *client.Order) (*client.Payment, error) {
	payment := &client.Payment{
		ID:            generateTransactionID(),
		OrderID:       order.ID,
		AppointmentID: order.AppointmentID,
		CustomerID:    order.CustomerID,
		Amount:        order.Amount,
		Currency:      order.Currency,
		Status:        client.PaymentStatusPending,
		PaymentMethod: client.PaymentMethodWeChatPay,
		TransactionID: generateTransactionID(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.logger.Debug("Mock支付记录创建成功",
		map[string]interface{}{
			"payment_id": payment.ID,
			"order_id":   order.ID,
			"amount":     payment.Amount,
		})

	return payment, nil
}

// rollbackAppointmentCreation 回滚预约创建
func (s *OrderIntegrationServiceReal) rollbackAppointmentCreation(ctx context.Context, appointmentID string) error {
	// 删除预约记录
	err := s.appointmentService.DeleteAppointment(appointmentID)
	if err != nil {
		s.logger.Error("删除预约记录失败",
			map[string]interface{}{
				"appointment_id": appointmentID,
				"error": err,
			})
		return err
	}

	s.logger.Warn("预约创建已回滚", "appointment_id", appointmentID)
	return nil
}

// rollbackOrderAndAppointment 回滚订单和预约
func (s *OrderIntegrationServiceReal) rollbackOrderAndAppointment(ctx context.Context, appointmentID, orderID string) error {
	// 删除订单记录（调用Kratos服务）
	err := s.kratosClient.CancelOrder(ctx, orderID, "预约创建失败，自动取消订单")
	if err != nil {
		s.logger.Error("取消订单失败",
			map[string]interface{}{
				"order_id": orderID,
				"error": err,
			})
		// 继续删除预约，即使订单取消失败
	}

	// 删除预约记录
	if err := s.rollbackAppointmentCreation(ctx, appointmentID); err != nil {
		return err
	}

	return nil
}

// publishAppointmentCreatedEvent 发布预约创建事件
func (s *OrderIntegrationServiceReal) publishAppointmentCreatedEvent(ctx context.Context, appointment *entity.Appointment, order *client.Order, payment interface{}) error {
	if s.eventService == nil {
		return nil
	}

	// 发布预约创建事件
	err := s.eventService.PublishAppointmentCreated(ctx, appointment, order)
	if err != nil {
		s.logger.Error("发布预约创建事件失败", "error", err)
		return err
	}

	// 可以在这里添加更多事件发布逻辑

	return nil
}

// buildAppointmentResponse 构建预约响应
func (s *OrderIntegrationServiceReal) buildAppointmentResponse(appointment *entity.Appointment, order *client.Order, payment *client.Payment) *dto.AppointmentResponse {
	response := &dto.AppointmentResponse{
		ID:              appointment.ID.String(),
		CustomerID:      appointment.CustomerID.String(),
		StaffID:         appointment.StaffID.String(),
		ServiceID:       appointment.ServiceID.String(),
		StartTime:       appointment.StartTime,
		EndTime:         appointment.EndTime,
		Status:          appointment.Status,
		Notes:           appointment.Notes,
		CreatedAt:       appointment.CreatedAt,
		UpdatedAt:       appointment.UpdatedAt,
	}

	// 添加订单信息到响应
	if order != nil {
		// 可以在响应中包含订单信息，或者创建扩展的响应结构
		s.logger.Debug("包含订单信息在响应中",
			map[string]interface{}{
				"order_id":     order.ID,
				"order_number": order.OrderNumber,
			})
	}

	// 添加支付信息到响应
	if payment != nil {
		s.logger.Debug("包含支付信息在响应中",
			map[string]interface{}{
				"payment_id":     payment.ID,
				"payment_method": payment.PaymentMethod,
				"transaction_id": payment.TransactionID,
			})
	}

	return response
}

// calculateServicePrice 计算服务价格（与原实现相同）
func (s *OrderIntegrationServiceReal) calculateServicePrice(serviceID string) float64 {
	priceMap := map[string]float64{
		"cardiology-consultation": 300.00,
		"general-consultation":   150.00,
		"specialist-consultation": 500.00,
		"health-checkup":         800.00,
		"vaccination":            200.00,
		"dental-consultation":     400.00,
		"pediatric-consultation":   350.00,
		"orthopedics-consultation": 450.00,
	}

	if price, exists := priceMap[serviceID]; exists {
		return price
	}

	return 200.00
}

// generateOrderNumber 生成订单号（与原实现相同）
func (s *OrderIntegrationServiceReal) generateOrderNumber() string {
	timestamp := time.Now().Format("20060102150405")
	random := uuid.New().String()[:8]
	return fmt.Sprintf("ORD%s%s", timestamp, random)
}

// generateTransactionID 生成交易ID（与原实现相同）
func (s *OrderIntegrationServiceReal) generateTransactionID() string {
	timestamp := time.Now().Format("20060102150405")
	random := time.Now().Nanosecond() / 1000000
	return fmt.Sprintf("TXN%s%d", timestamp, random)
}

// GetOrderStatus 获取订单状态
func (s *OrderIntegrationServiceReal) GetOrderStatus(ctx context.Context, orderID string) (*client.Order, error) {
	order, err := s.kratosClient.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("获取订单状态失败: %w", err)
	}
	return order, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderIntegrationServiceReal) UpdateOrderStatus(ctx context.Context, orderID string, status client.OrderStatus, paymentStatus client.PaymentStatus) error {
	req := &client.UpdateOrderStatusRequest{
		Status:        string(status),
		PaymentStatus: string(paymentStatus),
	}

	err := s.kratosClient.UpdateOrderStatus(ctx, orderID, req)
	if err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	s.logger.Info("订单状态更新成功",
		map[string]interface{}{
			"order_id":      orderID,
			"status":        status,
			"payment_status": paymentStatus,
		})

	return nil
}

// GetPaymentStatus 获取支付状态
func (s *OrderIntegrationServiceReal) GetPaymentStatus(ctx context.Context, paymentID string) (*client.Payment, error) {
	if s.plutusClient == nil {
		s.logger.Warn("Plutus客户端未配置，无法获取支付状态")
		return nil, fmt.Errorf("Plutus客户端未配置")
	}

	payment, err := s.plutusClient.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("获取支付状态失败: %w", err)
	}

	return payment, nil
}

// QueryPaymentStatus 查询支付状态详情
func (s *OrderIntegrationServiceReal) QueryPaymentStatus(ctx context.Context, paymentID string) (*client.PaymentStatusQuery, error) {
	if s.plutusClient == nil {
		s.logger.Warn("Plutus客户端未配置，无法查询支付状态")
		return nil, fmt.Errorf("Plutus客户端未配置")
	}

	statusQuery, err := s.plutusClient.QueryPaymentStatus(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("查询支付状态失败: %w", err)
	}

	return statusQuery, nil
}

// RefundPayment 退款
func (s *OrderIntegrationServiceReal) RefundPayment(ctx context.Context, paymentID string, amount float64, reason string) error {
	if s.plutusClient == nil {
		s.logger.Warn("Plutus客户端未配置，无法进行退款")
		return fmt.Errorf("Plutus客户端未配置")
	}

	req := &client.RefundRequest{
		PaymentID: paymentID,
		Amount:    amount,
		Reason:    reason,
	}

	err := s.plutusClient.RefundPayment(ctx, req)
	if err != nil {
		return fmt.Errorf("退款失败: %w", err)
	}

	s.logger.Info("退款请求成功",
		map[string]interface{}{
			"payment_id": paymentID,
			"amount":     amount,
			"reason":     reason,
		})

	return nil
}