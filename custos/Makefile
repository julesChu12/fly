.PHONY: build test clean run dev help migrate

# Module configuration
MAIN_PACKAGE := cmd/userd/main.go
SWAGGER_OUTPUT := docs

# Include shared Swagger targets
include ../scripts/makefile-swagger.mk

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application (auto-generates Swagger docs)"
	@echo "  test         - Run tests with coverage"
	@echo "  clean        - Clean build artifacts and generated docs"
	@echo "  run          - Run the application (auto-generates Swagger docs)"
	@echo "  dev          - Setup development environment"
	@echo "  lint         - Run linter (if available)"
	@echo "  swagger      - Generate Swagger documentation"
	@echo "  migrate      - Show migration status"
	@echo "  migrate-up   - Apply all pending migrations"
	@echo "  migrate-down - Rollback last migration"
	@echo "  migrate-new  - Create new migration (use NAME=migration_name)"
	@echo "  help         - Show this help message"

build: pre-build-swagger
	@chmod +x ./scripts/build.sh
	@./scripts/build.sh

test:
	@chmod +x ./scripts/test.sh
	@./scripts/test.sh

clean: swagger-clean
	@rm -rf ./bin
	@rm -f coverage.out coverage.html
	@echo "✅ Clean completed!"

run: pre-build-swagger build
	@./bin/userd

dev:
	@chmod +x ./scripts/dev.sh
	@./scripts/dev.sh

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

migrate:
	@chmod +x ./scripts/migrate.sh
	@./scripts/migrate.sh status

migrate-up:
	@chmod +x ./scripts/migrate.sh
	@./scripts/migrate.sh up development

migrate-down:
	@chmod +x ./scripts/migrate.sh
	@./scripts/migrate.sh down development

migrate-new:
	@chmod +x ./scripts/migrate.sh
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-new NAME=migration_name"; \
		exit 1; \
	fi
	@./scripts/migrate.sh new development $(NAME)