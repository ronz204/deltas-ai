// Package proof turns what a verification run observed into a verdict on a claim.
package proof

import (
	"fmt"

	"github.com/ronz204/deltas-ai/internal/domain/claim"
)

// Verdict is how a claim stands after a check.
type Verdict int

const (
	// Holds means the claim's command passed and a test it runs cites the claim.
	Holds Verdict = iota + 1
	// Violated means the claim's command failed.
	Violated
	// Unverified means there is no evidence either way. It never passes for Holds.
	Unverified
)

func (v Verdict) String() string {
	switch v {
	case Holds:
		return "Holds"
	case Violated:
		return "Violated"
	case Unverified:
		return "Unverified"
	}
	return "unknown"
}

// Outcome is a verdict with the reason a person needs to act on it.
type Outcome struct {
	Verdict Verdict
	Reason  string // empty for Holds
}

// Judge decides a claim from its tag, the exit code of its command and whether a test that
// command runs cites the claim. exit and cited only matter for a Test tag.
func Judge(tag claim.Tag, exit int, cited bool) Outcome {
	switch {
	case tag.Kind == claim.Manual && tag.Arg != "":
		return Outcome{Unverified, "checked by hand: " + tag.Arg}
	case tag.Kind == claim.Manual:
		return Outcome{Unverified, "checked by hand"}
	case tag.Kind != claim.Test:
		return Outcome{Unverified, "nothing verifies this claim yet"}
	case exit != 0:
		return Outcome{Violated, fmt.Sprintf("the command exited %d", exit)}
	case !cited:
		return Outcome{Unverified, "the command exited 0 but no test it runs cites the claim ID"}
	}
	return Outcome{Verdict: Holds}
}
