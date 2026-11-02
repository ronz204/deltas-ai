// Package linking cross-checks a project's specs against its test files.
package linking

import (
	"path"
	"regexp"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/claimset"
)

type File struct{ Path, Text string }

var scriptExts = map[string]bool{".go": true, ".py": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true, ".mts": true, ".cts": true}

// DefaultTestFiles is delta's own rule for what a test file is, as name patterns: the same as the match
// that testing.conf installs.
var DefaultTestFiles = []string{"*_test.go", "test_*.py", "*_test.py", "*.test.*", "*.spec.*"}

var defaultMatcher = TestMatcher(nil)

// IsTestFile applies DefaultTestFiles.
func IsTestFile(p string) bool { return defaultMatcher(p) }

// TestMatcher says which files are tests: those whose name matches one of the patterns (as in path.Match,
// on the name alone), or DefaultTestFiles when there are none. A Markdown file never is one, so a
// *.spec.md contract is not.
func TestMatcher(patterns []string) func(string) bool {
	if len(patterns) == 0 {
		patterns = DefaultTestFiles
	}
	return func(p string) bool {
		base := path.Base(p)
		if path.Ext(base) == ".md" {
			return false
		}
		for _, pattern := range patterns {
			if ok, _ := path.Match(pattern, base); ok {
				return true
			}
		}
		return false
	}
}

// idPattern matches INV-01, INV_01 and INV01 in any case, but not INV-011.
func idPattern(id claimset.ID) *regexp.Regexp {
	prefix, num, _ := strings.Cut(id.String(), "-")
	return regexp.MustCompile(`(?i)` + prefix + `[-_]?` + num + `(?:[^0-9]|$)`)
}

// Cited counts a citation only under the paths the command names: IDs are unique per spec, not per project.
// shortcut: paths are found by shape and quotes are ignored; a command naming none runs every test file.
func Cited(id claimset.ID, command string, files []File) bool {
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
