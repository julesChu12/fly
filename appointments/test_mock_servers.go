package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MockKratosServer Mock Kratos订单服务
type MockKratosServer struct {
	server *http.Server
	port   int
}

// NewMockKratosServer 创建Mock Kratos服务
func NewMockKratosServer(port int) *MockKratosServer {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "kratos-mock",
			"timestamp": time.Now(),
		})
	})

	// 创建订单
	router.POST("/api/orders", func(c *gin.Context) {
		var req struct {
			CustomerID    string  `json:"customer_id" binding:"required"`
			AppointmentID string  `json:"appointment_id" binding:"required"`
			StaffID       string  `json:"staff_id" binding:"required"`
			ServiceID     string  `json:"service_id" binding:"required"`
			Amount        float64 `json:"amount" binding:"required,gt=0"`
			Currency      string  `json:"currency"`
			Description   string  `json:"description"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   gin.H{"message": err.Error()},
			})
			return
		}

		// 创建订单响应
		order := gin.H{
			"id":              uuid.New().String(),
			"order_number":    fmt.Sprintf("ORD%d%s", time.Now().Unix(), uuid.New().String()[:8]),
			"customer_id":     req.CustomerID,
			"appointment_id":  req.AppointmentID,
			"staff_id":        req.StaffID,
			"service_id":      req.ServiceID,
			"amount":          req.Amount,
			"currency":        req.Currency,
			"status":          "pending",
			"payment_status":  "pending",
			"description":     req.Description,
			"order_time":      time.Now(),
			"payment_deadline": time.Now().Add(30 * time.Minute),
			"created_at":      time.Now(),
			"updated_at":      time.Now(),
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    order,
		})
	})

	// 获取订单
	router.GET("/api/orders/:id", func(c *gin.Context) {
		orderID := c.Param("id")

		order := gin.H{
			"id":              orderID,
			"order_number":    "ORD123456789",
			"customer_id":     "customer-001",
			"appointment_id":  "appointment-001",
			"staff_id":        "staff-001",
			"service_id":      "cardiology-consultation",
			"amount":          300.00,
			"currency":        "CNY",
			"status":          "pending",
			"payment_status":  "pending",
			"description":     "测试订单",
			"order_time":      time.Now(),
			"payment_deadline": time.Now().Add(30 * time.Minute),
			"created_at":      time.Now(),
			"updated_at":      time.Now(),
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    order,
		})
	})

	// 更新订单状态
	router.PUT("/api/orders/:id/status", func(c *gin.Context) {
		var req struct {
			Status        string `json:"status" binding:"required"`
			PaymentStatus string `json:"payment_status"`
			Reason        string `json:"reason"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   gin.H{"message": err.Error()},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "订单状态更新成功",
		})
	})

	// 取消订单
	router.DELETE("/api/orders/:id", func(c *gin.Context) {
		orderID := c.Param("id")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "订单取消成功",
			"order_id": orderID,
		})
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return &MockKratosServer{
		server: server,
		port:   port,
	}
}

// Start 启动Mock Kratos服务
func (m *MockKratosServer) Start() error {
	fmt.Printf("🚀 Mock Kratos服务启动在端口 %d\n", m.port)
	return m.server.ListenAndServe()
}

// Stop 停止Mock Kratos服务
func (m *MockKratosServer) Stop() error {
	fmt.Println("🛑 Mock Kratos服务停止")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.server.Shutdown(ctx)
}

// MockPlutusServer Mock Plutus支付服务
type MockPlutusServer struct {
	server *http.Server
	port   int
}

// NewMockPlutusServer 创建Mock Plutus服务
func NewMockPlutusServer(port int) *MockPlutusServer {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "plutus-mock",
			"timestamp": time.Now(),
		})
	})

	// 创建支付
	router.POST("/api/payments", func(c *gin.Context) {
		var req struct {
			OrderID       string  `json:"order_id" binding:"required"`
			AppointmentID string  `json:"appointment_id" binding:"required"`
			CustomerID    string  `json:"customer_id" binding:"required"`
			Amount        float64 `json:"amount" binding:"required,gt=0"`
			Currency      string  `json:"currency"`
			PaymentMethod string  `json:"payment_method" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   gin.H{"message": err.Error()},
			})
			return
		}

		// 创建支付响应
		payment := gin.H{
			"id":            uuid.New().String(),
			"order_id":      req.OrderID,
			"appointment_id": req.AppointmentID,
			"customer_id":    req.CustomerID,
			"amount":        req.Amount,
			"currency":      req.Currency,
			"status":        "pending",
			"payment_method": req.PaymentMethod,
			"transaction_id": fmt.Sprintf("TXN%s", uuid.New().String()[:8]),
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    payment,
		})
	})

	// 获取支付状态
	router.GET("/api/payments/:id", func(c *gin.Context) {
		paymentID := c.Param("id")

		payment := gin.H{
			"id":            paymentID,
			"order_id":      "order-001",
			"appointment_id": "appointment-001",
			"customer_id":    "customer-001",
			"amount":        300.00,
			"currency":      "CNY",
			"status":        "pending",
			"payment_method": "wechat_pay",
			"transaction_id": "TXN123456789",
			"created_at":    time.Now(),
			"updated_at":    time.Now(),
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    payment,
		})
	})

	// 查询支付状态
	router.GET("/api/payments/:id/query", func(c *gin.Context) {
		paymentID := c.Param("id")

		query := gin.H{
			"payment_id":    paymentID,
			"status":        "pending",
			"amount":        "300.00",
			"paid_amount":   "0.00",
			"refund_amount": "0.00",
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    query,
		})
	})

	// 退款
	router.POST("/api/payments/:id/refund", func(c *gin.Context) {
		paymentID := c.Param("id")

		var req struct {
			Amount float64 `json:"amount" binding:"required,gt=0"`
			Reason string  `json:"reason" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   gin.H{"message": err.Error()},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "退款请求成功",
			"payment_id": paymentID,
			"amount":     req.Amount,
			"reason":     req.Reason,
		})
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return &MockPlutusServer{
		server: server,
		port:   port,
	}
}

// Start 启动Mock Plutus服务
func (m *MockPlutusServer) Start() error {
	fmt.Printf("💰 Mock Plutus支付服务启动在端口 %d\n", m.port)
	return m.server.ListenAndServe()
}

// Stop 停止Mock Plutus服务
func (m *MockPlutusServer) Stop() error {
	fmt.Println("🛑 Mock Plutus支付服务停止")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.server.Shutdown(ctx)
}

// 主函数 - 启动Mock服务
func main() {
	fmt.Println("🎭 启动Mock服务集群...")
	fmt.Println("这些服务用于测试预约系统的真实集成")

	// 启动Mock Kratos服务
	kratosServer := NewMockKratosServer(8081)
	go func() {
		if err := kratosServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Mock Kratos服务启动失败: %v", err)
		}
	}()

	// 等待Kratos服务启动
	time.Sleep(1 * time.Second)

	// 启动Mock Plutus服务
	plutusServer := NewMockPlutusServer(8082)
	go func() {
		if err := plutusServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Mock Plutus服务启动失败: %v", err)
		}
	}()

	// 等待Plutus服务启动
	time.Sleep(1 * time.Second)

	fmt.Println("\n✅ Mock服务集群启动完成!")
	fmt.Println("📍 Mock Kratos服务: http://localhost:8081")
	fmt.Println("💰 Mock Plutus服务: http://localhost:8082")
	fmt.Println("\n🧪 现在可以运行集成测试:")
	fmt.Println("   go run test_integration_real_services.go")
	fmt.Println("\n🛑 按 Ctrl+C 停止所有服务")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 正在停止Mock服务...")

	// 优雅关闭服务
	if err := kratosServer.Stop(); err != nil {
		log.Printf("停止Kratos服务失败: %v", err)
	}

	if err := plutusServer.Stop(); err != nil {
		log.Printf("停止Plutus服务失败: %v", err)
	}

	fmt.Println("👋 Mock服务集群已停止")
}