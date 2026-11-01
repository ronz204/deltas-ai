// Package status summarises where every slice stands.
package status

import (
	"context"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
	"github.com/ronz204/deltas-ai/source/domain/lifecycle"
	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
	"github.com/ronz204/deltas-ai/source/domain/specfile"
)

type Specs interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, slice string) ([]byte, error)
	Path(slice string) string
}

type Reporter struct {
	specs  Specs
	scopes map[string][]string
}

// New builds the reporter. scopes holds, for the slices that have them, the paths delta.conf says they depend on.
func New(specs Specs, scopes map[string][]string) *Reporter {
	return &Reporter{specs: specs, scopes: scopes}
}

// Row counts a slice's claims by tag; Issues counts the problems found, and the counts cover only clean claims.
type Row struct {
	Slice  string
	Status lifecycle.Status // empty when the frontmatter could not be read
	Claims int
	Test   int
	Manual int
	None   int
	Issues int
	Scope  linkcheck.Scope // the part of the project the slice depends on
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
		file := r.specs.Path(name)
		sp, ds := specfile.ParseSlice(file, name, src)
		row := Row{Slice: name, Status: sp.Status, Claims: len(sp.Claims), Issues: len(ds), Scope: linkcheck.NewScope(file, r.scopes[name])}
		for _, c := range sp.Claims {
			switch c.Tag.Method {
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
