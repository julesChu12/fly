# Multi-stage, multi-environment Dockerfile for Appointments Service
# Usage:
#   Development: docker build --target dev -t appointments:dev .
#   Production:  docker build --target prod -t appointments:latest .

# ========================================
# Base Stage - Common dependencies
# ========================================
FROM golang:1.25.1-alpine AS base

WORKDIR /app

# Install base packages
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy go mod files and local dependencies
COPY appointments/go.mod appointments/go.sum ./
COPY mora /mora
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod edit -replace github.com/julesChu12/fly/mora=/mora && \
    go mod download

# ========================================
# Development Stage
# ========================================
FROM base AS dev

# Install development tools
RUN apk add --no-cache curl make

# Copy source code
COPY appointments/ ./
COPY appointments/configs /app/configs

# Build with debug info
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -gcflags="all=-N -l" -o appointments ./cmd/appointments

# Set proper permissions before switching user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8086 9086

# Default command for development
CMD ["./appointments", "serve"]

# ========================================
# Builder Stage - Production Build
# ========================================
FROM base AS builder

# Copy source code
COPY appointments/ ./
COPY appointments/configs /app/configs

# Build optimized binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o appointments ./cmd/appointments

# ========================================
# Production Stage
# ========================================
FROM base AS prod

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/appointments /app/appointments

# Create necessary directories
RUN mkdir -p /app/logs /app/data && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8086/health || exit 1

# Expose ports
EXPOSE 8086 9086

# Default command
CMD ["/app/appointments", "serve"]

# ========================================
# Testing Stage
# ========================================
FROM base AS test

# Copy source code
COPY appointments/ ./
COPY appointments/configs /app/configs

# Run tests
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go test -v -race -coverprofile=coverage.out ./... && \
    go tool cover -html=coverage.out -o coverage.html

# ========================================
# Security Scanning Stage
# ========================================
FROM scratch AS security

# Copy binary for security scanning
COPY --from=builder /app/appointments /appointments

# Metadata
LABEL maintainer="fly-team" \
      version="1.0.0" \
      description="Appointments Booking Service"