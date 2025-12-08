.PHONY: help build test run clean dev lint proto grpc tidy docker-build docker-run docker-clean \
	docker-dev docker-dev-down docker-prod docker-prod-down docker-prod-full \
	deploy-dev deploy-prod deploy-prod-safe rollback init-db migrate health logs \
	dev-watch migration-create backup env-setup env-dev env-prod docs

# Default target
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  run           - Build and run the application"
	@echo "  dev           - Setup development environment"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  proto         - Generate protobuf files"
	@echo "  grpc          - Start gRPC server"
	@echo "  tidy          - Clean up dependencies"
	@echo ""
	@echo "Environment management:"
	@echo "  env-setup     - Setup environment files"
	@echo "  env-dev       - Load development environment"
	@echo "  env-prod      - Load production environment"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build  - Build Docker image for current environment"
	@echo "  docker-run    - Run Docker container for current environment"
	@echo "  docker-dev    - Start development environment with Docker"
	@echo "  docker-prod   - Start production environment with Docker"
	@echo "  docker-clean  - Clean Docker resources"
	@echo ""
	@echo "Deployment commands:"
	@echo "  deploy-dev    - Deploy to development environment"
	@echo "  deploy-prod   - Deploy to production environment"

# Variables
APP_NAME := plutus
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Environment files
ENV_DIR := configs
ENV_DEV := $(ENV_DIR)/.env.dev
ENV_PROD := $(ENV_DIR)/.env.prod
ENV_EXAMPLE := $(ENV_DIR)/.env.example

# Default environment
APP_ENV ?= development

# Load environment-specific variables (if exists)
-include $(ENV_DEV)

# ========================================
# Environment Management
# ========================================

# Setup environment files
env-setup:
	@if [ ! -f $(ENV_DEV) ]; then \
		echo "Creating development environment file..."; \
		cp $(ENV_EXAMPLE) $(ENV_DEV); \
	fi
	@if [ ! -f $(ENV_PROD) ]; then \
		echo "Creating production environment file..."; \
		cp $(ENV_EXAMPLE) $(ENV_PROD); \
		echo "⚠️  Please edit $(ENV_PROD) with production values"; \
	fi
	@echo "✅ Environment files setup complete"

# Load development environment
env-dev:
	@echo "🔧 Loading development environment..."
	@export $(shell cat $(ENV_DEV) | xargs) && echo "✅ Development environment loaded"

# Load production environment
env-prod:
	@echo "🚀 Loading production environment..."
	@export $(shell cat $(ENV_PROD) | xargs) && echo "✅ Production environment loaded"

# ========================================
# Build & Test Commands
# ========================================

# Build the application
build:
	@echo "Building $(APP_NAME) for $(APP_ENV) environment..."
	@mkdir -p bin
	@if [ -f $(ENV_DIR)/.env.$(APP_ENV) ]; then \
		source $(ENV_DIR)/.env.$(APP_ENV); \
	fi && \
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/plutus/
	@echo "✅ Build complete: bin/$(APP_NAME)"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Test coverage report: coverage.html"

# Build and run the application
run: build
	@echo "🚀 Running $(APP_NAME) in $(APP_ENV) mode..."
	@if [ -f $(ENV_DIR)/.env.$(APP_ENV) ]; then \
		source $(ENV_DIR)/.env.$(APP_ENV); \
	fi && \
	./bin/$(APP_NAME)

# Setup development environment
dev:
	@echo "🔧 Setting up development environment..."
	@go mod download
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "✅ Development environment ready"

# Run linter
lint:
	@echo "🔍 Running linter..."
	@golangci-lint run

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

# Generate protobuf files
proto:
	@echo "🔧 Generating protobuf files..."
	@mkdir -p api/proto/plutus/v1
	@protoc --go_out=. --go-opt=paths=source_relative \
		--go-grpc_out=. --go-grpc-opt=paths=source_relative \
		api/proto/plutus/v1/plutus.proto
	@echo "✅ Protobuf files generated to api/proto/plutus/v1/"

# Generate Swagger documentation
docs:
	@echo "📚 Generating Swagger documentation..."
	@mkdir -p docs
	@swag init -g cmd/$(APP_NAME)/main.go -o docs
	@echo "✅ Documentation generated in docs/"

# Start gRPC server
grpc:
	@echo "🔌 Starting gRPC server..."
	@if [ -f $(ENV_DIR)/.env.$(APP_ENV) ]; then \
		source $(ENV_DIR)/.env.$(APP_ENV); \
	fi && \
	go run cmd/plutus/main.go -mode grpc

# Clean up dependencies
tidy:
	@echo "🧹 Cleaning up dependencies..."
	@go mod tidy
	@go mod verify

# ========================================
# Docker Commands
# ========================================

# Build Docker image for current environment
docker-build:
	@echo "🐳 Building Docker image for $(APP_ENV) environment..."
	@if [ -f $(ENV_DIR)/.env.$(APP_ENV) ]; then \
		source $(ENV_DIR)/.env.$(APP_ENV); \
	fi && \
	docker build \
		--target $${BUILD_TARGET:-dev} \
		--build-arg GO_VERSION=$${GO_VERSION:-1.25} \
		--build-arg BUILD_OS=$${BUILD_OS:-linux} \
		--build-arg BUILD_ARCH=$${BUILD_ARCH:-amd64} \
		-t $${DOCKER_IMAGE_NAME:-plutus-service}:$${DOCKER_TAG:-$(APP_ENV)} \
		-t $${DOCKER_IMAGE_NAME:-plutus-service}:$(VERSION) \
		.
	@echo "✅ Docker image built"

# Run Docker container for current environment
docker-run:
	@echo "🐳 Running Docker container for $(APP_ENV) environment..."
	@if [ -f $(ENV_DIR)/.env.$(APP_ENV) ]; then \
		source $(ENV_DIR)/.env.$(APP_ENV); \
	fi && \
	docker run --rm -d \
		--name $${DOCKER_IMAGE_NAME:-plutus-service}-$(APP_ENV) \
		-p $${SERVER_HTTP_PORT:-8085}:$${SERVER_HTTP_PORT:-8085} \
		-p $${SERVER_GRPC_PORT:-9085}:$${SERVER_GRPC_PORT:-9085} \
		-e APP_ENV=$(APP_ENV) \
		-e CONFIG_PATH=/app/configs/plutus.yaml \
		$${DOCKER_IMAGE_NAME:-plutus-service}:$${DOCKER_TAG:-$(APP_ENV)}

# Start development environment with Docker
docker-dev:
	@echo "🐳 Starting development environment..."
	@if [ -f $(ENV_DEV) ]; then \
		APP_ENV=development docker-compose --env-file $(ENV_DEV) up --build; \
	else \
		APP_ENV=development docker-compose up --build; \
	fi

# Stop development environment
docker-dev-down:
	@echo "🛑 Stopping development environment..."
	@if [ -f $(ENV_DEV) ]; then \
		APP_ENV=development docker-compose --env-file $(ENV_DEV) down; \
	else \
		APP_ENV=development docker-compose down; \
	fi

# Start production environment with Docker
docker-prod:
	@echo "🚀 Starting production environment..."
	@if [ -f $(ENV_PROD) ]; then \
		APP_ENV=production docker-compose --env-file $(ENV_PROD) --profile production up --build -d; \
	else \
		APP_ENV=production docker-compose --profile production up --build -d; \
	fi

# Stop production environment
docker-prod-down:
	@echo "🛑 Stopping production environment..."
	@if [ -f $(ENV_PROD) ]; then \
		APP_ENV=production docker-compose --env-file $(ENV_PROD) --profile production down; \
	else \
		APP_ENV=production docker-compose --profile production down; \
	fi

# Clean Docker resources
docker-clean:
	@echo "🧹 Cleaning Docker resources..."
	@docker system prune -f
	@docker volume prune -f
	@docker network prune -f
	@echo "✅ Docker cleanup complete"