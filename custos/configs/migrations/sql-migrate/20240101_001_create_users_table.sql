-- +migrate Up
-- 创建用户基础表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar VARCHAR(255),
    status ENUM('active','disabled','locked','deleted','merged') DEFAULT 'active',
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    user_type ENUM('customer','staff','partner') DEFAULT 'customer' COMMENT '用户类型',
    tenant_id BIGINT UNSIGNED NULL COMMENT '租户ID（多租户）',
    token_version INT DEFAULT 0 COMMENT '强制下线版本号',
    merged_into_user_id BIGINT UNSIGNED NULL COMMENT '若合并：指向主账户ID',
    last_login_at TIMESTAMP NULL COMMENT '最近登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status),
    INDEX idx_role (role),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_token_version (token_version),
    CONSTRAINT fk_users_merged_into FOREIGN KEY (merged_into_user_id)
        REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS users;