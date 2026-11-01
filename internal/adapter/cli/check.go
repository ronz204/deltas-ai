package cli

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ronz204/deltas-ai/internal/domain/proof"
	"github.com/ronz204/deltas-ai/internal/usecase/verify"
)

func checkCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "check [slice]",
		Short: "Run the command of every test claim and say whether it holds",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rep, err := d.Checker.Check(cmd.Context(), optional(args))
			if err != nil {
				return err
			}
			printDiags(cmd.ErrOrStderr(), rep.Diags)
			count := printResults(cmd.OutOrStdout(), rep.Results)
			if count[proof.Violated] > 0 {
				return ErrFound
			}
			return foundIn(rep.Diags)
		},
	}
}

// printResults prints one row per claim and a summary line, and returns how many claims got
// each verdict.
func printResults(w io.Writer, results []verify.Result) map[proof.Verdict]int {
	count := map[proof.Verdict]int{}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for _, r := range results {
		count[r.Outcome.Verdict]++
		row := fmt.Sprintf("%s\t%s\t%s", r.Slice, r.ID, r.Outcome.Verdict)
		if r.Outcome.Reason != "" {
			row += "\t" + r.Outcome.Reason
		}
		fmt.Fprintln(tw, row)
	}
	tw.Flush()
	fmt.Fprintf(w, "%d holds, %d violated, %d unverified\n", count[proof.Holds], count[proof.Violated], count[proof.Unverified])
	return count
}
