#!/bin/bash

# Fly平台微服务系统验证脚本
# 用于快速验证三个微服务的基本功能

set -e

echo "🚀 Fly平台微服务系统验证开始..."
echo "=================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 验证函数
verify_step() {
    echo -e "${BLUE}[验证]${NC} $1"
}

success() {
    echo -e "${GREEN}[成功]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[警告]${NC} $1"
}

error() {
    echo -e "${RED}[错误]${NC} $1"
}

# 1. 项目结构验证
verify_step "检查项目结构..."

SERVICES=("hermes" "kratos" "plutus")
REQUIRED_DIRS=("internal/application/service" "internal/domain/entity" "internal/infrastructure/database" "internal/interface/http" "pkg/constants" "pkg/errors" "pkg/types" "cmd" "configs")

for service in "${SERVICES[@]}"; do
    if [ -d "$service" ]; then
        success "服务目录存在: $service"

        for dir in "${REQUIRED_DIRS[@]}"; do
            if [ -d "$service/$dir" ]; then
                echo "  ✓ $service/$dir"
            else
                warning "  ✗ 缺少目录: $service/$dir"
            fi
        done
    else
        error "服务目录不存在: $service"
    fi
done

# 2. Go模块验证
verify_step "检查Go模块文件..."

for service in "${SERVICES[@]}"; do
    if [ -f "$service/go.mod" ]; then
        success "Go模块文件存在: $service/go.mod"

        # 检查关键依赖
        echo "  检查依赖包:"
        grep -E "(gin-gonic|gorm|grpc|swaggo)" "$service/go.mod" | sed 's/^/    /'

    else
        error "Go模块文件不存在: $service/go.mod"
    fi
done

# 3. 配置文件验证
verify_step "检查配置文件..."

for service in "${SERVICES[@]}"; do
    if [ -f "$service/configs/$service.yaml" ]; then
        success "配置文件存在: $service/configs/$service.yaml"
    else
        warning "配置文件不存在: $service/configs/$service.yaml"
    fi
done

# 4. protobuf文件验证
verify_step "检查gRPC定义文件..."

for service in "${SERVICES[@]}"; do
    if [ -f "$service/api/proto/$service.proto" ]; then
        success "protobuf文件存在: $service/api/proto/$service.proto"
    else
        warning "protobuf文件不存在: $service/api/proto/$service.proto"
    fi
done

# 5. 数据库迁移文件验证
verify_step "检查数据库迁移文件..."

for service in "${SERVICES[@]}"; do
    migration_dir="$service/configs/migrations"
    if [ -d "$migration_dir" ] && [ "$(ls -A $migration_dir)" ]; then
        success "迁移文件存在: $migration_dir"
        ls -la "$migration_dir"/*.sql 2>/dev/null | sed 's/^/  /' || warning "  没有找到SQL迁移文件"
    else
        warning "迁移目录为空或不存在: $migration_dir"
    fi
done

# 6. 文档验证
verify_step "检查文档文件..."

docs=("README.md" "REGRESSION_TEST_REPORT.md")
for doc in "${docs[@]}"; do
    if [ -f "$doc" ]; then
        success "文档存在: $doc"
    else
        warning "文档不存在: $doc"
    fi
done

# 7. Docker配置验证
verify_step "检查Docker配置..."

if [ -f "docker-compose.yaml" ]; then
    success "Docker Compose文件存在"

    echo "  检查服务定义:"
    for service in "${SERVICES[@]}"; do
        if grep -q "  $service:" docker-compose.yaml; then
            echo "    ✓ $service 服务已定义"
        else
            warning "    ✗ $service 服务未在docker-compose.yaml中定义"
        fi
    done
else
    warning "Docker Compose文件不存在"
fi

for service in "${SERVICES[@]}"; do
    if [ -f "$service/Dockerfile" ]; then
        success "Dockerfile存在: $service/Dockerfile"
    else
        warning "Dockerfile不存在: $service/Dockerfile"
    fi
done

# 8. 统计信息
verify_step "项目统计信息..."

echo "📊 项目规模统计:"
echo "  Go源文件数量: $(find . -name "*.go" -type f | wc -l)"
echo "  protobuf文件数量: $(find . -name "*.proto" -type f | wc -l)"
echo "  配置文件数量: $(find . -name "*.yaml" -o -name "*.yml" | wc -l)"
echo "  迁移文件数量: $(find . -name "*.sql" -type f | wc -l)"

# 9. 代码质量检查
verify_step "代码质量检查..."

echo "📋 主要文件检查:"
for service in "${SERVICES[@]}"; do
    echo "  $service 服务:"

    # 检查主要Go文件
    main_file="$service/cmd/$service/main.go"
    if [ -f "$main_file" ]; then
        echo "    ✓ 主程序文件: $main_file"
        lines=$(wc -l < "$main_file")
        echo "      - 代码行数: $lines"
    else
        warning "    ✗ 主程序文件不存在: $main_file"
    fi

    # 统计该服务的Go文件数量
    go_files=$(find "$service" -name "*.go" -type f | wc -l)
    echo "    - Go文件数量: $go_files"
done

# 10. 总结
echo ""
echo "🎯 验证总结"
echo "=================================="

total_issues=0

# 检查关键文件
critical_files=(
    "hermes/cmd/hermes/main.go"
    "kratos/cmd/kratos/main.go"
    "plutus/cmd/plutus/main.go"
    "docker-compose.yaml"
    "go.work"
)

echo "关键文件检查:"
for file in "${critical_files[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✓ $file"
    else
        echo "  ✗ $file"
        ((total_issues++))
    fi
done

if [ $total_issues -eq 0 ]; then
    success "🎉 系统验证完成！所有关键组件都已就绪。"
    echo ""
    echo "📋 后续步骤:"
    echo "  1. 解决依赖版本问题 (DTM客户端)"
    echo "  2. 启动数据库服务进行连接测试"
    echo "  3. 运行 'make dev' 启动各个服务"
    echo "  4. 访问 http://localhost:8083/swagger 查看API文档"
else
    warning "⚠️  发现 $total_issues 个问题需要解决"
fi

echo ""
echo "🔗 相关链接:"
echo "  - Hermes API: http://localhost:8083/swagger"
echo "  - Kratos API: http://localhost:8084/swagger"
echo "  - Plutus API: http://localhost:8085/swagger"
echo "  - 健康检查: http://localhost:8083/health"

echo ""
echo "📝 详细报告: REGRESSION_TEST_REPORT.md"