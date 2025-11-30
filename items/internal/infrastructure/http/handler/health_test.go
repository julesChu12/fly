package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("health check returns healthy status", func(t *testing.T) {
		router := gin.New()
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response.Status)
		assert.Equal(t, "items-service", response.Service)
		assert.Equal(t, "1.0.0", response.Version)
		assert.NotZero(t, response.Timestamp)
	})

	t.Run("health check response structure", func(t *testing.T) {
		router := gin.New()
		router.GET("/health", HealthCheck)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 验证Content-Type是application/json
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		// 验证响应是有效的JSON
		var healthMap map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &healthMap)
		require.NoError(t, err)

		// 验证必需的字段存在
		assert.Contains(t, healthMap, "status")
		assert.Contains(t, healthMap, "timestamp")
		assert.Contains(t, healthMap, "service")
		assert.Contains(t, healthMap, "version")
	})

	t.Run("health check timestamp is recent", func(t *testing.T) {
		router := gin.New()
		router.GET("/health", HealthCheck)

		beforeCall := getCurrentTime()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		afterCall := getCurrentTime()

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// 验证时间戳在合理范围内（调用前后之间）
		assert.True(t, response.Timestamp.After(beforeCall) || response.Timestamp.Equal(beforeCall))
		assert.True(t, response.Timestamp.Before(afterCall) || response.Timestamp.Equal(afterCall))
	})
}

// getCurrentTime 获取当前时间的辅助函数，便于测试
func getCurrentTime() time.Time {
	return time.Now()
}