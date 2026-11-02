package tracking

import (
	"context"
	"fmt"

	"github.com/ronz204/deltas-ai/package/domain/specfile"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

type NotReady struct{ Reason string }

func (e *NotReady) Error() string { return e.Reason }

func (r *Reporter) Ready(ctx context.Context, slice string) error {
	src, err := r.specs.Load(ctx, slice)
	if err != nil {
		return err
	}
	sp, ds := specfile.ParseSlice(r.specs.Path(slice), slice, src)
	switch {
	case len(ds) > 0:
		return &NotReady{fmt.Sprintf("the spec of %s has %d problems; run delta spec lint %s", slice, len(ds), slice)}
	case sp.Status == states.Draft:
		return &NotReady{fmt.Sprintf("%s is a draft; a person approves it first with delta approve %s", slice, slice)}
	}
	return nil
}
