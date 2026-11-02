package terminal

import (
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/ronz204/deltas-ai/package/domain/findings"
	"github.com/ronz204/deltas-ai/package/usecase/deployer"
	"github.com/ronz204/deltas-ai/package/usecase/drafting"
	"github.com/ronz204/deltas-ai/package/usecase/sentinel"
	"github.com/ronz204/deltas-ai/package/usecase/tracing"
	"github.com/ronz204/deltas-ai/package/usecase/tracking"
	"github.com/ronz204/deltas-ai/package/usecase/verifier"
	"github.com/ronz204/deltas-ai/package/usecase/weighing"
	"github.com/spf13/cobra"
)

var ErrFound = errors.New("problems found")

type ExitError struct{ Code int }

func (e *ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Code) }

type Deps struct {
	Author    *drafting.Author
	Checker   *verifier.Checker
	Report    *tracking.Reporter
	Tracer    *tracing.Tracer
	Guard     *sentinel.Guard
	Installer *deployer.Installer
	Meter     *weighing.Meter
	Version   string
}

func New(d Deps) *cobra.Command {
	root := &cobra.Command{
		Use:           "delta",
		Short:         "Keep a verifiable contract between a project and its code",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(specCmd(d), linkCmd(d), approveCmd(d), builtCmd(d), readyCmd(d), lintCmd(d), checkCmd(d), statusCmd(d), staleCmd(d), installCmd(d), updateCmd(d), guardsCmd(d), budgetCmd(d), versionCmd(d))
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

func foundIn(ds []findings.Diagnostic) error {
	if slices.ContainsFunc(ds, func(d findings.Diagnostic) bool { return !d.Warn }) {
		return ErrFound
	}
	return nil
}
