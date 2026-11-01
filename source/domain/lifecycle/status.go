// Package lifecycle holds the status of a slice and the moves the CLI allows between statuses.
package lifecycle

import (
	"fmt"
	"slices"
)

type Status string

const (
	Draft    Status = "draft"
	Approved Status = "approved"
	Built    Status = "built"
	Verified Status = "verified"
	Stale    Status = "stale"
)

var statuses = []Status{Draft, Approved, Built, Verified, Stale}

// moves lists, for each status, where a slice can go next. Any status can also fall back to draft (see Can).
var moves = map[Status][]Status{
	Draft:    {Approved},
	Approved: {Built},
	Built:    {Verified},
	Verified: {Stale},
	Stale:    {Built},
}

func ParseStatus(s string) (Status, error) {
	if st := Status(s); slices.Contains(statuses, st) {
		return st, nil
	}
	return "", fmt.Errorf("unknown status %q; want one of %v", s, statuses)
}

// Can says whether a slice may go from one status to another. A changed contract is a draft again, so every
// status but draft may return to it.
func Can(from, to Status) bool {
	if to == Draft {
		return from != Draft && slices.Contains(statuses, from)
	}
	return slices.Contains(moves[from], to)
}
