package deployer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/conffile"
)

type Files interface {
	ReadFile(ctx context.Context, rel string) ([]byte, error)
	WriteFile(ctx context.Context, rel string, content []byte) error
	RemoveFile(ctx context.Context, rel string) error
	Name() string
}

type Asset struct {
	Path    string
	Content []byte
	Managed bool
}

type Harness interface {
	Assets(project string) ([]Asset, error)
	Settings() ([]byte, error)
}

type Binary interface {
	Locate(name string) (path string, self bool)
}

type Installer struct {
	files   Files
	harness Harness
	bin     Binary
	conf    conffile.Conf
}

func New(files Files, harness Harness, bin Binary, conf conffile.Conf) *Installer {
	return &Installer{files: files, harness: harness, bin: bin, conf: conf}
}

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

type Step struct {
	Done bool
	What string
}

var skeleton = []Asset{{Path: "deltas/specs/.gitkeep"}}

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

func (i *Installer) assets() ([]Asset, error) { return i.harness.Assets(i.files.Name()) }

func (i *Installer) copy(ctx context.Context, assets []Asset) ([]Step, error) {
	var written, same int
	var kept []string
	for _, a := range assets {
		have, err := i.files.ReadFile(ctx, a.Path)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			if err := i.files.WriteFile(ctx, a.Path, a.Content); err != nil {
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
