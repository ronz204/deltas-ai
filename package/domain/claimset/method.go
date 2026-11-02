package claimset

import (
	"fmt"
	"strings"
)

type Method int

const (
	Test Method = iota + 1
	Manual
	None
)

type Tag struct {
	Method Method
	Arg    string // the command for Test, the optional step for Manual, empty for None
}

func (t Tag) String() string {
	switch t.Method {
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

func ParseTag(s string) (Tag, error) {
	name, arg, _ := strings.Cut(s, ":")
	name, arg = strings.TrimSpace(name), strings.TrimSpace(arg)
	switch name {
	case "test":
		if arg == "" {
			return Tag{}, fmt.Errorf("tag %q needs a command, as in [test: go test ./...]", s)
		}
		return Tag{Method: Test, Arg: arg}, nil
	case "manual":
		return Tag{Method: Manual, Arg: arg}, nil
	case "none":
		if arg != "" {
			return Tag{}, fmt.Errorf("tag %q takes no argument", s)
		}
		return Tag{Method: None}, nil
	}
	return Tag{}, fmt.Errorf("unknown tag %q; use test: <command>, manual or none", s)
}
