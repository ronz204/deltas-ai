package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ronz204/deltas-ai/sources/adapter/cmd"
	"github.com/ronz204/deltas-ai/sources/gateway/bundle"
	"github.com/ronz204/deltas-ai/sources/gateway/files"
	"github.com/ronz204/deltas-ai/sources/gateway/runner"
	"github.com/ronz204/deltas-ai/sources/usecase/author"
	"github.com/ronz204/deltas-ai/sources/usecase/status"
	"github.com/ronz204/deltas-ai/sources/usecase/verify"
)

// commandTimeout bounds each claim's command, so one that hangs cannot hang a check.
const commandTimeout = 10 * time.Minute

// run executes one command line and returns the exit status: 0 ok, 1 problems found, 2 error.
func run(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer, cwd string) int {
	root := files.FindRoot(cwd)
	store := files.New(root)
	app := cmd.New(cmd.Deps{
		Author:  author.New(store, bundle.Bundle{}),
		Checker: verify.New(store, store, runner.New(root, commandTimeout, errOut)),
		Report:  status.New(store),
	})
	app.SetArgs(args)
	app.SetIn(in)
	app.SetOut(out)
	app.SetErr(errOut)
	switch err := app.ExecuteContext(ctx); {
	case err == nil:
		return 0
	case errors.Is(err, cmd.ErrFound):
		return 1
	default:
		fmt.Fprintln(errOut, "delta:", err)
		return 2
	}
}
