#!/bin/bash

# UTF8MB4 编码验证脚本
# 用于测试数据库是否能正确处理 4 字节 UTF8 字符

echo "🔍 开始验证 UTF8MB4 编码配置..."

# 连接到 MySQL 并验证字符集配置
echo "📊 检查数据库字符集配置..."
docker exec fly-mysql mysql -u fly_user -prootpassword items_dev -e "
SELECT
    SCHEMA_NAME as '数据库',
    DEFAULT_CHARACTER_SET_NAME as '字符集',
    DEFAULT_COLLATION_NAME as '排序规则'
FROM information_schema.SCHEMATA
WHERE SCHEMA_NAME = 'items_dev';

SELECT
    TABLE_NAME as '表名',
    TABLE_COLLATION as '表排序规则'
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'items_dev';

SELECT
    COLUMN_NAME as '列名',
    DATA_TYPE as '数据类型',
    CHARACTER_SET_NAME as '字符集',
    COLLATION_NAME as '排序规则'
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'items_dev'
  AND TABLE_NAME IN ('categories', 'items')
  AND DATA_TYPE IN ('varchar', 'text', 'char');
"

# 测试插入包含 4 字节 UTF8 字符的数据
echo "🧪 测试插入 4 字节 UTF8 字符..."
docker exec fly-mysql mysql -u fly_user -prootpassword items_dev -e "
INSERT INTO categories (id, name, description, sort_order, status) VALUES
('test-utf8mb4-001', 'Emoji 测试分类 🎉🔥💯', '这是一个包含 Emoji 和特殊字符的描述：👨‍💻👩‍🎨🌈✨', 999, 'ACTIVE');

INSERT INTO items (id, name, description, type, price, category_id, status, sku, stock) VALUES
('test-utf8mb4-002', '测试商品 🎁', '商品描述包含复杂 Emoji：🎪🎭🎨🏆🎯', 'PRODUCT', 99.99, 'test-utf8mb4-001', 'ACTIVE', 'TEST-001', 10);
"

# 验证数据是否正确存储
echo "✅ 验证数据是否正确存储..."
docker exec fly-mysql mysql -u fly_user -prootpassword items_dev -e "
SELECT id, name, description, status FROM categories WHERE id = 'test-utf8mb4-001';
SELECT id, name, description, status, sku FROM items WHERE id = 'test-utf8mb4-002';
"

# 检查字符长度是否正确（4 字节字符应该占用 1 个字符位置）
echo "📏 验证字符长度计算..."
docker exec fly-mysql mysql -u fly_user -prootpassword items_dev -e "
SELECT
    name,
    CHAR_LENGTH(name) as '字符数',
    LENGTH(name) as '字节数',
    description,
    CHAR_LENGTH(description) as '字符数',
    LENGTH(description) as '字节数'
FROM categories
WHERE id = 'test-utf8mb4-001';
"

echo "🎉 UTF8MB4 编码验证完成！"