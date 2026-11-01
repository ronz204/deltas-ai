package linkcheck

import (
	"testing"
)

func TestScopeIsEveryCodeFileUnlessPathsAreDeclared(t *testing.T) {
	s := NewScope("deltas/specifies/billing.spec.md", nil)
	cases := map[string]bool{
		"billing/total.go":                  true,
		"shared/money/round.go":             true,
		"deltas/specifies/billing.spec.md":  true,
		"README.md":                         false,
		"docs/guide.md":                     false,
		"deltas/specifies/payments.spec.md": false,
	}
	for file, want := range cases {
		if got := s.Covers(file); got != want {
			t.Errorf("Covers(%q) = %v, want %v", file, got, want)
		}
	}
}

func TestDeclaredPathsNarrowTheScope(t *testing.T) {
	s := NewScope("deltas/specifies/billing.spec.md", []string{"billing", "shared/money"})
	cases := map[string]bool{
		"billing/total.go":                  true,
		"billing/deep/x_test.go":            true,
		"shared/money/round.go":             true,
		"deltas/specifies/billing.spec.md":  true,
		"billing2/x.go":                     false,
		"payments/pay.go":                   false,
		"shared/other.go":                   false,
		"billing/NOTES.md":                  false,
		"deltas/specifies/payments.spec.md": false,
	}
	for file, want := range cases {
		if got := s.Covers(file); got != want {
			t.Errorf("Covers(%q) = %v, want %v", file, got, want)
		}
	}
}

func TestTestMatcher(t *testing.T) {
	own := TestMatcher(nil)
	for file, want := range map[string]bool{"a/b_test.go": true, "x/test_a.py": true, "a.test.ts": true, "a.go": false, "deltas/specifies/a.spec.md": false} {
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
