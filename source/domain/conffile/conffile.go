// Package conffile reads delta.conf, the optional settings file at the root of a project.
package conffile

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/ronz204/deltas-ai/source/domain/budget"
)

const Name = "delta.conf"

// Conf is what delta.conf says over delta's defaults. A line that cannot be used is left out and told in Problems.
type Conf struct {
	Budget    budget.Ceilings
	TestFiles []string            // patterns on a file's name; empty means delta's own
	Timeout   time.Duration       // for each claim's command; zero means the default
	Tools     []string            // programs that must be on the PATH
	Scopes    map[string][]string // slice -> the paths it depends on; a slice not listed depends on every code file
	Problems  []string
}

// Parse reads key=value lines; blank lines and lines starting with # are skipped. The keys are the budget
// ceilings (see budget.ParseCeilings), test_files, command_timeout, required_tools and scope.<slice>.
func Parse(text string) Conf {
	c := Conf{Budget: budget.ParseCeilings(text), Scopes: map[string][]string{}}
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok {
			c.problem(i, "expected key=value")
			continue
		}
		switch {
		case budget.IsKey(key):
		case key == "test_files":
			c.testFiles(i, value)
		case key == "command_timeout":
			c.timeout(i, value)
		case key == "required_tools":
			c.Tools = words(value)
		case strings.HasPrefix(key, "scope."):
			c.scope(i, strings.TrimPrefix(key, "scope."), value)
		default:
			c.problem(i, fmt.Sprintf("unknown key %q", key))
		}
	}
	return c
}

func (c *Conf) problem(i int, msg string) {
	c.Problems = append(c.Problems, fmt.Sprintf("%s:%d: %s", Name, i+1, msg))
}

func (c *Conf) testFiles(i int, value string) {
	for _, p := range words(value) {
		if _, err := path.Match(p, "x"); err != nil {
			c.problem(i, fmt.Sprintf("test_files: %q is not a valid pattern", p))
			return
		}
	}
	c.TestFiles = words(value)
}

func (c *Conf) timeout(i int, value string) {
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		c.problem(i, fmt.Sprintf("command_timeout: %q is not a positive duration such as 10m", value))
		return
	}
	c.Timeout = d
}

func (c *Conf) scope(i int, slice, value string) {
	var paths []string
	for _, p := range words(value) {
		p = strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(strings.ReplaceAll(p, `\`, "/"), "./"), "..."), "/")
		if p != "" && p != "." {
			paths = append(paths, path.Clean(p))
		}
	}
	if slice == "" || len(paths) == 0 {
		c.problem(i, "scope.<slice> needs a slice name and at least one path inside the project")
		return
	}
	c.Scopes[slice] = paths
}

// words splits a value on commas and spaces.
func words(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
}
