package tracing

import (
	"context"
	"slices"
)

type Marks interface {
	Snapshot(ctx context.Context, slice string) error
	Compare(ctx context.Context, slice string) (changed, gone []string, known bool, err error)
	Drop(ctx context.Context, slice string) error
}

type Store interface {
	Footprint(ctx context.Context, slice string) ([]string, error)
	SaveFootprint(ctx context.Context, slice string, files []string) error
}

type Tracer struct {
	marks Marks
	store Store
}

func New(marks Marks, store Store) *Tracer { return &Tracer{marks: marks, store: store} }

func (t *Tracer) Begin(ctx context.Context, slice string) error {
	return t.marks.Snapshot(ctx, slice)
}

func (t *Tracer) Finish(ctx context.Context, slice string) (files []string, recorded bool, err error) {
	changed, gone, known, err := t.marks.Compare(ctx, slice)
	if err != nil || !known {
		return nil, false, err
	}
	files, err = t.store.Footprint(ctx, slice)
	if err != nil {
		return nil, false, err
	}
	files = append(files, changed...)
	files = slices.DeleteFunc(files, func(f string) bool { return slices.Contains(gone, f) })
	slices.Sort(files)
	files = slices.Compact(files)
	if err := t.store.SaveFootprint(ctx, slice, files); err != nil {
		return nil, false, err
	}
	_ = t.marks.Drop(ctx, slice)
	return files, true, nil
}
