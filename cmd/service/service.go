package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/h4-poc/service/cmd/service/server"
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Application API server",
	Long:  `Application API server is a service for managing applications.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(server.NewRunCommand())
	rootCmd.AddCommand(server.NewVersionCommand())
}
