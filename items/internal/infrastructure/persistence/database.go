package persistence

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/julesChu12/fly/items/internal/infrastructure/persistence/model"
)

// DB 数据库连接
type DB struct {
	*gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg *viper.Viper) (*DB, error) {
	dsn := buildDSN(cfg)

	// GORM 配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.GetInt("database.max_idle_conns"))
	sqlDB.SetMaxOpenConns(cfg.GetInt("database.max_open_conns"))
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.GetInt("database.conn_max_lifetime")) * time.Second)

	log.Println("Database connected successfully")

	return &DB{DB: db}, nil
}

// buildDSN 构建数据库连接字符串
func buildDSN(cfg *viper.Viper) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=%t&loc=%s",
		cfg.GetString("database.username"),
		cfg.GetString("database.password"),
		cfg.GetString("database.host"),
		cfg.GetInt("database.port"),
		cfg.GetString("database.database"),
		cfg.GetString("database.charset"),
		cfg.GetString("database.collation"),
		cfg.GetBool("database.parse_time"),
		cfg.GetString("database.loc"),
	)
}

// AutoMigrate 自动迁移数据库表
func (db *DB) AutoMigrate() error {
	log.Println("Running database migrations...")

	// 迁移模型 - 使用嵌入的 gorm.DB 的 AutoMigrate 方法
	if err := db.DB.AutoMigrate(
		&model.Category{},
		&model.Item{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migrations completed")
	return nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}