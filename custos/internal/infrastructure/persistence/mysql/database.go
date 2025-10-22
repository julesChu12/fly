package mysql

import (
	"fmt"

	"github.com/julesChu12/fly/custos/internal/domain/entity"
	moradb "github.com/julesChu12/fly/mora/pkg/db"
	"gorm.io/gorm"
)

// Database wraps mora db.Client
type Database struct {
	client *moradb.Client
}

// NewDatabase creates a new database instance using mora/pkg/db
func NewDatabase(dsn string, debug bool) (*Database, error) {
	logLevel := "warn"
	if debug {
		logLevel = "info"
	}

	cfg := moradb.Config{
		Driver:          "mysql",
		DSN:             dsn,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3600, // 1 hour
		LogLevel:        logLevel,
	}

	client, err := moradb.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Database{client: client}, nil
}

// AutoMigrate automatically migrates database schema
func (d *Database) AutoMigrate() error {
	return d.client.AutoMigrate(
		&entity.Tenant{}, // Add Tenant table
		&entity.User{},
		&entity.Session{},
	)
}

// DB returns the underlying GORM DB instance for backward compatibility
func (d *Database) DB() *gorm.DB {
	return d.client.DB()
}

// Client returns the mora db.Client for advanced usage
func (d *Database) Client() *moradb.Client {
	return d.client
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.client.Close()
}
