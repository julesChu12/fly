package config

import (
	"fmt"
	"time"

	"github.com/julesChu12/fly/mora/pkg/config"
	"github.com/spf13/viper"
)

// Manager 配置管理器
type Manager struct {
	viper *viper.Viper
}

// NewManager 创建配置管理器
func NewManager(configPaths []string, envPrefix string) (*Manager, error) {
	loader := config.New().
		WithYAML(configPaths...).
		WithEnvPrefix(envPrefix)

	v, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	return &Manager{
		viper: v,
	}, nil
}

// GetString 获取字符串配置
func (m *Manager) GetString(key string) string {
	return m.viper.GetString(key)
}

// GetInt 获取整数配置
func (m *Manager) GetInt(key string) int {
	return m.viper.GetInt(key)
}

// GetBool 获取布尔配置
func (m *Manager) GetBool(key string) bool {
	return m.viper.GetBool(key)
}

// GetDuration 获取时间间隔配置
func (m *Manager) GetDuration(key string) time.Duration {
	return m.viper.GetDuration(key)
}

// GetBytes 获取字节数配置
func (m *Manager) GetBytes(key string) uint64 {
	return m.viper.GetUint64(key)
}

// GetStringSlice 获取字符串数组配置
func (m *Manager) GetStringSlice(key string) []string {
	return m.viper.GetStringSlice(key)
}

// Set 设置配置值
func (m *Manager) Set(key string, value interface{}) {
	m.viper.Set(key, value)
}

// IsSet 检查配置是否存在
func (m *Manager) IsSet(key string) bool {
	return m.viper.IsSet(key)
}

// Unmarshal 解析配置到结构体
func (m *Manager) Unmarshal(key string, rawVal interface{}) error {
	return m.viper.UnmarshalKey(key, rawVal)
}

// AllSettings 获取所有配置
func (m *Manager) AllSettings() map[string]interface{} {
	return m.viper.AllSettings()
}