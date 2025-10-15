# Kratos JWT 认证集成指南

## 概述

Kratos 已成功集成 Custos 作为认证服务，通过 gRPC 调用 Custos 来验证 JWT token。

## 架构设计

### 方案 A: gRPC 调用 Custos（已实现）

```
Client → Kratos (HTTP/gRPC) → Custos gRPC Service (ValidateToken)
```

**优势**：
- ✅ Custos 集中管理 token 生命周期
- ✅ 支持 session 验证、强制下线等高级功能
- ✅ 安全性高（Custos 可以撤销 token）

## 实现组件

### 1. Custos gRPC Client
**文件**: `internal/infrastructure/custos/client.go`

提供与 Custos 服务的 gRPC 通信：
```go
client, err := custos.NewClient("localhost:50051")
defer client.Close()

result, err := client.ValidateToken(ctx, token)
```

### 2. HTTP JWT 中间件
**文件**: `internal/interface/http/middleware/auth.go`

Gin 中间件，用于验证 HTTP 请求中的 JWT token：
```go
authMW := middleware.NewAuthMiddleware(custosClient, skipPaths)
router.Use(authMW.RequireAuth())
```

### 3. gRPC JWT 拦截器
**文件**: `internal/interface/grpc/middleware/auth.go`

gRPC 拦截器，用于验证 gRPC 请求中的 JWT token：
```go
authInterceptor := middleware.NewGRPCAuthInterceptor(custosClient)
grpc.ChainUnaryInterceptor(authInterceptor.UnaryInterceptor())
```

### 4. Context 辅助方法
**文件**: `pkg/utils/context.go`

从 context 中提取用户信息：
```go
userID, err := utils.GetUserIDFromContext(ctx)
tenantID, err := utils.GetTenantIDFromContext(ctx)
username := utils.GetUsernameFromContext(ctx)
```

## 配置

### configs/kratos.yaml

```yaml
auth:
  jwt_secret: "your-jwt-secret-key"  # 可选，暂未使用
  token_expire: 3600s
  custos_endpoint: "localhost:50051"  # Custos gRPC 服务地址
  skip_paths: ["/health", "/ready", "/metrics", "/swagger/*"]
```

## 使用方式

### 1. 启动 Custos 服务

首先确保 Custos 服务正在运行：

```bash
cd /path/to/custos
go run cmd/userd/main.go
```

默认 gRPC 端口: `50051`

### 2. 启动 Kratos 服务

```bash
cd /path/to/kratos
go run cmd/kratos/main.go
```

Kratos 会自动连接到 Custos：
```
Successfully connected to Custos at localhost:50051
```

### 3. 客户端调用

#### HTTP API

在 Authorization header 中携带 JWT token：

```bash
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/api/orders
```

#### gRPC API

在 metadata 中携带 JWT token：

```go
md := metadata.Pairs("authorization", "Bearer "+token)
ctx := metadata.NewOutgoingContext(context.Background(), md)

resp, err := client.GetOrder(ctx, &pb.GetOrderRequest{OrderId: 123})
```

### 4. 业务代码中获取用户信息

#### HTTP Handler

```go
func (h *OrderHandler) CreateOrder(c *gin.Context) {
    // 从 Gin context 中获取用户信息
    userID := middleware.MustGetUserID(c)
    tenantID, _ := middleware.GetTenantID(c)
    username := middleware.GetUsername(c)

    // 业务逻辑...
}
```

#### gRPC Handler

```go
func (s *OrderServiceServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
    // 从 context 中获取用户信息
    userID, err := utils.GetUserIDFromContext(ctx)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "user not authenticated")
    }

    tenantID, _ := utils.GetTenantIDFromContext(ctx)
    username := utils.GetUsernameFromContext(ctx)

    // 业务逻辑...
}
```

## 认证流程

### HTTP 请求流程

```
1. Client 发送请求，Header 携带: Authorization: Bearer <token>
2. Kratos HTTP 中间件提取 token
3. 调用 Custos gRPC ValidateToken(token)
4. Custos 验证 token 并返回用户信息
5. Kratos 将用户信息注入到 Gin Context
6. 业务 Handler 从 Context 获取用户信息
```

### gRPC 请求流程

```
1. Client 发送 gRPC 请求，Metadata 携带: authorization: Bearer <token>
2. Kratos gRPC 拦截器提取 token
3. 调用 Custos gRPC ValidateToken(token)
4. Custos 验证 token 并返回用户信息
5. Kratos 将用户信息注入到 gRPC Context
6. 业务 Handler 从 Context 获取用户信息
```

## JWT Token 结构

Custos 签发的 JWT token 包含以下 claims：

```json
{
  "user_id": 123,
  "username": "john_doe",
  "role": "USER_TYPE_CUSTOMER",
  "session_id": "uuid-string",
  "tenant_id": 456,
  "iss": "custos",
  "sub": "123",
  "exp": 1234567890,
  "iat": 1234567890
}
```

Kratos 从 ValidateToken 响应中获取以下信息并注入到 Context：
- `user_id`: 用户 ID
- `username`: 用户名
- `email`: 用户邮箱
- `tenant_id`: 租户 ID（用于多租户隔离）
- `user_type`: 用户类型（customer/admin/operator）

## 跳过认证的路径

以下路径会跳过 JWT 认证：
- `/health` - 健康检查
- `/ready` - 就绪检查
- `/metrics` - Prometheus 指标
- `/swagger/*` - API 文档

可以在 `configs/kratos.yaml` 中配置 `auth.skip_paths` 来添加更多路径。

## 故障处理

### 1. Custos 服务不可用

如果 Custos 服务未启动或连接失败，Kratos 会输出警告并继续启动（不带认证中间件）：

```
Warning: failed to connect to Custos at localhost:50051: connection refused
Running without authentication middleware
```

### 2. Token 验证失败

客户端会收到 `401 Unauthorized` 响应：

```json
{
  "code": "UNAUTHORIZED",
  "message": "invalid token"
}
```

### 3. 连接超时

Token 验证的超时时间为 3 秒，超时后会返回认证失败。

## 最佳实践

1. **生产环境配置**
   - 使用环境变量配置 `custos_endpoint`
   - 确保 Custos 和 Kratos 之间的网络稳定性
   - 考虑使用 Kubernetes Service 或服务网格进行服务发现

2. **安全建议**
   - 始终使用 HTTPS/TLS 传输 JWT token
   - Token 应该有合理的过期时间
   - 生产环境应该使用真实的 JWT secret key

3. **性能优化**
   - Custos ValidateToken 调用会增加约 3-10ms 延迟
   - 可以考虑添加本地缓存（带 TTL）来减少 Custos 调用
   - 使用 gRPC 连接池复用连接

## 下一步

### 可选增强功能

1. **本地 Token 缓存**
   - 缓存已验证的 token（带 TTL）
   - 减少 Custos 服务压力

2. **Token 刷新**
   - 实现 refresh token 机制
   - 自动刷新即将过期的 token

3. **细粒度权限控制**
   - 集成 Casbin 实现 RBAC
   - 支持基于角色的访问控制

4. **多租户隔离增强**
   - 在所有 Repository 方法中强制 tenant_id 过滤
   - 防止跨租户数据访问

## 相关文件

- `internal/infrastructure/custos/client.go` - Custos gRPC 客户端
- `internal/interface/http/middleware/auth.go` - HTTP 认证中间件
- `internal/interface/grpc/middleware/auth.go` - gRPC 认证拦截器
- `pkg/utils/context.go` - Context 辅助方法
- `cmd/kratos/main.go` - 主程序（集成认证）
- `configs/kratos.yaml` - 配置文件
