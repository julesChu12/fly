package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"github.com/julesChu12/fly/mora/pkg/db"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// Config 数据库配置
type Config struct {
	Driver          string        `yaml:"driver" json:"driver"`
	DSN             string        `yaml:"dsn" json:"dsn"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	LogLevel        string        `yaml:"log_level" json:"log_level"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Driver:          "mysql",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        "warn",
	}
}

// Manager 数据库管理器
type Manager struct {
	gormDB *gorm.DB
	logger *logger.Logger
	config *Config
}

// NewManager 创建数据库管理器
func NewManager(config *Config, logger *logger.Logger) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 使用 mora 的数据库包装器
	gormConfig := db.Config{
		DSN:             config.DSN,
		MaxIdleConns:    config.MaxIdleConns,
		MaxOpenConns:    config.MaxOpenConns,
		ConnMaxLifetime: int(config.ConnMaxLifetime.Seconds()),
	}

	gormClient, err := db.New(gormConfig)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败: %w", err)
	}

	gormDB := gormClient.DB()

	return &Manager{
		gormDB: gormDB,
		logger: logger,
		config: config,
	}, nil
}

// GetDB 获取 GORM 数据库实例
func (m *Manager) GetDB() *gorm.DB {
	return m.gormDB
}

// HealthCheck 健康检查
func (m *Manager) HealthCheck() error {
	sqlDB, err := m.gormDB.DB()
	if err != nil {
		return fmt.Errorf("获取底层 SQL 连接失败: %w", err)
	}

	return sqlDB.Ping()
}

// Close 关闭数据库连接
func (m *Manager) Close() error {
	sqlDB, err := m.gormDB.DB()
	if err != nil {
		return fmt.Errorf("获取底层 SQL 连接失败: %w", err)
	}

	return sqlDB.Close()
}

// GetStats 获取连接池统计信息
func (m *Manager) GetStats() map[string]interface{} {
	sqlDB, err := m.gormDB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

