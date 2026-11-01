package ledger

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/judgment"
	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
	"github.com/ronz204/deltas-ai/source/usecase/verify"
)

var ctx = context.Background()

// tree is a working tree with some changed files: its fingerprint covers the changed files a scope covers.
type tree struct {
	changed map[string]string
	err     error
}

func (t *tree) Tree(_ context.Context, scope linkcheck.Scope) (string, bool, error) {
	var parts []string
	for name, text := range t.changed {
		if scope.Covers(name) {
			parts = append(parts, name+"="+text)
		}
	}
	slices.Sort(parts)
	return strings.Join(parts, ";"), len(parts) > 0, t.err
}

var (
	billing  = linkcheck.NewScope("deltas/specifies/billing.spec.md", []string{"billing"})
	payments = linkcheck.NewScope("deltas/specifies/payments.spec.md", []string{"payments"})
)

func TestARecordedSliceStandsOnlyOnTheTreeItSaw(t *testing.T) {
	tr := &tree{changed: map[string]string{"billing/total.go": "v1", "payments/pay.go": "v1"}}
	l := New(t.TempDir(), "/repo", tr)
	if err := l.Record(ctx, []verify.Seen{{Slice: "billing", Scope: billing}, {Slice: "payments", Scope: payments, Violated: true}}); err != nil {
		t.Fatal(err)
	}
	by, err := l.Standings(ctx, map[string]linkcheck.Scope{"billing": billing, "payments": payments})
	if err != nil || by["billing"] != judgment.Passing || by["payments"] != judgment.Failing {
		t.Fatalf("Standings = %v, %v", by, err)
	}

	tr.changed["billing/total.go"] = "v2"
	by, _ = l.Standings(ctx, map[string]linkcheck.Scope{"billing": billing, "payments": payments})
	if by["billing"] != judgment.Unchecked || by["payments"] != judgment.Failing {
		t.Fatalf("after billing changed: %v; only billing should need a new check", by)
	}
}

func TestAChangeOutsideTheScopeOfASliceDoesNotMatterToIt(t *testing.T) {
	tr := &tree{changed: map[string]string{"README.md": "x", "payments/pay.go": "v1"}}
	l := New(t.TempDir(), "/repo", tr)
	by, err := l.Standings(ctx, map[string]linkcheck.Scope{"billing": billing, "payments": payments})
	if err != nil {
		t.Fatal(err)
	}
	if by["billing"] != judgment.Passing {
		t.Errorf("billing = %v: nothing it depends on changed, so there is nothing to prove", by["billing"])
	}
	if by["payments"] != judgment.Unchecked {
		t.Errorf("payments = %v: it changed and was never checked", by["payments"])
	}
}

func TestACleanTreeHasNothingToProve(t *testing.T) {
	by, err := New(t.TempDir(), "/repo", &tree{}).Standings(ctx, map[string]linkcheck.Scope{"billing": billing})
	if err != nil || by["billing"] != judgment.Passing {
		t.Fatalf("Standings = %v, %v; want it to stand", by, err)
	}
}

func TestRecordsKeepEarlierSlicesAndAreKeptPerProject(t *testing.T) {
	cache, tr := t.TempDir(), &tree{changed: map[string]string{"billing/a.go": "x", "payments/b.go": "x"}}
	one, other := New(cache, "/one", tr), New(cache, "/other", tr)
	if err := one.Record(ctx, []verify.Seen{{Slice: "billing", Scope: billing}}); err != nil {
		t.Fatal(err)
	}
	if err := one.Record(ctx, []verify.Seen{{Slice: "payments", Scope: payments}}); err != nil {
		t.Fatal(err)
	}
	both := map[string]linkcheck.Scope{"billing": billing, "payments": payments}
	by, _ := one.Standings(ctx, both)
	if by["billing"] != judgment.Passing || by["payments"] != judgment.Passing {
		t.Fatalf("a second record lost the first: %v", by)
	}
	if by, _ := other.Standings(ctx, both); by["billing"] != judgment.Unchecked {
		t.Fatalf("another project saw this project's record: %v", by)
	}
}

func TestADamagedRecordIsReplacedNotFatal(t *testing.T) {
	tr := &tree{changed: map[string]string{"billing/a.go": "x"}}
	l := New(t.TempDir(), "/repo", tr)
	one := map[string]linkcheck.Scope{"billing": billing}
	if err := l.Record(ctx, []verify.Seen{{Slice: "billing", Scope: billing}}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(l.file, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if by, err := l.Standings(ctx, one); err != nil || by["billing"] != judgment.Unchecked {
		t.Fatalf("damaged: %v, %v; want Unchecked and no error", by, err)
	}
	if err := l.Record(ctx, []verify.Seen{{Slice: "billing", Scope: billing}}); err != nil {
		t.Fatalf("recording over a damaged file: %v", err)
	}
	if by, _ := l.Standings(ctx, one); by["billing"] != judgment.Passing {
		t.Fatal("the record was not repaired")
	}
}

func TestNothingIsLeftBesideTheRecord(t *testing.T) {
	cache := t.TempDir()
	l := New(cache, "/repo", &tree{changed: map[string]string{"billing/a.go": "x"}})
	for i := 0; i < 2; i++ {
		if err := l.Record(ctx, []verify.Seen{{Slice: "billing", Scope: billing}}); err != nil {
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
	seen := []verify.Seen{{Slice: "billing", Scope: billing}}
	if err := New(t.TempDir(), "/repo", &tree{err: boom}).Record(ctx, seen); !errors.Is(err, boom) {
		t.Errorf("Record = %v, want the tree's error", err)
	}
	if _, err := New(t.TempDir(), "/repo", &tree{err: boom}).Standings(ctx, map[string]linkcheck.Scope{"billing": billing}); !errors.Is(err, boom) {
		t.Errorf("Standings = %v, want the tree's error", err)
	}
	if err := New("", "/repo", &tree{}).Record(ctx, seen); err == nil {
		t.Error("Record without a cache directory succeeded")
	}
}
