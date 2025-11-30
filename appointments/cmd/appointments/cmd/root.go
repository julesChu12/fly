package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "appointments",
	Short: "Appointments Service - 预约管理服务",
	Long:  `Appointments Service 是一个微服务，负责管理客户预约信息，包括预约的创建、查询、状态更新和日历视图功能。支持HTTP和gRPC双协议访问。`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)

	// 设置全局标志
	rootCmd.PersistentFlags().StringP("config", "c", "", "配置文件路径 (默认: configs/appointments.yaml)")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "日志级别 (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")
}

func handleRootCommand(cmd *cobra.Command, args []string) {
	fmt.Println(`
Appointments Service - 预约管理服务

可用命令:
  serve     启动HTTP和gRPC服务器
  version   显示版本信息
  help      显示帮助信息

使用 'appointments help [command]' 查看具体命令的帮助信息。`)
	os.Exit(0)
}
