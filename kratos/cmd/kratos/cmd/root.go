package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kratos",
	Short: "Kratos Order Service",
	Long: `Kratos is the order management service in the Fly platform.
Named after the Greek god of strength and power, it handles order creation, lifecycle management, and transaction coordination.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
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
