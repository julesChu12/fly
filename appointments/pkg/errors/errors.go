package errors

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// ErrorCode 错误码类型
type ErrorCode string

// 预定义错误码
const (
	// 通用错误码
	CodeInternalError    ErrorCode = "INTERNAL_ERROR"
	CodeInvalidArgument  ErrorCode = "INVALID_ARGUMENT"
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	CodeTimeout          ErrorCode = "TIMEOUT"
	CodeRateLimited      ErrorCode = "RATE_LIMITED"

	// 业务错误码
	CodeBusinessError       ErrorCode = "BUSINESS_ERROR"
	CodeValidationError      ErrorCode = "VALIDATION_ERROR"
	CodeConflictError       ErrorCode = "CONFLICT_ERROR"
	CodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"

	// 系统错误码
	CodeDatabaseError     ErrorCode = "DATABASE_ERROR"
	CodeNetworkError      ErrorCode = "NETWORK_ERROR"
	CodeAuthenticationError ErrorCode = "AUTHENTICATION_ERROR"
	CodeAuthorizationError  ErrorCode = "AUTHORIZATION_ERROR"
)

// ErrorLevel 错误级别
type ErrorLevel string

const (
	LevelFatal ErrorLevel = "FATAL"
	LevelError ErrorLevel = "ERROR"
	LevelWarn  ErrorLevel = "WARN"
	LevelInfo  ErrorLevel = "INFO"
)

// AppError 应用错误结构
type AppError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Level       ErrorLevel             `json:"level"`
	Cause       error                  `json:"-"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Stack       []string               `json:"stack,omitempty"`
	Retryable   bool                   `json:"retryable"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	Delay       time.Duration `json:"delay"`
	Backoff     BackoffType   `json:"backoff"`
}

// BackoffType 退避类型
type BackoffType string

const (
	BackoffFixed     BackoffType = "fixed"
	BackoffLinear    BackoffType = "linear"
	BackoffExponential BackoffType = "exponential"
)

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Context != nil && len(e.Context) > 0 {
		return fmt.Sprintf("[%s] %s (context: %v)", e.Code, e.Message, e.Context)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is 支持errors.Is
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return false
}

// NewAppError 创建新的应用错误
func NewAppError(code ErrorCode, message string, level ErrorLevel) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
		Retryable: false,
		Stack:     captureStack(),
	}
}

// WithCause 添加原因错误
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithContext 添加上下文信息
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithRetryable 设置可重试
func (e *AppError) WithRetryable(retryable bool) *AppError {
	e.Retryable = retryable
	return e
}

// WithRetryPolicy 设置重试策略
func (e *AppError) WithRetryPolicy(maxAttempts int, delay time.Duration, backoff BackoffType) *AppError {
	e.Retryable = true
	e.RetryPolicy = &RetryPolicy{
		MaxAttempts: maxAttempts,
		Delay:       delay,
		Backoff:     backoff,
	}
	return e
}

// Log 记录错误日志
func (e *AppError) Log(log *logger.Logger) {
	fields := map[string]interface{}{
		"error_code":    e.Code,
		"error_message": e.Message,
		"error_level":   e.Level,
		"timestamp":     e.Timestamp,
		"retryable":     e.Retryable,
	}

	// 添加上下文信息
	for k, v := range e.Context {
		fields[k] = v
	}

	// 添加堆栈信息
	if len(e.Stack) > 0 {
		fields["stack"] = strings.Join(e.Stack, "\n")
	}

	// 添加原因错误
	if e.Cause != nil {
		fields["cause"] = e.Cause.Error()
	}

	// 根据级别记录日志
	switch e.Level {
	case LevelFatal:
		log.Fatal("应用错误", fields)
	case LevelError:
		log.Error("应用错误", fields)
	case LevelWarn:
		log.Warn("应用错误", fields)
	case LevelInfo:
		log.Info("��用错误", fields)
	default:
		log.Error("未知级别错误", fields)
	}
}

// captureStack 捕获调用堆栈
func captureStack() []string {
	const maxStackDepth = 10
	var stack []string

	for i := 2; i < maxStackDepth+2; i++ { // 跳过captureStack和调用函数
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// 格式化: function (file:line)
		stack = append(stack, fmt.Sprintf("%s (%s:%d)", fn.Name(), file, line))
	}

	return stack
}

// 预定义错误创建函数

// InternalError 内部错误
func InternalError(message string) *AppError {
	return NewAppError(CodeInternalError, message, LevelError)
}

// InvalidArgument 参数错误
func InvalidArgument(message string) *AppError {
	return NewAppError(CodeInvalidArgument, message, LevelWarn)
}

// NotFound 未找到错误
func NotFound(resource string, id interface{}) *AppError {
	return NewAppError(CodeNotFound, fmt.Sprintf("%s not found: %v", resource, id), LevelWarn).
		WithRetryable(false)
}

// Timeout 超时错误
func Timeout(operation string, timeout time.Duration) *AppError {
	return NewAppError(CodeTimeout, fmt.Sprintf("%s timeout after %v", operation, timeout), LevelWarn).
		WithRetryable(true).
		WithRetryPolicy(3, 1*time.Second, BackoffExponential)
}

// DatabaseError 数据库错误
func DatabaseError(operation string, cause error) *AppError {
	return NewAppError(CodeDatabaseError, fmt.Sprintf("Database %s failed", operation), LevelError).
		WithCause(cause).
		WithRetryable(true).
		WithRetryPolicy(3, 2*time.Second, BackoffExponential).
		WithContext("operation", operation)
}

// ValidationError 验证错误
func ValidationError(field string, value interface{}, reason string) *AppError {
	return NewAppError(CodeValidationError, fmt.Sprintf("Validation failed for %s", field), LevelWarn).
		WithContext("field", field).
		WithContext("value", value).
		WithContext("reason", reason).
		WithRetryable(false)
}

// ExternalServiceError 外部服务错误
func ExternalServiceError(service string, operation string, cause error) *AppError {
	return NewAppError(CodeExternalServiceError, fmt.Sprintf("External service %s %s failed", service, operation), LevelError).
		WithCause(cause).
		WithContext("service", service).
		WithContext("operation", operation).
		WithRetryable(true).
		WithRetryPolicy(5, 1*time.Second, BackoffExponential)
}

// ConflictError 冲突错误
func ConflictError(resource string, reason string) *AppError {
	return NewAppError(CodeConflictError, fmt.Sprintf("Conflict in %s: %s", resource, reason), LevelWarn).
		WithContext("resource", resource).
		WithContext("reason", reason).
		WithRetryable(true).
		WithRetryPolicy(2, 500*time.Millisecond, BackoffLinear)
}

// BusinessError 业务错误
func BusinessError(message string) *AppError {
	return NewAppError(CodeBusinessError, message, LevelWarn).
		WithRetryable(false)
}

// RateLimited 限流错误
func RateLimited(limit int, window time.Duration) *AppError {
	return NewAppError(CodeRateLimited, fmt.Sprintf("Rate limited: %d requests per %v", limit, window), LevelWarn).
		WithRetryable(true).
		WithRetryPolicy(3, 2*time.Second, BackoffExponential).
		WithContext("limit", limit).
		WithContext("window", window)
}

// PermissionDenied 权限拒绝错误
func PermissionDenied(resource string, action string) *AppError {
	return NewAppError(CodePermissionDenied, fmt.Sprintf("Permission denied for %s on %s", action, resource), LevelWarn).
		WithRetryable(false).
		WithContext("resource", resource).
		WithContext("action", action)
}

// NetworkError 网络错误
func NetworkError(operation string, cause error) *AppError {
	return NewAppError(CodeNetworkError, fmt.Sprintf("Network %s failed", operation), LevelError).
		WithCause(cause).
		WithRetryable(true).
		WithRetryPolicy(5, 1*time.Second, BackoffExponential).
		WithContext("operation", operation)
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retryable
	}
	return false
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) ErrorCode {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return CodeInternalError
}

// WrapError 包装普通错误为应用错误
func WrapError(err error, code ErrorCode, message string) *AppError {
	appErr := NewAppError(code, message, LevelError)
	return appErr.WithCause(err)
}