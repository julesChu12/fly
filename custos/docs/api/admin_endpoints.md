# Admin API Endpoints

本文档描述了 Custos 系统中的管理员（Admin）API 端点。所有端点都需要管理员权限。

## 认证要求

所有 Admin API 都需要：
1. 有效的 JWT Access Token（通过 `Authorization: Bearer <token>` header）
2. 用户必须拥有 `admin` 角色

## 端点列表

### 1. 用户管理

#### 1.1 列出用户（带分页和过滤）

```http
GET /api/v1/admin/users
```

**查询参数：**
- `page` (int, optional): 页码，默认 1
- `page_size` (int, optional): 每页数量，默认 20，最大 100
- `status` (string, optional): 按状态过滤 (active|inactive|frozen|disabled|locked|deleted|merged)
- `role` (string, optional): 按角色过滤 (admin|user|guest)
- `user_type` (string, optional): 按用户类型过滤 (customer|staff|partner)
- `keyword` (string, optional): 搜索关键词（匹配用户名或邮箱）
- `tenant_id` (uint, optional): 租户ID过滤

**响应示例：**
```json
{
  "users": [
    {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "nickname": "Administrator",
      "avatar": "https://example.com/avatar.jpg",
      "status": "active",
      "role": "admin",
      "user_type": "staff",
      "tenant_id": null,
      "token_version": 0,
      "merged_into_user_id": null,
      "last_login_at": "2025-01-27T10:00:00Z",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-27T10:00:00Z"
    }
  ],
  "total": 150,
  "page": 1,
  "page_size": 20,
  "total_pages": 8
}
```

#### 1.2 获取单个用户详情

```http
GET /api/v1/admin/users/:id
```

**路径参数：**
- `id` (uint): 用户ID

**响应示例：**
```json
{
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "status": "active",
    "role": "admin",
    "token_version": 0,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-27T10:00:00Z"
  },
  "roles": ["admin"],
  "active_sessions": 2
}
```

#### 1.3 更新用户状态

```http
PATCH /api/v1/admin/users/:id/status
```

**请求体：**
```json
{
  "status": "disabled",
  "reason": "Violation of terms of service"
}
```

**支持的状态值：**
- `active` - 激活
- `inactive` - 未激活
- `frozen` - 冻结
- `disabled` - 禁用
- `locked` - 锁定
- `deleted` - 删除

**响应示例：**
```json
{
  "user_id": 123,
  "old_status": "active",
  "new_status": "disabled",
  "updated_at": "2025-01-27T10:30:00Z",
  "message": "user status updated successfully"
}
```

**注意：** 当状态更新为 `disabled`、`locked` 或 `deleted` 时，系统会自动：
- 撤销该用户的所有活跃会话
- 递增 `token_version` 以使所有现有 token 失效

#### 1.4 强制用户下线

```http
POST /api/v1/admin/users/:id/force-logout
```

**请求体（可选）：**
```json
{
  "session_id": "uuid-session-id",
  "reason": "Security incident"
}
```

- 如果提供 `session_id`：仅撤销指定会话
- 如果不提供 `session_id`：撤销用户的所有会话并递增 token_version

**响应示例：**
```json
{
  "user_id": 123,
  "sessions_revoked": 3,
  "token_version_old": 0,
  "token_version_new": 1,
  "message": "user forcefully logged out"
}
```

---

### 2. 角色与权限管理

#### 2.1 分配角色给用户

```http
POST /api/v1/admin/users/:id/roles
```

**请求体：**
```json
{
  "role": "admin"
}
```

**支持的角色：**
- `admin` - 管理员
- `user` - 普通用户
- `guest` - 访客

**响应示例：**
```json
{
  "message": "role assigned successfully"
}
```

#### 2.2 获取用户的角色

```http
GET /api/v1/admin/users/:id/roles
```

**响应示例：**
```json
{
  "user_id": 123,
  "roles": ["admin", "user"],
  "permissions": [
    "users:read",
    "users:write",
    "admin:access"
  ]
}
```

#### 2.3 添加策略规则

```http
POST /api/v1/admin/policies
```

**请求体：**
```json
{
  "subject": "admin",
  "object": "/api/v1/admin/*",
  "action": "GET"
}
```

**响应示例：**
```json
{
  "message": "policy added successfully"
}
```

#### 2.4 移除策略规则

```http
DELETE /api/v1/admin/policies
```

**请求体：**
```json
{
  "subject": "admin",
  "object": "/api/v1/admin/*",
  "action": "GET"
}
```

**响应示例：**
```json
{
  "message": "policy removed successfully"
}
```

---

### 3. 系统统计

#### 3.1 获取系统统计信息

```http
GET /api/v1/admin/stats
```

**响应示例：**
```json
{
  "total_users": 1500,
  "active_users": 1200,
  "inactive_users": 200,
  "frozen_users": 50,
  "deleted_users": 50,
  "total_sessions": 2500,
  "active_sessions": 1800,
  "users_by_role": {
    "admin": 10,
    "user": 1400,
    "guest": 90
  },
  "users_by_type": {
    "customer": 1300,
    "staff": 150,
    "partner": 50
  },
  "new_users_today": 25,
  "new_users_this_week": 180
}
```

---

## 错误响应

所有端点在出错时返回标准错误格式：

```json
{
  "error": "error message description"
}
```

**常见 HTTP 状态码：**
- `200 OK` - 请求成功
- `400 Bad Request` - 请求参数错误
- `401 Unauthorized` - 未认证
- `403 Forbidden` - 权限不足
- `404 Not Found` - 资源不存在
- `500 Internal Server Error` - 服务器内部错误

---

## 使用示例

### 示例 1: 列出所有激活的管理员

```bash
curl -X GET "http://localhost:8081/api/v1/admin/users?status=active&role=admin&page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 示例 2: 禁用用户账户

```bash
curl -X PATCH "http://localhost:8081/api/v1/admin/users/123/status" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "disabled",
    "reason": "Suspicious activity detected"
  }'
```

### 示例 3: 强制用户全部下线

```bash
curl -X POST "http://localhost:8081/api/v1/admin/users/123/force-logout" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

### 示例 4: 获取系统统计

```bash
curl -X GET "http://localhost:8081/api/v1/admin/stats" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 安全注意事项

1. **永远不要在日志中记录敏感信息**（如 token、密码等）
2. **强制下线功能慎用**，会影响用户体验
3. **定期审计管理员操作**，确保没有滥用权限
4. **使用 HTTPS** 传输所有敏感数据
5. **实施 IP 白名单**（可选）限制管理员访问来源

---

## 数据模型

### UserListFilter

用于 ListUsers 的过滤器：

```go
type UserListFilter struct {
    Status   *string  // 用户状态
    Role     *string  // 用户角色
    UserType *string  // 用户类型
    TenantID *uint    // 租户ID
    Keyword  *string  // 搜索关键词
}
```

### AdminUserInfo

管理员视图的用户信息：

```go
type AdminUserInfo struct {
    ID               uint
    Username         string
    Email            string
    Nickname         string
    Avatar           string
    Status           string
    Role             string
    UserType         string
    TenantID         *uint
    TokenVersion     int
    MergedIntoUserID *uint
    LastLoginAt      *time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

---

## 版本历史

- **v1.0.0** (2025-01-27): 初始版本，包含完整的用户管理、角色管理和系统统计功能
