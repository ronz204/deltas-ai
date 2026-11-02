package terminal

import (
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"

	"github.com/ronz204/deltas-ai/package/domain/budget"
	"github.com/spf13/cobra"
)

func budgetCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "budget",
		Short: "Weigh the context loaded on every turn against its ceilings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			report, err := d.Meter.Always(cmd.Context())
			if err != nil {
				return err
			}
			printBudget(cmd.OutOrStdout(), report)
			problems := report.Findings()
			printDiags(cmd.ErrOrStderr(), problems)
			return foundIn(problems)
		},
	}
}

func printBudget(w io.Writer, report budget.Report) {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "KIND\tFILE\tSIZE\tLIMIT\tTOKENS")
	for _, it := range report.Items {
		fmt.Fprintf(tw, "%s\t%s\t%d %s\t%s\t%d\n", it.Kind, it.Path, it.Size, it.Unit, limitText(it.Limit), it.Tokens)
	}
	tw.Flush()
	fmt.Fprintf(w, "total: about %d tokens of %s\n", report.Total, limitText(report.Limit))
}

func limitText(limit int) string {
	if limit == 0 {
		return "-"
	}
	return strconv.Itoa(limit)
}
