package tracking

import (
	"context"

	"github.com/ronz204/deltas-ai/package/domain/budget"
	"github.com/ronz204/deltas-ai/package/domain/claimset"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/domain/specfile"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

type Specs interface {
	List(ctx context.Context) ([]string, error)
	Load(ctx context.Context, slice string) ([]byte, error)
	Path(slice string) string
}

type Baselines interface {
	Drift(ctx context.Context, slice string, scope linking.Scope) (changed []string, known bool, err error)
}

type Files interface {
	Glob(ctx context.Context, patterns []string) ([]budget.File, error)
}

type Reporter struct {
	specs     Specs
	baselines Baselines
	files     Files
}

func New(specs Specs) *Reporter {
	return &Reporter{specs: specs}
}

func (r *Reporter) WithDrift(b Baselines) *Reporter {
	r.baselines = b
	return r
}

func (r *Reporter) WithDocs(f Files) *Reporter {
	r.files = f
	return r
}

type Row struct {
	Slice  string
	Status states.Status
	Claims int
	Test   int
	Manual int
	None   int
	Issues int
	Scope  linking.Scope

	Changed []string
}

func (r *Reporter) Report(ctx context.Context) ([]Row, error) {
	drifted, _ := r.Drifts(ctx)
	changed := map[string][]string{}
	for _, d := range drifted {
		changed[d.Slice] = d.Files
	}
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
		row := Row{Slice: name, Status: sp.Status, Claims: len(sp.Claims), Issues: len(ds), Scope: linking.NewScope(file, sp.Claims), Changed: changed[name]}
		for _, c := range sp.Claims {
			switch c.Tag.Method {
			case claimset.Test:
				row.Test++
			case claimset.Manual:
				row.Manual++
			default:
				row.None++
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}
