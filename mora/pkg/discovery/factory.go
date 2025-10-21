package discovery

import (
	"fmt"
	"strings"
)

// DiscoveryType 服务发现类型
type DiscoveryType string

const (
	TypeEnv    DiscoveryType = "env"    // 环境变量模式
	TypeConsul DiscoveryType = "consul" // Consul 模式
)

// Config 服务发现配置
type Config struct {
	Type   DiscoveryType  `yaml:"type" json:"type"`     // 服务发现类型
	Consul *ConsulConfig  `yaml:"consul" json:"consul"` // Consul 配置
}

// New 根据配置创建服务发现实例
func New(cfg *Config) (Discovery, error) {
	if cfg == nil {
		// 默认使用环境变量模式
		return NewEnvDiscovery(), nil
	}

	switch strings.ToLower(string(cfg.Type)) {
	case string(TypeEnv), "":
		return NewEnvDiscovery(), nil
	case string(TypeConsul):
		return NewConsulDiscovery(cfg.Consul)
	default:
		return nil, fmt.Errorf("unsupported discovery type: %s", cfg.Type)
	}
}

// MustNew 创建服务发现实例，失败时 panic
func MustNew(cfg *Config) Discovery {
	discovery, err := New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create discovery: %v", err))
	}
	return discovery
}
