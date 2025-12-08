# Multi-stage, multi-environment Dockerfile for Plutus Service
# Usage:
#   Development: docker build --target dev -t plutus:dev .
#   Production:  docker build --target prod -t plutus:latest .

# ========================================
# Base Stage - Common dependencies
# ========================================
FROM golang:1.25.1-alpine AS base

ARG GO_VERSION=1.25.1
ARG BUILD_OS=linux
ARG BUILD_ARCH=amd64

WORKDIR /app

# Install base packages
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy go mod files and local dependencies
COPY plutus/go.mod plutus/go.sum ./
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
COPY plutus/ ./

# Build with debug info
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -gcflags="all=-N -l" -o plutus ./cmd/plutus

# Switch to non-root user
USER appuser

# Default configuration
ENV SERVER_HTTP_PORT=8084
ENV SERVER_GRPC_PORT=9084

# Expose ports
EXPOSE 8084 9084

# Default command for development
CMD ["./plutus", "serve"]

# ========================================
# Builder Stage - Production Build
# ========================================
FROM base AS builder

# Copy source code
COPY plutus/ ./

# Build optimized binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${BUILD_OS} GOARCH=${BUILD_ARCH} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o plutus ./cmd/plutus

# ========================================
# Production Stage
# ========================================
FROM base AS prod

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/plutus .
COPY --from=builder /app/configs ./configs

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Default configuration
ENV SERVER_HTTP_PORT=8084
ENV SERVER_GRPC_PORT=9084

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8084/health || exit 1

# Expose ports
EXPOSE 8084 9084

# Default command
CMD ["./plutus", "serve"]