package project

import (
	"github.com/spf13/cobra"
)

// TODO: Implement project command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage project",
	}

	return cmd
}
