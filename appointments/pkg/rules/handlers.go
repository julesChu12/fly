package rules

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// ApproveActionHandler 批��动作处理器
type ApproveActionHandler struct {
	logger *logger.Logger
}

// NewApproveActionHandler 创建批准动作处理器
func NewApproveActionHandler(logger *logger.Logger) *ApproveActionHandler {
	return &ApproveActionHandler{logger: logger}
}

// CanHandle 检查是否能处理指定动作
func (h *ApproveActionHandler) CanHandle(action RuleAction) bool {
	return action == ActionApprove
}

// Execute 执行动作
func (h *ApproveActionHandler) Execute(ctx context.Context, action RuleAction, parameters map[string]interface{}, data map[string]interface{}) (interface{}, error) {
	h.logger.Debug("执行批准动作",
		map[string]interface{}{
			"parameters": parameters,
			"data_keys":  getKeys(data),
		})

	// 获取批准信息
	message := "请求已批准"
	if msg, ok := parameters["message"].(string); ok {
		message = msg
	}

	// 获取审批人
	approver := "系统自动批准"
	if appr, ok := parameters["approver"].(string); ok {
		approver = appr
	}

	result := map[string]interface{}{
		"approved":   true,
		"message":    message,
		"approver":   approver,
		"approved_at": time.Now(),
		"action":     "approve",
	}

	return result, nil
}

// RejectActionHandler 拒绝动作处理器
type RejectActionHandler struct {
	logger *logger.Logger
}

// NewRejectActionHandler 创建拒绝动作处理器
func NewRejectActionHandler(logger *logger.Logger) *RejectActionHandler {
	return &RejectActionHandler{logger: logger}
}

// CanHandle 检查是否能处理指定动作
func (h *RejectActionHandler) CanHandle(action RuleAction) bool {
	return action == ActionReject
}

// Execute 执行动作
func (h *RejectActionHandler) Execute(ctx context.Context, action RuleAction, parameters map[string]interface{}, data map[string]interface{}) (interface{}, error) {
	h.logger.Debug("执行拒绝动作",
		map[string]interface{}{
			"parameters": parameters,
			"data_keys":  getKeys(data),
		})

	// 获取拒绝原因
	reason := "不符合业务规则"
	if r, ok := parameters["reason"].(string); ok {
		reason = r
	}

	// 获取拒绝代码
	code := "REJECTED"
	if c, ok := parameters["code"].(string); ok {
		code = c
	}

	result := map[string]interface{}{
		"approved":     false,
		"rejected":     true,
		"reason":       reason,
		"code":         code,
		"rejected_at":  time.Now(),
		"action":       "reject",
	}

	return result, nil
}

// ReviewActionHandler 审核动作处理器
type ReviewActionHandler struct {
	logger *logger.Logger
}

// NewReviewActionHandler 创建审核动作处理器
func NewReviewActionHandler(logger *logger.Logger) *ReviewActionHandler {
	return &ReviewActionHandler{logger: logger}
}

// CanHandle 检查是否能处理指定动作
func (h *ReviewActionHandler) CanHandle(action RuleAction) bool {
	return action == ActionReview
}

// Execute 执行动作
func (h *ReviewActionHandler) Execute(ctx context.Context, action RuleAction, parameters map[string]interface{}, data map[string]interface{}) (interface{}, error) {
	h.logger.Debug("执行审核动作",
		map[string]interface{}{
			"parameters": parameters,
			"data_keys":  getKeys(data),
		})

	// 获取审核级别
	level := "L1" // 默认L1审核
	if l, ok := parameters["level"].(string); ok {
		level = l
	}

	// 获取审核人列表
	reviewers := []string{"管理员"}
	if revs, ok := parameters["reviewers"].([]interface{}); ok {
		reviewers = make([]string, len(revs))
		for i, rev := range revs {
			if str, ok := rev.(string); ok {
				reviewers[i] = str
			}
		}
	}

	// 获取截止时间
	var deadline *time.Time
	if dl, ok := parameters["deadline"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, dl); err == nil {
			deadline = &parsed
		}
	}

	result := map[string]interface{}{
		"requires_review": true,
		"level":          level,
		"reviewers":      reviewers,
		"created_at":     time.Now(),
		"action":         "review",
	}

	if deadline != nil {
		result["deadline"] = *deadline
	}

	return result, nil
}

// ModifyActionHandler 修改动作处理器
type ModifyActionHandler struct {
	logger *logger.Logger
}

// NewModifyActionHandler 创建修改动作处理器
func NewModifyActionHandler(logger *logger.Logger) *ModifyActionHandler {
	return &ModifyActionHandler{logger: logger}
}

// CanHandle 检查是否能处理指定动作
func (h *ModifyActionHandler) CanHandle(action RuleAction) bool {
	return action == ActionModify
}

// Execute 执行动作
func (h *ModifyActionHandler) Execute(ctx context.Context, action RuleAction, parameters map[string]interface{}, data map[string]interface{}) (interface{}, error) {
	h.logger.Debug("执行修改动作",
		map[string]interface{}{
			"parameters": parameters,
			"data_keys":  getKeys(data),
		})

	// 获取修改操作
	modifications, ok := parameters["modifications"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("修��动作需要modifications参数")
	}

	// 执行修改
	modified := make(map[string]interface{})
	for key, value := range modifications {
		modified[key] = value

		// 如果指定了目标字段，则修改原数据
		if targetField, ok := parameters["target_field"].(string); ok {
			if strings.HasPrefix(key, targetField+".") {
				actualKey := strings.TrimPrefix(key, targetField+".")
				data[actualKey] = value
			}
		}
	}

	result := map[string]interface{}{
		"modified":       true,
		"modifications":  modifications,
		"original_data":  data,
		"modified_at":    time.Now(),
		"action":         "modify",
	}

	return result, nil
}

// NotifyActionHandler 通知动作处理器
type NotifyActionHandler struct {
	logger *logger.Logger
}

// NewNotifyActionHandler 创建通知动作处理器
func NewNotifyActionHandler(logger *logger.Logger) *NotifyActionHandler {
	return &NotifyActionHandler{logger: logger}
}

// CanHandle 检查是否能处理指定动作
func (h *NotifyActionHandler) CanHandle(action RuleAction) bool {
	return action == ActionNotify
}

// Execute 执行动作
func (h *NotifyActionHandler) Execute(ctx context.Context, action RuleAction, parameters map[string]interface{}, data map[string]interface{}) (interface{}, error) {
	h.logger.Debug("执行通知动作",
		map[string]interface{}{
			"parameters": parameters,
			"data_keys":  getKeys(data),
		})

	// 获取通知内容
	title := "系统通知"
	if t, ok := parameters["title"].(string); ok {
		title = t
	}

	message := "规则引擎触发的通知"
	if msg, ok := parameters["message"].(string); ok {
		message = msg
	}

	// 获取通知渠道
	channels := []string{"system"}
	if chs, ok := parameters["channels"].([]interface{}); ok {
		channels = make([]string, len(chs))
		for i, ch := range chs {
			if str, ok := ch.(string); ok {
				channels[i] = str
			}
		}
	}

	// 获取接收人
	recipients := []string{}
	if recs, ok := parameters["recipients"].([]interface{}); ok {
		recipients = make([]string, len(recs))
		for i, rec := range recs {
			if str, ok := rec.(string); ok {
				recipients[i] = str
			}
		}
	}

	// 创建通知对象
	notification := map[string]interface{}{
		"id":          fmt.Sprintf("notify_%d", time.Now().Unix()),
		"title":       title,
		"message":     message,
		"channels":    channels,
		"recipients":  recipients,
		"data":        data,
		"created_at":  time.Now(),
		"priority":    "normal",
		"action":      "notify",
	}

	// 设置优先级
	if priority, ok := parameters["priority"].(string); ok {
		notification["priority"] = priority
	}

	// 记录通知日志
	h.logger.Info("发送通知",
		map[string]interface{}{
			"notification_id": notification["id"],
			"title":           title,
			"channels":        channels,
			"recipients":      recipients,
		})

	return notification, nil
}

// DelayActionHandler 延迟动作处理器
type DelayActionHandler struct {
	logger *logger.Logger
}

// NewDelayActionHandler 创建延迟动作处理器
func NewDelayActionHandler(logger *logger.Logger) *DelayActionHandler {
	return &DelayActionHandler{logger: logger}
}

// CanHandle 检查是否能处理指定动作
func (h *DelayActionHandler) CanHandle(action RuleAction) bool {
	return action == ActionDelay
}

// Execute 执行动作
func (h *DelayActionHandler) Execute(ctx context.Context, action RuleAction, parameters map[string]interface{}, data map[string]interface{}) (interface{}, error) {
	h.logger.Debug("执行延迟动作",
		map[string]interface{}{
			"parameters": parameters,
			"data_keys":  getKeys(data),
		})

	// 获取延迟时间
	delayDuration := 1 * time.Second // 默认延迟1秒
	if delay, ok := parameters["delay_duration"].(string); ok {
		if parsed, err := time.ParseDuration(delay); err == nil {
			delayDuration = parsed
		}
	} else if delayMs, ok := parameters["delay_ms"].(float64); ok {
		delayDuration = time.Duration(delayMs) * time.Millisecond
	}

	h.logger.Debug("开始延迟",
		map[string]interface{}{
			"duration": delayDuration,
		})

	// 执行延迟
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delayDuration):
		// 延迟完成
	}

	result := map[string]interface{}{
		"delayed":        true,
		"delay_duration": delayDuration,
		"completed_at":   time.Now(),
		"action":         "delay",
	}

	return result, nil
}

// getKeys 获取map的所有键
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// DefaultActionHandlers 返回默认的动作处理器集合
func DefaultActionHandlers(logger *logger.Logger) []ActionHandler {
	return []ActionHandler{
		NewApproveActionHandler(logger),
		NewRejectActionHandler(logger),
		NewReviewActionHandler(logger),
		NewModifyActionHandler(logger),
		NewNotifyActionHandler(logger),
		NewDelayActionHandler(logger),
	}
}