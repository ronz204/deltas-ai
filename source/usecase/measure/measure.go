// Package measure weighs the context a project loads on every turn.
package measure

import (
	"context"

	"github.com/ronz204/deltas-ai/package/domain/budget"
)

type Files interface {
	Glob(ctx context.Context, patterns []string) ([]budget.File, error)
}

type Meter struct {
	files    Files
	ceilings budget.Ceilings
}

// New builds the meter of a project whose budget.conf sets these ceilings.
func New(files Files, ceilings budget.Ceilings) *Meter {
	return &Meter{files: files, ceilings: ceilings}
}

// Always weighs what is loaded on every turn against the ceilings.
func (m *Meter) Always(ctx context.Context) (budget.Report, error) {
	files, err := m.files.Glob(ctx, budget.Patterns)
	if err != nil {
		return budget.Report{}, err
	}
	return budget.Measure(files, m.ceilings), nil
}
