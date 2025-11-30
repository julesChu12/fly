// Package main provides the entry point for Custos User & Authentication Service
// Custos 服务的主入口，提供用户认证和授权功能，支持HTTP和gRPC双协议
package main

import (
	"log"
	"os"

	"github.com/julesChu12/fly/custos/cmd/cmd"
)

// @title Custos User & Authentication Service API
// @version 1.0
// @description 用户认证和授权服务API文档，提供用户注册、登录、OAuth、RBAC等功能
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.fly.com/support
// @contact.email support@fly.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name 认证
// @tag.description 用户认证相关接口（注册、登录、刷新令牌等）

// @tag.name 密码管理
// @tag.description 密码相关接口（验证、修改密码等）

// @tag.name OAuth
// @tag.description OAuth 第三方登录相关接口

// @tag.name 用户配置
// @tag.description 用户个人配置管理接口

// @tag.name 管理员
// @tag.description 管理员功能相关接口（需要管理员权限）

// @tag.name 租户
// @tag.description 多租户管理相关接口

// @tag.name 健康检查
// @tag.description 服务健康检查接口

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
