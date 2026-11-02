package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/ronz204/deltas-ai/package/domain/conffile"
	"github.com/ronz204/deltas-ai/package/gateway/diskbase"
	"github.com/ronz204/deltas-ai/package/gateway/embedded"
	"github.com/ronz204/deltas-ai/package/gateway/executor"
	"github.com/ronz204/deltas-ai/package/gateway/gitshell"
	"github.com/ronz204/deltas-ai/package/gateway/recorder"
	"github.com/ronz204/deltas-ai/package/usecase/deployer"
	"github.com/ronz204/deltas-ai/package/usecase/drafting"
	"github.com/ronz204/deltas-ai/package/usecase/sentinel"
	"github.com/ronz204/deltas-ai/package/usecase/tracking"
	"github.com/ronz204/deltas-ai/package/usecase/verifier"
	"github.com/ronz204/deltas-ai/package/usecase/weighing"
	"github.com/ronz204/deltas-ai/source/adapter/cmd"
)

// commandTimeout bounds each claim's command, so one that hangs cannot hang a check.
const commandTimeout = 10 * time.Minute

// run executes one command line and returns the exit status: 0 ok, 1 problems found, 2 error.
// delta hook answers to Claude Code instead: it exits 2 only to block, and 1 when delta itself failed.
func run(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer, cwd string) int {
	root := diskbase.FindRoot(cwd)
	conf := loadConf(root)
	vault := diskbase.New(root).Matching(conf.Files.Match).Skipping(conf.Files.Skip)
	programs := executor.New(root, commandTimeout, errOut)
	git := gitshell.New(root)
	cache, _ := os.UserCacheDir() // without one, checks are not remembered and Stop asks for them again
	checks := recorder.New(cache, root, git)
	report := tracking.New(vault).WithDrift(checks, vault).WithDocs(vault)
	installer := deployer.New(vault, embedded.Bundle{}, programs, conf)
	checker := verifier.New(vault, vault, programs, checks)
	app := cmd.New(cmd.Deps{
		Author:    drafting.New(vault, embedded.Bundle{}),
		Checker:   checker,
		Report:    report,
		Guard:     sentinel.New(root, report, installer, checks, checker),
		Installer: installer,
		Meter:     weighing.New(vault, conf.Budget),
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

// loadConf reads budget.conf and testing.conf; one that is missing leaves its defaults, and one that cannot be
// read leaves them too and says so.
func loadConf(root string) conffile.Conf {
	vault := diskbase.New(root)
	var problems []string
	read := func(name string) string {
		text, err := vault.ReadFile(context.Background(), name)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			problems = append(problems, fmt.Sprintf("%s could not be read: %v", name, err))
		}
		return string(text)
	}
	c := conffile.Parse(read(conffile.BudgetName), read(conffile.TestingName))
	c.Problems = append(problems, c.Problems...)
	return c
}
