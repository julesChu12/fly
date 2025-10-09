

-- 001_initial_schema.sql
-- Plutus (Payment & Wallet Service) - 初始数据库结构
-- 能力：钱包余额 / 交易流水 / 支付渠道 / 多租户 / 幂等键 / 对账预留

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- Table: wallets
-- 每个 (tenant_id, customer_id) 唯一钱包；
-- 建议在应用层对余额修改使用事务 + SELECT ... FOR UPDATE。
-- ============================================
CREATE TABLE IF NOT EXISTS wallets (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '钱包ID',
  tenant_id BIGINT NOT NULL COMMENT '租户ID（多租户隔离）',
  customer_id BIGINT NOT NULL COMMENT '客户ID（Hermes.customers.id）',
  balance DECIMAL(14,2) NOT NULL DEFAULT 0 COMMENT '当前余额',
  currency CHAR(3) NOT NULL DEFAULT 'CNY' COMMENT '货币：ISO 4217',
  status ENUM('active','frozen') NOT NULL DEFAULT 'active' COMMENT '钱包状态',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY uk_tenant_customer (tenant_id, customer_id),
  KEY idx_tenant (tenant_id),
  KEY idx_customer (customer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户钱包账户';

-- ============================================
-- Table: transactions
-- 不可变更的交易流水；所有资金变动必须写入该表；
-- 幂等键 idempotency_key 保障外部回调/重试的幂等性；
-- 可与订单（Kratos.orders）及支付渠道对接。
-- ============================================
CREATE TABLE IF NOT EXISTS transactions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '交易ID',
  tenant_id BIGINT NOT NULL COMMENT '租户ID',
  wallet_id BIGINT NOT NULL COMMENT '关联钱包ID（FK wallets.id）',
  order_id BIGINT NULL COMMENT 'Kratos.orders.id；纯充值可为空',
  type ENUM('recharge','consume','refund') NOT NULL COMMENT '交易类型',
  amount DECIMAL(14,2) NOT NULL COMMENT '交易金额（正数）',
  currency CHAR(3) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  channel ENUM('wallet','wechat','alipay','stripe','paypal','bank','other') NOT NULL DEFAULT 'wallet' COMMENT '支付渠道',
  status ENUM('pending','success','failed') NOT NULL DEFAULT 'pending' COMMENT '交易状态',
  idempotency_key VARCHAR(128) NULL COMMENT '幂等键（外部请求唯一标识）',
  reference_no VARCHAR(128) NULL COMMENT '外部渠道流水号/参考号',
  meta JSON NULL COMMENT '扩展信息：回执/错误码/签名校验等',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  KEY idx_wallet (wallet_id),
  KEY idx_tenant (tenant_id),
  KEY idx_order (order_id),
  KEY idx_status (status),
  UNIQUE KEY uk_tenant_idemp (tenant_id, idempotency_key),
  CONSTRAINT fk_tx_wallet FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资金交易流水（不可变更）';

-- ============================================
-- Table: payment_channels
-- 保存第三方支付网关配置与回执存档；按租户隔离；
-- 生产中建议把敏感字段加密或放入密钥管理系统。
-- ============================================
CREATE TABLE IF NOT EXISTS payment_channels (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '渠道配置ID',
  tenant_id BIGINT NOT NULL COMMENT '租户ID',
  name VARCHAR(64) NOT NULL COMMENT '渠道名（展示）',
  type ENUM('wechat','alipay','stripe','paypal','bank','other') NOT NULL COMMENT '渠道类型',
  config JSON NULL COMMENT '配置JSON：app_id、mch_id、api_key、证书路径、回调URL等',
  status ENUM('active','inactive') NOT NULL DEFAULT 'active' COMMENT '状态',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY uk_tenant_name (tenant_id, name),
  KEY idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='第三方支付渠道配置';

-- ============================================
-- Table: channel_callbacks
-- 可选：保存第三方回调原文，用于审计与调试；
-- 大字段建议冷热分层或短期存储策略。
-- ============================================
CREATE TABLE IF NOT EXISTS channel_callbacks (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '回调记录ID',
  tenant_id BIGINT NOT NULL COMMENT '租户ID',
  channel_id BIGINT NULL COMMENT 'payment_channels.id（可空）',
  reference_no VARCHAR(128) NULL COMMENT '外部单号/回执号',
  request_body MEDIUMTEXT NULL COMMENT '回调请求原文',
  response_body MEDIUMTEXT NULL COMMENT '回调应答原文',
  http_status INT NULL COMMENT 'HTTP 状态码',
  verified TINYINT(1) NOT NULL DEFAULT 0 COMMENT '签名是否通过校验',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '回调时间',
  KEY idx_tenant (tenant_id),
  KEY idx_channel (channel_id),
  KEY idx_reference (reference_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='第三方回调留痕';

-- ============================================
-- View/Report: wallet_balances (可选)
-- 聚合余额示例视图，方便报表与排错（MySQL 8+ 可创建视图）。
-- ============================================
-- CREATE OR REPLACE VIEW wallet_balances AS
-- SELECT w.id AS wallet_id, w.tenant_id, w.customer_id, w.balance, w.currency,
--        COUNT(t.id) AS tx_count,
--        SUM(CASE WHEN t.type='recharge' AND t.status='success' THEN t.amount ELSE 0 END) AS total_recharge,
--        SUM(CASE WHEN t.type='consume'  AND t.status='success' THEN t.amount ELSE 0 END) AS total_consume,
--        SUM(CASE WHEN t.type='refund'   AND t.status='success' THEN t.amount ELSE 0 END) AS total_refund
-- FROM wallets w
-- LEFT JOIN transactions t ON t.wallet_id = w.id
-- GROUP BY w.id, w.tenant_id, w.customer_id, w.balance, w.currency;

SET FOREIGN_KEY_CHECKS = 1;