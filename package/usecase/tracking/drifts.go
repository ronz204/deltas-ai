package tracking

import (
	"context"
	"slices"

	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/domain/specfile"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

type Drift struct {
	Slice  string
	Files  []string
	Marked bool
}

func (r *Reporter) Stale(ctx context.Context) ([]Drift, error) {
	if r.baselines == nil || r.saver == nil {
		return nil, nil
	}
	names, err := r.specs.List(ctx)
	if err != nil {
		return nil, err
	}
	var out []Drift
	for _, name := range names {
		src, err := r.specs.Load(ctx, name)
		if err != nil {
			return out, err
		}
		file := r.specs.Path(name)
		sp, ds := specfile.ParseSlice(file, name, src)
		if len(ds) > 0 || (sp.Status != states.Verified && sp.Status != states.Stale) {
			continue
		}
		files, known, err := r.baselines.Drift(ctx, name, linking.NewScope(file, sp.Claims))
		if err != nil {
			return out, err
		}
		if sp.Status == states.Stale {
			files = slices.DeleteFunc(files, func(f string) bool { return f == file })
		}
		if !known || len(files) == 0 {
			continue
		}
		d := Drift{Slice: name, Files: files}
		if sp.Status == states.Verified {
			edited, err := specfile.SetStatus(src, states.Stale)
			if err != nil {
				return out, err
			}
			if err := r.saver.Save(ctx, name, edited); err != nil {
				return out, err
			}
			d.Marked = true
		}
		out = append(out, d)
	}
	return out, nil
}
