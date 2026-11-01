package cli

import "github.com/spf13/cobra"

func lintCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "lint",
		Short: "Check every spec and cross it with the project's tests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ds, err := d.Checker.Lint(cmd.Context())
			if err != nil {
				return err
			}
			printDiags(cmd.ErrOrStderr(), ds)
			return foundIn(ds)
		},
	}
}
