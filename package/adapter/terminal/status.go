package terminal

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/ronz204/deltas-ai/package/usecase/tracking"
	"github.com/spf13/cobra"
)

func statusCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show every slice and whether delta is set up here",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rows, err := d.Report.Report(cmd.Context())
			if err != nil {
				return err
			}
			if err := printRows(cmd.OutOrStdout(), rows); err != nil {
				return err
			}
			if row, err := d.Report.Docs(cmd.Context()); err == nil {
				printDocs(cmd.OutOrStdout(), row)
			}
			for _, problem := range d.Installer.Check(cmd.Context()) {
				fmt.Fprintln(cmd.ErrOrStderr(), "warning:", problem)
			}
			return nil
		},
	}
}

func printRows(out io.Writer, rows []tracking.Row) error {
	if len(rows) == 0 {
		fmt.Fprintln(out, "no specs yet; start with: delta spec new <slice>")
		return nil
	}
	tw := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "SLICE\tSTATE\tCLAIMS\tTEST\tMANUAL\tNONE\tISSUES")
	for _, r := range rows {
		state := string(r.Status)
		if state == "" {
			state = "-"
		}
		if n := len(r.Changed); n > 0 {
			state = fmt.Sprintf("%s (%d files changed)", state, n)
		}
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%d\t%d\n", r.Slice, state, r.Claims, r.Test, r.Manual, r.None, r.Issues)
	}
	return tw.Flush()
}

func printDocs(out io.Writer, row tracking.DocsRow) {
	switch {
	case row.Total == 0:
	case row.Next == "":
		fmt.Fprintf(out, "docs: all %d written\n", row.Total)
	default:
		fmt.Fprintf(out, "docs: %d of %d written; next: /scriber %s\n", row.Filled, row.Total, row.Next)
	}
}
