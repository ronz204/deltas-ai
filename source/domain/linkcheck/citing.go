// Package linkcheck cross-checks a project's specs against its test files.
package linkcheck

import (
	"path"
	"regexp"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/assertion"
)

type File struct{ Path, Text string }

var scriptExts = map[string]bool{".go": true, ".py": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true, ".mts": true, ".cts": true}

// IsTestFile ignores *.spec.md contracts: only script extensions count.
func IsTestFile(p string) bool {
	base := path.Base(p)
	ext := path.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	switch {
	case ext == ".go":
		return strings.HasSuffix(stem, "_test")
	case ext == ".py":
		return strings.HasPrefix(stem, "test_") || strings.HasSuffix(stem, "_test")
	case scriptExts[ext]:
		return strings.HasSuffix(stem, ".test") || strings.HasSuffix(stem, ".spec")
	}
	return false
}

// idPattern matches INV-01, INV_01 and INV01 in any case, but not INV-011.
func idPattern(id assertion.ID) *regexp.Regexp {
	prefix, num, _ := strings.Cut(id.String(), "-")
	return regexp.MustCompile(`(?i)` + prefix + `[-_]?` + num + `(?:[^0-9]|$)`)
}

// Cited counts a citation only under the paths the command names: IDs are unique per spec, not per project.
// shortcut: paths are found by shape and quotes are ignored; a command naming none runs every test file.
func Cited(id assertion.ID, command string, files []File) bool {
	scope := pathsIn(command)
	re := idPattern(id)
	for _, f := range files {
		if inScope(f.Path, scope) && re.MatchString(f.Text) {
			return true
		}
	}
	return false
}

func pathsIn(command string) []string {
	var paths []string
	for _, arg := range strings.Fields(command) {
		if p, ok := asPath(arg); ok {
			paths = append(paths, p)
		}
	}
	return paths
}

func asPath(arg string) (string, bool) {
	if strings.HasPrefix(arg, "-") {
		return "", false
	}
	p := strings.ReplaceAll(arg, `\`, "/")
	startsHere := p == "." || strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../")
	namesScript := strings.Contains(p, "/") && scriptExts[path.Ext(p)]
	if !startsHere && !namesScript {
		return "", false
	}
	p = strings.TrimSuffix(p, "...") // ./pkg/... means the package and everything under it
	p = strings.TrimSuffix(p, "/")
	return path.Clean(p), true
}

func inScope(file string, paths []string) bool {
	if len(paths) == 0 {
		return true
	}
	for _, p := range paths {
		if p == "." || file == p || strings.HasPrefix(file, p+"/") {
			return true
		}
	}
	return false
}
