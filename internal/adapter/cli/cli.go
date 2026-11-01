// Package cli turns arguments into calls to the use cases, and their results into text and an exit status.
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/ronz204/deltas-ai/internal/domain/findings"
	"github.com/ronz204/deltas-ai/internal/usecase/author"
	"github.com/ronz204/deltas-ai/internal/usecase/status"
	"github.com/ronz204/deltas-ai/internal/usecase/verify"
)

// ErrFound means a command ran fine and found problems, already printed: exit 1 with nothing more.
var ErrFound = errors.New("problems found")

type Deps struct {
	Author  *author.Author
	Checker *verify.Checker
	Report  *status.Reporter
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
	root.AddCommand(specCmd(d), linkCmd(d), lintCmd(d), checkCmd(d), statusCmd(d), commitMsgCmd())
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
