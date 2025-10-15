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

## 🗄️ 数据库设计（MySQL）

详见 `configs/migrations/001_initial_schema.sql`

### customers (客户表)

```sql
CREATE TABLE customers (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT NOT NULL COMMENT '租户ID（多租户隔离）',
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(255),
    tags TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant (tenant_id),
    INDEX idx_phone (phone),
    INDEX idx_email (email),
    UNIQUE KEY uk_tenant_email (tenant_id, email)
);
```

- **多租户隔离**：通过 `tenant_id` 实现数据隔离
- **唯一约束**：每个租户内 email 唯一
- **索引优化**：对租户、手机号、邮箱建立索引

### contacts (联系方式表)

```sql
CREATE TABLE contacts (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id BIGINT NOT NULL COMMENT '租户ID（多租户隔离）',
    customer_id INT UNSIGNED NOT NULL,
    type VARCHAR(50) NOT NULL,
    value VARCHAR(255) NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tenant (tenant_id),
    INDEX idx_customer_id (customer_id),
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
);
```

- **联系方式类型**：支持电话、邮箱、地址、微信等多种类型
- **主联系方式**：`is_primary` 标记主要联系方式
- **级联删除**：删除客户时自动删除关联的联系方式

---

## 📋 API接口

### HTTP REST API

| 方法 | 路径 | 描述 | 状态 |
|------|------|------|------|
| POST | `/api/customers` | 创建客户 | ✅ 已实现 |
| GET | `/api/customers/{id}` | 获取客户信息 | ✅ 已实现 |
| GET | `/api/customers/{id}/contacts` | 获取客户及联系方式 | ✅ 已实现 |
| PUT | `/api/customers/{id}` | 更新客户信息 | ✅ 已实现 |
| DELETE | `/api/customers/{id}` | 删除客户 | ✅ 已实现 |
| GET | `/api/customers` | 客户列表查询（分页） | ✅ 已实现 |

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

- Prometheus指标暴露在 `:8083/metrics`
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

# 查看覆盖率报告
open coverage.html
```

---

## 🔌 API 使用示例

### 创建客户

```bash
curl -X POST http://localhost:8080/api/customers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "张三",
    "phone": "13800138000",
    "email": "zhangsan@example.com",
    "tags": "VIP,企业客户"
  }'

# 响应示例
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "name": "张三",
    "phone": "13800138000",
    "email": "zhangsan@example.com",
    "tags": "VIP,企业客户",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

### 获取客户信息

```bash
curl -X GET http://localhost:8080/api/customers/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 获取客户及联系方式

```bash
curl -X GET http://localhost:8080/api/customers/1/contacts \
  -H "Authorization: Bearer $TOKEN"

# 响应示例
{
  "code": 0,
  "message": "success",
  "data": {
    "customer": {
      "id": 1,
      "name": "张三",
      "phone": "13800138000",
      "email": "zhangsan@example.com"
    },
    "contacts": [
      {
        "id": 1,
        "customer_id": 1,
        "type": "phone",
        "value": "13800138000",
        "is_primary": true
      },
      {
        "id": 2,
        "customer_id": 1,
        "type": "email",
        "value": "zhangsan@example.com",
        "is_primary": true
      }
    ]
  }
}
```

### 更新客户信息

```bash
curl -X PUT http://localhost:8080/api/customers/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "张三（更新）",
    "phone": "13900139000",
    "tags": "VVIP,企业客户,长期合作"
  }'
```

### 客户列表查询

```bash
# 分页查询
curl -X GET "http://localhost:8080/api/customers?page=1&page_size=20" \
  -H "Authorization: Bearer $TOKEN"

# 按手机号查询
curl -X GET "http://localhost:8080/api/customers?phone=138" \
  -H "Authorization: Bearer $TOKEN"
```

### 删除客户

```bash
curl -X DELETE http://localhost:8080/api/customers/1 \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📊 实现状态

### ✅ 已完成 (95%)

#### 核心功能
- ✅ 客户 CRUD 完整实现
- ✅ 联系方式管理
- ✅ 多租户数据隔离
- ✅ 分页查询支持
- ✅ 数据校验和业务规则检查

#### API 接口
- ✅ HTTP REST API 完整实现
- ✅ gRPC API 接口定义和实现
- ✅ Swagger API 文档生成
- ✅ JWT 认证集成

#### 数据层
- ✅ MySQL 数据库集成
- ✅ Redis 缓存支持
- ✅ 数据库迁移脚本
- ✅ 完整的数据库 Schema

#### 可观测性
- ✅ OpenTelemetry 链路追踪集成
- ✅ Prometheus 指标收集
- ✅ 结构化日志输出
- ✅ 健康检查端点

#### 分布式特性
- ✅ DTM 分布式事务集成
- ✅ Redis 缓存策略
- ✅ gRPC 服务间通信

#### 部署
- ✅ Dockerfile
- ✅ Makefile 构建工具
- ✅ 配置文件管理

### 🔧 待完善 (5%)

- ⏳ 客户标签管理功能增强
- ⏳ 客户分组和批量操作
- ⏳ 更多联系方式类型支持
- ⏳ 客户画像和分析功能
- ⏳ 性能压测报告

---

## 📖 文档

### API文档

- Swagger UI: `http://localhost:8083/swagger/index.html`
- OpenAPI 3.0规范文件: `/docs/swagger.json`

### 架构文档

详细的架构设计文档正在完善中，当前可以参考代码结构了解系统设计：

```text
hermes/
├── cmd/hermes/           # 应用入口
├── internal/
│   ├── application/      # 应用服务层
│   ├── domain/          # 领域层
│   ├── infrastructure/  # 基础设施层
│   └── interface/       # 接口层
├── pkg/                 # 公共类型定义
├── configs/             # 配置文件
└── docs/               # 自动生成的API文档
```

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
