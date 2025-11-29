.PHONY: help build test run clean dev lint proto grpc tidy

# Default target
help:
	@echo "Available commands:"
	@echo "  build     - Build the application"
	@echo "  test      - Run tests"
	@echo "  run       - Build and run the application"
	@echo "  dev       - Setup development environment"
	@echo "  lint      - Run linter"
	@echo "  clean     - Clean build artifacts"
	@echo "  proto     - Generate protobuf files"
	@echo "  grpc      - Start gRPC server"
	@echo "  tidy      - Clean up dependencies"

# Variables
APP_NAME := items
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/$(APP_NAME)
	@echo "Build complete: bin/$(APP_NAME)"

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Test coverage report: coverage.html"

# Build and run the application
run: build
	@echo "Running $(APP_NAME)..."
	./bin/$(APP_NAME)

# Setup development environment
dev:
	@echo "Setting up development environment..."
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Development environment ready"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf docs/
	@echo "Clean complete"

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@mkdir -p api/proto/generated
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto

# Generate Swagger documentation
docs:
	@echo "Generating Swagger documentation..."
	@mkdir -p docs
	swag init -g cmd/$(APP_NAME)/main.go -o docs

# Start gRPC server
grpc:
	@echo "Starting gRPC server..."
	go run cmd/$(APP_NAME)/main.go -mode grpc

# Clean up dependencies
tidy:
	@echo "Cleaning up dependencies..."
	go mod tidy
	go mod verify

# Initialize database
init-db:
	@echo "Initializing database..."
	go run cmd/$(APP_NAME)/main.go -init-db

# Run database migrations
migrate:
	@echo "Running database migrations..."
	go run cmd/$(APP_NAME)/main.go -migrate

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8086:8086 -p 50056:50056 $(APP_NAME):$(VERSION)