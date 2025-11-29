# Cap CRM 后端 API 接口需求文档

## 📋 项目概述

基于前端 API 服务层的实现，需要开发完整的 RESTful API 后端服务来支持 Cap CRM 系统的所有业务功能。

---

## 🔐 认证模块 (Authentication Module)

### 基础配置
- **Base URL**: `/api/v1/auth`
- **认证方式**: JWT (JSON Web Token)

### 接口列表

#### 1. 用户登录
```http
POST /api/v1/auth/login
```
**请求体:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```
**响应:**
```json
{
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "张三",
      "role": "admin",
      "avatar": "https://example.com/avatar.jpg"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 2. 刷新 Token
```http
POST /api/v1/auth/refresh
```
**请求体:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### 3. 注册用户
```http
POST /api/v1/auth/register
```
**请求体:**
```json
{
  "email": "newuser@example.com",
  "password": "password123",
  "name": "新用户",
  "role": "user"
}
```

#### 4. 忘记密码
```http
POST /api/v1/auth/forgot-password
```
**请求体:**
```json
{
  "email": "user@example.com"
}
```

#### 5. 重置密码
```http
POST /api/v1/auth/reset-password
```
**请求体:**
```json
{
  "token": "reset_token_here",
  "newPassword": "newpassword123"
}
```

#### 6. 获取当前用户信息
```http
GET /api/v1/auth/me
```
**Headers:** `Authorization: Bearer {token}`

---

## 👥 客户管理模块 (Customer Management)

### 基础配置
- **Base URL**: `/api/v1/customers`

### 数据模型
```typescript
interface Customer {
  id: number;
  name: string;
  email: string;
  phone: string;
  gender: 'male' | 'female' | 'other';
  birthday: string; // ISO 8601
  address: string;
  company?: string;
  position?: string;
  source?: string;
  tags: string[];
  notes?: string;
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
}
```

### 接口列表

#### 1. 获取客户列表
```http
GET /api/v1/customers
```
**查询参数:**
- `page`: number (默认: 1)
- `limit`: number (默认: 10)
- `search`: string (搜索关键词)
- `status`: 'active' | 'inactive'
- `source`: string
- `tags`: string[]

#### 2. 获取单个客户
```http
GET /api/v1/customers/{id}
```

#### 3. 创建客户
```http
POST /api/v1/customers
```
**请求体:**
```json
{
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "gender": "male",
  "birthday": "1990-01-01",
  "address": "北京市朝阳区",
  "company": "ABC公司",
  "position": "经理",
  "source": "网络推广",
  "tags": ["VIP", "老客户"],
  "notes": "重要客户",
  "status": "active"
}
```

#### 4. 更新客户
```http
PUT /api/v1/customers/{id}
```

#### 5. 删除客户
```http
DELETE /api/v1/customers/{id}
```

#### 6. 批量删除客户
```http
DELETE /api/v1/customers/batch
```
**请求体:**
```json
{
  "ids": [1, 2, 3]
}
```

#### 7. 添加客户标签
```http
POST /api/v1/customers/{id}/tags
```
**请求体:**
```json
{
  "tags": ["VIP", "重要客户"]
}
```

#### 8. 移除客户标签
```http
DELETE /api/v1/customers/{id}/tags
```
**请求体:**
```json
{
  "tags": ["VIP"]
}
```

#### 9. 获取客户统计
```http
GET /api/v1/customers/statistics
```

---

## 📅 预约管理模块 (Appointment Management)

### 基础配置
- **Base URL**: `/api/v1/appointments`

### 数据模型
```typescript
interface Appointment {
  id: number;
  customerId: number;
  employeeId: number;
  serviceId: number;
  startTime: string; // ISO 8601
  endTime: string; // ISO 8601
  status: 'pending' | 'confirmed' | 'completed' | 'cancelled';
  notes?: string;
  createdAt: string;
  updatedAt: string;
  customer?: Customer;
  employee?: Employee;
  service?: Service;
}
```

### 接口列表

#### 1. 获取预约列表
```http
GET /api/v1/appointments
```
**查询参数:**
- `page`, `limit`
- `customerId`: number
- `employeeId`: number
- `serviceId`: number
- `status`: string
- `startDate`: string
- `endDate`: string

#### 2. 获取单个预约
```http
GET /api/v1/appointments/{id}
```

#### 3. 创建预约
```http
POST /api/v1/appointments
```
**请求体:**
```json
{
  "customerId": 1,
  "employeeId": 2,
  "serviceId": 3,
  "startTime": "2025-01-15T10:00:00Z",
  "endTime": "2025-01-15T11:00:00Z",
  "notes": "首次预约",
  "status": "pending"
}
```

#### 4. 更新预约
```http
PUT /api/v1/appointments/{id}
```

#### 5. 删除预约
```http
DELETE /api/v1/appointments/{id}
```

#### 6. 批量删除预约
```http
DELETE /api/v1/appointments/batch
```

#### 7. 获取可用时间槽
```http
GET /api/v1/appointments/available-slots
```
**查询参数:**
- `employeeId`: number
- `serviceId`: number
- `date`: string

#### 8. 获取日历事件
```http
GET /api/v1/appointments/calendar
```
**查询参数:**
- `startDate`: string
- `endDate`: string
- `employeeId`?: number

---

## 👨‍💼 员工管理模块 (Employee Management)

### 基础配置
- **Base URL**: `/api/v1/employees`

### 数据模型
```typescript
interface Employee {
  id: number;
  name: string;
  email: string;
  phone: string;
  position: string;
  department: string;
  hireDate: string; // ISO 8601
  status: 'active' | 'inactive';
  address: string;
  avatarUrl?: string;
  roleId?: number;
  createdAt: string;
  updatedAt: string;
}
```

### 接口列表

#### 1. 获取员工列表
```http
GET /api/v1/employees
```
**查询参数:**
- `page`, `limit`
- `department`: string
- `position`: string
- `status`: string
- `search`: string

#### 2. 获取单个员工
```http
GET /api/v1/employees/{id}
```

#### 3. 创建员工
```http
POST /api/v1/employees
```
**请求体:**
```json
{
  "name": "李四",
  "email": "lisi@example.com",
  "phone": "13900139000",
  "position": "发型师",
  "department": "技术部",
  "hireDate": "2024-01-01",
  "status": "active",
  "address": "北京市海淀区",
  "avatarUrl": "https://example.com/avatar.jpg",
  "roleId": 2
}
```

#### 4. 更新员工
```http
PUT /api/v1/employees/{id}
```

#### 5. 删除员工
```http
DELETE /api/v1/employees/{id}
```

#### 6. 批量删除员工
```http
DELETE /api/v1/employees/batch
```

#### 7. 获取员工统计
```http
GET /api/v1/employees/statistics
```

---

## 💇 服务管理模块 (Service Management)

### 基础配置
- **Base URL**: `/api/v1/services`

### 数据模型
```typescript
interface Service {
  id: number;
  name: string;
  description?: string;
  price: number;
  duration: number; // 分钟
  category?: string;
  isActive: boolean;
  imageUrl?: string;
  createdAt: string;
  updatedAt: string;
}
```

### 接口列表

#### 1. 获取服务列表
```http
GET /api/v1/services
```
**查询参数:**
- `page`, `limit`
- `category`: string
- `isActive`: boolean
- `search`: string
- `minPrice`: number
- `maxPrice`: number

#### 2. 获取单个服务
```http
GET /api/v1/services/{id}
```

#### 3. 创建服务
```http
POST /api/v1/services
```
**请求体:**
```json
{
  "name": "洗剪吹套餐",
  "description": "包含洗头、剪发、吹干造型",
  "price": 128,
  "duration": 60,
  "category": "基础服务",
  "isActive": true,
  "imageUrl": "https://example.com/service.jpg"
}
```

#### 4. 更新服务
```http
PUT /api/v1/services/{id}
```

#### 5. 删除服务
```http
DELETE /api/v1/services/{id}
```

#### 6. 批量删除服务
```http
DELETE /api/v1/services/batch
```

#### 7. 获取服务分类
```http
GET /api/v1/services/categories
```

#### 8. 获取服务统计
```http
GET /api/v1/services/statistics
```

---

## 📦 产品管理模块 (Product Management)

### 基础配置
- **Base URL**: `/api/v1/products`

### 数据模型
```typescript
interface Product {
  id: number;
  name: string;
  description?: string;
  price: number;
  stock: number;
  category?: string;
  brand?: string;
  sku?: string;
  barcode?: string;
  imageUrl?: string;
  isActive: boolean;
  cost?: number;
  createdAt: string;
  updatedAt: string;
}
```

### 接口列表

#### 1. 获取产品列表
```http
GET /api/v1/products
```
**查询参数:**
- `page`, `limit`
- `category`: string
- `brand`: string
- `isActive`: boolean
- `search`: string
- `lowStock`: boolean (库存不足)

#### 2. 获取单个产品
```http
GET /api/v1/products/{id}
```

#### 3. 创建产品
```http
POST /api/v1/products
```
**请求体:**
```json
{
  "name": "洗发水",
  "description": "专业洗发水，适合各种发质",
  "price": 68,
  "stock": 100,
  "category": "洗护用品",
  "brand": "欧莱雅",
  "sku": "PROD-001",
  "barcode": "1234567890123",
  "imageUrl": "https://example.com/product.jpg",
  "isActive": true,
  "cost": 35
}
```

#### 4. 更新产品
```http
PUT /api/v1/products/{id}
```

#### 5. 删除产品
```http
DELETE /api/v1/products/{id}
```

#### 6. 批量删除产品
```http
DELETE /api/v1/products/batch
```

#### 7. 更新产品库存
```http
PATCH /api/v1/products/{id}/stock
```
**请求体:**
```json
{
  "stock": 150
}
```

#### 8. 获取产品分类
```http
GET /api/v1/products/categories
```

#### 9. 获取产品品牌
```http
GET /api/v1/products/brands
```

#### 10. 获取产品统计
```http
GET /api/v1/products/statistics
```

---

## 💰 支付管理模块 (Payment Management)

### 基础配置
- **Base URL**: `/api/v1/payments`

### 数据模型
```typescript
interface Payment {
  id: number;
  customerId: number;
  appointmentId?: number;
  employeeId?: number;
  amount: number;
  method: 'cash' | 'card' | 'mobile' | 'transfer';
  status: 'pending' | 'completed' | 'failed' | 'refunded';
  description?: string;
  transactionId?: string;
  paymentDate: string; // ISO 8601
  createdAt: string;
  updatedAt: string;
}
```

### 接口列表

#### 1. 获取支付记录
```http
GET /api/v1/payments
```
**查询参数:**
- `page`, `limit`
- `customerId`: number
- `appointmentId`: number
- `employeeId`: number
- `method`: string
- `status`: string
- `startDate`: string
- `endDate`: string

#### 2. 获取单个支付
```http
GET /api/v1/payments/{id}
```

#### 3. 创建支付
```http
POST /api/v1/payments
```
**请求体:**
```json
{
  "customerId": 1,
  "appointmentId": 10,
  "employeeId": 2,
  "amount": 588.00,
  "method": "card",
  "description": "洗剪吹服务费用",
  "transactionId": "TXN123456789",
  "paymentDate": "2025-01-15T14:30:00Z"
}
```

#### 4. 更新支付
```http
PUT /api/v1/payments/{id}
```

#### 5. 删除支付
```http
DELETE /api/v1/payments/{id}
```

#### 6. 退款
```http
POST /api/v1/payments/{id}/refund
```
**请求体:**
```json
{
  "reason": "客户申请退款"
}
```

#### 7. 获取支付统计
```http
GET /api/v1/payments/statistics
```
**查询参数:**
- `startDate`: string
- `endDate`: string

---

## 🔧 通用接口

### 1. 健康检查
```http
GET /api/v1/health
```
**响应:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-15T10:00:00Z",
  "version": "1.0.0",
  "database": "connected"
}
```

### 2. API 版本信息
```http
GET /api/v1/version
```
**响应:**
```json
{
  "version": "1.0.0",
  "apiVersion": "v1",
  "buildDate": "2025-01-15T10:00:00Z"
}
```

### 3. 错误响应格式
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": [
      {
        "field": "email",
        "message": "邮箱格式不正确"
      }
    ]
  }
}
```

---

## 📊 分页响应格式

所有列表接口都应返回标准化的分页响应：

```json
{
  "data": {
    "items": [
      // 实体对象数组
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "totalPages": 10,
      "hasNext": true,
      "hasPrev": false
    }
  }
}
```

---

## 🔐 安全要求

### 1. 认证
- 所有需要认证的接口都应在请求头中包含 JWT Token
- `Authorization: Bearer {token}`

### 2. 权限控制
- 基于角色的访问控制 (RBAC)
- 常见角色：admin, manager, employee, user

### 3. 输入验证
- 所有请求数据都应进行严格的验证
- 防止 SQL 注入、XSS 等安全漏洞

### 4. 速率限制
- API 调用频率限制
- 防止暴力攻击

---

## 🔄 WebSocket 实时功能

### 连接地址
`ws://localhost:3001` (开发环境)

### 事件类型

#### 1. 客户事件
- `customer:created` - 新客户创建
- `customer:updated` - 客户信息更新
- `customer:deleted` - 客户删除

#### 2. 预约事件
- `appointment:created` - 新预约创建
- `appointment:updated` - 预约状态更新
- `appointment:deleted` - 预约删除

#### 3. 支付事件
- `payment:created` - 新支付创建
- `payment:updated` - 支付状态更新
- `payment:refunded` - 支付退款

---

## 📝 数据库要求

### 必需表结构

1. **users** - 用户表
2. **customers** - 客户表
3. **appointments** - 预约表
4. **employees** - 员工表
5. **services** - 服务表
6. **products** - 产品表
7. **payments** - 支付表
8. **roles** - 角色表
9. **user_roles** - 用户角色关联表

### 字段要求
- 所有表都应包含 `id` (主键)、`created_at`、`updated_at` 字段
- 使用合适的索引优化查询性能
- 设置适当的外键约束

---

## 🚀 部署要求

### 环境变量
```bash
# 数据库配置
DATABASE_URL=
DB_HOST=
DB_PORT=
DB_NAME=
DB_USER=
DB_PASSWORD=

# JWT 配置
JWT_SECRET=
JWT_EXPIRES_IN=
JWT_REFRESH_EXPIRES_IN=

# 服务器配置
PORT=3000
NODE_ENV=production
CORS_ORIGIN=
```

### 技术栈建议
- **后端**: Node.js + Express/Koa/Fastify
- **数据库**: PostgreSQL/MySQL
- **ORM**: Prisma/TypeORM/Sequelize
- **认证**: jsonwebtoken + bcrypt
- **WebSocket**: Socket.io
- **文档**: Swagger/OpenAPI

---

## 📋 开发优先级

### Phase 1: 核心功能 (必需)
1. 认证系统 (登录/注册/token刷新)
2. 客户管理 CRUD
3. 预约管理 CRUD
4. 基础权限控制

### Phase 2: 业务功能 (重要)
1. 员工管理
2. 服务管理
3. 产品管理
4. 支付管理

### Phase 3: 增强功能 (可选)
1. WebSocket 实时更新
2. 统计报表
3. 文件上传
4. 邮件通知

### Phase 4: 高级功能 (长期)
1. 高级权限管理
2. 审计日志
3. 数据导入导出
4. API 限流和监控

---

## 📞 技术支持

如需进一步的技术咨询或有任何实现问题，请参考前端代码中的 API 服务实现：
- `src/services/api/` - API 客户端实现
- `src/components/` - 前端组件实现
- `CLAUDE.md` - 项目架构文档

**后端 API 开发完成后，Cap CRM 系统将完全就绪！** 🎉