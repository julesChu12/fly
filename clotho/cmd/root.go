package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clotho",
	Short: "Clotho API Orchestration Layer",
	Long: `Clotho is the API orchestration layer in the Fly monorepo ecosystem.
It exposes HTTP/REST APIs externally and orchestrates calls to internal domain services via gRPC.
Clotho does not implement business logic - it only handles request routing, authentication middleware, and response aggregation.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 添加所有子命令
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	// 这里可以添加全局配置初始化逻辑
}

func init() {
	cobra.OnInitialize(initConfig)
}
