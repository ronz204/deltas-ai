package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ronz204/deltas-ai/internal/adapter/cli"
	"github.com/ronz204/deltas-ai/internal/gateway/bundle"
	"github.com/ronz204/deltas-ai/internal/gateway/files"
	"github.com/ronz204/deltas-ai/internal/gateway/runner"
	"github.com/ronz204/deltas-ai/internal/usecase/author"
	"github.com/ronz204/deltas-ai/internal/usecase/status"
	"github.com/ronz204/deltas-ai/internal/usecase/verify"
)

// commandTimeout bounds each claim's command, so one that hangs cannot hang a check.
const commandTimeout = 10 * time.Minute

// run wires the layers together and executes one command line. It returns the exit status:
// 0 all good, 1 problems were found, 2 the command could not run.
func run(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer, cwd string) int {
	root := files.FindRoot(cwd)
	store := files.New(root)
	cmd := cli.New(cli.Deps{
		Author:  author.New(store, bundle.Bundle{}),
		Checker: verify.New(store, store, runner.New(root, commandTimeout, errOut)),
		Report:  status.New(store),
	})
	cmd.SetArgs(args)
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	switch err := cmd.ExecuteContext(ctx); {
	case err == nil:
		return 0
	case errors.Is(err, cli.ErrFound):
		return 1
	default:
		fmt.Fprintln(errOut, "delta:", err)
		return 2
	}
}
