.PHONY: proto build test clean swagger

# Module configuration
MAIN_PACKAGE := cmd/appointments/main.go
SWAGGER_OUTPUT := docs

# Include shared Swagger targets
include ../scripts/makefile-swagger.mk

# 构建应用
build: pre-build-swagger
	@echo "Building Appointments service..."
	@go build -o bin/appointments cmd/appointments/main.go

# 运行测试
test:
	@echo "Running tests for Appointments..."
	@go test ./...

# 清理生成文件
clean: swagger-clean
	@echo "Cleaning generated files..."
	@rm -rf bin/ api/proto/appointments/

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# 运行服务
run: pre-build-swagger
	@echo "Running Appointments service..."
	@go run cmd/appointments/main.go

# Docker构建
docker-build:
	@echo "Building Docker image for Appointments..."
	@docker build -t appointments:latest .

# 开发环境运行
dev: deps proto swagger run