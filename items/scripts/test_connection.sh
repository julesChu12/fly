#!/bin/bash

# 测试数据库连接字符集的脚本

echo "🔍 测试不同的数据库连接方式..."

echo ""
echo "📊 1. 默认连接（可能乱码）:"
docker exec fly-mysql mysql -u fly_user -prootpassword items_dev -e "
SELECT name, description FROM categories LIMIT 1;
"

echo ""
echo "✅ 2. UTF8MB4 连接（正确显示）:"
docker exec fly-mysql mysql -u fly_user -prootpassword items_dev --default-character-set=utf8mb4 -e "
SELECT name, description FROM categories LIMIT 1;
"

echo ""
echo "🔧 3. 检查当前连接字符集:"
docker exec fly-mysql mysql -u fly_user -prootpassword items_dev --default-character-set=utf8mb4 -e "
SHOW VARIABLES LIKE 'character_set%';
"

echo ""
echo "📏 4. 验证字符长度:"
docker exec fly-mysql mysql -u fly_user -prootpassword items_dev --default-character-set=utf8mb4 -e "
SELECT
    name,
    CHAR_LENGTH(name) as '字符数',
    LENGTH(name) as '字节数',
    description,
    CHAR_LENGTH(description) as '描述字符数',
    LENGTH(description) as '描述字节数'
FROM categories
WHERE name LIKE '%美容%'
LIMIT 1;
"

echo ""
echo "💡 5. 解决方案提示:"
echo "   - 应用连接时需添加: charset=utf8mb4&collation=utf8mb4_unicode_ci"
echo "   - 命令行连接使用: --default-character-set=utf8mb4"
echo "   - GUI 客户端设置连接字符集为 utf8mb4"