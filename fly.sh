#!/bin/bash
# Fly Platform Service Manager

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="/tmp"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Services
SERVICES=("custos" "hermes" "kratos" "plutus" "clotho")

start() {
    echo -e "${GREEN}🚀 Starting Fly Platform...${NC}\n"

    # Start Docker infrastructure
    echo "📦 Starting infrastructure (MySQL, Redis)..."
    cd "$PROJECT_ROOT" && docker compose up -d mysql redis
    sleep 5

    # Start microservices
    for service in "${SERVICES[@]}"; do
        case $service in
            custos)
                port=8081
                cmd="cd \"$PROJECT_ROOT/custos\" && ./bin/userd serve --config configs/custos.yaml"
                ;;
            hermes)
                port=8083
                cmd="cd \"$PROJECT_ROOT/hermes\" && ./bin/hermes serve --config configs/hermes.yaml"
                ;;
            kratos)
                port=8084
                cmd="cd \"$PROJECT_ROOT/kratos\" && ./bin/kratos serve --config configs/kratos.yaml"
                ;;
            plutus)
                port=8085
                cmd="cd \"$PROJECT_ROOT/plutus\" && ./bin/plutus serve --config configs/plutus.yaml"
                ;;
            clotho)
                port=8080
                cmd="cd \"$PROJECT_ROOT/clotho\" && ./bin/clotho serve --config configs/clotho.yaml"
                ;;
        esac

        echo "🔧 Starting $service on port $port..."
        bash -c "$cmd" > "$LOG_DIR/${service}.log" 2>&1 &
        sleep 2
    done

    echo -e "\n${GREEN}✅ All services started!${NC}\n"
    status
}

stop() {
    echo -e "${YELLOW}🛑 Stopping Fly Platform...${NC}\n"

    # Stop microservices
    pkill -f 'userd|hermes|kratos|plutus|clotho' 2>/dev/null

    # Stop Docker infrastructure
    cd "$PROJECT_ROOT" && docker compose stop mysql redis

    echo -e "\n${GREEN}✅ All services stopped!${NC}"
}

restart() {
    stop
    sleep 2
    start
}

status() {
    echo -e "${GREEN}📊 Fly Platform Status:${NC}\n"

    # Check Docker
    echo "Infrastructure:"
    docker ps --format "  {{.Names}}: {{.Status}}" | grep -E "fly-mysql|fly-redis" || echo "  Not running"

    echo ""
    echo "Microservices:"

    # Check services
    check_service "Custos" 8081 "/api/v1/health"
    check_service "Hermes" 8083 "/health"
    check_service "Kratos" 8084 "/health"
    check_service "Plutus" 8082 "/health"
    check_service "Clotho" 8080 "/health"
}

check_service() {
    local name=$1
    local port=$2
    local endpoint=$3

    if curl -s "http://localhost:${port}${endpoint}" > /dev/null 2>&1; then
        echo -e "  ✅ ${name} (port ${port}): ${GREEN}RUNNING${NC}"
    else
        echo -e "  ❌ ${name} (port ${port}): ${RED}DOWN${NC}"
    fi
}

logs() {
    local service=$1
    if [ -z "$service" ]; then
        echo "Usage: $0 logs <service>"
        echo "Services: custos, hermes, kratos, plutus, clotho"
        exit 1
    fi

    if [ -f "$LOG_DIR/${service}.log" ]; then
        tail -f "$LOG_DIR/${service}.log"
    else
        echo "Log file not found: $LOG_DIR/${service}.log"
        exit 1
    fi
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    logs)
        logs "$2"
        ;;
    *)
        echo "Fly Platform Service Manager"
        echo ""
        echo "Usage: $0 {start|stop|restart|status|logs <service>}"
        echo ""
        echo "Commands:"
        echo "  start    - Start all services"
        echo "  stop     - Stop all services"
        echo "  restart  - Restart all services"
        echo "  status   - Show service status"
        echo "  logs     - Tail logs for a specific service"
        echo ""
        echo "Example:"
        echo "  $0 start"
        echo "  $0 logs custos"
        exit 1
esac
