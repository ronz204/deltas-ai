package status

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/lifecycle"
	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
)

type fake map[string]string

func (f fake) Path(s string) string { return "deltas/specifies/" + s + ".spec.md" }
func (f fake) List(context.Context) ([]string, error) {
	var names []string
	for _, n := range []string{"a", "b", "c"} {
		if _, ok := f[n]; ok {
			names = append(names, n)
		}
	}
	return names, nil
}
func (f fake) Save(_ context.Context, s string, c []byte) error { f[s] = string(c); return nil }
func (f fake) Load(_ context.Context, s string) ([]byte, error) {
	if src, ok := f[s]; ok {
		return []byte(src), nil
	}
	return nil, errors.New("missing")
}

func TestReport(t *testing.T) {
	specs := fake{
		"a": "---\nformat: 1\nslice: a\nstatus: approved\n---\n" +
			"- INV-01 [test: go test ./a]\n  When x, y.\n" +
			"- INV-02 [manual]\n  When x, y.\n" +
			"- INV-03 [none]\n  When x, y.\n" +
			"- ACC-01 [none]\n  Given a, when b, then c.\n",
		"b": "---\nformat: 1\nslice: b\nstatus: verified\n---\n",
		"c": "not a spec",
	}
	rows, err := New(specs, nil).Report(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []Row{
		{Slice: "a", Status: lifecycle.Approved, Claims: 4, Test: 1, Manual: 1, None: 2},
		{Slice: "b", Status: lifecycle.Verified},
		{Slice: "c", Issues: 1},
	}
	scopes := make([]linkcheck.Scope, len(rows))
	for i := range rows {
		scopes[i], rows[i].Scope = rows[i].Scope, linkcheck.Scope{}
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("rows:\n got %+v\nwant %+v", rows, want)
	}
	if a := scopes[0]; !a.Covers("a/x.go") || !a.Covers("b/x.go") || a.Covers("README.md") || !a.Covers("deltas/specifies/a.spec.md") {
		t.Fatalf("the scope of a = %+v, want every code file and its spec", a)
	}
}

func TestReportNarrowsOnlyTheSlicesDeltaConfLists(t *testing.T) {
	specs := fake{
		"a": "---\nformat: 1\nslice: a\nstatus: approved\n---\n",
		"b": "---\nformat: 1\nslice: b\nstatus: approved\n---\n",
	}
	rows, err := New(specs, map[string][]string{"a": {"a", "shared"}}).Report(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if a := rows[0].Scope; !a.Covers("shared/x.go") || a.Covers("b/x.go") {
		t.Errorf("a = %+v, want only a, shared and its spec", a)
	}
	if b := rows[1].Scope; !b.Covers("a/x.go") {
		t.Errorf("b = %+v, want every code file: deltas.config says nothing about it", b)
	}
}

func TestReportOfAnEmptyProject(t *testing.T) {
	rows, err := New(fake{}, nil).Report(context.Background())
	if err != nil || len(rows) != 0 {
		t.Fatalf("Report = %v, %v", rows, err)
	}
}

func TestReadyOnlyForACleanApprovedOrStaleSlice(t *testing.T) {
	spec := func(status string) string {
		return "---\nformat: 1\nslice: a\nstatus: " + status + "\n---\n- INV-01 [none]\n  When x, y.\n"
	}
	cases := map[string]struct {
		src   string
		ready bool
	}{
		"approved": {spec("approved"), true},
		"stale":    {spec("stale"), true},
		"draft":    {spec("draft"), false},
		"built":    {spec("built"), false},
		"verified": {spec("verified"), false},
		"broken":   {"not a spec", false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err := New(fake{"a": c.src}, nil).Ready(context.Background(), "a")
			var notReady *NotReady
			if c.ready != (err == nil) || (!c.ready && !errors.As(err, &notReady)) {
				t.Fatalf("Ready = %v, want ready = %v", err, c.ready)
			}
		})
	}
}

func TestReadyFailsPlainlyForAMissingSpec(t *testing.T) {
	err := New(fake{}, nil).Ready(context.Background(), "a")
	var notReady *NotReady
	if err == nil || errors.As(err, &notReady) {
		t.Fatalf("Ready = %v, want a plain error", err)
	}
}

// drifts says, per slice, which files changed since it was verified; a slice it does not list was never verified here.
type drifts map[string][]string

func (d drifts) Drift(_ context.Context, slice string, _ linkcheck.Scope) ([]string, bool, error) {
	files, ok := d[slice]
	return files, ok, nil
}

func spec(slice, state string) string {
	return "---\nformat: 1\nslice: " + slice + "\nstatus: " + state + "\n---\n- INV-01 [none]\n  When x, y.\n"
}

func TestStaleMarksAVerifiedSliceWhoseFilesChanged(t *testing.T) {
	specs := fake{"a": spec("a", "verified"), "b": spec("b", "verified")}
	r := New(specs, nil).WithDrift(drifts{"a": {"a/x.go"}, "b": {}}, specs)
	got, err := r.Stale(context.Background())
	if want := []Drift{{Slice: "a", Files: []string{"a/x.go"}, Marked: true}}; err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Stale = %+v, %v; want %+v", got, err, want)
	}
	if specs["a"] != spec("a", "stale") || specs["b"] != spec("b", "verified") {
		t.Fatalf("stored a = %q, b = %q; only a goes stale, and only its status line changes", specs["a"], specs["b"])
	}
}

func TestStaleKeepsReportingAStaleSliceWithoutMarkingItAgain(t *testing.T) {
	specs := fake{"a": spec("a", "stale")}
	r := New(specs, nil).WithDrift(drifts{"a": {"a/x.go", "deltas/specifies/a.spec.md"}}, specs)
	got, err := r.Stale(context.Background())
	if want := []Drift{{Slice: "a", Files: []string{"a/x.go"}}}; err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Stale = %+v, %v; want %+v (its own spec is not news)", got, err, want)
	}
}

func TestStaleLeavesAloneWhatWasNeverVerifiedOrIsNotVerified(t *testing.T) {
	specs := fake{"a": spec("a", "verified"), "b": spec("b", "built"), "c": "not a spec"}
	r := New(specs, nil).WithDrift(drifts{"b": {"b/x.go"}, "c": {"c/x.go"}}, specs)
	if got, err := r.Stale(context.Background()); err != nil || len(got) != 0 {
		t.Fatalf("Stale = %+v, %v; a slice with no baseline, one still being built and a broken spec are never stale", got, err)
	}
	if specs["a"] != spec("a", "verified") || specs["b"] != spec("b", "built") {
		t.Fatalf("a spec was rewritten: %v", specs)
	}
}

func TestStaleIsOffUntilDriftIsWiredIn(t *testing.T) {
	specs := fake{"a": spec("a", "verified")}
	if got, err := New(specs, nil).Stale(context.Background()); err != nil || got != nil {
		t.Fatalf("Stale = %+v, %v", got, err)
	}
}

type brokenDrift struct{}

func (brokenDrift) Drift(context.Context, string, linkcheck.Scope) ([]string, bool, error) {
	return nil, true, errors.New("git is gone")
}

func TestStaleReportsWhyItCouldNotTell(t *testing.T) {
	specs := fake{"a": spec("a", "verified")}
	if _, err := New(specs, nil).WithDrift(brokenDrift{}, specs).Stale(context.Background()); err == nil {
		t.Fatal("Stale hid an error")
	}
}

func TestReportShowsWhatWentStaleAndStillWorksWhenDriftCannotBeTold(t *testing.T) {
	specs := fake{"a": spec("a", "verified")}
	rows, err := New(specs, nil).WithDrift(drifts{"a": {"a/x.go"}}, specs).Report(context.Background())
	if err != nil || rows[0].Status != lifecycle.Stale || !reflect.DeepEqual(rows[0].Changed, []string{"a/x.go"}) {
		t.Fatalf("rows = %+v, %v; want a stale with its changed file", rows, err)
	}
	specs = fake{"a": spec("a", "verified")}
	rows, err = New(specs, nil).WithDrift(brokenDrift{}, specs).Report(context.Background())
	if err != nil || rows[0].Status != lifecycle.Verified {
		t.Fatalf("rows = %+v, %v; a status report must not fail because git did", rows, err)
	}
}
