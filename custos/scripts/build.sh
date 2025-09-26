#!/bin/bash

set -e

echo "Building Custos User Service..."

# Clean previous builds
rm -rf ./bin

# Create bin directory
mkdir -p ./bin

# Build the main service
echo "Building userd..."
go build -o ./bin/userd ./cmd/userd

echo "Build completed successfully!"
echo "Binary location: ./bin/userd"