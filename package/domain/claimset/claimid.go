// Package claimset is one statement of a contract, with a stable ID and a way to verify it.
package claimset

import (
	"fmt"
	"strings"
)

type Kind int

const (
	Invariant Kind = iota + 1
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

type ID struct {
	text string
	kind Kind
}

func ParseID(s string) (ID, error) {
	prefix, num, _ := strings.Cut(s, "-")
	kind, ok := prefixes[prefix]
	if !ok || num == "" || strings.Trim(num, "0123456789") != "" {
		return ID{}, fmt.Errorf("claim ID %q must look like INV-01 (invariant) or ACC-01 (acceptance)", s)
	}
	return ID{text: s, kind: kind}, nil
}

func (id ID) String() string { return id.text }

func (id ID) Kind() Kind { return id.kind }
