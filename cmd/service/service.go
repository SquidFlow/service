package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/h4-poc/service/cmd/service/server"
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
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := strings.Split(f.File, "service/")[1]
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
	log.SetReportCaller(true)

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

	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().Bool("json-logs", false, "Output logs in JSON format")
}

func preRun(cmd *cobra.Command, args []string) {
	logLevel, err := cmd.Flags().GetString("log-level")
	if err != nil {
		log.Fatal(err)
	}
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	log.SetLevel(level)

	jsonLogs, err := cmd.Flags().GetBool("json-logs")
	if err != nil {
		log.Fatal(err)
	}
	if jsonLogs {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	log.WithFields(log.Fields{
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
			log.Fatalf("Application panic: %v", r)
		}
	}()

	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Failed to start application")
	}
}

func getCurrentWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.WithError(err).Error("Failed to get working directory")
		return "unknown"
	}
	return dir
}
