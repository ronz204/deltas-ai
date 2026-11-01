// Package cmd turns arguments into calls to the use cases, and their results into text and an exit status.
package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/ronz204/deltas-ai/source/domain/findings"
	"github.com/ronz204/deltas-ai/source/usecase/author"
	"github.com/ronz204/deltas-ai/source/usecase/guard"
	"github.com/ronz204/deltas-ai/source/usecase/measure"
	"github.com/ronz204/deltas-ai/source/usecase/setup"
	"github.com/ronz204/deltas-ai/source/usecase/status"
	"github.com/ronz204/deltas-ai/source/usecase/verify"
)

// ErrFound means a command ran fine and found problems, already printed: exit 1 with nothing more.
var ErrFound = errors.New("problems found")

// ExitError makes the program exit with Code; whatever it had to say is already printed.
type ExitError struct{ Code int }

func (e *ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Code) }

type Deps struct {
	Author    *author.Author
	Checker   *verify.Checker
	Report    *status.Reporter
	Guard     *guard.Guard
	Installer *setup.Installer
	Meter     *measure.Meter
}

// New builds the delta command. Findings go to stderr and results to stdout.
func New(d Deps) *cobra.Command {
	root := &cobra.Command{
		Use:           "delta",
		Short:         "Keep a verifiable contract between a project and its code",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(specCmd(d), linkCmd(d), lintCmd(d), checkCmd(d), statusCmd(d), commitMsgCmd(), installCmd(d), guardsCmd(d), budgetCmd(d))
	return root
}

func optional(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func printDiags(w io.Writer, ds []findings.Diagnostic) {
	for _, d := range ds {
		fmt.Fprintln(w, d)
	}
}

// foundIn is ErrFound when ds holds an error; warnings alone do not fail a run.
func foundIn(ds []findings.Diagnostic) error {
	for _, d := range ds {
		if !d.Warn {
			return ErrFound
		}
	}
	return nil
}
