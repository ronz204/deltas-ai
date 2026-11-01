// Package setup installs delta into a project without overriding what the project already has.
package setup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

const (
	hooksDir = ".githooks"
	hookName = "commit-msg"
	hookPath = hooksDir + "/" + hookName
)

type Files interface {
	ReadFile(ctx context.Context, rel string) ([]byte, error)
	WriteFile(ctx context.Context, rel string, content []byte, perm fs.FileMode) error
}

type Git interface {
	HooksPath(ctx context.Context) (string, error)
	SetHooksPath(ctx context.Context, path string) error
}

type Hooks interface {
	GitHook(name string) ([]byte, error)
}

// Binary finds a program on the PATH and says whether it is the one running.
type Binary interface {
	Locate(name string) (path string, self bool)
}

type Installer struct {
	files Files
	git   Git
	hooks Hooks
	bin   Binary
}

func New(files Files, git Git, hooks Hooks, bin Binary) *Installer {
	return &Installer{files: files, git: git, hooks: hooks, bin: bin}
}

// Check lists what is wrong with how delta is installed here, one sentence each; an empty list
// means all is well. It never fails: what it cannot read is reported as a problem too.
func (i *Installer) Check(ctx context.Context) []string {
	var problems []string
	current, err := i.git.HooksPath(ctx)
	if err != nil {
		return []string{"could not read git's configuration: " + err.Error()}
	}
	_, err = i.files.ReadFile(ctx, hookPath)
	missing := errors.Is(err, fs.ErrNotExist)
	pointed := strings.TrimPrefix(current, "./") == hooksDir
	switch {
	case current == "" && missing:
		problems = append(problems, "git does not check commit messages yet: run delta init")
	case current == "":
		problems = append(problems, "core.hooksPath is not set, so git ignores "+hookPath+": run delta init")
	case pointed && missing:
		problems = append(problems, hookPath+" is missing: run delta init")
	case !pointed:
		problems = append(problems, fmt.Sprintf("core.hooksPath is %s, so %s does not run on its own", current, hookPath))
	}
	switch path, self := i.bin.Locate("delta"); {
	case path == "":
		problems = append(problems, "delta is not on the PATH, so the git hook cannot run it")
	case !self:
		problems = append(problems, fmt.Sprintf("the delta on the PATH (%s) is not this program, so the git hook would run it instead", path))
	}
	return problems
}

// Step says what was done, or why something was left as it was.
type Step struct {
	Done bool
	What string
}

// Run installs the commit-msg git hook and points core.hooksPath at it. It is safe to run again,
// and it never replaces a different hook or a core.hooksPath that is already set to something else.
func (i *Installer) Run(ctx context.Context) ([]Step, error) {
	want, err := i.hooks.GitHook(hookName)
	if err != nil {
		return nil, err
	}
	have, err := i.files.ReadFile(ctx, hookPath)
	missing := errors.Is(err, fs.ErrNotExist)
	if err != nil && !missing {
		return nil, err
	}
	pointed, err := i.pointHooksPath(ctx)
	if err != nil {
		return nil, err
	}

	var file Step
	switch {
	case missing:
		if err := i.files.WriteFile(ctx, hookPath, want, 0o755); err != nil {
			return nil, err
		}
		file = Step{true, "installed " + hookPath}
	case bytes.Equal(have, want):
		file = Step{false, hookPath + " is already installed"}
	default:
		file = Step{false, hookPath + " exists and differs: left as it is"}
	}
	return []Step{file, pointed}, nil
}

// pointHooksPath is done before the hook is written, so a directory that is not a git repository
// fails here and nothing is left behind.
func (i *Installer) pointHooksPath(ctx context.Context) (Step, error) {
	current, err := i.git.HooksPath(ctx)
	if err != nil {
		return Step{}, err
	}
	switch strings.TrimPrefix(current, "./") {
	case hooksDir:
		return Step{false, "core.hooksPath already points to " + hooksDir}, nil
	case "":
		if err := i.git.SetHooksPath(ctx, hooksDir); err != nil {
			return Step{}, err
		}
		return Step{true, "set core.hooksPath to " + hooksDir}, nil
	}
	return Step{false, fmt.Sprintf("core.hooksPath is %s: left as it is; call %s from there to check commit messages", current, hookPath)}, nil
}
