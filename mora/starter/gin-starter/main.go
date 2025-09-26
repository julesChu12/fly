package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/auth"

	ginauth "github.com/julesChu12/fly/mora/adapters/gin"
	_ "github.com/julesChu12/fly/mora/starter/gin-starter/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	// JWTSecret is the secret key for JWT signing
	JWTSecret = "your-super-secret-key-change-in-production"
	// TokenTTL is the time-to-live for access tokens
	TokenTTL = 10 * time.Minute
)

// @title Mora API
// @version 1.0
// @description Mora能力库演示API - 提供JWT认证和业务接口示例
// @termsOfService http://swagger.io/terms/

// @contact.name Mora API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	r := gin.Default()

	// Configure auth middleware
	authConfig := ginauth.AuthMiddlewareConfig{
		Secret:    JWTSecret,
		SkipPaths: []string{"/health", "/login", "/swagger/*"},
	}

	// Apply auth middleware globally (except for skip paths)
	r.Use(ginauth.AuthMiddleware(authConfig))

	// Public routes (no authentication required)
	r.GET("/health", healthHandler)
	r.POST("/login", loginHandler)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Protected routes (authentication required)
	r.GET("/profile", profileHandler)
	r.GET("/protected", protectedHandler)

	// Business API routes
	api := r.Group("/api/v1")
	{
		api.GET("/orders", getOrdersHandler)
		api.POST("/orders", createOrderHandler)
		api.GET("/users", getUsersHandler)
	}

	r.Run(":8080")
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
	Time   string `json:"time" example:"2023-12-25T15:30:45Z"`
}

// @Summary Health Check
// @Description 系统健康检查接口
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "ok",
		Time:   time.Now().Format(time.RFC3339),
	})
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int    `json:"expires_in" example:"600"`
	UserID      string `json:"user_id" example:"user-123"`
	Username    string `json:"username" example:"admin"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error" example:"authentication failed"`
	Message string `json:"message" example:"invalid username or password"`
}

// @Summary User Login
// @Description 用户登录接口，返回Access Token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /login [post]
func loginHandler(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	// Mock authentication - in production, validate against UserService
	if req.Username == "admin" && req.Password == "password" {
		// Generate access token
		token, err := auth.GenerateToken("user-123", req.Username, JWTSecret, TokenTTL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "token generation failed",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, LoginResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int(TokenTTL.Seconds()),
			UserID:      "user-123",
			Username:    req.Username,
		})
		return
	}

	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error:   "authentication failed",
		Message: "invalid username or password",
	})
}

// ProfileResponse represents profile response
type ProfileResponse struct {
	UserID   string `json:"user_id" example:"user-123"`
	Username string `json:"username" example:"admin"`
	Subject  string `json:"subject" example:"user-123"`
	Exp      string `json:"exp" example:"2023-12-25T16:30:45Z"`
	Iat      string `json:"iat" example:"2023-12-25T15:30:45Z"`
}

// @Summary Get User Profile
// @Description 获取当前用户的个人信息
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProfileResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /profile [get]
func profileHandler(c *gin.Context) {
	userID := ginauth.GetUserID(c)
	claims := ginauth.GetClaims(c)

	if claims == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "failed to get user claims",
		})
		return
	}

	c.JSON(http.StatusOK, ProfileResponse{
		UserID:   userID,
		Username: claims.Username,
		Subject:  claims.Subject,
		Exp:      claims.ExpiresAt.Time.Format(time.RFC3339),
		Iat:      claims.IssuedAt.Time.Format(time.RFC3339),
	})
}

// ProtectedResponse represents protected endpoint response
type ProtectedResponse struct {
	Message string `json:"message" example:"This is a protected endpoint"`
	UserID  string `json:"user_id" example:"user-123"`
	Time    string `json:"time" example:"2023-12-25T15:30:45Z"`
}

// @Summary Protected Endpoint
// @Description 受保护的接口示例
// @Tags Protected
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProtectedResponse
// @Failure 401 {object} ErrorResponse
// @Router /protected [get]
func protectedHandler(c *gin.Context) {
	userID := ginauth.GetUserID(c)
	c.JSON(http.StatusOK, ProtectedResponse{
		Message: "This is a protected endpoint",
		UserID:  userID,
		Time:    time.Now().Format(time.RFC3339),
	})
}

// Order represents order information
type Order struct {
	ID     string  `json:"id" example:"order-1"`
	UserID string  `json:"user_id" example:"user-123"`
	Amount float64 `json:"amount" example:"100.00"`
	Status string  `json:"status" example:"completed"`
}

// OrdersResponse represents orders list response
type OrdersResponse struct {
	Orders []Order `json:"orders"`
	Total  int     `json:"total" example:"2"`
}

// @Summary Get Orders
// @Description 获取用户订单列表
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} OrdersResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/orders [get]
func getOrdersHandler(c *gin.Context) {
	userID := ginauth.GetUserID(c)

	// Mock orders data - in production, query from database
	orders := []Order{
		{ID: "order-1", UserID: userID, Amount: 100.00, Status: "completed"},
		{ID: "order-2", UserID: userID, Amount: 250.50, Status: "pending"},
	}

	c.JSON(http.StatusOK, OrdersResponse{
		Orders: orders,
		Total:  len(orders),
	})
}

// CreateOrderRequest represents create order request
type CreateOrderRequest struct {
	Amount      float64 `json:"amount" binding:"required" example:"100.00"`
	Description string  `json:"description" example:"订单描述"`
}

// CreateOrderResponse represents create order response
type CreateOrderResponse struct {
	Order Order `json:"order"`
}

// @Summary Create Order
// @Description 创建新订单
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateOrderRequest true "创建订单请求"
// @Success 201 {object} CreateOrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/orders [post]
func createOrderHandler(c *gin.Context) {
	userID := ginauth.GetUserID(c)

	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	// Mock order creation
	order := Order{
		ID:     "order-" + time.Now().Format("20060102150405"),
		UserID: userID,
		Amount: req.Amount,
		Status: "created",
	}

	c.JSON(http.StatusCreated, CreateOrderResponse{
		Order: order,
	})
}

// User represents user information
type User struct {
	ID       string `json:"id" example:"user-123"`
	Username string `json:"username" example:"admin"`
	Role     string `json:"role" example:"admin"`
}

// UsersResponse represents users list response
type UsersResponse struct {
	Users     []User `json:"users"`
	Total     int    `json:"total" example:"2"`
	RequestBy string `json:"request_by" example:"user-123"`
}

// @Summary Get Users
// @Description 获取用户列表（管理员功能）
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UsersResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users [get]
func getUsersHandler(c *gin.Context) {
	userID := ginauth.GetUserID(c)

	// Mock users data
	users := []User{
		{ID: "user-123", Username: "admin", Role: "admin"},
		{ID: "user-456", Username: "user1", Role: "user"},
	}

	c.JSON(http.StatusOK, UsersResponse{
		Users:     users,
		Total:     len(users),
		RequestBy: userID,
	})
}
