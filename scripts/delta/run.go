package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ronz204/deltas-ai/source/adapter/cmd"
	"github.com/ronz204/deltas-ai/source/gateway/bundle"
	"github.com/ronz204/deltas-ai/source/gateway/gitcli"
	"github.com/ronz204/deltas-ai/source/gateway/runner"
	"github.com/ronz204/deltas-ai/source/gateway/vaults"
	"github.com/ronz204/deltas-ai/source/usecase/author"
	"github.com/ronz204/deltas-ai/source/usecase/guard"
	"github.com/ronz204/deltas-ai/source/usecase/measure"
	"github.com/ronz204/deltas-ai/source/usecase/setup"
	"github.com/ronz204/deltas-ai/source/usecase/status"
	"github.com/ronz204/deltas-ai/source/usecase/verify"
)

// commandTimeout bounds each claim's command, so one that hangs cannot hang a check.
const commandTimeout = 10 * time.Minute

// run executes one command line and returns the exit status: 0 ok, 1 problems found, 2 error.
// delta hook answers to Claude Code instead: it exits 2 only to block, and 1 when delta itself failed.
func run(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer, cwd string) int {
	root := vaults.FindRoot(cwd)
	vault := vaults.New(root)
	report := status.New(vault)
	programs := runner.New(root, commandTimeout, errOut)
	installer := setup.New(vault, gitcli.New(root), bundle.Bundle{}, programs)
	app := cmd.New(cmd.Deps{
		Author:    author.New(vault, bundle.Bundle{}),
		Checker:   verify.New(vault, vault, programs),
		Report:    report,
		Guard:     guard.New(root, report, installer),
		Installer: installer,
		Meter:     measure.New(vault),
	})
	app.SetArgs(args)
	app.SetIn(in)
	app.SetOut(out)
	app.SetErr(errOut)
	var exit *cmd.ExitError
	switch err := app.ExecuteContext(ctx); {
	case err == nil:
		return 0
	case errors.Is(err, cmd.ErrFound):
		return 1
	case errors.As(err, &exit):
		return exit.Code
	default:
		fmt.Fprintln(errOut, "delta:", err)
		return 2
	}
}
