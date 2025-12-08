# Multi-stage, multi-environment Dockerfile for Clotho Service
# Usage:
#   Development: docker build --target dev -t clotho:dev .
#   Production:  docker build --target prod -t clotho:latest .

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
COPY clotho/go.mod clotho/go.sum ./
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
COPY clotho/ ./

# Build with debug info
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -gcflags="all=-N -l" -o clotho ./cmd/clotho

# Switch to non-root user
USER appuser

# Default configuration
ENV CONFIG_PATH=/app/configs/clotho.yaml
ENV SERVER_PORT=8080
ENV SERVER_HOST=0.0.0.0

# Expose ports
EXPOSE 8080 9080

# Default command for development
CMD ["./clotho", "serve"]

# ========================================
# Builder Stage - Production Build
# ========================================
FROM base AS builder

# Copy source code
COPY clotho/ ./

# Build optimized binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${BUILD_OS} GOARCH=${BUILD_ARCH} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o clotho ./cmd/clotho

# ========================================
# Production Stage
# ========================================
FROM alpine:latest AS prod

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/clotho ./clotho
COPY --from=builder /app/configs ./configs

# Create necessary directories and set permissions
RUN mkdir -p /app/logs /app/data && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Default configuration
ENV CONFIG_PATH=/app/configs/clotho.yaml
ENV SERVER_PORT=8080
ENV SERVER_HOST=0.0.0.0

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Expose ports
EXPOSE 8080 9080

# Default command
CMD ["./clotho", "serve"]