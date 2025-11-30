package rules

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// DefaultRuleEngine 默认规则引擎实现
type DefaultRuleEngine struct {
	rules          map[string]*Rule
	mu             sync.RWMutex
	logger         *logger.Logger
	config         *RuleEngineConfig
	repository     RuleRepository
	handlers       map[RuleAction]ActionHandler
	stats          *RuleEngineStats
	cache          map[string]interface{}
	cacheMu        sync.RWMutex
	lastCacheClean time.Time
}

// NewDefaultRuleEngine 创建默认规则引擎
func NewDefaultRuleEngine(
	config *RuleEngineConfig,
	repository RuleRepository,
	logger *logger.Logger,
) *DefaultRuleEngine {
	if config == nil {
		config = DefaultRuleEngineConfig()
	}

	engine := &DefaultRuleEngine{
		rules:          make(map[string]*Rule),
		logger:         logger,
		config:         config,
		repository:     repository,
		handlers:       make(map[RuleAction]ActionHandler),
		stats:          &RuleEngineStats{LastResetTime: time.Now()},
		cache:          make(map[string]interface{}),
		lastCacheClean: time.Now(),
	}

	// 启动缓存清理协程
	go engine.startCacheCleanup()

	return engine
}

// ExecuteRules 执行规则
func (e *DefaultRuleEngine) ExecuteRules(ctx context.Context, data map[string]interface{}) (*ExecutionContext, error) {
	startTime := time.Now()

	// 创建执行上下文
	execCtx := &ExecutionContext{
		Context:      ctx,
		Data:         data,
		Variables:    make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		StartTime:    startTime,
		RuleResults:  make([]*RuleResult, 0),
		TraceEnabled: e.config.EnableTracing,
	}

	e.logger.Debug("开始执行规则",
		map[string]interface{}{
			"data_size":   len(data),
			"rules_count": len(e.rules),
		})

	// 设置超时
	if e.config.DefaultTimeout > 0 {
		var cancel context.CancelFunc
		execCtx.Context, cancel = context.WithTimeout(ctx, e.config.DefaultTimeout)
		defer cancel()
	}

	// 获取激活的规则并按优先级排序
	activeRules := e.getActiveRulesSorted()

	// 并发执行规则（受配置限制）
	e.executeRulesConcurrently(execCtx, activeRules)

	// 更新统计信息
	e.updateExecutionStats(execCtx, time.Since(startTime))

	e.logger.Debug("规则执行完成",
		map[string]interface{}{
			"execution_time": time.Since(startTime),
			"rules_matched":  len(execCtx.RuleResults),
			"total_rules":    len(activeRules),
		})

	return execCtx, nil
}

// executeRulesConcurrently 并发执行规则
func (e *DefaultRuleEngine) executeRulesConcurrently(execCtx *ExecutionContext, rules []*Rule) {
	if e.config.MaxConcurrentRules <= 1 {
		// 串行执行
		e.executeRulesSequentially(execCtx, rules)
		return
	}

	// 并发执行
	ruleChan := make(chan *Rule, len(rules))
	resultChan := make(chan *RuleResult, len(rules))
	var wg sync.WaitGroup

	// 启动worker协程
	workerCount := min(e.config.MaxConcurrentRules, len(rules))
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rule := range ruleChan {
				result := e.executeRule(execCtx, rule)
				resultChan <- result
			}
		}()
	}

	// 分发规则
	for _, rule := range rules {
		ruleChan <- rule
	}
	close(ruleChan)

	// 等待所有worker完成
	wg.Wait()
	close(resultChan)

	// 收集结果
	for result := range resultChan {
		execCtx.RuleResults = append(execCtx.RuleResults, result)
	}
}

// executeRulesSequentially 串行执行规则
func (e *DefaultRuleEngine) executeRulesSequentially(execCtx *ExecutionContext, rules []*Rule) {
	for _, rule := range rules {
		result := e.executeRule(execCtx, rule)
		execCtx.RuleResults = append(execCtx.RuleResults, result)
	}
}

// executeRule 执行单个规则
func (e *DefaultRuleEngine) executeRule(execCtx *ExecutionContext, rule *Rule) *RuleResult {
	startTime := time.Now()
	result := &RuleResult{
		RuleID:     rule.ID,
		RuleName:   rule.Name,
		Action:     rule.Action,
		Timestamp:  startTime,
		Conditions: make([]bool, len(rule.Conditions)),
	}

	defer func() {
		result.ExecTime = time.Since(startTime)
		if r := recover(); r != nil {
			result.Error = fmt.Errorf("规则执行异常: %v", r)
			e.logger.Error("规则执行异常",
				map[string]interface{}{
					"rule_id": rule.ID,
					"panic":   r,
				})
		}
	}()

	// 执行条件检查
	score, matched := e.evaluateConditions(execCtx, rule.Conditions, result.Conditions)
	result.Score = score
	result.Matched = matched

	// 更新规则统计
	e.updateRuleStats(rule, matched, result.ExecTime)

	if matched {
		// 执行动作
		actionResult, err := e.executeAction(execCtx, rule.Action, rule.Parameters)
		result.Result = actionResult
		result.Error = err

		if err != nil {
			e.logger.Error("规则动作执行失败",
				map[string]interface{}{
					"rule_id": rule.ID,
					"action":  rule.Action,
					"error":   err,
				})
		}
	}

	return result
}

// evaluateConditions 评估规则条件
func (e *DefaultRuleEngine) evaluateConditions(execCtx *ExecutionContext, conditions []RuleCondition, conditionResults []bool) (float64, bool) {
	if len(conditions) == 0 {
		return 0.0, false
	}

	totalWeight := 0.0
	matchedWeight := 0.0
	allMatched := true

	for i, condition := range conditions {
		fieldValue := e.getFieldValue(execCtx, condition.Field)
		matched := e.evaluateCondition(fieldValue, condition.Operator, condition.Value)
		conditionResults[i] = matched

		totalWeight += condition.Weight
		if matched {
			matchedWeight += condition.Weight
		} else {
			allMatched = false
		}
	}

	// 计算匹配分数
	score := 0.0
	if totalWeight > 0 {
		score = matchedWeight / totalWeight
	}

	// 根据策略决定是否匹配（这里使用AND策略：所有条件都必须匹配）
	isMatched := allMatched && totalWeight > 0

	return score, isMatched
}

// evaluateCondition 评估单个条件
func (e *DefaultRuleEngine) evaluateCondition(fieldValue interface{}, operator Operator, compareValue interface{}) bool {
	switch operator {
	case OpEqual:
		return e.compareEqual(fieldValue, compareValue)
	case OpNotEqual:
		return !e.compareEqual(fieldValue, compareValue)
	case OpGreater:
		return e.compareGreater(fieldValue, compareValue)
	case OpGreaterEqual:
		return e.compareGreaterEqual(fieldValue, compareValue)
	case OpLess:
		return e.compareLess(fieldValue, compareValue)
	case OpLessEqual:
		return e.compareLessEqual(fieldValue, compareValue)
	case OpContains:
		return e.compareContains(fieldValue, compareValue)
	case OpNotContains:
		return !e.compareContains(fieldValue, compareValue)
	case OpIn:
		return e.compareIn(fieldValue, compareValue)
	case OpNotIn:
		return !e.compareIn(fieldValue, compareValue)
	case OpRegex:
		return e.compareRegex(fieldValue, compareValue)
	case OpBetween:
		return e.compareBetween(fieldValue, compareValue)
	case OpIsNull:
		return fieldValue == nil
	case OpIsNotNull:
		return fieldValue != nil
	default:
		return false
	}
}

// getFieldValue 获取字段值
func (e *DefaultRuleEngine) getFieldValue(execCtx *ExecutionContext, fieldPath string) interface{} {
	// 支持嵌套字段访问，如 "user.profile.age"
	parts := strings.Split(fieldPath, ".")
	var value interface{} = execCtx.Data

	for _, part := range parts {
		if mapValue, ok := value.(map[string]interface{}); ok {
			value = mapValue[part]
		} else {
			return nil
		}
	}

	return value
}

// 比较函数实现
func (e *DefaultRuleEngine) compareEqual(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func (e *DefaultRuleEngine) compareGreater(a, b interface{}) bool {
	return e.compareNumeric(a, b, func(a, b float64) bool { return a > b })
}

func (e *DefaultRuleEngine) compareGreaterEqual(a, b interface{}) bool {
	return e.compareNumeric(a, b, func(a, b float64) bool { return a >= b })
}

func (e *DefaultRuleEngine) compareLess(a, b interface{}) bool {
	return e.compareNumeric(a, b, func(a, b float64) bool { return a < b })
}

func (e *DefaultRuleEngine) compareLessEqual(a, b interface{}) bool {
	return e.compareNumeric(a, b, func(a, b float64) bool { return a <= b })
}

func (e *DefaultRuleEngine) compareContains(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Contains(aStr, bStr)
}

func (e *DefaultRuleEngine) compareIn(a, b interface{}) bool {
	if list, ok := b.([]interface{}); ok {
		for _, item := range list {
			if e.compareEqual(a, item) {
				return true
			}
		}
	}
	return false
}

func (e *DefaultRuleEngine) compareRegex(a, b interface{}) bool {
	pattern, err := regexp.Compile(fmt.Sprintf("%v", b))
	if err != nil {
		return false
	}
	return pattern.MatchString(fmt.Sprintf("%v", a))
}

func (e *DefaultRuleEngine) compareBetween(a, b interface{}) bool {
	if rangeValue, ok := b.([]interface{}); ok && len(rangeValue) == 2 {
		return e.compareGreaterEqual(a, rangeValue[0]) && e.compareLessEqual(a, rangeValue[1])
	}
	return false
}

func (e *DefaultRuleEngine) compareNumeric(a, b interface{}, compareFunc func(float64, float64) bool) bool {
	aFloat, aOk := e.toFloat64(a)
	bFloat, bOk := e.toFloat64(b)
	return aOk && bOk && compareFunc(aFloat, bFloat)
}

func (e *DefaultRuleEngine) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}

// executeAction 执行动作
func (e *DefaultRuleEngine) executeAction(execCtx *ExecutionContext, action RuleAction, parameters map[string]interface{}) (interface{}, error) {
	if handler, exists := e.handlers[action]; exists {
		return handler.Execute(execCtx.Context, action, parameters, execCtx.Data)
	}

	// 默认动作处理
	switch action {
	case ActionApprove:
		return map[string]interface{}{"approved": true}, nil
	case ActionReject:
		return map[string]interface{}{"approved": false}, nil
	case ActionReview:
		return map[string]interface{}{"requires_review": true}, nil
	case ActionDelay:
		delay := parameters["delay_duration"]
		if delayStr, ok := delay.(string); ok {
			if duration, err := time.ParseDuration(delayStr); err == nil {
				time.Sleep(duration)
			}
		}
		return map[string]interface{}{"delayed": true}, nil
	default:
		return nil, fmt.Errorf("不支持的动作: %s", action)
	}
}

// getActiveRulesSorted 获取激活的规则并按优先级排序
func (e *DefaultRuleEngine) getActiveRulesSorted() []*Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var activeRules []*Rule
	for _, rule := range e.rules {
		if rule.Status == StatusActive {
			activeRules = append(activeRules, rule)
		}
	}

	// 按优先级排序（降序）
	sort.Slice(activeRules, func(i, j int) bool {
		return activeRules[i].Priority > activeRules[j].Priority
	})

	return activeRules
}

// updateRuleStats 更新规则统计信息
func (e *DefaultRuleEngine) updateRuleStats(rule *Rule, matched bool, execTime time.Duration) {
	rule.HitCount++
	if matched {
		rule.LastHitAt = &[]time.Time{time.Now()}[0]
	}

	// 更新平均执行时间
	rule.AvgExecTime = (rule.AvgExecTime + execTime.Seconds()) / 2
}

// updateExecutionStats 更新执行统计信息
func (e *DefaultRuleEngine) updateExecutionStats(execCtx *ExecutionContext, duration time.Duration) {
	e.stats.TotalExecutions++

	// 计算成功的执行次数
	successCount := 0
	for _, result := range execCtx.RuleResults {
		if result.Error == nil {
			successCount++
		}
	}

	if successCount == len(execCtx.RuleResults) {
		e.stats.SuccessfulExecs++
	} else {
		e.stats.FailedExecs++
	}

	// 更新平均执行时间
	e.stats.AvgExecTime = (e.stats.AvgExecTime + float64(duration.Milliseconds())) / 2
}

// startCacheCleanup 启动缓存清理协程
func (e *DefaultRuleEngine) startCacheCleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.cleanCache()
		}
	}
}

// cleanCache 清理过期缓存
func (e *DefaultRuleEngine) cleanCache() {
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	// 简单的定期清理，实际应用中可以根据TTL进行更精确的清理
	if len(e.cache) > e.config.CacheSize {
		// 清理一半的缓存
		toClean := len(e.cache) / 2
		count := 0
		for key := range e.cache {
			delete(e.cache, key)
			count++
			if count >= toClean {
				break
			}
		}
	}

	e.lastCacheClean = time.Now()
}

// min 返回两个整数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
