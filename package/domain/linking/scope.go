package linking

import (
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
)

type Scope struct {
	spec  string
	paths []string
}

func NewScope(spec string, claims []claimset.Claim) Scope {
	return Scope{spec: spec, paths: testPaths(claims)}
}

func testPaths(claims []claimset.Claim) []string {
	var paths []string
	for _, c := range claims {
		if c.Tag.Method != claimset.Test {
			continue
		}
		named := pathsIn(c.Tag.Arg)
		if len(named) == 0 {
			return nil
		}
		paths = append(paths, named...)
	}
	return paths
}

func (s Scope) Covers(file string) bool {
	switch {
	case file == s.spec:
		return true
	case strings.HasSuffix(file, ".md"), isDelta(file):
		return false
	}
	return underAny(file, s.paths)
}

func isDelta(file string) bool {
	for _, dir := range []string{".claude/", "deltas/"} {
		if strings.HasPrefix(file, dir) {
			return true
		}
	}
	return false
}
