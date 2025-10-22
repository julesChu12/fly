# ==============================================================================
# Swagger Documentation Generator - Shared Makefile
# ==============================================================================
# Usage: Include this file in your module's Makefile
# Example: include ../scripts/makefile-swagger.mk
#
# Required variables:
#   - MAIN_PACKAGE: Path to main.go (e.g., cmd/service/main.go)
#
# Optional variables:
#   - SWAGGER_OUTPUT: Output directory (default: docs)
# ==============================================================================

SWAGGER_OUTPUT ?= docs

.PHONY: check-swagger swagger-clean swagger-force pre-build-swagger

# Check if swag is installed
check-swagger:
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "❌ Error: swag is not installed"; \
		echo "Install it with: go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi

# Pre-build step: ensure Swagger docs are up to date
pre-build-swagger: check-swagger
	@echo "🔄 Checking Swagger documentation..."
	@if [ ! -f $(SWAGGER_OUTPUT)/docs.go ] || \
	   find internal/interface/http/handler -type f -name "*.go" -newer $(SWAGGER_OUTPUT)/docs.go 2>/dev/null | grep -q .; then \
		echo "📝 Generating Swagger documentation..."; \
		swag init -g $(MAIN_PACKAGE) --output $(SWAGGER_OUTPUT) --parseDependency --parseInternal; \
		echo "✅ Swagger documentation updated"; \
	else \
		echo "✅ Swagger documentation is up to date"; \
	fi

# Force regenerate Swagger documentation
swagger-force: check-swagger
	@echo "📝 Generating Swagger docs..."
	@swag init -g $(MAIN_PACKAGE) --output $(SWAGGER_OUTPUT) --parseDependency --parseInternal
	@echo "✅ Swagger documentation generated in $(SWAGGER_OUTPUT)/"

# Alias for backward compatibility
swagger: swagger-force

# Clean Swagger generated files
swagger-clean:
	@rm -f $(SWAGGER_OUTPUT)/docs.go $(SWAGGER_OUTPUT)/swagger.json $(SWAGGER_OUTPUT)/swagger.yaml
	@echo "✅ Swagger documentation cleaned"

# Help for Swagger targets
swagger-help:
	@echo "Swagger targets:"
	@echo "  swagger         - Force regenerate Swagger documentation"
	@echo "  swagger-force   - Force regenerate Swagger documentation"
	@echo "  swagger-clean   - Remove generated Swagger files"
	@echo "  check-swagger   - Check if swag tool is installed"
