package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/h4-poc/service/cmd/service/server"
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Application API server",
	Long:  `Application API server is a service for managing applications.`,
}

var (
	Version   string
	BuildTime string
	GitCommit string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(server.RunCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `All software has versions. This is Application API's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Application API v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
	},
}
