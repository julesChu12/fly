# Fly 项目模块开发进度分析报告

**生成日期**: 2025-09-29
**分析范围**: Fly Monorepo 全部模块
**版本**: v1.0

---

## 📋 执行摘要

Fly 是一个采用 Monorepo 架构的分布式微服务项目，目标为小微企业提供轻量级 CRM 后端服务。项目采用清晰的分层架构，包含基础设施层、业务域层和 API 编排层，具备良好的可扩展性和开发效率。

**当前开发状态**: 🟢 核心基础设施完成，业务模块开发中

---

## 🏗️ 项目架构概览

```text

fly/
├── mora/             # 🟢 通用能力库 (已完成)
├── custos/           # 🟢 用户域服务 (已完成)
├── clotho/           # 🟡 API编排层 (开发中)
├── observability/    # 🟡 可观测性 (部分完成)
└── [未来模块]         # 🔴 业务域服务 (待开发)
    ├── orders/       # 订单管理
    ├── customers/    # 客户管理
    ├── services/     # 服务管理
    ├── appointments/ # 预约调度
    ├── inventory/    # 库存管理
    ├── payments/     # 支付结算
    ├── marketing/    # 营销管理
    └── analytics/    # 报表分析
```text


---

## 📊 模块详细分析

### 1. Mora - 通用能力库

**📍 模块位置**: `/mora/`
**🚀 开发状态**: ✅ 已完成 (95%)
**🎯 职责边界**: 跨服务基础设施组件，为所有业务服务提供通用能力

#### 功能实现情况

| 功能模块 | 状态 | 实现度 |
|---------|------|--------|
| JWT认证 | ✅ 完成 | 100% |
| 配置管理 | ✅ 完成 | 100% |
| 日志组件 | ✅ 完成 | 100% |
| 数据库适配器 | ✅ 完成 | 90% |
| 缓存组件 | ✅ 完成 | 90% |
| 消息队列 | ⚠️ 基础实现 | 60% |
| 工具库 | ✅ 完成 | 80% |
#### Custos 技术栈选择

- **核心语言**: Go 1.25.1
- **Web框架适配**: Gin + GoZero 支持
- **数据库支持**: MySQL, PostgreSQL, SQLite (GORM + SQLx)
- **缓存**: Redis
- **认证**: JWT + 自定义Claims
- **可观测性**: OpenTelemetry

#### Custos 技术栈优势

✅ **框架无关设计**: 核心功能独立于具体框架，支持多种适配器
✅ **类型安全**: 充分利用 Go 的类型系统，减少运行时错误
✅ **高性能**: Go 原生并发优势，适合高并发场景
✅ **易于测试**: 接口驱动设计，便于单元测试和模拟

#### Custos 技术栈劣势

⚠️ **学习成本**: 开发者需要理解适配器模式和接口设计
⚠️ **抽象层级**: 增加了一定的复杂度，简单项目可能过度设计

#### Custos 对外接口

**类型**: Go Package 导入
**接入方式**:

```go
import "github.com/julesChu12/fly/mora/pkg/auth"
import "github.com/julesChu12/fly/mora/adapters/gin"
```text


#### Mora 未来扩展方向

- 完善消息队列组件 (RabbitMQ, Kafka 支持)
- 增加更多数据库适配器 (MongoDB, ClickHouse)
- 完善链路追踪和指标收集

---

### 2. Custos - 用户域服务

**📍 模块位置**: `/custos/`
**🚀 开发状态**: ✅ 已完成 (90%)
**🎯 职责边界**: 统一身份认证与权限管理中心，是整个系统的唯一 Token 签发者

#### 功能实现情况

| 功能模块 | 状态 | 实现度 |
|---------|------|--------|
| 用户注册/登录 | ✅ 完成 | 100% |
| JWT Token管理 | ✅ 完成 | 100% |
| OAuth2.0集成 | ✅ 完成 | 85% |
| RBAC权限控制 | ✅ 完成 | 90% |
| 会话管理 | ✅ 完成 | 95% |
| 用户生命周期 | ✅ 完成 | 85% |
| 强制下线 | ✅ 完成 | 100% |
#### Custos 技术栈选择
- **Web框架**: Gin
- **数据库**: MySQL + GORM
- **权限框架**: Casbin
- **认证**: JWT + OAuth2.0
- **数据迁移**: sql-migrate
- **测试**: testify

#### Custos 技术栈优势
✅ **成熟生态**: Gin + GORM + Casbin 是 Go 生态中的成熟组合
✅ **权限灵活**: Casbin 支持复杂的权限控制策略
✅ **OAuth便利**: 支持主流第三方登录 (Google, GitHub)
✅ **数据一致性**: MySQL事务保证数据一致性

#### Custos 技术栈劣势
⚠️ **单点依赖**: MySQL 成为性能瓶颈时需要分库分表
⚠️ **Casbin复杂性**: 权限策略配置需要学习成本

#### Custos 对外接口


**1. HTTP REST API (端口 8081)**

```text

基础认证:
  POST /api/v1/auth/register     - 用户注册
  POST /api/v1/auth/login        - 用户登录
  POST /api/v1/auth/refresh      - 刷新Token
  POST /api/v1/auth/logout       - 用户登出
  POST /api/v1/auth/logout-all   - 全设备登出

OAuth认证:
  GET  /api/v1/oauth/:provider/login     - OAuth登录URL
  GET  /api/v1/oauth/:provider/callback  - OAuth回调处理
  POST /api/v1/oauth/:provider/bind      - 绑定OAuth账号
  DELETE /api/v1/oauth/:provider/unbind  - 解绑OAuth账号
  GET  /api/v1/oauth/bindings            - 查询OAuth绑定

用户管理:
  GET  /api/v1/user/profile              - 获取用户档案

管理功能:
  GET    /api/v1/admin/users             - 用户列表
  GET    /api/v1/admin/users/:id         - 获取用户详情
  PATCH  /api/v1/admin/users/:id/status  - 更新用户状态
  PATCH  /api/v1/admin/users/:id/role    - 更新用户角色
  POST   /api/v1/admin/users/:id/force-logout - 强制用户下线
  GET    /api/v1/admin/stats             - 系统统计
```text



**2. gRPC 内部接口 (端口 9001 - 计划中)**

```protobuf
service CustosService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
  rpc GetUserPreferences(GetUserPreferencesRequest) returns (GetUserPreferencesResponse);
  rpc UpdateUserPreferences(UpdateUserPreferencesRequest) returns (UpdateUserPreferencesResponse);
  rpc GetUserStatistics(GetUserStatisticsRequest) returns (GetUserStatisticsResponse);
}
```text


#### Custos 未来扩展方向

- 完善 gRPC 接口实现
- 增加多租户支持 (tenant_id 已预留)
- 完善用户偏好设置功能
- 增加用户行为统计分析

---

### 3. Clotho - API编排层

**📍 模块位置**: `/clotho/`
**🚀 开发状态**: 🟡 开发中 (70%)
**🎯 职责边界**: 对外统一API门面，负责请求编排、认证校验、结果聚合，不包含业务逻辑

#### Clotho 功能实现情况

| 功能模块 | 状态 | 实现度 |
|---------|------|--------|
| HTTP路由框架 | ✅ 完成 | 100% |
| JWT认证中间件 | ✅ 完成 | 100% |
| 限流熔断中间件 | ✅ 完成 | 100% |
| 可观测性集成 | ✅ 完成 | 95% |
| gRPC客户端 | ⚠️ 部分实现 | 60% |
| 请求编排逻辑 | ⚠️ 基础实现 | 40% |
| Swagger文档 | ✅ 完成 | 90% |
#### Custos 技术栈选择
- **Web框架**: Gin
- **服务间通信**: gRPC
- **限流**: golang.org/x/time/rate + 自定义实现
- **熔断**: sony/gobreaker
- **可观测性**: OpenTelemetry + Prometheus + Jaeger
- **配置管理**: Viper
- **API文档**: Swagger/OpenAPI

#### Custos 技术栈优势
✅ **高性能**: Gin + gRPC 组合保证高吞吐量
✅ **可观测性完善**: 全链路追踪和指标收集
✅ **弹性设计**: 限流熔断保护下游服务
✅ **文档自动化**: Swagger 自动生成 API 文档

#### Custos 技术栈劣势
⚠️ **复杂度**: 多种中间件集成增加了系统复杂度
⚠️ **调试难度**: 分布式环境下问题定位较困难

#### Custos 对外接口


**HTTP REST API (端口 8080)**

```text

系统接口:
  GET /health                              - 健康检查
  GET /metrics                             - Prometheus指标
  GET /swagger/*any                        - API文档

用户代理接口:
  GET /api/v1/users/me                     - 当前用户信息
  GET /api/v1/users/:id                    - 获取用户信息

用户档案管理:
  GET /api/v1/profile/                     - 获取用户档案
  PUT /api/v1/profile/                     - 更新用户档案
  PUT /api/v1/profile/preferences          - 更新用户偏好
  GET /api/v1/profile/users/:id            - 获取其他用户公开档案

中间件监控:
  GET /api/v1/monitoring/stats             - 所有中间件统计
  GET /api/v1/monitoring/rate-limiter      - 限流器统计
  GET /api/v1/monitoring/circuit-breaker   - 熔断器统计
  POST /api/v1/monitoring/circuit-breaker/reset - 重置熔断器

未来业务接口 (计划):
  /api/v1/orders/*                         - 订单相关接口
  /api/v1/customers/*                      - 客户相关接口
  /api/v1/appointments/*                   - 预约相关接口
  /api/v1/services/*                       - 服务相关接口
```text


#### 限流熔断配置
```yaml
rate_limiter:
  global_rps: 1000.0      # 全局每秒请求数
  global_burst: 2000      # 全局突发请求数
  per_ip_rps: 10.0        # 每IP每秒请求数
  per_ip_burst: 20        # 每IP突发请求数
  per_user_rps: 100.0     # 每用户每秒请求数
  per_user_burst: 200     # 每用户突发请求数

circuit_breaker:
  max_requests: 5         # 半开状态最大请求数
  failure_threshold: 5    # 失败阈值
  failure_ratio: 0.6      # 失败率阈值
  min_requests: 10        # 最小请求数
```text


#### Clotho 未来扩展方向
- 完善 gRPC 客户端连接池
- 实现更多业务域的代理接口
- 增加 API 版本管理
- 完善错误处理和重试机制

---

### 4. Observability - 可观测性服务

**📍 模块位置**: `/observability/`
**🚀 开发状态**: 🟡 部分完成 (60%)
**🎯 职责边界**: 提供全栈监控、链路追踪、指标收集和日志聚合

#### Observability 功能实现情况

| 功能模块 | 状态 | 实现度 |
|---------|------|--------|
| Jaeger链路追踪 | ✅ 完成 | 95% |
| Prometheus指标 | ⚠️ 配置完成 | 70% |
| 应用指标集成 | ✅ 完成 | 85% |
| Grafana仪表板 | ❌ 未实现 | 0% |
| 日志聚合 | ❌ 未实现 | 0% |
| 告警规则 | ❌ 未实现 | 0% |
#### Custos 技术栈选择
- **链路追踪**: Jaeger (全链路)
- **指标收集**: Prometheus
- **应用指标**: OpenTelemetry
- **可视化**: 计划使用 Grafana
- **日志**: 计划使用 ELK Stack

#### Custos 对外接口


**监控访问地址**

```text

Jaeger UI:     http://localhost:16686  - 分布式链路追踪
Prometheus UI: http://localhost:9090   - 指标查询 (待启用)

OTLP接收端点:
  gRPC: localhost:4317                 - OpenTelemetry gRPC
  HTTP: localhost:4318                 - OpenTelemetry HTTP
```text



**应用指标端点**

```text

Custos指标:   http://localhost:8081/metrics
Clotho指标:   http://localhost:8080/metrics
Jaeger指标:   http://localhost:14269/metrics
```text


#### Observability 未来扩展方向
- 启用 Prometheus 监控
- 集成 Grafana 仪表板
- 实现日志聚合方案
- 配置告警规则和通知

---

## 🌐 容器化部署架构

**部署方式**: Docker Compose
**配置文件**: `/docker-compose.yaml`

```yaml
服务清单:
  mysql:        数据存储 (端口 3306)
  redis:        缓存服务 (端口 6379)
  custos:       用户服务 (端口 8081)
  clotho:       API网关 (端口 8080)
  jaeger:       链路追踪 (端口 16686)
  prometheus:   指标监控 (端口 9090) - 待启用
```text


**服务依赖关系**:
```text

mysql/redis → custos → clotho
                   ↓
              jaeger/prometheus
```text


---

## 🎯 CRM系统缺失模块分析

基于小微企业 CRM 系统需求 (理发店、洗车店等)，以下是缺失的业务模块：

### 🔴 高优先级 - 核心业务模块

#### 1. Customers - 客户管理服务
**职责**: 客户信息、客户分类、客户关系维护
**预计开发周期**: 2-3周
**技术栈建议**: Go + Gin + GORM + MySQL
**核心功能**:
- 客户档案管理 (联系方式、偏好、历史记录)
- 客户分类标签 (VIP、普通、潜在客户)
- 客户关系维护 (生日提醒、回访记录)

#### 2. Services - 服务管理服务
**职责**: 服务项目、价格体系、服务流程管理
**预计开发周期**: 2周
**技术栈建议**: Go + Gin + GORM + MySQL
**核心功能**:
- 服务项目管理 (理发、洗车项目定义)
- 价格体系管理 (基础价格、会员价格、促销价格)
- 服务流程配置 (标准化作业流程)

#### 3. Appointments - 预约调度服务
**职责**: 预约管理、时间调度、资源分配
**预计开发周期**: 3-4周
**技术栈建议**: Go + Gin + GORM + MySQL + Redis
**核心功能**:
- 预约时间管理 (时段分配、冲突检测)
- 员工调度 (技师/员工排班)
- 资源预定 (设备、房间预定)
- 预约提醒 (短信、微信通知)

#### 4. Orders - 订单管理服务
**职责**: 订单生命周期、计费结算、订单统计
**预计开发周期**: 2-3周
**技术栈建议**: Go + Gin + GORM + MySQL
**核心功能**:
- 订单创建与流转 (预约→服务→结算)
- 计费逻辑 (服务费、优惠计算)
- 订单状态管理 (待服务、进行中、已完成、已取消)

### 🟡 中优先级 - 运营支撑模块

#### 5. Payments - 支付结算服务
**职责**: 多渠道支付、对账、财务管理
**预计开发周期**: 3-4周
**技术栈建议**: Go + Gin + 支付SDK集成
**核心功能**:
- 多支付渠道 (现金、微信、支付宝、银行卡)
- 支付对账 (日结、月结)
- 财务报表 (收入统计、成本核算)

#### 6. Inventory - 库存管理服务
**职责**: 库存商品、进销存、库存预警
**预计开发周期**: 2-3周
**技术栈建议**: Go + Gin + GORM + MySQL
**核心功能**:
- 商品管理 (洗发水、汽车用品等)
- 进销存管理 (采购、销售、库存)
- 库存预警 (低库存提醒、过期提醒)

#### 7. Staff - 员工管理服务
**职责**: 员工档案、排班管理、绩效考核
**预计开发周期**: 2周
**技术栈建议**: Go + Gin + GORM + MySQL
**核心功能**:
- 员工档案管理 (基本信息、技能认证)
- 排班管理 (班次安排、请假管理)
- 绩效统计 (服务量、客户评价、收入统计)

### 🟢 低优先级 - 增值功能模块

#### 8. Marketing - 营销管理服务
**职责**: 会员体系、促销活动、客户营销
**预计开发周期**: 3-4周
**核心功能**:
- 会员等级体系 (积分、等级权益)
- 促销活动管理 (优惠券、折扣活动)
- 营销推送 (短信营销、生日营销)

#### 9. Analytics - 报表分析服务
**职责**: 业务数据分析、经营报表、决策支持
**预计开发周期**: 4-5周
**核心功能**:
- 经营报表 (日/周/月报表)
- 客户分析 (客户画像、流失分析)
- 业务分析 (服务热度、员工效率)

#### 10. Notifications - 通知服务
**职责**: 统一通知推送、消息模板、通知记录
**预计开发周期**: 2周
**核心功能**:
- 多渠道通知 (短信、微信、邮件)
- 消息模板管理
- 通知历史记录

---

## 🚀 开发优先级建议

### Phase 1: 核心业务MVP (预计 8-10周)
1. **Customers** (客户管理) - 3周
2. **Services** (服务管理) - 2周
3. **Appointments** (预约调度) - 4周
4. **Orders** (订单管理) - 3周

### Phase 2: 运营支撑 (预计 6-8周)
1. **Payments** (支付结算) - 4周
2. **Staff** (员工管理) - 2周
3. **Inventory** (库存管理) - 3周

### Phase 3: 增值功能 (预计 8-10周)
1. **Notifications** (通知服务) - 2周
2. **Marketing** (营销管理) - 4周
3. **Analytics** (报表分析) - 5周

---

## 📈 技术债务与优化建议

### 当前技术债务
1. **Observability**: Prometheus 未启用，缺少完整监控
2. **Clotho**: gRPC 客户端连接池未完善
3. **Mora**: 消息队列组件需要增强
4. **测试覆盖**: 整体测试覆盖率偏低

### 架构优化建议
1. **数据库分离**: 考虑读写分离，提升数据库性能
2. **缓存策略**: 增加分布式缓存，减少数据库压力
3. **配置中心**: 统一配置管理，支持动态配置更新
4. **服务治理**: 增加服务注册发现、健康检查

### 开发效率提升
1. **代码生成**: 基于 Proto 自动生成 CRUD 代码
2. **CI/CD**: 实现自动化测试、构建、部署流水线
3. **开发工具**: 提供本地开发环境一键搭建脚本

---

## 🎯 总结与建议

### 项目优势
✅ **架构清晰**: 分层明确，职责边界清楚
✅ **技术选型合理**: Go生态成熟稳定，适合快速开发
✅ **可扩展性强**: Monorepo + 微服务，便于团队协作
✅ **基础设施完善**: 认证、可观测性、容器化已就绪

### 关键建议
1. **先完善基础设施**: 优先解决 Prometheus 监控、完善测试覆盖
2. **按优先级开发**: 建议按 Phase 1 → Phase 2 → Phase 3 顺序开发
3. **重视数据模型**: 客户、订单、预约等核心实体设计要前瞻性考虑
4. **API设计统一**: 制定 API 设计规范，保证接口一致性

### 开发资源估算
- **当前基础设施**: 已完成约 70%
- **CRM核心功能**: 预计需要 22-28 周完成
- **团队规模建议**: 3-5 人(2个后端 + 1个前端 + 1个测试 + 1个DevOps)

---


**报告结束**



*本报告基于 2025-09-29 的代码分析生成，建议定期更新以反映最新开发进度*

