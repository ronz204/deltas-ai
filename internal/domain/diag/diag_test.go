package diag

import "testing"

func TestString(t *testing.T) {
	cases := []struct {
		name string
		d    Diagnostic
		want string
	}{
		{"with a line", Diagnostic{File: "a.spec.md", Line: 14, Code: "SPEC-TAG", Message: "bad tag"}, "a.spec.md:14: SPEC-TAG: bad tag"},
		{"whole file", Diagnostic{File: "a.spec.md", Code: "SPEC-KEY", Message: "missing"}, "a.spec.md: SPEC-KEY: missing"},
		{"warning", Diagnostic{File: "a_test.go", Line: 3, Code: "LINT-SKIP", Message: "skipped", Warn: true}, "a_test.go:3: warning LINT-SKIP: skipped"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.d.String(); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
