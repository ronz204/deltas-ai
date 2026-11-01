// Package cli is the command line: it turns arguments into calls to the use cases, and what they
// return into text and an exit status. It holds no rules of its own.
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/ronz204/deltas-ai/internal/domain/diag"
	"github.com/ronz204/deltas-ai/internal/usecase/author"
	"github.com/ronz204/deltas-ai/internal/usecase/status"
	"github.com/ronz204/deltas-ai/internal/usecase/verify"
)

// ErrFound is what a command returns when it ran fine and found problems: the caller exits 1
// without printing anything more, because the findings were already printed.
var ErrFound = errors.New("problems found")

// Deps are the use cases the commands call.
type Deps struct {
	Author  *author.Author
	Checker *verify.Checker
	Report  *status.Reporter
}

// New builds the delta command and its subcommands. Findings go to stderr and results to stdout.
func New(d Deps) *cobra.Command {
	root := &cobra.Command{
		Use:           "delta",
		Short:         "Keep a verifiable contract between a project and its code",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(specCmd(d), linkCmd(d), lintCmd(d), checkCmd(d), statusCmd(d), commitMsgCmd())
	return root
}

// optional returns the single optional argument, or "" when it was left out.
func optional(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func printDiags(w io.Writer, ds []diag.Diagnostic) {
	for _, d := range ds {
		fmt.Fprintln(w, d)
	}
}

// foundIn is ErrFound when ds holds an error; warnings alone do not fail a run.
func foundIn(ds []diag.Diagnostic) error {
	for _, d := range ds {
		if !d.Warn {
			return ErrFound
		}
	}
	return nil
}
