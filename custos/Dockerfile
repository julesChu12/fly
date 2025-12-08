# Multi-stage, multi-environment Dockerfile for Custos Service
# Usage:
#   Development: docker build --target dev -t custos:dev .
#   Production:  docker build --target prod -t custos:latest .

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
COPY custos/go.mod custos/go.sum ./
COPY mora /mora
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod edit -replace github.com/julesChu12/fly/mora=/mora && \
    go mod download

# ========================================
# Development Stage
# ========================================
FROM base AS dev

# Install development tools
RUN apk add --no-cache curl make && \
    go install github.com/air-verse/air@latest

# Copy source code
COPY custos/ ./
COPY custos/configs /app/configs
COPY mora /mora
COPY custos/.air.toml /app/.air.toml

# Build with debug info
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -gcflags="all=-N -l" -o custos ./cmd/main.go

# Set ownership of /app to appuser
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser
ENV PATH="/home/appuser/go/bin:${PATH}"

# Expose ports
EXPOSE 8081 9081

# Default command for development
CMD ["air"]

# ========================================
# Builder Stage - Production Build
# ========================================
FROM base AS builder

# Copy source code
COPY custos/ ./
COPY custos/configs /app/configs
COPY mora /mora

# Build optimized binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${BUILD_OS} GOARCH=${BUILD_ARCH} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o custos ./cmd/main.go

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
COPY --from=builder /app/custos /app/custos

# Create necessary directories
RUN mkdir -p /app/logs /app/data && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Default configuration path (can be overridden by docker-compose)
ENV CONFIG_PATH=/app/configs/custos.yaml

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8081/api/v1/health || exit 1

# Expose ports
EXPOSE 8081 9081

# Default command
CMD ["/app/custos", "serve"]