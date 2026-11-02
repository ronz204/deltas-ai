package linking

import (
	"path"
	"strings"
)

var DefaultTestFiles = []string{"*_test.go", "test_*.py", "*_test.py", "*.test.*", "*.spec.*"}

var DefaultIgnored = []string{
	"go.sum", "package-lock.json", "npm-shrinkwrap.json", "yarn.lock", "pnpm-lock.yaml", "bun.lockb",
	"Cargo.lock", "poetry.lock", "Pipfile.lock", "uv.lock", "composer.lock", "Gemfile.lock",
}

var defaultMatcher = TestMatcher(nil)

func IsTestFile(p string) bool { return defaultMatcher(p) }

func TestMatcher(patterns []string) func(string) bool {
	if len(patterns) == 0 {
		patterns = DefaultTestFiles
	}
	return func(p string) bool {
		name := path.Base(p)
		if path.Ext(name) == ".md" {
			return false
		}
		for _, pattern := range patterns {
			if ok, _ := path.Match(pattern, name); ok {
				return true
			}
		}
		return false
	}
}

func IgnoreMatcher(patterns []string) func(string) bool {
	if len(patterns) == 0 {
		patterns = DefaultIgnored
	}
	return func(p string) bool {
		for _, pattern := range patterns {
			if ignores(pattern, p) {
				return true
			}
		}
		return false
	}
}

func ValidIgnore(pattern string) bool {
	_, err := path.Match(strings.TrimSuffix(pattern, "/**"), "x")
	return err == nil
}

func ignores(pattern, file string) bool {
	if dir, folder := strings.CutSuffix(pattern, "/**"); folder {
		return strings.HasPrefix(file, dir+"/")
	}
	target := file
	if !strings.Contains(pattern, "/") {
		target = path.Base(file)
	}
	ok, _ := path.Match(pattern, target)
	return ok
}
