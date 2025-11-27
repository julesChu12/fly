package rules

import (
	"context"
	"time"
)

// RuleAction 规则动作类型
type RuleAction string

const (
	ActionApprove RuleAction = "approve"     // 批准
	ActionReject  RuleAction = "reject"      // 拒绝
	ActionReview  RuleAction = "review"      // 人工审核
	ActionModify  RuleAction = "modify"      // 修改数据
	ActionNotify  RuleAction = "notify"      // 发送通知
	ActionDelay   RuleAction = "delay"       // 延迟处理
)

// RuleStatus 规则状态
type RuleStatus string

const (
	StatusActive   RuleStatus = "active"     // 激活
	StatusInactive RuleStatus = "inactive"   // 停用
	StatusTesting  RuleStatus = "testing"    // 测试中
)

// Operator 操作符类型
type Operator string

const (
	OpEqual         Operator = "eq"         // 等于
	OpNotEqual      Operator = "ne"         // 不���于
	OpGreater       Operator = "gt"         // 大于
	OpGreaterEqual  Operator = "gte"        // 大于等于
	OpLess          Operator = "lt"         // 小于
	OpLessEqual     Operator = "lte"        // 小于等于
	OpContains      Operator = "contains"   // 包含
	OpNotContains   Operator = "not_contains" // 不包含
	OpIn           Operator = "in"         // 在列表中
	OpNotIn        Operator = "not_in"     // 不在列表中
	OpRegex        Operator = "regex"      // 正则匹配
	OpBetween      Operator = "between"    // 在范围内
	OpIsNull       Operator = "is_null"    // 为空
	OpIsNotNull    Operator = "is_not_null" // 不为空
)

// RuleCondition 规则条件
type RuleCondition struct {
	Field    string      `json:"field"`     // 字段名
	Operator Operator    `json:"operator"`  // 操作符
	Value    interface{} `json:"value"`     // 比较值
	Weight   float64     `json:"weight"`    // 权重 (0-1)
}

// Rule 规则定义
type Rule struct {
	ID          string           `json:"id"`           // 规则ID
	Name        string           `json:"name"`         // 规则名称
	Description string           `json:"description"`  // 规则描述
	Priority    int              `json:"priority"`     // 优先级 (数字越大优先级越高)
	Conditions  []RuleCondition  `json:"conditions"`   // 规则条件
	Action      RuleAction       `json:"action"`       // 规则动作
	Parameters  map[string]interface{} `json:"parameters"`  // 动作参数
	Status      RuleStatus       `json:"status"`       // 规则状态
	Tags        []string         `json:"tags"`         // 标签
	CreatedAt   time.Time        `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time        `json:"updated_at"`   // 更新时间
	Version     int              `json:"version"`      // 版本号

	// 运行时属性
	HitCount    int64            `json:"hit_count"`    // 命中次数
	LastHitAt   *time.Time       `json:"last_hit_at"`  // 最后命中时间
	AvgExecTime float64          `json:"avg_exec_time"` // 平均执行时间
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	Context    context.Context                `json:"-"`
	Data       map[string]interface{}        `json:"data"`         // 输入数据
	Variables  map[string]interface{}        `json:"variables"`    // 变量
	Metadata   map[string]interface{}        `json:"metadata"`     // 元数据
	StartTime  time.Time                      `json:"start_time"`   // 开始时间
	RuleResults []*RuleResult                 `json:"rule_results"` // 规则执行结果
	TraceEnabled bool                         `json:"trace_enabled"` // 是否启用跟踪
}

// RuleResult 规则执行结果
type RuleResult struct {
	RuleID      string        `json:"rule_id"`      // 规则ID
	RuleName    string        `json:"rule_name"`    // 规则名称
	Matched     bool          `json:"matched"`      // 是否匹配
	Score       float64       `json:"score"`        // 匹配分数
	Action      RuleAction    `json:"action"`      // 执行动作
	Result      interface{}   `json:"result"`       // 执行结果
	Error       error         `json:"error"`        // 错误信息
	ExecTime    time.Duration `json:"exec_time"`    // 执行时间
	Conditions  []bool        `json:"conditions"`   // 各条件匹配结果
	Timestamp   time.Time     `json:"timestamp"`    // 执行时间
}

// RuleEngineConfig 规则引擎配置
type RuleEngineConfig struct {
	MaxConcurrentRules    int           `yaml:"max_concurrent_rules"`
	DefaultTimeout        time.Duration `yaml:"default_timeout"`
	EnableMetrics         bool          `yaml:"enable_metrics"`
	EnableTracing         bool          `yaml:"enable_tracing"`
	CacheSize             int           `yaml:"cache_size"`
	CacheTTL              time.Duration `yaml:"cache_ttl"`
	HotReloadEnabled      bool          `yaml:"hot_reload_enabled"`
	RuleFiles             []string      `yaml:"rule_files"`
}

// DefaultRuleEngineConfig 默认配置
func DefaultRuleEngineConfig() *RuleEngineConfig {
	return &RuleEngineConfig{
		MaxConcurrentRules: 100,
		DefaultTimeout:     30 * time.Second,
		EnableMetrics:      true,
		EnableTracing:      true,
		CacheSize:          1000,
		CacheTTL:           1 * time.Hour,
		HotReloadEnabled:   true,
		RuleFiles:          []string{"rules/*.json"},
	}
}

// RuleEngine 规则引擎接口
type RuleEngine interface {
	// ExecuteRules 执行规则
	ExecuteRules(ctx context.Context, data map[string]interface{}) (*ExecutionContext, error)

	// AddRule 添加规则
	AddRule(rule *Rule) error

	// RemoveRule 删除规则
	RemoveRule(ruleID string) error

	// UpdateRule 更新规则
	UpdateRule(rule *Rule) error

	// GetRule 获取规则
	GetRule(ruleID string) (*Rule, error)

	// ListRules 列出规则
	ListRules(filter *RuleFilter) ([]*Rule, error)

	// EnableRule 启用规则
	EnableRule(ruleID string) error

	// DisableRule 禁用规则
	DisableRule(ruleID string) error

	// ValidateRule 验证规则
	ValidateRule(rule *Rule) error

	// GetStats 获取统计信息
	GetStats(ctx context.Context) (*RuleEngineStats, error)

	// ReloadRules 重新加载规则
	ReloadRules(ctx context.Context) error
}

// RuleFilter 规则过滤器
type RuleFilter struct {
	Status   RuleStatus `json:"status,omitempty"`
	Tags     []string   `json:"tags,omitempty"`
	Action   RuleAction `json:"action,omitempty"`
	Priority *int       `json:"priority,omitempty"`
	Limit    int        `json:"limit,omitempty"`
	Offset   int        `json:"offset,omitempty"`
}

// RuleEngineStats 规则引擎统计信息
type RuleEngineStats struct {
	TotalRules       int64     `json:"total_rules"`        // 总规则数
	ActiveRules      int64     `json:"active_rules"`       // 激活规则数
	TotalExecutions  int64     `json:"total_executions"`   // 总执行次数
	SuccessfulExecs  int64     `json:"successful_execs"`   // 成功执行次数
	FailedExecs      int64     `json:"failed_execs"`       // 失败执行次数
	AvgExecTime      float64   `json:"avg_exec_time_ms"`   // 平均执行时间
	CacheHitRate     float64   `json:"cache_hit_rate"`     // 缓存命中率
	MostHitRules     []string  `json:"most_hit_rules"`     // 最常命中的规则
	LastExecutions   []string  `json:"last_executions"`    // 最近执行的规则
	LastResetTime    time.Time `json:"last_reset_time"`    // 最后重置时间
}

// ActionHandler 动作处理器接口
type ActionHandler interface {
	CanHandle(action RuleAction) bool
	Execute(ctx context.Context, action RuleAction, parameters map[string]interface{}, data map[string]interface{}) (interface{}, error)
}

// RuleRepository 规则存储接口
type RuleRepository interface {
	Save(ctx context.Context, rule *Rule) error
	Get(ctx context.Context, ruleID string) (*Rule, error)
	List(ctx context.Context, filter *RuleFilter) ([]*Rule, error)
	Delete(ctx context.Context, ruleID string) error
	Exists(ctx context.Context, ruleID string) (bool, error)
}