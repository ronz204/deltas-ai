package claim

import (
	"fmt"
	"strings"
)

// TagKind is how a claim is verified.
type TagKind int

const (
	// Test names a command that exits 0 when the claim holds, unit or end to end.
	Test TagKind = iota + 1
	// Manual is checked by a person; Arg may describe the step.
	Manual
	// None means nothing verifies the claim yet; it stays visible as verification debt.
	None
)

// Tag is the verification tag between the brackets of a claim header.
type Tag struct {
	Kind TagKind
	Arg  string // the command for Test, the optional step for Manual, empty for None
}

// String is the text that goes between the brackets, so ParseTag(t.String()) gives t back.
func (t Tag) String() string {
	switch t.Kind {
	case Test:
		return "test: " + t.Arg
	case Manual:
		if t.Arg != "" {
			return "manual: " + t.Arg
		}
		return "manual"
	}
	return "none"
}

// ParseTag reads the text between the brackets: "test: <command>", "manual", "manual: <step>"
// or "none".
func ParseTag(s string) (Tag, error) {
	name, arg, _ := strings.Cut(s, ":")
	name, arg = strings.TrimSpace(name), strings.TrimSpace(arg)
	switch name {
	case "test":
		if arg == "" {
			return Tag{}, fmt.Errorf("tag %q needs a command, as in [test: go test ./...]", s)
		}
		return Tag{Kind: Test, Arg: arg}, nil
	case "manual":
		return Tag{Kind: Manual, Arg: arg}, nil
	case "none":
		if arg != "" {
			return Tag{}, fmt.Errorf("tag %q takes no argument", s)
		}
		return Tag{Kind: None}, nil
	}
	return Tag{}, fmt.Errorf("unknown tag %q; use test: <command>, manual or none", s)
}
