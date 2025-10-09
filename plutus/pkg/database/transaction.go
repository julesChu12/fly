package database

import (
	"context"

	"gorm.io/gorm"
)

// TransactionFunc is a function type for database transactions
type TransactionFunc func(tx *gorm.DB) error

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func WithTransaction(ctx context.Context, db *gorm.DB, fn TransactionFunc) error {
	return db.WithContext(ctx).Transaction(fn)
}

// GetDB returns the underlying *gorm.DB from context or the provided db
// This is useful for methods that may be called within a transaction or standalone
func GetDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok && tx != nil {
		return tx
	}
	return db
}
