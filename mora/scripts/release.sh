#!/bin/bash
set -e

# Mora 模块发布脚本
# 使用方法: ./scripts/release.sh <version>
# 示例: ./scripts/release.sh v1.0.0

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "❌ 错误: 请提供版本号"
    echo "使用方法: ./scripts/release.sh <version>"
    echo "示例: ./scripts/release.sh v1.0.0"
    exit 1
fi

# 验证版本号格式 (v1.0.0)
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ 错误: 版本号格式不正确，应为 v1.0.0 格式"
    exit 1
fi

echo "🚀 开始发布 mora $VERSION..."
echo ""

# 获取脚本所在目录的父目录（mora 根目录）
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
MORA_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"
FLY_ROOT="$( cd "$MORA_DIR/.." && pwd )"

cd "$MORA_DIR"

# 1. 检查是否有未提交的更改
echo "📋 步骤 1: 检查工作区状态..."
if [ -n "$(git status --porcelain)" ]; then
    echo "⚠️  警告: mora 目录有未提交的更改"
    read -p "是否继续发布? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ 发布已取消"
        exit 1
    fi
fi

# 2. 运行测试
echo ""
echo "📋 步骤 2: 运行测试..."
if ! go test ./...; then
    echo "❌ 测试失败，请修复后重试"
    exit 1
fi
echo "✅ 所有测试通过"

# 3. 检查是否已存在该标签
echo ""
echo "📋 步骤 3: 检查标签..."
cd "$FLY_ROOT"
if git rev-parse "mora/$VERSION" >/dev/null 2>&1; then
    echo "⚠️  警告: 标签 mora/$VERSION 已存在"
    read -p "是否覆盖? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ 发布已取消"
        exit 1
    fi
    git tag -d "mora/$VERSION" 2>/dev/null || true
    git push origin ":refs/tags/mora/$VERSION" 2>/dev/null || true
fi

# 4. 创建标签
echo ""
echo "📋 步骤 4: 创建 Git 标签..."
TAG_MESSAGE="Release mora $VERSION

Mora 是一个框架无关的通用能力库，提供：
- JWT/JWK 认证支持
- 结构化日志 (Zap)
- 配置管理 (YAML + ENV)
- 数据库适配器 (GORM + SQLX)
- Redis 缓存和分布式锁
- 消息队列抽象
- OpenTelemetry 可观测性
- 框架适配器 (Gin + Go-Zero)"

git tag -a "mora/$VERSION" -m "$TAG_MESSAGE"
echo "✅ 标签 mora/$VERSION 已创建"

# 5. 推送标签
echo ""
echo "📋 步骤 5: 推送标签到远程仓库..."
if git push origin "mora/$VERSION"; then
    echo "✅ 标签已推送到远程仓库"
else
    echo "❌ 推送失败，请检查网络连接和权限"
    exit 1
fi

# 6. 显示发布信息
echo ""
echo "🎉 发布成功!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Mora $VERSION 已发布"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📦 其他模块现在可以使用以下命令导入:"
echo ""
echo "   go get github.com/julesChu12/fly/mora@$VERSION"
echo ""
echo "   或在 go.mod 中添加:"
echo "   require github.com/julesChu12/fly/mora $VERSION"
echo ""
echo "🔗 标签地址:"
echo "   https://github.com/julesChu12/fly/releases/tag/mora/$VERSION"
echo ""

