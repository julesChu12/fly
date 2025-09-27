# Mora

**Mora** 来自希腊神话中的 *Moirai*（命运三女神），她们掌控着众生的命运之线。  
作为一个 Golang 能力库，Mora 承载着“分配与秩序”的寓意：  
它为所有服务提供通用的基础能力模块，让项目在规则与清晰边界下快速启航。  

Mora 并不是一个具体的网关或框架，而是一个 **能力源泉**：  
- 在 `pkg/` 中沉淀通用模块（auth/logger/config/...）  
- 在 `adapters/` 中提供框架适配层  
- 在 `starter/` 中演示 API 层如何 orchestrate（编排）能力与领域服务  

---

## 项目结构
```
mora/
  ├── go.mod
  ├── pkg/                    # 核心能力包（框架无关）✅
  │   ├── auth/              # JWT token 生成与校验 ✅
  │   │   ├── claims.go      # JWT Claims 结构定义
  │   │   ├── jwt.go         # JWT 生成与验证
  │   │   └── jwks.go        # JWK 支持
  │   ├── logger/            # 日志封装 ✅
  │   │   ├── logger.go      # Zap 日志封装
  │   │   └── context.go     # 上下文追踪支持（含 OTel 集成）
  │   ├── config/            # 配置加载 ✅
  │   │   └── loader.go      # YAML/ENV 配置加载
  │   ├── observability/     # 监控和链路追踪 ✅
  │   │   ├── observability.go # OpenTelemetry 初始化
  │   │   ├── config.go      # 可观测性配置
  │   │   └── utils.go       # TraceID/SpanID 工具
  │   ├── db/                # 数据库封装 ✅
  │   │   ├── gorm.go        # GORM 封装
  │   │   └── sqlx.go        # SQLX 封装
  │   ├── cache/             # Redis 缓存封装 ✅
  │   │   ├── redis.go       # Redis 基础操作
  │   │   └── lock.go        # 分布式锁
  │   ├── mq/                # 消息队列封装 ✅
  │   │   ├── mq.go          # 消息队列接口
  │   │   ├── memory.go      # 内存队列实现
  │   │   └── redis.go       # Redis 队列实现
  │   └── utils/             # 通用工具 ✅
  │       ├── crypto.go      # 加密工具
  │       ├── string.go      # 字符串工具
  │       └── time.go        # 时间工具
  │
  ├── adapters/              # 框架适配层 ✅
  │   ├── gin/               # Gin 框架适配 ✅
  │   │   ├── auth_middleware.go # JWT 认证中间件
  │   │   └── otel_middleware.go # OpenTelemetry 中间件
  │   └── gozero/            # Go-Zero 框架适配 ✅
  │       ├── auth_middleware.go # JWT 认证中间件
  │       ├── context.go     # 上下文工具
  │       └── otel_middleware.go # OpenTelemetry gRPC 拦截器
  │
  ├── starter/               # 示例应用 ✅
  │   ├── gin-starter/       # Gin 演示应用
  │   │   ├── main.go        # 完整的 REST API 示例
  │   │   └── docs/          # Swagger 文档
  │   └── gozero-starter/    # Go-Zero 演示应用
  │       ├── main.go        # Go-Zero 服务示例
  │       ├── api/           # API 定义
  │       ├── etc/           # 配置文件
  │       └── internal/      # 内部实现
  │
  └── docs/
      └── usage-examples.md
```

---

## 模块说明

### pkg/
- **auth/**  
  提供 JWT/JWK 的生成与验证工具方法：  
  - `GenerateToken(userID, secret, ttl)`  
  - `ValidateToken(token, secret)` → 返回 `Claims`（含 userID）  
  - **不依赖 DB，不依赖 User Service**  
  - **仅提供 JWT/JWK 工具方法，不负责用户认证或状态管理**  

- **logger/**  
  封装日志库（zap/logx），统一输出格式，支持 traceId。

- **config/**  
  支持 YAML/ENV 配置加载，未来可扩展远程配置中心。

- **observability/**  
  OpenTelemetry 可观测性支持，提供链路追踪、指标收集和日志关联。

- **db/**  
  数据库封装，基于 sqlx 或 gorm。

- **cache/**  
  Redis 工具，支持常见模式（缓存 aside、分布式锁）。

- **mq/**  
  消息队列封装，支持内存和 Redis 实现。

- **utils/**  
  工具函数（string、time、crypto 等）。

---

### adapters/
- **gin/**  
  提供 gin 中间件包装，如：  
  - `AuthMiddleware(secret)`：调用 `pkg/auth` 校验 token，将 userID 注入 gin.Context。  
  - `ObservabilityMiddleware(serviceName)`：添加 OpenTelemetry 链路追踪支持。

- **gozero/**  
  提供 go-zero 的中间件包装：  
  - `AuthMiddleware(secret)`：JWT 认证中间件  
  - `ServerOption()` / `ClientOption()`：gRPC OpenTelemetry 拦截器  

---

### starter/
- **gin-starter/**  
  演示 API 层如何编排 User Service 与 Auth 模块：  
  - `/login`：模拟调用 User Service 验证用户名密码，成功后用 `pkg/auth` 签发 token。  
  - `/ping`：受保护接口，使用 `AuthMiddleware` 验证 token，返回 userID。  

运行方式：
```bash
# Gin 示例应用
cd starter/gin-starter
go run main.go
# 访问 http://localhost:8080/swagger/ 查看 API 文档

# Go-Zero 示例应用
cd starter/gozero-starter
go run main.go -f etc/mora-api.yaml
# 默认运行在 http://localhost:8888
```

### API 接口示例
- **公开接口**：
  - `GET /health` - 健康检查
  - `POST /login` - 用户登录（返回 JWT Token）
- **认证接口**：
  - `GET /profile` - 获取用户信息
  - `GET /protected` - 受保护的示例接口
  - `GET /api/v1/orders` - 获取订单列表
  - `POST /api/v1/orders` - 创建订单
  - `GET /api/v1/users` - 获取用户列表

---

## 实现状态

### ✅ 已完成
- **核心能力包（pkg/）**：
  - `auth/` - JWT Token 生成与验证，支持 JWKS
  - `logger/` - 基于 Zap 的结构化日志，支持链路追踪
  - `config/` - 统一配置加载（YAML + ENV）
  - `observability/` - OpenTelemetry 可观测性支持
  - `db/` - 数据库抽象层（GORM + SQLX）
  - `cache/` - Redis 缓存与分布式锁
  - `mq/` - 消息队列抽象（内存 + Redis 实现）
  - `utils/` - 通用工具集（加密、字符串、时间）

- **框架适配器（adapters/）**：
  - `gin/` - Gin 框架认证中间件 + OpenTelemetry 中间件
  - `gozero/` - Go-Zero 框架认证中间件 + OpenTelemetry 中间件

- **演示应用（starter/）**：
  - `gin-starter/` - 完整的 Gin REST API（含 Swagger 文档）
  - `gozero-starter/` - Go-Zero 微服务示例

- **测试覆盖**：
  - 所有核心包都有完整的单元测试
  - 测试通过率 100%（50 个 Go 文件）
  - 支持多种数据库和缓存后端

### 📋 开发计划
- 扩展更多 MQ 实现（Kafka/RabbitMQ）
- 添加更多数据库驱动支持（MongoDB、ClickHouse 等）
- 完善 CI/CD 脚手架和自动化测试
- 增加部署示例和最佳实践
- 添加更多框架适配器（Echo、Fiber 等）
- 完善监控和告警集成

---

## 设计原则
- **核心能力包（pkg/）框架无关**  
- **adapters/** 作为防腐层，负责将能力包接入 gin/go-zero 等框架  
- **starter/** 演示完整场景，API 层是 orchestrator（编排器），连接 Auth 与 User Service  
- **User Service 属于领域服务**，负责用户表/权限表，不与 Auth 模块耦合  
- **Mora 不负责用户认证逻辑（登录/刷新/状态管理），这些属于 UserService (Custos)**  

---

## 下一步
- 持续完善核心模块的功能和性能优化
- 扩展更多框架适配器（Echo、Fiber 等）
- 增加更多业务场景的示例和最佳实践
- 优化文档和开发者体验  

---

## 语言支持
- [中文版 README](README.md)
- [English README](README_EN.md)
