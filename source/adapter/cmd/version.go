package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd's first line is how the git hook tells this program from git-delta, so it must not change.
func versionCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of delta",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "delta (deltas-ai) %s\n", d.Version)
		},
	}
}
