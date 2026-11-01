package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func installCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Install delta in this project: git hook, harness files and Claude Code settings",
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
