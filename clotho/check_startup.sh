#!/bin/bash

echo "=== 检查 Clotho 服务启动 ==="

# 1. 编译项目
echo "1. 编译项目..."
go build ./cmd/clotho
if [ $? -eq 0 ]; then
    echo "✅ 编译成功"
else
    echo "❌ 编译失败"
    exit 1
fi

# 2. 启动服务并检查
echo -e "\n2. 启动服务..."
./clotho serve &
CLOTHO_PID=$!

# 等待服务启动
echo "等待服务启动..."
sleep 3

# 3. 检查健康检查端点
echo -e "\n3. 检查健康检查端点..."
curl -s http://localhost:8080/health
if [ $? -eq 0 ]; then
    echo -e "\n✅ 服务启动成功！"
else
    echo -e "\n❌ 服务启动失败"
fi

# 4. 清理
echo -e "\n4. 清理进程..."
kill $CLOTHO_PID 2>/dev/null
wait $CLOTHO_PID 2>/dev/null

echo -e "\n=== 检查完成 ==="