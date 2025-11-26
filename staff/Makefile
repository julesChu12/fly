.PHONY: proto build test clean swagger

# Module configuration
MAIN_PACKAGE := cmd/staff/main.go
SWAGGER_OUTPUT := docs

# Include shared Swagger targets
include ../scripts/makefile-swagger.mk

# 构建应用
build: pre-build-swagger
	@echo "Building Staff service..."
	@go build -o bin/staff cmd/staff/main.go

# 运行测试
test:
	@echo "Running tests for Staff..."
	@go test ./...

# 清理生成文件
clean: swagger-clean
	@echo "Cleaning generated files..."
	@rm -rf bin/ api/proto/staff/

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# 运行服务
run: pre-build-swagger
	@echo "Running Staff service..."
	@go run cmd/staff/main.go

# Docker构建
docker-build:
	@echo "Building Docker image for Staff..."
	@docker build -t staff:latest .

# 开发环境运行
dev: deps proto swagger run