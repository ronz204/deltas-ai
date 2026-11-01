// Package slice describes where a slice is in its life. Which moves between statuses are legal
// is the state machine's job and arrives with it.
package slice

import (
	"fmt"
	"strings"
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
	for _, st := range statuses {
		if string(st) == s {
			return st, nil
		}
	}
	names := make([]string, len(statuses))
	for i, st := range statuses {
		names[i] = string(st)
	}
	return "", fmt.Errorf("unknown status %q; want one of %s", s, strings.Join(names, ", "))
}
