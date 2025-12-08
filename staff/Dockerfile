# Multi-stage, multi-environment Dockerfile for Staff Service
# Usage:
#   Development: docker build --target dev -t staff:dev .
#   Production:  docker build --target prod -t staff:latest .

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
COPY staff/go.mod staff/go.sum ./
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
COPY staff/ ./

# Build with debug info
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -gcflags="all=-N -l" -o staff ./cmd/staff/main.go

# Set proper permissions before switching user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Default configuration
ENV SERVER_HTTP_PORT=8090
ENV SERVER_GRPC_PORT=9090

# Expose ports
EXPOSE 8090 9090

# Default command for development
CMD ["./staff", "serve"]

# ========================================
# Builder Stage - Production Build
# ========================================
FROM base AS builder

# Copy source code
COPY staff/ ./

# Build optimized binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${BUILD_OS} GOARCH=${BUILD_ARCH} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o staff ./cmd/staff/main.go

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
COPY --from=builder /app/staff /app/staff

# Create necessary directories
RUN mkdir -p /app/logs /app/data && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Default configuration
ENV SERVER_HTTP_PORT=8090
ENV SERVER_GRPC_PORT=9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8090/health || exit 1

# Expose ports
EXPOSE 8090 9090

# Default command
CMD ["/app/staff", "serve"]