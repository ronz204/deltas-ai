package linkcheck

import (
	"testing"
)

func TestScopeIsEveryCodeFile(t *testing.T) {
	s := NewScope("deltas/specs/billing.spec.md")
	cases := map[string]bool{
		"billing/total.go":              true,
		"shared/money/round.go":         true,
		"deltas/specs/billing.spec.md":  true,
		"README.md":                     false,
		"docs/guide.md":                 false,
		"deltas/specs/payments.spec.md": false,
		"deltas/budget.conf":            false,
		"deltas/deltas.manifest":        false,
		".claude/settings.json":         false,
		"go.mod":                        true,
		"src/deltas/x.go":               true,
	}
	for file, want := range cases {
		if got := s.Covers(file); got != want {
			t.Errorf("Covers(%q) = %v, want %v", file, got, want)
		}
	}
}

func TestTestMatcher(t *testing.T) {
	own := TestMatcher(nil)
	for file, want := range map[string]bool{"a/b_test.go": true, "x/test_a.py": true, "a.test.ts": true, "a.go": false, "deltas/specs/a.spec.md": false} {
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
