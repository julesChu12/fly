FROM golang:1.25.1-alpine AS builder

# Set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o clotho cmd/clotho/*.go

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/clotho .

# Build a small image
FROM alpine:3.18

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /dist/clotho .

# Copy config files
COPY configs/ ./configs/

# Create non-root user
RUN adduser -D -s /bin/sh clotho
USER clotho

# Expose port
EXPOSE 8080

# Command to run
ENTRYPOINT ["./clotho"]
CMD ["serve"]