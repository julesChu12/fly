-- OAuth 账号绑定表
CREATE TABLE IF NOT EXISTS user_oauth (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    provider VARCHAR(64) NOT NULL COMMENT '提供方: google/github/wechat',
    provider_uid VARCHAR(128) NOT NULL COMMENT '提供方用户唯一ID',
    access_token VARCHAR(255) COMMENT '第三方Access Token（可选保存）',
    refresh_token VARCHAR(255) COMMENT '第三方Refresh Token（可选保存）',
    expires_at TIMESTAMP NULL COMMENT '第三方Token过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_provider_uid (provider, provider_uid),
    KEY idx_user_provider (user_id, provider),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 刷新Token表
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    token_hash CHAR(64) NOT NULL COMMENT 'SHA-256 hash of refresh token',
    is_used BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    KEY idx_user_id (user_id),
    KEY idx_token_hash (token_hash),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 会话管理表
CREATE TABLE IF NOT EXISTS sessions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    session_id CHAR(36) NOT NULL COMMENT '会话ID（UUID）',
    refresh_token_id BIGINT UNSIGNED NULL,
    device_id VARCHAR(128) COMMENT '设备ID（可选）',
    user_agent VARCHAR(500) COMMENT '用户代理',
    ip VARCHAR(45) COMMENT '登录IP (IPv4/IPv6)',
    revoked BOOLEAN DEFAULT FALSE COMMENT '是否已撤销',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_session_id (session_id),
    KEY idx_user_id (user_id),
    KEY idx_refresh_token_id (refresh_token_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (refresh_token_id) REFERENCES refresh_tokens(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- JWK 密钥元数据表
CREATE TABLE IF NOT EXISTS jwk_keys (
    kid VARCHAR(64) PRIMARY KEY COMMENT 'Key ID',
    alg VARCHAR(16) NOT NULL COMMENT '算法，如 RS256/ES256',
    public_jwk JSON NOT NULL COMMENT '公钥（JWK 格式）',
    active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rotated_at TIMESTAMP NULL COMMENT '轮换时间',
    retired_at TIMESTAMP NULL COMMENT '退役时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 更新用户表，添加多租户和强制下线支持
ALTER TABLE users
ADD COLUMN user_type ENUM('customer','staff','partner') DEFAULT 'customer' COMMENT '用户类型',
ADD COLUMN tenant_id BIGINT UNSIGNED NULL COMMENT '租户ID（多租户）',
ADD COLUMN token_version INT DEFAULT 0 COMMENT '强制下线版本号',
ADD COLUMN merged_into_user_id BIGINT UNSIGNED NULL COMMENT '若合并：指向主账户ID',
ADD COLUMN last_login_at TIMESTAMP NULL COMMENT '最近登录时间',
ADD INDEX idx_tenant_id (tenant_id),
ADD INDEX idx_token_version (token_version);

-- 更新用户状态枚举，添加合并状态
ALTER TABLE users MODIFY COLUMN status ENUM('active','disabled','locked','deleted','merged') DEFAULT 'active';

-- 用户扩展信息表
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id BIGINT UNSIGNED PRIMARY KEY,
    nickname VARCHAR(64),
    avatar VARCHAR(255),
    gender ENUM('male','female','other') DEFAULT 'other',
    birthday DATE,
    extra JSON COMMENT '扩展信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;