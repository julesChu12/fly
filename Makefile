.PHONY: help dev down clean ps logs

# Default target
help:
	@echo "Available commands:"
	@echo "  dev           - Start all services in development mode"
	@echo "  down          - Stop all services"
	@echo "  clean         - Stop all services and remove all data"
	@echo "  ps            - Show the status of all services"
	@echo "  logs          - Show the logs of all services"
	@echo "  env-setup     - Setup environment files"

# Environment files
ENV_DEV := .env.dev
ENV_EXAMPLE := .env.example

# Start all services
dev: env-setup
	@echo "Starting all services in development mode..."
	@docker-compose -f docker-compose.yaml --env-file $(ENV_DEV) up --build -d

# Stop all services
down:
	@echo "Stopping all services..."
	@docker-compose -f docker-compose.yaml down

# Clean all services
clean: down
	@echo "Cleaning all services..."
	@docker system prune -af
	@docker volume prune -f

# Show status
ps:
	@docker-compose -f docker-compose.yaml ps

# Show logs
logs:
	@docker-compose -f docker-compose.yaml logs -f

# Setup environment files
env-setup:
	@if [ ! -f $(ENV_DEV) ]; then \
		echo "Creating development environment file..."; \
		cp $(ENV_EXAMPLE) $(ENV_DEV); \
	fi
	@echo "✅ Environment files setup complete"
