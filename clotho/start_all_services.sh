#!/bin/bash

echo "=== 启动所有 Fly 微服务 ==="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 创建日志目录
mkdir -p logs

# 函数：启动服务
start_service() {
    local service_name=$1
    local service_dir=$2
    local service_port=$3
    local log_file=$4

    echo -e "${BLUE}启动 $service_name (端口 $service_port)...${NC}"

    # 检查服务目录是否存在
    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}错误: 服务目录 $service_dir 不存在${NC}"
        return 1
    fi

    # 检查端口是否已被占用
    if lsof -i :$service_port >/dev/null 2>&1; then
        echo -e "${YELLOW}警告: 端口 $service_port 已被占用${NC}"
        return 0
    fi

    # 启动服务
    cd "$service_dir" || exit 1
    nohup go run ./cmd/${service_dir##*/} serve > "../clotho/logs/${log_file}" 2>&1 &
    local pid=$!
    cd - >/dev/null

    # 保存PID
    echo $pid > ../clotho/logs/${log_file}.pid

    # 等待一下让服务启动
    sleep 2

    # 检查服务是否成功启动
    if kill -0 $pid 2>/dev/null; then
        echo -e "${GREEN}✅ $service_name 启动成功 (PID: $pid)${NC}"
    else
        echo -e "${RED}❌ $service_name 启动失败${NC}"
        return 1
    fi
}

# 函数：停止服务
stop_service() {
    local service_name=$1
    local log_file=$2
    local pid_file="../clotho/logs/${log_file}.pid"

    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        echo -e "${YELLOW}停止 $service_name (PID: $pid)...${NC}"
        kill $pid 2>/dev/null
        sleep 1
        if kill -0 $pid 2>/dev/null; then
            kill -9 $pid 2>/dev/null
        fi
        rm -f "$pid_file"
        echo -e "${GREEN}✅ $service_name 已停止${NC}"
    fi
}

# 解析命令行参数
if [ "$1" = "stop" ]; then
    echo "停止所有服务..."
    stop_service "Custos" "custos.log"
    stop_service "Hermes" "hermes.log"
    stop_service "Appointments" "appointments.log"
    stop_service "Staff" "staff.log"
    stop_service "Plutus" "plutus.log"
    stop_service "Kratos" "kratos.log"
    stop_service "Items" "items.log"
    stop_service "Clotho" "clotho.log"
    echo -e "${GREEN}所有服务已停止${NC}"
    exit 0
fi

# 获取当前工作目录的父目录
BASE_DIR=$(pwd)/..
echo "基础目录: $BASE_DIR"

# 启动顺序
echo "按照依赖顺序启动服务..."
echo "========================"

# 1. 启动基础服务（数据库等）
echo -e "\n${BLUE}阶段 1: 基础服务${NC}"
echo -e "MySQL 应该已在 3306 端口运行"

# 2. 启动核心服务
echo -e "\n${BLUE}阶段 2: 核心服务${NC}"
start_service "Custos" "$BASE_DIR/custos" 9001 "custos.log"
start_service "Hermes" "$BASE_DIR/hermes" 8080 "hermes.log"

# 3. 启动业务服务
echo -e "\n${BLUE}阶段 3: 业务服务${NC}"
start_service "Appointments" "$BASE_DIR/appointments" 8083 "appointments.log"
start_service "Staff" "$BASE_DIR/staff" 8084 "staff.log"
start_service "Plutus" "$BASE_DIR/plutus" 8085 "plutus.log"
start_service "Kratos" "$BASE_DIR/kratos" 8082 "kratos.log"
start_service "Items" "$BASE_DIR/items" 8086 "items.log"

# 4. 启动 API 网关
echo -e "\n${BLUE}阶段 4: API 网关${NC}"
# 检查是否有端口冲突
if lsof -i :8080 >/dev/null 2>&1; then
    echo -e "${YELLOW}端口 8080 已被占用，检查是否是 Hermes...${NC}"
    if curl -s http://localhost:8080/health | grep -q "Hermes"; then
        echo -e "${RED}错误: Hermes 占用了 Clotho 的端口 (8080)${NC}"
        echo "请修改 clotho.yaml 中的端口配置"
        exit 1
    fi
fi

# 使用备用端口启动 Clotho
echo -e "${BLUE}启动 Clotho 在端口 8087（避免与 Hermes 冲突）...${NC}"
sed 's/port: "8080"/port: "8087"/' configs/clotho.yaml > configs/clotho_backup.yaml && mv configs/clotho_backup.yaml configs/clotho.yaml
nohup go run ./cmd/clotho serve --port 8087 > logs/clotho.log 2>&1 &
CLOTHO_PID=$!
echo $CLOTHO_PID > logs/clotho.log.pid

# 等待 Clotho 启动
sleep 3

# 验证所有服务
echo -e "\n${BLUE}验证服务状态...${NC}"
echo "========================"
./check_system_health.sh

# 显示服务信息
echo -e "\n${GREEN}=== 服务启动完成 ===${NC}"
echo -e "${BLUE}服务访问地址：${NC}"
echo -e "  - Clotho API 网关: http://localhost:8087"
echo -e "  - Health 检查: http://localhost:8087/health"
echo -e "  - API 文档: http://localhost:8087/swagger/index.html"
echo -e "\n${YELLOW}日志位置: ./logs/${NC}"
echo -e "${YELLOW}停止所有服务: ./start_all_services.sh stop${NC}"