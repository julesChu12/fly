#!/bin/bash

echo "=================================="
echo "Fly 微服务生态系统状态检查"
echo "=================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 检查函数
check_docker_service() {
    local service_name=$1
    local container_pattern=$2
    local health_check=$3

    echo -n -e "${BLUE}$service_name: ${NC}"

    # 检查 Docker 容器
    if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "$container_pattern" >/dev/null 2>&1; then
        local status=$(docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "$container_pattern" | awk '{print $2}')
        if echo "$status" | grep -q "healthy\|Up"; then
            echo -e " ${GREEN}✅ 运行中 ($status)${NC}"
            return 0
        else
            echo -e " ${YELLOW}⚠️ 容器存在但状态异常 ($status)${NC}"
            return 1
        fi
    else
        echo -e " ${RED}❌ 未运行${NC}"
        return 2
    fi
}

check_port_service() {
    local service_name=$1
    local port=$2
    local health_url=$3

    echo -n -e "${BLUE}$service_name (端口 $port): ${NC}"

    # 检查端口是否开放
    if lsof -i :$port >/dev/null 2>&1; then
        echo -n "${GREEN}✅ 端口开放${NC}"

        # 如果提供了健康检查URL，尝试访问
        if [ -n "$health_url" ]; then
            if curl -s --max-time 2 "$health_url" >/dev/null 2>&1; then
                echo -e " - ${GREEN}健康检查通过${NC}"
            else
                echo -e " - ${YELLOW}健康检查失败${NC}"
            fi
        else
            echo ""
        fi
        return 0
    else
        echo -e "${RED}❌ 端口关闭${NC}"
        return 1
    fi
}

check_service_directory() {
    local service_name=$1
    local directory=$2

    if [ -d "../$directory" ]; then
        echo -e "  ${BLUE}目录存在: ${GREEN}../$directory${NC}"
        if [ -f "../$directory/go.mod" ]; then
            echo -e "  ${BLUE}Go模块: ${GREEN}已配置${NC}"
        fi
    else
        echo -e "  ${YELLOW}目录不存在: ../$directory${NC}"
    fi
}

echo "----------------------------------"
echo "📦 基础设施服务"
echo "----------------------------------"

# MySQL
check_docker_service "MySQL" "mysql" ""
echo -n "  连接测试: "
if mysql -h localhost -P 3306 -u root -p123456 -e "SELECT 1" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ 可连接${NC}"
elif mysql -h localhost -P 3306 -u custos -pcustospassword -e "USE clotho_db; SELECT 1" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ 可连接 (使用 clotho_db)${NC}"
else
    echo -e "${RED}❌ 连接失败${NC}"
fi

# Redis
check_docker_service "Redis" "redis" ""
echo -n "  连接测试: "
if redis-cli -h localhost -p 6379 ping 2>/dev/null | grep -q PONG; then
    echo -e "${GREEN}✅ 可连接${NC}"
else
    echo -e "${RED}❌ 连接失败${NC}"
fi

# Jaeger
check_docker_service "Jaeger" "jaeger" ""
if [ $? -ne 0 ]; then
    echo -e "  ${YELLOW}提示: Jaeger 是可选的追踪服务${NC}"
fi

# Prometheus
check_docker_service "Prometheus" "prometheus" ""
if [ $? -ne 0 ]; then
    echo -e "  ${YELLOW}提示: Prometheus 是可选的监控服务${NC}"
fi

echo ""
echo "----------------------------------"
echo "🏢 业务域服务"
echo "----------------------------------"

# 检查服务目录和可能的运行状态
services=("custos:9001:http://localhost:9002/health" "hermes:8080:http://localhost:8080/health"
          "kratos:8082:http://localhost:8082/health" "appointments:8083:http://localhost:8083/health"
          "staff:8084:http://localhost:8084/health" "plutus:8085:http://localhost:8085/health"
          "items:8086:http://localhost:8086/health")

for service_info in "${services[@]}"; do
    IFS=':' read -r service_name gRPC_port health_url <<< "$service_info"

    # 检查目录
    check_service_directory "$service_name" "$service_name"

    # 检查gRPC端口（如果有的话）
    if [ -n "$gRPC_port" ] && [ "$gRPC_port" != "8080" ]; then
        check_port_service "$service_name gRPC" "$gRPC_port" ""
    fi

    # 检查HTTP端口
    http_port=$(echo "$health_url" | sed 's/.*:\([0-9]*\)\/.*/\1/')
    check_port_service "$service_name HTTP" "$http_port" "$health_url"

    echo ""
done

echo "----------------------------------"
echo "🌐 Clotho API 网关"
echo "----------------------------------"

# Clotho 可能在多个端口运行
for port in 8087 8088 8080; do
    if lsof -i :$port >/dev/null 2>&1; then
        check_port_service "Clotho" "$port" "http://localhost:$port/health"
        break
    fi
done

echo ""
echo "----------------------------------"
echo "📊 服务汇总"
echo "----------------------------------"

# 统计服务状态
total_infrastructure=4
running_infrastructure=$(docker ps --format "{{.Names}}" | grep -E "mysql|redis|jaeger|prometheus" | wc -l)

total_business=7
running_business=0
for port in 8080 8081 8082 8083 8084 8085 8086; do
    if lsof -i :$port >/dev/null 2>&1; then
        ((running_business++))
    fi
done

echo -e "基础设施服务: ${GREEN}$running_infrastructure/$total_infrastructure${NC}"
echo -e "业务域服务: ${GREEN}$running_business/$total_business${NC}"

if [ $running_infrastructure -eq $total_infrastructure ] && [ $running_business -eq $total_business ]; then
    echo -e "\n${GREEN}🎉 所有服务正常运行！${NC}"
else
    echo -e "\n${YELLOW}⚠️ 部分服务未运行${NC}"
    echo -e "${BLUE}💡 建议：${NC}"
    echo "1. 检查 Docker 是否正常运行: docker ps"
    echo "2. 启动所有服务: ./start_all_services.sh"
    echo "3. 使用 Docker Compose: docker-compose up -d"
    echo "4. 查看服务日志: docker logs <service-name>"
fi

echo ""
echo "----------------------------------"
echo "🔧 快速操作指南"
echo "----------------------------------"
echo "启动所有服务:"
echo "  ./start_all_services.sh"
echo ""
echo "停止所有服务:"
echo "  ./start_all_services.sh stop"
echo ""
echo "查看服务健康:"
echo "  ./check_system_health.sh"
echo ""
echo "重启特定服务:"
echo "  docker restart <container-name>"