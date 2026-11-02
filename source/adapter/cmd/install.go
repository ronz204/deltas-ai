package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func installCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Install delta in this project: harness files and Claude Code settings",
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

func updateCmd(d Deps) *cobra.Command {
	var check bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Bring the harness files up to date with this delta, keeping what the project changed",
		Long: "Installs the harness files that are missing, replaces those the project has not touched since delta installed " +
			"them, removes those this delta no longer ships, and leaves alone every file the project changed.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if check {
				pending, err := d.Installer.Pending(cmd.Context())
				if err != nil {
					return err
				}
				for _, line := range pending {
					fmt.Fprintln(cmd.OutOrStdout(), line)
				}
				if len(pending) > 0 {
					return ErrFound
				}
				return nil
			}
			steps, err := d.Installer.Update(cmd.Context())
			if err != nil {
				return err
			}
			for _, s := range steps {
				fmt.Fprintln(cmd.OutOrStdout(), s.What)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "write nothing; list what would change and exit 1 if anything would")
	return cmd
}
