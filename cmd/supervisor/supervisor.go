package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/h4-poc/service/cmd/supervisor/commands"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/util"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:               "supervisor",
		Short:             "supervisor H4 management",
		Long:              `Supervisor is a tool for initializing and managing application deployments.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	}

	ctx := context.Background()
	lgr := log.FromLogrus(logrus.NewEntry(logrus.New()), &log.LogrusConfig{Level: "info"})
	ctx = log.WithLogger(ctx, lgr)
	ctx = util.ContextWithCancelOnSignals(ctx, syscall.SIGINT, syscall.SIGTERM)

	rootCmd.AddCommand(commands.NewVersionCommand())
	rootCmd.AddCommand(commands.NewBootstrapCmd())
	rootCmd.AddCommand(commands.NewProjectCmd())
	rootCmd.AddCommand(commands.NewStatusCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
