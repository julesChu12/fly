// Package main provides the entry point for Hermes Customer Service
// Hermes服务的主入口，提供客户管理功能，支持HTTP和gRPC双协议
package main

import (
	"log"
	"os"

	"github.com/julesChu12/fly/hermes/cmd/hermes/cmd"
)

// @title Hermes Customer Service API
// @version 1.0
// @description 客户管理服务API文档，支持客户信息和联系方式的增删改查操作
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.fly.com/support
// @contact.email support@fly.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

// @tag.name 客户管理
// @tag.description 客户信息管理相关接口

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
