# Appointments Service - 预约管理服务

## 概述

Appointments Service 是 Fly CRM 系统的预约管理微服务，负责处理客户预约的创建、查询、状态更新和日历视图等功能。

## 功能特性

- **预约管理**: 创建、更新、删除预约
- **状态管理**: 预约状态流转（待确认→已确认→进行中→已完成）
- **冲突检测**: 自动检查时间冲突
- **可用时间**: 查询员工可用时间段
- **日历视图**: 提供日历格式的预约视图
- **提醒系统**: 支持预约提醒功能
- **统计分析**: 预约数据统计和分析

## 技术架构

- **HTTP端口**: 8083
- **gRPC端口**: 9083
- **数据库**: MySQL (默认: appointments)
- **框架**: Gin (HTTP), gRPC
- **ORM**: GORM
- **日志**: 结构化日志

## 项目结构

```
appointments/
├── cmd/appointments/          # 应用入口
│   ├── main.go               # 主程序
│   └── cmd/                  # Cobra命令
├── internal/
│   ├── application/service/  # 应用服务层
│   ├── domain/               # 领域层
│   │   ├── entity/          # 实体
│   │   ├── repository/      # 仓储接口
│   │   └── dto/             # 数据传输对象
│   ├── interface/           # 接口层
│   │   ├── http/           # HTTP接口
│   │   └── grpc/           # gRPC接口
│   └── infrastructure/     # 基础设施层
│       └── database/       # 数据库实现
├── configs/                 # 配置文件
├── docs/                   # API文档
├── proto/                  # Protocol Buffers
└── Makefile               # 构建脚本
```

## 快速开始

### 1. 安装依赖

```bash
make deps
```

### 2. 配置数据库

编辑 `configs/appointments.yaml` 文件，配置数据库连接信息：

```yaml
database:
  host: localhost
  port: "3306"
  username: your_username
  password: your_password
  database: appointments
```

### 3. 启动服务

```bash
# 开发模式
make dev

# 或者直接运行
make run
```

### 4. 访问服务

- **HTTP API**: http://localhost:8083
- **Swagger文档**: http://localhost:8083/swagger/index.html
- **健康检查**: http://localhost:8083/health

## API接口

### 预约管理

- `GET /api/appointments` - 获取预约列表
- `POST /api/appointments` - 创建预约
- `GET /api/appointments/{id}` - 获取预约详情
- `PUT /api/appointments/{id}` - 更新预约
- `DELETE /api/appointments/{id}` - 删除预约
- `PUT /api/appointments/{id}/status` - 更新预约状态

### 日历和可用时间

- `GET /api/appointments/calendar` - 获取日历视图
- `GET /api/appointments/availability` - 检查可用时间
- `POST /api/appointments/conflict-check` - 检查时间冲突

### 客户和员工相关

- `GET /api/appointments/customer/{customerId}` - 获取客户预约
- `GET /api/appointments/employee/{employeeId}` - 获取员工预约
- `GET /api/appointments/employee/{employeeId}/upcoming` - 获取即将到来的预约

## 预约状态

- `pending` - 待确认
- `confirmed` - 已确认
- `in_progress` - 进行中
- `completed` - 已完成
- `cancelled` - 已取消
- `no_show` - 未到店

## 开发命令

```bash
# 构建
make build

# 运行测试
make test

# 生成Swagger文档
make swagger

# 清理构建文件
make clean

# Docker构建
make docker-build
```

## 环境变量

可以通过环境变量覆盖配置文件设置：

```bash
# 服务器端口
export APPOINTMENTS_SERVER_HTTP_PORT=8083
export APPOINTMENTS_SERVER_GRPC_PORT=9083

# 数据库配置
export APPOINTMENTS_DATABASE_HOST=localhost
export APPOINTMENTS_DATABASE_PASSWORD=your_password

# 日志级别
export APPOINTMENTS_LOGGER_LEVEL=debug
```

## 命令行参数

```bash
# 启动服务
./appointments serve \
  --config configs/appointments.yaml \
  --http-port 8083 \
  --grpc-port 9083 \
  --db-dsn "user:password@tcp(localhost:3306)/appointments"

# 查看版本
./appointments version

# 查看帮助
./appointments help
```

## 数据库表结构

主要表：`appointments`

字段说明：
- `id`: 预约ID (UUID)
- `customer_id`: 客户ID (UUID)
- `employee_id`: 员工ID (UUID)
- `service_id`: 服务ID (UUID)
- `start_time`: 开始时间
- `end_time`: 结束时间
- `status`: 预约状态
- `notes`: 备注
- `reminder`: 是否提醒
- `reminder_time`: 提醒时间
- `created_at`: 创建时间
- `updated_at`: 更新时间
- `deleted_at`: 删除时间（软删除）

## 监控和日志

服务提供以下监控端点：

- `/health` - 健康检查
- 结构化日志输出
- 追踪ID支持

## 测试

运行单元测试：

```bash
make test
```

运行特定测试：

```bash
go test ./internal/application/service/
go test ./internal/infrastructure/database/
```

## 部署

### Docker部署

```bash
# 构���镜像
make docker-build

# 运行容器
docker run -p 8083:8083 -p 9083:9083 appointments:latest
```

### 配置管理

生产环境中建议使用以下配置：

- 日志级别：info 或 warn
- 数据库连接池：根据负载调整
- 启用文件日志记录
- 配置日志轮转

## 故障排除

常见问题：

1. **数据库连接失败**
   - 检查数据库配置
   - 确认数据库服务运行状态
   - 检查网络连接

2. **端口冲突**
   - 修改配置文件中的端口设置
   - 检查端口占用情况

3. **权限问题**
   - 确认数据库用户权限
   - 检查文件系统权限

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交变更
4. 创建 Pull Request

## 许可证

Apache 2.0 License