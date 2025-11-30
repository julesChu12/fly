package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	AppName   = "Appointments Service"
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = "go1.25.1"
	Compiler  = "gc"
	Platform  = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version information for Appointments Service, including build details.`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func init() {
	// Add version command to root command
}

func printVersion() {
	fmt.Printf(`
%s
  Version:    %s
  Git Commit: %s
  Build Time: %s
  Go Version: %s
  Compiler:   %s
  Platform:   %s

Copyright (c) 2024 Fly CRM. All rights reserved.
`, AppName, Version, GitCommit, BuildTime, GoVersion, Compiler, Platform)
}
