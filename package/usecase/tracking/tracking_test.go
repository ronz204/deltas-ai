package tracking

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ronz204/deltas-ai/package/domain/budget"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/domain/states"
)

type fake map[string]string

func (f fake) Path(s string) string { return "deltas/specs/" + s + ".spec.md" }
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
	rows, err := New(specs).Report(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []Row{
		{Slice: "a", Status: states.Approved, Claims: 4, Test: 1, Manual: 1, None: 2},
		{Slice: "b", Status: states.Verified},
		{Slice: "c", Issues: 1},
	}
	scopes := make([]linking.Scope, len(rows))
	for i := range rows {
		scopes[i], rows[i].Scope = rows[i].Scope, linking.Scope{}
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("rows:\n got %+v\nwant %+v", rows, want)
	}
	if a := scopes[0]; !a.Covers("a/x.go") || a.Covers("b/x.go") || a.Covers("README.md") || !a.Covers("deltas/specs/a.spec.md") {
		t.Fatalf("the scope of a = %+v, want only the code under ./a and its spec", a)
	}
}

func TestReportOfAnEmptyProject(t *testing.T) {
	rows, err := New(fake{}).Report(context.Background())
	if err != nil || len(rows) != 0 {
		t.Fatalf("Report = %v, %v", rows, err)
	}
}

func TestReadyForACleanSliceThatIsNotADraft(t *testing.T) {
	spec := func(status string) string {
		return "---\nformat: 1\nslice: a\nstatus: " + status + "\n---\n- INV-01 [none]\n  When x, y.\n"
	}
	cases := map[string]struct {
		src   string
		ready bool
	}{
		"approved": {spec("approved"), true},
		"built":    {spec("built"), true},
		"verified": {spec("verified"), true},
		"stale":    {spec("stale"), true},
		"draft":    {spec("draft"), false},
		"broken":   {"not a spec", false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err := New(fake{"a": c.src}).Ready(context.Background(), "a")
			var notReady *NotReady
			if c.ready != (err == nil) || (!c.ready && !errors.As(err, &notReady)) {
				t.Fatalf("Ready = %v, want ready = %v", err, c.ready)
			}
		})
	}
}

func TestReadyFailsPlainlyForAMissingSpec(t *testing.T) {
	err := New(fake{}).Ready(context.Background(), "a")
	var notReady *NotReady
	if err == nil || errors.As(err, &notReady) {
		t.Fatalf("Ready = %v, want a plain error", err)
	}
}

type drifts map[string][]string

func (d drifts) Drift(_ context.Context, slice string, _ linking.Scope) ([]string, bool, error) {
	files, ok := d[slice]
	return files, ok, nil
}

func spec(slice, state string) string {
	return "---\nformat: 1\nslice: " + slice + "\nstatus: " + state + "\n---\n- INV-01 [none]\n  When x, y.\n"
}

func TestDriftsListsAVerifiedSliceWhoseFilesChangedAndNeverWritesItsSpec(t *testing.T) {
	specs := fake{"a": spec("a", "verified"), "b": spec("b", "verified")}
	r := New(specs).WithDrift(drifts{"a": {"a/x.go"}, "b": {}})
	got, err := r.Drifts(context.Background())
	if want := []Drift{{Slice: "a", Files: []string{"a/x.go"}}}; err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Drifts = %+v, %v; want %+v", got, err, want)
	}
	if specs["a"] != spec("a", "verified") || specs["b"] != spec("b", "verified") {
		t.Fatalf("stored a = %q, b = %q; telling the drift must not change a spec", specs["a"], specs["b"])
	}
}

func TestDriftsKeepsReportingASliceMarkedStaleByAnOlderDelta(t *testing.T) {
	specs := fake{"a": spec("a", "stale")}
	r := New(specs).WithDrift(drifts{"a": {"a/x.go", "deltas/specs/a.spec.md"}})
	got, err := r.Drifts(context.Background())
	if want := []Drift{{Slice: "a", Files: []string{"a/x.go"}}}; err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Drifts = %+v, %v; want %+v (its own spec is not news)", got, err, want)
	}
}

func TestDriftsLeavesAloneWhatWasNeverVerifiedOrIsNotVerified(t *testing.T) {
	specs := fake{"a": spec("a", "verified"), "b": spec("b", "built"), "c": "not a spec"}
	r := New(specs).WithDrift(drifts{"b": {"b/x.go"}, "c": {"c/x.go"}})
	if got, err := r.Drifts(context.Background()); err != nil || len(got) != 0 {
		t.Fatalf("Drifts = %+v, %v; a slice with no baseline, one still being built and a broken spec have no drift", got, err)
	}
}

func TestDriftsIsOffUntilItIsWiredIn(t *testing.T) {
	specs := fake{"a": spec("a", "verified")}
	if got, err := New(specs).Drifts(context.Background()); err != nil || got != nil {
		t.Fatalf("Drifts = %+v, %v", got, err)
	}
}

type brokenDrift struct{}

func (brokenDrift) Drift(context.Context, string, linking.Scope) ([]string, bool, error) {
	return nil, true, errors.New("git is gone")
}

func TestDriftsReportsWhyItCouldNotTell(t *testing.T) {
	specs := fake{"a": spec("a", "verified")}
	if _, err := New(specs).WithDrift(brokenDrift{}).Drifts(context.Background()); err == nil {
		t.Fatal("Drifts hid an error")
	}
}

func TestReportShowsTheDriftOfAVerifiedSliceAndStillWorksWhenDriftCannotBeTold(t *testing.T) {
	specs := fake{"a": spec("a", "verified")}
	rows, err := New(specs).WithDrift(drifts{"a": {"a/x.go"}}).Report(context.Background())
	if err != nil || rows[0].Status != states.Verified || !reflect.DeepEqual(rows[0].Changed, []string{"a/x.go"}) {
		t.Fatalf("rows = %+v, %v; want a verified slice with its changed file", rows, err)
	}
	rows, err = New(specs).WithDrift(brokenDrift{}).Report(context.Background())
	if err != nil || rows[0].Status != states.Verified || len(rows[0].Changed) != 0 {
		t.Fatalf("rows = %+v, %v; a status report must not fail because git did", rows, err)
	}
}

type docFiles []budget.File

func (d docFiles) Glob(context.Context, []string) ([]budget.File, error) { return d, nil }

func TestDocsCountsTheWrittenOnesAndNamesTheNextInOrder(t *testing.T) {
	files := docFiles{
		{Path: "deltas/docs/approach.md", Text: "# Approach\n\n<!-- todo -->\n"},
		{Path: "deltas/docs/overview.md", Text: "# Overview\n\nA real doc.\n"},
		{Path: "deltas/docs/structure.md", Text: "# Structure\n\n<!-- todo -->\n"},
		{Path: "deltas/docs/expertise.md", Text: "# Expertise\n\nDone.\n"},
		{Path: "deltas/docs/deep/other.md", Text: "<!-- not a doc -->"},
	}
	got, err := New(fake{}).WithDocs(files).Docs(context.Background())
	if want := (DocsRow{Total: 4, Filled: 2, Next: "structure"}); err != nil || got != want {
		t.Fatalf("Docs = %+v, %v; want %+v", got, err, want)
	}
}

func TestDocsHasNoNextWhenAllAreWrittenOrThereAreNone(t *testing.T) {
	done := docFiles{{Path: "deltas/docs/overview.md", Text: "ok"}}
	if got, _ := New(fake{}).WithDocs(done).Docs(context.Background()); got != (DocsRow{Total: 1, Filled: 1}) {
		t.Errorf("Docs = %+v", got)
	}
	if got, err := New(fake{}).WithDocs(docFiles{}).Docs(context.Background()); err != nil || got != (DocsRow{}) {
		t.Errorf("Docs of a project with none = %+v, %v", got, err)
	}
	if got, err := New(fake{}).Docs(context.Background()); err != nil || got != (DocsRow{}) {
		t.Errorf("Docs with nothing to read files from = %+v, %v", got, err)
	}
}

func TestDocsNamesAnExtraTemplateWhenTheUsualOnesAreDone(t *testing.T) {
	files := docFiles{{Path: "deltas/docs/overview.md", Text: "ok"}, {Path: "deltas/docs/security.md", Text: "<!-- x -->"}}
	if got, _ := New(fake{}).WithDocs(files).Docs(context.Background()); got.Next != "security" {
		t.Errorf("Docs = %+v, want security next", got)
	}
}
