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

type Files interface {
	ReadFile(ctx context.Context, rel string) ([]byte, error)
	WriteFile(ctx context.Context, rel string, content []byte, perm fs.FileMode) error
	RemoveFile(ctx context.Context, rel string) error // also the folders it leaves empty
	Name() string                                     // what the project is called
}

// Asset is a file the harness puts in a project: where, what and with which permissions. A managed asset is
// kept up to date by delta update; any other is installed once and then belongs to the project.
type Asset struct {
	Path    string
	Content []byte
	Mode    fs.FileMode
	Managed bool // kept up to date by delta update; any other asset is installed once and is then the project's
}

// Harness is what delta ships to install.
type Harness interface {
	Assets(project string) ([]Asset, error) // skills, rules, the config files and CLAUDE.md
	Settings() ([]byte, error)              // the part of .claude/settings.json that delta needs
}

// Binary finds a program on the PATH and says whether it is the one running.
type Binary interface {
	Locate(name string) (path string, self bool)
}

type Installer struct {
	files   Files
	harness Harness
	bin     Binary
	conf    conffile.Conf
}

// New builds the installer of a project whose config files say conf.
func New(files Files, harness Harness, bin Binary, conf conffile.Conf) *Installer {
	return &Installer{files: files, harness: harness, bin: bin, conf: conf}
}

// Check lists what is wrong with how delta is installed here, one sentence each; an empty list
// means all is well. It never fails: what it cannot read is reported as a problem too.
func (i *Installer) Check(ctx context.Context) []string {
	problems := append([]string(nil), i.conf.Problems...)
	switch path, self := i.bin.Locate("delta"); {
	case path == "":
		problems = append(problems, "delta is not on the PATH, so the hooks of Claude Code cannot run it")
	case !self:
		problems = append(problems, fmt.Sprintf("the delta on the PATH (%s) is not this program, so the hooks of Claude Code would run it instead", path))
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

// Run installs the files of the harness and the settings Claude Code needs. It is safe to run again, and it
// never replaces a file that differs from the harness's.
func (i *Installer) Run(ctx context.Context) ([]Step, error) {
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
	return append(append(copied, recorded...), settings), nil
}

// assets is what the harness ships to this project.
func (i *Installer) assets() ([]Asset, error) { return i.harness.Assets(i.files.Name()) }

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

func noun(n int) string {
	if n == 1 {
		return "file"
	}
	return "files"
}
