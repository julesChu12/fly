# Plutus 修复总结

## 已完成的 P0 和 P1 级别修复

### ✅ P0 - 关键修复（影响功能正确性）

#### 1. 修复事务原子性问题（`internal/application/service/wallet_service.go`）
**问题**: Recharge、Consume、Refund 方法中，创建交易记录和更新余额不在同一数据库事务中，违反原子性要求。

**解决方案**:
- 创建了 `pkg/database/transaction.go` 提供事务辅助方法 `WithTransaction()`
- 重构了 Recharge、Consume、Refund 方法，使用数据库事务包装所有资金操作
- 确保交易记录创建和余额更新在同一事务中完成，失败时自动回滚

**影响的文件**:
- `pkg/database/transaction.go` (新建)
- `internal/application/service/wallet_service.go` (修改)
- `cmd/plutus/main.go` (修改 - 添加 db 参数)

#### 2. 修复 LockForUpdate 参数错误（`internal/application/service/wallet_service.go:259`）
**问题**: 调用 `LockForUpdate(ctx, req.CustomerID)` 传入的是 customer ID，但应该传入 wallet ID。

**解决方案**:
- 先通过 `GetByCustomerID()` 获取 wallet，拿到 wallet.ID
- 在事务内部使用 `FOR UPDATE` 锁定正确的 wallet 行
- 重新检查锁定后的余额和状态，避免竞态条件

#### 3. 实现 tenant_id 认证中间件
**问题**: 代码中从 context 获取 tenant_id，但没有中间件设置它，导致多租户隔离失效。

**解决方案**:
- 创建了 `internal/interface/http/middleware/auth.go`
- 实现了 `TenantMiddleware()` 从 HTTP header `X-Tenant-ID` 提取租户 ID
- 实现了 `UserIDMiddleware()` 和 `TraceIDMiddleware()` 用于追踪
- 在 `router.go` 中应用到所有 `/api/*` 路由

**新增文件**:
- `internal/interface/http/middleware/auth.go`

### ✅ P1 - 基础设施完善

#### 4. 实现健康检查和 Metrics 端点
**解决方案**:
- 创建了 `pkg/observability/health.go` 提供健康检查服务
  - `/health` - 整体健康状态（含数据库检查）
  - `/health/liveness` - 存活检查
  - `/health/readiness` - 就绪检查
- 创建了 `pkg/observability/metrics.go` 提供 Prometheus metrics
  - `/metrics` - Prometheus 格式的指标数据
- 在 main.go 中启动两个后台服务器:
  - Health check server: 端口 8081
  - Metrics server: 端口 9090

**新增文件**:
- `pkg/observability/health.go`
- `pkg/observability/metrics.go`

#### 5. 添加 Dockerfile
**解决方案**:
- 创建多阶段构建 Dockerfile
  - Builder 阶段: 使用 golang:1.25-alpine 编译
  - Final 阶段: 使用 alpine:latest 运行
- 非 root 用户运行，提高安全性
- 暴露端口: 8085 (HTTP), 9085 (gRPC), 8081 (Health), 9090 (Metrics)
- 创建 `.dockerignore` 减小镜像大小

**新增文件**:
- `Dockerfile`
- `.dockerignore`

## 项目状态

### ✅ 已修复的核心问题
1. ✅ 数据库事务原子性
2. ✅ 行锁参数错误
3. ✅ 多租户认证中间件
4. ✅ 健康检查端点
5. ✅ Metrics 端点
6. ✅ Docker 容器化

### 📋 后续待实现功能（P2 和 P3）

#### P2 - 完善功能
- [ ] gRPC 服务实现（README 声明但未实现）
- [ ] 单元测试（提高代码质量）
- [ ] JWT 认证中间件（增强安全性）

#### P3 - 未来规划
- [ ] DTM 分布式事务集成
- [ ] 外部支付网关（WeChat、Alipay、Stripe）
- [ ] 限流功能
- [ ] Webhook 通知机制
- [ ] Mora 框架集成（或移除相关声明）

## 如何使用

### 本地运行
```bash
# 构建
make build

# 运行
make run

# 测试
make test
```

### Docker 运行
```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run

# 或者使用 docker 命令
docker build -t plutus:latest .
docker run -p 8085:8085 -p 8081:8081 -p 9090:9090 plutus:latest
```

### API 调用示例

**注意**: 所有 API 调用必须包含 `X-Tenant-ID` header！

```bash
# 创建钱包
curl -X POST http://localhost:8085/api/wallets \
  -H "X-Tenant-ID: 1" \
  -H "Content-Type: application/json" \
  -d '{"customer_id": 1001, "currency": "CNY"}'

# 充值
curl -X POST http://localhost:8085/api/transactions/recharge \
  -H "X-Tenant-ID: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1001,
    "amount": 100.50,
    "currency": "CNY",
    "channel": "wechat"
  }'

# 健康检查
curl http://localhost:8081/health

# Metrics
curl http://localhost:9090/metrics
```

## 技术改进点

1. **事务安全**: 所有资金操作现在都在数据库事务中，保证 ACID 特性
2. **并发安全**: Consume 操作使用 `FOR UPDATE` 行锁，防止超额扣款
3. **多租户隔离**: 通过中间件强制要求 tenant_id，确保数据隔离
4. **可观测性**: 提供健康检查和 metrics，方便监控和运维
5. **容器化**: Docker 支持，便于部署和扩展
6. **安全性**: Docker 非 root 用户运行，减少安全风险

## 重要提醒

⚠️ **生产环境部署前需要**:
1. 配置正确的数据库连接（`configs/plutus.yaml`）
2. 设置 JWT secret（如果实现 JWT 认证）
3. 配置生产级别的日志系统
4. 实现真实的 Prometheus metrics 收集（当前是静态示例）
5. 添加单元测试和集成测试
6. 实施 gRPC 服务（如果需要）
7. 配置反向代理（Nginx/Traefik）和 TLS 证书
