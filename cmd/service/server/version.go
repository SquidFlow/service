package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/h4-poc/service/pkg/store"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show cli version",
		Run: func(_ *cobra.Command, _ []string) {
			s := store.Get()
			fmt.Printf("Version: %s\n", s.Version.Version)
			fmt.Printf("BuildDate: %s\n", s.Version.BuildDate)
			fmt.Printf("GitCommit: %s\n", s.Version.GitCommit)
			fmt.Printf("GoVersion: %s\n", s.Version.GoVersion)
			fmt.Printf("GoCompiler: %s\n", s.Version.GoCompiler)
			fmt.Printf("Platform: %s\n", s.Version.Platform)
		},
	}

	return cmd
}
