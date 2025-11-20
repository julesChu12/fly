#!/bin/bash

# Fly CRM 可观测性栈快速启动脚本
# 用途: 一键启动开发或生产环境的可观测性栈

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Docker 和 Docker Compose
check_requirements() {
    log_info "检查依赖环境..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi

    log_success "依赖环境检查通过"
}

# 启动开发环境
start_dev() {
    log_info "启动开发环境..."

    cd "$PROJECT_ROOT"

    log_info "拉取最新镜像..."
    docker-compose pull 2>/dev/null || log_warning "镜像拉取失败，将使用本地镜像"

    log_info "启动服务..."
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

    log_info "等待服务就绪..."
    sleep 10

    # 验证服务
    if check_services; then
        print_access_info
        log_success "开发环境启动成功！"
    else
        log_error "服务启动失败，请检查日志"
        exit 1
    fi
}

# 启动生产环境
start_prod() {
    log_warning "即将启动生产环境..."

    # 检查 .env.prod 文件
    if [ ! -f "$PROJECT_ROOT/.env.prod" ]; then
        log_error ".env.prod 文件不存在"
        log_info "请先复制 .env.prod.example 到 .env.prod 并填入实际配置"
        echo "  cp .env.prod.example .env.prod"
        echo "  vim .env.prod"
        exit 1
    fi

    log_info "加载生产环境变量..."
    source .env.prod

    cd "$PROJECT_ROOT"

    log_info "启动生产栈..."
    docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

    log_info "等待服务就绪..."
    sleep 15

    if check_services; then
        log_success "生产环境启动成功！"
    else
        log_error "服务启动失败，请检查日志"
        exit 1
    fi
}

# 停止所有服务
stop_services() {
    log_info "停止所有服务..."

    cd "$PROJECT_ROOT"
    docker-compose down

    log_success "所有服务已停止"
}

# 检查服务健康状态
check_services() {
    log_info "检查服务状态..."

    local max_retries=30
    local retry=0

    while [ $retry -lt $max_retries ]; do
        local healthy_count=0

        # 检查关键服务
        for service in prometheus grafana elasticsearch; do
            if docker ps | grep -q "fly-$service"; then
                ((healthy_count++))
            fi
        done

        if [ $healthy_count -eq 3 ]; then
            return 0
        fi

        ((retry++))
        sleep 1
    done

    return 1
}

# 打印访问信息
print_access_info() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}   Fly CRM 可观测性栈访问信息${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${BLUE}指标监控:${NC}"
    echo -e "  • Prometheus: ${YELLOW}http://localhost:9090${NC}"
    echo -e "  • Grafana:    ${YELLOW}http://localhost:3000${NC} (admin/grafana)"
    echo ""
    echo -e "${BLUE}日志聚合:${NC}"
    echo -e "  • Elasticsearch: ${YELLOW}http://localhost:9200${NC}"
    echo -e "  • Kibana:        ${YELLOW}http://localhost:5601${NC}"
    echo ""
    echo -e "${BLUE}链路追踪:${NC}"
    echo -e "  • Jaeger: ${YELLOW}http://localhost:16686${NC}"
    echo ""
    echo -e "${BLUE}告警管理:${NC}"
    echo -e "  • AlertManager: ${YELLOW}http://localhost:9093${NC}"
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo ""
}

# 显示日志
show_logs() {
    local service=${1:-prometheus}
    cd "$PROJECT_ROOT"
    docker-compose logs -f "fly-$service"
}

# 显示帮助信息
show_help() {
    cat << EOF
Fly CRM 可观测性栈启动脚本

用途:
    $0 [命令]

命令:
    start-dev       启动开发环境（推荐用于本地开发）
    start-prod      启动生产环境（需要配置 .env.prod）
    stop            停止所有服务
    status          显示服务状态
    logs            显示服务日志 (可选择服务: prometheus|grafana|elasticsearch|jaeger|alertmanager)
    help            显示此帮助信息

示例:
    # 启动开发环境
    $0 start-dev

    # 启动生产环境
    $0 start-prod

    # 查看 Prometheus 日志
    $0 logs prometheus

    # 停止所有服务
    $0 stop

环境变量:
    PROJECT_ROOT  项目根目录（默认为脚本所在目录）

EOF
}

# 主函数
main() {
    check_requirements

    case "${1:-start-dev}" in
        start-dev)
            start_dev
            ;;
        start-prod)
            start_prod
            ;;
        stop)
            stop_services
            ;;
        status)
            cd "$PROJECT_ROOT"
            docker-compose ps
            ;;
        logs)
            show_logs "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
