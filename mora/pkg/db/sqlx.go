package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// SQLXClient wraps sqlx database instance
type SQLXClient struct {
	db *sqlx.DB
}

// NewSQLX creates a new database client using sqlx
func NewSQLX(cfg Config) (*SQLXClient, error) {
	db, err := sqlx.Connect(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return &SQLXClient{db: db}, nil
}

// DB returns the underlying sqlx DB instance
func (c *SQLXClient) DB() *sqlx.DB {
	return c.db
}

// Close closes the database connection
func (c *SQLXClient) Close() error {
	return c.db.Close()
}

// Ping tests the database connection
func (c *SQLXClient) Ping() error {
	return c.db.Ping()
}

// Stats returns database connection pool statistics
func (c *SQLXClient) Stats() sql.DBStats {
	return c.db.Stats()
}

// SQLXTransaction represents a database transaction with sqlx
type SQLXTransaction struct {
	tx *sqlx.Tx
}

// Begin starts a new transaction
func (c *SQLXClient) Begin() (*SQLXTransaction, error) {
	tx, err := c.db.Beginx()
	if err != nil {
		return nil, err
	}
	return &SQLXTransaction{tx: tx}, nil
}

// BeginTx starts a new transaction with context and options
func (c *SQLXClient) BeginTx(ctx context.Context, opts *sql.TxOptions) (*SQLXTransaction, error) {
	tx, err := c.db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &SQLXTransaction{tx: tx}, nil
}

// Tx returns the transaction's sqlx Tx instance
func (tx *SQLXTransaction) Tx() *sqlx.Tx {
	return tx.tx
}

// Commit commits the transaction
func (tx *SQLXTransaction) Commit() error {
	return tx.tx.Commit()
}

// Rollback rolls back the transaction
func (tx *SQLXTransaction) Rollback() error {
	return tx.tx.Rollback()
}

// WithTransaction executes a function within a transaction
func (c *SQLXClient) WithTransaction(ctx context.Context, fn func(*SQLXTransaction) error) error {
	tx, err := c.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rollbackErr)
		}
		return err
	}

	return tx.Commit()
}

// Query Operations

// Get gets a single record into dest
func (c *SQLXClient) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return c.db.GetContext(ctx, dest, query, args...)
}

// Select gets multiple records into dest
func (c *SQLXClient) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return c.db.SelectContext(ctx, dest, query, args...)
}

// Exec executes a query without returning any rows
func (c *SQLXClient) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (c *SQLXClient) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return c.db.QueryxContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (c *SQLXClient) QueryRow(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return c.db.QueryRowxContext(ctx, query, args...)
}

// NamedExec executes a named query
func (c *SQLXClient) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return c.db.NamedExecContext(ctx, query, arg)
}

// NamedQuery executes a named query that returns rows
func (c *SQLXClient) NamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return c.db.NamedQueryContext(ctx, query, arg)
}

// Prepared Statements

// PreparedStatement wraps a prepared statement
type PreparedStatement struct {
	stmt *sqlx.Stmt
}

// Prepare creates a prepared statement
func (c *SQLXClient) Prepare(ctx context.Context, query string) (*PreparedStatement, error) {
	stmt, err := c.db.PreparexContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &PreparedStatement{stmt: stmt}, nil
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	return ps.stmt.Close()
}

// Exec executes the prepared statement
func (ps *PreparedStatement) Exec(ctx context.Context, args ...interface{}) (sql.Result, error) {
	return ps.stmt.ExecContext(ctx, args...)
}

// Get executes the prepared statement and scans the result into dest
func (ps *PreparedStatement) Get(ctx context.Context, dest interface{}, args ...interface{}) error {
	return ps.stmt.GetContext(ctx, dest, args...)
}

// Select executes the prepared statement and scans the results into dest
func (ps *PreparedStatement) Select(ctx context.Context, dest interface{}, args ...interface{}) error {
	return ps.stmt.SelectContext(ctx, dest, args...)
}

// Query executes the prepared statement and returns rows
func (ps *PreparedStatement) Query(ctx context.Context, args ...interface{}) (*sqlx.Rows, error) {
	return ps.stmt.QueryxContext(ctx, args...)
}

// QueryRow executes the prepared statement and returns a single row
func (ps *PreparedStatement) QueryRow(ctx context.Context, args ...interface{}) *sqlx.Row {
	return ps.stmt.QueryRowxContext(ctx, args...)
}

// Transaction Query Operations

// Get gets a single record into dest within transaction
func (tx *SQLXTransaction) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return tx.tx.GetContext(ctx, dest, query, args...)
}

// Select gets multiple records into dest within transaction
func (tx *SQLXTransaction) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return tx.tx.SelectContext(ctx, dest, query, args...)
}

// Exec executes a query without returning any rows within transaction
func (tx *SQLXTransaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return tx.tx.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows within transaction
func (tx *SQLXTransaction) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return tx.tx.QueryxContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row within transaction
func (tx *SQLXTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return tx.tx.QueryRowxContext(ctx, query, args...)
}

// NamedExec executes a named query within transaction
func (tx *SQLXTransaction) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return tx.tx.NamedExecContext(ctx, query, arg)
}

// NamedQuery executes a named query that returns rows within transaction
func (tx *SQLXTransaction) NamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return tx.tx.NamedQuery(query, arg)
}

// Helper Functions

// In creates an IN clause for the given slice
func In(query string, args ...interface{}) (string, []interface{}, error) {
	return sqlx.In(query, args...)
}

// Named replaces named parameters in a query
func Named(query string, arg interface{}) (string, []interface{}, error) {
	return sqlx.Named(query, arg)
}
