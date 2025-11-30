package service

import (
	"context"
	"fmt"
	"time"

	"github.com/julesChu12/fly/appointments/internal/domain/dto"
	"github.com/julesChu12/fly/appointments/pkg/rules"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// RuleEngineService 规则引擎服务
type RuleEngineService struct {
	ruleEngine rules.RuleEngine
	logger     *logger.Logger
}

// NewRuleEngineService 创建规则引擎服务
func NewRuleEngineService(
	ruleEngine rules.RuleEngine,
	logger *logger.Logger,
) *RuleEngineService {
	return &RuleEngineService{
		ruleEngine: ruleEngine,
		logger:     logger,
	}
}

// ValidateAppointmentWithRules 使用规则验证预约
func (s *RuleEngineService) ValidateAppointmentWithRules(
	ctx context.Context,
	req *dto.CreateAppointmentRequest,
) (*RuleEngineResult, error) {
	// 构建规则输入数据
	data := s.buildAppointmentData(req)

	s.logger.Info("开始规则引擎验证预约",
		map[string]interface{}{
			"customer_id": req.CustomerID,
			"staff_id":    req.StaffID,
			"start_time":  req.StartTime,
		})

	// 执行规则
	execCtx, err := s.ruleEngine.ExecuteRules(ctx, data)
	if err != nil {
		s.logger.Error("规则引擎执行失败",
			map[string]interface{}{
				"error": err,
			})
		return nil, fmt.Errorf("规则引擎执行失败: %w", err)
	}

	// 分析执行结果
	result := s.analyzeRuleResults(execCtx)

	s.logger.Info("规则引擎验证完成",
		map[string]interface{}{
			"approved":        result.Approved,
			"requires_review": result.RequiresReview,
			"rejected":        result.Rejected,
			"rules_matched":   result.RulesMatched,
			"execution_time":  execCtx.StartTime,
		})

	return result, nil
}

// ValidateAppointmentUpdateWithRules 使用规则验证预约更新
func (s *RuleEngineService) ValidateAppointmentUpdateWithRules(
	ctx context.Context,
	appointmentID string,
	req *dto.UpdateAppointmentRequest,
	currentAppointment *dto.AppointmentResponse,
) (*RuleEngineResult, error) {
	// 构建规则输入数据（包含当前预约信息）
	data := s.buildAppointmentUpdateData(appointmentID, req, currentAppointment)

	s.logger.Info("开始规则引擎验证预约更新",
		map[string]interface{}{
			"appointment_id": appointmentID,
			"has_changes":    len(s.getChangedFields(req)) > 0,
		})

	// 执行规则
	execCtx, err := s.ruleEngine.ExecuteRules(ctx, data)
	if err != nil {
		s.logger.Error("规则引擎执行预约更新验证失败",
			map[string]interface{}{
				"appointment_id": appointmentID,
				"error":          err,
			})
		return nil, fmt.Errorf("规则引擎执行失败: %w", err)
	}

	// 分析执行结果
	result := s.analyzeRuleResults(execCtx)

	s.logger.Info("规则引擎预约更新验证完成",
		map[string]interface{}{
			"appointment_id":  appointmentID,
			"approved":        result.Approved,
			"requires_review": result.RequiresReview,
			"rejected":        result.Rejected,
		})

	return result, nil
}

// ProcessAppointmentPricingWithRules 使用规则处理预约定价
func (s *RuleEngineService) ProcessAppointmentPricingWithRules(
	ctx context.Context,
	appointment *dto.AppointmentResponse,
	basePrice float64,
) (*PricingResult, error) {
	// 构建定价数据
	data := s.buildPricingData(appointment, basePrice)

	s.logger.Info("开始规则引擎处理定价",
		map[string]interface{}{
			"appointment_id": appointment.ID,
			"base_price":     basePrice,
			"service_id":     appointment.ServiceID,
		})

	// 执行定价规则
	execCtx, err := s.ruleEngine.ExecuteRules(ctx, data)
	if err != nil {
		s.logger.Error("规则引擎定价处理失败",
			map[string]interface{}{
				"appointment_id": appointment.ID,
				"error":          err,
			})
		return nil, fmt.Errorf("规则引擎定价处理失败: %w", err)
	}

	// 分析定价结果
	pricingResult := s.analyzePricingResults(execCtx, basePrice)

	s.logger.Info("规则引擎定价处理完成",
		map[string]interface{}{
			"appointment_id":   appointment.ID,
			"final_price":      pricingResult.FinalPrice,
			"original_price":   pricingResult.OriginalPrice,
			"discount_amount":  pricingResult.DiscountAmount,
			"surcharge_amount": pricingResult.SurchargeAmount,
		})

	return pricingResult, nil
}

// TriggerNotificationRules 触发通知规则
func (s *RuleEngineService) TriggerNotificationRules(
	ctx context.Context,
	event string,
	data map[string]interface{},
) ([]*NotificationResult, error) {
	// 添加事件信息到数据中
	data["event"] = event
	data["event_time"] = time.Now()

	s.logger.Info("触发通知规则",
		map[string]interface{}{
			"event":     event,
			"data_keys": len(data),
		})

	// 执行通知规则
	execCtx, err := s.ruleEngine.ExecuteRules(ctx, data)
	if err != nil {
		s.logger.Error("通知规则执行失败",
			map[string]interface{}{
				"event": event,
				"error": err,
			})
		return nil, fmt.Errorf("通知规则执行失败: %w", err)
	}

	// 提取通知结果
	notifications := s.extractNotificationResults(execCtx)

	s.logger.Info("通知规则执行完成",
		map[string]interface{}{
			"event":         event,
			"notifications": len(notifications),
		})

	return notifications, nil
}

// RuleEngineResult 规则引擎执行结果
type RuleEngineResult struct {
	Approved       bool                   `json:"approved"`        // 是否批准
	Rejected       bool                   `json:"rejected"`        // 是否拒绝
	RequiresReview bool                   `json:"requires_review"` // 是否需要审核
	RulesMatched   int                    `json:"rules_matched"`   // 匹配的规则数量
	RuleResults    []*rules.RuleResult    `json:"rule_results"`    // 详细规则结果
	Actions        []string               `json:"actions"`         // 执行的动作
	Messages       []string               `json:"messages"`        // 消息列表
	Modifications  map[string]interface{} `json:"modifications"`   // 修改的数据
	Requires       map[string]interface{} `json:"requires"`        // 额外要求
	ExecutionTime  time.Duration          `json:"execution_time"`  // 执行时间
}

// PricingResult 定价结果
type PricingResult struct {
	OriginalPrice    float64            `json:"original_price"`    // 原价
	FinalPrice       float64            `json:"final_price"`       // 最终价格
	DiscountAmount   float64            `json:"discount_amount"`   // 折扣金额
	SurchargeAmount  float64            `json:"surcharge_amount"`  // 附加费金额
	Discounts        []*PricingModifier `json:"discounts"`         // 折扣列表
	Surcharges       []*PricingModifier `json:"surcharges"`        // 附加费列表
	PricingBreakdown map[string]float64 `json:"pricing_breakdown"` // 价格明细
	AppliedRules     []string           `json:"applied_rules"`     // 应用的规则
}

// PricingModifier 价格修饰符
type PricingModifier struct {
	Type        string  `json:"type"`        // 类型: discount, surcharge
	Name        string  `json:"name"`        // 名称
	Amount      float64 `json:"amount"`      // 金额
	Percentage  float64 `json:"percentage"`  // 百分比
	Description string  `json:"description"` // 描述
	RuleID      string  `json:"rule_id"`     // 规则ID
}

// NotificationResult 通知结果
type NotificationResult struct {
	ID         string                 `json:"id"`         // 通知ID
	Title      string                 `json:"title"`      // 标题
	Message    string                 `json:"message"`    // 消息内容
	Channels   []string               `json:"channels"`   // 通知渠道
	Recipients []string               `json:"recipients"` // 接收人
	Priority   string                 `json:"priority"`   // 优先级
	Data       map[string]interface{} `json:"data"`       // 附加数据
	CreatedAt  time.Time              `json:"created_at"` // 创建时间
	RuleID     string                 `json:"rule_id"`    // 触发规则ID
}

// buildAppointmentData 构建预约数据
func (s *RuleEngineService) buildAppointmentData(req *dto.CreateAppointmentRequest) map[string]interface{} {
	data := map[string]interface{}{
		"customer_id":      req.CustomerID,
		"staff_id":         req.StaffID,
		"service_id":       req.ServiceID,
		"start_time":       req.StartTime,
		"end_time":         req.EndTime,
		"duration_minutes": req.EndTime.Sub(req.StartTime).Minutes(),
		"status":           "pending",
		"created_at":       time.Now(),
	}

	// 添加客户信息（示例数据，实际应从数据库获取）
	data["customer"] = map[string]interface{}{
		"id":                req.CustomerID,
		"is_new":            true, // 实际需要查询客户历史
		"appointment_count": 0,
		"vip_level":         "normal",
	}

	// 添加服务人员信息（示例数据）
	data["staff"] = map[string]interface{}{
		"id":                      req.StaffID,
		"availability":            true,
		"is_specialist":           false,
		"daily_appointment_count": 0,
	}

	// 添加服务信息（示例数据）
	data["service"] = map[string]interface{}{
		"id":                  req.ServiceID,
		"requires_specialist": false,
		"base_price":          100.0,
		"duration":            req.EndTime.Sub(req.StartTime),
	}

	// 添加定价信息
	data["pricing"] = map[string]interface{}{
		"base_price":             100.0,
		"service_fee_multiplier": 1.0,
		"discount_percentage":    0.0,
		"peak_hour_surcharge":    false,
		"holiday_surcharge":      false,
		"new_customer_discount":  false,
		"bulk_discount":          false,
	}

	return data
}

// buildAppointmentUpdateData 构建预约更新数据
func (s *RuleEngineService) buildAppointmentUpdateData(
	appointmentID string,
	req *dto.UpdateAppointmentRequest,
	current *dto.AppointmentResponse,
) map[string]interface{} {
	data := s.buildAppointmentData(&dto.CreateAppointmentRequest{
		CustomerID: current.CustomerID,
		StaffID:    current.StaffID,
		ServiceID:  current.ServiceID,
		StartTime:  current.StartTime,
		EndTime:    current.EndTime,
	})

	data["appointment_id"] = appointmentID
	data["status"] = "updating"
	data["original_data"] = current

	// 添加更新字段
	if req.StartTime != nil {
		data["new_start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		data["new_end_time"] = *req.EndTime
	}
	if req.Notes != nil {
		data["new_notes"] = *req.Notes
	}
	if req.Status != nil {
		data["new_status"] = *req.Status
	}

	return data
}

// buildPricingData 构建定价数据
func (s *RuleEngineService) buildPricingData(appointment *dto.AppointmentResponse, basePrice float64) map[string]interface{} {
	data := s.buildAppointmentData(&dto.CreateAppointmentRequest{
		CustomerID: appointment.CustomerID,
		StaffID:    appointment.StaffID,
		ServiceID:  appointment.ServiceID,
		StartTime:  appointment.StartTime,
		EndTime:    appointment.EndTime,
	})

	data["appointment_id"] = appointment.ID
	if pricing, ok := data["pricing"].(map[string]interface{}); ok {
		pricing["base_price"] = basePrice
		data["pricing"] = pricing
	} else {
		// 如果pricing不存在或类型不对，创建新的pricing map
		data["pricing"] = map[string]interface{}{
			"base_price": basePrice,
		}
	}

	return data
}

// analyzeRuleResults 分析规则执行结果
func (s *RuleEngineService) analyzeRuleResults(execCtx *rules.ExecutionContext) *RuleEngineResult {
	result := &RuleEngineResult{
		RuleResults:   execCtx.RuleResults,
		Actions:       []string{},
		Messages:      []string{},
		Modifications: make(map[string]interface{}),
		Requires:      make(map[string]interface{}),
	}

	for _, ruleResult := range execCtx.RuleResults {
		if ruleResult.Matched {
			result.RulesMatched++

			// 记录动作
			result.Actions = append(result.Actions, string(ruleResult.Action))

			// 处理不同的动作类型
			switch ruleResult.Action {
			case rules.ActionApprove:
				result.Approved = true
				if ruleResult.Result != nil {
					if msg, ok := ruleResult.Result.(map[string]interface{})["message"].(string); ok {
						result.Messages = append(result.Messages, msg)
					}
				}

			case rules.ActionReject:
				result.Rejected = true
				if ruleResult.Result != nil {
					if msg, ok := ruleResult.Result.(map[string]interface{})["reason"].(string); ok {
						result.Messages = append(result.Messages, "拒绝: "+msg)
					}
				}

			case rules.ActionReview:
				result.RequiresReview = true
				if ruleResult.Result != nil {
					if res, ok := ruleResult.Result.(map[string]interface{}); ok {
						result.Requires["review_level"] = res["level"]
						result.Requires["reviewers"] = res["reviewers"]
					}
				}

			case rules.ActionModify:
				if ruleResult.Result != nil {
					if res, ok := ruleResult.Result.(map[string]interface{}); ok {
						if mods, ok := res["modifications"].(map[string]interface{}); ok {
							// 合并修改
							for k, v := range mods {
								result.Modifications[k] = v
							}
						}
					}
				}

			case rules.ActionNotify:
				if ruleResult.Result != nil {
					if msg, ok := ruleResult.Result.(map[string]interface{})["message"].(string); ok {
						result.Messages = append(result.Messages, "通知: "+msg)
					}
				}
			}
		}
	}

	// 计算执行时间
	result.ExecutionTime = time.Since(execCtx.StartTime)

	return result
}

// analyzePricingResults 分析定价结果
func (s *RuleEngineService) analyzePricingResults(execCtx *rules.ExecutionContext, basePrice float64) *PricingResult {
	result := &PricingResult{
		OriginalPrice:    basePrice,
		FinalPrice:       basePrice,
		Discounts:        []*PricingModifier{},
		Surcharges:       []*PricingModifier{},
		PricingBreakdown: make(map[string]float64),
		AppliedRules:     []string{},
	}

	currentPrice := basePrice

	for _, ruleResult := range execCtx.RuleResults {
		if ruleResult.Matched && ruleResult.Action == rules.ActionModify {
			result.AppliedRules = append(result.AppliedRules, ruleResult.RuleID)

			if ruleResult.Result != nil {
				if res, ok := ruleResult.Result.(map[string]interface{}); ok {
					if mods, ok := res["modifications"].(map[string]interface{}); ok {
						// 处理定价修改
						if multiplier, ok := mods["pricing.service_fee_multiplier"].(float64); ok {
							currentPrice = currentPrice * multiplier
						}

						if discount, ok := mods["pricing.discount_percentage"].(float64); ok && discount > 0 {
							discountAmount := currentPrice * discount
							result.Discounts = append(result.Discounts, &PricingModifier{
								Type:        "discount",
								Name:        "规则折扣",
								Amount:      discountAmount,
								Percentage:  discount * 100,
								Description: "规则引擎应用的折扣",
								RuleID:      ruleResult.RuleID,
							})
							currentPrice -= discountAmount
						}
					}
				}
			}
		}
	}

	result.FinalPrice = currentPrice
	result.DiscountAmount = result.OriginalPrice - result.FinalPrice + result.SurchargeAmount
	result.PricingBreakdown["base_price"] = result.OriginalPrice
	result.PricingBreakdown["final_price"] = result.FinalPrice
	result.PricingBreakdown["discount_amount"] = result.DiscountAmount
	result.PricingBreakdown["surcharge_amount"] = result.SurchargeAmount

	return result
}

// extractNotificationResults 提取通知结果
func (s *RuleEngineService) extractNotificationResults(execCtx *rules.ExecutionContext) []*NotificationResult {
	var notifications []*NotificationResult

	for _, ruleResult := range execCtx.RuleResults {
		if ruleResult.Matched && ruleResult.Action == rules.ActionNotify {
			if ruleResult.Result != nil {
				if res, ok := ruleResult.Result.(map[string]interface{}); ok {
					notification := &NotificationResult{
						ID:        res["id"].(string),
						Title:     res["title"].(string),
						Message:   res["message"].(string),
						Priority:  res["priority"].(string),
						Data:      res["data"].(map[string]interface{}),
						CreatedAt: res["created_at"].(time.Time),
						RuleID:    ruleResult.RuleID,
					}

					if channels, ok := res["channels"].([]interface{}); ok {
						for _, ch := range channels {
							if str, ok := ch.(string); ok {
								notification.Channels = append(notification.Channels, str)
							}
						}
					}

					if recipients, ok := res["recipients"].([]interface{}); ok {
						for _, rec := range recipients {
							if str, ok := rec.(string); ok {
								notification.Recipients = append(notification.Recipients, str)
							}
						}
					}

					notifications = append(notifications, notification)
				}
			}
		}
	}

	return notifications
}

// getChangedFields 获取变更的字段
func (s *RuleEngineService) getChangedFields(req *dto.UpdateAppointmentRequest) []string {
	var fields []string
	if req.StartTime != nil {
		fields = append(fields, "start_time")
	}
	if req.EndTime != nil {
		fields = append(fields, "end_time")
	}
	if req.Notes != nil {
		fields = append(fields, "notes")
	}
	if req.Status != nil {
		fields = append(fields, "status")
	}
	return fields
}
