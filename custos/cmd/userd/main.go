package main

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
	"github.com/julesChu12/fly/custos/internal/domain/service/rbac"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
	"github.com/julesChu12/fly/custos/internal/infrastructure/migrate"
	"github.com/julesChu12/fly/custos/internal/infrastructure/persistence/mysql"
	"github.com/julesChu12/fly/custos/internal/interface/http/handler"
	"github.com/julesChu12/fly/custos/internal/interface/http/middleware"
	"github.com/julesChu12/fly/custos/internal/interface/http/router"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

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

	userRepo := mysql.NewUserRepository(db.DB())
	sessionRepo := mysql.NewSessionRepository(db.DB())
	refreshTokenRepo := mysql.NewRefreshTokenRepository(db.DB())
	userOAuthRepo := mysql.NewUserOAuthRepository(db.DB())

	tokenService := token.NewTokenService(cfg.JWT.SecretKey, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	authSvc := authService.NewAuthService(userRepo, sessionRepo, refreshTokenRepo, tokenService)
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

	authHandler := handler.NewAuthHandler(registerUC, loginUC, refreshUC, logoutUC, logoutAllUC)
	userHandler := handler.NewUserHandler()
	oauthHandler := handler.NewOAuthHandler(oauthSvc, tokenService)
	adminHandler := handler.NewAdminHandler(userRepo, rbacSvc)
	healthHandler := handler.NewHealthHandler()
	authMW := middleware.NewAuthMiddleware(tokenService, sessionRepo)

	routerHandler := router.NewRouter(authHandler, userHandler, oauthHandler, adminHandler, healthHandler, authMW)
	ginEngine := routerHandler.SetupRoutes()

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      ginEngine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
