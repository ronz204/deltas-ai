// Package slice describes where a slice is in its life. Which moves between statuses are legal
// is the state machine's job and arrives with it.
package slice

import (
	"fmt"
	"slices"
)

// Status is the stage a slice's contract has reached.
type Status string

const (
	Draft    Status = "draft"
	Approved Status = "approved"
	Built    Status = "built"
	Verified Status = "verified"
	Stale    Status = "stale"
)

var statuses = []Status{Draft, Approved, Built, Verified, Stale}

// ParseStatus turns the text of a frontmatter value into a Status.
func ParseStatus(s string) (Status, error) {
	if st := Status(s); slices.Contains(statuses, st) {
		return st, nil
	}
	return "", fmt.Errorf("unknown status %q; want one of %v", s, statuses)
}
