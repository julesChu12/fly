FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go.mod for hermes and local mora module; later set replace and download deps
COPY hermes/go.mod hermes/go.sum ./
COPY mora ./mora
COPY hermes/. .
RUN go mod edit -replace github.com/julesChu12/fly/mora=./mora && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/hermes/main.go                                                                          

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /build/main .
COPY --from=builder /build/configs ./configs

# Create non-root user
RUN adduser -D -s /bin/sh hermes
RUN chown -R hermes:hermes /root
USER hermes

EXPOSE 8080
CMD ["./main", "serve"]
