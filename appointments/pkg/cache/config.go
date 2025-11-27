package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	Redis  RedisConfig  `yaml:"redis"`
	Memory MemoryConfig `yaml:"memory"`
	TTL    TTLConfig    `yaml:"ttl"`
	Cleanup CleanupConfig `yaml:"cleanup"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Address     string        `yaml:"address"`
	Password    string        `yaml:"password"`
	DB          int           `yaml:"db"`
	PoolSize    int           `yaml:"pool_size"`
	MaxRetries  int           `yaml:"max_retries"`
	Timeout     time.Duration `yaml:"timeout"`
	KeyPrefix   string        `yaml:"key_prefix"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

// MemoryConfig 内存缓存配置
type MemoryConfig struct {
	Enabled    bool          `yaml:"enabled"`
	MaxSize    int           `yaml:"max_size"`
	DefaultTTL time.Duration `yaml:"default_ttl"`
	GCInterval time.Duration `yaml:"gc_interval"`
}

// TTLConfig TTL配置
type TTLConfig struct {
	// 预约相关缓存
	AppointmentSingle time.Duration `yaml:"appointment_single"` // 预约详情缓存时间
	AppointmentList  time.Duration `yaml:"appointment_list"`   // 预约列表缓存时间

	// 实时性缓存
	AvailabilityCheck time.Duration `yaml:"availability_check"` // 可用性检查缓存时间
	ConflictCheck    time.Duration `yaml:"conflict_check"`     // 冲突检查缓存时间

	// 规则引擎缓存
	RulesValidation time.Duration `yaml:"rules_validation"` // 规则验证缓存时间
	RulesPricing    time.Duration `yaml:"rules_pricing"`     // 定价规则缓存时间

	// 外部服务缓存
	ServiceData time.Duration `yaml:"service_data"` // 服务数据缓存时间
	ConfigData  time.Duration `yaml:"config_data"`  // 配置数据缓存时间

	// 默认TTL
	Default time.Duration `yaml:"default"`
}

// CleanupConfig 清理配置
type CleanupConfig struct {
	Interval   time.Duration `yaml:"interval"`
	Retention   time.Duration `yaml:"retention"`
	BatchSize   int           `yaml:"batch_size"`
	Enabled     bool          `yaml:"enabled"`
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Redis: RedisConfig{
			Address:     "localhost:6379",
			Password:    "",
			DB:          2, // 使用DB 2作为业务缓存（DB 1用于幂等性）
			PoolSize:    10,
			MaxRetries:  3,
			Timeout:     5 * time.Second,
			KeyPrefix:   "appointments",
			IdleTimeout: 5 * time.Minute,
		},
		Memory: MemoryConfig{
			Enabled:    true,
			MaxSize:    1000,
			DefaultTTL: 10 * time.Minute,
			GCInterval: 5 * time.Minute,
		},
		TTL: TTLConfig{
			// 预约相关缓存（中等TTL，平衡性能和一致性）
			AppointmentSingle: 15 * time.Minute,
			AppointmentList:  10 * time.Minute,

			// 实时性缓存（短TTL，保证实时性）
			AvailabilityCheck: 2 * time.Minute,
			ConflictCheck:    5 * time.Minute,

			// 规则引擎缓存（较长TTL，规则变更不频繁）
			RulesValidation: 30 * time.Minute,
			RulesPricing:    1 * time.Hour,

			// 外部服务缓存（长TTL，外部数据变更不频繁）
			ServiceData: 2 * time.Hour,
			ConfigData:  24 * time.Hour,

			// 默认TTL
			Default: 10 * time.Minute,
		},
		Cleanup: CleanupConfig{
			Interval:   10 * time.Minute,
			Retention:   30 * 24 * time.Hour, // 30天
			BatchSize:   100,
			Enabled:     true,
		},
	}
}

// CacheKeyBuilder 缓存键构建器
type CacheKeyBuilder struct {
	prefix string
	parts  []string
}

// NewCacheKeyBuilder 创建缓存键构建器
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{
		prefix: prefix,
		parts:  make([]string, 0),
	}
}

// Add 添加部分
func (b *CacheKeyBuilder) Add(part string) *CacheKeyBuilder {
	b.parts = append(b.parts, part)
	return b
}

// AddHash 添加哈希部分
func (b *CacheKeyBuilder) AddHash(data string) *CacheKeyBuilder {
	hash := generateHash(data)
	b.parts = append(b.parts, hash)
	return b
}

// Build 构建缓存键
func (b *CacheKeyBuilder) Build() string {
	if b.prefix != "" {
		return b.prefix + ":" + strings.Join(b.parts, ":")
	}
	return strings.Join(b.parts, ":")
}

// BuildWithPrefix 使用指定前缀构建键
func (b *CacheKeyBuilder) BuildWithPrefix(prefix string) string {
	return prefix + ":" + b.Build()
}

// Reset 重置构建器
func (b *CacheKeyBuilder) Reset() *CacheKeyBuilder {
	b.parts = b.parts[:0]
	return b
}

// FilterHashBuilder 过滤条件哈希构建器
type FilterHashBuilder struct {
	pairs map[string]interface{}
}

// NewFilterHashBuilder 创建过滤条件哈希构建器
func NewFilterHashBuilder() *FilterHashBuilder {
	return &FilterHashBuilder{
		pairs: make(map[string]interface{}),
	}
}

// Add 添加过滤条件
func (b *FilterHashBuilder) Add(key string, value interface{}) *FilterHashBuilder {
	b.pairs[key] = value
	return b
}

// AddMap 批量添加过滤条件
func (b *FilterHashBuilder) AddMap(pairs map[string]interface{}) *FilterHashBuilder {
	for k, v := range pairs {
		b.pairs[k] = v
	}
	return b
}

// Build 构建哈希值
func (b *FilterHashBuilder) Build() string {
	data := fmt.Sprintf("%v", b.pairs)
	return generateHash(data)
}

// Reset 重置构建器
func (b *FilterHashBuilder) Reset() *FilterHashBuilder {
	b.pairs = make(map[string]interface{})
	return b
}

// generateHash 生成SHA256哈希
func generateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8]) // 使用前8位作为短哈希
}

// CacheKeyTemplate 缓存键模板
type CacheKeyTemplate struct {
	template string
}

// NewCacheKeyTemplate 创建缓存键模板
func NewCacheKeyTemplate(template string) *CacheKeyTemplate {
	return &CacheKeyTemplate{
		template: template,
	}
}

// Format 格式化缓存键
func (t *CacheKeyTemplate) Format(args ...interface{}) string {
	return fmt.Sprintf(t.template, args...)
}

// 预定义的缓存键模板
var (
	// 预约相关
	AppointmentSingleKey = NewCacheKeyTemplate("appointments:single:%s")
	AppointmentListKey  = NewCacheKeyTemplate("appointments:list:%s")
	AvailabilityKey     = NewCacheKeyTemplate("appointments:availability:%s:%s")
	ConflictKey        = NewCacheKeyTemplate("appointments:conflict:%s:%s:%s")

	// 规则引擎
	RulesValidationKey = NewCacheKeyTemplate("rules:validation:%s")
	RulesPricingKey    = NewCacheKeyTemplate("rules:pricing:%s")

	// 外部服务
	ServicePriceKey = NewCacheKeyTemplate("service:price:%s")
	ServiceStaffKey = NewCacheKeyTemplate("service:staff:%s")

	// 配置数据
	ConfigEnumKey = NewCacheKeyTemplate("config:enum:%s")
)

// CacheKeyGenerator 缓存键生成器
type CacheKeyGenerator struct {
	prefix string
}

// NewCacheKeyGenerator 创建缓存键生成器
func NewCacheKeyGenerator(prefix string) *CacheKeyGenerator {
	return &CacheKeyGenerator{
		prefix: prefix,
	}
}

// GenerateAppointmentKey 生成预约缓存键
func (g *CacheKeyGenerator) GenerateAppointmentKey(id string) string {
	return fmt.Sprintf("%s:appointments:single:%s", g.prefix, id)
}

// GenerateAppointmentListKey 生成预约列表缓存键
func (g *CacheKeyGenerator) GenerateAppointmentListKey(filterHash string) string {
	return fmt.Sprintf("%s:appointments:list:%s", g.prefix, filterHash)
}

// GenerateAvailabilityKey 生成可用性缓存键
func (g *CacheKeyGenerator) GenerateAvailabilityKey(staffID, date string) string {
	return fmt.Sprintf("%s:appointments:availability:%s:%s", g.prefix, staffID, date)
}

// GenerateConflictKey 生成冲突检查缓存键
func (g *CacheKeyGenerator) GenerateConflictKey(staffID string, startTime, endTime time.Time) string {
	start := startTime.Format("20060102150405")
	end := endTime.Format("20060102150405")
	return fmt.Sprintf("%s:appointments:conflict:%s:%s:%s", g.prefix, staffID, start, end)
}

// GenerateRulesValidationKey 生成规则验证缓存键
func (g *CacheKeyGenerator) GenerateRulesValidationKey(dataHash string) string {
	return fmt.Sprintf("%s:rules:validation:%s", g.prefix, dataHash)
}

// GenerateRulesPricingKey 生成定价规则缓存键
func (g *CacheKeyGenerator) GenerateRulesPricingKey(appointmentID string) string {
	return fmt.Sprintf("%s:rules:pricing:%s", g.prefix, appointmentID)
}

// GenerateServicePriceKey 生成服务价格缓存键
func (g *CacheKeyGenerator) GenerateServicePriceKey(serviceID string) string {
	return fmt.Sprintf("%s:service:price:%s", g.prefix, serviceID)
}

// GenerateServiceStaffKey 生成员工信息缓存键
func (g *CacheKeyGenerator) GenerateServiceStaffKey(staffID string) string {
	return fmt.Sprintf("%s:service:staff:%s", g.prefix, staffID)
}

// GenerateConfigEnumKey 生成配置枚举缓存键
func (g *CacheKeyGenerator) GenerateConfigEnumKey(enumType string) string {
	return fmt.Sprintf("%s:config:enum:%s", g.prefix, enumType)
}

// ValidateConfig 验证配置
func (c *CacheConfig) Validate() error {
	// 验证Redis配置
	if c.Redis.Address == "" {
		return fmt.Errorf("Redis地址不能为空")
	}
	if c.Redis.DB < 0 || c.Redis.DB > 15 {
		return fmt.Errorf("Redis数据库编号必须在0-15之间")
	}
	if c.Redis.PoolSize <= 0 {
		return fmt.Errorf("Redis连接池大小必须大于0")
	}

	// 验证内存配置
	if c.Memory.Enabled && c.Memory.MaxSize <= 0 {
		return fmt.Errorf("启用内存缓存时，最大大小必须大于0")
	}

	// 验证TTL配置
	if c.TTL.AppointmentSingle <= 0 {
		return fmt.Errorf("预约详情缓存TTL必须大于0")
	}
	if c.TTL.AppointmentList <= 0 {
		return fmt.Errorf("预约列表缓存TTL必须大于0")
	}
	if c.TTL.AvailabilityCheck <= 0 {
		return fmt.Errorf("可用性检查缓存TTL必须大于0")
	}

	// 验证清理配置
	if c.Cleanup.Enabled && c.Cleanup.Interval <= 0 {
		return fmt.Errorf("启用清理时，清理间隔必须大于0")
	}

	return nil
}

// GetTTLByType 根据类型获取TTL
func (c *CacheConfig) GetTTLByType(cacheType string) time.Duration {
	switch cacheType {
	case "appointment_single":
		return c.TTL.AppointmentSingle
	case "appointment_list":
		return c.TTL.AppointmentList
	case "availability_check":
		return c.TTL.AvailabilityCheck
	case "conflict_check":
		return c.TTL.ConflictCheck
	case "rules_validation":
		return c.TTL.RulesValidation
	case "rules_pricing":
		return c.TTL.RulesPricing
	case "service_data":
		return c.TTL.ServiceData
	case "config_data":
		return c.TTL.ConfigData
	default:
		return c.TTL.Default
	}
}

// IsMemoryEnabled 检查内存缓存是否启用
func (c *CacheConfig) IsMemoryEnabled() bool {
	return c.Memory.Enabled && c.Memory.MaxSize > 0
}

// IsRedisEnabled 检查Redis缓存是否启用
func (c *CacheConfig) IsRedisEnabled() bool {
	return c.Redis.Address != "" && c.Redis.PoolSize > 0
}