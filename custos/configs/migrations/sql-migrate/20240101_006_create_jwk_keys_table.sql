-- +migrate Up
-- 创建JWK密钥元数据表
CREATE TABLE IF NOT EXISTS jwk_keys (
    kid VARCHAR(64) PRIMARY KEY COMMENT 'Key ID',
    alg VARCHAR(16) NOT NULL COMMENT '算法，如 RS256/ES256',
    public_jwk JSON NOT NULL COMMENT '公钥（JWK 格式）',
    active BOOLEAN DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rotated_at TIMESTAMP NULL COMMENT '轮换时间',
    retired_at TIMESTAMP NULL COMMENT '退役时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS jwk_keys;