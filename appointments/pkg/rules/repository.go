package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// MemoryRuleRepository 内存规则存储实现
type MemoryRuleRepository struct {
	rules map[string]*Rule
	mu    sync.RWMutex
}

// NewMemoryRuleRepository 创建内存规则存储
func NewMemoryRuleRepository() *MemoryRuleRepository {
	return &MemoryRuleRepository{
		rules: make(map[string]*Rule),
	}
}

// Save 保存规则
func (r *MemoryRuleRepository) Save(ctx context.Context, rule *Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ruleCopy := *rule
	r.rules[rule.ID] = &ruleCopy
	return nil
}

// Get 获取规则
func (r *MemoryRuleRepository) Get(ctx context.Context, ruleID string) (*Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, exists := r.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("规则不存在: %s", ruleID)
	}

	ruleCopy := *rule
	return &ruleCopy, nil
}

// List 列出规则
func (r *MemoryRuleRepository) List(ctx context.Context, filter *RuleFilter) ([]*Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var rules []*Rule
	for _, rule := range r.rules {
		if matchesFilter(rule, filter) {
			ruleCopy := *rule
			rules = append(rules, &ruleCopy)
		}
	}

	// 应用分页
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		if start >= len(rules) {
			return nil, nil
		}
		end := start + filter.Limit
		if end > len(rules) {
			end = len(rules)
		}
		rules = rules[start:end]
	}

	return rules, nil
}

// Delete 删除规则
func (r *MemoryRuleRepository) Delete(ctx context.Context, ruleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rules[ruleID]; !exists {
		return fmt.Errorf("规则不存在: %s", ruleID)
	}

	delete(r.rules, ruleID)
	return nil
}

// Exists 检查规则是否存在
func (r *MemoryRuleRepository) Exists(ctx context.Context, ruleID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.rules[ruleID]
	return exists, nil
}

// FileRuleRepository 文件规则存储实现
type FileRuleRepository struct {
	ruleDir string
	logger  *logger.Logger
	mu      sync.RWMutex
}

// NewFileRuleRepository 创建文件规则存储
func NewFileRuleRepository(ruleDir string, logger *logger.Logger) *FileRuleRepository {
	return &FileRuleRepository{
		ruleDir: ruleDir,
		logger:  logger,
	}
}

// Save 保存规则
func (r *FileRuleRepository) Save(ctx context.Context, rule *Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 确保目录存在
	if err := os.MkdirAll(r.ruleDir, 0755); err != nil {
		return fmt.Errorf("��建规则目录失败: %w", err)
	}

	// 序列化规则
	data, err := json.MarshalIndent(rule, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化规则失败: %w", err)
	}

	// 写入文件
	filename := filepath.Join(r.ruleDir, rule.ID+".json")
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入规则文件失败: %w", err)
	}

	r.logger.Debug("规则保存到文件成功",
		map[string]interface{}{
			"rule_id":   rule.ID,
			"file_path": filename,
		})

	return nil
}

// Get 获取规则
func (r *FileRuleRepository) Get(ctx context.Context, ruleID string) (*Rule, error) {
	filename := filepath.Join(r.ruleDir, ruleID+".json")

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("规则不存在: %s", ruleID)
		}
		return nil, fmt.Errorf("读取规则文件失败: %w", err)
	}

	var rule Rule
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, fmt.Errorf("反序列化规则失败: %w", err)
	}

	return &rule, nil
}

// List 列出规则
func (r *FileRuleRepository) List(ctx context.Context, filter *RuleFilter) ([]*Rule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 读取目录中的所有JSON文件
	files, err := ioutil.ReadDir(r.ruleDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Rule{}, nil
		}
		return nil, fmt.Errorf("读取规则目录失败: %w", err)
	}

	var rules []*Rule
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filename := filepath.Join(r.ruleDir, file.Name())
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			r.logger.Warn("读取规则文件失败",
				map[string]interface{}{
					"file_path": filename,
					"error":     err,
				})
			continue
		}

		var rule Rule
		if err := json.Unmarshal(data, &rule); err != nil {
			r.logger.Warn("反序列化规则失败",
				map[string]interface{}{
					"file_path": filename,
					"error":     err,
				})
			continue
		}

		if matchesFilter(&rule, filter) {
			rules = append(rules, &rule)
		}
	}

	// 按优先级排序
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	// 应用分页
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		if start >= len(rules) {
			return nil, nil
		}
		end := start + filter.Limit
		if end > len(rules) {
			end = len(rules)
		}
		rules = rules[start:end]
	}

	return rules, nil
}

// Delete 删除规则
func (r *FileRuleRepository) Delete(ctx context.Context, ruleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	filename := filepath.Join(r.ruleDir, ruleID+".json")
	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("规则不存在: %s", ruleID)
		}
		return fmt.Errorf("删除规则文件失败: %w", err)
	}

	r.logger.Debug("规则文件删除成功",
		map[string]interface{}{
			"rule_id":   ruleID,
			"file_path": filename,
		})

	return nil
}

// Exists 检查规则是否存在
func (r *FileRuleRepository) Exists(ctx context.Context, ruleID string) (bool, error) {
	filename := filepath.Join(r.ruleDir, ruleID+".json")
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// LoadRulesFromDirectory 从目录加载规则文件
func (r *FileRuleRepository) LoadRulesFromDirectory(patterns []string) ([]*Rule, error) {
	var allRules []*Rule

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			r.logger.Warn("无效的文件模式",
				map[string]interface{}{
					"pattern": pattern,
					"error":   err,
				})
			continue
		}

		for _, file := range matches {
			rules, err := r.loadRulesFromFile(file)
			if err != nil {
				r.logger.Warn("加载规则文件失败",
					map[string]interface{}{
						"file_path": file,
						"error":     err,
					})
				continue
			}
			allRules = append(allRules, rules...)
		}
	}

	// 按优先级排序
	sort.Slice(allRules, func(i, j int) bool {
		return allRules[i].Priority > allRules[j].Priority
	})

	return allRules, nil
}

// loadRulesFromFile 从文件加载规则
func (r *FileRuleRepository) loadRulesFromFile(filename string) ([]*Rule, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 尝试解析为单个规则
	var singleRule Rule
	if err := json.Unmarshal(data, &singleRule); err == nil && singleRule.ID != "" {
		singleRule.CreatedAt = time.Now()
		singleRule.UpdatedAt = time.Now()
		return []*Rule{&singleRule}, nil
	}

	// 尝试解析为规则数组
	var rules []*Rule
	if err := json.Unmarshal(data, &rules); err == nil {
		for _, rule := range rules {
			if rule.CreatedAt.IsZero() {
				rule.CreatedAt = time.Now()
			}
			if rule.UpdatedAt.IsZero() {
				rule.UpdatedAt = time.Now()
			}
		}
		return rules, nil
	}

	return nil, fmt.Errorf("文件格式无效，无法解析为规则")
}

// matchesFilter 检查规则是否匹配过滤器
func matchesFilter(rule *Rule, filter *RuleFilter) bool {
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