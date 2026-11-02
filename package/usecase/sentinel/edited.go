package sentinel

import (
	"context"
	"strings"
)

func (g *Guard) after(ctx context.Context, e Event) (Decision, error) {
	if !fileTools[e.Tool] {
		return Decision{}, nil
	}
	rel, ok := g.relative(e.Path)
	if !ok {
		return Decision{}, nil
	}
	lines, err := g.lintOf(ctx, rel)
	if err != nil || len(lines) == 0 {
		return Decision{}, err
	}
	return Decision{Block: true, Reason: strings.Join(lines, "\n")}, nil
}

func (g *Guard) lintOf(ctx context.Context, rel string) ([]string, error) {
	if g.linter == nil || !g.linter.IsTest(rel) {
		return nil, nil
	}
	ds, err := g.linter.Lint(ctx)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, d := range ds {
		if d.File == rel {
			lines = append(lines, "- "+d.String())
		}
	}
	if len(lines) == 0 {
		return nil, nil
	}
	return append([]string{"delta lint found problems in the test file you just wrote:"}, lines...), nil
}
