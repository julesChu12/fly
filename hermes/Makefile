.PHONY: proto build test clean swagger

# 生成protobuf代码
proto:
	@echo "Generating protobuf code for Hermes..."
	@mkdir -p api/proto/hermes/v1
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/hermes.proto

# 生成Swagger文档
swagger:
	@echo "Generating Swagger docs for Hermes..."
	@swag init -g cmd/hermes/main.go --output docs

# 构建应用
build:
	@echo "Building Hermes service..."
	@go build -o bin/hermes cmd/hermes/main.go

# 运行测试
test:
	@echo "Running tests for Hermes..."
	@go test ./...

# 清理生成文件
clean:
	@echo "Cleaning generated files..."
	@rm -rf bin/ docs/ api/proto/hermes/

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# 运行服务
run:
	@echo "Running Hermes service..."
	@go run cmd/hermes/main.go

# Docker构建
docker-build:
	@echo "Building Docker image for Hermes..."
	@docker build -t hermes:latest .

# 开发环境运行
dev: deps proto swagger run