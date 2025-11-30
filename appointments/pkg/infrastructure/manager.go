package infrastructure

import (
	"fmt"

	"github.com/julesChu12/fly/appointments/pkg/cache"
	"github.com/julesChu12/fly/appointments/pkg/config"
	"github.com/julesChu12/fly/appointments/pkg/database"
	"github.com/julesChu12/fly/appointments/pkg/events"
	"github.com/julesChu12/fly/appointments/pkg/observability"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/mora/pkg/mq"
)

// Config 基础设施配置
type Config struct {
	// 配置管理
	ConfigPaths []string `yaml:"config_paths" json:"config_paths"`
	EnvPrefix   string   `yaml:"env_prefix" json:"env_prefix"`

	// 缓存配置
	Cache struct {
		RedisAddr     string `yaml:"redis_addr" json:"redis_addr"`
		RedisPassword string `yaml:"redis_password" json:"redis_password"`
		RedisDB       int    `yaml:"redis_db" json:"redis_db"`
		Prefix        string `yaml:"prefix" json:"prefix"`
	} `yaml:"cache" json:"cache"`

	// 数据库配置
	Database *database.Config `yaml:"database" json:"database"`

	// 消息队列配置
	MessageQueue struct {
		Driver  string            `yaml:"driver" json:"driver"`   // memory, redis, kafka
		DSN     string            `yaml:"dsn" json:"dsn"`         // connection string
		Options map[string]string `yaml:"options" json:"options"` // additional options
	} `yaml:"message_queue" json:"message_queue"`

	// 可观测性配置
	Observability *observability.Config `yaml:"observability" json:"observability"`

	// 日志配置
	Logger struct {
		Level  string `yaml:"level" json:"level"`
		Format string `yaml:"format" json:"format"`
		Output string `yaml:"output" json:"output"`
	} `yaml:"logger" json:"logger"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		ConfigPaths: []string{"configs/appointments.yaml"},
		EnvPrefix:   "APPOINTMENTS",
		Database:    database.DefaultConfig(),
		Observability: observability.DefaultConfig(),
		Logger: struct {
			Level  string `yaml:"level" json:"level"`
			Format string `yaml:"format" json:"format"`
			Output string `yaml:"output" json:"output"`
		}{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		MessageQueue: struct {
			Driver  string            `yaml:"driver" json:"driver"`
			DSN     string            `yaml:"dsn" json:"dsn"`
			Options map[string]string `yaml:"options" json:"options"`
		}{
			Driver: "memory",
			Options: make(map[string]string),
		},
		Cache: struct {
			RedisAddr     string `yaml:"redis_addr" json:"redis_addr"`
			RedisPassword string `yaml:"redis_password" json:"redis_password"`
			RedisDB       int    `yaml:"redis_db" json:"redis_db"`
			Prefix        string `yaml:"prefix" json:"prefix"`
		}{
			RedisAddr: "localhost:6379",
			RedisDB:   2,
			Prefix:    "appointments",
		},
	}
}

// Manager 基础设施管理器
type Manager struct {
	config      *Config
	configMgr   *config.Manager
	logger      *logger.Logger
	obsMgr      *observability.Manager
	dbMgr       *database.Manager
	cacheAdapter *cache.Adapter
	eventBus    events.EventBus
	mqClient    mq.Client
}

// NewManager 创建基础设施管理器
func NewManager(cfg *Config) (*Manager, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 1. 初始化配置管理器
	configMgr, err := config.NewManager(cfg.ConfigPaths, cfg.EnvPrefix)
	if err != nil {
		return nil, fmt.Errorf("初始化配置管理器失败: %w", err)
	}

	// 2. 初始化日志
	log, err := logger.New(logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	log.Info("开始初始化基础设施")

	// 3. 初始化可观测性
	obsMgr, err := observability.NewManager(cfg.Observability, log)
	if err != nil {
		log.Error("初始化可观测性失败", "error", err)
		return nil, fmt.Errorf("初始化可观测性失败: %w", err)
	}

	// 4. 初始化数据库
	dbMgr, err := database.NewManager(cfg.Database, log)
	if err != nil {
		log.Error("初始化数据库失败", "error", err)
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 5. 初始化缓存
	cacheAdapter, err := cache.NewAdapter(
		cfg.Cache.RedisAddr,
		cfg.Cache.RedisPassword,
		cfg.Cache.RedisDB,
		cfg.Cache.Prefix,
	)
	if err != nil {
		log.Error("初始化缓存失败", "error", err)
		return nil, fmt.Errorf("初始化缓存失败: %w", err)
	}

	// 6. 初始化消息队列
	mqConfig := mq.Config{
		Driver:  cfg.MessageQueue.Driver,
		DSN:     cfg.MessageQueue.DSN,
		Options: cfg.MessageQueue.Options,
	}

	mqClient, err := mq.New(mqConfig)
	if err != nil {
		log.Error("初始化消息队列失败", "error", err)
		return nil, fmt.Errorf("初始化消息队列失败: %w", err)
	}

	// 7. 初始化事件总线
	eventBus := events.NewAdapter(mqClient, log)

	log.Info("基础设施初始化完成")

	return &Manager{
		config:       cfg,
		configMgr:    configMgr,
		logger:       log,
		obsMgr:       obsMgr,
		dbMgr:        dbMgr,
		cacheAdapter: cacheAdapter,
		eventBus:     eventBus,
		mqClient:     mqClient,
	}, nil
}

// GetConfigManager 获取配置管理器
func (m *Manager) GetConfigManager() *config.Manager {
	return m.configMgr
}

// GetLogger 获取日志器
func (m *Manager) GetLogger() *logger.Logger {
	return m.logger
}

// GetObservabilityManager 获取可观测性管理器
func (m *Manager) GetObservabilityManager() *observability.Manager {
	return m.obsMgr
}

// GetDatabaseManager 获取数据库管理器
func (m *Manager) GetDatabaseManager() *database.Manager {
	return m.dbMgr
}

// GetCacheAdapter 获取缓存适配器
func (m *Manager) GetCacheAdapter() *cache.Adapter {
	return m.cacheAdapter
}

// GetEventBus 获取事件总线
func (m *Manager) GetEventBus() events.EventBus {
	return m.eventBus
}

// GetMQClient 获取消息队列客户端
func (m *Manager) GetMQClient() mq.Client {
	return m.mqClient
}

// HealthCheck 健康检查
func (m *Manager) HealthCheck() map[string]error {
	results := make(map[string]error)

	// 数据库健康检查
	if err := m.dbMgr.HealthCheck(); err != nil {
		results["database"] = err
	} else {
		results["database"] = nil
	}

	return results
}

// Close 关闭所有资源
func (m *Manager) Close() error {
	m.logger.Info("开始关闭基础设施资源")

	var errors []error

	// 关闭事件总线
	if m.eventBus != nil {
		if err := m.eventBus.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭事件总线失败: %w", err))
		}
	}

	// 关闭消息队列
	if m.mqClient != nil {
		if err := m.mqClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭消息队列失败: %w", err))
		}
	}

	// 关闭缓存
	if m.cacheAdapter != nil {
		if err := m.cacheAdapter.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭缓存失败: %w", err))
		}
	}

	// 关闭数据库
	if m.dbMgr != nil {
		if err := m.dbMgr.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭数据库失败: %w", err))
		}
	}

	// 关闭可观测性
	if m.obsMgr != nil {
		if err := m.obsMgr.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭可观测性失败: %w", err))
		}
	}

	if len(errors) > 0 {
		m.logger.Error("部分资源关闭失败", "errors", errors)
		return fmt.Errorf("关闭资源时发生错误")
	}

	m.logger.Info("基础设施资源关闭完成")
	return nil
}

// GetStats 获取基础设施统计信息
func (m *Manager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 数据库统计
	if m.dbMgr != nil {
		stats["database"] = m.dbMgr.GetStats()
	}

	// 事件总线统计
	if m.eventBus != nil {
		stats["event_bus"] = m.eventBus.GetStats()
	}

	return stats
}