# TODO README - AI Implementation Plan

本文件列出需要交由 AI 模型落地的主要 TODO，围绕缺失的预约（appointments）与员工（staff）服务，以及 Clotho 作为编排层尚未覆盖的 API 能力。执行这些 TODO 时，请严格遵循仓库根目录 README 中描述的模块结构与 go.work 管理方式。

## 1. 新增预约服务 (`appointments/`)
1. **创建模块骨架**：在仓库根目录新增 `appointments` 目录，包含 `cmd/appointments`、`internal/{interface,application,domain,infrastructure}`、`configs/appointments.yaml`、`docs/`、`Makefile` 等，与现有服务（如 `hermes`）保持一致；在 `go.work` 中加入该模块。
2. **API 实现**：按照 `API_Documentation.md` 中的预约章节（GET/POST/PUT/DELETE `/appointments`、日历视图、可用时间、状态更新）实现 HTTP 适配层；若存在 gRPC 需求，生成 `api/proto` 定义并落地 `internal/interface/grpc`。
3. **领域建模与应用层**：设计预约实体、状态枚举、冲突检测规则、提醒策略；实现用例服务（创建/重排/取消/状态流转）并加入必要的事务或分布式锁策略。
4. **基础设施层**：提供 MySQL migrations、仓储接口实现、缓存/异步任务（提醒通知）支持，并补全配置项（数据库、Redis、队列等）。
5. **测试与文档**：补充单元测试（`testify`），并在 `docs/` 中更新 swagger、时序图及注意事项；`make test` 需覆盖创建/冲突检测/提醒。

## 2. 新增员工服务 (`staff/`)
1. **模块骨架**：和预约服务相同的目录结构（`cmd/staff` 等）与 go.work 配置。
2. **API 实现**：覆盖 `API_Documentation.md` 的员工接口（列表、CRUD、状态更新、角色列表、关联预约查询）；在 HTTP 层补充过滤、分页、排序逻辑，并支撑未来的 gRPC 访问。
3. **领域与应用逻辑**：实现员工实体、角色/部门/岗位管理、可用性日历（供预约模块调用），并提供批量导入/导出占位接口。
4. **基础设施与安全**：实现仓储（MySQL/GORM）、头像存储占位、与 Custos 的权限同步（定期从 Custos 拉取角色/用户绑定）；暴露缓存层接口供 Clotho 缓存员工摘要。
5. **测试与文档**：围绕 CRUD、状态切换、权限校验补充测试，生成 swagger，并在 `docs/` 提供 ER 图与流程说明。

## 3. Clotho API 支持缺口
当前 `clotho/internal/infrastructure/http/router.go` 仅注册了 `/users`、`/profile`、`/monitoring` 路由，并只封装了 `custos` gRPC 客户端，尚未代理其他服务。需要交由 AI 完成以下事项：
1. **服务客户端与配置**：为 hermes（客户）、kratos（订单）、plutus（支付）、新建的 appointments/staff 服务分别实现 HTTP/gRPC 客户端或适配层，扩展 `services.<name>.address` 配置与连接管理（熔断、重试、负载均衡）。
2. **应用用例**：在 `internal/application/usecase` 内为每个后端服务新增用例（CustomerProxy、OrderProxy、PaymentProxy、AppointmentProxy、StaffProxy 等），统一鉴权上下文，并在必要时实现跨服务编排（例如创建预约时联合 staff 可用性检查与 Hermes 客户数据）。
3. **HTTP Handler & 路由**：新增路由组 `/customers`、`/orders`、`/payments`、`/appointments`、`/employees` 等，完整覆盖 API 文档给出的接口；确保请求/响应体与 `API_Documentation.md` 一致，并在 swagger 注释中同步描述。
4. **认证与授权联动**：复用 Custos token 校验，在 Clotho 层根据 `permissions` 字段做细粒度授权拦截；对需要员工/预约权限的路由添加策略占位。
5. **监控与测试**：为新增代理编写集成测试（可使用 mock server 或 fake gRPC 服务），扩展 Prometheus 指标与日志字段（记录目标服务延迟/错误率）；更新 `docs/` 以描述新的路由编排关系。

## 4. 交付与验证要求
- 更新 `README.md`、`Fly_项目模块开发进度分析报告.md` 以反映新模块与 Clotho 能力。
- 运行 `go work sync`、各模块 `make build && make test`、`./scripts/test-compile.sh`，并在 PR 中附测试结果。
- 对涉及数据库的变更，提供迁移脚本与回滚说明。

完成以上 TODO 后，Clotho 将成为真正覆盖所有业务域的 API 门面，同时预约与员工服务也会具备独立部署能力。
