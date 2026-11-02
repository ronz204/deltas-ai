package terminal

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/ronz204/deltas-ai/package/domain/judgment"
	"github.com/ronz204/deltas-ai/package/usecase/verifier"
	"github.com/spf13/cobra"
)

func checkCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "check [slice...]",
		Short: "Run the command of every test claim and say whether it holds",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rep, err := checkAll(cmd.Context(), d.Checker, args)
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

func checkAll(ctx context.Context, c *verifier.Checker, slices []string) (verifier.Report, error) {
	if len(slices) == 0 {
		return c.Check(ctx, "")
	}
	var all verifier.Report
	for _, slice := range slices {
		rep, err := c.Check(ctx, slice)
		if err != nil {
			return all, err
		}
		all.Diags = append(all.Diags, rep.Diags...)
		all.Results = append(all.Results, rep.Results...)
		all.Verified = append(all.Verified, rep.Verified...)
	}
	return all, nil
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
