package judgment

import (
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
)

func TestJudge(t *testing.T) {
	test := assertion.Tag{Kind: assertion.Test, Arg: "go test ./x"}
	cases := []struct {
		name  string
		tag   assertion.Tag
		exit  int
		cited bool
		want  Verdict
	}{
		{"passes and is cited", test, 0, true, Holds},
		{"passes but nothing cites it", test, 0, false, Unverified},
		{"fails", test, 1, true, Violated},
		{"fails and nothing cites it", test, 2, false, Violated},
		{"manual", assertion.Tag{Kind: assertion.Manual}, 0, true, Unverified},
		{"manual with a step", assertion.Tag{Kind: assertion.Manual, Arg: "open it"}, 0, true, Unverified},
		{"none", assertion.Tag{Kind: assertion.None}, 0, true, Unverified},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Judge(c.tag, c.exit, c.cited)
			if got.Verdict != c.want {
				t.Fatalf("got %v (%s), want %v", got.Verdict, got.Reason, c.want)
			}
			if (got.Verdict == Holds) != (got.Reason == "") {
				t.Fatalf("reason %q does not fit verdict %v: only Holds comes without one", got.Reason, got.Verdict)
			}
		})
	}
}
