// Package judgment turns what a verification run observed into a verdict on a claim.
package judgment

import (
	"fmt"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
)

type Verdict int

const (
	Holds Verdict = iota + 1
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

type Outcome struct {
	Verdict Verdict
	Reason  string // empty for Holds
}

// Judge decides a claim; exit and cited only matter for a Test tag.
func Judge(tag assertion.Tag, exit int, cited bool) Outcome {
	switch {
	case tag.Kind == assertion.Manual && tag.Arg != "":
		return Outcome{Unverified, "checked by hand: " + tag.Arg}
	case tag.Kind == assertion.Manual:
		return Outcome{Unverified, "checked by hand"}
	case tag.Kind != assertion.Test:
		return Outcome{Unverified, "nothing verifies this claim yet"}
	case exit != 0:
		return Outcome{Violated, fmt.Sprintf("the command exited %d", exit)}
	case !cited:
		return Outcome{Unverified, "the command exited 0 but no test it runs cites the claim ID"}
	}
	return Outcome{Verdict: Holds}
}
