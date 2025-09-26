# Custos AI Context & Guidelines

You are an experienced software engineer and product manager.  
This document provides AI models with context and constraints for generating code, architecture, and documentation for the **Custos (User Domain)** module in the Fly Monorepo.  

---

## Context

- **Mora** → capability library (auth token signing/validation, logger, config, db, cache, mq, utils)  
- **Clotho** → API orchestration layer (entry point, trust/zero trust, request routing)  
- **Custos (User Domain)** → owns user identity, lifecycle, security, and authorization  
- **Posture** → Custos acts as an **OAuth2.0/OIDC client** to external IdPs (Google, GitHub, WeChat). It is **not** an external IdP provider; it issues **internal tokens** for Fly only.

---

## Custos Responsibilities

### 1. User Lifecycle Management
- User registration (C-end self-service, B-end admin-created)  
- Activation / Freeze / Deletion  
- Profile management (nickname, avatar, email, phone, extended profile)  

### 2. Authentication
- Username + password login  
- Phone/email OTP login (C-end)  
- OAuth2.0 third-party login (Google, WeChat, Apple ID, etc.)  
- Access/Refresh token mechanism (support rotation + state table)  
- Multi-session management (web, mobile, tablet)  
- Forced logout (token_version strategy)  
  → combined with session-level revocation for fine-grained control (see Security section).

### 3. Security
- Password hashing (bcrypt/argon2)  
- Login failure limit (anti-brute-force)  
- Two-factor authentication (2FA/MFA)  
- Login & audit logs  
- Abnormal login detection  
- Token/session revocation strategies:
  - **Global**: via `users.token_version` field (all tokens invalidated).
  - **Per-session**: via `sessions.revoked` flag (specific device/session invalidated).
  - **Rotation**: Refresh tokens rotated on use, invalidated once used or revoked.
  - **Hybrid**: short-lived Access Tokens + token_version for global kicks + session revocation for device-level kicks.

### 4. Authorization (via Casbin)
- RBAC implemented with Casbin  
- Custos does not maintain custom `roles/permissions` tables  
- Casbin `casbin_rule` table stores role & permission policies  
- Custos integrates Casbin Enforcer for runtime checks  
- Provides wrapper APIs for managing roles/permissions  
- Future: ABAC using Casbin models  

### 5. OAuth2.0 Federation (Client posture)
- Act as **OAuth2.0/OIDC client** to external IdPs (Google, GitHub, WeChat).
- Implement callback endpoints: `/oauth/{provider}/callback` (authorization-code exchange).
- Normalize external identities into `user_oauth` (one user can bind multiple providers).
- External tokens are used only to fetch identity; Custos then issues **internal JWT** for Fly.
- **Non-goal**: do not expose `/authorize` or `/token` as an IdP for third parties.

### 6. Audit & Observability
- Login events (success/failure, IP, UA)  
- Permission change logs  
- Security events (forced logout, reused refresh token detection)  
- Export to MQ/ES/Prometheus  

### 7. Identity Linking & Account Merge
- Support binding multiple external identities (wechat/google/github) to a single local user.
- Provide **account merge** flow (strong re-auth on the target account):
  1) verify owner of the primary account;
  2) migrate `user_oauth`, profile and domain references to the primary `user_id`;
  3) mark secondary account `status=merged`.
- Record bind/unbind/merge events in audit logs.

### 8. Internal Token Authority & Key Management
- Custos is the **internal token issuer** (calls Mora `auth` to sign JWT).
- Provide internal **JWKS** endpoint for service verification (Clotho/Orders/Payments).
- Implement **key rotation** with `kid`; old keys remain available for verification until retired.
- Keys are stored in KMS or filesystem; DB keeps **public key metadata** only.

---

## Out of Scope
The User Domain **does not handle**:  
- Trust/Zero Trust (device, IP, network validation → handled by Clotho)  
- Infrastructure capabilities (logging, config, db, mq → handled by Mora)  
- Other business domains (orders, payments, etc.)  

---

The schema below models local users, profile extensions, external identity bindings, refresh tokens, and (optional) session & key metadata to support rotation and granular revocation.

## Database Schema (DDL)

### users
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,                           -- 用户ID，主键
    username VARCHAR(64) UNIQUE,                                    -- 本地用户名（可选，第三方登录可为空）
    email VARCHAR(128) UNIQUE,                                      -- 邮箱（可选）
    phone VARCHAR(32) UNIQUE,                                       -- 手机号（可选）
    password_hash VARCHAR(255),                                     -- 本地密码哈希（bcrypt/argon2）
    user_type ENUM('customer','staff','partner') DEFAULT 'customer',-- 用户类型
    tenant_id BIGINT NULL,                                          -- 租户ID（多租户）
    status ENUM('active','disabled','locked','deleted','merged')    -- 账户状态，含合并态
           DEFAULT 'active',
    token_version INT DEFAULT 0,                                    -- 强制下线版本号
    merged_into_user_id BIGINT NULL,                                -- 若合并：指向主账户ID
    last_login_at DATETIME NULL,                                    -- 最近登录时间
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,                  -- 创建时间
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
               ON UPDATE CURRENT_TIMESTAMP,                         -- 更新时间
    CONSTRAINT fk_users_merged_into FOREIGN KEY (merged_into_user_id)
        REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX idx_users_tenant ON users(tenant_id);
```

### user_profiles
```sql
CREATE TABLE user_profiles (
    user_id BIGINT PRIMARY KEY, -- 用户ID，关联 users.id
    nickname VARCHAR(64), -- 昵称
    avatar VARCHAR(255), -- 头像URL
    gender ENUM('male','female','other') DEFAULT 'other', -- 性别
    birthday DATE, -- 出生日期
    extra JSON, -- 扩展信息
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE -- 级联删除
);
```

### user_oauth
```sql
CREATE TABLE user_oauth (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,              -- 绑定记录ID
    user_id BIGINT NOT NULL,                           -- 本地用户ID
    provider VARCHAR(64) NOT NULL,                     -- 提供方: wechat/google/github
    provider_uid VARCHAR(128) NOT NULL,                -- 提供方用户唯一ID: openid/sub/id
    access_token VARCHAR(255),                         -- 第三方Access Token（可选保存）
    refresh_token VARCHAR(255),                        -- 第三方Refresh Token（可选保存）
    expires_at DATETIME,                               -- 第三方Token过期时间
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,     -- 创建时间
    UNIQUE(provider, provider_uid),                    -- 同一提供方+UID唯一
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_user_oauth_user_provider ON user_oauth(user_id, provider);
```

### refresh_tokens
```sql
CREATE TABLE refresh_tokens (
    id BIGINT PRIMARY KEY AUTO_INCREMENT, -- 刷新Token记录ID
    user_id BIGINT NOT NULL, -- 用户ID
    token_hash CHAR(64) NOT NULL, -- 刷新Token的哈希值
    is_used BOOLEAN DEFAULT FALSE, -- 是否已使用（用于一次性刷新）
    expires_at DATETIME NOT NULL, -- 过期时间
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE -- 级联删除
);
```

### sessions (optional but recommended)
```sql
CREATE TABLE sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,                -- 会话记录ID
    user_id BIGINT NOT NULL,                             -- 用户ID
    session_id CHAR(36) NOT NULL,                        -- 会话ID（UUID）
    refresh_token_id BIGINT NULL,                        -- 关联的刷新Token记录
    device_id VARCHAR(128),                              -- 设备ID（可选）
    user_agent VARCHAR(255),                             -- UA
    ip VARCHAR(45),                                      -- 登录IP (IPv4/IPv6)
    revoked BOOLEAN DEFAULT FALSE,                       -- 是否已撤销
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,       -- 创建时间
    last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,     -- 最近活跃
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (refresh_token_id) REFERENCES refresh_tokens(id) ON DELETE SET NULL,
    UNIQUE(session_id)
);
CREATE INDEX idx_sessions_user ON sessions(user_id);
```

### jwk_keys (public metadata for rotation)
```sql
CREATE TABLE jwk_keys (
    kid VARCHAR(64) PRIMARY KEY,                         -- Key ID
    alg VARCHAR(16) NOT NULL,                            -- 算法，如 RS256/ES256
    public_jwk JSON NOT NULL,                            -- 公钥（JWK 格式）
    active BOOLEAN DEFAULT TRUE,                         -- 是否激活
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,       -- 创建时间
    rotated_at DATETIME NULL,                            -- 轮换时间
    retired_at DATETIME NULL                             -- 退役时间
);
```

---

## Public API Surface (called by Clotho)
- `POST /v1/auth/login` → local username/password login
- `POST /v1/auth/refresh` → rotate refresh token, return new access token
- `POST /v1/auth/logout` → revoke current session
- `POST /v1/auth/force-logout` → admin/ops revoke by user_id or session_id
- `GET  /v1/users/me` → current user info
- `GET  /v1/oauth/{provider}/login` → redirect to IdP authorize URL
- `POST /v1/oauth/{provider}/callback` → exchange code for token, bind or create user, issue internal JWT
- `POST /v1/oauth/{provider}/bind` → bind third-party identity to current user
- `POST /v1/account/merge` → merge secondary account into primary (strong re-auth required)
- `GET  /internal/jwks.json` → internal JWKS for service verification

---

## Instructions to AI
When generating code or architecture:  
- Assume **RBAC = Casbin**, no custom role/permission tables  
- Generate Golang structs, APIs, middleware, and services based on Custos responsibilities  
- Use Mora for common utilities (auth token signing, config, db connection, etc.)  
- Integrate Casbin Enforcer in middleware for RBAC checks  
- Generate starter code for login, logout, token refresh, forced logout, audit logging, Casbin role assignment  
- Follow clean architecture principles (Mora = capability library, Custos = user domain, Clotho = API orchestrator)  
- Treat Custos as an **IdP client**, not an external IdP provider; do **not** implement `/authorize` or `/token` for third parties.  
- Prefer per-tenant RBAC using Casbin "with domains" if multi-tenancy is enabled (scope decisions by tenant_id).  
- Implement key rotation with `kid`, and design token claims: {sub, tid, jti, iat, exp, iss, aud, scope, roles, ver=token_version}.  
- Implement hybrid forced logout: use token_version for global revocation, sessions.revoked for device-level revocation, and refresh token rotation for high security.
