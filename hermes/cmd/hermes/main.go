// Package main provides the entry point for Hermes Customer Service
// Hermes服务的主入口，提供客户管理功能，支持HTTP和gRPC双协议
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/hermes/internal/application/service"
	"github.com/julesChu12/fly/hermes/internal/infrastructure/database"
	"github.com/julesChu12/fly/hermes/internal/interface/grpc"
	"github.com/julesChu12/fly/hermes/internal/interface/http"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	grpcLib "google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title Hermes Customer Service API
// @version 1.0
// @description 客户管理服务API文档，支持客户信息和联系方式的增删改查操作
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.fly.com/support
// @contact.email support@fly.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8083
// @BasePath /api

// @tag.name 客户管理
// @tag.description 客户信息管理相关接口

func main() {
	log.Println("Starting Hermes Customer Service...")

	// 初始化数据库连接
	db, err := initDatabase()
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// 初始化Repository层
	customerRepo := database.NewCustomerRepository(db)
	contactRepo := database.NewContactRepository(db)

	// 初始化Service层
	customerService := service.NewCustomerService(customerRepo, contactRepo)

	// 启动HTTP服务器
	go func() {
		if err := startHTTPServer(customerService); err != nil {
			log.Fatal("HTTP server failed:", err)
		}
	}()

	// 启动gRPC服务器
	go func() {
		if err := startGRPCServer(customerService); err != nil {
			log.Fatal("gRPC server failed:", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Hermes service...")

	// 关闭数据库连接
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	log.Println("Hermes service stopped")
}

// initDatabase 初始化数据库连接
func initDatabase() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/hermes?charset=utf8mb4&parseTime=True&loc=Local"
	}

	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// startHTTPServer 启动HTTP服务器，提供REST API和Swagger文档
func startHTTPServer(customerService service.CustomerService) error {
	r := gin.New()

	// 添加中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "hermes",
		})
	})

	// API路由组
	api := r.Group("/api")
	customerHandler := http.NewCustomerHandler(customerService)
	customerHandler.RegisterRoutes(api)

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("HTTP server starting on port %s", port)
	return r.Run(":" + port)
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(customerService service.CustomerService) error {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "9080"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	s := grpcLib.NewServer()

	// 注册gRPC服务
	customerHandler := grpc.NewCustomerGRPCHandler(customerService)
	_ = customerHandler // 暂时避免未使用变量警告

	log.Printf("gRPC server starting on port %s", port)
	return s.Serve(lis)
}
