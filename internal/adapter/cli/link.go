package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func linkCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "link <slice> <claim-id> <command>",
		Short: "Verify a claim with a command: its tag becomes test: <command>",
		Long:  "Verify a claim with a command: its tag becomes test: <command>. Quote the command so it is one argument.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			slice, id, command := args[0], args[1], args[2]
			if err := d.Author.Link(cmd.Context(), slice, id, command); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "linked %s of %s\n", id, slice)
			return nil
		},
	}
}
