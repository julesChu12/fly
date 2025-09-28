# Clotho 可观测性指南

## 概述

Clotho 是 Fly 生态系统中的 API 编排层，集成了完整的可观测性解决方案，包括分布式追踪、指标收集、日志聚合和健康监控。

## 核心功能

### 1. API 编排与代理

- **Profile API**: 完整的用户资料管理 API
  - `GET /api/v1/profile` - 获取当前用户完整档案
  - `PUT /api/v1/profile` - 更新用户资料
  - `PUT /api/v1/profile/preferences` - 更新用户偏好设置
  - `GET /api/v1/profile/users/:id` - 获取其他用户公开资料

- **User API**: 基础用户信息 API
  - `GET /api/v1/users/me` - 获取当前用户基本信息
  - `GET /api/v1/users/:id` - 根据 ID 获取用户信息

### 2. 生产级中间件

#### 限流 (Rate Limiting)
- **全局限流**: 控制整体请求速率
- **IP 限流**: 防止单个 IP 滥用
- **用户限流**: 认证用户的精确控制
- **自动清理**: 定期清理空闲限流器

**配置示例**:
```yaml
rate_limiter:
  global_rps: 1000.0      # 全局每秒 1000 请求
  global_burst: 2000      # 全局突发容量
  per_ip_rps: 10.0        # 每 IP 每秒 10 请求
  per_ip_burst: 20        # 每 IP 突发容量
  per_user_rps: 100.0     # 每用户每秒 100 请求
  per_user_burst: 200     # 每用户突发容量
```

#### 熔断 (Circuit Breaker)
- **智能熔断**: 基于错误率和阈值的自动熔断
- **端点隔离**: 每个 API 端点独立熔断
- **半开状态**: 自动探测服务恢复
- **状态监控**: 实时监控熔断器状态

**配置示例**:
```yaml
circuit_breaker:
  max_requests: 5         # 半开状态最大请求数
  failure_threshold: 5    # 失败次数阈值
  failure_ratio: 0.6      # 失败率阈值
  min_requests: 10        # 最小评估请求数
```

### 3. 完整的可观测性

#### OpenTelemetry 分布式追踪
- **自动链路追踪**: HTTP 请求自动生成 trace
- **gRPC 调用追踪**: 内部服务调用链路
- **上下文传播**: 跨服务 trace 上下文传递
- **多种导出器**: 支持 OTLP、Jaeger、标准输出

**配置示例**:
```yaml
observability:
  service_name: "clotho"
  exporter_url: "http://localhost:4317"
  sample_ratio: 1.0
  tracing:
    enabled: true
    jaeger:
      enabled: true
      endpoint: "http://localhost:14268/api/traces"
```

#### Prometheus 指标收集
- **HTTP 指标**: 请求数、响应时间、状态码分布
- **业务指标**: 用户登录、资料更新、API 调用统计
- **基础设施指标**: 限流、熔断、gRPC 调用
- **自定义指标**: 支持业务特定指标

**核心指标**:
```
clotho_http_request_duration_seconds    # HTTP 请求耗时
clotho_http_requests_total              # HTTP 请求总数
clotho_grpc_request_duration_seconds    # gRPC 请求耗时
clotho_rate_limit_exceeded_total        # 限流超限次数
clotho_circuit_breaker_state            # 熔断器状态
clotho_user_logins_total                # 用户登录统计
clotho_profile_updates_total            # 资料更新统计
```

#### 实时监控端点
- `GET /api/v1/monitoring/stats` - 综合统计信息
- `GET /api/v1/monitoring/rate-limiter` - 限流器状态
- `GET /api/v1/monitoring/circuit-breaker` - 熔断器状态
- `POST /api/v1/monitoring/circuit-breaker/reset` - 重置熔断器

### 4. API 文档

#### Swagger/OpenAPI 自动生成
- **完整 API 文档**: 自动生成所有端点文档
- **交互式界面**: 支持在线测试 API
- **模型定义**: 详细的请求/响应模型
- **认证说明**: JWT Bearer Token 认证指导

访问地址: `GET /swagger/index.html`

## 快速开始

### 1. 服务启动

```bash
# 构建应用
go build -o clotho .

# 启动服务
./clotho serve
```

### 2. 健康检查

```bash
curl http://localhost:8080/health
```

**响应示例**:
```json
{
  "status": "healthy",
  "timestamp": "2023-10-01T10:00:00Z",
  "service": "clotho",
  "version": "0.1.0",
  "uptime": "1h30m45s"
}
```

### 3. 查看指标

```bash
# Prometheus 指标
curl http://localhost:8080/metrics

# 限流器状态
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/monitoring/rate-limiter

# 熔断器状态
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/monitoring/circuit-breaker
```

### 4. API 文档

在浏览器中访问: `http://localhost:8080/swagger/index.html`

## 可观测性栈集成

### Prometheus + Grafana

1. **Prometheus 配置** (`prometheus.yml`):
```yaml
scrape_configs:
  - job_name: 'clotho'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

2. **Grafana 仪表板**: 导入预配置的 Clotho 仪表板

### Jaeger 分布式追踪

1. **启动 Jaeger**:
```bash
docker run -d -p 16686:16686 -p 14268:14268 jaegertracing/all-in-one:latest
```

2. **配置追踪导出**:
```yaml
observability:
  tracing:
    enabled: true
    jaeger:
      enabled: true
      endpoint: "http://localhost:14268/api/traces"
```

3. **访问 Jaeger UI**: `http://localhost:16686`

### ELK 日志聚合

Clotho 使用结构化日志，便于 ELK 栈集成:

```json
{
  "level": "info",
  "time": "2023-10-01T10:00:00Z",
  "caller": "handler/profile.go:70",
  "msg": "HTTP request completed",
  "method": "GET",
  "path": "/api/v1/profile",
  "status": "200",
  "duration_ms": 45.2,
  "client_ip": "192.168.1.100",
  "trace_id": "abc123def456"
}
```

## 性能调优

### 限流器配置优化

```yaml
rate_limiter:
  # 根据服务容量调整全局限流
  global_rps: 5000.0      # 高性能服务器
  global_burst: 10000

  # 根据用户分级调整
  per_user_rps: 200.0     # VIP 用户更高限制
  per_user_burst: 400

  # 清理策略优化
  cleanup_interval: "2m"   # 更频繁清理
  max_idle_time: "5m"     # 更短空闲时间
```

### 熔断器敏感度调整

```yaml
circuit_breaker:
  # 保守策略 - 更容忍错误
  failure_threshold: 10
  failure_ratio: 0.8
  min_requests: 20

  # 激进策略 - 快速熔断
  failure_threshold: 3
  failure_ratio: 0.4
  min_requests: 5
```

## 监控告警

### 关键指标告警

1. **高错误率告警**:
```yaml
- alert: HighErrorRate
  expr: rate(clotho_http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 1m
  annotations:
    summary: "Clotho 服务错误率过高"
```

2. **限流触发告警**:
```yaml
- alert: RateLimitExceeded
  expr: rate(clotho_rate_limit_exceeded_total[5m]) > 10
  for: 30s
  annotations:
    summary: "Clotho 限流频繁触发"
```

3. **熔断器开启告警**:
```yaml
- alert: CircuitBreakerOpen
  expr: clotho_circuit_breaker_state == 2
  for: 0s
  annotations:
    summary: "Clotho 熔断器已开启"
```

## 故障排查

### 1. 请求链路追踪

通过 Jaeger UI 查看完整请求链路:
1. 访问 `http://localhost:16686`
2. 选择 `clotho` 服务
3. 根据时间范围搜索 trace
4. 分析请求在各个组件中的耗时

### 2. 指标分析

使用 Grafana 仪表板分析:
- **请求量趋势**: 识别流量模式
- **错误率分析**: 定位问题端点
- **响应时间分布**: 发现性能瓶颈
- **限流/熔断状态**: 了解保护机制触发情况

### 3. 日志聚合查询

在 Kibana 中使用查询:
```
# 查看特定用户的请求
user_id:12345 AND level:info

# 查看错误日志
level:error AND service:clotho

# 查看慢请求
duration_ms:>1000 AND path:/api/v1/profile
```

## 最佳实践

### 1. 指标命名约定

- 使用服务名前缀: `clotho_*`
- 描述性命名: `http_request_duration_seconds`
- 一致的标签: `method`, `path`, `status`

### 2. 追踪策略

- 生产环境使用采样: `sample_ratio: 0.1`
- 关键路径强制追踪
- 包含业务上下文信息

### 3. 日志结构化

- 使用 JSON 格式
- 包含 trace_id 关联
- 敏感信息脱敏

### 4. 监控层次化

- **基础设施层**: CPU、内存、网络
- **应用层**: HTTP 指标、业务指标
- **业务层**: 用户行为、转化率

## 扩展指南

### 添加自定义指标

```go
// 在 handler 中使用
middleware.RecordAPICall(c, "user_profile_update", true)

// 直接使用 registry
if registry, exists := c.Get("metrics"); exists {
    m := registry.(*observability.MetricsRegistry)
    m.IncrementCounter("custom_metric_total", "label_value")
}
```

### 集成新的追踪导出器

```go
// 在 Mora observability 包中扩展
func InitWithCustomExporter(cfg Config) (CleanupFunc, error) {
    // 添加自定义导出器支持
}
```

### 自定义中间件

```go
func CustomMetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 业务特定指标收集
        c.Next()
    }
}
```

通过这套完整的可观测性解决方案，Clotho 提供了生产级的监控、追踪和分析能力，帮助开发团队快速定位问题、优化性能并确保服务稳定性。