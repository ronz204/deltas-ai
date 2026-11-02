package judgment

import (
	"fmt"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
)

type Verdict int

const (
	Holds Verdict = iota + 1
	Violated

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
	Reason  string
}

func Judge(tag claimset.Tag, exit int, cited bool) Outcome {
	switch {
	case tag.Method == claimset.Manual:
		reason := "checked by hand"
		if tag.Arg != "" {
			reason += ": " + tag.Arg
		}
		return Outcome{Unverified, reason}
	case tag.Method != claimset.Test:
		return Outcome{Unverified, "nothing verifies this claim yet"}
	case exit != 0:
		return Outcome{Violated, fmt.Sprintf("the command exited %d", exit)}
	case !cited:
		return Outcome{Unverified, "the command exited 0 but no test it runs cites the claim ID"}
	}
	return Outcome{Verdict: Holds}
}
