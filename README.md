

# Fly Monorepo

**Fly** 是一个采用 Monorepo 模式管理的分布式服务架构项目。  
它的目标是通过清晰的模块边界和共享能力库，加速独立开发者从需求到交付的效率。

---

## 📦 仓库结构

```
fly/
  go.work
  mora/        # 通用能力库
  custos/      # 用户域（身份 & 权限中心）
  clotho/      # API 编排层
  orders/      # (未来扩展) 订单域
  payments/    # (未来扩展) 支付域
  docs/        # 文档与架构说明
  deploy/      # 部署配置 (docker-compose / k8s)
```

---

## 🔹 模块说明

### 1. Mora （能力库）
- 定位：框架无关的通用能力库  
- 功能：JWT/JWK、日志、配置、数据库、缓存、消息队列、工具方法  
- 特点：不依赖业务逻辑，可被所有服务使用  

### 2. Custos （用户域）
- 定位：身份与权限中心  
- 功能：用户表、账号体系、本地/外部 OAuth2.0 登录、内部 Token 签发、RBAC、强制下线  
- 特点：唯一 Token 签发者，其他服务只消费 Custos Token  

### 3. Clotho （API 编排层）
- 定位：对外统一 API 门面  
- 功能：Token 校验、请求编排、聚合结果返回  
- 特点：不签发 Token，只做编排和网关逻辑  

### 4. Orders / Payments （未来扩展）
- 独立的业务域，专注各自领域逻辑  
- 通过 Clotho 对外统一输出  

---

## 🔹 架构原则
1. **分层清晰**：Mora → Custos → Clotho → 业务域  
2. **统一 Token 签发**：所有身份认证由 Custos 管理  
3. **API 门面统一**：Clotho 是唯一对外暴露层  
4. **开发友好**：Monorepo + go.work，保证多模块协作丝滑  

---

## 🔹 下一步
- 完善各模块内的 README（详细说明能力和边界）  
- 增加 docker-compose 一键启动环境  
- 集成 CI/CD（按目录触发构建）  
- 补充监控与日志收集方案  