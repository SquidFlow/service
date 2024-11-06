package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/h4-poc/service/cmd/service/server"
	"github.com/h4-poc/service/pkg/log"
	"github.com/h4-poc/service/pkg/store"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Application API server",
	Long: `Application API server is a service for managing applications.
It provides RESTful APIs for managing applications, projects, and security configurations.`,
	PersistentPreRun: preRun,
}

func init() {
	logger := log.FromLogrus(logrus.NewEntry(logrus.New()), &log.LogrusConfig{
		Level:  "debug",
		Format: log.FormatterText,
	})
	log.SetDefault(logger)

	store.Get().Version = store.Version{
		Version:    version,
		GitCommit:  commit,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		GoCompiler: runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	rootCmd.AddCommand(server.NewRunCommand())
	rootCmd.AddCommand(server.NewVersionCommand())

	log.G().AddPFlags(rootCmd)
}

func preRun(cmd *cobra.Command, args []string) {
	logger := log.G()
	if err := logger.Configure(); err != nil {
		fmt.Printf("Failed to configure logger: %v\n", err)
		os.Exit(1)
	}

	log.G().WithFields(log.Fields{
		"version":     version,
		"commit":      commit,
		"build_date":  buildDate,
		"go_version":  runtime.Version(),
		"platform":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		"pid":         os.Getpid(),
		"working_dir": getCurrentWorkingDir(),
	}).Info("Starting application server")
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.G().Fatalf("Application panic: %v", r)
		}
	}()

	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		log.G().WithError(err).Fatal("Failed to start application")
	}
}

func getCurrentWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.G().WithError(err).Error("Failed to get working directory")
		return "unknown"
	}
	return dir
}
