#!/bin/bash

# sql-migrate migration management script

set -e

COMMAND=$1
ENVIRONMENT=${2:-development}

if [ -z "$COMMAND" ]; then
    echo "Usage: $0 [up|down|status|new] [environment]"
    echo ""
    echo "Commands:"
    echo "  up      - Apply all pending migrations"
    echo "  down    - Rollback the last migration"
    echo "  status  - Show migration status"
    echo "  new     - Create a new migration file"
    echo ""
    echo "Environments: development, test, production"
    exit 1
fi

# Check if sql-migrate is installed
if ! command -v sql-migrate &> /dev/null; then
    echo "sql-migrate is not installed. Installing..."
    go install github.com/rubenv/sql-migrate/...@latest
fi

CONFIG_FILE="configs/dbconfig.yml"

case "$COMMAND" in
    "up")
        echo "Applying migrations for environment: $ENVIRONMENT"
        sql-migrate up -config="$CONFIG_FILE" -env="$ENVIRONMENT"
        ;;
    "down")
        echo "Rolling back last migration for environment: $ENVIRONMENT"
        sql-migrate down -config="$CONFIG_FILE" -env="$ENVIRONMENT" -limit=1
        ;;
    "status")
        echo "Migration status for environment: $ENVIRONMENT"
        sql-migrate status -config="$CONFIG_FILE" -env="$ENVIRONMENT"
        ;;
    "new")
        if [ -z "$3" ]; then
            echo "Usage: $0 new [environment] [migration_name]"
            exit 1
        fi
        MIGRATION_NAME=$3
        TIMESTAMP=$(date +%Y%m%d_%H%M%S)
        FILENAME="configs/migrations/sql-migrate/${TIMESTAMP}_${MIGRATION_NAME}.sql"

        cat > "$FILENAME" << EOF
-- +migrate Up
-- TODO: Add your migration here

-- +migrate Down
-- TODO: Add your rollback here
EOF

        echo "Created new migration: $FILENAME"
        ;;
    *)
        echo "Unknown command: $COMMAND"
        exit 1
        ;;
esac