package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/h4-poc/service/cmd/supervisor/commands"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:               "supervisor",
		Short:             "supervisor H4 management",
		Long:              `Supervisor is a tool for initializing and managing application deployments.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	}

	rootCmd.AddCommand(commands.NewVersionCommand())
	rootCmd.AddCommand(commands.NewBootstrapCmd())
	rootCmd.AddCommand(commands.NewProjectCmd())
	rootCmd.AddCommand(commands.NewStatusCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
