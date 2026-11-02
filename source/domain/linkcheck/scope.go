package linkcheck

import "strings"

// Scope is the part of a project a slice depends on: every code file, since a change anywhere can break a claim.
// Markdown is documentation and never counts, except the slice's own spec. What delta itself installs and keeps
// (.claude/ and deltas/) is not code either: running delta update must not make every slice stale.
type Scope struct {
	spec string
}

// NewScope is the scope of a spec file.
func NewScope(spec string) Scope { return Scope{spec: spec} }

// Covers says whether a change to file (relative to the project, with forward slashes) can affect the slice.
func (s Scope) Covers(file string) bool {
	switch {
	case file == s.spec:
		return true
	case strings.HasSuffix(file, ".md"), isDelta(file):
		return false
	}
	return true
}

// isDelta says whether a file belongs to delta's own installation in the project root.
func isDelta(file string) bool {
	for _, dir := range []string{".claude/", "deltas/"} {
		if strings.HasPrefix(file, dir) {
			return true
		}
	}
	return false
}
