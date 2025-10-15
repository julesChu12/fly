package cmd

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	pb "github.com/julesChu12/fly/hermes/api/proto"
	"github.com/julesChu12/fly/hermes/internal/application/service"
	"github.com/julesChu12/fly/hermes/internal/domain/repository"
	"github.com/julesChu12/fly/hermes/internal/infrastructure/database"
	grpcInterface "github.com/julesChu12/fly/hermes/internal/interface/grpc"
	httpInterface "github.com/julesChu12/fly/hermes/internal/interface/http"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	grpcLib "google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Hermes HTTP and gRPC servers",
	Long:  `Start the Hermes HTTP and gRPC servers to handle customer management requests.`,
	Run:   runServer,
}

var (
	configPath string
	httpPort   string
	grpcPort   string
	dbDSN      string
)

func init() {
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "configs/hermes.yaml", "Path to configuration file")
	serveCmd.Flags().StringVarP(&httpPort, "http-port", "p", "", "HTTP server port (default: 8080 or from config)")
	serveCmd.Flags().StringVarP(&grpcPort, "grpc-port", "g", "", "gRPC server port (default: 9080 or from config)")
	serveCmd.Flags().StringVarP(&dbDSN, "db-dsn", "d", "", "Database DSN (overrides config)")
}

func runServer(cmd *cobra.Command, args []string) {
	log.Println("Starting Hermes Customer Service...")

	// Get database DSN
	dsn := dbDSN
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/hermes?charset=utf8mb4&parseTime=True&loc=Local"
	}

	// Initialize database connection
	db, err := initDatabase(dsn)
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Initialize Repository layer
	customerRepo := database.NewCustomerRepository(db)
	contactRepo := database.NewContactRepository(db)

	// Initialize Service layer
	customerService := service.NewCustomerService(customerRepo, contactRepo)
	contactService := service.NewContactService(contactRepo)

	// Get ports
	httpPortStr := httpPort
	if httpPortStr == "" {
		httpPortStr = os.Getenv("PORT")
	}
	if httpPortStr == "" {
		httpPortStr = "8080"
	}

	grpcPortStr := grpcPort
	if grpcPortStr == "" {
		grpcPortStr = os.Getenv("GRPC_PORT")
	}
	if grpcPortStr == "" {
		grpcPortStr = "9080"
	}

	// Start servers
	httpServer := startHTTPServer(customerService, contactService, httpPortStr)
	grpcServer, grpcListener := startGRPCServer(customerService, contactRepo, grpcPortStr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Hermes service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	// Shutdown gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		grpcListener.Close()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		log.Println("gRPC server shutdown timeout, forcing stop")
		grpcServer.Stop()
	case <-stopped:
		log.Println("gRPC server stopped gracefully")
	}

	// Close database connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	log.Println("Hermes service stopped")
}

// initDatabase initializes database connection
func initDatabase(dsn string) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// startHTTPServer starts HTTP server providing REST API and Swagger documentation
func startHTTPServer(customerService service.CustomerService, contactService service.ContactService, port string) *http.Server {
	r := gin.New()

	// Add middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Add CORS middleware
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

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "hermes",
		})
	})

	// API routes
	api := r.Group("/api")
	customerHandler := httpInterface.NewCustomerHandler(customerService)
	customerHandler.RegisterRoutes(api)

	contactHandler := httpInterface.NewContactHandler(contactService)
	contactHandler.RegisterRoutes(api)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("HTTP server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	return srv
}

// startGRPCServer starts gRPC server
func startGRPCServer(customerService service.CustomerService, contactRepo repository.ContactRepository, port string) (*grpcLib.Server, net.Listener) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	s := grpcLib.NewServer()

	// Register gRPC services
	customerHandler := grpcInterface.NewCustomerGRPCHandler(customerService)
	pb.RegisterCustomerServiceServer(s, customerHandler)

	contactHandler := grpcInterface.NewContactGRPCHandler(contactRepo)
	pb.RegisterContactServiceServer(s, contactHandler)

	// Start server in goroutine
	go func() {
		log.Printf("gRPC server starting on port %s", port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	return s, lis
}
