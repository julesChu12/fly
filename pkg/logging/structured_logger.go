package logging

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var levelNames = map[LogLevel]string{
	LevelTrace: "TRACE",
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

// LogEntry 结构化日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
}

// StructuredLogger 结构化日志器
type StructuredLogger struct {
	baseLogger *logger.Logger
	level      LogLevel
	output     LogOutput
	mu         sync.RWMutex
	hooks      []LogHook

	// 性能统计
	totalLogs     int64
	errorCount    int64
	warnCount     int64
	lastResetTime time.Time
}

// LogOutput 日志输出接口
type LogOutput interface {
	Write(entry *LogEntry) error
	Close() error
}

// LogHook 日志钩子接口
type LogHook interface {
	Fire(entry *LogEntry) error
	Levels() []LogLevel
}

// ContextKey 上下文键类型
type ContextKey string

const (
	TraceIDKey ContextKey = "trace_id"
	SpanIDKey  ContextKey = "span_id"
	UserIDKey  ContextKey = "user_id"
	RequestIDKey ContextKey = "request_id"
)

// NewStructuredLogger 创建结构化日志器
func NewStructuredLogger(baseLogger *logger.Logger) *StructuredLogger {
	return &StructuredLogger{
		baseLogger:    baseLogger,
		level:         LevelInfo,
		output:        &ConsoleOutput{},
		hooks:         make([]LogHook, 0),
		lastResetTime: time.Now(),
	}
}

// SetLevel 设置日志级别
func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// AddHook 添加日志钩子
func (l *StructuredLogger) AddHook(hook LogHook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook)
}

// SetOutput 设置日志输出
func (l *StructuredLogger) SetOutput(output LogOutput) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// Trace 记录追踪级别日志
func (l *StructuredLogger) Trace(message string, fields ...map[string]interface{}) {
	l.log(LevelTrace, message, fields...)
}

// Debug 记录调试级别日志
func (l *StructuredLogger) Debug(message string, fields ...map[string]interface{}) {
	l.log(LevelDebug, message, fields...)
}

// Info 记录信息级别日志
func (l *StructuredLogger) Info(message string, fields ...map[string]interface{}) {
	l.log(LevelInfo, message, fields...)
}

// Warn 记录警告级别日志
func (l *StructuredLogger) Warn(message string, fields ...map[string]interface{}) {
	l.log(LevelWarn, message, fields...)
	atomic.AddInt64(&l.warnCount, 1)
}

// Error 记录错误级别日志
func (l *StructuredLogger) Error(message string, fields ...map[string]interface{}) {
	l.log(LevelError, message, fields...)
	atomic.AddInt64(&l.errorCount, 1)
}

// Fatal 记录致命错误日志
func (l *StructuredLogger) Fatal(message string, fields ...map[string]interface{}) {
	l.log(LevelFatal, message, fields...)
	os.Exit(1)
}

// WithContext 从上下文中提取信息记录日志
func (l *StructuredLogger) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: l,
		ctx:    ctx,
	}
}

// WithFields 添加字段
func (l *StructuredLogger) WithFields(fields map[string]interface{}) *FieldLogger {
	return &FieldLogger{
		logger: l,
		fields: fields,
	}
}

// WithField 添加单个字段
func (l *StructuredLogger) WithField(key string, value interface{}) *FieldLogger {
	return l.WithFields(map[string]interface{}{key: value})
}

// log 核心日志记录方法
func (l *StructuredLogger) log(level LogLevel, message string, fields ...map[string]interface{}) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	output := l.output
	hooks := make([]LogHook, len(l.hooks))
	copy(hooks, l.hooks)
	l.mu.RUnlock()

	// 创建日志条目
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    make(map[string]interface{}),
		Caller:    getCaller(),
	}

	// 合并字段
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entry.Fields[k] = v
		}
	}

	// 添加性��统计字段
	entry.Fields["log_stats"] = map[string]interface{}{
		"total_logs":   atomic.LoadInt64(&l.totalLogs),
		"error_count":  atomic.LoadInt64(&l.errorCount),
		"warn_count":   atomic.LoadInt64(&l.warnCount),
		"uptime":       time.Since(l.lastResetTime).String(),
	}

	atomic.AddInt64(&l.totalLogs, 1)

	// 执行钩子
	for _, hook := range hooks {
		if shouldFireHook(hook, level) {
			if err := hook.Fire(entry); err != nil {
				// 钩子执行失败不应该影响主日志记录
				fmt.Printf("Log hook error: %v\n", err)
			}
		}
	}

	// 输出日志
	if output != nil {
		if err := output.Write(entry); err != nil {
			fmt.Printf("Log output error: %v\n", err)
		}
	}

	// 同时输出到mora logger（保持兼容性）
	l.outputToMora(entry)
}

// outputToMora 输出到mora logger
func (l *StructuredLogger) outputToMora(entry *LogEntry) {
	if l.baseLogger == nil {
		return
	}

	fields := make(map[string]interface{})
	for k, v := range entry.Fields {
		fields[k] = v
	}
	fields["caller"] = entry.Caller

	switch entry.Level {
	case LevelTrace, LevelDebug:
		l.baseLogger.Debug(entry.Message, fields)
	case LevelInfo:
		l.baseLogger.Info(entry.Message, fields)
	case LevelWarn:
		l.baseLogger.Warn(entry.Message, fields)
	case LevelError, LevelFatal:
		l.baseLogger.Error(entry.Message, fields)
	}
}

// GetStats 获取日志统计
func (l *StructuredLogger) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_logs":     atomic.LoadInt64(&l.totalLogs),
		"error_count":    atomic.LoadInt64(&l.errorCount),
		"warn_count":     atomic.LoadInt64(&l.warnCount),
		"uptime":         time.Since(l.lastResetTime).String(),
		"logs_per_second": float64(atomic.LoadInt64(&l.totalLogs)) / time.Since(l.lastResetTime).Seconds(),
	}
}

// ResetStats 重置统计
func (l *StructuredLogger) ResetStats() {
	atomic.StoreInt64(&l.totalLogs, 0)
	atomic.StoreInt64(&l.errorCount, 0)
	atomic.StoreInt64(&l.warnCount, 0)
	l.lastResetTime = time.Now()
}

// getCaller 获取调用者信息
func getCaller() string {
	_, file, line, ok := runtime.Caller(4) // 跳过4层调用栈
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// shouldFireHook 检查是否应该执行钩子
func shouldFireHook(hook LogHook, level LogLevel) bool {
	levels := hook.Levels()
	for _, l := range levels {
		if l == level {
			return true
		}
	}
	return false
}

// ContextLogger 带上下文的日志器
type ContextLogger struct {
	logger *StructuredLogger
	ctx    context.Context
}

// Info 记录信息级别日志（带上下文）
func (cl *ContextLogger) Info(message string, fields ...map[string]interface{}) {
	entryFields := make(map[string]interface{})

	// 从上下文中提取信息
	if traceID := cl.ctx.Value(TraceIDKey); traceID != nil {
		entryFields["trace_id"] = traceID
	}
	if spanID := cl.ctx.Value(SpanIDKey); spanID != nil {
		entryFields["span_id"] = spanID
	}
	if userID := cl.ctx.Value(UserIDKey); userID != nil {
		entryFields["user_id"] = userID
	}
	if requestID := cl.ctx.Value(RequestIDKey); requestID != nil {
		entryFields["request_id"] = requestID
	}

	// 合并其他字段
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			entryFields[k] = v
		}
	}

	cl.logger.Info(message, entryFields)
}

// Debug 记录调试级别日志（带上下文）
func (cl *ContextLogger) Debug(message string, fields ...map[string]interface{}) {
	cl.logger.Debug(message, fields...)
}

// Warn 记录警告级别日志（带上下文）
func (cl *ContextLogger) Warn(message string, fields ...map[string]interface{}) {
	cl.logger.Warn(message, fields...)
}

// Error 记录错误级别日志（带上下文）
func (cl *ContextLogger) Error(message string, fields ...map[string]interface{}) {
	cl.logger.Error(message, fields...)
}

// FieldLogger 带固定字段的日志器
type FieldLogger struct {
	logger *StructuredLogger
	fields map[string]interface{}
}

// Info 记录信息级别日志（带固定字段）
func (fl *FieldLogger) Info(message string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})

	// 添加固定字段
	for k, v := range fl.fields {
		allFields[k] = v
	}

	// 合并临时字段
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}

	fl.logger.Info(message, allFields)
}

// Debug 记录调试级别日志（带固定字段）
func (fl *FieldLogger) Debug(message string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})

	for k, v := range fl.fields {
		allFields[k] = v
	}

	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}

	fl.logger.Debug(message, allFields)
}

// Warn 记录警告级别日志（带固定字段）
func (fl *FieldLogger) Warn(message string, fields ...map[string]interface{}) {
	fl.logger.Warn(message, append([]map[string]interface{}{fl.fields}, fields...)...)
}

// Error 记录错误级别日志（带固定字段）
func (fl *FieldLogger) Error(message string, fields ...map[string]interface{}) {
	fl.logger.Error(message, append([]map[string]interface{}{fl.fields}, fields...)...)
}

// WithFields 添加更多字段
func (fl *FieldLogger) WithFields(fields map[string]interface{}) *FieldLogger {
	newFields := make(map[string]interface{})

	// 复制现有字段
	for k, v := range fl.fields {
		newFields[k] = v
	}

	// 添加新字段
	for k, v := range fields {
		newFields[k] = v
	}

	return &FieldLogger{
		logger: fl.logger,
		fields: newFields,
	}
}

// WithField 添加单个字段
func (fl *FieldLogger) WithField(key string, value interface{}) *FieldLogger {
	return fl.WithFields(map[string]interface{}{key: value})
}

// ConsoleOutput 控制台输出实现
type ConsoleOutput struct{}

// Write 写入日志条目到控制台
func (c *ConsoleOutput) Write(entry *LogEntry) error {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
	levelName := levelNames[entry.Level]

	// 基础格式
	fmt.Printf("[%s] %s %s", timestamp, levelName, entry.Message)

	// 添加调用者信息
	if entry.Caller != "" {
		fmt.Printf(" (%s)", entry.Caller)
	}

	// 添加字段信息
	if len(entry.Fields) > 0 {
		fmt.Printf(" %v", entry.Fields)
	}

	fmt.Println()
	return nil
}

// Close 关闭输出
func (c *ConsoleOutput) Close() error {
	return nil
}

// 全局结构化日志器实例
var GlobalStructuredLogger *StructuredLogger

// InitStructuredLogger 初始化全局结构化日志器
func InitStructuredLogger(baseLogger *logger.Logger) {
	GlobalStructuredLogger = NewStructuredLogger(baseLogger)
}

// Info 全局Info方法
func Info(message string, fields ...map[string]interface{}) {
	if GlobalStructuredLogger != nil {
		GlobalStructuredLogger.Info(message, fields...)
	}
}

// Error 全局Error方法
func Error(message string, fields ...map[string]interface{}) {
	if GlobalStructuredLogger != nil {
		GlobalStructuredLogger.Error(message, fields...)
	}
}

// Debug 全局Debug方法
func Debug(message string, fields ...map[string]interface{}) {
	if GlobalStructuredLogger != nil {
		GlobalStructuredLogger.Debug(message, fields...)
	}
}

// Warn 全局Warn方法
func Warn(message string, fields ...map[string]interface{}) {
	if GlobalStructuredLogger != nil {
		GlobalStructuredLogger.Warn(message, fields...)
	}
}