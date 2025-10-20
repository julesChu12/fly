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

	_ "github.com/julesChu12/fly/custos/api/swagger" // Swagger docs
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
	cfg := config.MustLoad()

	// Override with command line flags if provided
	if httpPort != "" {
		cfg.App.Port = httpPort
	}
	if dbDSN != "" {
		// Override database DSN (need to update config structure to support this)
		log.Printf("Note: Database DSN override via flag not yet implemented in config structure")
	}

	// Initialize logger
	loggerConfig := logger.Config{
		Level:  "info",
		Format: "json",
	}
	l, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	db, err := mysql.NewDatabase(cfg.Database.DSN(), cfg.App.Env == "development")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Get mora db.Client for repository initialization
	dbClient := db.Client()

	// Get raw SQL DB connection for migrations
	sqlDB, err := db.DB().DB()
	if err != nil {
		log.Fatalf("Failed to get raw database connection: %v", err)
	}

	// Run migrations using sql-migrate
	migrationManager := migrate.NewMigrationManager(sqlDB, *l)
	if err := migrationManager.Up(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
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
	oauthSvc := oauth.NewService(cfg, userRepo, userOAuthRepo)

	// Initialize RBAC service
	rbacModelPath := "configs/rbac_model.conf"
	rbacSvc, err := rbac.NewRBACService(db.DB(), rbacModelPath)
	if err != nil {
		log.Fatalf("Failed to initialize RBAC service: %v", err)
	}

	registerUC := auth.NewRegisterUseCase(authSvc)
	loginUC := auth.NewLoginUseCase(authSvc)
	refreshUC := auth.NewRefreshUseCase(authSvc)
	logoutUC := auth.NewLogoutUseCase(authSvc)
	logoutAllUC := auth.NewLogoutAllUseCase(authSvc)

	authHandlerMain := handler.NewAuthHandler(registerUC, loginUC, refreshUC, logoutUC, logoutAllUC)
	passwordHandler := authHandler.NewPasswordHandler(passwordService, userRepo)
	userHandler := handler.NewUserHandler()
	oauthHandler := handler.NewOAuthHandler(oauthSvc, tokenService)
	adminHandler := handler.NewAdminHandler(userRepo, sessionRepo, rbacSvc)
	profileHandler := handler.NewProfileHandler(userRepo, userProfileRepo)
	healthHandler := handler.NewHealthHandler()
	authMW := middleware.NewAuthMiddleware(tokenService, sessionRepo)

	routerHandler := router.NewRouter(authHandlerMain, userHandler, oauthHandler, adminHandler, profileHandler, healthHandler, authMW)
	ginEngine := routerHandler.SetupRoutes()

	// Register password routes
	api := ginEngine.Group("/api/v1")
	passwordHandler.RegisterPasswordRoutes(api, authMW.RequireAuth(), nil)

	// Swagger documentation route
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize gRPC server
	grpcServer := grpcInterface.NewCustosGRPCServer(userRepo, sessionRepo, tokenService)

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
		log.Printf("gRPC server starting on port %s", grpcPortStr)
		if err := grpcServer.Start(grpcPortStr); err != nil {
			log.Fatalf("gRPC server failed to start: %v", err)
		}
	}()

	go func() {
		log.Printf("HTTP server starting on port %s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown gRPC server
	grpcServer.Stop()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Servers exited")
}
