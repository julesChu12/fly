# Items Service - 统一商品管理服务

Items 服务是 Fly 生态系统中的统一商品管理微服务，负责管理服务和产品数据。

## 功能特性

- **统一商品模型**: 支持服务类型和产品类型的统一管理
- **多维度查询**: 按类型、分类、状态等条件进行复杂查询
- **分类管理**: 灵活的商品分类体系
- **库存管理**: 产品库存追踪和预警
- **员工分配**: 服务的员工分配管理
- **统计分析**: 商品销量、收入等统计指标

## 架构设计

### 技术栈
- **语言**: Go 1.21+
- **框架**: Gin (HTTP), gRPC (内部通信)
- **数据库**: MySQL 8.0+
- **缓存**: Redis
- **日志**: Zap
- **监控**: OpenTelemetry
- **API文档**: Swagger

### 目录结构
```
items/
├── cmd/items/           # 应用入口点
├── internal/            # 内部代码
│   ├── application/     # 应用服务层
│   ├── domain/         # 领域模型
│   │   ├── item/       # 商品领域
│   │   └── category/   # 分类领域
│   └── infrastructure/ # 基础设施层
│       ├── http/       # HTTP ��口
│       └── persistence/ # 数据持久化
├── api/proto/          # gRPC Proto 定义
├── configs/            # 配置文件
├── docs/              # API 文档
└── test/              # 测试代码
```

## 核心概念

### 商品类型 (ItemType)
- `SERVICE`: 服务类商品（需要预约和员工分配）
- `PRODUCT`: 产品类商品（需要库存和成本核算）

### 商品状态 (ItemStatus)
- `ACTIVE`: 活跃状态
- `INACTIVE`: 停用状态
- `DRAFT`: 草稿状态
- `ARCHIVED`: 归档状态

## 快速开始

### 环境要求
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 安装运行

1. **克隆项目**
   ```bash
   git clone https://github.com/julesChu12/fly.git
   cd fly/items
   ```

2. **安装依赖**
   ```bash
   make dev
   ```

3. **配置环境**
   ```bash
   # 设置环境配置文件
   make env-setup
   # 编辑 configs/.env.dev 或 configs/.env.prod
   vim configs/.env.dev
   ```

4. **运行服务**
   ```bash
   # Docker 方式（推荐）
   make docker-dev

   # 或本地运行
   make run
   ```

服务将在以下地址启动：
- HTTP API: http://localhost:8086
- gRPC: localhost:50056
- Swagger文档: http://localhost:8086/swagger/index.html

## 🔌 端口隔离策略

为了避免多服务环境中的端口冲突，我们采用环境特定的端口映射策略。

### 📊 端口分配方案

#### 开发环境 (.env.dev)
| 服务 | 容器端口 | 主机端口 | 用途 |
|------|----------|----------|------|
| **Items HTTP** | 8086 | **18086** | API 服务 |
| **Items gRPC** | 50056 | **15056** | gRPC 服务 |
| **MySQL** | 3306 | **13306** | 数据库 |
| **Redis** | 6379 | **16379** | 缓存 |

#### 生产环境 (.env.prod)
| 服务 | 容器端口 | 主机端口 | 用途 |
|------|----------|----------|------|
| **Items HTTP** | 8086 | **8086** | API 服务 |
| **Items gRPC** | 50056 | **50056** | gRPC 服务 |
| **MySQL** | 3306 | **3306** | 数据库 |
| **Redis** | 6379 | **6379** | 缓存 |

### 🛡️ 隔离机制

1. **环境前缀规则**
   - **开发环境**: 使用 1xxxx 端口段
   - **测试环境**: 使用 2xxxx 端口段
   - **生产环境**: 使用标准端口

2. **容器内部端口保持不变**
   - 所有服务在容器内部仍使用标准端口
   - 只有映射到主机的端口发生变化

3. **网络隔离**
   - 每个环境使用独立的 Docker 网络
   - 网络命名: `fly-{env}-network`

### 🔧 端口访问示例

#### 开发环境访问
```bash
# HTTP API
curl http://localhost:18086/health

# gRPC 服务
grpcurl -plaintext localhost:15056

# 数据库连接
mysql -h localhost -P 13306 -u fly_user -p

# Redis 连接
redis-cli -h localhost -p 16379
```

#### 生产环境访问
```bash
# HTTP API
curl http://localhost:8086/health

# gRPC 服务
grpcurl -plaintext localhost:50056

# 数据库连接
mysql -h localhost -P 3306 -u fly_user -p

# Redis 连接
redis-cli -h localhost -p 6379
```

## API 文档

### 商品管理

#### 创建商品
```http
POST /api/v1/items
Content-Type: application/json

{
  "name": "专业理发服务",
  "description": "包含洗剪吹的专业理发服务",
  "type": "SERVICE",
  "price": 88.00,
  "category_id": 1,
  "duration": 60,
  "staff_required": true
}
```

#### 查询商品列表
```http
GET /api/v1/items?type=SERVICE&status=ACTIVE&page=1&page_size=20
```

#### 更新商品
```http
PUT /api/v1/items/{id}
Content-Type: application/json

{
  "price": 98.00,
  "status": "ACTIVE"
}
```

#### 删除商品
```http
DELETE /api/v1/items/{id}
```

### 分类管理

#### 创建分类
```http
POST /api/v1/categories
Content-Type: application/json

{
  "name": "美容美发",
  "description": "美容和美发相关服务",
  "parent_id": null
}
```

#### 查询分类树
```http
GET /api/v1/categories/tree
```

## 开发指南

### 运行测试
```bash
make test
```

### 代码规范
```bash
make lint
```

### 生成 API 文档
```bash
make docs
```

### 数据库迁移
```bash
make migrate
```

## 配置说明

### 环境变量配置

项目使用环境变量进行配置管理，支持开发、生产等多环境：

#### 配置文件结构
```
configs/
├── .env.example    # 配置模板
├── .env.dev        # 开发环境配置
└── .env.prod       # 生产环境配置
```

#### 主要配置项

```bash
# 应用配置
APP_ENV=development
SERVER_HOST=0.0.0.0
SERVER_PORT=18086          # HTTP 服务端口
GRPC_PORT=15056            # gRPC 服务端口

# 数据库配置
DB_HOST=localhost
DB_PORT=13306              # MySQL 端口
DB_USERNAME=fly_user
DB_PASSWORD=rootpassword
DB_DATABASE=items_dev

# Redis 配置
REDIS_ADDRESS=localhost:16379
REDIS_PASSWORD=redispassword
REDIS_DB=1

# 日志配置
LOG_LEVEL=debug
LOG_FORMAT=json
```

### 配置管理命令

```bash
# 设置环境文件
make env-setup

# 加载开发环境
make env-dev

# 加载生产环境
make env-prod
```

## 部署

### Docker 环境部署

#### 开发环境
```bash
# 启动开发环境（包含数据库和缓存）
make docker-dev

# 查看服务状态
docker ps

# 查看日志
make logs

# 停止服务
make docker-dev-down
```

#### 生产环境
```bash
# 启动生产环境
make docker-prod

# 启动完整生产环境（包含监控和服务发现）
make docker-prod-full

# 停止生产环境
make docker-prod-down
```

### 安全部署
```bash
# 安全部署（带备份）
make deploy-prod-safe

# 回滚到上一版本
make rollback
```

### 环境管理
1. **环境隔离**: 开发、生产环境使用独立端口
2. **配置管理**: 通过 `.env` 文件管理环境变量
3. **网络隔离**: 每个环境使用独立的 Docker 网络
4. **健康检查**: 自动监控服务状态

## 监控和日志

- 使用 OpenTelemetry 进行分布式追踪
- Zap 结构化日志
- Prometheus 指标收集

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码
4. 创建 Pull Request

## 许可证

Apache License 2.0