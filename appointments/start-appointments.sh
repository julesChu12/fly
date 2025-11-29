#!/bin/bash

# Appointments 服务启动脚本
# 使用方法: ./start-appointments.sh [选项]

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置
DEFAULT_HTTP_PORT="8083"
DEFAULT_GRPC_PORT="9083"
DEFAULT_CONFIG="configs/appointments.yaml"
DEFAULT_ENV="development"

# 显示帮助
show_help() {
    echo -e "${BLUE}Appointments 预约管理服务启动脚本${NC}"
    echo ""
    echo "使用方法:"
    echo "  $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示此帮助信息"
    echo "  -p, --http-port <port>  HTTP服务端口 (默认: $DEFAULT_HTTP_PORT)"
    echo "  -g, --grpc-port <port>  gRPC服务端口 (默认: $DEFAULT_GRPC_PORT)"
    echo "  -c, --config <path>     配置文件路径 (默认: $DEFAULT_CONFIG)"
    echo "  -e, --env <env>         运行环境 (默认: $DEFAULT_ENV)"
    echo "  -d, --dev              开发模式 (等同于 -e development)"
    echo "  --prod                  生产模式 (等同于 -e production)"
    echo "  --debug                调试模式 (显示详细日志)"
    echo "  --dry-run             显示将要执行的命令但不实际运行"
    echo "  --no-logs              禁用日志输出到文件"
    echo "  --stop                 停止运行中的服务"
    echo ""
    echo "示例:"
    echo "  $0                     # 使用默认配置启动"
    echo "  $0 -p 9083              # 指定HTTP端口"
    echo "  $0 -c local.yaml        # 使用本地配置文件"
    echo "  $0 --debug              # 调试模式启动"
    echo "  $0 --stop               # 停止服务"
    echo ""
}

# 解析命令行参数
HTTP_PORT="$DEFAULT_HTTP_PORT"
GRPC_PORT="$DEFAULT_GRPC_PORT"
CONFIG_FILE="$DEFAULT_CONFIG"
ENVIRONMENT="$DEFAULT_ENV"
DEV_MODE=false
PROD_MODE=false
DEBUG_MODE=false
DRY_RUN=false
NO_LOGS=false
STOP_SERVICE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -p|--http-port)
            HTTP_PORT="$2"
            shift 2
            ;;
        -g|--grpc-port)
            GRPC_PORT="$2"
            shift 2
            ;;
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -d|--dev)
            DEV_MODE=true
            ENVIRONMENT="development"
            shift
            ;;
        --prod)
            PROD_MODE=true
            ENVIRONMENT="production"
            shift
            ;;
        --debug)
            DEBUG_MODE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --no-logs)
            NO_LOGS=true
            shift
            ;;
        --stop)
            STOP_SERVICE=true
            shift
            ;;
        *)
            echo -e "${RED}错误: 未知选项 '$1'${NC}"
            echo "使用 $0 --help 查看帮助信息"
            exit 1
            ;;
    esac
done

# 停止服务
if [ "$STOP_SERVICE" = true ]; then
    echo -e "${BLUE}正在停止 Appointments 服务...${NC}"

    # 查找并停止进程
    APPOINTMENTS_PIDS=$(ps aux | grep "[a]ppointments.*serve" | grep -v grep | awk '{print $2}')

    if [ -n "$APPOINTMENTS_PIDS" ]; then
        echo -e "${YELLOW}找到以下进程:${NC}"
        echo "$APPOINTMENTS_PIDS"

        echo -n "正在停止..."
        echo "$APPOINTMENTS_PIDS" | xargs kill -TERM 2>/dev/null || true
        sleep 2

        # 如果进程仍在运行，强制杀死
        REMAINING=$(ps aux | grep "[a]ppointments.*serve" | grep -v grep | awk '{print $2}' | wc -l)
        if [ "$REMAINING" -gt 0 ]; then
            echo "强制停止剩余进程..."
            ps aux | grep "[a]ppointments.*serve" | grep -v grep | awk '{print $2}' | xargs kill -9 2>/dev/null || true
        fi

        echo -e "${GREEN}✅ Appointments 服务已停止${NC}"
    else
        echo -e "${YELLOW}⚠️ 没有找到运行中的 Appointments 服务${NC}"
    fi

    # 清理端口占用
    echo -n "检查端口占用..."
    for port in 8083 9083; do
        PID=$(lsof -ti :$port 2>/dev/null || true)
        if [ -n "$PID" ]; then
            kill -9 $PID 2>/dev/null || true
            echo -n "."
        fi
    done
    echo " ${GREEN}完成${NC}"

    exit 0
fi

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}错误: 配置文件 '$CONFIG_FILE' 不存在${NC}"
    echo -e "${YELLOW}提示: 使用 '$0 --help' 查看使用说明${NC}"
    exit 1
fi

# 检查端口是否被占用
check_port() {
    local port=$1
    local service_name=$2

    if netstat -an 2>/dev/null | grep ":$port " | grep -q "LISTEN"; then
        echo -e "${RED}错误: 端口 $port 已被占用 ($service_name)${NC}"
        echo -e "${YELLOW}请使用 'pkill -f \"appointments.*serve\"' 停止相关进程${NC}"
        exit 1
    fi
}

check_port $DEFAULT_HTTP_PORT "HTTP"
check_port $DEFAULT_GRPC_PORT "gRPC"

# 构建启动命令
CMD="go run ./cmd/appointments/main.go serve"

# 添加参数
if [ "$CONFIG_FILE" != "$DEFAULT_CONFIG" ]; then
    CMD="$CMD --config $CONFIG_FILE"
fi

if [ "$HTTP_PORT" != "$DEFAULT_HTTP_PORT" ]; then
    CMD="$CMD --http-port $HTTP_PORT"
fi

if [ "$GRPC_PORT" != "$DEFAULT_GRPC_PORT" ]; then
    CMD="$CMD --grpc-port $GRPC_PORT"
fi

# 添加环境变量
if [ "$ENVIRONMENT" != "$DEFAULT_ENV" ]; then
    CMD="APPOINTMENTS_ENV=$ENVIRONMENT $CMD"
fi

# 构建日志选项
LOG_FILE="logs/appointments_$(date +%Y%m%d_%H%M%S).log"
if [ "$NO_LOGS" = false ]; then
    CMD="$CMD > $LOG_FILE 2>&1 &"
else
    CMD="$CMD 2>&1 &"
fi

# 显示启动信息
echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}  Appointments 预约管理服务启动${NC}"
echo -e "${BLUE}================================${NC}"
echo ""
echo -e "${BLUE}配置信息:${NC}"
echo -e "  HTTP端口: ${GREEN}$HTTP_PORT${NC}"
echo -e "  gRPC端口: ${GREEN}$GRPC_PORT${NC}"
echo -e "  配置文件: ${GREEN}$CONFIG_FILE${NC}"
echo -e "  运行环境: ${GREEN}$ENVIRONMENT${NC}"
echo -e "  调试模式: ${GREEN}${DEBUG_MODE:-否}${NC}"
echo -e "  日志文件: ${GREEN}${LOG_FILE:-无}${NC}"
echo ""

if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW} Dry Run 模式 - 将要执行的命令:${NC}"
    echo -e "${YELLOW}$CMD${NC}"
    exit 0
fi

# 创建日志目录
mkdir -p logs

# 启动服务
echo -e "${BLUE}正在启动服务...${NC}"
echo -e "${YELLOW}执行命令: $CMD${NC}"
echo ""

# 执行命令
eval $CMD

# 获取进程ID
APPOINTMENTS_PID=$!
echo -e "${GREEN}✅ Appointments 服务已启动 (PID: $APPOINTMENTS_PID)${NC}"

# 等待服务启动
echo -n "等待服务启动..."
for i in {1..10}; do
    sleep 1
    echo -n "."
    if curl -s http://localhost:$HTTP_PORT/health >/dev/null 2>&1; then
        echo ""
        echo -e "${GREEN}✅ 服务启动成功!${NC}"
        break
    fi
done

echo ""
echo -e "${BLUE}服务访问地址:${NC}"
echo -e "  HTTP API: ${GREEN}http://localhost:$HTTP_PORT${NC}"
echo -e "  gRPC:   ${GREEN}localhost:$GRPC_PORT${NC}"
echo -e "  健康检查: ${GREEN}http://localhost:$HTTP_PORT/health${NC}"
echo ""
echo -e "${BLUE}常用命令:${NC}"
echo -e "  停止服务: ${YELLOW}$0 --stop${NC}"
echo -e "  查看日志: ${YELLOW}tail -f $LOG_FILE${NC}"
echo -e "  API文档: ${YELLOW}http://localhost:$HTTP_PORT/swagger/index.html${NC}"
echo ""

# 保存PID到文件
echo $APPOINTMENTS_PID > appointments.pid
echo -e "${BLUE}PID 已保存到: appointments.pid${NC}"