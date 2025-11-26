// Package main provides the entry point for Appointments Service
// 预约服务的主入口，提供预约管理功能，支持HTTP和gRPC双协议
package main

import (
	"log"
	"os"

	"github.com/julesChu12/fly/appointments/cmd/appointments/cmd"
)

// @title Appointments Service API
// @version 1.0
// @description 预约管理服务API文档，支持预约的创建、查询、状态更新和日历视图操作
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.fly.com/support
// @contact.email support@fly.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8083
// @BasePath /api

// @tag.name 预约管理
// @tag.description 预约信息管理相关接口

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}