package verify

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/judgment"
	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
)

var ctx = context.Background()

type fakeSpecs map[string]string

func (f fakeSpecs) Path(s string) string { return "deltas/specs/" + s + ".spec.md" }
func (f fakeSpecs) List(context.Context) ([]string, error) {
	var names []string
	for _, n := range []string{"billing", "broken", "payments", "svc/auth/sessions"} { // fixed order, like a sorted listing
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

func (f fakeSpecs) Save(_ context.Context, s string, c []byte) error {
	f[s] = string(c)
	return nil
}

type fakeTests []linkcheck.File

func (f fakeTests) TestFiles(context.Context) ([]linkcheck.File, error) { return f, nil }
func (f fakeTests) IsTest(rel string) bool                              { return linkcheck.IsTestFile(rel) }

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

func newVerify(specs fakeSpecs, run *fakeRunner) *Checker { return New(specs, tests, run, nil, nil) }

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
	if len(rep.Diags) == 0 || rep.Diags[0].File != "deltas/specs/broken.spec.md" {
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

type fakeMarks struct {
	got  map[string]bool
	seen []Seen
	err  error
}

func (f *fakeMarks) Record(_ context.Context, seen []Seen) error {
	f.seen, f.got = seen, map[string]bool{}
	for _, s := range seen {
		f.got[s.Slice] = s.Violated
	}
	return f.err
}

func TestCheckRemembersWhichSlicesWereViolated(t *testing.T) {
	specs := fakeSpecs{
		"billing":  specOf("billing", passing),
		"broken":   "no frontmatter",
		"payments": specOf("payments", failing),
	}
	marks := &fakeMarks{}
	run := &fakeRunner{exits: map[string]int{"go test ./failing": 1}}
	if _, err := New(specs, tests, run, marks, nil).Check(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if want := map[string]bool{"billing": false, "payments": true}; !reflect.DeepEqual(marks.got, want) {
		t.Fatalf("recorded %v, want %v (a spec with problems was never checked, so it is not recorded)", marks.got, want)
	}
	if billing := marks.seen[0].Scope; !billing.Covers("payments/pay.go") || !billing.Covers("deltas/specs/billing.spec.md") || billing.Covers("README.md") {
		t.Fatalf("the scope of billing = %+v, want every code file and its spec when deltas.config names no paths", billing)
	}
}

func TestCheckStillReportsWhenItCannotBeRemembered(t *testing.T) {
	specs := fakeSpecs{"billing": specOf("billing", passing)}
	rep, err := New(specs, tests, &fakeRunner{}, &fakeMarks{err: errors.New("disk full")}, nil).Check(ctx, "")
	if err != nil || len(rep.Results) != 1 {
		t.Fatalf("Check = %+v, %v; the results must survive a failed record", rep, err)
	}
	if len(rep.Diags) != 1 || rep.Diags[0].Code != CodeRecord || !rep.Diags[0].Warn {
		t.Fatalf("diags = %+v, want one CHECK-RECORD warning", rep.Diags)
	}
}

func TestCheckRecordsNothingWhenNoSliceWasChecked(t *testing.T) {
	marks := &fakeMarks{}
	if _, err := New(fakeSpecs{"broken": "no frontmatter"}, tests, &fakeRunner{}, marks, nil).Check(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if marks.got != nil {
		t.Fatalf("recorded %v for a project with nothing checked", marks.got)
	}
}

func TestCheckRecordsTheNarrowScopeDeltaConfDeclares(t *testing.T) {
	specs := fakeSpecs{"billing": specOf("billing", passing)}
	marks := &fakeMarks{}
	scopes := map[string][]string{"billing": {"billing", "shared/money"}}
	if _, err := New(specs, tests, &fakeRunner{}, marks, scopes).Check(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if s := marks.seen[0].Scope; !s.Covers("shared/money/round.go") || s.Covers("payments/pay.go") {
		t.Fatalf("scope = %+v, want only billing, shared/money and the spec", s)
	}
}

func builtSpec(slice string, claims ...string) string {
	return strings.Replace(specOf(slice, claims...), "status: approved", "status: built", 1)
}

func TestCheckVerifiesABuiltSliceWhoseTestClaimsAllHold(t *testing.T) {
	specs := fakeSpecs{"billing": builtSpec("billing", passing, manual, untagged)}
	rep, err := newVerify(specs, &fakeRunner{}).Check(ctx, "")
	if err != nil || !reflect.DeepEqual(rep.Verified, []string{"billing"}) {
		t.Fatalf("Check = %+v, %v", rep, err)
	}
	if want := builtSpec("billing", passing, manual, untagged); specs["billing"] != strings.Replace(want, "status: built", "status: verified", 1) {
		t.Fatalf("stored %q", specs["billing"])
	}
}

func TestCheckLeavesTheStatusAloneUnlessItIsBuiltAndClean(t *testing.T) {
	cases := map[string]string{
		"a claim is violated":     builtSpec("billing", passing, failing),
		"a claim is unverified":   builtSpec("billing", passing, uncited),
		"no test claim":           builtSpec("billing", manual, untagged),
		"approved, not built yet": specOf("billing", passing),
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			specs := fakeSpecs{"billing": spec}
			run := &fakeRunner{exits: map[string]int{"go test ./failing": 1}}
			rep, err := New(specs, tests, run, nil, nil).Check(ctx, "")
			if err != nil || len(rep.Verified) != 0 || specs["billing"] != spec {
				t.Fatalf("Check = %+v, %v; stored %q", rep, err, specs["billing"])
			}
		})
	}
}

func TestCheckSaysWhichSlicesEndVerified(t *testing.T) {
	specs := fakeSpecs{
		"billing":  builtSpec("billing", passing),
		"payments": strings.Replace(builtSpec("payments", passing), "status: built", "status: verified", 1),
	}
	marks := &fakeMarks{}
	if _, err := New(specs, tests, &fakeRunner{}, marks, nil).Check(ctx, ""); err != nil {
		t.Fatal(err)
	}
	for _, s := range marks.seen {
		if !s.Verified {
			t.Errorf("%s is verified after the check but Seen says it is not", s.Slice)
		}
	}
}

func TestCheckDoesNotCallASliceVerifiedUnlessItHolds(t *testing.T) {
	specs := fakeSpecs{
		"billing":  builtSpec("billing", passing, failing),
		"payments": specOf("payments", passing),
	}
	marks := &fakeMarks{}
	run := &fakeRunner{exits: map[string]int{"go test ./failing": 1}}
	if _, err := New(specs, tests, run, marks, nil).Check(ctx, ""); err != nil {
		t.Fatal(err)
	}
	for _, s := range marks.seen {
		if s.Verified {
			t.Errorf("%s is not verified after the check but Seen says it is", s.Slice)
		}
	}
}
