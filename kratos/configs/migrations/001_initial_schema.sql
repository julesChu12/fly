

-- 001_initial_schema.sql
-- Kratos Order Service - 初始数据库结构
-- 负责订单主数据、订单明细、订单状态日志
-- 能力：多租户 / 状态机 / 幂等性 / 审计追踪

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- Table: orders
-- ============================================
CREATE TABLE IF NOT EXISTS orders (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '订单ID',
  tenant_id BIGINT NOT NULL COMMENT '租户ID（多租户隔离）',
  order_no VARCHAR(64) NOT NULL UNIQUE COMMENT '业务订单号（幂等创建标识）',
  customer_id BIGINT NOT NULL COMMENT '客户ID（来自Hermes）',
  total_amount DECIMAL(12,2) NOT NULL COMMENT '订单总金额',
  currency CHAR(3) NOT NULL DEFAULT 'CNY' COMMENT '货币类型',
  status ENUM('pending','paid','fulfilled','canceled') NOT NULL DEFAULT 'pending' COMMENT '订单状态机',
  remark VARCHAR(255) NULL COMMENT '订单备注',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  INDEX idx_tenant (tenant_id),
  INDEX idx_customer (customer_id),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';

-- ============================================
-- Table: order_items
-- ============================================
CREATE TABLE IF NOT EXISTS order_items (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '订单明细ID',
  tenant_id BIGINT NOT NULL COMMENT '租户ID（多租户隔离）',
  order_id BIGINT NOT NULL COMMENT '关联订单ID',
  product_id BIGINT NULL COMMENT '商品ID（来自产品服务或外部系统）',
  product_name VARCHAR(256) NOT NULL COMMENT '商品名称（快照）',
  sku VARCHAR(128) NULL COMMENT 'SKU 编码（选填）',
  quantity INT NOT NULL DEFAULT 1 COMMENT '数量',
  unit_price DECIMAL(12,2) NOT NULL COMMENT '单价',
  total_price DECIMAL(12,2) GENERATED ALWAYS AS (quantity * unit_price) STORED COMMENT '行小计（自动计算）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  INDEX idx_order (order_id),
  INDEX idx_tenant (tenant_id),
  CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表';

-- ============================================
-- Table: order_status_logs
-- ============================================
CREATE TABLE IF NOT EXISTS order_status_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
  tenant_id BIGINT NOT NULL COMMENT '租户ID',
  order_id BIGINT NOT NULL COMMENT '订单ID',
  from_status ENUM('pending','paid','fulfilled','canceled') NULL COMMENT '变更前状态',
  to_status ENUM('pending','paid','fulfilled','canceled') NOT NULL COMMENT '变更后状态',
  reason VARCHAR(255) NULL COMMENT '状态变更原因',
  operator_id BIGINT NULL COMMENT '操作人ID（来自Custos）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  INDEX idx_order (order_id),
  INDEX idx_tenant (tenant_id),
  CONSTRAINT fk_order_logs_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单状态变更日志表';

-- ============================================
-- Table: order_audit
-- ============================================
CREATE TABLE IF NOT EXISTS order_audit (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '审计ID',
  tenant_id BIGINT NOT NULL COMMENT '租户ID',
  order_id BIGINT NOT NULL COMMENT '订单ID',
  action VARCHAR(64) NOT NULL COMMENT '操作类型（CREATE/UPDATE/DELETE）',
  actor VARCHAR(128) NULL COMMENT '操作者',
  payload JSON NULL COMMENT '操作内容（JSON 格式快照）',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  INDEX idx_order (order_id),
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单审计记录表';

SET FOREIGN_KEY_CHECKS = 1;