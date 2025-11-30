package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julesChu12/fly/custos/internal/application/usecase/auth"
	"github.com/julesChu12/fly/custos/internal/config"
	authService "github.com/julesChu12/fly/custos/internal/domain/service/auth"
	"github.com/julesChu12/fly/custos/internal/domain/service/oauth"
	"github.com/julesChu12/fly/custos/internal/domain/service/password"
	"github.com/julesChu12/fly/custos/internal/domain/service/rbac"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
	"github.com/julesChu12/fly/custos/internal/infrastructure/migrate"
	"github.com/julesChu12/fly/custos/internal/infrastructure/persistence/mysql"
	grpcInterface "github.com/julesChu12/fly/custos/internal/interface/grpc"
	"github.com/julesChu12/fly/custos/internal/interface/http/handler"
	authHandler "github.com/julesChu12/fly/custos/internal/interface/http/handler/auth"
	"github.com/julesChu12/fly/custos/internal/interface/http/middleware"
	"github.com/julesChu12/fly/custos/internal/interface/http/router"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Custos HTTP and gRPC servers",
	Long:  `Start the Custos HTTP and gRPC servers to handle user authentication and authorization requests.`,
	Run:   runServer,
}

var (
	configPath string
	httpPort   string
	grpcPort   string
	dbDSN      string
)

func init() {
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "configs/custos.yaml", "Path to configuration file")
	serveCmd.Flags().StringVarP(&httpPort, "http-port", "p", "", "HTTP server port (default: 8081 or from config)")
	serveCmd.Flags().StringVarP(&grpcPort, "grpc-port", "g", "", "gRPC server port (default: 9001 or from config)")
	serveCmd.Flags().StringVarP(&dbDSN, "db-dsn", "d", "", "Database DSN (overrides config)")
}

func runServer(cmd *cobra.Command, args []string) {
	// 如果命令行指定了不同的配置文件，通过环境变量传递给配置加载器
	if configPath != "configs/custos.yaml" {
		log.Printf("Using custom config file: %s", configPath)
		os.Setenv("CONFIG_PATH", configPath)
	}

	cfg := config.MustLoad()

	// 显示加载的配置信息
	log.Printf("✅ Configuration loaded successfully")
	log.Printf("   Database: %s:%d (%s@%s)", cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Database)
	// Redis 配置可能在其他地方处理，这里暂时跳过
	log.Printf("   Note: Redis configuration loaded")

	// Override with command line flags if provided
	if httpPort != "" {
		log.Printf("   HTTP Port override: %s -> %s", cfg.App.Port, httpPort)
		cfg.App.Port = httpPort
	}
	if dbDSN != "" {
		// Override database DSN (need to update config structure to support this)
		log.Printf("Note: Database DSN override via flag not yet implemented in config structure")
	}

	// Initialize logger from configuration
	loggerConfig := logger.Config{
		Level:        cfg.Log.Level,
		Format:       cfg.Log.Format,
		OutputPath:   cfg.Log.OutputPath,
		MaxSize:      cfg.Log.MaxSize,
		MaxBackups:   cfg.Log.MaxBackups,
		MaxAge:       cfg.Log.MaxAge,
		Compress:     cfg.Log.Compress,
		EnableStdout: cfg.Log.EnableStdout,
		EnableFile:   cfg.Log.EnableFile,
	}

	// In development mode, override to console format for better readability
	if cfg.IsDev() {
		loggerConfig.Format = "console"
		loggerConfig.Level = "debug"
	}

	l, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	l.Infof("Starting Custos service in %s mode", cfg.App.Env)
	l.Infof("Log configuration - Level: %s, Format: %s, Stdout: %v, File: %v (%s)",
		loggerConfig.Level, loggerConfig.Format, loggerConfig.EnableStdout,
		loggerConfig.EnableFile, loggerConfig.OutputPath)

	db, err := mysql.NewDatabase(cfg.Database.DSN(), cfg.App.Env == "development")
	if err != nil {
		l.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Get mora db.Client for repository initialization
	dbClient := db.Client()

	// Get raw SQL DB connection for migrations
	sqlDB, err := db.DB().DB()
	if err != nil {
		l.Fatalf("Failed to get raw database connection: %v", err)
	}

	// Run migrations using sql-migrate
	migrationManager := migrate.NewMigrationManager(sqlDB, *l)
	if err := migrationManager.Up(); err != nil {
		l.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories with mora db.Client
	userRepo := mysql.NewUserRepository(dbClient)
	sessionRepo := mysql.NewSessionRepository(dbClient)
	refreshTokenRepo := mysql.NewRefreshTokenRepository(dbClient)
	userOAuthRepo := mysql.NewUserOAuthRepository(dbClient)
	userProfileRepo := mysql.NewUserProfileRepository(db.DB())

	// Initialize password service
	passwordService := password.NewPasswordService()

	tokenService := token.NewTokenService(cfg.JWT.SecretKey, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	authSvc := authService.NewAuthService(userRepo, sessionRepo, refreshTokenRepo, tokenService, passwordService)
	oauthSvc := oauth.NewService(cfg, userRepo, userOAuthRepo, userProfileRepo)

	// Initialize RBAC service
	rbacModelPath := "configs/rbac_model.conf"
	rbacSvc, err := rbac.NewRBACService(db.DB(), rbacModelPath)
	if err != nil {
		l.Fatalf("Failed to initialize RBAC service: %v", err)
	}

	registerUC := auth.NewRegisterUseCase(authSvc, userProfileRepo)
	loginUC := auth.NewLoginUseCase(authSvc, userProfileRepo)
	refreshUC := auth.NewRefreshUseCase(authSvc, userProfileRepo)
	logoutUC := auth.NewLogoutUseCase(authSvc)
	logoutAllUC := auth.NewLogoutAllUseCase(authSvc)

	authHandlerMain := handler.NewAuthHandler(registerUC, loginUC, refreshUC, logoutUC, logoutAllUC)
	passwordHandler := authHandler.NewPasswordHandler(passwordService, userRepo)
	oauthHandler := handler.NewOAuthHandler(oauthSvc, userProfileRepo, tokenService)
	adminHandler := handler.NewAdminHandler(userRepo, userProfileRepo, sessionRepo, rbacSvc)
	profileHandler := handler.NewProfileHandler(userRepo, userProfileRepo)
	healthHandler := handler.NewHealthHandler()
	authMW := middleware.NewAuthMiddleware(tokenService, sessionRepo)

	routerHandler := router.NewRouter(authHandlerMain, oauthHandler, adminHandler, profileHandler, healthHandler, authMW)
	ginEngine := routerHandler.SetupRoutes()

	// Register password routes
	api := ginEngine.Group("/api/v1")
	passwordHandler.RegisterPasswordRoutes(api, authMW.RequireAuth(), nil)

	// Swagger documentation route
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize gRPC server
	grpcServer := grpcInterface.NewCustosGRPCServer(userRepo, userProfileRepo, sessionRepo, tokenService)

	// Override gRPC port if provided
	grpcPortStr := "9001"
	if grpcPort != "" {
		grpcPortStr = grpcPort
	}

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      ginEngine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start gRPC server in background
	go func() {
		l.Infof("gRPC server starting on port %s", grpcPortStr)
		if err := grpcServer.Start(grpcPortStr); err != nil {
			l.Fatalf("gRPC server failed to start: %v", err)
		}
	}()

	go func() {
		l.Infof("HTTP server starting on port %s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown gRPC server
	grpcServer.Stop()

	if err := srv.Shutdown(ctx); err != nil {
		l.Fatalf("Server forced to shutdown: %v", err)
	}

	l.Info("Servers exited successfully")
}
