#!/bin/bash

set -e

echo "Building Custos User Service..."

# Clean previous builds
rm -rf ./bin

# Create bin directory
mkdir -p ./bin

# Build the main service
echo "Building custos..."
go build -o ./bin/custos ./cmd/main.go

echo "Build completed successfully!"
echo "Binary location: ./bin/custos"