package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/kratos/internal/infrastructure/custos"
	"github.com/stretchr/testify/assert"
)

// MockCustosClient implements a mock Custos client for testing
type MockCustosClient struct {
	ValidateTokenFunc func(ctx context.Context, token string) (*custos.TokenValidationResult, error)
}

func (m *MockCustosClient) ValidateToken(ctx context.Context, token string) (*custos.TokenValidationResult, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, token)
	}
	return &custos.TokenValidationResult{IsValid: true}, nil
}

func TestAuthMiddleware_RequireAuth_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock Custos client
	mockClient := &MockCustosClient{
		ValidateTokenFunc: func(ctx context.Context, token string) (*custos.TokenValidationResult, error) {
			if token == "valid-token" {
				return &custos.TokenValidationResult{
					IsValid:  true,
					UserID:   123,
					Username: "testuser",
					Email:    "test@example.com",
					TenantID: 456,
					UserType: "customer",
				}, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	// Create test router
	router := gin.New()
	authMW := NewAuthMiddleware(mockClient, []string{"/health", "/swagger/*"})
	router.Use(authMW.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get(UserIDKey)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
		})
	})

	// Create test request with valid token
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "123")
}

func TestAuthMiddleware_RequireAuth_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock Custos client
	mockClient := &MockCustosClient{}
	authMW := NewAuthMiddleware(mockClient, []string{})

	router := gin.New()
	router.Use(authMW.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Create test request without token
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing authorization token")
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock Custos client that returns error
	mockClient := &MockCustosClient{
		ValidateTokenFunc: func(ctx context.Context, token string) (*custos.TokenValidationResult, error) {
			return nil, errors.New("validation failed")
		},
	}

	authMW := NewAuthMiddleware(mockClient, []string{})
	router := gin.New()
	router.Use(authMW.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Create test request with invalid token
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to validate token")
}

func TestAuthMiddleware_ShouldSkipPath(t *testing.T) {
	tests := []struct {
		name      string
		skipPaths []string
		testPath  string
		expected  bool
	}{
		{
			name:      "exact match",
			skipPaths: []string{"/health", "/metrics"},
			testPath:  "/health",
			expected:  true,
		},
		{
			name:      "wildcard match",
			skipPaths: []string{"/swagger/*"},
			testPath:  "/swagger/index.html",
			expected:  true,
		},
		{
			name:      "no match",
			skipPaths: []string{"/health"},
			testPath:  "/api/orders",
			expected:  false,
		},
		{
			name:      "wildcard deep match",
			skipPaths: []string{"/api/v1/*"},
			testPath:  "/api/v1/orders",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authMW := &AuthMiddleware{
				skipPaths: tt.skipPaths,
			}
			result := authMW.shouldSkipPath(tt.testPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("user ID exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(UserIDKey, uint(123))

		userID, err := GetUserID(c)
		assert.NoError(t, err)
		assert.Equal(t, uint(123), userID)
	})

	t.Run("user ID not found", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		_, err := GetUserID(c)
		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
	})

	t.Run("user ID wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(UserIDKey, "invalid")

		_, err := GetUserID(c)
		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
	})
}

func TestGetTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("tenant ID exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(TenantIDKey, uint(456))

		tenantID, err := GetTenantID(c)
		assert.NoError(t, err)
		assert.Equal(t, uint(456), tenantID)
	})

	t.Run("tenant ID not found", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		tenantID, err := GetTenantID(c)
		assert.NoError(t, err)
		assert.Equal(t, uint(0), tenantID) // Tenant ID is optional
	})

	t.Run("tenant ID wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(TenantIDKey, "invalid")

		tenantID, err := GetTenantID(c)
		assert.NoError(t, err)
		assert.Equal(t, uint(0), tenantID) // Should return 0 for invalid type
	})
}

func TestGetUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("username exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(UsernameKey, "testuser")

		username := GetUsername(c)
		assert.Equal(t, "testuser", username)
	})

	t.Run("username not found", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		username := GetUsername(c)
		assert.Equal(t, "", username)
	})

	t.Run("username wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(UsernameKey, 123)

		username := GetUsername(c)
		assert.Equal(t, "", username)
	})
}

func TestGetUserType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("user type exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(UserTypeKey, "admin")

		userType := GetUserType(c)
		assert.Equal(t, "admin", userType)
	})

	t.Run("user type not found", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		userType := GetUserType(c)
		assert.Equal(t, "", userType)
	})
}

func TestExtractToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		headerValue  string
		expected    string
	}{
		{
			name:        "valid bearer token",
			headerValue: "Bearer token123",
			expected:    "token123",
		},
		{
			name:        "empty header",
			headerValue:  "",
			expected:    "",
		},
		{
			name:        "missing bearer prefix",
			headerValue: "token123",
			expected:    "",
		},
		{
			name:        "wrong prefix",
			headerValue: "Basic token123",
			expected:    "",
		},
		{
			name:        "bearer with extra space",
			headerValue: "Bearer  token123",
			expected:    " token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", "/", nil)
			if tt.headerValue != "" {
				c.Request.Header.Set("Authorization", tt.headerValue)
			}

			token := extractToken(c)
			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestAuthMiddleware_ContextInjection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock Custos client with full user data
	mockClient := &MockCustosClient{
		ValidateTokenFunc: func(ctx context.Context, token string) (*custos.TokenValidationResult, error) {
			return &custos.TokenValidationResult{
				IsValid:  true,
				UserID:   123,
				Username: "testuser",
				Email:    "test@example.com",
				TenantID: 456,
				UserType: "customer",
			}, nil
		},
	}

	authMW := NewAuthMiddleware(mockClient, []string{})
	router := gin.New()
	router.Use(authMW.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		userID := MustGetUserID(c)
		username := GetUsername(c)
		tenantID, _ := GetTenantID(c)
		userType := GetUserType(c)
		email, _ := c.Get(EmailKey)

		c.JSON(http.StatusOK, gin.H{
			"user_id":   userID,
			"username": username,
			"tenant_id": tenantID,
			"user_type": userType,
			"email":    email,
		})
	})

	// Create test request with valid token
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "123")
	assert.Contains(t, w.Body.String(), "testuser")
	assert.Contains(t, w.Body.String(), "456")
	assert.Contains(t, w.Body.String(), "customer")
}

func TestMustGetUserID_Panic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	assert.Panics(t, func() {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		// User ID not set, should panic
		MustGetUserID(c)
	}, "user_id not found in context")
}

func TestAuthMiddleware_SkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock Custos client
	mockClient := &MockCustosClient{}

	authMW := NewAuthMiddleware(mockClient, []string{"/health", "/api/docs/*"})
	router := gin.New()
	router.Use(authMW.RequireAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})
	router.GET("/api/docs/index.html", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	tests := []struct {
		name      string
		path      string
		expectAuth bool
	}{
		{
			name:      "skip exact match",
			path:      "/health",
			expectAuth: false,
		},
		{
			name:      "skip wildcard match",
			path:      "/api/docs/index.html",
			expectAuth: false,
		},
		{
			name:      "require auth",
			path:      "/test",
			expectAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			router.ServeHTTP(w, req)

			if tt.expectAuth {
				assert.Equal(t, http.StatusUnauthorized, w.Code, "Should require auth for path: %s", tt.path)
			} else {
				assert.Equal(t, http.StatusOK, w.Code, "Should skip auth for path: %s", tt.path)
			}
		})
	}
}
