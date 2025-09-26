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
  ├── pkg/
  │   ├── auth/        # JWT token 生成与校验（无 DB、无框架依赖）
  │   ├── logger/      # 日志封装
  │   ├── config/      # 配置加载
  │   ├── db/          # 数据库封装
  │   ├── cache/       # Redis 缓存封装
  │   ├── mq/          # 消息队列封装
  │   └── utils/       # 通用工具
  │
  ├── adapters/
  │   ├── gin/         # gin 框架中间件适配
  │   │   └── authmw.go
  │   └── gozero/      # (可选) go-zero 框架中间件适配
  │
  ├── starter/
  │   └── gin-starter/ # 最小 gin 示例工程
  │       ├── main.go
  │       └── user_service_mock.go
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

- **db/**  
  数据库封装，基于 sqlx 或 gorm。  

- **cache/**  
  Redis 工具，支持常见模式（缓存 aside、分布式锁）。  

- **mq/**  
  消息队列封装，支持 Kafka/RabbitMQ。  

- **utils/**  
  工具函数（string、time、crypto 等）。  

---

### adapters/
- **gin/**  
  提供 gin 中间件包装，如：  
  - `AuthMiddleware(secret)`：调用 `pkg/auth` 校验 token，将 userID 注入 gin.Context。  

- **gozero/** (可选)  
  提供 go-zero 的中间件包装。  

---

### starter/
- **gin-starter/**  
  演示 API 层如何编排 User Service 与 Auth 模块：  
  - `/login`：模拟调用 User Service 验证用户名密码，成功后用 `pkg/auth` 签发 token。  
  - `/ping`：受保护接口，使用 `AuthMiddleware` 验证 token，返回 userID。  

运行方式：  
```bash
cd starter/gin-starter
go run main.go
```

---

## 设计原则
- **核心能力包（pkg/）框架无关**  
- **adapters/** 作为防腐层，负责将能力包接入 gin/go-zero 等框架  
- **starter/** 演示完整场景，API 层是 orchestrator（编排器），连接 Auth 与 User Service  
- **User Service 属于领域服务**，负责用户表/权限表，不与 Auth 模块耦合  
- **Mora 不负责用户认证逻辑（登录/刷新/状态管理），这些属于 UserService (Custos)**  

---

## 下一步
- 扩展 Redis 缓存封装  
- 增加 go-zero starter  
- 提供 CI/CD 脚手架和部署示例  
