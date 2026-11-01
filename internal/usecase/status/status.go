// Package status summarises where every slice stands.
package status

import (
	"context"

	"github.com/ronz204/deltas-ai/internal/domain/claim"
	"github.com/ronz204/deltas-ai/internal/domain/slice"
	"github.com/ronz204/deltas-ai/internal/domain/spec"
)

// Specs is where specs are read from.
type Specs interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, slice string) ([]byte, error)
	Path(slice string) string
}

// Reporter summarises the slices of a project.
type Reporter struct{ specs Specs }

func New(specs Specs) *Reporter { return &Reporter{specs: specs} }

// Row is one slice. Test, Manual and None count its claims by how they are verified; the
// unverified debt is everything that is not Test. Issues counts the problems found in its spec,
// and when there are any the other numbers cover only the claims that parsed cleanly.
type Row struct {
	Slice  string
	Status slice.Status // empty when the frontmatter could not be read
	Claims int
	Test   int
	Manual int
	None   int
	Issues int
}

// Report returns one row per slice, in name order.
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
		sp, ds := spec.ParseSlice(r.specs.Path(name), name, src)
		row := Row{Slice: name, Status: sp.Status, Claims: len(sp.Claims), Issues: len(ds)}
		for _, c := range sp.Claims {
			switch c.Tag.Kind {
			case claim.Test:
				row.Test++
			case claim.Manual:
				row.Manual++
			default:
				row.None++
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}
