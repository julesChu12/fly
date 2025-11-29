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

3. **配置数据库**
   ```bash
   # 编辑 configs/items.yaml
   vim configs/items.yaml
   ```

4. **运行服务**
   ```bash
   make run
   ```

服务将在以下地址启动：
- HTTP API: http://localhost:8086
- gRPC: localhost:50056
- Swagger文档: http://localhost:8086/swagger/index.html

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

主要配置项位于 `configs/items.yaml`：

```yaml
server:
  port: "8086"          # HTTP 服务端口
  read_timeout: 30      # 读超时时间（秒）
  write_timeout: 30     # 写超时时间（秒）

database:
  host: "localhost"     # 数据库主机
  port: 3306           # 数据库端口
  database: "items_db"  # 数据库名
  username: "root"      # 用户名
  password: ""          # 密码

redis:
  address: "localhost:6379"  # Redis 地址
  db: 0                    # Redis 数据库编号
```

## 部署

### Docker 部署
```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run
```

### 生产部署
1. 设置环境变量
2. 配置数据库连接
3. 设置 Redis 连接
4. 启动服务

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