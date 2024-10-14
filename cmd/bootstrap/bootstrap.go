package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"github.com/h4-poc/service/cmd/bootstrap/project"
	"github.com/h4-poc/service/cmd/bootstrap/run"
	"github.com/h4-poc/service/cmd/bootstrap/status"
)

var (
	cfgFile   string
	Version   string
	BuildTime string
	GitCommit string
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap application management",
		Long:  `Bootstrap is a tool for initializing and managing application deployments.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("Executing PersistentPreRunE for command: %s", cmd.Name())
			if cmd.Name() == "version" {
				return nil
			}

			if cfgFile == "" {
				cfgFile = os.Getenv("BOOTSTRAP_CONFIG")
			}
			if cfgFile == "" {
				return fmt.Errorf("config file is required. Use -c/--config flag or set BOOTSTRAP_CONFIG environment variable")
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (required if BOOTSTRAP_CONFIG env var is not set)")

	rootCmd.AddCommand(run.Cmd())
	rootCmd.AddCommand(status.Cmd())
	rootCmd.AddCommand(project.Cmd())
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
