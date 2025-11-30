package rules

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// AddRule 添加规则
func (e *DefaultRuleEngine) AddRule(rule *Rule) error {
	if err := e.ValidateRule(rule); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 检查规则是否已存在
	if _, exists := e.rules[rule.ID]; exists {
		return fmt.Errorf("规则已存在: %s", rule.ID)
	}

	// 设置默认值
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	if rule.Status == "" {
		rule.Status = StatusActive
	}
	if rule.Version == 0 {
		rule.Version = 1
	}

	// 保存到内存
	e.rules[rule.ID] = rule

	// 保存到存储
	if e.repository != nil {
		if err := e.repository.Save(context.Background(), rule); err != nil {
			delete(e.rules, rule.ID)
			return fmt.Errorf("保存规则到存储失败: %w", err)
		}
	}

	e.logger.Info("规则添加成功",
		map[string]interface{}{
			"rule_id":   rule.ID,
			"rule_name": rule.Name,
			"priority":  rule.Priority,
		})

	return nil
}

// RemoveRule 删除规则
func (e *DefaultRuleEngine) RemoveRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 检查规则是否存在
	rule, exists := e.rules[ruleID]
	if !exists {
		return fmt.Errorf("规则不存在: %s", ruleID)
	}

	// 从内存删除
	delete(e.rules, ruleID)

	// 从存储删除
	if e.repository != nil {
		if err := e.repository.Delete(context.Background(), ruleID); err != nil {
			// 恢复到内存
			e.rules[ruleID] = rule
			return fmt.Errorf("从存储删除规则失败: %w", err)
		}
	}

	e.logger.Info("规则删除成功",
		map[string]interface{}{
			"rule_id":   ruleID,
			"rule_name": rule.Name,
		})

	return nil
}

// UpdateRule 更新规则
func (e *DefaultRuleEngine) UpdateRule(rule *Rule) error {
	if err := e.ValidateRule(rule); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 检查规则是否存在
	existingRule, exists := e.rules[rule.ID]
	if !exists {
		return fmt.Errorf("规则不存在: %s", rule.ID)
	}

	// 保留原有信息
	rule.CreatedAt = existingRule.CreatedAt
	rule.HitCount = existingRule.HitCount
	rule.LastHitAt = existingRule.LastHitAt
	rule.AvgExecTime = existingRule.AvgExecTime
	rule.UpdatedAt = time.Now()
	rule.Version = existingRule.Version + 1

	// 更新内存
	e.rules[rule.ID] = rule

	// 更新存储
	if e.repository != nil {
		if err := e.repository.Save(context.Background(), rule); err != nil {
			// 恢复原规则
			e.rules[rule.ID] = existingRule
			return fmt.Errorf("更新规则到存储失败: %w", err)
		}
	}

	e.logger.Info("规则更新成功",
		map[string]interface{}{
			"rule_id":   rule.ID,
			"rule_name": rule.Name,
			"version":   rule.Version,
		})

	return nil
}

// GetRule 获取规则
func (e *DefaultRuleEngine) GetRule(ruleID string) (*Rule, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("规则不存在: %s", ruleID)
	}

	// 返回副本以避免外部修改
	ruleCopy := *rule
	return &ruleCopy, nil
}

// ListRules 列出规则
func (e *DefaultRuleEngine) ListRules(filter *RuleFilter) ([]*Rule, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var rules []*Rule

	// 应用过滤器
	for _, rule := range e.rules {
		if e.matchesFilter(rule, filter) {
			ruleCopy := *rule
			rules = append(rules, &ruleCopy)
		}
	}

	// 应用分页
	if filter != nil {
		if filter.Limit > 0 {
			start := filter.Offset
			if start >= len(rules) {
				return []*Rule{}, nil
			}
			end := start + filter.Limit
			if end > len(rules) {
				end = len(rules)
			}
			rules = rules[start:end]
		}
	}

	return rules, nil
}

// matchesFilter 检查规则是否匹配过滤器
func (e *DefaultRuleEngine) matchesFilter(rule *Rule, filter *RuleFilter) bool {
	if filter == nil {
		return true
	}

	// 状态过滤
	if filter.Status != "" && rule.Status != filter.Status {
		return false
	}

	// 动作过滤
	if filter.Action != "" && rule.Action != filter.Action {
		return false
	}

	// 优先级过滤
	if filter.Priority != nil && rule.Priority != *filter.Priority {
		return false
	}

	// 标签过滤
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, filterTag := range filter.Tags {
			for _, ruleTag := range rule.Tags {
				if ruleTag == filterTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

// EnableRule 启用规则
func (e *DefaultRuleEngine) EnableRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return fmt.Errorf("规则不存在: %s", ruleID)
	}

	if rule.Status == StatusActive {
		return nil // 已经是激活状态
	}

	rule.Status = StatusActive
	rule.UpdatedAt = time.Now()

	// 更新存储
	if e.repository != nil {
		if err := e.repository.Save(context.Background(), rule); err != nil {
			rule.Status = StatusInactive // 回滚
			return fmt.Errorf("更新规则状态失败: %w", err)
		}
	}

	e.logger.Info("规则启用成功",
		map[string]interface{}{
			"rule_id":   ruleID,
			"rule_name": rule.Name,
		})

	return nil
}

// DisableRule 禁用规则
func (e *DefaultRuleEngine) DisableRule(ruleID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return fmt.Errorf("规则不存在: %s", ruleID)
	}

	if rule.Status == StatusInactive {
		return nil // 已经是停用状态
	}

	rule.Status = StatusInactive
	rule.UpdatedAt = time.Now()

	// 更新存储
	if e.repository != nil {
		if err := e.repository.Save(context.Background(), rule); err != nil {
			rule.Status = StatusActive // 回滚
			return fmt.Errorf("更新规则状态失败: %w", err)
		}
	}

	e.logger.Info("规则禁用成功",
		map[string]interface{}{
			"rule_id":   ruleID,
			"rule_name": rule.Name,
		})

	return nil
}

// ValidateRule 验证规则
func (e *DefaultRuleEngine) ValidateRule(rule *Rule) error {
	if rule.ID == "" {
		return fmt.Errorf("规则ID不能为空")
	}

	if rule.Name == "" {
		return fmt.Errorf("规则名称不能为空")
	}

	if len(rule.Conditions) == 0 {
		return fmt.Errorf("规则条件不能为空")
	}

	// 验证条件
	for i, condition := range rule.Conditions {
		if condition.Field == "" {
			return fmt.Errorf("条件%d的字段不能为空", i+1)
		}
		if condition.Operator == "" {
			return fmt.Errorf("条件%d的操作符不能为空", i+1)
		}
		if condition.Weight <= 0 || condition.Weight > 1 {
			return fmt.Errorf("条件%d的权重必须在0-1之间", i+1)
		}
	}

	// 验证操作符
	validOperators := map[Operator]bool{
		OpEqual:        true,
		OpNotEqual:     true,
		OpGreater:      true,
		OpGreaterEqual: true,
		OpLess:         true,
		OpLessEqual:    true,
		OpContains:     true,
		OpNotContains:  true,
		OpIn:           true,
		OpNotIn:        true,
		OpRegex:        true,
		OpBetween:      true,
		OpIsNull:       true,
		OpIsNotNull:    true,
	}

	for _, condition := range rule.Conditions {
		if !validOperators[condition.Operator] {
			return fmt.Errorf("不支持的操作符: %s", condition.Operator)
		}
	}

	// 验证动作
	validActions := map[RuleAction]bool{
		ActionApprove: true,
		ActionReject:  true,
		ActionReview:  true,
		ActionModify:  true,
		ActionNotify:  true,
		ActionDelay:   true,
	}

	if !validActions[rule.Action] {
		return fmt.Errorf("不支持的动作: %s", rule.Action)
	}

	return nil
}

// GetStats 获取统计信息
func (e *DefaultRuleEngine) GetStats(ctx context.Context) (*RuleEngineStats, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// 更新基本统计
	e.stats.TotalRules = int64(len(e.rules))
	e.stats.ActiveRules = 0

	var mostHitRules []string
	type ruleHitCount struct {
		id    string
		count int64
	}
	var hitCounts []ruleHitCount

	for _, rule := range e.rules {
		if rule.Status == StatusActive {
			e.stats.ActiveRules++
		}
		hitCounts = append(hitCounts, ruleHitCount{
			id:    rule.ID,
			count: rule.HitCount,
		})
	}

	// 排序获取最常命中的规则（前10个）
	sort.Slice(hitCounts, func(i, j int) bool {
		return hitCounts[i].count > hitCounts[j].count
	})

	for i := 0; i < len(hitCounts) && i < 10; i++ {
		if hitCounts[i].count > 0 {
			mostHitRules = append(mostHitRules, hitCounts[i].id)
		}
	}
	e.stats.MostHitRules = mostHitRules

	// 返回统计信息副本
	statsCopy := *e.stats
	return &statsCopy, nil
}

// ReloadRules 重新加载规则
func (e *DefaultRuleEngine) ReloadRules(ctx context.Context) error {
	if e.repository == nil {
		return fmt.Errorf("未配置规则存储，无法重新加载")
	}

	e.logger.Info("开始重新加载规则")

	// 从存储加载所有规则
	rules, err := e.repository.List(ctx, &RuleFilter{})
	if err != nil {
		return fmt.Errorf("从存储加载规则失败: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 清空现有规则
	oldRules := e.rules
	e.rules = make(map[string]*Rule)

	// 加载新规则
	for _, rule := range rules {
		e.rules[rule.ID] = rule
	}

	e.logger.Info("规则重新加载完成",
		map[string]interface{}{
			"old_count": len(oldRules),
			"new_count": len(e.rules),
		})

	return nil
}

// RegisterActionHandler 注册动作处理器
func (e *DefaultRuleEngine) RegisterActionHandler(handler ActionHandler) {
	for action := range map[RuleAction]bool{
		ActionApprove: true,
		ActionReject:  true,
		ActionReview:  true,
		ActionModify:  true,
		ActionNotify:  true,
		ActionDelay:   true,
	} {
		if handler.CanHandle(action) {
			e.handlers[action] = handler
		}
	}
}

// GetCacheStats 获取缓存统计
func (e *DefaultRuleEngine) GetCacheStats() map[string]interface{} {
	e.cacheMu.RLock()
	defer e.cacheMu.RUnlock()

	return map[string]interface{}{
		"cache_size":      len(e.cache),
		"cache_limit":     e.config.CacheSize,
		"last_clean_time": e.lastCacheClean,
	}
}
