package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ronz204/deltas-ai/source/adapter/hook"
)

// guardsCmd is what Claude Code runs for every hook event: the event arrives as JSON on stdin.
func guardsCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Answer a Claude Code hook event read from stdin",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if code := hook.Run(cmd.Context(), d.Guard, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()); code != 0 {
				return &ExitError{Code: code}
			}
			return nil
		},
	}
}
