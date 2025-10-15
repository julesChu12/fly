package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hermes",
	Short: "Hermes Customer Service",
	Long: `Hermes is the customer management service in the Fly platform.
It provides customer information and contact management functionality with HTTP REST API and gRPC protocol support.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	// Global configuration initialization logic can be added here
}

func init() {
	cobra.OnInitialize(initConfig)
}
