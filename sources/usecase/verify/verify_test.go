package verify

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ronz204/deltas-ai/sources/domain/judgment"
	"github.com/ronz204/deltas-ai/sources/domain/linkcheck"
)

var ctx = context.Background()

type fakeSpecs map[string]string

func (f fakeSpecs) Path(s string) string { return "deltas/" + s + ".spec.md" }
func (f fakeSpecs) List(context.Context) ([]string, error) {
	var names []string
	for _, n := range []string{"billing", "broken", "payments"} { // fixed order, like a sorted listing
		if _, ok := f[n]; ok {
			names = append(names, n)
		}
	}
	return names, nil
}
func (f fakeSpecs) Load(_ context.Context, s string) ([]byte, error) {
	src, ok := f[s]
	if !ok {
		return nil, errors.New("no spec for " + s)
	}
	return []byte(src), nil
}

type fakeTests []linkcheck.File

func (f fakeTests) TestFiles(context.Context) ([]linkcheck.File, error) { return f, nil }

type fakeRunner struct {
	exits map[string]int
	errs  map[string]error
	ran   []string
}

func (f *fakeRunner) Run(_ context.Context, command string) (int, error) {
	f.ran = append(f.ran, command)
	return f.exits[command], f.errs[command]
}

func specOf(slice string, claims ...string) string {
	out := "---\nformat: 1\nslice: " + slice + "\nstatus: approved\n---\n"
	for _, c := range claims {
		out += c
	}
	return out
}

const (
	passing  = "- INV-01 [test: go test ./billing]\n  When x, y.\n"
	failing  = "- INV-02 [test: go test ./failing]\n  When x, y.\n"
	uncited  = "- INV-03 [test: go test ./billing]\n  When x, y.\n"
	manual   = "- INV-04 [manual: look at it]\n  When x, y.\n"
	untagged = "- INV-05 [none]\n  When x, y.\n"
)

var tests = fakeTests{{Path: "billing/total_test.go", Text: "// INV-01\n// INV-02\n"}}

func newVerify(specs fakeSpecs, run *fakeRunner) *Checker { return New(specs, tests, run) }

func TestCheckJudgesEveryKindOfClaim(t *testing.T) {
	specs := fakeSpecs{"billing": specOf("billing", passing, failing, uncited, manual, untagged)}
	run := &fakeRunner{exits: map[string]int{"go test ./failing": 1}}
	rep, err := newVerify(specs, run).Check(ctx, "")
	if err != nil || len(rep.Diags) != 0 {
		t.Fatalf("Check = %+v, %v", rep, err)
	}
	var got []string
	for _, r := range rep.Results {
		got = append(got, r.ID.String()+" "+r.Outcome.Verdict.String())
	}
	want := []string{"INV-01 Holds", "INV-02 Violated", "INV-03 Unverified", "INV-04 Unverified", "INV-05 Unverified"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("verdicts = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(run.ran, []string{"go test ./billing", "go test ./failing", "go test ./billing"}) {
		t.Fatalf("ran %v: only test claims have a command to run", run.ran)
	}
}

func TestCheckIsScopedToTheNamedSlice(t *testing.T) {
	specs := fakeSpecs{"billing": specOf("billing", passing), "payments": specOf("payments", passing)}
	rep, err := newVerify(specs, &fakeRunner{}).Check(ctx, "payments")
	if err != nil || len(rep.Results) != 1 || rep.Results[0].Slice != "payments" {
		t.Fatalf("Check(payments) = %+v, %v", rep, err)
	}
	if _, err := newVerify(specs, &fakeRunner{}).Check(ctx, "ghost"); err == nil {
		t.Fatal("an unknown slice was accepted")
	}
}

func TestCheckDoesNotRunABrokenSpec(t *testing.T) {
	specs := fakeSpecs{"billing": specOf("billing", passing), "broken": "no frontmatter"}
	run := &fakeRunner{}
	rep, err := newVerify(specs, run).Check(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Diags) == 0 || rep.Diags[0].File != "deltas/broken.spec.md" {
		t.Fatalf("the broken spec was not reported: %+v", rep.Diags)
	}
	if len(rep.Results) != 1 || rep.Results[0].Slice != "billing" {
		t.Fatalf("results = %+v, want only billing's", rep.Results)
	}
}

func TestCheckAnUnrunnableCommandIsUnverifiedWithItsReason(t *testing.T) {
	specs := fakeSpecs{"billing": specOf("billing", passing)}
	run := &fakeRunner{errs: map[string]error{"go test ./billing": errors.New("not found")}}
	rep, _ := newVerify(specs, run).Check(ctx, "")
	got := rep.Results[0].Outcome
	if got.Verdict != judgment.Unverified || got.Reason != "could not run the command: not found" {
		t.Fatalf("outcome = %+v", got)
	}
}

func TestCheckStopsWhenCancelled(t *testing.T) {
	specs := fakeSpecs{"billing": specOf("billing", passing)}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := newVerify(specs, &fakeRunner{}).Check(cancelled, ""); err == nil {
		t.Fatal("a cancelled run reported success")
	}
}

func TestSpecLint(t *testing.T) {
	specs := fakeSpecs{"billing": specOf("billing", passing), "broken": specOf("other", passing)}
	ds, err := newVerify(specs, &fakeRunner{}).SpecLint(ctx, "")
	if err != nil || len(ds) != 1 || ds[0].Code != "SPEC-NAME" {
		t.Fatalf("SpecLint = %v, %v", ds, err)
	}
	if ds, _ := newVerify(specs, &fakeRunner{}).SpecLint(ctx, "billing"); len(ds) != 0 {
		t.Fatalf("billing alone is fine, got %v", ds)
	}
}

func TestLintCrossesSpecsWithTests(t *testing.T) {
	inv02 := "- INV-02 [test: go test ./billing]\n  When x, y.\n"
	specs := fakeSpecs{"billing": specOf("billing", passing, inv02, uncited)}
	ds, err := newVerify(specs, &fakeRunner{}).Lint(ctx)
	if err != nil || len(ds) != 1 || ds[0].Code != linkcheck.CodeNoCite || ds[0].Line != 10 {
		t.Fatalf("Lint = %v, %v", ds, err)
	}
}

func TestLintReportsOnlySpecProblemsWhenThereAreAny(t *testing.T) {
	specs := fakeSpecs{"billing": specOf("billing", uncited), "broken": "no frontmatter"}
	ds, err := newVerify(specs, &fakeRunner{}).Lint(ctx)
	if err != nil || len(ds) == 0 {
		t.Fatalf("Lint = %v, %v", ds, err)
	}
	for _, d := range ds {
		if d.Code == linkcheck.CodeNoCite {
			t.Fatalf("a broken project was crossed with its tests: %v", d)
		}
	}
}
