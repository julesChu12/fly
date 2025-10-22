BINARY_NAME=kratos
MAIN_PACKAGE_DIR=./cmd/kratos
MAIN_PACKAGE=$(MAIN_PACKAGE_DIR)/main.go
SWAGGER_OUTPUT=docs

# Include shared Swagger targets
include ../scripts/makefile-swagger.mk

.PHONY: build run test clean deps

# Build the application
build: pre-build-swagger
	go build -o bin/$(BINARY_NAME) $(MAIN_PACKAGE_DIR)

# Run the application
run: pre-build-swagger
	go run $(MAIN_PACKAGE_DIR)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean: swagger-clean
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	go mod download
	go mod tidy

# Install tools
install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Run all checks
check: fmt lint test

# Docker build
docker-build:
	docker build -t fly-kratos .

# Docker run
docker-run:
	docker run -p 8082:8082 -p 9092:9092 fly-kratos

# Development setup
dev-setup: deps install-tools swagger

.DEFAULT_GOAL := build