package linking

import "path"

var DefaultTestFiles = []string{"*_test.go", "test_*.py", "*_test.py", "*.test.*", "*.spec.*"}

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
