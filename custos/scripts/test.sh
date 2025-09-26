#!/bin/bash

set -e

echo "Running Custos tests..."

# Run unit tests with coverage
echo "Running unit tests..."
go test -v -race -coverprofile=coverage.out ./...

# Display coverage report
echo "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "Tests completed!"
echo "Coverage report: coverage.html"