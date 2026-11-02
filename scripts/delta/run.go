package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/ronz204/deltas-ai/source/adapter/cmd"
	"github.com/ronz204/deltas-ai/source/domain/budget"
	"github.com/ronz204/deltas-ai/source/domain/conffile"
	"github.com/ronz204/deltas-ai/source/gateway/bundle"
	"github.com/ronz204/deltas-ai/source/gateway/gitcli"
	"github.com/ronz204/deltas-ai/source/gateway/ledger"
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
	conf := loadConf(root)
	vault := vaults.New(root).Matching(conf.TestFiles)
	timeout := commandTimeout
	if conf.Timeout > 0 {
		timeout = conf.Timeout
	}
	programs := runner.New(root, timeout, errOut)
	git := gitcli.New(root)
	cache, _ := os.UserCacheDir() // without one, checks are not remembered and Stop asks for them again
	checks := ledger.New(cache, root, git)
	report := status.New(vault, conf.Scopes).WithDrift(checks, vault).WithDocs(vault)
	installer := setup.New(vault, git, bundle.Bundle{}, programs, conf)
	checker := verify.New(vault, vault, programs, checks, conf.Scopes)
	app := cmd.New(cmd.Deps{
		Author:    author.New(vault, bundle.Bundle{}),
		Checker:   checker,
		Report:    report,
		Guard:     guard.New(root, report, installer, checks, checker),
		Installer: installer,
		Meter:     measure.New(vault, conf.Budget),
		Version:   version,
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

// loadConf reads deltas.config. Without one it falls back to the budget ceilings of the older .claude/budget.conf,
// and a file that cannot be read leaves the defaults and says so.
func loadConf(root string) conffile.Conf {
	vault := vaults.New(root)
	text, err := vault.ReadFile(context.Background(), conffile.Name)
	switch {
	case err == nil:
		return conffile.Parse(string(text))
	case !errors.Is(err, fs.ErrNotExist):
		c := conffile.Parse("")
		c.Problems = append(c.Problems, fmt.Sprintf("%s could not be read: %v", conffile.Name, err))
		return c
	}
	c := conffile.Parse("")
	if legacy, err := vault.ReadFile(context.Background(), ".claude/budget.conf"); err == nil {
		c.Budget = budget.ParseCeilings(string(legacy))
	}
	return c
}
