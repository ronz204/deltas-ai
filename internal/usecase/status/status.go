// Package status summarises where every slice stands.
package status

import (
	"context"

	"github.com/ronz204/deltas-ai/internal/domain/assertion"
	"github.com/ronz204/deltas-ai/internal/domain/lifecycle"
	"github.com/ronz204/deltas-ai/internal/domain/specfile"
)

type Specs interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, slice string) ([]byte, error)
	Path(slice string) string
}

type Reporter struct{ specs Specs }

func New(specs Specs) *Reporter { return &Reporter{specs: specs} }

// Row counts a slice's claims by tag; Issues counts the problems found, and the counts cover only clean claims.
type Row struct {
	Slice  string
	Status lifecycle.Status // empty when the frontmatter could not be read
	Claims int
	Test   int
	Manual int
	None   int
	Issues int
}

func (r *Reporter) Report(ctx context.Context) ([]Row, error) {
	names, err := r.specs.List(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]Row, 0, len(names))
	for _, name := range names {
		src, err := r.specs.Load(ctx, name)
		if err != nil {
			return nil, err
		}
		sp, ds := specfile.ParseSlice(r.specs.Path(name), name, src)
		row := Row{Slice: name, Status: sp.Status, Claims: len(sp.Claims), Issues: len(ds)}
		for _, c := range sp.Claims {
			switch c.Tag.Kind {
			case assertion.Test:
				row.Test++
			case assertion.Manual:
				row.Manual++
			default:
				row.None++
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}
