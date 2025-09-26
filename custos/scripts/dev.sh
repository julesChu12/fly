#!/bin/bash

set -e

echo "Setting up development environment..."

# Copy environment file if it doesn't exist
if [ ! -f .env ]; then
    cp configs/local.env.example .env
    echo "Created .env file from template"
    echo "Please update the .env file with your database credentials"
fi

# Install dependencies
echo "Installing Go dependencies..."
go mod tidy

# Build the application
echo "Building application..."
./scripts/build.sh

echo "Development environment setup complete!"
echo ""
echo "Next steps:"
echo "1. Update .env with your database configuration"
echo "2. Create MySQL database: custos_dev"
echo "3. Run the service: ./bin/userd"