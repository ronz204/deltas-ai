package tracking

import (
	"context"
	"maps"
	"slices"

	"github.com/ronz204/deltas-ai/package/domain/guides"
)

type DocsRow struct {
	Total  int
	Filled int
	Next   string
}

func (r *Reporter) Docs(ctx context.Context) (DocsRow, error) {
	if r.files == nil {
		return DocsRow{}, nil
	}
	found, err := r.files.Glob(ctx, []string{guides.Dir + "/*.md"})
	if err != nil {
		return DocsRow{}, err
	}
	var row DocsRow
	open := map[string]bool{}
	for _, f := range found {
		name, ok := guides.Name(f.Path)
		if !ok {
			continue
		}
		row.Total++
		if guides.Filled(f.Text) {
			row.Filled++
		} else {
			open[name] = true
		}
	}
	for _, name := range append(slices.Clone(guides.Order), slices.Sorted(maps.Keys(open))...) {
		if open[name] {
			row.Next = name
			break
		}
	}
	return row, nil
}
