# Cap CRM 后端接口文档

## 基础信息

- **项目名称**: Cap CRM (客户关系管理系统)
- **API版本**: v1.0
- **基础URL**: `http://localhost:3000/api/v1`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON

## 通用规范

### 请求头
```
Content-Type: application/json
Authorization: Bearer {token}
Accept: application/json
```

### 统一响应格式
```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 错误响应格式
```json
{
  "code": 400,
  "message": "参数错误",
  "error": "name is required",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 分页参数
- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 20, 最大: 100)
- `sort`: 排序字段 (默认: createdAt)
- `order`: 排序方式 (asc|desc, 默认: desc)

---

## 🔐 认证模块 (Authentication)

### 用户登录
```http
POST /auth/login
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
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "refresh_token_here",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "张三",
      "role": "admin"
    },
    "expiresIn": 86400
  }
}
```

### 用户登出
```http
POST /auth/logout
Authorization: Bearer {token}
```

### 刷新Token
```http
POST /auth/refresh
```

**请求体:**
```json
{
  "refreshToken": "refresh_token_here"
}
```

### 获取当前用户信息
```http
GET /auth/me
Authorization: Bearer {token}
```

### 注册用户
```http
POST /auth/register
```

**请求体:**
```json
{
  "email": "newuser@example.com",
  "password": "password123",
  "name": "李四",
  "phone": "13800138000",
  "roleId": 2
}
```

---

## 👥 员工管理 (Employees)

### 获取员工列表
```http
GET /employees?page=1&limit=20&search=张三&status=active
Authorization: Bearer {token}
```

### 创建员工
```http
POST /employees
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "name": "王五",
  "email": "wangwu@example.com",
  "phone": "13800138001",
  "roleId": 3,
  "department": "销售部",
  "position": "销售经理",
  "avatar": "https://example.com/avatar.jpg",
  "status": "active"
}
```

### 获取员工详情
```http
GET /employees/:id
Authorization: Bearer {token}
```

### 更新员工信息
```http
PUT /employees/:id
Authorization: Bearer {token}
```

### 删除员工
```http
DELETE /employees/:id
Authorization: Bearer {token}
```

### 更新员工状态
```http
PUT /employees/:id/status
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "status": "inactive"
}
```

### 获取角色列表
```http
GET /employees/roles
Authorization: Bearer {token}
```

---

## 🤝 客户管理 (Customers)

### 获取客户列表
```http
GET /customers?page=1&limit=20&search=张三&tags=VIP&status=active
Authorization: Bearer {token}
```

### 创建客户
```http
POST /customers
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "name": "张先生",
  "email": "zhang@example.com",
  "phone": "13800138000",
  "gender": "male",
  "birthday": "1990-01-01",
  "address": "北京市朝阳区",
  "company": "某公司",
  "position": "经理",
  "source": "推荐",
  "tags": ["VIP", "重要客户"],
  "notes": "重要客户，需要重点维护"
}
```

### 获取客户详情
```http
GET /customers/:id
Authorization: Bearer {token}
```

### 更新客户信息
```http
PUT /customers/:id
Authorization: Bearer {token}
```

### 删除客户
```http
DELETE /customers/:id
Authorization: Bearer {token}
```

### 搜索客户
```http
GET /customers/search?q=张先生&limit=10
Authorization: Bearer {token}
```

### 获取客户标签
```http
GET /customers/tags
Authorization: Bearer {token}
```

### 添加客户标签
```http
POST /customers/:id/tags
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "tags": ["VIP", "潜在客户"]
}
```

### 获取客户联系人
```http
GET /customers/:id/contacts
Authorization: Bearer {token}
```

### 添加客户联系人
```http
POST /customers/:id/contacts
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "name": "李小姐",
  "phone": "13800138001",
  "email": "li@example.com",
  "position": "助理",
  "isPrimary": true
}
```

### 获取客户交互历史
```http
GET /customers/:id/history?page=1&limit=20
Authorization: Bearer {token}
```

### 添加客户备注
```http
POST /customers/:id/notes
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "content": "客户今天来电咨询产品信息",
  "type": "call",
  "nextFollowUp": "2024-01-15T10:00:00Z"
}
```

---

## 📅 预约管理 (Appointments)

### 获取预约列表
```http
GET /appointments?page=1&limit=20&date=2024-01-01&status=confirmed
Authorization: Bearer {token}
```

### 创建预约
```http
POST /appointments
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "customerId": 1,
  "employeeId": 2,
  "serviceId": 3,
  "startTime": "2024-01-15T14:00:00Z",
  "endTime": "2024-01-15T15:00:00Z",
  "notes": "客户需要详细咨询",
  "status": "pending",
  "reminder": true,
  "reminderTime": "2024-01-15T13:00:00Z"
}
```

### 获取预约详情
```http
GET /appointments/:id
Authorization: Bearer {token}
```

### 更新预约
```http
PUT /appointments/:id
Authorization: Bearer {token}
```

### 取消预约
```http
DELETE /appointments/:id
Authorization: Bearer {token}
```

### 更新预约状态
```http
PUT /appointments/:id/status
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "status": "completed",
  "completionNotes": "服务已完成，客户满意"
}
```

### 获取日历视图
```http
GET /appointments/calendar?start=2024-01-01&end=2024-01-31&employeeId=2
Authorization: Bearer {token}
```

### 检查可用时间
```http
GET /appointments/availability?date=2024-01-15&employeeId=2&serviceId=3
Authorization: Bearer {token}
```

---

## 💳 支付管理 (Payments)

### 获取支付记录
```http
GET /payments?page=1&limit=20&customerId=1&status=paid&startDate=2024-01-01&endDate=2024-01-31
Authorization: Bearer {token}
```

### 创建支付记录
```http
POST /payments
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "customerId": 1,
  "appointmentId": 5,
  "amount": 299.00,
  "method": "wechat",
  "type": "service",
  "description": "理发服务支付",
  "status": "pending",
  "dueDate": "2024-01-15T18:00:00Z"
}
```

### 获取支付详情
```http
GET /payments/:id
Authorization: Bearer {token}
```

### 更新支付记录
```http
PUT /payments/:id
Authorization: Bearer {token}
```

### 删除支付记录
```http
DELETE /payments/:id
Authorization: Bearer {token}
```

### 申请退款
```http
POST /payments/:id/refund
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "reason": "客户不满意服务",
  "refundAmount": 299.00
}
```

### 获取支付统计
```http
GET /payments/statistics?startDate=2024-01-01&endDate=2024-01-31
Authorization: Bearer {token}
```

### 获取支付方式列表
```http
GET /payments/methods
Authorization: Bearer {token}
```

**响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "微信支付",
      "code": "wechat",
      "enabled": true
    },
    {
      "id": 2,
      "name": "支付宝",
      "code": "alipay",
      "enabled": true
    },
    {
      "id": 3,
      "name": "现金",
      "code": "cash",
      "enabled": true
    }
  ]
}
```

---

## 🛍️ 产品服务管理 (Products & Services)

### 服务管理

#### 获取服务列表
```http
GET /services?page=1&limit=20&category=hair&status=active
Authorization: Bearer {token}
```

#### 创建服务
```http
POST /services
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "name": "洗剪吹套餐",
  "description": "包含洗发、剪发、吹干服务",
  "price": 68.00,
  "duration": 60,
  "category": "hair",
  "image": "https://example.com/service.jpg",
  "status": "active",
  "sortOrder": 1
}
```

#### 获取服务详情
```http
GET /services/:id
Authorization: Bearer {token}
```

#### 更新服务
```http
PUT /services/:id
Authorization: Bearer {token}
```

#### 删除服务
```http
DELETE /services/:id
Authorization: Bearer {token}
```

#### 更新服务状态
```http
PUT /services/:id/status
Authorization: Bearer {token}
```

### 产品管理

#### 获取产品列表
```http
GET /products?page=1&limit=20&category=shampoo&status=active&stock=true
Authorization: Bearer {token}
```

#### 创建产品
```http
POST /products
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "name": "洗发水",
  "description": "专业洗发水，适合各种发质",
  "price": 45.00,
  "cost": 25.00,
  "stock": 50,
  "minStock": 10,
  "categoryId": 1,
  "sku": "SHAMPOO-001",
  "image": "https://example.com/product.jpg",
  "status": "active"
}
```

#### 更新库存
```http
PUT /products/:id/stock
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "stock": 45,
  "operation": "subtract",
  "reason": "销售出库"
}
```

#### 获取产品分类
```http
GET /products/categories
Authorization: Bearer {token}
```

---

## 📊 数据统计 (Analytics)

### 仪表板数据
```http
GET /dashboard/overview
Authorization: Bearer {token}
```

**响应:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "totalCustomers": 1250,
    "totalRevenue": 125000.50,
    "todayAppointments": 15,
    "pendingPayments": 2500.00,
    "newCustomersThisMonth": 45,
    "revenueGrowth": 12.5
  }
}
```

### 收入统计
```http
GET /dashboard/revenue?startDate=2024-01-01&endDate=2024-01-31&groupBy=day
Authorization: Bearer {token}
```

### 客户统计
```http
GET /dashboard/customers?startDate=2024-01-01&endDate=2024-01-31
Authorization: Bearer {token}
```

### 预约统计
```http
GET /dashboard/appointments?startDate=2024-01-01&endDate=2024-01-31
Authorization: Bearer {token}
```

### 业绩统计
```http
GET /dashboard/performance?startDate=2024-01-01&endDate=2024-01-31&employeeId=2
Authorization: Bearer {token}
```

---

## 📱 通用接口 (Common)

### 文件上传

#### 上传头像
```http
POST /upload/avatar
Content-Type: multipart/form-data
Authorization: Bearer {token}
```

#### 上传文档
```http
POST /upload/document
Content-Type: multipart/form-data
Authorization: Bearer {token}
```

#### 上传图片
```http
POST /upload/image
Content-Type: multipart/form-data
Authorization: Bearer {token}
```

### 系统配置

#### 获取系统设置
```http
GET /settings
Authorization: Bearer {token}
```

#### 更新系统设置
```http
PUT /settings
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "shopName": "美发沙龙",
  "phone": "400-123-4567",
  "email": "contact@example.com",
  "address": "北京市朝阳区某某街道123号",
  "businessHours": {
    "monday": "09:00-20:00",
    "tuesday": "09:00-20:00",
    "wednesday": "09:00-20:00",
    "thursday": "09:00-20:00",
    "friday": "09:00-20:00",
    "saturday": "09:00-21:00",
    "sunday": "10:00-19:00"
  }
}
```

### 消息通知

#### 获取通知列表
```http
GET /notifications?page=1&limit=20&unread=true
Authorization: Bearer {token}
```

#### 标记已读
```http
PUT /notifications/:id/read
Authorization: Bearer {token}
```

#### 发送通知
```http
POST /notifications/send
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "userId": 2,
  "title": "预约提醒",
  "content": "您有一个预约将在30分钟后开始",
  "type": "appointment_reminder"
}
```

---

## 🔍 搜索过滤 (Search & Filter)

### 通用搜索
```http
GET /search/customers?q=张先生&limit=10
GET /search/appointments?q=2024-01-15&limit=10
GET /search/products?q=洗发水&limit=10
Authorization: Bearer {token}
```

### 高级过滤
```http
POST /filter/customers
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "filters": {
    "tags": ["VIP"],
    "totalSpent": { "min": 1000, "max": 5000 },
    "lastVisitDate": { "start": "2024-01-01", "end": "2024-01-31" },
    "status": "active"
  },
  "sort": { "field": "lastVisitDate", "order": "desc" },
  "page": 1,
  "limit": 20
}
```

---

## 🔗 关联数据 (Related Data)

### 获取关联信息
```http
GET /customers/:id/appointments?page=1&limit=20
GET /customers/:id/payments?page=1&limit=20
GET /employees/:id/appointments?page=1&limit=20
GET /services/:id/appointments?page=1&limit=20
Authorization: Bearer {token}
```

---

## 📡 实时功能 (Real-time)

### WebSocket 连接
```
WS /realtime/connect?token={token}
```

### 订阅频道
```json
{
  "type": "subscribe",
  "channels": ["notifications", "appointments"]
}
```

### 实时通知格式
```json
{
  "type": "notification",
  "channel": "notifications",
  "data": {
    "id": 1,
    "title": "新预约",
    "message": "客户张先生预约了明天14:00的服务",
    "timestamp": "2024-01-15T10:00:00Z"
  }
}
```

---

## 🛡️ 权限管理 (Permissions)

### 获取权限列表
```http
GET /permissions
Authorization: Bearer {token}
```

### 获取角色列表
```http
GET /roles
Authorization: Bearer {token}
```

### 创建角色
```http
POST /roles
Authorization: Bearer {token}
```

**请求体:**
```json
{
  "name": "发型师",
  "description": "负责提供发型设计服务",
  "permissions": [
    "customers:read",
    "appointments:read",
    "appointments:update",
    "services:read"
  ]
}
```

### 获取用户权限
```http
GET /users/:id/permissions
Authorization: Bearer {token}
```

---

## 📝 数据导入导出 (Import/Export)

### 导入客户数据
```http
POST /import/customers
Content-Type: multipart/form-data
Authorization: Bearer {token}
```

### 导出数据
```http
GET /export/customers?format=csv&startDate=2024-01-01&endDate=2024-01-31
GET /export/appointments?format=excel&month=2024-01
GET /export/payments?format=pdf&quarter=2024-Q1
Authorization: Bearer {token}
```

---

## 状态码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权，需要登录 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 422 | 数据验证失败 |
| 500 | 服务器内部错误 |

## 数据字典

### 预约状态
- `pending`: 待确认
- `confirmed`: 已确认
- `in_progress`: 进行中
- `completed`: 已完成
- `cancelled`: 已取消
- `no_show`: 未到店

### 支付状态
- `pending`: 待支付
- `paid`: 已支付
- `refunded`: 已退款
- `partial_refund`: 部分退款
- `cancelled`: 已取消

### 用户状态
- `active`: 正常
- `inactive`: 停用
- `locked`: 锁定
- `pending`: 待激活

### 支付方式
- `wechat`: 微信支付
- `alipay`: 支付宝
- `cash`: 现金
- `card`: 银行卡
- `other`: 其他

---

## 更新日志

### v1.0.0 (2024-01-15)
- 初始版本发布
- 包含所有核心功能接口
- 支持JWT认证
- 提供完整的CRUD操作