package terminal

import (
	"fmt"

	"github.com/spf13/cobra"
)

func staleCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "stale",
		Short: "Mark verified slices stale when the files they depend on changed, and list the stale ones",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			drifted, err := d.Report.Stale(cmd.Context())
			if err != nil {
				return err
			}
			for _, s := range drifted {
				verb := "is stale"
				if s.Marked {
					verb = "became stale"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s: %d files changed since it was verified\n", s.Slice, verb, len(s.Files))
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
