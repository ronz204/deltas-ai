package terminal

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
