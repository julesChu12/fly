# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o plutus ./cmd/plutus

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Shanghai

# Create non-root user
RUN addgroup -g 1000 plutus && \
    adduser -D -u 1000 -G plutus plutus

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/plutus .

# Copy config files
COPY --from=builder /app/configs ./configs

# Change ownership
RUN chown -R plutus:plutus /app

# Switch to non-root user
USER plutus

# Expose ports
# 8085: HTTP API
# 9085: gRPC (when implemented)
# 8081: Health check
# 9090: Metrics
EXPOSE 8085 9085 8081 9090

# Run the application
CMD ["./plutus"]
