package linking

import "strings"

type Scope struct {
	spec string
}

func NewScope(spec string) Scope { return Scope{spec: spec} }

func (s Scope) Covers(file string) bool {
	switch {
	case file == s.spec:
		return true
	case strings.HasSuffix(file, ".md"), isDelta(file):
		return false
	}
	return true
}

func isDelta(file string) bool {
	for _, dir := range []string{".claude/", "deltas/"} {
		if strings.HasPrefix(file, dir) {
			return true
		}
	}
	return false
}
