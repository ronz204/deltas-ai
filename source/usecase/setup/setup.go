// Package setup installs delta into a project without overriding what the project already has.
package setup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/conffile"
)

const (
	hooksDir = ".githooks"
	hookName = "commit-msg"
	hookPath = hooksDir + "/" + hookName
)

type Files interface {
	ReadFile(ctx context.Context, rel string) ([]byte, error)
	WriteFile(ctx context.Context, rel string, content []byte, perm fs.FileMode) error
	RemoveFile(ctx context.Context, rel string) error // also the folders it leaves empty
	Name() string                                     // what the project is called
}

type Git interface {
	HooksPath(ctx context.Context) (string, error)
	SetHooksPath(ctx context.Context, path string) error
}

// Asset is a file the harness puts in a project: where, what and with which permissions. A managed asset is
// kept up to date by delta update; any other is installed once and then belongs to the project.
type Asset struct {
	Path    string
	Content []byte
	Mode    fs.FileMode
	Managed bool // kept up to date by delta update; any other asset is installed once and is then the project's
	Base    bool // part of the base preset: left out when deltas.config says base=off
}

// Harness is what delta ships to install.
type Harness interface {
	GitHook(name string) ([]byte, error)
	Assets(project string) ([]Asset, error) // skills, rules, deltas.config and CLAUDE.md
	Settings() ([]byte, error)              // the part of .claude/settings.json that delta needs
}

// Binary finds a program on the PATH and says whether it is the one running.
type Binary interface {
	Locate(name string) (path string, self bool)
}

type Installer struct {
	files   Files
	git     Git
	harness Harness
	bin     Binary
	conf    conffile.Conf
}

// New builds the installer of a project whose deltas.config says conf.
func New(files Files, git Git, harness Harness, bin Binary, conf conffile.Conf) *Installer {
	return &Installer{files: files, git: git, harness: harness, bin: bin, conf: conf}
}

// Check lists what is wrong with how delta is installed here, one sentence each; an empty list
// means all is well. It never fails: what it cannot read is reported as a problem too.
func (i *Installer) Check(ctx context.Context) []string {
	problems := append([]string(nil), i.conf.Problems...)
	for _, tool := range i.conf.Tools {
		if path, _ := i.bin.Locate(tool); path == "" {
			problems = append(problems, fmt.Sprintf("%s is required by deltas.config and is not on the PATH", tool))
		}
	}
	current, err := i.git.HooksPath(ctx)
	if err != nil {
		return append(problems, "could not read git's configuration: "+err.Error())
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
	if pending, err := i.Pending(ctx); err != nil {
		problems = append(problems, "could not compare the installed harness with this program's: "+err.Error())
	} else if len(pending) > 0 {
		problems = append(problems, fmt.Sprintf("the installed harness is out of date (%d %s would change): run delta update", len(pending), noun(len(pending))))
	}
	return problems
}

// Step says what was done, or why something was left as it was.
type Step struct {
	Done bool
	What string
}

// skeleton is what every project gets before the harness: the folder that holds the contracts.
var skeleton = []Asset{{Path: "deltas/specs/.gitkeep", Mode: 0o644}}

// Run installs the commit-msg git hook and points core.hooksPath at it, then the files of the harness and
// the settings Claude Code needs. It is safe to run again, and it never replaces a file that differs from
// the harness's, a different hook, or a core.hooksPath that is already set to something else.
func (i *Installer) Run(ctx context.Context) ([]Step, error) {
	want, err := i.harness.GitHook(hookName)
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
	steps := []Step{file, pointed}

	assets, err := i.assets()
	if err != nil {
		return nil, err
	}
	copied, err := i.copy(ctx, append(append([]Asset(nil), skeleton...), assets...))
	if err != nil {
		return nil, err
	}
	settings, err := i.installSettings(ctx)
	if err != nil {
		return nil, err
	}
	recorded, err := i.record(ctx)
	if err != nil {
		return nil, err
	}
	return append(append(append(steps, copied...), recorded...), settings), nil
}

// assets is what the harness ships to this project: all of it, but the base preset when deltas.config turns it off.
func (i *Installer) assets() ([]Asset, error) {
	all, err := i.harness.Assets(i.files.Name())
	if err != nil || !i.conf.NoBase {
		return all, err
	}
	var out []Asset
	for _, a := range all {
		if !a.Base {
			out = append(out, a)
		}
	}
	return out, nil
}

// copy writes the assets that are not there yet. One that exists and differs is a choice of the project's.
func (i *Installer) copy(ctx context.Context, assets []Asset) ([]Step, error) {
	var written, same int
	var kept []string
	for _, a := range assets {
		have, err := i.files.ReadFile(ctx, a.Path)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			if err := i.files.WriteFile(ctx, a.Path, a.Content, a.Mode); err != nil {
				return nil, err
			}
			written++
		case err != nil:
			return nil, err
		case bytes.Equal(have, a.Content):
			same++
		default:
			kept = append(kept, a.Path)
		}
	}
	var steps []Step
	if written > 0 {
		steps = append(steps, Step{true, fmt.Sprintf("installed %d %s from the harness", written, noun(written))})
	}
	if len(kept) > 0 {
		steps = append(steps, Step{false, "left as they are, since they differ from the harness: " + strings.Join(kept, ", ")})
	}
	if written == 0 && len(kept) == 0 {
		steps = append(steps, Step{false, fmt.Sprintf("the %d %s of the harness are already installed", same, noun(same))})
	}
	return steps, nil
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

func noun(n int) string {
	if n == 1 {
		return "file"
	}
	return "files"
}
