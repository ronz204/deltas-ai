package recorder

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ronz204/deltas-ai/package/domain/diffing"
	"github.com/ronz204/deltas-ai/package/domain/judgment"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/usecase/verifier"
)

var ctx = context.Background()

type tree struct {
	changed map[string]string
	all     map[string]string
	err     error
	since   []string
	began   int
}

func (t *tree) Baseline(context.Context) (diffing.Baseline, error) {
	t.began++
	return diffing.Baseline{Head: "h", Dirty: map[string]string{"a.go": "1"}}, t.err
}

func (t *tree) Since(_ context.Context, b diffing.Baseline) ([]string, error) {
	if b.Head != "h" || b.Dirty["a.go"] != "1" {
		return nil, errors.New("the baseline was not kept as it was taken")
	}
	return t.since, t.err
}

func (t *tree) Files(_ context.Context, scope linking.Scope) (map[string]string, error) {
	out := map[string]string{}
	for name, hash := range t.all {
		if scope.Covers(name) {
			out[name] = hash
		}
	}
	return out, t.err
}

func (t *tree) Tree(_ context.Context, scope linking.Scope) (string, bool, error) {
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
	billing  = linking.NewScope("deltas/specs/billing.spec.md", nil)
	payments = linking.NewScope("deltas/specs/payments.spec.md", nil)
)

func TestARecordedSliceStandsOnlyOnTheTreeItSaw(t *testing.T) {
	tr := &tree{changed: map[string]string{"billing/total.go": "v1", "payments/pay.go": "v1"}}
	l := New(t.TempDir(), "/repo", tr)
	if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing}, {Slice: "payments", Scope: payments, Violated: true}}); err != nil {
		t.Fatal(err)
	}
	by, err := l.Standings(ctx, map[string]linking.Scope{"billing": billing, "payments": payments})
	if err != nil || by["billing"] != judgment.Passing || by["payments"] != judgment.Failing {
		t.Fatalf("Standings = %v, %v", by, err)
	}

	tr.changed["billing/total.go"] = "v2"
	by, _ = l.Standings(ctx, map[string]linking.Scope{"billing": billing, "payments": payments})
	if by["billing"] != judgment.Unchecked || by["payments"] != judgment.Unchecked {
		t.Fatalf("after the code changed: %v; every slice depends on every code file, so each needs a new check", by)
	}
}

func TestAChangeToDocumentationDoesNotMatterToASlice(t *testing.T) {
	l := New(t.TempDir(), "/repo", &tree{})
	by, err := l.Standings(ctx, map[string]linking.Scope{"billing": billing, "payments": payments})
	if err != nil || by["billing"] != judgment.Passing || by["payments"] != judgment.Passing {
		t.Fatalf("Standings = %v, %v; want both to stand", by, err)
	}
	tr := &tree{changed: map[string]string{"README.md": "x"}}
	by, err = New(t.TempDir(), "/repo", tr).Standings(ctx, map[string]linking.Scope{"billing": billing})
	if err != nil || by["billing"] != judgment.Passing {
		t.Fatalf("a change to README.md: %v, %v; nothing it depends on changed, so there is nothing to prove", by, err)
	}
	tr.changed["payments/pay.go"] = "v1"
	if by, _ = New(t.TempDir(), "/repo", tr).Standings(ctx, map[string]linking.Scope{"billing": billing}); by["billing"] != judgment.Unchecked {
		t.Errorf("billing = %v: code changed and it was never checked", by["billing"])
	}
}

func TestACleanTreeHasNothingToProve(t *testing.T) {
	by, err := New(t.TempDir(), "/repo", &tree{}).Standings(ctx, map[string]linking.Scope{"billing": billing})
	if err != nil || by["billing"] != judgment.Passing {
		t.Fatalf("Standings = %v, %v; want it to stand", by, err)
	}
}

func TestRecordsKeepEarlierSlicesAndAreKeptPerProject(t *testing.T) {
	cache, tr := t.TempDir(), &tree{changed: map[string]string{"billing/a.go": "x", "payments/b.go": "x"}}
	one, other := New(cache, "/one", tr), New(cache, "/other", tr)
	if err := one.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing}}); err != nil {
		t.Fatal(err)
	}
	if err := one.Record(ctx, []verifier.Seen{{Slice: "payments", Scope: payments}}); err != nil {
		t.Fatal(err)
	}
	both := map[string]linking.Scope{"billing": billing, "payments": payments}
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
	one := map[string]linking.Scope{"billing": billing}
	if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing}}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(l.file, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if by, err := l.Standings(ctx, one); err != nil || by["billing"] != judgment.Unchecked {
		t.Fatalf("damaged: %v, %v; want Unchecked and no error", by, err)
	}
	if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing}}); err != nil {
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
		if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing}}); err != nil {
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
	seen := []verifier.Seen{{Slice: "billing", Scope: billing}}
	if err := New(t.TempDir(), "/repo", &tree{err: boom}).Record(ctx, seen); !errors.Is(err, boom) {
		t.Errorf("Record = %v, want the tree's error", err)
	}
	if _, err := New(t.TempDir(), "/repo", &tree{err: boom}).Standings(ctx, map[string]linking.Scope{"billing": billing}); !errors.Is(err, boom) {
		t.Errorf("Standings = %v, want the tree's error", err)
	}
	if err := New("", "/repo", &tree{}).Record(ctx, seen); err == nil {
		t.Error("Record without a cache directory succeeded")
	}
}

func TestASliceThatWasNeverVerifiedHereHasNoDriftToTell(t *testing.T) {
	tr := &tree{all: map[string]string{"billing/total.go": "v1"}}
	l := New(t.TempDir(), "/repo", tr)
	if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing}}); err != nil {
		t.Fatal(err)
	}
	if changed, known, err := l.Drift(ctx, "billing", billing); err != nil || known || changed != nil {
		t.Fatalf("Drift = %v, %v, %v; a check that did not verify the slice sets no baseline", changed, known, err)
	}
}

func TestDriftListsTheFilesThatChangedSinceTheSliceWasVerified(t *testing.T) {
	tr := &tree{all: map[string]string{"billing/total.go": "v1", "billing/tax.go": "v1", "payments/pay.go": "v1"}}
	l := New(t.TempDir(), "/repo", tr)
	if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing, Verified: true}}); err != nil {
		t.Fatal(err)
	}
	if changed, known, err := l.Drift(ctx, "billing", billing); err != nil || !known || len(changed) != 0 {
		t.Fatalf("Drift = %v, %v, %v; nothing changed yet", changed, known, err)
	}
	tr.all["billing/total.go"] = "v2"
	tr.all["payments/pay.go"] = "v2"
	delete(tr.all, "billing/tax.go")
	tr.all["billing/new.go"] = "v1"
	changed, known, err := l.Drift(ctx, "billing", billing)
	if want := []string{"billing/new.go", "billing/tax.go", "billing/total.go", "payments/pay.go"}; err != nil || !known || !slices.Equal(changed, want) {
		t.Fatalf("Drift = %v, %v, %v; want %v", changed, known, err, want)
	}
}

func TestDriftIsKeptAcrossACheckThatDoesNotVerify(t *testing.T) {
	tr := &tree{all: map[string]string{"billing/total.go": "v1"}}
	l := New(t.TempDir(), "/repo", tr)
	if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing, Verified: true}}); err != nil {
		t.Fatal(err)
	}
	tr.all["billing/total.go"] = "v2"
	if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing, Violated: true}}); err != nil {
		t.Fatal(err)
	}
	if changed, known, _ := l.Drift(ctx, "billing", billing); !known || len(changed) != 1 {
		t.Fatalf("Drift = %v, %v; a failing check must not become the new baseline", changed, known)
	}
	if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing, Verified: true}}); err != nil {
		t.Fatal(err)
	}
	if changed, _, _ := l.Drift(ctx, "billing", billing); len(changed) != 0 {
		t.Fatalf("Drift = %v; a new verification is the new baseline", changed)
	}
}

func TestDriftFailureIsReported(t *testing.T) {
	boom := errors.New("boom")
	tr := &tree{all: map[string]string{"billing/total.go": "v1"}}
	l := New(t.TempDir(), "/repo", tr)
	if err := l.Record(ctx, []verifier.Seen{{Slice: "billing", Scope: billing, Verified: true}}); err != nil {
		t.Fatal(err)
	}
	tr.err = boom
	if _, _, err := l.Drift(ctx, "billing", billing); !errors.Is(err, boom) {
		t.Fatalf("Drift error = %v, want %v", err, boom)
	}
}

func TestASessionKeepsTheBaselineItStartedWith(t *testing.T) {
	tr := &tree{since: []string{"billing/total.go"}}
	l := New(t.TempDir(), "/repo", tr)
	if _, known, err := l.Touched(ctx, "s1"); err != nil || known {
		t.Fatalf("before Begin: known = %v, %v", known, err)
	}
	if err := l.Begin(ctx, "s1"); err != nil {
		t.Fatal(err)
	}
	files, known, err := l.Touched(ctx, "s1")
	if err != nil || !known || !slices.Equal(files, []string{"billing/total.go"}) {
		t.Fatalf("Touched = %v, %v, %v", files, known, err)
	}
	if _, known, _ := l.Touched(ctx, "other"); known {
		t.Fatal("another session read the baseline of s1")
	}
}

func TestBeginAgainDoesNotReplaceTheBaseline(t *testing.T) {
	tr := &tree{}
	l := New(t.TempDir(), "/repo", tr)
	for range 3 {
		if err := l.Begin(ctx, "s1"); err != nil {
			t.Fatal(err)
		}
	}
	if tr.began != 1 {
		t.Fatalf("the baseline was taken %d times; a resumed session must keep the first one", tr.began)
	}
}

func TestBeginDropsTheBaselinesOfOldSessions(t *testing.T) {
	l := New(t.TempDir(), "/repo", &tree{})
	if err := l.Begin(ctx, "old"); err != nil {
		t.Fatal(err)
	}
	old := l.sessionFile("old")
	then := time.Now().Add(-2 * sessionTTL)
	if err := os.Chtimes(old, then, then); err != nil {
		t.Fatal(err)
	}
	if err := l.Begin(ctx, "new"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("the old baseline is still there: %v", err)
	}
	if _, err := os.Stat(l.sessionFile("new")); err != nil {
		t.Fatalf("the new baseline is gone: %v", err)
	}
}

func TestASessionWithoutACacheIsUnknown(t *testing.T) {
	l := New("", "/repo", &tree{})
	if l.Begin(ctx, "s1") == nil {
		t.Fatal("Begin without a cache directory must say so")
	}
	if _, known, err := l.Touched(ctx, "s1"); known || err != nil {
		t.Fatalf("known = %v, %v", known, err)
	}
}
