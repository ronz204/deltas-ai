package terminal

import (
	"github.com/ronz204/deltas-ai/package/adapter/listener"
	"github.com/spf13/cobra"
)

func guardsCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Answer a Claude Code hook event read from stdin",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if code := listener.Run(cmd.Context(), d.Guard, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()); code != 0 {
				return &ExitError{Code: code}
			}
			return nil
		},
	}
}
