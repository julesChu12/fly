package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	Driver          string `json:"driver" yaml:"driver" env:"DRIVER"` // mysql, postgres, sqlite
	DSN             string `json:"dsn" yaml:"dsn" env:"DSN"`          // Data Source Name
	MaxOpenConns    int    `json:"max_open_conns" yaml:"max_open_conns" env:"MAX_OPEN_CONNS"`
	MaxIdleConns    int    `json:"max_idle_conns" yaml:"max_idle_conns" env:"MAX_IDLE_CONNS"`
	ConnMaxLifetime int    `json:"conn_max_lifetime" yaml:"conn_max_lifetime" env:"CONN_MAX_LIFETIME"` // seconds
	LogLevel        string `json:"log_level" yaml:"log_level" env:"LOG_LEVEL"`                         // silent, error, warn, info
}

// DefaultConfig returns default database configuration
func DefaultConfig() Config {
	return Config{
		Driver:          "mysql",
		DSN:             "",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3600, // 1 hour
		LogLevel:        "warn",
	}
}

// Client wraps GORM database instance
type Client struct {
	db *gorm.DB
}

// New creates a new database client using GORM
func New(cfg Config) (*Client, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	case "sqlite":
		dialector = sqlite.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// Configure GORM logger
	var logLevel logger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Warn
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return &Client{db: db}, nil
}

// DB returns the underlying GORM DB instance
func (c *Client) DB() *gorm.DB {
	return c.db
}

// Close closes the database connection
func (c *Client) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping tests the database connection
func (c *Client) Ping() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats returns database connection pool statistics
func (c *Client) Stats() sql.DBStats {
	sqlDB, err := c.db.DB()
	if err != nil {
		return sql.DBStats{}
	}
	return sqlDB.Stats()
}

// Transaction Management

// Transaction represents a database transaction
type Transaction struct {
	tx *gorm.DB
}

// Begin starts a new transaction
func (c *Client) Begin() *Transaction {
	return &Transaction{tx: c.db.Begin()}
}

// BeginTx starts a new transaction with context and options
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) *Transaction {
	return &Transaction{tx: c.db.WithContext(ctx).Begin(opts)}
}

// DB returns the transaction's GORM DB instance
func (tx *Transaction) DB() *gorm.DB {
	return tx.tx
}

// Commit commits the transaction
func (tx *Transaction) Commit() error {
	return tx.tx.Commit().Error
}

// Rollback rolls back the transaction
func (tx *Transaction) Rollback() error {
	return tx.tx.Rollback().Error
}

// WithTransaction executes a function within a transaction
func (c *Client) WithTransaction(ctx context.Context, fn func(*Transaction) error) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transaction := &Transaction{tx: tx}
		return fn(transaction)
	})
}

// WithTransactionTx executes a function within a transaction with options
func (c *Client) WithTransactionTx(ctx context.Context, opts *sql.TxOptions, fn func(*Transaction) error) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		transaction := &Transaction{tx: tx}
		return fn(transaction)
	}, opts)
}

// CRUD Operations Helpers

// Create creates a new record
func (c *Client) Create(ctx context.Context, value interface{}) error {
	return c.db.WithContext(ctx).Create(value).Error
}

// Save saves/updates a record
func (c *Client) Save(ctx context.Context, value interface{}) error {
	return c.db.WithContext(ctx).Save(value).Error
}

// First finds the first record matching the query
func (c *Client) First(ctx context.Context, dest interface{}, conds ...interface{}) error {
	return c.db.WithContext(ctx).First(dest, conds...).Error
}

// Find finds all records matching the query
func (c *Client) Find(ctx context.Context, dest interface{}, conds ...interface{}) error {
	return c.db.WithContext(ctx).Find(dest, conds...).Error
}

// Update updates records with conditions
func (c *Client) Update(ctx context.Context, column string, value interface{}, conds ...interface{}) error {
	return c.db.WithContext(ctx).Model(nil).Where(conds[0], conds[1:]...).Update(column, value).Error
}

// Updates updates multiple columns with conditions
func (c *Client) Updates(ctx context.Context, values interface{}, conds ...interface{}) error {
	return c.db.WithContext(ctx).Model(nil).Where(conds[0], conds[1:]...).Updates(values).Error
}

// Delete deletes records with conditions
func (c *Client) Delete(ctx context.Context, value interface{}, conds ...interface{}) error {
	return c.db.WithContext(ctx).Delete(value, conds...).Error
}

// Count counts records matching the conditions
func (c *Client) Count(ctx context.Context, model interface{}, count *int64, conds ...interface{}) error {
	query := c.db.WithContext(ctx).Model(model)
	if len(conds) > 0 {
		query = query.Where(conds[0], conds[1:]...)
	}
	return query.Count(count).Error
}

// Exists checks if records exist with the given conditions
func (c *Client) Exists(ctx context.Context, model interface{}, conds ...interface{}) (bool, error) {
	var count int64
	err := c.Count(ctx, model, &count, conds...)
	return count > 0, err
}

// Paginate performs pagination query
func (c *Client) Paginate(ctx context.Context, dest interface{}, page, pageSize int, conds ...interface{}) error {
	offset := (page - 1) * pageSize
	query := c.db.WithContext(ctx)
	if len(conds) > 0 {
		query = query.Where(conds[0], conds[1:]...)
	}
	return query.Offset(offset).Limit(pageSize).Find(dest).Error
}

// PaginateResult represents pagination result
type PaginateResult struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// PaginateWithCount performs pagination with total count
func (c *Client) PaginateWithCount(ctx context.Context, model interface{}, dest interface{}, page, pageSize int, conds ...interface{}) (*PaginateResult, error) {
	var total int64

	// Count total records
	countQuery := c.db.WithContext(ctx).Model(model)
	if len(conds) > 0 {
		countQuery = countQuery.Where(conds[0], conds[1:]...)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Get paginated data
	offset := (page - 1) * pageSize
	dataQuery := c.db.WithContext(ctx).Model(model)
	if len(conds) > 0 {
		dataQuery = dataQuery.Where(conds[0], conds[1:]...)
	}
	if err := dataQuery.Offset(offset).Limit(pageSize).Find(dest).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch paginated data: %w", err)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &PaginateResult{
		Data:       dest,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Migration helpers

// AutoMigrate runs automatic migration for given models
func (c *Client) AutoMigrate(models ...interface{}) error {
	return c.db.AutoMigrate(models...)
}

// Raw executes raw SQL
func (c *Client) Raw(ctx context.Context, sql string, values ...interface{}) *gorm.DB {
	return c.db.WithContext(ctx).Raw(sql, values...)
}

// Exec executes raw SQL
func (c *Client) Exec(ctx context.Context, sql string, values ...interface{}) error {
	return c.db.WithContext(ctx).Exec(sql, values...).Error
}
