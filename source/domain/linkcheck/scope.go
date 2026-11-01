package linkcheck

import (
	"slices"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
)

// Scope is the part of a project a slice depends on: the paths its commands name, plus its own spec.
// Markdown is documentation and never counts, except the spec itself.
// shortcut: code a command reaches but does not name (a shared package) is outside the scope; upgrade with a
// list of extra paths per slice in delta.conf.
type Scope struct {
	paths []string // empty means every path
	spec  string
}

// ScopeOf is the scope of a spec file whose claims are given. A test command that names no path, or
// names ".", looks at the whole project, and so does the scope.
func ScopeOf(spec string, claims []assertion.Claim) Scope {
	s := Scope{spec: spec}
	for _, c := range claims {
		if c.Tag.Method != assertion.Test {
			continue
		}
		found := pathsIn(c.Tag.Arg)
		if len(found) == 0 || slices.Contains(found, ".") {
			return Scope{spec: spec}
		}
		s.paths = append(s.paths, found...)
	}
	slices.Sort(s.paths)
	s.paths = slices.Compact(s.paths)
	return s
}

// Covers says whether a change to file (relative to the project, with forward slashes) can affect the slice.
func (s Scope) Covers(file string) bool {
	switch {
	case file == s.spec:
		return true
	case strings.HasSuffix(file, ".md"):
		return false
	}
	return inScope(file, s.paths)
}
