# 可观测性服务配置

## 当前状态

### ✅ 已启用的服务

#### Jaeger (分布式追踪)
- **状态**: 正常运行
- **访问地址**: http://localhost:16686
- **功能**: 分布式追踪和链路分析
- **端口**: 
  - 16686: Jaeger UI
  - 4317: OTLP gRPC
  - 4318: OTLP HTTP

### ⏸️ 暂时禁用的服务

#### Prometheus (指标监控)
- **状态**: 暂时禁用
- **原因**: 网络连接问题，无法拉取镜像
- **启用方法**: 
  1. 取消注释 `docker-compose.yaml` 中的 Prometheus 配置
  2. 运行 `docker-compose up prometheus -d`

## 配置文件

### prometheus.yml
已创建基本的 Prometheus 配置文件，包含以下监控目标：
- Prometheus 自身监控
- Custos 服务监控 (端口 8081)
- Clotho 服务监控 (端口 8080)
- Jaeger 监控 (端口 14269)

## 启用步骤

### 1. 启用 Prometheus
```bash
# 取消注释 docker-compose.yaml 中的 Prometheus 配置
# 然后运行：
docker-compose up prometheus -d
```

### 2. 验证服务
```bash
# 检查所有服务状态
docker-compose ps

# 访问 Jaeger UI
open http://localhost:16686

# 访问 Prometheus UI (启用后)
open http://localhost:9090
```

## 网络问题解决

如果遇到网络连接问题，可以尝试：

1. **使用镜像加速器**
   ```bash
   # 配置 Docker 镜像加速器
   # 编辑 ~/.docker/daemon.json
   {
     "registry-mirrors": [
       "https://docker.mirrors.ustc.edu.cn",
       "https://hub-mirror.c.163.com"
     ]
   }
   ```

2. **手动拉取镜像**
   ```bash
   # 在网络条件好的时候手动拉取
   docker pull prom/prometheus:latest
   docker pull jaegertracing/all-in-one:1.53
   ```

3. **使用代理**
   ```bash
   # 设置 Docker 代理
   export HTTP_PROXY=http://proxy:port
   export HTTPS_PROXY=http://proxy:port
   ```

## 监控配置

### 应用指标端点
确保应用暴露了 `/metrics` 端点：
- Custos: `http://custos:8081/metrics`
- Clotho: `http://clotho:8080/metrics`

### 自定义指标
可以在应用中添加自定义指标：
```go
// 示例：添加请求计数器
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
)
```
