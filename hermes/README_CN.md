# Hermes - 客户服务

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
```

## 📖 文档

### API文档

- Swagger UI: `http://localhost:8083/swagger/index.html`
- OpenAPI 3.0规范文件: `/docs/swagger.json`

## 🏗️ 系统架构

### 模块职责划分

- **Mora** → 能力库（认证令牌签名/验证、日志、配置、数据库、缓存、消息队列、工具函数）
- **Clotho** → API 编排层（入口点、信任/零信任、请求路由）
- **Hermes (客户域)** → 负责客户信息管理、联系方式管理、客户生命周期
- **Kratos** → 订单服务（订单创建、状态管理、订单查询）
- **Plutus** → 支付钱包服务（支付处理、钱包管理、交易记录）

---

## Hermes 核心职责

### 1. 客户生命周期管理

- 客户信息创建（姓名、电话、邮箱、标签）
- 客户信息更新和删除
- 客户状态管理（活跃、禁用、删除）
- 多租户客户隔离（基于 tenant_id）

### 2. 联系方式管理

- 多种联系方式类型支持（电话、邮箱、地址、微信等）
- 主要联系方式标记（is_primary）
- 联系方式与客户的关联管理
- 联系方式的分组和分类

### 3. 数据校验和业务规则

- 邮箱唯一性验证（同一租户内）
- 输入参数验证和业务规则检查
- 数据完整性约束
- 错误处理和异常管理

### 4. 分页查询和性能优化

- 支持大数据量的分页查询
- 数据库索引优化（tenant_id、phone、email）
- 查询性能监控和优化
- 缓存策略（未来扩展）

### 5. 分布式事务支持

- DTM 分布式事务管理器集成
- SAGA 模式支持（长流程业务）
- TCC 模式支持（短流程高一致性）
- XA 模式支持（数据库事务）

### 6. 双协议支持

- HTTP REST API（面向 Web 应用）
- gRPC 接口（面向微服务间通信）
- 统一的业务逻辑层
- 协议无关的数据模型

### 7. 可观测性和监控

- 分布式链路追踪（OpenTelemetry）
- 业务指标监控（Prometheus）
- 结构化日志输出
- 健康检查端点

---

## 数据库架构（DDL）

### customers 客户表

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

### contacts 联系方式表

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

---

## 项目结构

```text
hermes/
├── cmd/
│   └── hermes/                 # 主应用程序入口点
├── configs/
│   ├── hermes.yaml             # 服务配置文件
│   └── migrations/             # 数据库迁移文件
│       └── 001_initial_schema.sql
├── docs/
│   ├── docs.go                 # Swagger 文档生成
│   ├── swagger.json            # OpenAPI 3.0 规范
│   └── swagger.yaml            # Swagger YAML 格式
├── internal/
│   ├── application/            # 应用层（用例）
│   │   └── service/           # 应用服务实现
│   │       ├── customer_service.go
│   │       └── customer_service_test.go
│   ├── domain/                # 域层（业务逻辑）
│   │   ├── entity/            # 域实体
│   │   │   └── customer.go    # 客户和联系方式实体
│   │   └── repository/        # 仓储接口
│   │       └── interfaces.go  # 数据访问接口定义
│   ├── infrastructure/        # 基础设施层
│   │   └── database/          # 数据持久化
│   │       ├── customer_repository.go
│   │       └── contact_repository.go
│   └── interface/             # 接口层（HTTP/gRPC 处理器）
│       ├── grpc/              # gRPC 接口实现
│       │   ├── customer_handler.go
│       │   └── contact_handler.go
│       └── http/              # HTTP 接口实现
│           ├── customer_handler.go
│           └── customer_handler_test.go
├── pkg/                       # Hermes 特定包
│   ├── constants/             # 域特定常量
│   │   └── constants.go       # 服务常量定义
│   ├── dtm/                   # 分布式事务管理
│   │   └── manager.go         # DTM 管理器实现
│   ├── errors/                # 域特定错误
│   │   └── errors.go          # 错误类型定义
│   └── types/                 # 域特定类型
│       └── types.go           # 请求/响应类型定义
├── api/
│   └── proto/                 # Protocol Buffers 定义
│       ├── hermes.proto       # gRPC 服务定义
│       ├── hermes.pb.go       # 生成的 Go 代码
│       └── hermes_grpc.pb.go  # 生成的 gRPC 代码
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 层职责说明

### 域层 (`internal/domain/`)

**实体** (`entity/`):
- `Customer`: 核心客户实体，包含业务规则和验证逻辑
- `Contact`: 联系方式实体，支持多种联系方式类型

**仓储** (`repository/`):
- `CustomerRepository`: 客户数据访问契约
- `ContactRepository`: 联系方式数据访问契约

### 应用层 (`internal/application/`)

**服务** (`service/`):
- `CustomerService`: 客户管理业务逻辑
- 包含客户 CRUD 操作和业务规则验证
- 支持联系方式关联查询

### 基础设施层 (`internal/infrastructure/`)

**持久化** (`database/`):
- `customer_repository.go`: MySQL 客户数据访问实现
- `contact_repository.go`: MySQL 联系方式数据访问实现
- 使用 GORM 进行数据库操作

### 接口层 (`internal/interface/`)

**HTTP** (`http/`):
- `customer_handler.go`: HTTP 请求处理器
- 支持 RESTful API 设计
- 集成 Swagger 文档生成

**gRPC** (`grpc/`):
- `customer_handler.go`: gRPC 服务处理器
- 支持微服务间通信
- Protocol Buffers 序列化

---

## 核心特性

### 客户管理
- 客户信息 CRUD 操作
- 多租户数据隔离
- 邮箱唯一性验证
- 客户标签管理

### 联系方式管理
- 多种联系方式类型（电话、邮箱、地址、微信）
- 主要联系方式标记
- 联系方式与客户关联
- 联系方式分组管理

### 数据校验
- 输入参数验证
- 业务规则检查
- 数据完整性约束
- 错误处理和异常管理

### 分页查询
- 支持大数据量分页
- 数据库索引优化
- 查询性能监控
- 可配置页面大小

### 分布式事务
- DTM 集成支持
- SAGA 模式（长流程）
- TCC 模式（高一致性）
- XA 模式（数据库事务）

### 双协议支持
- HTTP REST API
- gRPC 接口
- 统一业务逻辑
- 协议无关设计

---

## 未来增强

### 缓存支持
- Redis 集成用于会话存储
- 查询结果缓存
- 性能优化

### 事件驱动架构
- 客户生命周期事件
- 审计日志
- 消息队列集成

### 高级查询
- 全文搜索支持
- 复杂查询条件
- 数据导出功能

### 监控和可观测性
- 业务指标监控
- 性能分析
- 告警机制

---

## 依赖关系

### 外部依赖
- **Gin**: HTTP Web 框架
- **GORM**: ORM 数据库操作
- **DTM**: 分布式事务管理
- **Protocol Buffers**: gRPC 序列化

### 内部依赖（Mora 生态）
- **Mora**: 共享能力库
- **Clotho**: API 编排（未来）
- **Kratos**: 订单服务
- **Plutus**: 支付钱包服务

---

## 配置管理

环境变量通过 `configs/hermes.yaml` 管理：
- 数据库连接设置
- 服务端口配置
- 缓存配置
- 可观测性配置
- DTM 服务器配置

---

## 测试策略

- 域逻辑单元测试
- 用例集成测试
- HTTP 处理器测试
- gRPC 服务测试
- 数据库迁移测试

---

## 开发工作流

1. **域优先**: 从域实体和业务规则开始
2. **仓储模式**: 定义数据访问接口
3. **用例实现**: 实现应用逻辑
4. **基础设施**: 添加持久化和外部服务实现
5. **接口层**: 创建 HTTP/gRPC 处理器和中间件
6. **测试**: 添加全面的测试覆盖

---

## AI 助手指南

- 使用清晰、描述性的函数和变量名
- 为复杂的业务逻辑添加全面的注释
- 遵循 Go 约定和最佳实践
- 保持一致的错误处理模式
- 记录架构决策和权衡

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
