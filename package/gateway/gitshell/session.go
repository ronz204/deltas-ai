package gitshell

import (
	"context"
	"slices"

	"github.com/ronz204/deltas-ai/package/domain/diffing"
)

func (g *Git) Baseline(ctx context.Context) (diffing.Baseline, error) {
	head, changed, err := g.head(ctx)
	if err != nil {
		return diffing.Baseline{}, err
	}
	names, err := g.dirtyNames(ctx, changed)
	if err != nil {
		return diffing.Baseline{}, err
	}
	dirty := make(map[string]string, len(names))
	for _, name := range names {
		dirty[name] = g.contentHash(name)
	}
	return diffing.Baseline{Head: head, Dirty: dirty}, nil
}

func (g *Git) Since(ctx context.Context, b diffing.Baseline) ([]string, error) {
	changed := []string{"ls-files", "-z", "--cached"}
	if b.Head != "" {
		changed = []string{"diff", "--name-only", "--no-renames", "-z", b.Head}
	}
	names, err := g.dirtyNames(ctx, changed)
	if err != nil {
		return nil, err
	}
	for name := range b.Dirty {
		names = append(names, name)
	}
	slices.Sort(names)
	names = slices.Compact(names)
	return slices.DeleteFunc(names, func(name string) bool {
		before, wasDirty := b.Dirty[name]
		return wasDirty && before == g.contentHash(name)
	}), nil
}

func (g *Git) dirtyNames(ctx context.Context, changed []string) ([]string, error) {
	names, err := g.names(ctx, changed...)
	if err != nil {
		return nil, err
	}
	untracked, err := g.names(ctx, "ls-files", "-z", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	return append(names, untracked...), nil
}
