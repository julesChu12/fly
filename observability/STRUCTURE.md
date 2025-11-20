# Observability 栈完整项目结构

```
observability/
├── README.md                           # 英文文档（95% 完成）
├── README_CN.md                        # 中文文档（95% 完成）
├── .gitignore                          # Git 忽略配置
├── start.sh                            # 快速启动脚本
│
├── prometheus/
│   ├── prometheus.yml                  # Prometheus 配置（已更新）
│   └── rules.yml                       # 告警规则
│
├── grafana/
│   └── provisioning/
│       ├── datasources/
│       │   └── prometheus.yaml         # 数据源配置
│       └── dashboards/
│           └── dashboards.yaml         # 仪表板配置
│
├── alertmanager/
│   ├── alertmanager.yml                # AlertManager 配置（生产环境填充密钥）
│   └── templates.tmpl                  # 告警通知模板
│
├── logstash/
│   ├── pipeline.conf                   # Logstash 管道配置
│   └── templates/
│       └── logs-template.json          # Elasticsearch 索引模板
│
├── kibana/
│   └── kibana.yml                      # Kibana 配置
│
└── [项目根目录]/
    ├── docker-compose.yml              # 基础配置（已更新）
    ├── docker-compose.dev.yml          # 开发环境配置
    ├── docker-compose.prod.yml         # 生产环境配置
    ├── .env.prod.example               # 生产环境变量示例
    └── .env.prod                       # 生产环境变量（需手动创建和填充）
```

---

## 📊 服务组成

### 指标收集和监控
- **Prometheus** (v2.50.0): 时间序列数据库和监控系统
- **Grafana** (10.2.2): 数据可视化和仪表板
- **mysqld_exporter** (0.15.0): MySQL 监控导出器
- **redis_exporter** (latest): Redis 监控导出器

### 日志聚合
- **Elasticsearch** (8.10.0): 分布式搜索和分析引擎
- **Logstash** (8.10.0): 日志处理和管道
- **Kibana** (8.10.0): 日志查询和可视化

### 告警和追踪
- **AlertManager** (0.26.0): 告警管理和通知
- **Jaeger** (1.53): 分布式链路追踪（来自 docker-compose.yml）

---

## 🚀 快速开始

### 开发环境（推荐）

```bash
# 进入项目目录
cd /Users/yt/Documents/developer/fly

# 方式 1: 使用快速启动脚本
bash observability/start.sh start-dev

# 方式 2: 手动启动
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# 访问服务
# Grafana: http://localhost:3000 (admin/grafana)
# Prometheus: http://localhost:9090
# Kibana: http://localhost:5601
# Jaeger: http://localhost:16686
# AlertManager: http://localhost:9093
```

### 生产环境

```bash
# 1. 配置生产环境变量
cp .env.prod.example .env.prod
vim .env.prod  # 填入实际的密码和服务器地址

# 2. 启动生产栈
bash observability/start.sh start-prod

# 或手动启动
source .env.prod
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

---

## 💾 数据持久化

### 开发环境
- **存储类型**: 本地 tmpfs（内存存储）
- **保留期**: 7 天
- **优点**: 快速、占用 SSD 少
- **适用**: 本地开发和测试

### 生产环境
- **存储类型**: NFS 共享存储
- **保留期**: 30-90 天
- **优点**: 高可用、数据持久化
- **配置**: 需设置 NFS_SERVER 环境变量

#### 卷清单

| 卷名 | 用途 | 大小 | 保留期 |
|------|------|------|--------|
| prometheus_data | Prometheus TSDB | 5GB-50GB | 7-90 天 |
| grafana_data | 仪表板和配置 | 1GB | 永久 |
| elasticsearch_data | 日志索引 | 10GB-100GB | 30 天 |
| alertmanager_data | 告警历史 | 100MB | 7 天 |

---

## 🔧 配置说明

### Prometheus 配置（prometheus.yml）

**抓取间隔**:
- 应用服务: 10 秒
- 导出器: 15-30 秒
- Jaeger: 30 秒

**保留策略**:
- 开发: 7 天
- 生产: 90 天

### Grafana 配置

**数据源**:
- Prometheus (默认)
- Elasticsearch / Loki
- Jaeger

**预配置仪表板**:
- 应用健康概览
- 服务性能分析
- 基础设施监控

### AlertManager 配置

**通知渠道**:
- 邮件（SMTP）
- 钉钉 / 企业微信（Webhook）

**预定义告警**:
- 服务宕机 (critical)
- 高错误率 (warning)
- 高响应时间 (warning)
- 数据库告警 (warning)

### Logstash 管道

**输入源**:
- TCP 端口 5000
- 文件日志
- Docker 日志

**处理逻辑**:
- JSON 解析
- 时间戳标准化
- 字段提取和丰富

**输出**:
- Elasticsearch 索引

---

## 📈 监控指标

### 应用指标
```promql
# HTTP 请求率
rate(http_requests_total[5m])

# 错误率
rate(http_requests_total{status=~"5.."}[5m])

# 响应时间 P95
histogram_quantile(0.95, http_request_duration_seconds_bucket)
```

### 基础设施指标
```promql
# MySQL 连接数
mysql_global_status_threads_connected

# Redis 内存使用率
redis_memory_used_bytes / redis_memory_max_bytes

# 磁盘使用率
100 - (node_filesystem_avail_bytes / node_filesystem_size_bytes * 100)
```

---

## 🛠️ 常见操作

### 查看服务状态
```bash
docker-compose ps
```

### 查看特定服务日志
```bash
# 使用脚本
bash observability/start.sh logs prometheus

# 或直接使用 docker
docker logs fly-prometheus -f
docker logs fly-grafana -f
docker logs fly-elasticsearch -f
```

### 重新加载配置
```bash
# Prometheus 重新加载规则
curl -X POST http://localhost:9090/-/reload

# Grafana 重新加载数据源
docker exec fly-grafana grafana-cli admin reload-provisioning-dashboards
```

### 清理数据
```bash
# 停止服务
docker-compose down

# 删除卷
docker volume rm fly_prometheus_data
docker volume rm fly_grafana_data
docker volume rm fly_elasticsearch_data

# 重新启动
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d
```

### 备份数据
```bash
# Prometheus 快照
docker exec fly-prometheus promtool tsdb snapshot /prometheus/snapshots/backup

# Grafana 导出仪表板
curl http://localhost:3000/api/dashboards/db/dashboard-name \
  -H "Authorization: Bearer $GRAFANA_TOKEN" > dashboard.json

# Elasticsearch 快照
curl -X PUT "localhost:9200/_snapshot/backup" -H 'Content-Type: application/json' -d'{
  "type": "fs",
  "settings": {
    "location": "/backups/elasticsearch"
  }
}'
```

---

## ⚙️ 故障排查

### Prometheus 无法连接到应用
```bash
# 检查应用是否暴露 /metrics
curl http://localhost:8081/metrics

# 查看 Prometheus 日志
docker logs fly-prometheus | grep -i error

# 验证网络
docker exec fly-prometheus ping hermes
```

### Elasticsearch 磁盘占用过大
```bash
# 删除老索引
curl -X DELETE "localhost:9200/logs-2025.09.*"

# 配置索引生命周期管理
curl -X PUT "localhost:9200/_ilm/policy/logs-policy" -H 'Content-Type: application/json' -d'{
  "policy": "logs-policy",
  "phases": {
    "delete": {
      "min_age": "30d",
      "actions": {"delete": {}}
    }
  }
}'
```

### 告警未发送
```bash
# 检查 AlertManager 状态
docker logs fly-alertmanager

# 验证邮件配置
docker exec fly-alertmanager amtool config routes

# 查看待发告警
curl http://localhost:9093/api/v1/alerts
```

---

## 📚 相关文档

- [README.md](./README.md) - 英文完整指南
- [README_CN.md](./README_CN.md) - 中文完整指南
- [Prometheus 文档](https://prometheus.io/docs/)
- [Grafana 文档](https://grafana.com/docs/grafana/)
- [Elasticsearch 文档](https://www.elastic.co/guide/en/elasticsearch/reference/)

---

## 📋 完成清单

- ✅ Prometheus 配置和告警规则
- ✅ Grafana 仪表板和数据源配置
- ✅ ELK Stack (Elasticsearch/Logstash/Kibana)
- ✅ AlertManager 邮件/Webhook 通知
- ✅ 开发环境配置 (docker-compose.dev.yml)
- ✅ 生产环境配置 (docker-compose.prod.yml)
- ✅ 数据持久化策略
- ✅ 快速启动脚本
- ✅ 完整文档（中英文）
- ✅ MySQL/Redis 导出器

---

**完成度**: 95% ✅
**最后更新**: 2025-10-28
**维护者**: Fly CRM 团队
