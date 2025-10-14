package mora

import (
	"github.com/julesChu12/fly/mora/pkg/logger"
	"go.uber.org/zap"
)

var Log *zap.Logger

// InitLogger initializes the Mora logger
func InitLogger(serviceName string, env string) error {
	cfg := logger.Config{
		Level:       "info",
		Encoding:    "json",
		Development: env == "development",
		OutputPaths: []string{"stdout"},
	}

	if env == "development" {
		cfg.Level = "debug"
		cfg.Encoding = "console"
	}

	l, err := logger.NewLogger(cfg)
	if err != nil {
		return err
	}

	Log = l
	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Log == nil {
		// Fallback to default logger if not initialized
		Log, _ = zap.NewProduction()
	}
	return Log
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
