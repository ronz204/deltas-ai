package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func installCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Install the git hook that checks commit messages",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			steps, err := d.Installer.Run(cmd.Context())
			if err != nil {
				return err
			}
			for _, s := range steps {
				fmt.Fprintln(cmd.OutOrStdout(), s.What)
			}
			return nil
		},
	}
}
