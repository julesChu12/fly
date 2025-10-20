FROM golang:1.25.1-alpine AS builder

# Set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

# Move to working directory /build
WORKDIR /build

# Copy go.mod for clotho and the local mora module
COPY clotho/go.mod clotho/go.sum ./
COPY mora ./mora

# Copy the source code into the container (service subdir)
COPY clotho/. .

# After copying source (which includes go.mod), set replace and download deps
RUN go mod edit -replace github.com/julesChu12/fly/mora=./mora && go mod download

# Build the application
RUN go build -ldflags="-s -w" -o clotho ./cmd/clotho

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

# Copy config files from service directory when context is repo root
COPY clotho/configs/ ./configs/

# Create non-root user
RUN adduser -D -s /bin/sh clotho
USER clotho

# Expose port
EXPOSE 8080

# Command to run
ENTRYPOINT ["./clotho"]
CMD ["serve"]