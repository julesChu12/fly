.PHONY: build test clean run dev help

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  test     - Run tests with coverage"
	@echo "  clean    - Clean build artifacts"
	@echo "  run      - Run the application"
	@echo "  dev      - Setup development environment"
	@echo "  lint     - Run linter (if available)"
	@echo "  help     - Show this help message"

build:
	@chmod +x ./scripts/build.sh
	@./scripts/build.sh

test:
	@chmod +x ./scripts/test.sh
	@./scripts/test.sh

clean:
	@rm -rf ./bin
	@rm -f coverage.out coverage.html
	@echo "Clean completed!"

run: build
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