package linkcheck

import (
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
)

func TestScopeIsEveryCodeFileUnlessPathsAreDeclared(t *testing.T) {
	s := NewScope("deltas/billing.spec.md", nil)
	cases := map[string]bool{
		"billing/total.go":        true,
		"shared/money/round.go":   true,
		"deltas/billing.spec.md":  true,
		"README.md":               false,
		"docs/guide.md":           false,
		"deltas/payments.spec.md": false,
	}
	for file, want := range cases {
		if got := s.Covers(file); got != want {
			t.Errorf("Covers(%q) = %v, want %v", file, got, want)
		}
	}
}

func TestDeclaredPathsNarrowTheScope(t *testing.T) {
	s := NewScope("deltas/billing.spec.md", []string{"billing", "shared/money"})
	cases := map[string]bool{
		"billing/total.go":        true,
		"billing/deep/x_test.go":  true,
		"shared/money/round.go":   true,
		"deltas/billing.spec.md":  true,
		"billing2/x.go":           false,
		"payments/pay.go":         false,
		"shared/other.go":         false,
		"billing/NOTES.md":        false,
		"deltas/payments.spec.md": false,
	}
	for file, want := range cases {
		if got := s.Covers(file); got != want {
			t.Errorf("Covers(%q) = %v, want %v", file, got, want)
		}
	}
}

func TestTestMatcher(t *testing.T) {
	own := TestMatcher(nil)
	for file, want := range map[string]bool{"a/b_test.go": true, "x/test_a.py": true, "a.test.ts": true, "a.go": false, "deltas/a.spec.md": false} {
		if got := own(file); got != want {
			t.Errorf("default: %q = %v, want %v", file, got, want)
		}
	}
	custom := TestMatcher([]string{"*_spec.rb", "*Test.java"})
	for file, want := range map[string]bool{
		"spec/models/user_spec.rb": true,
		"src/UserTest.java":        true,
		"a_test.go":                false, // the patterns replace the defaults
		"docs/user_spec.rb.md":     false,
		"src/User.java":            false,
	} {
		if got := custom(file); got != want {
			t.Errorf("custom: %q = %v, want %v", file, got, want)
		}
	}
}

func TestUnderNamesFilesFromTheFolderACommandRunsIn(t *testing.T) {
	files := []File{{Path: "a/x_test.go", Text: "1"}, {Path: "svc/auth/y_test.go", Text: "2"}, {Path: "svc/other/z_test.go", Text: "3"}}
	if got := Under(files, ""); len(got) != 3 {
		t.Fatalf("the root keeps every file: %v", got)
	}
	got := Under(files, "svc/auth")
	if len(got) != 1 || got[0].Path != "y_test.go" || got[0].Text != "2" {
		t.Fatalf("Under = %v", got)
	}
	id, _ := assertion.ParseID("INV-01")
	cited := []File{{Path: "svc/auth/login/login_test.go", Text: "// INV-01"}}
	if !Cited(id, "go test ./login", Under(cited, "svc/auth")) || Cited(id, "go test ./login", cited) {
		t.Fatal("a path in the command is relative to the folder it runs from")
	}
}
