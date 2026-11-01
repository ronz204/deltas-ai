// Package lifecycle holds the status of a slice; the legal moves between statuses arrive with the state machine.
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

func ParseStatus(s string) (Status, error) {
	if st := Status(s); slices.Contains(statuses, st) {
		return st, nil
	}
	return "", fmt.Errorf("unknown status %q; want one of %v", s, statuses)
}
