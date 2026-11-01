package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ronz204/deltas-ai/internal/domain/messages"
)

// commitMsgCmd is what a git commit-msg hook runs: git passes the path of the message file.
func commitMsgCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "commit-msg <file>",
		Short: "Check a commit message file against the Conventional Commits rules",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			ds := messages.Validate(args[0], string(message))
			printDiags(cmd.ErrOrStderr(), ds)
			return foundIn(ds)
		},
	}
}
