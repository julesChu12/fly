# Mora 模块发布和使用指南

## 📦 已发布版本

- **v1.0.0** (2025-01-20) - 初始稳定版本

## 🚀 在其他模块中使用 Mora

### 方式 1: 使用版本号（推荐用于生产环境）

在 `go.mod` 中添加：

```go
require (
    github.com/julesChu12/fly/mora v1.0.0
)
```

然后运行：
```bash
go get github.com/julesChu12/fly/mora@v1.0.0
go mod tidy
```

### 方式 2: 使用最新版本

```bash
go get github.com/julesChu12/fly/mora@latest
go mod tidy
```

### 方式 3: 使用特定主版本

```bash
go get github.com/julesChu12/fly/mora@v1
go mod tidy
```

### 方式 4: 本地开发（使用 replace）

在 `go.mod` 中添加：

```go
require (
    github.com/julesChu12/fly/mora v1.0.0
)

// 本地开发时使用本地路径
replace github.com/julesChu12/fly/mora => ../mora
```

## 📝 使用示例

### 导入 Mora 包

```go
import (
    "github.com/julesChu12/fly/mora/pkg/auth"
    "github.com/julesChu12/fly/mora/pkg/logger"
    "github.com/julesChu12/fly/mora/pkg/config"
    "github.com/julesChu12/fly/mora/pkg/db"
    "github.com/julesChu12/fly/mora/pkg/cache"
    "github.com/julesChu12/fly/mora/pkg/observability"
    
    // 框架适配器
    ginAdapter "github.com/julesChu12/fly/mora/adapters/gin"
    gozeroAdapter "github.com/julesChu12/fly/mora/adapters/gozero"
)
```

### 基本使用

```go
// 1. 初始化日志
logger, _ := logger.New(logger.Config{
    Level:  "info",
    Format: "json",
})

// 2. 加载配置
cfg, _ := config.New().
    WithYAML("config.yaml").
    WithEnvPrefix("APP").
    Load()

// 3. 初始化数据库
dbClient, _ := db.NewGORMClient(cfg.GetString("database.dsn"))

// 4. 初始化缓存
cacheClient, _ := cache.NewRedisClient(cache.Config{
    Address: cfg.GetString("redis.address"),
})

// 5. 使用 JWT
token, _ := auth.GenerateToken("user123", "secret", time.Hour)
claims, _ := auth.ValidateToken(token, "secret")
```

## 🔄 版本更新

### 更新到新版本

```bash
# 更新到最新版本
go get -u github.com/julesChu12/fly/mora@latest

# 更新到特定版本
go get github.com/julesChu12/fly/mora@v1.1.0

# 清理未使用的依赖
go mod tidy
```

### 查看可用版本

```bash
go list -m -versions github.com/julesChu12/fly/mora
```

## 🛠️ 发布新版本

### 使用发布脚本

```bash
cd mora
./scripts/release.sh v1.1.0
```

### 手动发布

```bash
# 1. 运行测试
go test ./...

# 2. 创建标签
git tag -a mora/v1.1.0 -m "Release mora v1.1.0"

# 3. 推送标签
git push origin mora/v1.1.0
```

## 📚 更多信息

- [完整文档](README.md)
- [变更日志](CHANGELOG.md)
- [使用示例](starter/)

## 🔗 相关链接

- GitHub 仓库: https://github.com/julesChu12/fly
- 标签地址: https://github.com/julesChu12/fly/releases/tag/mora/v1.0.0

