package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/julesChu12/fly/mora/pkg/logger"
	migrate "github.com/rubenv/sql-migrate"
)

//go:embed sql-migrate/*.sql
var migrations embed.FS

// MigrationManager handles database migrations using sql-migrate
type MigrationManager struct {
	db     *sql.DB
	logger logger.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB, logger logger.Logger) *MigrationManager {
	return &MigrationManager{
		db:     db,
		logger: logger,
	}
}

// Up applies all pending migrations
func (m *MigrationManager) Up() error {
	migrationSource := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "sql-migrate",
	}

	n, err := migrate.Exec(m.db, "mysql", migrationSource, migrate.Up)
	if err != nil {
		m.logger.Error("Failed to apply migrations", "error", err)
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	m.logger.Info("Applied migrations", "count", n)
	return nil
}

// Down rolls back the last migration
func (m *MigrationManager) Down() error {
	migrationSource := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "sql-migrate",
	}

	n, err := migrate.ExecMax(m.db, "mysql", migrationSource, migrate.Down, 1)
	if err != nil {
		m.logger.Error("Failed to rollback migration", "error", err)
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	m.logger.Info("Rolled back migrations", "count", n)
	return nil
}

// Status returns the current migration status
func (m *MigrationManager) Status() ([]*migrate.MigrationRecord, error) {
	migrationSource := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "sql-migrate",
	}

	records, err := migrate.GetMigrationRecords(m.db, "mysql")
	if err != nil {
		m.logger.Error("Failed to get migration status", "error", err)
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}

	planned, err := migrationSource.FindMigrations()
	if err != nil {
		m.logger.Error("Failed to find migrations", "error", err)
		return nil, fmt.Errorf("failed to find migrations: %w", err)
	}

	for _, migration := range planned {
		found := false
		for _, record := range records {
			if record.Id == migration.Id {
				found = true
				break
			}
		}
		if !found {
			records = append(records, &migrate.MigrationRecord{
				Id:        migration.Id,
				AppliedAt: time.Time{},
			})
		}
	}

	return records, nil
}