package states

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

var moves = map[Status][]Status{
	Draft:    {Approved},
	Approved: {Built, Verified},
	Built:    {Built, Verified},
	Verified: {Built},
	Stale:    {Built, Verified},
}

func ParseStatus(s string) (Status, error) {
	if st := Status(s); slices.Contains(statuses, st) {
		return st, nil
	}
	return "", fmt.Errorf("unknown status %q; want one of %v", s, statuses)
}

func Can(from, to Status) bool {
	if to == Draft {
		return from != Draft && slices.Contains(statuses, from)
	}
	return slices.Contains(moves[from], to)
}
