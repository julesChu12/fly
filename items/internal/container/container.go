package container

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/julesChu12/fly/items/internal/application/service"
	"github.com/julesChu12/fly/items/internal/infrastructure/http/handler"
	grpcHandlers "github.com/julesChu12/fly/items/internal/interface/grpc/handlers"
	grpcServerPkg "github.com/julesChu12/fly/items/internal/interface/grpc"
	"github.com/julesChu12/fly/items/internal/infrastructure/persistence/repository"
	"github.com/spf13/viper"
	moraLogger "github.com/julesChu12/fly/mora/pkg/logger"
)

// Container 依赖注入容器
type Container struct {
	DB            *gorm.DB
	ItemHandler   *handler.ItemHandler
	CategoryHandler *handler.CategoryHandler
	SearchHandler *handler.SearchHandler
	StatsHandler  *handler.StatsHandler
	// gRPC 相关
	ItemGRPCServer   *grpcHandlers.ItemServer
	CategoryGRPCServer *grpcHandlers.CategoryServer
	GRPCServer      *grpcServerPkg.Server
}

// NewContainer 创建新的依赖注入容器
func NewContainer(cfg *viper.Viper) (*Container, error) {
	// 初始化数据库连接
	db, err := initDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 自动迁移数据库表
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	// 初始化仓储层
	itemRepo := repository.NewItemRepository(db)
	// categoryRepo := repository.NewCategoryRepository(db) // 待实现

	// 初始化服务层
	itemService := service.NewItemService(itemRepo)

	// 初始化日志器
	appLogger := moraLogger.NewDefault()

	// 创建分类服务（使用领域服务，因为应用层服务不存在）
	// 这里我们需要一个分类仓储，暂时跳过或使用占位符
	// categoryService := category.NewService(categoryRepo) // 需要 categoryRepo

	// 初始化HTTP处理器层
	itemHandler := handler.NewItemHandler(itemService)
	categoryHandler := handler.NewCategoryHandler()
	searchHandler := handler.NewSearchHandler(itemHandler)
	statsHandler := handler.NewStatsHandler(itemHandler)

	// 初始化gRPC处理器
	itemGRPCServer := grpcHandlers.NewItemServer(itemService, appLogger)

	// 暂时跳过分类gRPC服务器，因为没有分类服务
	// categoryGRPCServer := grpcHandlers.NewCategoryServer(categoryService, appLogger)
	var categoryGRPCServer *grpcHandlers.CategoryServer

	// 初始化gRPC服务器
	grpcPort := cfg.GetString("grpc.port")
	if grpcPort == "" {
		grpcPort = "50056" // 默认gRPC端口
	}

	gRPCSrv := grpcServerPkg.NewServer(itemService, nil, grpcPort, appLogger) // categoryService 暂时为 nil

	return &Container{
		DB:            db,
		ItemHandler:   itemHandler,
		CategoryHandler: categoryHandler,
		SearchHandler: searchHandler,
		StatsHandler:  statsHandler,
		// gRPC 相关
		ItemGRPCServer:   itemGRPCServer,
		CategoryGRPCServer: categoryGRPCServer,
		GRPCServer:      gRPCSrv,
	}, nil
}

// initDatabase 初始化数据库连接
func initDatabase(cfg *viper.Viper) (*gorm.DB, error) {
	// 构建DSN - 添加完整的字符集参数以防止乱码，禁用TLS用于开发环境
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=%t&loc=%s&interpolateParams=true&multiStatements=true&tls=false",
		cfg.GetString("database.username"),
		cfg.GetString("database.password"),
		cfg.GetString("database.host"),
		cfg.GetInt("database.port"),
		cfg.GetString("database.database"),
		cfg.GetString("database.charset"),
		"utf8mb4_unicode_ci",
		cfg.GetBool("database.parse_time"),
		cfg.GetString("database.loc"),
	)

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	// 获取底层的sqlDB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.GetInt("database.max_idle_conns"))
	sqlDB.SetMaxOpenConns(cfg.GetInt("database.max_open_conns"))
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.GetInt("database.conn_max_lifetime")) * time.Second)

	log.Printf("Database connected successfully to %s:%d",
		cfg.GetString("database.host"),
		cfg.GetInt("database.port"))

	return db, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	// 这里会调用模型的AutoMigrate方法
	// 由于我们使用手动SQL创建了表，这里暂时跳过
	// 在生产环境中，建议使用GORM的AutoMigrate
	log.Println("Database tables already exist (created by manual migration)")
	return nil
}

// Close 关闭容器资源
func (c *Container) Close() error {
	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}