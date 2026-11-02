package sentinel

import (
	"context"

	"github.com/ronz204/deltas-ai/package/domain/findings"
	"github.com/ronz204/deltas-ai/package/domain/judgment"
	"github.com/ronz204/deltas-ai/package/domain/linking"
)

type Event struct {
	Name string
	Tool string
	Path string

	Session    string
	StopActive bool
}

type Decision struct {
	Block   bool
	Reason  string
	Context string
}

type Health interface {
	Check(ctx context.Context) []string
}

type Evidence interface {
	Standings(ctx context.Context, slices map[string]linking.Scope) (map[string]judgment.Standing, error)
	Begin(ctx context.Context, session string) error
	Touched(ctx context.Context, session string) (files []string, known bool, err error)
}

type Linter interface {
	Lint(ctx context.Context) ([]findings.Diagnostic, error)
	IsTest(rel string) bool
}
