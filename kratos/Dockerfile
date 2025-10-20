# Build stage
FROM golang:1.25.1-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go.mod for kratos and local mora module; later set replace and download deps
COPY kratos/go.mod kratos/go.sum ./
COPY mora ./mora
COPY kratos/. .
RUN go mod edit -replace github.com/julesChu12/fly/mora=./mora && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kratos ./cmd/kratos                                                                              

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /build/kratos .
COPY --from=builder /build/configs ./configs

# Create non-root user
RUN adduser -D -s /bin/sh kratos
RUN chown -R kratos:kratos /root
USER kratos

# Expose ports
EXPOSE 8081 9081

# Run the application
CMD ["./kratos", "serve"]
