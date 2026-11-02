package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/ronz204/deltas-ai/package/domain/findings"
	"github.com/ronz204/deltas-ai/package/usecase/tracking"
	"github.com/spf13/cobra"
)

func approveCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "approve <slice>",
		Short: "Sign off a draft spec so Builder can build it (for a person to run)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return moved(cmd, args[0], "approved", d.Author.Approve)
		},
	}
}

func builtCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "built <slice>",
		Short: "Record that an approved slice is built; delta check then verifies it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return moved(cmd, args[0], "built", d.Author.Built)
		},
	}
}

func readyCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "ready <slice>",
		Short: "Succeed only if the slice has a clean spec and is approved, ready for Builder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := d.Report.Ready(cmd.Context(), args[0])
			var notReady *tracking.NotReady
			if errors.As(err, &notReady) {
				fmt.Fprintln(cmd.ErrOrStderr(), notReady.Reason)
				return ErrFound
			}
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "%s is ready to build\n", args[0])
			}
			return err
		},
	}
}

// moved runs one status change; a spec with problems is reported and left as it was.
func moved(cmd *cobra.Command, slice, to string, move func(context.Context, string) ([]findings.Diagnostic, error)) error {
	ds, err := move(cmd.Context(), slice)
	if err != nil {
		return err
	}
	if len(ds) > 0 {
		printDiags(cmd.ErrOrStderr(), ds)
		fmt.Fprintln(cmd.ErrOrStderr(), "nothing was changed")
		return ErrFound
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s is %s\n", slice, to)
	return nil
}
