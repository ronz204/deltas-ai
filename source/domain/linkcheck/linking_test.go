package linkcheck

import (
	"reflect"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
	"github.com/ronz204/deltas-ai/source/domain/specfile"
)

func id(t *testing.T, s string) assertion.ID {
	t.Helper()
	v, err := assertion.ParseID(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestIsTestFile(t *testing.T) {
	yes := []string{"total_test.go", "a/b/total_test.go", "total.test.ts", "src/x.spec.tsx", "e2e/consecutivo.e2e.test.ts", "test_total.py", "total_test.py", "x.test.mjs"}
	no := []string{"total.go", "test.go", "test_total.go", "invoicing.spec.md", "deltas/x.spec.md", "total.ts", "contest.ts", "x.test.md", "total.py", "testing.py"}
	for _, p := range yes {
		if !IsTestFile(p) {
			t.Errorf("%s should be a test file", p)
		}
	}
	for _, p := range no {
		if IsTestFile(p) {
			t.Errorf("%s should not be a test file", p)
		}
	}
}

func TestIDPattern(t *testing.T) {
	yes := []string{"INV-01", "// INV-01: rounds", "func TestINV01_Rounds(t *testing.T)", "def test_inv_01_rounds():", "it('INV-01: rounds'", "inv-01", "INV_01", "(INV-01)", "INV01"}
	no := []string{"", "INV-011", "INV-1", "INV-02", "ACC-01", "INV-0", "INV01234", "INV"}
	want := id(t, "INV-01")
	for _, s := range yes {
		if !idPattern(want).MatchString(s) {
			t.Errorf("%q should cite INV-01", s)
		}
	}
	for _, s := range no {
		if idPattern(want).MatchString(s) {
			t.Errorf("%q should not cite INV-01", s)
		}
	}
	if idPattern(id(t, "INV-1")).MatchString("INV-01") {
		t.Error("INV-1 and INV-01 are different IDs")
	}
}

func TestCitedIsScopedToWhatTheCommandRuns(t *testing.T) {
	files := []File{
		{Path: "billing/total_test.go", Text: "// INV-01"},
		{Path: "payments/pay_test.go", Text: "// nothing here"},
		{Path: "web/src/cart.test.ts", Text: "it('INV-01: adds', () => {})"},
	}
	inv := id(t, "INV-01")
	cases := []struct {
		name, command string
		want          bool
	}{
		{"the package that cites it", "go test ./billing", true},
		{"a package that does not", "go test ./payments", false},
		{"all packages", "go test ./...", true},
		{"the root", "go test .", true},
		{"a pattern under a package", "go test ./billing/...", true},
		{"windows separators", `go test .\billing`, true},
		{"a test file", "bunx vitest run web/src/cart.test.ts", true},
		{"another test file", "bunx vitest run web/src/other.test.ts", false},
		{"no path at all", "npm test", true},
		{"flags and values are not paths", "go test -run TestA/Sub -count=1 ./payments", false},
		{"two paths, one cites", "go test ./payments ./billing", true},
		{"outside the root", "go test ../billing", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Cited(inv, c.command, files); got != c.want {
				t.Fatalf("Cited(%q) = %v, want %v", c.command, got, c.want)
			}
		})
	}
	if Cited(inv, "npm test", nil) {
		t.Error("with no test files nothing is cited")
	}
}

func slice(t *testing.T, file string, claims ...assertion.Claim) Slice {
	t.Helper()
	return Slice{File: file, Spec: specfile.Spec{Claims: claims}}
}

func TestCheck(t *testing.T) {
	test := func(cmd string) assertion.Tag { return assertion.Tag{Method: assertion.Test, Arg: cmd} }
	slices := []Slice{
		slice(t, "deltas/billing.spec.md",
			assertion.Claim{ID: id(t, "INV-01"), Tag: test("go test ./billing"), Line: 7},
			assertion.Claim{ID: id(t, "INV-02"), Tag: test("go test ./billing"), Line: 9},
			assertion.Claim{ID: id(t, "INV-03"), Tag: assertion.Tag{Method: assertion.None}, Line: 11}),
		slice(t, "deltas/payments.spec.md",
			assertion.Claim{ID: id(t, "INV-01"), Tag: test("go test ./payments"), Line: 7}),
	}
	files := []File{
		{Path: "billing/total_test.go", Text: "// INV-01\nfunc TestINV01(t *testing.T) {}\n// INV-99 was removed\nt.Skip(\"later\")\n"},
		{Path: "payments/pay_test.go", Text: "// no claim cited here\nt.Skip(\"x\")\n"},
		{Path: "web/cart.test.ts", Text: "it.only('ACC-05 is gone', () => {})\n"},
	}
	got := Check(slices, files)
	type short struct {
		File, Code string
		Line       int
		Warn       bool
	}
	var have []short
	for _, d := range got {
		have = append(have, short{d.File, d.Code, d.Line, d.Warn})
	}
	want := []short{
		{"deltas/billing.spec.md", CodeNoCite, 9, false},  // INV-02: no test cites it
		{"deltas/payments.spec.md", CodeNoCite, 7, false}, // INV-01 is cited only under ./billing
		{"billing/total_test.go", CodeUnknownID, 3, true},
		{"billing/total_test.go", CodeSkip, 4, true},
		{"web/cart.test.ts", CodeUnknownID, 1, true},
		{"web/cart.test.ts", CodeSkip, 1, true},
	}
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("diagnostics:\n got %+v\nwant %+v", have, want)
	}
}

func TestCheckAcceptsACleanProject(t *testing.T) {
	slices := []Slice{slice(t, "deltas/a.spec.md", assertion.Claim{ID: id(t, "ACC-01"), Tag: assertion.Tag{Method: assertion.Test, Arg: "go test ./a"}, Line: 6})}
	files := []File{{Path: "a/a_test.go", Text: "func TestACC01(t *testing.T) {}\n"}}
	if got := Check(slices, files); len(got) != 0 {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}
