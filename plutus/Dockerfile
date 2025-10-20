# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go.mod for plutus and local mora module; later set replace and download deps
COPY plutus/go.mod plutus/go.sum ./
COPY mora ./mora
COPY plutus/. .
RUN go mod edit -replace github.com/julesChu12/fly/mora=./mora && go mod download

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
COPY --from=builder /build/plutus .

# Copy config files
COPY --from=builder /build/configs ./configs

# Change ownership
RUN chown -R plutus:plutus /app

# Switch to non-root user
USER plutus

# Expose ports
# 8082: HTTP API
# 9082: gRPC
EXPOSE 8082 9082

# Run the application
CMD ["./plutus", "serve"]
