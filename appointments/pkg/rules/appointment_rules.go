package rules

import (
	"time"

	"github.com/google/uuid"
)

// 预定义的预约业务规则

// CreateAppointmentApprovalRules 创建预约审批规则
func CreateAppointmentApprovalRules() []*Rule {
	return []*Rule{
		{
			ID:          "appointment_time_business_hours",
			Name:        "工作时间预约检查",
			Description: "预约时间必须在营业时间内（9:00-18:00）",
			Priority:    100,
			Conditions: []RuleCondition{
				{
					Field:    "start_time.hour",
					Operator: OpGreaterEqual,
					Value:    9,
					Weight:   0.5,
				},
				{
					Field:    "start_time.hour",
					Operator: OpLessEqual,
					Value:    17,
					Weight:   0.5,
				},
			},
			Action: ActionApprove,
			Status: StatusActive,
			Tags:   []string{"time", "business_hours"},
		},
		{
			ID:          "appointment_advanced_booking",
			Name:        "预约提前时间检查",
			Description: "预约必须提前至少2小时",
			Priority:    90,
			Conditions: []RuleCondition{
				{
					Field:    "start_time",
					Operator: OpGreaterEqual,
					Value:    "now + 2h",
					Weight:   1.0,
				},
			},
			Action: ActionApprove,
			Status: StatusActive,
			Tags:   []string{"time", "advanced_booking"},
		},
		{
			ID:          "appointment_max_advance_days",
			Name:        "最大预约天数限制",
			Description: "预约不能提前超过30天",
			Priority:    80,
			Conditions: []RuleCondition{
				{
					Field:    "start_time",
					Operator: OpLessEqual,
					Value:    "now + 720h", // 30天
					Weight:   1.0,
				},
			},
			Action: ActionApprove,
			Status: StatusActive,
			Tags:   []string{"time", "max_advance"},
		},
		{
			ID:          "appointment_duration_limit",
			Name:        "预约时长限制",
			Description: "单次预约时长不能超过4小时",
			Priority:    70,
			Conditions: []RuleCondition{
				{
					Field:    "duration_minutes",
					Operator: OpLessEqual,
					Value:    240,
					Weight:   1.0,
				},
			},
			Action: ActionApprove,
			Status: StatusActive,
			Tags:   []string{"duration", "limit"},
		},
		{
			ID:          "appointment_vip_customer",
			Name:        "VIP客户优先处理",
			Description: "VIP客户预约自动批准",
			Priority:    110,
			Conditions: []RuleCondition{
				{
					Field:    "customer.vip_level",
					Operator: OpIn,
					Value:    []interface{}{"gold", "platinum", "diamond"},
					Weight:   1.0,
				},
			},
			Action: ActionApprove,
			Parameters: map[string]interface{}{
				"message":  "VIP客户预约已自动批准",
				"approver": "VIP自动审批系统",
			},
			Status: StatusActive,
			Tags:   []string{"customer", "vip"},
		},
		{
			ID:          "appointment_weekend_surcharge",
			Name:        "周末预约附加费检查",
			Description: "周末预约需要人工审核附加费",
			Priority:    60,
			Conditions: []RuleCondition{
				{
					Field:    "start_time.weekday",
					Operator: OpIn,
					Value:    []interface{}{0, 6}, // 周日和周六
					Weight:   1.0,
				},
			},
			Action: ActionReview,
			Parameters: map[string]interface{}{
				"level":     "L2",
				"reason":    "周末预约需要审核附加费",
				"reviewers": []string{"财务部门", "运营部门"},
			},
			Status: StatusActive,
			Tags:   []string{"time", "weekend", "surcharge"},
		},
	}
}

// CreateAppointmentPricingRules 创建预约定价规则
func CreateAppointmentPricingRules() []*Rule {
	return []*Rule{
		{
			ID:          "appointment_peak_hour_pricing",
			Name:        "高峰时段定价",
			Description: "工作日17:00-20:00为高峰时段，需要加收20%服务费",
			Priority:    100,
			Conditions: []RuleCondition{
				{
					Field:    "start_time.weekday",
					Operator: OpBetween,
					Value:    []interface{}{1, 5}, // 周一到周五
					Weight:   0.5,
				},
				{
					Field:    "start_time.hour",
					Operator: OpBetween,
					Value:    []interface{}{17, 20},
					Weight:   0.5,
				},
			},
			Action: ActionModify,
			Parameters: map[string]interface{}{
				"modifications": map[string]interface{}{
					"pricing.service_fee_multiplier": 1.2,
					"pricing.peak_hour_surcharge":    true,
				},
				"target_field": "pricing",
			},
			Status: StatusActive,
			Tags:   []string{"pricing", "peak_hour", "surcharge"},
		},
		{
			ID:          "appointment_holiday_pricing",
			Name:        "节假日定价",
			Description: "法定节假日预约加收50%服务费",
			Priority:    95,
			Conditions: []RuleCondition{
				{
					Field:    "start_time.is_holiday",
					Operator: OpEqual,
					Value:    true,
					Weight:   1.0,
				},
			},
			Action: ActionModify,
			Parameters: map[string]interface{}{
				"modifications": map[string]interface{}{
					"pricing.service_fee_multiplier": 1.5,
					"pricing.holiday_surcharge":      true,
				},
				"target_field": "pricing",
			},
			Status: StatusActive,
			Tags:   []string{"pricing", "holiday", "surcharge"},
		},
		{
			ID:          "appointment_new_customer_discount",
			Name:        "新客户折扣",
			Description: "首次预约的新客户享受10%折扣",
			Priority:    85,
			Conditions: []RuleCondition{
				{
					Field:    "customer.is_new",
					Operator: OpEqual,
					Value:    true,
					Weight:   0.5,
				},
				{
					Field:    "customer.appointment_count",
					Operator: OpEqual,
					Value:    0,
					Weight:   0.5,
				},
			},
			Action: ActionModify,
			Parameters: map[string]interface{}{
				"modifications": map[string]interface{}{
					"pricing.discount_percentage":   0.1,
					"pricing.new_customer_discount": true,
				},
				"target_field": "pricing",
			},
			Status: StatusActive,
			Tags:   []string{"pricing", "discount", "new_customer"},
		},
		{
			ID:          "appointment_bulk_discount",
			Name:        "批量预约折扣",
			Description: "一次性预约5次以上享受15%折扣",
			Priority:    75,
			Conditions: []RuleCondition{
				{
					Field:    "bulk_booking_count",
					Operator: OpGreaterEqual,
					Value:    5,
					Weight:   1.0,
				},
			},
			Action: ActionModify,
			Parameters: map[string]interface{}{
				"modifications": map[string]interface{}{
					"pricing.discount_percentage": 0.15,
					"pricing.bulk_discount":       true,
				},
				"target_field": "pricing",
			},
			Status: StatusActive,
			Tags:   []string{"pricing", "discount", "bulk"},
		},
	}
}

// CreateAppointmentResourceRules 创建预约资源规则
func CreateAppointmentResourceRules() []*Rule {
	return []*Rule{
		{
			ID:          "appointment_staff_availability",
			Name:        "服务人员可用性检查",
			Description: "检查指定服务人员在该时间段是否可用",
			Priority:    120,
			Conditions: []RuleCondition{
				{
					Field:    "staff.availability",
					Operator: OpEqual,
					Value:    true,
					Weight:   1.0,
				},
			},
			Action: ActionApprove,
			Status: StatusActive,
			Tags:   []string{"resource", "staff", "availability"},
		},
		{
			ID:          "appointment_staff_workload",
			Name:        "服务人员工作负载检查",
			Description: "服务人员当天预约不能超过8个",
			Priority:    80,
			Conditions: []RuleCondition{
				{
					Field:    "staff.daily_appointment_count",
					Operator: OpLessEqual,
					Value:    8,
					Weight:   1.0,
				},
			},
			Action: ActionApprove,
			Parameters: map[string]interface{}{
				"message": "服务人员工作负载正常",
			},
			Status: StatusActive,
			Tags:   []string{"resource", "staff", "workload"},
		},
		{
			ID:          "appointment_specialist_requirement",
			Name:        "专家服务要求检查",
			Description: "某些特殊服务必须由专家提供服务",
			Priority:    100,
			Conditions: []RuleCondition{
				{
					Field:    "service.requires_specialist",
					Operator: OpEqual,
					Value:    true,
					Weight:   0.5,
				},
				{
					Field:    "staff.is_specialist",
					Operator: OpEqual,
					Value:    true,
					Weight:   0.5,
				},
			},
			Action: ActionApprove,
			Status: StatusActive,
			Tags:   []string{"resource", "staff", "specialist"},
		},
		{
			ID:          "appointment_equipment_availability",
			Name:        "设备可用性检查",
			Description: "检查所需设备在该时间段是否可用",
			Priority:    90,
			Conditions: []RuleCondition{
				{
					Field:    "equipment.availability",
					Operator: OpEqual,
					Value:    true,
					Weight:   1.0,
				},
			},
			Action: ActionApprove,
			Status: StatusActive,
			Tags:   []string{"resource", "equipment", "availability"},
		},
	}
}

// CreateAppointmentNotificationRules 创建预约通知规则
func CreateAppointmentNotificationRules() []*Rule {
	return []*Rule{
		{
			ID:          "appointment_confirmation_notification",
			Name:        "预约确认通知",
			Description: "预约创建成功后发送确认通知",
			Priority:    50,
			Conditions: []RuleCondition{
				{
					Field:    "status",
					Operator: OpEqual,
					Value:    "created",
					Weight:   1.0,
				},
			},
			Action: ActionNotify,
			Parameters: map[string]interface{}{
				"title":    "预约确认",
				"message":  "您的预约已创建成功",
				"channels": []interface{}{"email", "sms"},
				"priority": "high",
				"template": "appointment_confirmation",
			},
			Status: StatusActive,
			Tags:   []string{"notification", "confirmation"},
		},
		{
			ID:          "appointment_reminder_notification",
			Name:        "预约提醒通知",
			Description: "预约开始前24小时发送提醒通知",
			Priority:    40,
			Conditions: []RuleCondition{
				{
					Field:    "start_time",
					Operator: OpLessEqual,
					Value:    "now + 24h",
					Weight:   0.5,
				},
				{
					Field:    "start_time",
					Operator: OpGreater,
					Value:    "now + 23h",
					Weight:   0.5,
				},
			},
			Action: ActionNotify,
			Parameters: map[string]interface{}{
				"title":    "预约提醒",
				"message":  "您有一个预约将在明天进行",
				"channels": []interface{}{"sms", "push"},
				"priority": "normal",
				"template": "appointment_reminder",
			},
			Status: StatusActive,
			Tags:   []string{"notification", "reminder"},
		},
		{
			ID:          "appointment_cancellation_notification",
			Name:        "预约取消通知",
			Description: "预约取消时发送通知给相关方",
			Priority:    60,
			Conditions: []RuleCondition{
				{
					Field:    "status",
					Operator: OpEqual,
					Value:    "cancelled",
					Weight:   1.0,
				},
			},
			Action: ActionNotify,
			Parameters: map[string]interface{}{
				"title":      "预约取消",
				"message":    "您的预约已取消",
				"channels":   []interface{}{"email", "sms"},
				"priority":   "high",
				"recipients": []interface{}{"customer", "staff", "admin"},
				"template":   "appointment_cancellation",
			},
			Status: StatusActive,
			Tags:   []string{"notification", "cancellation"},
		},
	}
}

// GetAllAppointmentRules 获取所有预约相关规则
func GetAllAppointmentRules() []*Rule {
	var allRules []*Rule

	// 添加各种类型的规则
	allRules = append(allRules, CreateAppointmentApprovalRules()...)
	allRules = append(allRules, CreateAppointmentPricingRules()...)
	allRules = append(allRules, CreateAppointmentResourceRules()...)
	allRules = append(allRules, CreateAppointmentNotificationRules()...)

	// 为所有规则设置基本属性
	for i := range allRules {
		if allRules[i].ID == "" {
			allRules[i].ID = uuid.New().String()
		}
		if allRules[i].CreatedAt.IsZero() {
			allRules[i].CreatedAt = time.Now()
		}
		if allRules[i].UpdatedAt.IsZero() {
			allRules[i].UpdatedAt = time.Now()
		}
		if allRules[i].Version == 0 {
			allRules[i].Version = 1
		}
		if allRules[i].Status == "" {
			allRules[i].Status = StatusActive
		}
	}

	return allRules
}
