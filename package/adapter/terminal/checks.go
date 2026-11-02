package terminal

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/ronz204/deltas-ai/package/domain/judgment"
	"github.com/ronz204/deltas-ai/package/usecase/verifier"
	"github.com/spf13/cobra"
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
			for _, slice := range rep.Verified {
				fmt.Fprintf(cmd.OutOrStdout(), "%s is verified\n", slice)
			}
			if count[judgment.Violated] > 0 {
				return ErrFound
			}
			return foundIn(rep.Diags)
		},
	}
}

func printResults(w io.Writer, results []verifier.Result) map[judgment.Verdict]int {
	count := map[judgment.Verdict]int{}
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
	fmt.Fprintf(w, "%d holds, %d violated, %d unverified\n", count[judgment.Holds], count[judgment.Violated], count[judgment.Unverified])
	return count
}
