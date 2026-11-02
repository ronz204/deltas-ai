package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ronz204/deltas-ai/package/adapter/terminal"
	"github.com/ronz204/deltas-ai/package/gateway/diskbase"
	"github.com/ronz204/deltas-ai/package/gateway/embedded"
	"github.com/ronz204/deltas-ai/package/gateway/executor"
	"github.com/ronz204/deltas-ai/package/gateway/gitshell"
	"github.com/ronz204/deltas-ai/package/gateway/recorder"
	"github.com/ronz204/deltas-ai/package/usecase/deployer"
	"github.com/ronz204/deltas-ai/package/usecase/drafting"
	"github.com/ronz204/deltas-ai/package/usecase/sentinel"
	"github.com/ronz204/deltas-ai/package/usecase/tracing"
	"github.com/ronz204/deltas-ai/package/usecase/tracking"
	"github.com/ronz204/deltas-ai/package/usecase/verifier"
	"github.com/ronz204/deltas-ai/package/usecase/weighing"
	"github.com/spf13/cobra"
)

const commandTimeout = 10 * time.Minute

func run(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer, cwd string) int {
	app := newApp(diskbase.FindRoot(cwd), errOut)
	app.SetArgs(args)
	app.SetIn(in)
	app.SetOut(out)
	app.SetErr(errOut)
	return exitCode(app.ExecuteContext(ctx), errOut)
}

func newApp(root string, errOut io.Writer) *cobra.Command {
	conf := loadConf(root)
	vault := diskbase.New(root).Matching(conf.Files.Match).Skipping(conf.Files.Skip)
	programs := executor.New(root, commandTimeout, errOut)
	cache, _ := os.UserCacheDir()
	checks := recorder.New(cache, root, gitshell.New(root))
	report := tracking.New(vault).WithDrift(checks).WithFootprints(vault).WithDocs(vault)
	installer := deployer.New(vault, embedded.Bundle{}, programs, conf)
	checker := verifier.New(vault, vault, programs, checks).WithFootprints(vault)
	return terminal.New(terminal.Deps{
		Author:    drafting.New(vault, embedded.Bundle{}),
		Checker:   checker,
		Report:    report,
		Tracer:    tracing.New(checks, vault),
		Guard:     sentinel.New(root, report, installer, checks, checker),
		Installer: installer,
		Meter:     weighing.New(vault, conf.Budget),
		Version:   version,
	})
}

func exitCode(err error, errOut io.Writer) int {
	var exit *terminal.ExitError
	switch {
	case err == nil:
		return 0
	case errors.Is(err, terminal.ErrFound):
		return 1
	case errors.As(err, &exit):
		return exit.Code
	}
	fmt.Fprintln(errOut, "delta:", err)
	return 2
}
