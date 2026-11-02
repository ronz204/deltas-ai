package terminal

import (
	"fmt"

	"github.com/spf13/cobra"
)

func staleCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "stale",
		Short: "List the verified slices whose code changed since they were verified",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			drifted, err := d.Report.Drifts(cmd.Context())
			if err != nil {
				return err
			}
			for _, s := range drifted {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %d files changed since it was verified\n", s.Slice, len(s.Files))
				for _, f := range s.Files {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", f)
				}
			}
			if len(drifted) > 0 {
				return ErrFound
			}
			return nil
		},
	}
}
