// Package main provides the entry point for Staff Service
// 员工服务的主入口，提供员工管理功能，支持HTTP和gRPC双协议
package main

import (
	"os"

	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/staff/cmd/staff/cmd"
)

// @title Staff Service API
// @version 1.0
// @description 员工管理服务API文档，支持员工信息、角色权限、可用性管理等操作
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.fly.com/support
// @contact.email support@fly.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8084
// @BasePath /api

// @tag.name 员工管理
// @tag.description 员工信息管理相关接口

func main() {
	log := logger.NewDefault()
	if err := cmd.Execute(); err != nil {
		log.Error("Staff service failed to start", "error", err)
		os.Exit(1)
	}
}