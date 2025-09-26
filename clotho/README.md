Clotho

**Clotho** 源自希腊神话中的命运三女神之一，她负责纺织命运之线。  
在本项目中，Clotho 承担着 **API 层编排者** 的角色：  
它接收外部请求，调用 Custos（用户域）、订单域、支付域等领域服务，  
并将结果统一对外暴露，成为系统的「对外之线」。  

---

## 🎯 定位
- **不是网关**：Clotho 不负责限流、熔断、流量控制，这些交给 API Gateway 或 Service Mesh。  
- **不是领域服务**：Clotho 不维护业务数据，所有领域逻辑都在 Custos/Orders/Billing 等服务中。  
- **是编排层**：负责请求转发、聚合、统一对外接口。  

---

## 🏗️ 架构边界
- **输入**：HTTP/REST API，对外提供业务接口。  
- **输出**：通过 gRPC 调用内部领域服务（Custos/Orders/...）。  
- **职责**：  
  - 请求解析与路由  
  - 使用 Mora 的 Auth 中间件校验 Access Token  
  - 调用 Custos 完成用户认证/鉴权  
  - 调用其他领域服务完成业务编排  
  - 聚合结果，返回标准化响应  

---

## 📂 项目结构
```
clotho/
├── cmd/
│   └── clotho/            # 启动入口（cobra: serve, version...）
├── configs/
│   └── clotho.yaml
├── internal/
│   ├── application/
│   │   └── usecase/       # API 调用编排逻辑
│   │       ├── user_proxy.go
│   │       ├── order_proxy.go
│   │       └── payment_proxy.go
│   ├── infrastructure/
│   │   ├── client/        # gRPC 客户端
│   │   │   ├── custos_grpc.go
│   │   │   └── orders_grpc.go
│   │   └── http/          # 对外 HTTP API
│   │       ├── handler/
│   │       └── router.go
│   └── middleware/
│       └── auth.go
├── docs/
│   └── README.md
└── go.mod
```

---

## 🚦 请求流转
1. 外部客户端调用 Clotho 的 HTTP API  
2. Clotho 使用 **Mora Auth Middleware** 验证 Access Token  
3. 根据路由，Clotho 调用 Custos/Orders 等服务（gRPC）  
4. 聚合结果 → 返回 HTTP 响应  

---

## 🔑 关键特性
- 对外统一 API，隐藏内部服务细节  
- 内部 gRPC，高性能调用  
- 与 Custos 解耦，Custos 专注领域逻辑，Clotho 专注编排  
- 可扩展：未来可接入 Service Mesh / API Gateway 补充流控与安全  

---

## 🚀 下一步
- 实现基础的 HTTP 路由 + gRPC 调用示例  
- 集成 Mora 的 Auth 中间件  
- 提供健康检查接口 `/health`  
- CI/CD：配置独立流水线，支持单独发布  