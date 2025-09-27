package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger represents a logger instance
type Logger struct {
	*zap.SugaredLogger
}

// Config holds the logger configuration
type Config struct {
	Level  string `json:"level" yaml:"level"`   // debug, info, warn, error
	Format string `json:"format" yaml:"format"` // json, console
}

var defaultLogger *Logger

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	var config zap.Config
	if cfg.Format == "console" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = zap.NewAtomicLevelAt(level)

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}, nil
}

// NewDefault creates a logger with default configuration
func NewDefault() *Logger {
	if defaultLogger != nil {
		return defaultLogger
	}

	cfg := Config{
		Level:  "info",
		Format: "json",
	}

	if os.Getenv("ENV") == "development" {
		cfg.Format = "console"
		cfg.Level = "debug"
	}

	logger, err := New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create default logger: %v", err))
	}

	defaultLogger = logger
	return defaultLogger
}

// WithTraceID adds a trace ID to the logger context
func (l *Logger) WithTraceID(traceID string) *Logger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.With("trace_id", traceID),
	}
}

// WithContext extracts trace ID and span ID from context and adds them to logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}

	logger := l
	if traceID := GetTraceIDFromContext(ctx); traceID != "" {
		logger = logger.WithTraceID(traceID)
	}
	if spanID := GetSpanIDFromContext(ctx); spanID != "" {
		logger = &Logger{
			SugaredLogger: logger.SugaredLogger.With("span_id", spanID),
		}
	}
	return logger
}

// WithCtx is an alias for WithContext for convenience
func (l *Logger) WithCtx(ctx context.Context) *Logger {
	return l.WithContext(ctx)
}

// WithFields adds structured fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		SugaredLogger: l.SugaredLogger.With(args...),
	}
}

// Global logger functions using default logger

// WithCtx creates a logger with trace context from the default logger
func WithCtx(ctx context.Context) *Logger {
	return NewDefault().WithCtx(ctx)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	NewDefault().Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(template string, args ...interface{}) {
	NewDefault().Debugf(template, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	NewDefault().Info(args...)
}

// Infof logs a formatted info message
func Infof(template string, args ...interface{}) {
	NewDefault().Infof(template, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	NewDefault().Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(template string, args ...interface{}) {
	NewDefault().Warnf(template, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	NewDefault().Error(args...)
}

// Errorf logs a formatted error message
func Errorf(template string, args ...interface{}) {
	NewDefault().Errorf(template, args...)
}

// Fatal logs a fatal message and calls os.Exit(1)
func Fatal(args ...interface{}) {
	NewDefault().Fatal(args...)
}

// Fatalf logs a formatted fatal message and calls os.Exit(1)
func Fatalf(template string, args ...interface{}) {
	NewDefault().Fatalf(template, args...)
}
