# 新增服务文档

## 概述

本文档详细介绍了 Fly CRM 系统新增的两个核心服务：预约管理服务 (Appointments) 和员工管理服务 (Staff)。这两个服务完善了系统的业务闭环，为完整的客户关系管理提供了必要的支持。

---

## 📅 Appointments Service (预约管理服务)

### 服务概述

- **服务名称**: Appointments Service
- **端口配置**: HTTP 8083, gRPC 9083
- **定位**: 预约调度与时间管理中心
- **数据库名**: appointments

### 核心功能

#### 1. 预约管理
- **创建预约**: 支持客户、员工、服务的关联预约创建
- **预约查询**: 多维度查询（客户ID、员工ID、时间段、状态等）
- **预约更新**: 灵活的预约信息修改机制
- **预约删除**: 支持软删除，保留历史数据

#### 2. 状态管理
- **状态流转**: pending → confirmed → in_progress → completed
- **状态控制**: 支持取消 (cancelled) 和未到店 (no_show)
- **业务规则**: 状态转换验证，确保业务逻辑正确性

#### 3. 冲突检测
- **智能检测**: 自动检测时间冲突，避免重复预约
- **可用时间**: 实时查询员工可用时间段
- **建议推荐**: 冲突时提供其他可用时间建议

#### 4. 提醒系统
- **提醒设置**: 灵活的提醒时间配置
- **提醒触发**: 基于时间的自动提醒机制
- **提醒处理**: 批量处理待发送提醒

#### 5. 日历视图
- **视图查询**: 按员工、时间范围查询日历视图
- **事件展示**: 完整的预约事件信息展示
- **格式支持**: 支持多种日历数据格式

### API 接口

#### 基础 CRUD
```bash
POST   /api/appointments              # 创建预约
GET    /api/appointments              # 获取预约列表
GET    /api/appointments/{id}         # 获取预约详情
PUT    /api/appointments/{id}         # 更新预约
DELETE /api/appointments/{id}         # 删除预约
PUT    /api/appointments/{id}/status  # 更新预约状态
```

#### 业务接口
```bash
GET    /api/appointments/calendar         # 获取日历视图
GET    /api/appointments/availability     # 检查可用时间
POST   /api/appointments/conflict-check   # 检查时间冲突
GET    /api/appointments/customer/{id}    # 获取客户预约
GET    /api/appointments/employee/{id}     # 获取员工预约
GET    /api/appointments/employee/{id}/upcoming # 获取即将到来的预约
```

### 数据模型

#### Appointment (预约实体)
```go
type Appointment struct {
    ID           uuid.UUID         // 预约唯一标识
    CustomerID   uuid.UUID         // 客户ID
    EmployeeID   uuid.UUID         // 员工ID
    ServiceID    uuid.UUID         // 服务ID
    StartTime    time.Time         // 开始时间
    EndTime      time.Time         // 结束时间
    Notes        *string           // 备注信息
    Status       AppointmentStatus // 预约状态
    Reminder     bool              // 是否提醒
    ReminderTime *time.Time        // 提醒时间
    CreatedAt    time.Time         // 创建时间
    UpdatedAt    time.Time         // 更新时间
    DeletedAt    *time.Time        // 删除时间
}
```

#### 状态枚举
- `pending`: 待确认
- `confirmed`: 已确认
- `in_progress`: 进行中
- `completed`: 已完成
- `cancelled`: 已取消
- `no_show`: 未到店

### 技术特点

1. **双协议支持**: HTTP RESTful API + gRPC
2. **冲突检测**: 智能时间冲突算法
3. **状态机**: 完整的状态流转控制
4. **缓存优化**: 热点数据缓存机制
5. **事务支持**: 关键操作的事务保障
6. **链路追踪**: 分布式追踪支持

---

## 👥 Staff Service (员工管理服务)

### 服务概述

- **服务名称**: Staff Service
- **端口配置**: HTTP 8084, gRPC 9084
- **定位**: 员工信息与权限管理中心
- **数据库名**: staff

### 核心功能

#### 1. 员工管理
- **员工档案**: 完整的员工信息管理
- **组织架构**: 部门、岗位、职级管理
- **员工状态**: 在职、离职、休假、停职状态管理
- **入职管理**: 入职流程和档案建立

#### 2. 角色权限
- **角色定义**: 灵活的角色管理系统
- **权限分配**: 细粒度的权限控制
- **角色继承**: 支持角色的权限继承
- **权限审计**: 完整的权限变更记录

#### 3. 可用性管理
- **工作时间**: 灵活的工作时间配置
- **可用性设置**: 员工实时可用状态管理
- **排班管理**: 智能排班和可用性查询
- **技能标签**: 员工技能和专长管理

#### 4. 薪资管理
- **薪资设置**: 员工薪资信息管理
- **薪资调整**: 薪资变更记录
- **薪资统计**: 部门和角色薪资分析

#### 5. 统计分析
- **员工统计**: 各类员工数量统计
- **部门分析**: 部门人员分布分析
- **角色分析**: 角色分布和权限分析
- **成本分析**: 人力成本统计分析

### API 接口

#### 员工管理
```bash
POST   /api/staff              # 创建员工
GET    /api/staff              # 获取员工列表
GET    /api/staff/{id}         # 获取员工详情
PUT    /api/staff/{id}         # 更新员工信息
DELETE /api/staff/{id}         # 删除员工
PUT    /api/staff/{id}/status  # 更新员工状态
```

#### 角色管理
```bash
POST   /api/staff/roles        # 创建角色
GET    /api/staff/roles        # 获取角色列表
GET    /api/staff/roles/{id}   # 获取角色详情
PUT    /api/staff/roles/{id}   # 更新角色
DELETE /api/staff/roles/{id}   # 删除角色
```

#### 业务查询
```bash
GET    /api/staff/stats                    # 获取员工统计
GET    /api/staff/available               # 获取可用员工
GET    /api/staff/department/{dept}       # 按部门查询员工
GET    /api/staff/role/{roleId}           # 按角色查询员工
```

### 数据模型

#### Staff (员工实体)
```go
type Staff struct {
    ID               uuid.UUID  // 员工唯一标识
    UserID           *uuid.UUID // 关联用户ID
    Name             string     // 员工姓名
    Email            string     // 邮箱地址
    Phone            string     // 联系电话
    Gender           Gender     // 性别
    Birthday         *time.Time // 生日
    Avatar           *string    // 头像URL
    Department       string     // 所属部门
    Position         string     // 职位
    RoleID           uuid.UUID  // 角色ID
    Status           StaffStatus // 员工状态
    HireDate         *time.Time // 入职日期
    Salary           *float64   // 薪资
    Address          *string    // 地址
    EmergencyContact *string    // 紧急联系人
    Notes            *string    // 备注
    Skills           string     // 技能标签 (JSON)
    WorkingHours     *string    // 工作时间 (JSON)
    IsAvailable      bool       // 是否可用
    CreatedAt        time.Time  // 创建时间
    UpdatedAt        time.Time  // 更新时间
    DeletedAt        *time.Time // 删除时间
}
```

#### StaffRole (员工角色)
```go
type StaffRole struct {
    ID          uuid.UUID  // 角色唯一标识
    Name        string     // 角色名称
    Code        string     // 角色编码
    Description *string    // 角色描述
    Permissions string     // 权限列表 (JSON)
    IsDefault   bool       // 是否默认角色
    SortOrder   int        // 排序顺序
    Status      StaffStatus // 角色状态
    CreatedAt   time.Time  // 创建时间
    UpdatedAt   time.Time  // 更新时间
}
```

#### 状态枚举
- `active`: 在职
- `inactive`: 离职
- `on_leave`: 休假
- `suspended`: 停职

### 技术特点

1. **权限系统**: 基于RBAC的细粒度权限控制
2. **组织架构**: 灵活的部门层级管理
3. **可用性**: 实时员工可用状态管理
4. **技能标签**: 多维度员工技能管理
5. **统计分析**: 完整的人力资源统计
6. **数据安全**: 敏感信息加密存储

---

## 🔗 服务集成

### Clotho 编排层集成

Clotho 作为API编排层，已经集成了新服务的客户端：

```go
// Appointments 客户端
appointmentsClient := client.NewAppointmentsClient("http://localhost:8083", 30*time.Second)

// Staff 客户端
staffClient := client.NewStaffClient("http://localhost:8084", 30*time.Second)
```

### 路由配置

新增的路由组已经在 Clotho 中配置：

```go
// API v1 routes (auth required)
v1 := router.Group("/api/v1")
v1.Use(authMiddleware.ValidateToken())
{
    customers := v1.Group("/customers")
    appointments := v1.Group("/appointments")
    employees := v1.Group("/employees")
    orders := v1.Group("/orders")
    payments := v1.Group("/payments")
}
```

### 服务发现

所有服务支持服务发现和配置管理：

```yaml
# Clotho 配置示例
services:
  appointments:
    address: http://localhost:8083
  staff:
    address: http://localhost:8084
  hermes:
    address: http://localhost:8080
  kratos:
    address: http://localhost:8082
  plutus:
    address: http://localhost:8085
```

---

## 🚀 部署指南

### 环境要求

- Go 1.25.1+
- MySQL 8.0+
- Redis 6.0+
- Docker & Docker Compose

### 快速启动

```bash
# 1. 同步工作区
go work sync

# 2. 安装依赖
make deps

# 3. 构建服务
make build

# 4. 启动服务
make run
```

### Docker 部署

```bash
# 构建镜像
docker build -t appointments:latest ./appointments
docker build -t staff:latest ./staff

# 运行容器
docker run -d -p 8083:8083 appointments:latest
docker run -d -p 8084:8084 staff:latest
```

### 配置管理

每个服务都有独立的配置文件：

- `appointments/configs/appointments.yaml`
- `staff/configs/staff.yaml`

支持环境变量覆盖配置：

```bash
export APPOINTMENTS_SERVER_HTTP_PORT=8083
export STAFF_SERVER_HTTP_PORT=8084
```

---

## 🧪 测试策略

### 单元测试

```bash
# Appointments 测试
cd appointments && make test

# Staff 测试
cd staff && make test
```

### 集成测试

```bash
# API 接口测试
curl -X POST http://localhost:8083/api/appointments \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"...", "employee_id":"...", "service_id":"..."}'

curl -X POST http://localhost:8084/api/staff \
  -H "Content-Type: application/json" \
  -d '{"name":"张三", "email":"zhang@example.com", "department":"技术部"}'
```

### 压力测试

使用 JMeter 或 k6 进行压力测试：

```bash
# k6 压力测试示例
k6 run --vus 100 --duration 30s load_test.js
```

---

## 📊 监控与日志

### 健康检查

```bash
# Appointments 健康检查
curl http://localhost:8083/health

# Staff 健康检查
curl http://localhost:8084/health
```

### Prometheus 指标

所有服务都暴露 Prometheus 指标：

```bash
# Appointments 指标
curl http://localhost:8083/metrics

# Staff 指标
curl http://localhost:8084/metrics
```

### 日志管理

结构化日志输出，支持：

- JSON 格式
- 日志级别控制
- 文件轮转
- 链路追踪集成

---

## 🔮 未来规划

### 短期目标 (1-2个月)

1. **完善测试覆盖**: 单元测试覆盖率 > 90%
2. **性能优化**: 关键接口响应时间 < 200ms
3. **监控完善**: 集成 Jaeger 链路追踪
4. **文档完善**: API 文档自动生成

### 中期目标 (3-6个月)

1. **功能增强**:
   - 预约模板功能
   - 员工考勤集成
   - 客户满意度调研
2. **性能提升**:
   - 缓存策略优化
   - 数据库索引优化
   - 并发处理优化
3. **运维工具**:
   - 自动化部署脚本
   - 健康检查仪表板
   - 告警规则配置

### 长期目标 (6-12个月)

1. **智能化**:
   - 智能推荐系统
   - 预约预测分析
   - 员工绩效分析
2. **生态集成**:
   - 第三方日历集成
   - 支付网关集成
   - 消息推送集成
3. **国际化**:
   - 多语言支持
   - 多时区支持
   - 多货币支持

---

## 📞 技术支持

如有技术问题，请联系：

- **GitHub Issues**: [项目地址](https://github.com/julesChu12/fly)
- **技术文档**: [项目Wiki](https://github.com/julesChu12/fly/wiki)
- **API文档**: [Swagger UI](http://localhost:3000/swagger)

---

*最后更新时间: 2024-01-15*