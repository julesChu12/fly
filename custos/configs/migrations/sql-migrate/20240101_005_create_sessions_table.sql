-- +migrate Up
-- 创建会话管理表
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

-- +migrate Down
DROP TABLE IF EXISTS sessions;