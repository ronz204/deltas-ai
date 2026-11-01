package ledger

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/judgment"
)

var ctx = context.Background()

type tree struct {
	hash  string
	dirty bool
	err   error
}

func (t *tree) Tree(context.Context) (string, bool, error) { return t.hash, t.dirty, t.err }

func TestARecordedSliceStandsOnlyOnTheTreeItSaw(t *testing.T) {
	tr := &tree{hash: "a", dirty: true}
	l := New(t.TempDir(), "/repo", tr)
	if err := l.Record(ctx, map[string]bool{"billing": false, "pay": true}); err != nil {
		t.Fatal(err)
	}
	dirty, by, err := l.Standings(ctx, []string{"billing", "pay", "never"})
	if err != nil || !dirty {
		t.Fatalf("Standings = %v, %v, %v", dirty, by, err)
	}
	want := map[string]judgment.Standing{"billing": judgment.Passing, "pay": judgment.Failing, "never": judgment.Unchecked}
	for slice, w := range want {
		if by[slice] != w {
			t.Errorf("%s = %v, want %v", slice, by[slice], w)
		}
	}

	tr.hash = "b"
	if _, by, _ := l.Standings(ctx, []string{"billing"}); by["billing"] != judgment.Unchecked {
		t.Fatalf("after the tree changed billing = %v, want Unchecked", by["billing"])
	}
}

func TestACleanTreeHasNothingToProve(t *testing.T) {
	l := New(t.TempDir(), "/repo", &tree{})
	if dirty, by, err := l.Standings(ctx, []string{"billing"}); err != nil || dirty || by != nil {
		t.Fatalf("Standings = %v, %v, %v; want clean", dirty, by, err)
	}
}

func TestRecordsKeepEarlierSlicesAndAreKeptPerProject(t *testing.T) {
	cache, tr := t.TempDir(), &tree{hash: "a", dirty: true}
	one, other := New(cache, "/one", tr), New(cache, "/other", tr)
	if err := one.Record(ctx, map[string]bool{"billing": false}); err != nil {
		t.Fatal(err)
	}
	if err := one.Record(ctx, map[string]bool{"pay": false}); err != nil {
		t.Fatal(err)
	}
	_, by, _ := one.Standings(ctx, []string{"billing", "pay"})
	if by["billing"] != judgment.Passing || by["pay"] != judgment.Passing {
		t.Fatalf("a second record lost the first: %v", by)
	}
	if _, by, _ := other.Standings(ctx, []string{"billing"}); by["billing"] != judgment.Unchecked {
		t.Fatalf("another project saw this project's record: %v", by)
	}
}

func TestADamagedRecordIsReplacedNotFatal(t *testing.T) {
	cache, tr := t.TempDir(), &tree{hash: "a", dirty: true}
	l := New(cache, "/repo", tr)
	if err := l.Record(ctx, map[string]bool{"billing": false}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(l.file, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, by, err := l.Standings(ctx, []string{"billing"}); err != nil || by["billing"] != judgment.Unchecked {
		t.Fatalf("damaged: %v, %v; want Unchecked and no error", by, err)
	}
	if err := l.Record(ctx, map[string]bool{"billing": false}); err != nil {
		t.Fatalf("recording over a damaged file: %v", err)
	}
	if _, by, _ := l.Standings(ctx, []string{"billing"}); by["billing"] != judgment.Passing {
		t.Fatal("the record was not repaired")
	}
}

func TestNothingIsLeftBesideTheRecord(t *testing.T) {
	cache := t.TempDir()
	l := New(cache, "/repo", &tree{hash: "a", dirty: true})
	for i := 0; i < 2; i++ {
		if err := l.Record(ctx, map[string]bool{"billing": false}); err != nil {
			t.Fatal(err)
		}
	}
	entries, _ := os.ReadDir(filepath.Join(cache, "delta"))
	if len(entries) != 1 || strings.HasPrefix(entries[0].Name(), ".record") {
		t.Fatalf("the cache holds %d entries, want only the record", len(entries))
	}
}

func TestFailuresAreReportedNotHidden(t *testing.T) {
	boom := errors.New("git is gone")
	if err := New(t.TempDir(), "/repo", &tree{err: boom}).Record(ctx, map[string]bool{"a": false}); !errors.Is(err, boom) {
		t.Errorf("Record = %v, want the tree's error", err)
	}
	if _, _, err := New(t.TempDir(), "/repo", &tree{err: boom}).Standings(ctx, []string{"a"}); !errors.Is(err, boom) {
		t.Errorf("Standings = %v, want the tree's error", err)
	}
	if err := New("", "/repo", &tree{hash: "a", dirty: true}).Record(ctx, map[string]bool{"a": false}); err == nil {
		t.Error("Record without a cache directory succeeded")
	}
}
