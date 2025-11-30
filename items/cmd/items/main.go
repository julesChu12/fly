// Package main provides the entry point for Items Product Management Service
// Items 服务的主入口，提供商品管理、库存管理、分类管理等功能
package main

import (
	"log"
	"os"

	"github.com/julesChu12/fly/items/cmd/items/cmd"
)

// @title Items Product Management Service API
// @version 1.0
// @description 统一商品管理服务API文档，提供商品管理、库存管理、分类管理等功能
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.fly.com/support
// @contact.email support@fly.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8086
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token for authentication

// @tag.name 商品
// @tag.description 商品相关接口（增删改查、搜索等）

// @tag.name 库存
// @tag.description 库存管理相关接口

// @tag.name 分类
// @tag.description 商品分类相关接口

// @tag.name 健康检查
// @tag.description 服务健康检查接口

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}