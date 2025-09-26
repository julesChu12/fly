.PHONY: build test clean run dev help docker-build docker-run

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Run the application"
	@echo "  dev           - Setup development environment"
	@echo "  lint          - Run linter (if available)"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-stop   - Stop Docker Compose services"
	@echo "  help          - Show this help message"

build:
	@echo "Building Clotho..."
	@go build -ldflags="-s -w" -o bin/clotho cmd/clotho/*.go
	@echo "Build completed!"

test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Tests completed! Coverage report generated at coverage.html"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf ./bin
	@rm -f coverage.out coverage.html
	@docker system prune -f
	@echo "Clean completed!"

run: build
	@echo "Starting Clotho server..."
	@./bin/clotho serve

dev:
	@echo "Starting development server with hot reload..."
	@go run cmd/clotho/*.go serve

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running linter..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

docker-build:
	@echo "Building Docker image..."
	@docker build -t clotho:latest .
	@echo "Docker image built!"

docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d
	@echo "Services started! Clotho available at http://localhost:8080"
	@echo "Health check: curl http://localhost:8080/health"

docker-stop:
	@echo "Stopping Docker Compose services..."
	@docker-compose down
	@echo "Services stopped!"

docker-logs:
	@docker-compose logs -f clotho

init-deps:
	@echo "Installing development dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed!"