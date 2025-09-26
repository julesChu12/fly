# Mora – AI Context README

## 项目定位
- **mora** 是一个通用能力库（工具层）。  
- Mora 提供 JWT/JWK、logger、db、cache 等工具能力，但不会直接实现登录/刷新，这属于 UserService (Custos)。  
- Mora 负责 Access Token 的签名验证，支持 JWKS 或公钥方式校验。  

---

## 架构边界
1. **UserService 的职责**  
   - 签发 Access Token（短期）和 Refresh Token（长期）。  
   - 维护 Refresh Token 状态表（存储在 Redis 或数据库）。  
   - 处理 Refresh Token 的旋转与废弃。  
   - 提供认证相关 API：`/login` `/refresh` `/logout` `/introspect`。  
   - 负责 RBAC（细粒度权限控制）逻辑。  

2. **Mora 的职责**  
   - 接收客户端请求，要求附带 **Access Token**。  
   - 校验 Access Token 的签名与 claims（支持 JWKS 或公钥验证）。  
   - 如果 Access Token 过期，客户端需要去 **UserService** 调用 `/refresh`。  
   - Mora 不直接管理用户状态表，不维护黑名单。  
   - 提供可插拔的认证中间件，便于集成不同框架。  
   - 在需要细粒度鉴权时，可调用 UserService 的 `/introspect`。  

---

## Token 策略
- **Access Token**  
  - 有效期：10 分钟  
  - 用途：访问 Mora 的业务接口  
  - 存储：不存储，仅验证签名，支持 JWKS 或公钥验证  

- **Refresh Token**  
  - 有效期：7 天  
  - 用途：刷新 Access Token  
  - 存储：存储在 UserService 的 Redis/DB 状态表中  
  - 支持旋转：每次刷新生成新 Refresh Token，废弃旧的  

---

## Mora 的接口规则
- 所有业务接口必须要求 **Authorization: Bearer <AccessToken>**  
- Mora 不提供登录、刷新、退出接口，这些由 **UserService** 提供  
- Mora 只需要：  
  - 校验 Access Token 签名和 claims（支持 JWKS 或公钥）  
  - 拒绝过期或非法 token  
  - 认证中间件设计为可插拔，方便不同框架集成  
  - 细粒度权限控制（RBAC）由 UserService 负责  
  - 需要时调用 UserService 的 `/introspect` 进行鉴权  

---

## AI 使用提示
当你为 **mora** 生成代码时：  
- 不要实现登录/刷新逻辑，这些逻辑属于 Custos  
- Mora 内部只需要写 **Access Token 校验中间件**，支持 JWKS 或公钥验证  
- 认证中间件应设计为可插拔，方便在 Gin、go-zero 等框架中使用  
- Token 过期时，直接返回 401，让客户端去找 UserService  
- 如需示例，请用 Go (Gin、go-zero 框架) 展示如何拦截和校验 Access Token  

---

# PROMPT.md

# Mora Directory Context – AI Prompt

## Mora 目录职责说明

- Mora 是一个通用能力库，提供基础工具和能力支持，聚焦于 Access Token 的验证和相关工具的封装。  
- Mora 不实现登录、刷新、用户状态管理、RBAC 权限控制等业务逻辑，这些都由 UserService（Custos）负责。  
- Mora 支持基于 JWT 的 Access Token 验证，支持 JWKS 或直接公钥的验证方式。  
- Mora 提供认证中间件，设计为可插拔，方便集成到不同的 Go 框架（如 Gin、go-zero）。  
- Mora 仅负责校验 Access Token 签名和 claims，过期或非法 token 直接拒绝请求。  
- Mora 在需要时可调用 UserService 的 `/introspect` API 进行细粒度鉴权。  

## AI 生成代码注意事项

- 请勿在 Mora 实现登录、刷新、用户管理、RBAC 等业务逻辑。  
- 认证中间件应支持 JWKS 和公钥两种验证方式。  
- 认证中间件设计为可插拔，便于在不同框架中集成。  
- Token 过期或无效时，返回 401 状态码，由客户端去 UserService 处理刷新。  
- 如需示例，优先使用 Go 语言，结合 Gin 或 go-zero 框架演示 Access Token 的拦截与验证。  
- 保持 Mora 代码库的职责单一，专注于通用工具和 Access Token 验证。
