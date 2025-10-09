BINARY_NAME=plutus
MAIN_PACKAGE=./cmd/plutus

.PHONY: build run test clean swagger deps

# Build the application
build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PACKAGE)

# Run the application
run:
	go run $(MAIN_PACKAGE)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Generate swagger documentation
swagger:
	swag init -g $(MAIN_PACKAGE)/main.go --output docs

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
	docker build -t fly-plutus .

# Docker run
docker-run:
	docker run -p 8085:8085 -p 9085:9085 fly-plutus

# Development setup
dev-setup: deps install-tools swagger

.DEFAULT_GOAL := build