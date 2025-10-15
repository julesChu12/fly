package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "userd",
	Short: "Custos User & Authentication Service",
	Long: `Custos is the authentication and authorization service in the Fly platform.
Named after the Latin word for guardian/keeper, it handles user registration, authentication, JWT token management, and RBAC authorization.`,
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
