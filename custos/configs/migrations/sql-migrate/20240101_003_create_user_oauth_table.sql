-- +migrate Up
-- 创建OAuth账号绑定表
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

-- +migrate Down
DROP TABLE IF EXISTS user_oauth;