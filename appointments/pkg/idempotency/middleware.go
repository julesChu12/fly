package idempotency

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/julesChu12/fly/mora/pkg/logger"
)

// IdempotencyMiddleware 幂等性中间件
type IdempotencyMiddleware struct {
	manager IdempotencyManager
	logger  *logger.Logger
	config  *MiddlewareConfig
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	// 需要幂等性的HTTP方法
	RequiredMethods []string `yaml:"required_methods"`
	// 幂等性键的HTTP头
	IdempotencyKeyHeader string `yaml:"idempotency_key_header"`
	// 默认TTL
	DefaultTTL time.Duration `yaml:"default_ttl"`
	// 最大TTL
	MaxTTL time.Duration `yaml:"max_ttl"`
	// 是否记录请求体用于调试
	LogRequestBody bool `yaml:"log_request_body"`
}

// DefaultMiddlewareConfig 默认中间件配置
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		RequiredMethods:      []string{"POST", "PUT", "PATCH", "DELETE"},
		IdempotencyKeyHeader: "X-Idempotency-Key",
		DefaultTTL:           24 * time.Hour,
		MaxTTL:               7 * 24 * time.Hour,
		LogRequestBody:       false,
	}
}

// NewIdempotencyMiddleware 创建幂等性中间件
func NewIdempotencyMiddleware(manager IdempotencyManager, config *MiddlewareConfig, logger *logger.Logger) *IdempotencyMiddleware {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	return &IdempotencyMiddleware{
		manager: manager,
		logger:  logger,
		config:  config,
	}
}

// Middleware 返回Gin中间件函数
func (m *IdempotencyMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要幂等性处理
		if !m.requiresIdempotency(c.Request.Method) {
			c.Next()
			return
		}

		// 获取幂等性键
		idempotencyKey := m.getIdempotencyKey(c)
		if idempotencyKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "幂等性键不能为空",
				"code":  "MISSING_IDEMPOTENCY_KEY",
			})
			c.Abort()
			return
		}

		// 构建完整的键名
		fullKey := m.buildFullKey(c, idempotencyKey)

		// 检查幂等性
		isFirst, err := m.manager.CheckAndRecord(c.Request.Context(), fullKey, m.getDefaultTTL())
		if err != nil {
			m.logger.Error("幂等性检查失败",
				map[string]interface{}{
					"key":   fullKey,
					"error": err,
				})
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "幂等性检查失败",
				"code":  "IDEMPOTENCY_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		// 如果不是第一次请求，返回缓存的结果
		if !isFirst {
			m.handleDuplicateRequest(c, fullKey)
			return
		}

		// 记录请求体（用于调试）
		if m.config.LogRequestBody {
			m.logRequestBody(c, fullKey)
		}

		// 设置响应拦截器来保存结果
		c.Set("idempotency_key", fullKey)
		c.Next()
		m.saveResponse(c)
	}
}

// requiresIdempotency 检查是否需要幂等性处理
func (m *IdempotencyMiddleware) requiresIdempotency(method string) bool {
	for _, requiredMethod := range m.config.RequiredMethods {
		if method == requiredMethod {
			return true
		}
	}
	return false
}

// getIdempotencyKey 获取幂等性键
func (m *IdempotencyMiddleware) getIdempotencyKey(c *gin.Context) string {
	// 从HTTP头获取
	key := c.GetHeader(m.config.IdempotencyKeyHeader)
	if key != "" {
		return key
	}

	// 从查询参数获取
	key = c.Query("idempotency_key")
	if key != "" {
		return key
	}

	// 从用户信息和路径生成默认键
	userID := m.getUserID(c)
	if userID != "" {
		return m.generateDefaultKey(c, userID)
	}

	return ""
}

// buildFullKey 构建完整的键名
func (m *IdempotencyMiddleware) buildFullKey(c *gin.Context, idempotencyKey string) string {
	userID := m.getUserID(c)
	path := c.Request.URL.Path
	method := c.Request.Method

	return fmt.Sprintf("%s:%s:%s:%s", userID, method, path, idempotencyKey)
}

// getUserID 获取用户ID
func (m *IdempotencyMiddleware) getUserID(c *gin.Context) string {
	// 从上下文获取用户ID
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("%v", userID)
	}

	// 从请求头获取
	userID := c.GetHeader("X-User-ID")
	if userID != "" {
		return userID
	}

	// 默认使用IP地址
	return c.ClientIP()
}

// generateDefaultKey 生成默认键
func (m *IdempotencyMiddleware) generateDefaultKey(c *gin.Context, userID string) string {
	path := c.Request.URL.Path
	method := c.Request.Method
	timestamp := time.Now().Format("20060102150405")
	random := generateRandomString(8)

	return fmt.Sprintf("%s:%s:%s:%s:%s", userID, method, path, timestamp, random)
}

// handleDuplicateRequest 处理重复请求
func (m *IdempotencyMiddleware) handleDuplicateRequest(c *gin.Context, key string) {
	// 获取缓存的结果
	result, err := m.manager.GetResult(c.Request.Context(), key)
	if err != nil {
		m.logger.Error("获取幂等性结果失败",
			map[string]interface{}{
				"key":   key,
				"error": err,
			})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取重复请求结果失败",
			"code":  "GET_IDEMPOTENCY_RESULT_FAILED",
		})
		c.Abort()
		return
	}

	if result == nil {
		// 请求正在处理中
		c.JSON(http.StatusAccepted, gin.H{
			"message": "请求正在处理中",
			"code":    "REQUEST_PROCESSING",
		})
		c.Abort()
		return
	}

	// 返回缓存的结果
	m.returnCachedResult(c, result)
	c.Abort()
}

// saveResponse 保存响应结果
func (m *IdempotencyMiddleware) saveResponse(c *gin.Context) {
	key, exists := c.Get("idempotency_key")
	if !exists {
		return
	}

	idempotencyKey := key.(string)

	// 只保存成功的响应
	if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
		// 构建响应数据
		responseData := m.extractResponseData(c)

		// 保存到幂等性管理器
		ttl := m.getDefaultTTL()
		err := m.manager.SaveResult(c.Request.Context(), idempotencyKey, responseData, ttl)
		if err != nil {
			m.logger.Error("保存幂等性结果失败",
				map[string]interface{}{
					"key":   idempotencyKey,
					"error": err,
				})
		}
	}
}

// extractResponseData 提取响应数据
func (m *IdempotencyMiddleware) extractResponseData(c *gin.Context) interface{} {
	// 获取响应数据
	responseData, exists := c.Get("response_data")
	if exists {
		return responseData
	}

	// 如果没有响应数据，构建一个基本的响应结构
	return map[string]interface{}{
		"status":  c.Writer.Status(),
		"message": m.getStatusMessage(c.Writer.Status()),
		"data":    nil, // 可以在这里添加更多数据
	}
}

// returnCachedResult 返回缓存的结果
func (m *IdempotencyMiddleware) returnCachedResult(c *gin.Context, result interface{}) {
	// 尝试将结果转换为JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		m.logger.Error("序列化缓存结果失败",
			map[string]interface{}{
				"error": err,
			})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "返回缓存结果失败",
			"code":  "CACHE_RESULT_SERIALIZE_FAILED",
		})
		return
	}

	// 解析JSON以获取状态码
	var responseMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &responseMap); err == nil {
		if status, ok := responseMap["status"].(float64); ok {
			c.Writer.WriteHeader(int(status))
		}
	}

	// 设置响应头
	c.Header("X-Idempotency-Cached", "true")
	c.Header("X-Idempotency-Timestamp", time.Now().Format(time.RFC3339))

	// 写入响应数据
	c.Data(c.Writer.Status(), "application/json", jsonData)
}

// getDefaultTTL 获取默认TTL
func (m *IdempotencyMiddleware) getDefaultTTL() time.Duration {
	if m.config.DefaultTTL > 0 {
		return m.config.DefaultTTL
	}
	return 24 * time.Hour
}

// logRequestBody 记录请求体
func (m *IdempotencyMiddleware) logRequestBody(c *gin.Context, key string) {
	if !m.config.LogRequestBody {
		return
	}

	// 只记录小请求体
	maxSize := int64(1024) // 1KB
	if c.Request.ContentLength > maxSize {
		m.logger.Debug("请求体过大，跳过记录",
			map[string]interface{}{
				"key":            key,
				"content_length": c.Request.ContentLength,
			})
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		m.logger.Error("获取请求体失败",
			map[string]interface{}{
				"key":   key,
				"error": err,
			})
		return
	}

	m.logger.Debug("记录请求体",
		map[string]interface{}{
			"key":         key,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"body_length": len(body),
			"body":        string(body),
		})
}

// getStatusMessage 根据状态码获取消息
func (m *IdempotencyMiddleware) getStatusMessage(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "Success"
	case status >= 400 && status < 500:
		return "Client Error"
	case status >= 500:
		return "Server Error"
	default:
		return "Unknown"
	}
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
