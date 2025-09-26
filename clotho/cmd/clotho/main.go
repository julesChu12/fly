package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clotho",
	Short: "Clotho API Orchestration Layer",
	Long: `Clotho is the API orchestration layer in the Fly monorepo ecosystem.
It exposes HTTP/REST APIs externally and orchestrates calls to internal domain services via gRPC.
Clotho does not implement business logic - it only handles request routing, authentication middleware, and response aggregation.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}