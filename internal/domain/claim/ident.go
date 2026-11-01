// Package claim is the atom of a contract: one statement with a stable ID and a way to verify it.
package claim

import (
	"fmt"
	"strings"
)

// Kind says what a claim is, which decides the sentence notation it is written in.
type Kind int

const (
	// Invariant is written in EARS notation.
	Invariant Kind = iota + 1
	// Acceptance is written as Given/When/Then.
	Acceptance
)

func (k Kind) String() string {
	switch k {
	case Invariant:
		return "invariant"
	case Acceptance:
		return "acceptance"
	}
	return "unknown"
}

var prefixes = map[string]Kind{"INV": Invariant, "ACC": Acceptance}

// ID names a claim for good: tests cite it and the CLI links through it.
type ID struct {
	text string
	kind Kind
}

// ParseID accepts INV-<n> for invariants and ACC-<n> for acceptance criteria.
func ParseID(s string) (ID, error) {
	prefix, num, _ := strings.Cut(s, "-")
	kind, ok := prefixes[prefix]
	if !ok || num == "" || strings.Trim(num, "0123456789") != "" {
		return ID{}, fmt.Errorf("claim ID %q must look like INV-01 (invariant) or ACC-01 (acceptance)", s)
	}
	return ID{text: s, kind: kind}, nil
}

func (id ID) String() string { return id.text }

// Kind comes from the prefix; the zero ID has none.
func (id ID) Kind() Kind { return id.kind }
