package integration

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// TestContainer 测试容器配置
type TestContainer struct {
	MySQLContainer *mysql.MySQLContainer
	RedisContainer  *redis.RedisContainer
	MySQLHost      string
	MySQLPort      string
	RedisHost      string
	RedisPort      string
	Database        *gorm.DB
}

// SetupTestContainers 设置测试容器
func SetupTestContainers(ctx context.Context) (*TestContainer, error) {
	log.Println("🚀 启动集成测试容器...")

	// 启动MySQL容器
	mysqlContainer, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0"),
		mysql.WithDatabase("items_test"),
		mysql.WithUsername("testuser"),
		mysql.WithPassword("testpass"),
		mysql.WithScripts("./migrations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start MySQL container: %w", err)
	}

	mysqlHost, err := mysqlContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get MySQL host: %w", err)
	}
	mysqlPort, err := mysqlContainer.MappedPort(ctx, "3306/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get MySQL port: %w", err)
	}

	// 启动Redis容器
	redisContainer, err := redis.RunContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start Redis container: %w", err)
	}

	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis host: %w", err)
	}
	redisPort, err := redisContainer.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis port: %w", err)
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"testuser", "testpass", mysqlHost, mysqlPort.Port(), "items_test")

	db, err := gorm.Open(mysqlDriver.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 等待数据库就绪
	if err := waitForDatabase(db); err != nil {
		return nil, fmt.Errorf("database not ready: %w", err)
	}

	log.Printf("✅ 测试容器启动成功:")
	log.Printf("   MySQL: %s:%s", mysqlHost, mysqlPort.Port())
	log.Printf("   Redis: %s:%s", redisHost, redisPort.Port())

	return &TestContainer{
		MySQLContainer: mysqlContainer,
		RedisContainer:  redisContainer,
		MySQLHost:      mysqlHost,
		MySQLPort:      mysqlPort.Port(),
		RedisHost:      redisHost,
		RedisPort:      redisPort.Port(),
		Database:        db,
	}, nil
}

// Cleanup 清理测试容器
func (tc *TestContainer) Cleanup(ctx context.Context) error {
	log.Println("🧹 清理测试容器...")

	var errors []error

	if tc.Database != nil {
		if sqlDB, err := tc.Database.DB(); err == nil {
			sqlDB.Close()
		}
	}

	if tc.MySQLContainer != nil {
		if err := tc.MySQLContainer.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate MySQL container: %w", err))
		}
	}

	if tc.RedisContainer != nil {
		if err := tc.RedisContainer.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate Redis container: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	log.Println("✅ 测试容器清理完成")
	return nil
}

// waitForDatabase 等待数据库就绪
func waitForDatabase(db *gorm.DB) error {
	for i := 0; i < 30; i++ {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		if err := sqlDB.Ping(); err == nil {
			log.Println("✅ 数据库连接就绪")
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("database not ready after 30 seconds")
}

// ResetDatabase 重置数据库
func (tc *TestContainer) ResetDatabase(db *gorm.DB) error {
	log.Println("🔄 重置测试数据库...")

	// 获取所有表名
	var tables []string
	err := db.Raw("SHOW TABLES").Scan(&tables).Error
	if err != nil {
		return err
	}

	// 删除所有表数据（保留表结构）
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	log.Println("✅ 数据库重置完成")
	return nil
}

// GetMySQLDSN 获取MySQL连接字符串
func (tc *TestContainer) GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"testuser", "testpass", tc.MySQLHost, tc.MySQLPort, "items_test")
}

// GetRedisAddress 获取Redis连接地址
func (tc *TestContainer) GetRedisAddress() string {
	return fmt.Sprintf("%s:%s", tc.RedisHost, tc.RedisPort)
}