package linking

import (
	"testing"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
)

func TestScopeIsEveryCodeFile(t *testing.T) {
	s := NewScope("deltas/specs/billing.spec.md", nil)
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

func TestScopeFollowsThePathsOfTheTestCommands(t *testing.T) {
	test := func(command string) claimset.Claim {
		return claimset.Claim{Tag: claimset.Tag{Method: claimset.Test, Arg: command}}
	}
	cases := []struct {
		name   string
		claims []claimset.Claim
		file   string
		want   bool
	}{
		{"a named package", []claimset.Claim{test("go test ./billing")}, "billing/total.go", true},
		{"another package", []claimset.Claim{test("go test ./billing")}, "payments/pay.go", false},
		{"the package and below", []claimset.Claim{test("go test ./billing/...")}, "billing/deep/x.go", true},
		{"two commands", []claimset.Claim{test("go test ./billing"), test("go test ./shared/money")}, "shared/money/round.go", true},
		{"a command with no path", []claimset.Claim{test("go test ./billing"), test("make test")}, "payments/pay.go", true},
		{"the whole project", []claimset.Claim{test("go test ./...")}, "payments/pay.go", true},
		{"no test claim", []claimset.Claim{{Tag: claimset.Tag{Method: claimset.Manual}}}, "payments/pay.go", true},
		{"documentation never counts", []claimset.Claim{test("make test")}, "billing/NOTES.md", false},
	}
	for _, c := range cases {
		if got := NewScope("deltas/specs/billing.spec.md", c.claims).Covers(c.file); got != c.want {
			t.Errorf("%s: Covers(%q) = %v, want %v", c.name, c.file, got, c.want)
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
		"a_test.go":                false,
		"docs/user_spec.rb.md":     false,
		"src/User.java":            false,
	} {
		if got := custom(file); got != want {
			t.Errorf("custom: %q = %v, want %v", file, got, want)
		}
	}
}
