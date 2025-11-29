#!/bin/bash

echo "=== Clotho 系统健康检查 ==="
echo "检查时间: $(date)"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 服务列表
SERVICES=(
    "custos:9001:Custos (认证服务)"
    "hermes:8080:Hermes (客户服务)"
    "appointments:8083:Appointments (预约服务)"
    "staff:8084:Staff (员工服务)"
    "plutus:8085:Plutus (支付服务)"
    "kratos:8082:Kratos (订单服务)"
    "items:8086:Items (商品服务)"
    "clotho:8080:Clotho (API网关)"
)

# 检查单个服务
check_service() {
    local name=$1
    local port=$2
    local desc=$3

    echo -n "检查 $desc (端口 $port)... "

    # 检查端口是否被占用
    if lsof -i :$port >/dev/null 2>&1; then
        # 端口被占用，尝试健康检查
        if [ "$name" = "clotho" ]; then
            # Clotho 有特殊的健康检查端点
            if curl -s http://localhost:$port/health >/dev/null 2>&1; then
                echo -e "${GREEN}✅ 运行正常${NC}"
                return 0
            else
                echo -e "${YELLOW}⚠️  端口被占用但健康检查失败${NC}"
                return 1
            fi
        elif [ "$name" = "items" ]; then
            # Items 服务
            if curl -s http://localhost:$port/health >/dev/null 2>&1; then
                echo -e "${GREEN}✅ 运行正常${NC}"
                return 0
            else
                echo -e "${YELLOW}⚠️  端口被占用但健康检查失败${NC}"
                return 1
            fi
        else
            # 其他服务，假设有 /health 端点
            if curl -s http://localhost:$port/health >/dev/null 2>&1; then
                echo -e "${GREEN}✅ 运行正常${NC}"
                return 0
            else
                echo -e "${YELLOW}⚠️  端口被占用${NC}"
                return 1
            fi
        fi
    else
        echo -e "${RED}❌ 未运行${NC}"
        return 2
    fi
}

# 1. 检查所有服务
echo "1. 检查服务状态："
echo "------------------------"
FAILED_SERVICES=()
for service in "${SERVICES[@]}"; do
    IFS=':' read -r name port desc <<< "$service"
    if ! check_service "$name" "$port" "$desc"; then
        FAILED_SERVICES+=("$name")
    fi
done

# 2. 检查配置文件
echo -e "\n2. 检查配置文件："
echo "------------------------"
if [ -f "configs/clotho.yaml" ]; then
    echo -e "${GREEN}✅ clotho.yaml 存在${NC}"
else
    echo -e "${RED}❌ clotho.yaml 不存在${NC}"
fi

# 3. 检查编译状态
echo -e "\n3. 检查编译状态："
echo "------------------------"
echo "检查 clotho 编译..."
go build ./cmd/clotho >/dev/null 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ clotho 编译成功${NC}"
else
    echo -e "${RED}❌ clotho 编译失败${NC}"
fi

# 4. 生成启动建议
echo -e "\n4. 启动建议："
echo "------------------------"
if [ ${#FAILED_SERVICES[@]} -gt 0 ]; then
    echo -e "${YELLOW}需要启动以下服务：${NC}"
    for service in "${FAILED_SERVICES[@]}"; do
        case $service in
            "custos")
                echo "  - cd ../custos && go run ./cmd/custos serve"
                ;;
            "hermes")
                echo "  - cd ../hermes && go run ./cmd/hermes serve"
                ;;
            "appointments")
                echo "  - cd ../appointments && go run ./cmd/appointments serve"
                ;;
            "staff")
                echo "  - cd ../staff && go run ./cmd/staff serve"
                ;;
            "plutus")
                echo "  - cd ../plutus && go run ./cmd/plutus serve"
                ;;
            "kratos")
                echo "  - cd ../kratos && go run ./cmd/kratos serve"
                ;;
            "items")
                echo "  - cd ../items && go run ./cmd/items serve"
                ;;
            "clotho")
                echo "  - go run ./cmd/clotho serve"
                ;;
        esac
    done
else
    echo -e "${GREEN}✅ 所有服务都已运行${NC}"
    echo -e "可以运行: go run ./cmd/clotho serve"
fi

# 5. 检查数据库连接
echo -e "\n5. 数据库连接检查："
echo "------------------------"
if grep -q "mysql" configs/clotho.yaml; then
    DB_HOST=$(grep "host:" configs/clotho.yaml | awk '{print $2}')
    DB_PORT=$(grep "port:" configs/clotho.yaml | awk '{print $2}')
    if [ -n "$DB_HOST" ] && [ -n "$DB_PORT" ]; then
        echo -n "检查 MySQL ($DB_HOST:$DB_PORT)... "
        if nc -z localhost 3306 2>/dev/null; then
            echo -e "${GREEN}✅ 可访问${NC}"
        else
            echo -e "${RED}❌ 无法连接${NC}"
        fi
    fi
fi

echo -e "\n=== 检查完成 ==="