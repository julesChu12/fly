# Hermes - Customer Service (客户服务)

Hermes是Fly平台的客户管理服务，提供客户信息和联系方式的管理功能。该服务同时支持HTTP REST API和gRPC协议，集成了分布式链路追踪、缓存、分布式事务等企业级特性。

## 🚀 特性

### 核心功能
- **客户管理**: 客户信息的增删改查操作
- **联系方式管理**: 支持多种联系方式类型（电话、邮箱、地址、微信等）
- **数据校验**: 完善的输入验证和业务规则检查
- **分页查询**: 支持大数据量的分页查询

### 技术特性
- **双协议支持**: 同时提供HTTP REST API和gRPC接口
- **分布式链路追踪**: 集成Mora框架的可观测性能力
- **缓存支持**: Redis缓存提升性能
- **分布式事务**: DTM分布式事务管理器
- **API文档**: Swagger/OpenAPI 3.0文档
- **健康检查**: 完善的服务健康检查机制

## 📋 API接口

### HTTP REST API

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/customers` | 创建客户 |
| GET | `/api/customers/{id}` | 获取客户信息 |
| GET | `/api/customers/{id}/contacts` | 获取客户及联系方式 |
| PUT | `/api/customers/{id}` | 更新客户信息 |
| DELETE | `/api/customers/{id}` | 删除客户 |
| GET | `/api/customers` | 客户列表查询 |

### gRPC 接口

```protobuf
service CustomerService {
  rpc CreateCustomer(CreateCustomerRequest) returns (CreateCustomerResponse);
  rpc GetCustomer(GetCustomerRequest) returns (GetCustomerResponse);
  rpc GetCustomerWithContacts(GetCustomerRequest) returns (GetCustomerWithContactsResponse);
  rpc UpdateCustomer(UpdateCustomerRequest) returns (UpdateCustomerResponse);
  rpc DeleteCustomer(DeleteCustomerRequest) returns (DeleteCustomerResponse);
  rpc ListCustomers(ListCustomersRequest) returns (ListCustomersResponse);
}
```

## 🛠 开发与部署

### 前置要求
- Go 1.25+
- MySQL 8.0+
- Redis 6.0+
- Protocol Buffers Compiler (protoc)

### 本地开发

```bash
# 安装依赖
make deps

# 生成protobuf代码
make proto

# 生成Swagger文档
make swagger

# 运行服务
make dev
```

### Docker部署

```bash
# 构建镜像
make docker-build

# 使用docker-compose启动
docker-compose up hermes
```

## 🔧 配置

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `DATABASE_URL` | MySQL连接字符串 | `root:password@tcp(localhost:3306)/hermes` |
| `PORT` | HTTP服务端口 | `8080` |
| `GRPC_PORT` | gRPC服务端口 | `9080` |
| `REDIS_URL` | Redis连接地址 | `localhost:6379` |
| `DTM_SERVER` | DTM服务器地址 | `localhost:36790` |

### 配置文件
```yaml
# configs/hermes.yaml
server:
  http_port: 8080
  grpc_port: 9080

database:
  driver: mysql
  dsn: "root:password@tcp(localhost:3306)/hermes"
  max_open_conns: 100
  max_idle_conns: 10

cache:
  driver: redis
  address: "localhost:6379"
  password: ""
  db: 0

observability:
  tracing:
    endpoint: "http://localhost:4317"
    service_name: "hermes"
  metrics:
    enabled: true
    port: 9090
```

## 📊 监控与可观测性

### 健康检查
- HTTP: `GET /health`
- gRPC: 内置健康检查服务

### 指标监控
- Prometheus指标暴露在 `:9090/metrics`
- 包含业务指标和系统指标

### 链路追踪
- 集成OpenTelemetry
- 支持Jaeger和Zipkin

### 日志
- 结构化日志输出
- 支持多种日志级别
- 集成链路追踪ID

## 🔄 分布式事务

### DTM集成
Hermes集成了DTM分布式事务管理器，支持以下事务模式：

1. **SAGA模式**: 适用于长流程业务
2. **TCC模式**: 适用于短流程高一致性要求
3. **XA模式**: 适用于数据库事务

### 使用示例
```go
// 创建完整的订单业务流程
businessTx := &dtm.BusinessTransaction{
    CustomerID: 1,
    OrderItems: []dtm.OrderItem{
        {ProductID: 1, Quantity: 2, Price: 100.0},
    },
    PaymentAmount: 200.0,
}

err := dtmManager.ProcessOrderTransaction(ctx, businessTx)
```

## 🧪 测试

```bash
# 运行单元测试
make test

# 运行集成测试
make test-integration

# 生成测试覆盖率报告
make test-coverage
```

## 📖 文档

### API文档
- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI 3.0规范文件: `/docs/swagger.json`

### 架构文档
- [架构设计](docs/architecture.md)
- [数据库设计](docs/database.md)
- [API设计](docs/api.md)

## 🤝 贡献

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 Apache 2.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关服务

- [Kratos - 订单服务](../kratos/README.md)
- [Plutus - 支付钱包服务](../plutus/README.md)
- [Mora - 基础框架](../mora/README.md)
- [Custos - 认证授权服务](../custos/README.md)