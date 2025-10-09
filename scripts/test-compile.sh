#!/bin/bash

# 简单编译测试脚本
echo "🔧 测试服务编译..."
echo "=================================="

SERVICES=("hermes" "kratos" "plutus")

for service in "${SERVICES[@]}"; do
    echo "编译 $service 服务:"
    cd "$service"
    
    if go build ./cmd/$service 2>/dev/null; then
        echo "  ✅ 编译成功"
        rm -f $service 2>/dev/null || true
    else
        echo "  ❌ 编译失败"
        echo "  错误详情:"
        go build ./cmd/$service 2>&1 | head -5 | sed 's/^/    /'
    fi
    
    cd ..
    echo ""
done

echo "🎯 编译测试完成"
