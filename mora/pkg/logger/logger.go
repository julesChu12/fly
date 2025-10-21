package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger represents a logger instance
type Logger struct {
	*zap.SugaredLogger
}

// Config holds the logger configuration
type Config struct {
	Level      string `json:"level" yaml:"level"`           // debug, info, warn, error
	Format     string `json:"format" yaml:"format"`         // json, console
	OutputPath string `json:"output_path" yaml:"outputPath"` // log file path, empty for stdout only

	// File rotation settings (only used when OutputPath is set)
	MaxSize    int  `json:"max_size" yaml:"maxSize"`       // megabytes
	MaxBackups int  `json:"max_backups" yaml:"maxBackups"` // number of old log files to retain
	MaxAge     int  `json:"max_age" yaml:"maxAge"`         // days
	Compress   bool `json:"compress" yaml:"compress"`      // compress old log files

	// Output control
	EnableStdout bool `json:"enable_stdout" yaml:"enableStdout"` // enable stdout output (in addition to file)
	EnableFile   bool `json:"enable_file" yaml:"enableFile"`     // enable file output
}

var defaultLogger *Logger

// New creates a new logger instance with support for multiple outputs
func New(cfg Config) (*Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	// Determine encoder
	var encoder zapcore.Encoder
	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == "console" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Collect all write syncers
	var cores []zapcore.Core

	// Add stdout output if enabled
	if cfg.EnableStdout {
		stdoutCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, stdoutCore)
	}

	// Add file output if enabled
	if cfg.EnableFile && cfg.OutputPath != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.OutputPath,
			MaxSize:    cfg.MaxSize,    // megabytes
			MaxBackups: cfg.MaxBackups, // number of old log files
			MaxAge:     cfg.MaxAge,     // days
			Compress:   cfg.Compress,   // compress old files
			LocalTime:  true,           // use local time for backup filenames
		}

		fileCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(fileWriter),
			level,
		)
		cores = append(cores, fileCore)
	}

	// If no outputs are enabled, default to stdout
	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	// Combine all cores
	core := zapcore.NewTee(cores...)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

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
		Level:        "info",
		Format:       "json",
		EnableStdout: true,
		EnableFile:   false,
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
