package tracing

import (
	"context"
	"errors"
	"slices"
	"testing"
)

var ctx = context.Background()

type marks struct {
	changed, gone []string
	known         bool
	err           error
	began         []string
	dropped       []string
}

func (m *marks) Snapshot(_ context.Context, slice string) error {
	m.began = append(m.began, slice)
	return m.err
}

func (m *marks) Compare(context.Context, string) ([]string, []string, bool, error) {
	return m.changed, m.gone, m.known, m.err
}

func (m *marks) Drop(_ context.Context, slice string) error {
	m.dropped = append(m.dropped, slice)
	return nil
}

type store struct {
	files map[string][]string
	err   error
}

func (s *store) Footprint(_ context.Context, slice string) ([]string, error) {
	return slices.Clone(s.files[slice]), nil
}

func (s *store) SaveFootprint(_ context.Context, slice string, files []string) error {
	if s.err != nil {
		return s.err
	}
	s.files[slice] = files
	return nil
}

func TestBeginTakesTheSnapshotOfTheSlice(t *testing.T) {
	m := &marks{}
	if err := New(m, &store{}).Begin(ctx, "billing"); err != nil || !slices.Equal(m.began, []string{"billing"}) {
		t.Fatalf("began = %v, %v", m.began, err)
	}
}

func TestFinishRecordsWhatTheBuildChangedAndKeepsWhatWasThere(t *testing.T) {
	m := &marks{known: true, changed: []string{"src/new.go", "src/old.go"}, gone: []string{"src/gone.go"}}
	s := &store{files: map[string][]string{"billing": {"src/old.go", "src/gone.go", "src/kept.go"}}}
	files, recorded, err := New(m, s).Finish(ctx, "billing")
	want := []string{"src/kept.go", "src/new.go", "src/old.go"}
	if err != nil || !recorded || !slices.Equal(files, want) || !slices.Equal(s.files["billing"], want) {
		t.Fatalf("Finish = %v %v %v, stored %v; want %v", files, recorded, err, s.files["billing"], want)
	}
	if !slices.Equal(m.dropped, []string{"billing"}) {
		t.Fatalf("dropped = %v, want the snapshot gone once it was used", m.dropped)
	}
}

func TestFinishWithoutASnapshotChangesNothing(t *testing.T) {
	s := &store{files: map[string][]string{"billing": {"a.go"}}}
	files, recorded, err := New(&marks{}, s).Finish(ctx, "billing")
	if err != nil || recorded || files != nil || !slices.Equal(s.files["billing"], []string{"a.go"}) {
		t.Fatalf("Finish = %v %v %v, stored %v", files, recorded, err, s.files["billing"])
	}
}

func TestFinishKeepsTheSnapshotWhenTheFootprintCannotBeSaved(t *testing.T) {
	m := &marks{known: true, changed: []string{"a.go"}}
	s := &store{files: map[string][]string{}, err: errors.New("read-only disk")}
	if _, recorded, err := New(m, s).Finish(ctx, "billing"); err == nil || recorded || len(m.dropped) != 0 {
		t.Fatalf("recorded = %v, err = %v, dropped = %v; a failed save must leave the snapshot to try again", recorded, err, m.dropped)
	}
}

func TestFinishReportsWhyItCouldNotCompare(t *testing.T) {
	m := &marks{known: true, err: errors.New("git is gone")}
	if _, recorded, err := New(m, &store{files: map[string][]string{}}).Finish(ctx, "billing"); err == nil || recorded {
		t.Fatalf("recorded = %v, err = %v", recorded, err)
	}
}
