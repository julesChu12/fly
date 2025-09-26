package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Clotho",
	Long:  `Print the version number and build information of Clotho.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Clotho API Orchestration Layer")
		fmt.Println("Version: 0.1.0")
		fmt.Println("Build: development")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}