// Package measure weighs the context a project loads on every turn.
package measure

import (
	"context"
	"errors"
	"io/fs"

	"github.com/ronz204/deltas-ai/source/domain/budget"
)

const confPath = ".claude/budget.conf"

type Files interface {
	Glob(ctx context.Context, patterns []string) ([]budget.File, error)
	ReadFile(ctx context.Context, rel string) ([]byte, error)
}

type Meter struct{ files Files }

func New(files Files) *Meter { return &Meter{files: files} }

// Always weighs what is loaded on every turn against the ceilings in .claude/budget.conf, or
// against the defaults when the project has no such file.
func (m *Meter) Always(ctx context.Context) (budget.Report, error) {
	conf, err := m.files.ReadFile(ctx, confPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return budget.Report{}, err
	}
	files, err := m.files.Glob(ctx, budget.Patterns)
	if err != nil {
		return budget.Report{}, err
	}
	return budget.Measure(files, budget.ParseCeilings(string(conf))), nil
}
