package linkcheck

import (
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
)

func testClaim(cmd string) assertion.Claim {
	return assertion.Claim{Tag: assertion.Tag{Method: assertion.Test, Arg: cmd}}
}

func TestScopeFollowsThePathsTheCommandsName(t *testing.T) {
	s := ScopeOf("deltas/billing.spec.md", []assertion.Claim{
		testClaim("go test ./billing/..."),
		testClaim("npx vitest run tests/billing/total.test.ts"),
		{Tag: assertion.Tag{Method: assertion.Manual, Arg: "look at it"}},
	})
	cases := map[string]bool{
		"billing/total.go":                 true,
		"billing/deep/x_test.go":           true,
		"tests/billing/total.test.ts":      true,
		"deltas/billing.spec.md":           true,
		"billing2/x.go":                    false,
		"payments/pay.go":                  false,
		"README.md":                        false,
		"billing/NOTES.md":                 false,
		"deltas/payments.spec.md":          false,
		"tests/billing/total.test.ts.snap": false,
	}
	for file, want := range cases {
		if got := s.Covers(file); got != want {
			t.Errorf("Covers(%q) = %v, want %v", file, got, want)
		}
	}
}

func TestScopeIsTheWholeProjectWhenACommandNamesNoPath(t *testing.T) {
	for _, cmds := range [][]string{{"npm test"}, {"go test ./..."}, {"go test ./billing", "make test"}, {"pytest ."}} {
		var claims []assertion.Claim
		for _, c := range cmds {
			claims = append(claims, testClaim(c))
		}
		s := ScopeOf("deltas/billing.spec.md", claims)
		if !s.Covers("anything/at/all.go") || s.Covers("docs/guide.md") || !s.Covers("deltas/billing.spec.md") {
			t.Errorf("%v: scope = %+v, want every code file but no documentation", cmds, s)
		}
	}
}

func TestScopeWithoutTestClaimsCoversEveryCodeFile(t *testing.T) {
	s := ScopeOf("deltas/billing.spec.md", nil)
	if !s.Covers("main.go") || s.Covers("README.md") {
		t.Fatalf("scope = %+v", s)
	}
}
