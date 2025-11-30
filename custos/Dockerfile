# Step 1: Build stage
FROM golang:1.25.1-alpine AS builder

# Set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝代码并编译
COPY . .
RUN go build -ldflags="-s -w" -o userd ./cmd/userd

# Step 2: Runtime stage
FROM alpine:3.20

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 复制编译好的二进制
COPY --from=builder /build/userd /app/userd
COPY --from=builder /build/custos/configs /app/configs
COPY --from=builder /build/custos/docs /app/docs

# 创建非 root 用户
RUN adduser -D -s /bin/sh custos
USER custos

# 默认配置路径（可被 docker-compose 覆盖）
ENV CONFIG_PATH=/app/configs/custos.yaml

EXPOSE 8081 9001
ENTRYPOINT ["/app/userd"]
CMD ["serve"]
