# Build stage
FROM golang:1.25.1-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kratos ./cmd/kratos

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/kratos .
COPY --from=builder /app/configs ./configs

# Expose ports
EXPOSE 8082 9092

# Run the application
CMD ["./kratos"]
