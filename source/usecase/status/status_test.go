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

func (f fake) Path(s string) string { return "deltas/" + s + ".spec.md" }
func (f fake) List(context.Context) ([]string, error) {
	var names []string
	for _, n := range []string{"a", "b", "c"} {
		if _, ok := f[n]; ok {
			names = append(names, n)
		}
	}
	return names, nil
}
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
	if a := scopes[0]; !a.Covers("a/x.go") || !a.Covers("b/x.go") || a.Covers("README.md") || !a.Covers("deltas/a.spec.md") {
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
		t.Errorf("b = %+v, want every code file: delta.conf says nothing about it", b)
	}
}

func TestReportOfAnEmptyProject(t *testing.T) {
	rows, err := New(fake{}, nil).Report(context.Background())
	if err != nil || len(rows) != 0 {
		t.Fatalf("Report = %v, %v", rows, err)
	}
}
