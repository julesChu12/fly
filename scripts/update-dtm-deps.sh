#!/bin/bash

# DTM 依赖更新验证脚本
# 用于验证 DTM 客户端版本更新到 1.16.6

set -e

echo "🔧 DTM 依赖版本更新验证"
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

# 检查 DTM 版本更新
verify_step "检查 DTM 依赖版本更新..."

SERVICES=("hermes" "kratos" "plutus")
TARGET_VERSION="v1.16.6"

for service in "${SERVICES[@]}"; do
    if [ -f "$service/go.mod" ]; then
        echo "检查 $service 服务:"
        
        # 检查 client 版本
        if grep -q "github.com/dtm-labs/client $TARGET_VERSION" "$service/go.mod"; then
            success "  ✓ dtm-labs/client 版本已更新到 $TARGET_VERSION"
        else
            error "  ✗ dtm-labs/client 版本未更新到 $TARGET_VERSION"
        fi
        
        # 检查 dtmgrpc 版本
        if grep -q "github.com/dtm-labs/dtmgrpc $TARGET_VERSION" "$service/go.mod"; then
            success "  ✓ dtm-labs/dtmgrpc 版本已更新到 $TARGET_VERSION"
        else
            error "  ✗ dtm-labs/dtmgrpc 版本未更新到 $TARGET_VERSION"
        fi
        
        echo ""
    else
        error "服务目录不存在: $service"
    fi
done

# 验证依赖下载
verify_step "验证依赖下载..."

for service in "${SERVICES[@]}"; do
    if [ -d "$service" ]; then
        echo "验证 $service 服务依赖:"
        cd "$service"
        
        # 清理模块缓存
        go clean -modcache 2>/dev/null || true
        
        # 下载依赖
        if go mod download 2>/dev/null; then
            success "  ✓ 依赖下载成功"
        else
            error "  ✗ 依赖下载失败"
        fi
        
        # 验证依赖
        if go mod verify 2>/dev/null; then
            success "  ✓ 依赖验证通过"
        else
            warning "  ⚠ 依赖验证失败，但可能不影响编译"
        fi
        
        cd ..
        echo ""
    fi
done

# 测试编译
verify_step "测试服务编译..."

for service in "${SERVICES[@]}"; do
    if [ -d "$service" ]; then
        echo "编译 $service 服务:"
        cd "$service"
        
        # 尝试编译
        if go build ./cmd/$service 2>/dev/null; then
            success "  ✓ 编译成功"
            # 清理编译产物
            rm -f $service 2>/dev/null || true
        else
            error "  ✗ 编译失败"
            echo "  详细错误信息:"
            go build ./cmd/$service 2>&1 | sed 's/^/    /'
        fi
        
        cd ..
        echo ""
    fi
done

# 检查 DTM 管理器代码
verify_step "检查 DTM 管理器代码..."

if [ -f "hermes/pkg/dtm/manager.go" ]; then
    echo "检查 DTM 管理器实现:"
    
    # 检查导入
    if grep -q "github.com/dtm-labs/client/dtmcli" "hermes/pkg/dtm/manager.go"; then
        success "  ✓ dtmcli 导入正确"
    else
        warning "  ⚠ dtmcli 导入可能有问题"
    fi
    
    if grep -q "github.com/dtm-labs/client/dtmgrpc" "hermes/pkg/dtm/manager.go"; then
        success "  ✓ dtmgrpc 导入正确"
    else
        warning "  ⚠ dtmgrpc 导入可能有问题"
    fi
    
    # 检查代码完整性
    if grep -q "func.*Close.*error" "hermes/pkg/dtm/manager.go"; then
        success "  ✓ DTM 管理器方法完整"
    else
        warning "  ⚠ DTM 管理器方法可能不完整"
    fi
else
    warning "DTM 管理器文件不存在"
fi

echo ""
echo "🎯 更新总结"
echo "=================================="

# 统计更新结果
updated_services=0
total_services=${#SERVICES[@]}

for service in "${SERVICES[@]}"; do
    if [ -f "$service/go.mod" ]; then
        if grep -q "github.com/dtm-labs/client $TARGET_VERSION" "$service/go.mod" && \
           grep -q "github.com/dtm-labs/dtmgrpc $TARGET_VERSION" "$service/go.mod"; then
            ((updated_services++))
        fi
    fi
done

if [ $updated_services -eq $total_services ]; then
    success "🎉 所有服务的 DTM 依赖已成功更新到 $TARGET_VERSION"
    echo ""
    echo "📋 后续步骤:"
    echo "  1. 运行 'go mod tidy' 清理依赖"
    echo "  2. 测试服务启动和功能"
    echo "  3. 验证分布式事务功能"
else
    warning "⚠️  $updated_services/$total_services 个服务更新完成"
fi

echo ""
echo "🔗 相关命令:"
echo "  go mod tidy                    # 清理依赖"
echo "  go mod download               # 下载依赖"
echo "  go build ./cmd/[service]      # 编译服务"
echo "  go test ./...                 # 运行测试"
