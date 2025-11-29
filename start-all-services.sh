#!/bin/bash

# Fly 微服务统一启动脚本
# 启动所有核心服务：Custos, Staff, Plutus, Kratos, Appointments

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 显示帮助
show_help() {
    echo -e "${BLUE}Fly 微服务统一启动脚本${NC}"
    echo ""
    echo "使用方法:"
    echo "  $0 [选项] [服务名...]"
    echo ""
    echo "服务列表:"
    echo "  - custos: Custos认证服务 (HTTP:8081, gRPC:9001)"
    echo "  - staff: Staff员工服务 (HTTP:8084, gRPC:9084)"
    echo "  - plutus: Plutus支付服务 (HTTP:8085, gRPC:9085)"
    echo "  - kratos: Kratos订单服务 (HTTP:8082, gRPC:9092)"
    echo "  - appointments: Appointments预约服务 (HTTP:8083, gRPC:9083)"
    echo ""
    echo "选项:"
    echo "  -h, --help         显示此帮助信息"
    echo "  -s, --sequential   顺序启动服务（默认并行）"
    echo "  -t, --timeout N    每个服务启动超时时间（秒，默认10）"
    echo "  -d, --dev          开发模式启动"
    echo "  --stop             停止所有服务"
    echo "  --status           查看服务状态"
    echo "  --check            检查服务健康状态"
    echo "  --dry-run          显示将要执行的命令但不实际运行"
    echo ""
    echo "示例:"
    echo "  $0                           # 启动所有服务"
    echo "  $0 custos staff             # 只启动指定服务"
    echo "  $0 -s                        # 顺序启动所有服务"
    echo "  $0 -t 5                      # 设置5秒超时"
    echo "  $0 --stop                    # 停止所有服务"
    echo "  $0 --status                  # 查看服务状态"
    echo ""
}

# 解析命令行参数
SEQUENTIAL=false
TIMEOUT=10
DEV_MODE=false
STOP_SERVICES=false
CHECK_STATUS=false
CHECK_HEALTH=false
DRY_RUN=false
TARGET_SERVICES=()

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -s|--sequential)
            SEQUENTIAL=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -d|--dev)
            DEV_MODE=true
            shift
            ;;
        --stop)
            STOP_SERVICES=true
            shift
            ;;
        --status)
            CHECK_STATUS=true
            shift
            ;;
        --check)
            CHECK_HEALTH=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -*)
            echo -e "${RED}错误: 未知选项 '$1'${NC}"
            echo "使用 $0 --help 查看帮助信息"
            exit 1
            ;;
        *)
            TARGET_SERVICES+=("$1")
            shift
            ;;
    esac
done

# 如果没有指定服务，则启动所有服务
if [ ${#TARGET_SERVICES[@]} -eq 0 ]; then
    TARGET_SERVICES=("custos" "staff" "plutus" "kratos" "appointments")
fi

# 检查服务目录是否存在
check_service_directory() {
    local service=$1
    local service_dir="/Users/yt/Documents/developer/fly/$service"

    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}错误: 服务目录不存在 $service_dir${NC}"
        return 1
    fi
}

# 停止服务
stop_service() {
    local service=$1
    echo -n -e "${YELLOW}停止 $service 服务...${NC} "

    # 查找并停止相关进程
    case $service in
        "custos")
            pkill -f "userd.*serve" 2>/dev/null || true
            ;;
        "staff")
            pkill -f "staff.*serve" 2>/dev/null || true
            ;;
        "plutus")
            pkill -f "plutus.*serve" 2>/dev/null || true
            ;;
        "kratos")
            pkill -f "kratos.*serve" 2>/dev/null || true
            ;;
        "appointments")
            pkill -f "appointments.*serve" 2>/dev/null || true
            ;;
    esac

    sleep 1
    echo -e "${GREEN}✓${NC}"
}

# 停止所有服务
stop_all_services() {
    echo -e "${BLUE}正在停止所有 Fly 服务...${NC}"

    for service in "${TARGET_SERVICES[@]}"; do
        if check_service_directory "$service"; then
            stop_service "$service"
        fi
    done

    # 清理端口占用
    echo -n -e "${YELLOW}清理端口占用...${NC} "
    for port in 8081 8082 8083 8084 8085 9001 9002 9003 9083 9084 9085 9092; do
        PID=$(lsof -ti :$port 2>/dev/null || true)
        if [ -n "$PID" ]; then
            kill -9 $PID 2>/dev/null || true
        fi
    done
    echo -e "${GREEN}✓${NC}"

    echo -e "${GREEN}✅ 所有服务已停止${NC}"
}

# 检查服务状态
check_service_status() {
    local service=$1
    local name=""
    local grpc_port=""
    local http_port=""

    # 获取服务配置
    case $service in
        "custos")
            name="Custos认证服务"
            grpc_port="9001"
            http_port="8081"
            ;;
        "staff")
            name="Staff员工服务"
            grpc_port="9084"
            http_port="8084"
            ;;
        "plutus")
            name="Plutus支付服务"
            grpc_port="9085"
            http_port="8085"
            ;;
        "kratos")
            name="Kratos订单服务"
            grpc_port="9092"
            http_port="8082"
            ;;
        "appointments")
            name="Appointments预约服务"
            grpc_port="9083"
            http_port="8083"
            ;;
    esac

    # 检查进程
    local running=false
    case $service in
        "custos")
            ps aux | grep -q "[u]serd.*serve" && running=true
            ;;
        "staff")
            ps aux | grep -q "[s]taff.*serve" && running=true
            ;;
        "plutus")
            ps aux | grep -q "[p]lutus.*serve" && running=true
            ;;
        "kratos")
            ps aux | grep -q "[k]ratos.*serve" && running=true
            ;;
        "appointments")
            ps aux | grep -q "[a]ppointments.*serve" && running=true
            ;;
    esac

    # 检查端口
    local http_running=false
    local grpc_running=false

    if lsof -ti :$http_port >/dev/null 2>&1; then
        http_running=true
    fi

    if lsof -ti :$grpc_port >/dev/null 2>&1; then
        grpc_running=true
    fi

    # 显示状态
    if $running; then
        echo -e "${GREEN}● $service${NC} - 运行中"
    else
        echo -e "${RED}● $service${NC} - 已停止"
    fi

    echo -e "  HTTP端口 $http_port: ${http_running:+${GREEN}已占用${NC}}${http_running:-${RED}空闲${NC}}"
    echo -e "  gRPC端口 $grpc_port: ${grpc_running:+${GREEN}已占用${NC}}${grpc_running:-${RED}空闲${NC}}"
}

# 检查服务健康状态
check_service_health() {
    local service=$1
    local http_port=""

    # 获取服务HTTP端口
    case $service in
        "custos")
            http_port="8081"
            ;;
        "staff")
            http_port="8084"
            ;;
        "plutus")
            http_port="8085"
            ;;
        "kratos")
            http_port="8082"
            ;;
        "appointments")
            http_port="8083"
            ;;
    esac

    # 尝试访问健康检查端点
    local health_url="http://localhost:$http_port"
    case $service in
        "custos")
            health_url="$health_url/api/v1/health"
            ;;
        "staff"|"plutus"|"kratos"|"appointments")
            health_url="$health_url/health"
            ;;
    esac

    echo -n -e "检查 $service 健康状态... "

    if curl -s --max-time 3 "$health_url" >/dev/null 2>&1; then
        echo -e "${GREEN}健康${NC}"
        return 0
    else
        echo -e "${RED}不健康${NC}"
        return 1
    fi
}

# 启动服务
start_service() {
    local service=$1
    local service_dir="/Users/yt/Documents/developer/fly/$service"

    echo -e "${CYAN}启动 $service 服务...${NC}"

    if [ "$DRY_RUN" = true ]; then
        echo -e "${YELLOW}[Dry Run] cd $service_dir && 启动服务${NC}"
        return 0
    fi

    # 根据不同服务执行不同的启动命令
    cd "$service_dir"
    case $service in
        "custos")
            if [ -f "start-custos.sh" ]; then
                ./start-custos.sh ${DEV_MODE:+--debug} >/dev/null 2>&1 &
            else
                nohup go run ./cmd/userd/main.go serve > logs/custos_$(date +%Y%m%d_%H%M%S).log 2>&1 &
            fi
            ;;
        "staff")
            if [ -f "start-staff.sh" ]; then
                ./start-staff.sh ${DEV_MODE:+--debug} >/dev/null 2>&1 &
            else
                nohup go run ./cmd/staff/main.go serve > logs/staff_$(date +%Y%m%d_%H%M%S).log 2>&1 &
            fi
            ;;
        "plutus")
            mkdir -p logs
            nohup go run ./cmd/plutus serve > logs/plutus_$(date +%Y%m%d_%H%M%S).log 2>&1 &
            ;;
        "kratos")
            if [ -f "start-kratos.sh" ]; then
                ./start-kratos.sh ${DEV_MODE:+--debug} >/dev/null 2>&1 &
            else
                mkdir -p logs
                nohup go run ./cmd/kratos serve > logs/kratos_$(date +%Y%m%d_%H%M%S).log 2>&1 &
            fi
            ;;
        "appointments")
            if [ -f "start-appointments.sh" ]; then
                ./start-appointments.sh ${DEV_MODE:+--debug} >/dev/null 2>&1 &
            else
                mkdir -p logs
                nohup go run ./cmd/appointments/main.go serve > logs/appointments_$(date +%Y%m%d_%H%M%S).log 2>&1 &
            fi
            ;;
    esac

    # 等待服务启动
    local count=0
    while [ $count -lt $TIMEOUT ]; do
        if check_service_health "$service" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ $service 启动成功${NC}"
            return 0
        fi
        sleep 1
        count=$((count + 1))
        echo -n "."
    done

    echo ""
    echo -e "${RED}❌ $service 启动超时${NC}"
    return 1
}

# 主执行逻辑
main() {
    # 检查基础依赖
    echo -e "${BLUE}检查基础依赖...${NC}"

    # 检查 MySQL
    if docker ps | grep -q "mysql"; then
        echo -e "${GREEN}✓ MySQL 运行中${NC}"
    else
        echo -e "${RED}✗ MySQL 未运行${NC}"
    fi

    # 检查 Redis
    if docker ps | grep -q "redis"; then
        echo -e "${GREEN}✓ Redis 运行中${NC}"
    else
        echo -e "${RED}✗ Redis 未运行${NC}"
    fi

    echo ""

    # 执行相应操作
    if [ "$STOP_SERVICES" = true ]; then
        stop_all_services
        exit 0
    fi

    if [ "$CHECK_STATUS" = true ]; then
        echo -e "${BLUE}服务状态:${NC}"
        echo ""
        for service in "${TARGET_SERVICES[@]}"; do
            if check_service_directory "$service"; then
                check_service_status "$service"
                echo ""
            fi
        done
        exit 0
    fi

    if [ "$CHECK_HEALTH" = true ]; then
        echo -e "${BLUE}服务健康状态:${NC}"
        echo ""
        for service in "${TARGET_SERVICES[@]}"; do
            if check_service_directory "$service"; then
                check_service_health "$service"
            fi
        done
        echo ""
        exit 0
    fi

    # 启动服务
    echo -e "${BLUE}启动以下服务: ${CYAN}${TARGET_SERVICES[*]}${NC}"
    echo ""

    if [ "$SEQUENTIAL" = true ]; then
        # 顺序启动
        for service in "${TARGET_SERVICES[@]}"; do
            if check_service_directory "$service"; then
                start_service "$service"
                echo ""
            fi
        done
    else
        # 并行启动
        echo -e "${CYAN}并行启动服务...${NC}"

        # 启动 Custos 和基础服务先（依赖关系）
        for service in "${TARGET_SERVICES[@]}"; do
            if check_service_directory "$service"; then
                start_service "$service" &
            fi
        done

        # 等待所有服务启动完成
        wait

        echo ""
    fi

    # 最终状态检查
    echo -e "${BLUE}最终服务状态:${NC}"
    echo ""
    for service in "${TARGET_SERVICES[@]}"; do
        if check_service_directory "$service"; then
            check_service_status "$service"
            echo ""
        fi
    done

    # 显示访问地址
    echo -e "${BLUE}服务访问地址:${NC}"
    echo ""
    for service in "${TARGET_SERVICES[@]}"; do
        case $service in
            "custos")
                echo -e "  ${CYAN}Custos认证服务${NC}:"
                echo -e "    HTTP: http://localhost:8081"
                echo -e "    gRPC: localhost:9001"
                ;;
            "staff")
                echo -e "  ${CYAN}Staff员工服务${NC}:"
                echo -e "    HTTP: http://localhost:8084"
                echo -e "    gRPC: localhost:9084"
                ;;
            "plutus")
                echo -e "  ${CYAN}Plutus支付服务${NC}:"
                echo -e "    HTTP: http://localhost:8085"
                echo -e "    gRPC: localhost:9085"
                ;;
            "kratos")
                echo -e "  ${CYAN}Kratos订单服务${NC}:"
                echo -e "    HTTP: http://localhost:8082"
                echo -e "    gRPC: localhost:9092"
                ;;
            "appointments")
                echo -e "  ${CYAN}Appointments预约服务${NC}:"
                echo -e "    HTTP: http://localhost:8083"
                echo -e "    gRPC: localhost:9083"
                ;;
        esac
    done
    echo ""
}

# 执行主逻辑
main