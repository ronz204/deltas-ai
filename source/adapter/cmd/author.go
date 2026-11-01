package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func specCmd(d Deps) *cobra.Command {
	spec := &cobra.Command{Use: "spec", Short: "Create, check and write specs"}
	spec.AddCommand(specNewCmd(d), specLintCmd(d), specWriteCmd(d))
	return spec
}

func specNewCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "new <slice>",
		Short: "Create deltas/<slice>.spec.md from the template, as a draft",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := d.Author.Create(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "created", path)
			return nil
		},
	}
}

func specLintCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "lint [slice]",
		Short: "Check the structure of a spec, or of all of them",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ds, err := d.Checker.SpecLint(cmd.Context(), optional(args))
			if err != nil {
				return err
			}
			printDiags(cmd.ErrOrStderr(), ds)
			return foundIn(ds)
		},
	}
}

func specWriteCmd(d Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "write <slice>",
		Short: "Validate a spec read from stdin and save it as deltas/<slice>.spec.md",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}
			ds, err := d.Author.Write(cmd.Context(), args[0], content)
			if err != nil {
				return err
			}
			printDiags(cmd.ErrOrStderr(), ds)
			if foundIn(ds) != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "nothing was written")
				return ErrFound
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote the spec of %s\n", args[0])
			return nil
		},
	}
}
