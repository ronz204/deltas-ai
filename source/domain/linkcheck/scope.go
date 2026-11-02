package linkcheck

import "strings"

// Scope is the part of a project a slice depends on. Unless deltas.config names paths for the slice, that is
// every code file: a change anywhere can break a claim. Markdown is documentation and never counts, except
// the slice's own spec. What delta itself installs and keeps (.claude/, .githooks/, deltas/) is not code
// either: running delta update must not make every slice stale.
type Scope struct {
	paths []string // empty means every path
	spec  string
}

// NewScope is the scope of a spec file; paths, when given, narrow it to those files and folders.
func NewScope(spec string, paths []string) Scope { return Scope{paths: paths, spec: spec} }

// Covers says whether a change to file (relative to the project, with forward slashes) can affect the slice.
func (s Scope) Covers(file string) bool {
	switch {
	case file == s.spec:
		return true
	case strings.HasSuffix(file, ".md"), isDelta(file):
		return false
	}
	return inScope(file, s.paths)
}

// isDelta says whether a file belongs to delta's own installation in the project root.
func isDelta(file string) bool {
	for _, dir := range []string{".claude/", ".githooks/", "deltas/"} {
		if strings.HasPrefix(file, dir) {
			return true
		}
	}
	return false
}
