package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func statusCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show every slice: its state and how its claims are verified",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rows, err := d.Report.Report(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(rows) == 0 {
				fmt.Fprintln(out, "no specs yet; start with: delta spec new <slice>")
				return nil
			}
			tw := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "SLICE\tSTATE\tCLAIMS\tTEST\tMANUAL\tNONE\tISSUES")
			for _, r := range rows {
				state := string(r.Status)
				if state == "" {
					state = "-" // the frontmatter could not be read
				}
				fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%d\t%d\n", r.Slice, state, r.Claims, r.Test, r.Manual, r.None, r.Issues)
			}
			return tw.Flush()
		},
	}
}
