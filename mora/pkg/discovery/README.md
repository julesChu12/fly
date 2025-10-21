# Mora 服务发现使用文档

## 概述

Mora 提供了统一的服务发现抽象层，支持从环境变量到 Consul 的平滑演进，为微服务间通信提供灵活的地址解析能力。

## 核心特性

- ✅ **统一接口**：`Discovery` 接口统一所有服务发现实现
- ✅ **多种实现**：支持环境变量、Consul（未来可扩展 etcd、Kubernetes 等）
- ✅ **负载均衡**：内置轮询和随机负载均衡策略
- ✅ **健康检查**：自动过滤不健康的服务实例
- ✅ **平滑演进**：从环境变量无缝切换到 Consul
- ✅ **配置化**：通过配置文件选择服务发现类型

---

## 快速开始

### 1. 环境变量模式（当前推荐）

**适用场景：** 开发环境、Docker Compose 部署

```go
import "github.com/julesChu12/fly/mora/pkg/discovery"

// 创建环境变量服务发现实例
disc := discovery.NewEnvDiscovery()

// 获取服务地址
instance, err := disc.GetService(context.Background(), "custos")
if err != nil {
    log.Fatal(err)
}

fmt.Println(instance.Address()) // 输出: custos:9001
```

**环境变量配置：**

```bash
# 标准格式（推荐）
export CUSTOS_ADDRESS=custos:9001
export HERMES_ADDRESS=hermes:9080

# 兼容格式
export CUSTOS_GRPC_ADDRESS=custos:9001

# 分离格式
export CUSTOS_HOST=custos
export CUSTOS_PORT=9001
```

---

### 2. Consul 模式（生产推荐）

**适用场景：** 生产环境、多实例部署、动态扩缩容

```go
import "github.com/julesChu12/fly/mora/pkg/discovery"

// 创建 Consul 服务发现实例
consulCfg := &discovery.ConsulConfig{
    Address:    "127.0.0.1:8500",
    Datacenter: "dc1",
}

disc, err := discovery.NewConsulDiscovery(consulCfg)
if err != nil {
    log.Fatal(err)
}
defer disc.Close()

// 服务注册
instance := &discovery.ServiceInstance{
    ID:   "custos-1",
    Name: "custos",
    Host: "192.168.1.10",
    Port: 9001,
}
disc.Register(context.Background(), instance)

// 服务发现（自动负载均衡）
instance, err := disc.GetService(context.Background(), "custos")
fmt.Println(instance.Address())
```

---

### 3. 配置化使用（推荐）

**配置文件 (config.yaml):**

```yaml
discovery:
  type: env  # 或 consul
  consul:
    address: "127.0.0.1:8500"
    datacenter: "dc1"
    token: ""
    timeout: 10s
```

**代码集成：**

```go
import "github.com/julesChu12/fly/mora/pkg/discovery"

// 从配置创建服务发现实例
cfg := loadDiscoveryConfig()
disc, err := discovery.New(cfg)
if err != nil {
    log.Fatal(err)
}

// 或者使用 MustNew（失败时 panic）
disc := discovery.MustNew(cfg)
```

---

## 在 gRPC 客户端中使用

### Clotho 示例

```go
package main

import (
    "github.com/julesChu12/fly/clotho/internal/infrastructure/client"
    "github.com/julesChu12/fly/mora/pkg/discovery"
)

func main() {
    // 创建服务发现实例
    disc := discovery.NewEnvDiscovery()

    // 使用服务发现创建 Custos 客户端
    custosClient, err := client.NewCustosClientWithDiscovery(disc, 5*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    defer custosClient.Close()

    // 正常使用
    user, err := custosClient.GetUser(ctx, 1)
}
```

---

## Docker Compose 配置

### 环境变量模式

```yaml
services:
  clotho:
    environment:
      # 服务发现配置（标准化格式）
      - CUSTOS_ADDRESS=custos:9001
      - HERMES_ADDRESS=hermes:9080
      - KRATOS_ADDRESS=kratos:9092
      - PLUTUS_ADDRESS=plutus:9085
    networks:
      - fly-network
```

### Consul 模式

```yaml
services:
  consul:
    image: consul:latest
    ports:
      - "8500:8500"
    networks:
      - fly-network

  clotho:
    environment:
      - DISCOVERY_TYPE=consul
      - CONSUL_ADDRESS=consul:8500
    depends_on:
      - consul
    networks:
      - fly-network
```

---

## API 参考

### Discovery 接口

```go
type Discovery interface {
    // GetService 获取一个服务实例（带负载均衡）
    GetService(ctx context.Context, serviceName string) (*ServiceInstance, error)

    // GetServices 获取所有服务实例
    GetServices(ctx context.Context, serviceName string) ([]*ServiceInstance, error)

    // Register 注册服务实例
    Register(ctx context.Context, instance *ServiceInstance) error

    // Deregister 注销服务实例
    Deregister(ctx context.Context, instanceID string) error

    // Close 关闭客户端
    Close() error
}
```

### ServiceInstance 结构

```go
type ServiceInstance struct {
    ID       string            // 实例唯一标识
    Name     string            // 服务名称
    Host     string            // 主机地址
    Port     int               // 端口
    Metadata map[string]string // 元数据
    Healthy  bool              // 健康状态
}

// Address 返回完整地址
func (s *ServiceInstance) Address() string
```

---

## 负载均衡策略

### 轮询负载均衡

```go
balancer := discovery.NewRoundRobinBalancer()
instance, err := balancer.Select(instances)
```

### 随机负载均衡

```go
balancer := discovery.NewRandomBalancer()
instance, err := balancer.Select(instances)
```

### Consul 自定义负载均衡

```go
consul, _ := discovery.NewConsulDiscovery(cfg)
consul.SetLoadBalancer(discovery.NewRandomBalancer())
```

---

## 演进路径

### Phase 1: 环境变量模式（当前）

```
✅ 优点：
- 简单易用，无需额外组件
- 适合开发环境和小规模部署
- Docker Compose 原生支持

❌ 限制：
- 不支持多实例
- 无法动态更新
- 无健康检查
```

### Phase 2: Consul 模式（3-6个月后）

```
✅ 优点：
- 支持多实例和负载均衡
- 自动健康检查
- 动态服务注册/注销
- 跨环境一致性

🔧 迁移成本：
- 部署 Consul 集群
- 修改配置文件（type: env -> consul）
- 添加服务注册逻辑
```

### Phase 3: Kubernetes 原生模式（1年后）

```
✅ 优点：
- K8s Service 原生服务发现
- 自动 DNS 解析
- 无需额外组件

🔧 实现方式：
- 实现 KubernetesDiscovery
- 使用 K8s client-go
- 读取 Service/Endpoints API
```

---

## 最佳实践

### 1. 服务命名规范

```
✅ 推荐：
- custos, hermes, kratos, plutus (小写，无横杠)
- order-service (允许横杠，自动转为 ORDER_SERVICE)

❌ 避免：
- Custos (大写)
- custos_service (下划线)
```

### 2. 环境变量格式

```bash
# 优先级：
1. {SERVICE}_ADDRESS=host:port           # 最高优先级
2. {SERVICE}_GRPC_ADDRESS=host:port      # 兼容现有配置
3. {SERVICE}_HOST + {SERVICE}_PORT       # 分离格式
```

### 3. 错误处理

```go
instance, err := disc.GetService(ctx, "custos")
if err != nil {
    switch err.(type) {
    case *discovery.ErrServiceNotFound:
        log.Error("Service not configured")
    case *discovery.ErrNoHealthyInstance:
        log.Error("All instances unhealthy")
    default:
        log.Error("Discovery failed", "error", err)
    }
}
```

### 4. 上下文超时

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

instance, err := disc.GetService(ctx, "custos")
```

---

## 常见问题

### Q1: 如何从环境变量迁移到 Consul？

**A:** 只需修改配置文件，代码无需变更：

```yaml
# 之前
discovery:
  type: env

# 之后
discovery:
  type: consul
  consul:
    address: "consul:8500"
```

### Q2: 环境变量模式支持负载均衡吗？

**A:** 不支持。环境变量模式返回单实例。如需负载均衡，请使用 Consul 模式。

### Q3: 如何调试服务发现问题？

**A:** 启用日志并检查环境变量：

```bash
# 检查环境变量
env | grep -i custos

# 查看日志
LOG_LEVEL=debug go run main.go
```

### Q4: 服务名包含横杠如何处理？

**A:** 自动转换为下划线：

```bash
# 服务名: order-service
export ORDER_SERVICE_ADDRESS=orders:9002
```

---

## 示例代码

完整示例请参考：
- `mora/pkg/discovery/discovery_test.go` - 单元测试
- `clotho/internal/infrastructure/client/custos_grpc.go` - 集成示例

---

## 总结

Mora 服务发现模块提供了从简单到复杂的完整演进路径：

1. **现在**：使用 EnvDiscovery 快速开发
2. **3个月**：切换到 ConsulDiscovery 支持生产
3. **未来**：扩展到 etcd、K8s 等平台

统一的接口保证了迁移的平滑性，让您专注于业务逻辑而不是基础设施变更。
