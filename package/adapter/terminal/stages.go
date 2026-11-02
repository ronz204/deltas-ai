package terminal

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
			if err := moved(cmd, args[0], "built", d.Author.Built); err != nil {
				return err
			}
			printFootprint(cmd, d, args[0])
			return nil
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
				if err := d.Tracer.Begin(cmd.Context(), args[0]); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "warning: the footprint of the build cannot be followed:", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s is ready to build\n", args[0])
			}
			return err
		},
	}
}

func printFootprint(cmd *cobra.Command, d Deps, slice string) {
	files, recorded, err := d.Tracer.Finish(cmd.Context(), slice)
	switch {
	case err != nil:
		fmt.Fprintln(cmd.ErrOrStderr(), "warning: the footprint was not updated:", err)
	case !recorded:
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: no snapshot from delta ready %s; the footprint was not updated\n", slice)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "%s footprint: %d files\n", slice, len(files))
		for _, f := range files {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", f)
		}
		if len(files) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "A person can remove what is not %s's from deltas/specs/%s.files.\n", slice, slice)
		}
	}
}

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
