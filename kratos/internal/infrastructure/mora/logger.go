package mora

import (
	"github.com/julesChu12/fly/mora/pkg/logger"
	"go.uber.org/zap"
)

var Log *logger.Logger

// InitLogger initializes the Mora logger
func InitLogger(serviceName string, env string) error {
	cfg := logger.Config{
		Level:  "info",
		Format: "json",
	}

	if env == "development" {
		cfg.Level = "debug"
		cfg.Format = "console"
	}

	l, err := logger.New(cfg)
	if err != nil {
		return err
	}

	Log = l
	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *logger.Logger {
	if Log == nil {
		// Fallback to default logger if not initialized
		Log = logger.NewDefault()
	}
	return Log
}

// GetZapLogger returns the underlying zap.Logger from the mora logger
func GetZapLogger() *zap.Logger {
	l := GetLogger()
	return l.SugaredLogger.Desugar()
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
