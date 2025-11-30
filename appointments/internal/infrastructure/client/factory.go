package client

import (
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// ClientFactory 客户端工厂
type ClientFactory struct {
	logger *logger.Logger
}

// NewClientFactory 创建客户端工厂
func NewClientFactory(logger *logger.Logger) *ClientFactory {
	return &ClientFactory{
		logger: logger,
	}
}

// CreateKratosClient 创建Kratos客户端
func (f *ClientFactory) CreateKratosClient(config *KratosClientConfig) *KratosClient {
	if config == nil {
		config = DefaultKratosClientConfig()
	}

	return NewKratosClient(config, f.logger)
}

// CreatePlutusClient 创建Plutus客户端
func (f *ClientFactory) CreatePlutusClient(config *PlutusClientConfig) *PlutusClient {
	if config == nil {
		config = DefaultPlutusClientConfig()
	}

	return NewPlutusClient(config, f.logger)
}

// CreateAllClients 创建所有客户端
func (f *ClientFactory) CreateAllClients(kratosConfig *KratosClientConfig, plutusConfig *PlutusClientConfig) (*KratosClient, *PlutusClient) {
	kratosClient := f.CreateKratosClient(kratosConfig)
	plutusClient := f.CreatePlutusClient(plutusConfig)

	return kratosClient, plutusClient
}

// PlutusClientConfig Plutus客户端配置
type PlutusClientConfig struct {
	BaseURL        string        `yaml:"base_url"`
	APIKey         string        `yaml:"api_key"`
	APISecret      string        `yaml:"api_secret"`
	Timeout        time.Duration `yaml:"timeout"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
	EnableCircuit  bool          `yaml:"enable_circuit"`
	CircuitTimeout time.Duration `yaml:"circuit_timeout"`
}

// DefaultPlutusClientConfig 默认Plutus配置
func DefaultPlutusClientConfig() *PlutusClientConfig {
	return &PlutusClientConfig{
		BaseURL:        "http://localhost:8085", // Plutus服务地址
		APIKey:         "test_api_key",          // 测试环境API密钥
		APISecret:      "test_api_secret",       // 测试环境API密钥
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		EnableCircuit:  true,
		CircuitTimeout: 60 * time.Second,
	}
}
