# Observability 完整指南

**更新时间**: 2025-10-28
**完成度**: 95% ✅

## 📋 目录

1. [概述](#概述)
2. [架构设计](#架构设计)
3. [开发环境快速开始](#开发环境快速开始)
4. [生产环境部署](#生产环境部署)
5. [监控仪表板](#监控仪表板)
6. [告警管理](#告警管理)
7. [日志聚合](#日志聚合)
8. [常见问题](#常见问题)

---

## 概述

Fly CRM 可观测性栈提供了完整的三支柱监控解决方案：

| 柱子 | 工具 | 功能 |
|------|------|------|
| 📊 **指标** | Prometheus + Grafana | 实时性能监控和趋势分析 |
| 🔍 **追踪** | Jaeger | 分布式链路追踪 |
| 📝 **日志** | ELK Stack (Elasticsearch/Logstash/Kibana) | 日志聚合和分析 |

---

## 架构设计

### 完整架构图

```
┌─────────────────────────────────────────────────────────────┐
│                       Fly CRM 应用服务                        │
│  (Custos/Clotho/Hermes/Kratos/Plutus + MySQL/Redis)        │
└────────────┬──────────────────────────────┬────────────────┘
             │                              │
      指标收集 │                              │ 日志推送
             ▼                              ▼
    ┌─────────────────┐          ┌──────────────────┐
    │ Prometheus      │          │ Logstash         │
    │ (指标存储)      │          │ (日志处理)       │
    └────────┬────────┘          └────────┬─────────┘
             │                           │
             └─────────────┬─────────────┘
                           ▼
                  ┌──────────────────┐
                  │ Elasticsearch    │
                  │ (数据存储)       │
                  └────────┬─────────┘
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
   ┌─────────┐      ┌──────────┐      ┌──────────┐
   │Grafana  │      │ Kibana   │      │Jaeger UI │
   │(可视化) │      │(查询)    │      │(追踪)    │
   └─────────┘      └──────────┘      └──────────┘
        │                  │                │
        └──────────────────┼────────────────┘
                           ▼
                  ┌──────────────────┐
                  │ AlertManager     │
                  │ (告警管理)       │
                  └────────┬─────────┘
                           │
                  ┌────────▼────────┐
                  │ 邮件/钉钉/企业微信 │
                  └──────────────────┘
```

### 服务清单

| 服务 | 版本 | 端口 | 功能 |
|------|------|------|------|
| Prometheus | v2.50.0 | 9090 | 指标收集和存储 |
| Grafana | 10.2.2 | 3000 | 监控仪表板和数据可视化 |
| Elasticsearch | 8.10.0 | 9200 | 日志存储和搜索 |
| Logstash | 8.10.0 | 5000 | 日志收集和处理 |
| Kibana | 8.10.0 | 5601 | 日志查询和可视化 |
| Jaeger | 1.53 | 16686/4317 | 分布式追踪 |
| AlertManager | 0.26.0 | 9093 | 告警管理和通知 |
| mysqld_exporter | 0.15.0 | 9104 | MySQL 监控 |
| redis_exporter | latest | 9121 | Redis 监控 |

---

## 开发环境快速开始

### 1️⃣ 启动完整栈

```bash
# 进入项目目录
cd /Users/yt/Documents/developer/fly

# 启动开发环境（包含所有可观测性服务）
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# 查看服务状态
docker-compose ps
```

### 2️⃣ 访问各个服务

| 服务 | URL | 用户名 | 密码 |
|------|-----|--------|------|
| Grafana | http://localhost:3000 | admin | grafana |
| Prometheus | http://localhost:9090 | - | - |
| Jaeger | http://localhost:16686 | - | - |
| Kibana | http://localhost:5601 | - | - |
| AlertManager | http://localhost:9093 | - | - |

### 3️⃣ 验证数据流

#### 检查 Prometheus 指标抓取

```bash
# 查看 Prometheus 是否正常抓取指标
curl http://localhost:9090/api/v1/targets

# 查询某个服务的指标
curl 'http://localhost:9090/api/v1/query?query=up'
```

#### 检查 Jaeger 追踪

```bash
# 生成一些请求以创建追踪
curl http://localhost:8080/api/v1/users/me

# 访问 Jaeger UI
open http://localhost:16686
```

#### 检查日志收集

```bash
# 查看 Elasticsearch 索引
curl http://localhost:9200/_cat/indices

# 查询日志
curl 'http://localhost:9200/logs-*/_search?q=level:error'
```

### 4️⃣ 导入预制仪表板

开发环境已预配置了以下仪表板模板：

```
Grafana 自动导入:
├── Node Exporter - Nodes
├── MySQL Exporter
├── Redis Exporter
├── Jaeger Monitoring
└── Application Health Overview
```

---

## 生产环境部署

### 1️⃣ 环境准备

```bash
# 复制生产环境配置
cp .env.prod.example .env.prod

# 编辑配置文件，填入实际的密码和服务器地址
vim .env.prod

# 必填项：
# - GRAFANA_PASSWORD
# - ELASTICSEARCH_PASSWORD
# - KIBANA_PASSWORD
# - SMTP_HOST/SMTP_USER/SMTP_PASSWORD
# - NFS_SERVER（用于共享存储）
```

### 2️⃣ 启动生产环境

```bash
# 加载生产环境变量
source .env.prod

# 启动生产栈
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# 验证服务健康状态
docker-compose ps
```

### 3️⃣ 数据持久化配置

**生产环境使用 NFS 共享存储确保数据持久化：**

```bash
# NFS 服务器配置示例
# /etc/exports
/prometheus        *(rw,sync,no_subtree_check)
/grafana           *(rw,sync,no_subtree_check)
/elasticsearch     *(rw,sync,no_subtree_check)
/alertmanager      *(rw,sync,no_subtree_check)

# 重载 NFS 配置
exportfs -ra
```

**Docker 卷配置（自动）：**

- `prometheus_data`: 30-90 天的 TSDB 数据
- `grafana_data`: 仪表板和用户配置
- `elasticsearch_data`: 日志索引和数据
- `alertmanager_data`: 告警历史记录

### 4️⃣ 备份策略

```bash
#!/bin/bash
# 每日备份脚本

BACKUP_DIR="/backups/fly-crm"
DATE=$(date +%Y%m%d_%H%M%S)

# 备份 Prometheus 数据
docker exec fly-prometheus promtool tsdb snapshot /prometheus/snapshots/$DATE
tar czf $BACKUP_DIR/prometheus_$DATE.tar.gz /var/lib/docker/volumes/fly_prometheus_data/_data/

# 备份 Grafana 数据库
docker exec fly-grafana grafana-cli admin export-dashboard -s > $BACKUP_DIR/grafana_dashboards_$DATE.json

# 备份 Elasticsearch
curl -X PUT "localhost:9200/_snapshot/backup" -H 'Content-Type: application/json' -d'{
  "type": "fs",
  "settings": {
    "location": "'$BACKUP_DIR/elasticsearch'"
  }
}'

curl -X PUT "localhost:9200/_snapshot/backup/snapshot_$DATE?wait_for_completion=true"
```

---

## 监控仪表板

### Grafana 仪表板

#### 1. 应用健康概览

**路径**: 首页 → 应用健康概览

**关键指标**:
- 服务可用性 (up)
- 请求成功率 (request success rate)
- P95 响应时间
- 错误率趋势

#### 2. 应用性能监控

**关键指标**:
```promql
# HTTP 请求速率
rate(http_requests_total[5m])

# 错误率
rate(http_requests_total{status=~"5.."}[5m])

# 响应时间分布
histogram_quantile(0.95, http_request_duration_seconds_bucket)
```

#### 3. 基础设施监控

**MySQL**:
```
- 连接数: mysql_global_status_threads_connected
- 查询速率: rate(mysql_global_status_questions[5m])
- 慢查询: rate(mysql_global_status_slow_queries[5m])
- 缓存命中率: 100 - (mysql_global_status_innodb_buffer_pool_reads / (mysql_global_status_innodb_buffer_pool_reads + mysql_global_status_innodb_buffer_pool_read_requests) * 100)
```

**Redis**:
```
- 连接数: redis_connected_clients
- 内存使用: redis_memory_used_bytes / redis_memory_max_bytes
- 命中率: redis_keyspace_hits_total / (redis_keyspace_hits_total + redis_keyspace_misses_total)
- 操作速率: rate(redis_commands_processed_total[5m])
```

#### 4. 自定义仪表板创建

```bash
# 访问 Grafana
http://localhost:3000

# 步骤：
1. 点击 "+" → "Dashboard"
2. 选择 "Add new panel"
3. 选择数据源 (Prometheus/Elasticsearch/Jaeger)
4. 编写 PromQL 查询
5. 保存仪表板
```

---

## 告警管理

### 告警规则

告警规则定义在 `observability/prometheus/rules.yml`

#### 预定义告警

| 告警名称 | 条件 | 严重级别 |
|---------|------|---------|
| ServiceDown | up == 0 持续 1 分钟 | 🔴 Critical |
| HighErrorRate | 错误率 > 5% 持续 5 分钟 | 🟡 Warning |
| HighResponseTime | P95 > 1s 持续 5 分钟 | 🟡 Warning |
| MySQLHighConnections | 连接数 > 80 持续 5 分钟 | 🟡 Warning |
| RedisHighMemoryUsage | 内存使用率 > 80% 持续 5 分钟 | 🟡 Warning |

### 告警通知配置

#### 邮件通知设置

编辑 `observability/alertmanager/alertmanager.yml`:

```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_auth_username: 'your-email@gmail.com'
  smtp_auth_password: 'your-app-password'
  smtp_from: 'alerts@fly-crm.local'
```

#### 钉钉/企业微信集成

```yaml
receivers:
  - name: 'critical'
    webhook_configs:
      - url: 'http://webhook-receiver:5000/dingtalk'
        send_resolved: true
```

### 告警测试

```bash
# 测试告警规则
docker exec fly-prometheus promtool check rules /etc/prometheus/rules.yml

# 手动触发告警（停止某个服务）
docker-compose down hermes

# 查看告警状态
curl http://localhost:9090/api/v1/alerts

# 查看 AlertManager 告警
curl http://localhost:9093/api/v1/alerts
```

---

## 日志聚合

### Kibana 日志查询

#### 1. 创建索引模式

```bash
# 访问 Kibana
http://localhost:5601

# 步骤：
1. 点击 "Discover" 左侧的齿轮
2. 选择 "Index Patterns"
3. 点击 "Create index pattern"
4. 输入 "logs-*"
5. 选择 "@timestamp" 作为时间字段
```

#### 2. 查询日志

**示例查询**:

```
# 查询所有错误日志
level: "ERROR"

# 查询特定服务的日志
service: "hermes"

# 查询特定时间范围
@timestamp: [2025-10-28 TO 2025-10-29]

# 组合查询
service: "kratos" AND level: "ERROR" AND status_code: [500 TO 599]
```

#### 3. 创建仪表板

```
1. 点击 "Discover"
2. 应用过滤器
3. 点击 "Save"
4. 选择 "Save as dashboard panel"
5. 给面板命名并保存
```

### 日志来源

应用通过以下方式发送日志：

```go
// 应用日志配置示例
logger := zapLogger.With(
    zap.String("service", "hermes"),
    zap.String("version", "1.0.0"),
)

// 结构化日志格式
logger.Info("User login successful",
    zap.String("user_id", "123"),
    zap.String("method", "POST"),
    zap.String("path", "/api/auth/login"),
    zap.Int("status_code", 200),
    zap.Float64("response_time_ms", 123.45),
)
```

---

## 常见问题

### Q1: Prometheus 无法连接到应用服务怎么办？

**检查步骤**:

```bash
# 1. 检查应用是否暴露了 /metrics 端点
curl http://localhost:8081/metrics

# 2. 检查 Prometheus 配置是否正确
docker exec fly-prometheus cat /etc/prometheus/prometheus.yml

# 3. 查看 Prometheus 日志
docker logs fly-prometheus | grep -i error

# 4. 验证网络连接
docker exec fly-prometheus ping hermes
```

### Q2: Elasticsearch 磁盘空间占用过大怎么办？

**解决方案**:

```bash
# 1. 删除旧索引
curl -X DELETE "localhost:9200/logs-2025.10.01"

# 2. 配置索引生命周期管理 (ILM)
curl -X PUT "localhost:9200/_ilm/policy/logs-policy" -H 'Content-Type: application/json' -d'{
  "policy": "logs-policy",
  "phases": {
    "delete": {
      "min_age": "30d",
      "actions": {
        "delete": {}
      }
    }
  }
}'

# 3. 查看磁盘使用情况
curl "localhost:9200/_cat/allocation?v"
```

### Q3: 告警没有发送怎么办？

**检查步骤**:

```bash
# 1. 检查 AlertManager 是否正常运行
docker logs fly-alertmanager

# 2. 验证告警规则是否被正确加载
curl http://localhost:9090/api/v1/rules

# 3. 手动测试邮件配置
docker exec fly-alertmanager amtool config routes

# 4. 查看告警历史
curl http://localhost:9093/api/v1/alerts?silenced=false
```

### Q4: 开发环境和生产环境有什么区别？

**关键差异**:

| 方面 | 开发环境 | 生产环境 |
|------|---------|---------|
| 数据保留 | 7 天 | 90 天 |
| 存储 | 本地 tmpfs | NFS 共享 |
| 内存配置 | 128MB-512MB | 1GB-2GB |
| 匿名访问 | 允许 | 禁用 |
| 日志级别 | debug | info |
| 备份 | 否 | 是 |

### Q5: 如何自定义告警规则？

**步骤**:

```bash
# 1. 编辑规则文件
vim observability/prometheus/rules.yml

# 2. 添加新规则
- alert: MyCustomAlert
  expr: your_metric > 100
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Custom alert triggered"

# 3. 重新加载 Prometheus
curl -X POST http://localhost:9090/-/reload

# 4. 验证规则
curl http://localhost:9090/api/v1/rules
```

---

## 快速命令参考

```bash
# 启动开发环境
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# 启动生产环境
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# 停止所有服务
docker-compose down

# 查看特定服务日志
docker logs fly-prometheus -f

# 进入容器
docker exec -it fly-prometheus /bin/sh

# 手动触发告警测试
docker-compose down hermes && sleep 60 && docker-compose up hermes -d

# 清理过期数据
docker system prune -a

# 备份 Grafana 仪表板
curl http://localhost:3000/api/dashboards/db/my-dashboard -H "Authorization: Bearer $GRAFANA_TOKEN" > dashboard.json

# 导入仪表板
curl -X POST http://localhost:3000/api/dashboards/db -H "Content-Type: application/json" -d @dashboard.json
```

---

## 相关文档

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Grafana 官方文档](https://grafana.com/docs/grafana/)
- [Jaeger 官方文档](https://www.jaegertracing.io/docs/)
- [Elasticsearch 官方文档](https://www.elastic.co/guide/en/elasticsearch/reference/)
- [AlertManager 官方文档](https://prometheus.io/docs/alerting/latest/alertmanager/)

---

**文档完成度**: 95% ✅
**最后更新**: 2025-10-28
**维护者**: Fly CRM 团队
