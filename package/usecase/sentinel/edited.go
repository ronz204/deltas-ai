package sentinel

import (
	"context"
	"fmt"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/kbpaths"
)

func (g *Guard) after(ctx context.Context, e Event) (Decision, error) {
	if !fileTools[e.Tool] {
		return Decision{}, nil
	}
	rel, ok := g.relative(e.Path)
	if !ok {
		return Decision{}, nil
	}
	var lines []string
	if !kbpaths.Covers(rel) {
		drifted, err := g.slices.Stale(ctx)
		if err != nil {
			return Decision{}, err
		}
		for _, d := range drifted {
			if d.Marked {
				lines = append(lines, fmt.Sprintf("%s was verified and is now stale: %s changed. Check that its contract still holds: if it does, rebuild with /builder %s; if the contract is out of date, say so to the user, who decides it with /drafter.", d.Slice, fileList(d.Files), d.Slice))
			}
		}
	}
	lint, err := g.lintOf(ctx, rel)
	if err != nil {
		return Decision{}, err
	}
	if lines = append(lines, lint...); len(lines) == 0 {
		return Decision{}, nil
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
