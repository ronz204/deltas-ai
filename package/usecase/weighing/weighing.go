package weighing

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

func New(files Files, ceilings budget.Ceilings) *Meter {
	return &Meter{files: files, ceilings: ceilings}
}

func (m *Meter) Always(ctx context.Context) (budget.Report, error) {
	files, err := m.files.Glob(ctx, budget.Patterns)
	if err != nil {
		return budget.Report{}, err
	}
	return budget.Measure(files, m.ceilings), nil
}
