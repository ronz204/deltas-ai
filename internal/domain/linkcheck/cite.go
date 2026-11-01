// Package linkcheck cross-checks a project's specs against its test files.
package linkcheck

import (
	"path"
	"regexp"
	"strings"

	"github.com/ronz204/deltas-ai/internal/domain/assertion"
)

// File is a source file of the project: its path relative to the root, with forward slashes,
// and its text.
type File struct{ Path, Text string }

// scriptExts are the extensions of the languages whose tests and files the tool recognises.
var scriptExts = map[string]bool{".go": true, ".py": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true, ".mts": true, ".cts": true}

// IsTestFile recognises a test file by its name: *_test.go, test_*.py, *_test.py and
// *.test.* or *.spec.* in JavaScript or TypeScript (so a *.spec.md contract is not one).
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

// idPattern matches a claim ID however a test name can spell it: INV-01, INV_01 or INV01, in any
// case, and not as the start of a longer number.
func idPattern(id assertion.ID) *regexp.Regexp {
	prefix, num, _ := strings.Cut(id.String(), "-")
	return regexp.MustCompile(`(?i)` + prefix + `[-_]?` + num + `(?:[^0-9]|$)`)
}

// Cited reports whether a test file that command runs cites the claim. IDs are only unique inside
// one spec, so a citation counts only for the files under the paths the command names.
//
// shortcut: paths are found by shape (./x, ../x, or a script file with a slash) and quoting is
// ignored; a command that names none (npm test) is taken to run every test file. Name the
// package or file in the tag to narrow it.
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

// pathsIn returns the arguments of command that look like paths, cleaned.
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

// inScope reports whether file is one of the paths or under one of them. With no paths the
// command is taken to run everything.
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
