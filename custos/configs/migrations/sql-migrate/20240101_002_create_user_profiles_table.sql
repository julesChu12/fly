-- +migrate Up
-- 创建用户扩展信息表
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

-- +migrate Down
DROP TABLE IF EXISTS user_profiles;