package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "plutus",
	Short: "Plutus Payment & Wallet Service",
	Long: `Plutus is the payment and wallet service in the Fly platform.
Named after the Greek god of wealth, it handles payment processing, wallet management, and transaction recording.`,
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
