package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/custos/internal/application/usecase/auth"
	"github.com/julesChu12/fly/custos/internal/config"
	"github.com/julesChu12/fly/custos/internal/domain/repository"
	authService "github.com/julesChu12/fly/custos/internal/domain/service/auth"
	"github.com/julesChu12/fly/custos/internal/domain/service/oauth"
	"github.com/julesChu12/fly/custos/internal/domain/service/password"
	"github.com/julesChu12/fly/custos/internal/domain/service/rbac"
	"github.com/julesChu12/fly/custos/internal/domain/service/token"
	"github.com/julesChu12/fly/custos/internal/infrastructure/migrate"
	"github.com/julesChu12/fly/custos/internal/infrastructure/persistence/mysql"
	"github.com/julesChu12/fly/custos/internal/interface/http/handler"
	authHandler "github.com/julesChu12/fly/custos/internal/interface/http/handler/auth"
	"github.com/julesChu12/fly/custos/internal/interface/http/middleware"
	"github.com/julesChu12/fly/custos/internal/interface/http/router"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/stretchr/testify/require"
)

// TestServer holds the test environment
type TestServer struct {
	Router           *gin.Engine
	DB               *mysql.Database
	TokenService     *token.TokenService
	UserRepo         repository.UserRepository
	SessionRepo      repository.SessionRepository
	RefreshTokenRepo repository.RefreshTokenRepository
	UserOAuthRepo    repository.UserOAuthRepository
	UserProfileRepo  repository.UserProfileRepository
	RBACService      *rbac.RBACService
	Config           *config.Config
}

// SetupTestServer initializes a test server with all dependencies
func SetupTestServer(t *testing.T) *TestServer {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Load test configuration
	cfg := &config.Config{
		App: config.AppConfig{
			Port: "8080",
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "root",
			Password: "",
			Database: "custos_test",
			Charset:  "utf8mb4",
		},
		JWT: config.JWTConfig{
			SecretKey:       "test-secret-key-for-integration-tests",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
	}

	// Initialize logger
	loggerConfig := logger.Config{
		Level:  "error", // Set to error to reduce test output noise
		Format: "json",
	}
	l, err := logger.New(loggerConfig)
	require.NoError(t, err, "Failed to initialize logger")

	// Connect to database
	db, err := mysql.NewDatabase(cfg.Database.DSN(), true)
	require.NoError(t, err, "Failed to connect to test database")

	// Get raw SQL DB connection for migrations
	sqlDB, err := db.DB().DB()
	require.NoError(t, err, "Failed to get raw database connection")

	// Run migrations
	migrationManager := migrate.NewMigrationManager(sqlDB, *l)
	err = migrationManager.Up()
	require.NoError(t, err, "Failed to run migrations")

	// Clean up test data
	CleanupTestData(t, db)

	// Initialize repositories
	userRepo := mysql.NewUserRepository(db.DB())
	sessionRepo := mysql.NewSessionRepository(db.DB())
	refreshTokenRepo := mysql.NewRefreshTokenRepository(db.DB())
	userOAuthRepo := mysql.NewUserOAuthRepository(db.DB())
	userProfileRepo := mysql.NewUserProfileRepository(db.DB())

	// Initialize services
	passwordService := password.NewPasswordService()
	tokenService := token.NewTokenService(cfg.JWT.SecretKey, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	authSvc := authService.NewAuthService(userRepo, sessionRepo, refreshTokenRepo, tokenService, passwordService)
	oauthSvc := oauth.NewService(cfg, userRepo, userOAuthRepo)

	// Initialize RBAC service
	rbacModelPath := "../../configs/rbac_model.conf"
	rbacSvc, err := rbac.NewRBACService(db.DB(), rbacModelPath)
	require.NoError(t, err, "Failed to initialize RBAC service")

	// Initialize use cases
	registerUC := auth.NewRegisterUseCase(authSvc)
	loginUC := auth.NewLoginUseCase(authSvc)
	refreshUC := auth.NewRefreshUseCase(authSvc)
	logoutUC := auth.NewLogoutUseCase(authSvc)
	logoutAllUC := auth.NewLogoutAllUseCase(authSvc)

	// Initialize handlers
	authHandlerMain := handler.NewAuthHandler(registerUC, loginUC, refreshUC, logoutUC, logoutAllUC)
	passwordHandler := authHandler.NewPasswordHandler(passwordService, userRepo)
	userHandler := handler.NewUserHandler()
	oauthHandler := handler.NewOAuthHandler(oauthSvc, tokenService)
	adminHandler := handler.NewAdminHandler(userRepo, sessionRepo, rbacSvc)
	profileHandler := handler.NewProfileHandler(userRepo, userProfileRepo)
	healthHandler := handler.NewHealthHandler()
	authMW := middleware.NewAuthMiddleware(tokenService, sessionRepo)

	// Setup router
	routerHandler := router.NewRouter(authHandlerMain, userHandler, oauthHandler, adminHandler, profileHandler, healthHandler, authMW)
	ginEngine := routerHandler.SetupRoutes()

	// Register password routes
	api := ginEngine.Group("/api/v1")
	passwordHandler.RegisterPasswordRoutes(api, authMW.RequireAuth(), nil)

	return &TestServer{
		Router:           ginEngine,
		DB:               db,
		TokenService:     tokenService,
		UserRepo:         userRepo,
		SessionRepo:      sessionRepo,
		RefreshTokenRepo: refreshTokenRepo,
		UserOAuthRepo:    userOAuthRepo,
		UserProfileRepo:  userProfileRepo,
		RBACService:      rbacSvc,
		Config:           cfg,
	}
}

// Cleanup closes the database connection and cleans up test data
func (ts *TestServer) Cleanup(t *testing.T) {
	CleanupTestData(t, ts.DB)
	ts.DB.Close()
}

// CleanupTestData removes all test data from the database
func CleanupTestData(t *testing.T, db *mysql.Database) {
	// Order matters: delete child tables before parent tables
	tables := []string{
		"sessions",
		"refresh_tokens",
		"user_oauth_bindings",
		"user_profiles",
		"users",
	}

	for _, table := range tables {
		err := db.DB().Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		require.NoError(t, err, "Failed to clean up %s table", table)
	}
}

// MakeRequest makes an HTTP request to the test server
func (ts *TestServer) MakeRequest(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	return w
}

// MakeAuthenticatedRequest makes an authenticated HTTP request
func (ts *TestServer) MakeAuthenticatedRequest(method, path string, body interface{}, accessToken string) *httptest.ResponseRecorder {
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
	}
	return ts.MakeRequest(method, path, body, headers)
}

// RegisterTestUser registers a test user and returns the response
func (ts *TestServer) RegisterTestUser(t *testing.T, email, password string) map[string]interface{} {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/register", reqBody, nil)
	require.Equal(t, http.StatusOK, w.Code, "Failed to register test user: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Failed to unmarshal register response")

	return response
}

// LoginTestUser logs in a test user and returns access token
func (ts *TestServer) LoginTestUser(t *testing.T, email, password string) string {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	w := ts.MakeRequest(http.MethodPost, "/api/v1/auth/login", reqBody, nil)
	require.Equal(t, http.StatusOK, w.Code, "Failed to login test user: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Failed to unmarshal login response")

	accessToken, ok := response["access_token"].(string)
	require.True(t, ok, "Access token not found in login response")

	return accessToken
}

// RegisterAndLoginTestUser is a helper that registers and logs in a user
func (ts *TestServer) RegisterAndLoginTestUser(t *testing.T, email, password string) string {
	ts.RegisterTestUser(t, email, password)
	return ts.LoginTestUser(t, email, password)
}

// AssertJSONResponse asserts the JSON response matches expected values
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedFields map[string]interface{}) {
	require.Equal(t, expectedStatus, w.Code, "Unexpected status code: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Failed to unmarshal response")

	for key, expected := range expectedFields {
		actual, exists := response[key]
		require.True(t, exists, "Field %s not found in response", key)
		require.Equal(t, expected, actual, "Field %s has unexpected value", key)
	}
}

// AssertErrorResponse asserts the response is an error with expected message
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	require.Equal(t, expectedStatus, w.Code, "Unexpected status code: %s", w.Body.String())

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Failed to unmarshal error response")

	errorMsg, exists := response["error"].(string)
	require.True(t, exists, "Error field not found in response")
	require.Contains(t, errorMsg, expectedError, "Error message doesn't contain expected text")
}
