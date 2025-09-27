# SQL-Migrate Database Migrations

This project uses [sql-migrate](https://github.com/rubenv/sql-migrate) for database schema management.

## Migration Files Structure

```
configs/migrations/sql-migrate/
├── 20240101_001_create_users_table.sql
├── 20240101_002_create_user_profiles_table.sql
├── 20240101_003_create_user_oauth_table.sql
├── 20240101_004_create_refresh_tokens_table.sql
├── 20240101_005_create_sessions_table.sql
├── 20240101_006_create_jwk_keys_table.sql
└── 20240101_007_create_casbin_rule_table.sql
```

## Usage

### Using the Migration Script

```bash
# Apply all pending migrations
./scripts/migrate.sh up

# Rollback the last migration
./scripts/migrate.sh down

# Check migration status
./scripts/migrate.sh status

# Create a new migration
./scripts/migrate.sh new development create_new_table
```

### Using Makefile

```bash
# Check migration status
make migrate
```

### Programmatic Usage

Migrations are automatically applied when the application starts:

```go
migrationManager := migrate.NewMigrationManager(sqlDB, logger)
if err := migrationManager.Up(); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
}
```

## Configuration

Database configuration is managed in `configs/dbconfig.yml`:

```yaml
development:
  dialect: mysql
  datasource: "root:password@tcp(localhost:3306)/custos_dev?charset=utf8mb4&parseTime=True&loc=Local"
  dir: configs/migrations/sql-migrate
```

## Migration File Format

Each migration file follows the sql-migrate format:

```sql
-- +migrate Up
CREATE TABLE example (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL
);

-- +migrate Down
DROP TABLE example;
```