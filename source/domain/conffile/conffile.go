// Package conffile reads budget.conf and testing.conf, the optional settings files in the deltas/ folder at the root of a project.
package conffile

import (
	"fmt"
	"path"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/budget"
)

const (
	BudgetName  = "deltas/budget.conf"
	TestingName = "deltas/testing.conf"
)

// Conf is what the config files say over delta's defaults. A line that cannot be used is left out and told in Problems.
type Conf struct {
	Budget   budget.Ceilings
	Files    Files
	Problems []string
}

// Files is what testing.conf says about test files: which are tests and where not to look for them.
// An empty list means delta's own.
type Files struct {
	Match []string // patterns on a file's name
	Skip  []string // folders never searched
}

// Parse reads the text of budget.conf and of testing.conf. Both are key=value lines; blank lines and lines starting
// with # are skipped. budget.conf has the budget ceilings (see budget.ParseCeilings). testing.conf has match, the
// patterns of the test files, and skip, the folders never searched.
func Parse(budgetText, testingText string) Conf {
	c := Conf{Budget: budget.ParseCeilings(budgetText)}
	c.budget(budgetText)
	c.testing(testingText)
	return c
}

func (c *Conf) problem(name string, i int, msg string) {
	c.Problems = append(c.Problems, fmt.Sprintf("%s:%d: %s", name, i+1, msg))
}

func (c *Conf) budget(text string) {
	for _, l := range lines(text) {
		key, _, ok := strings.Cut(l.text, "=")
		switch key = strings.TrimSpace(key); {
		case !ok:
			c.problem(BudgetName, l.n, "expected key=value")
		case !budget.IsKey(key):
			c.problem(BudgetName, l.n, fmt.Sprintf("unknown key %q", key))
		}
	}
}

func (c *Conf) testing(text string) {
	for _, l := range lines(text) {
		key, value, ok := strings.Cut(l.text, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch {
		case !ok:
			c.problem(TestingName, l.n, "expected key=value")
		case key == "match":
			c.match(l.n, value)
		case key == "skip":
			c.Files.Skip = words(value)
		default:
			c.problem(TestingName, l.n, fmt.Sprintf("unknown key %q", key))
		}
	}
}

func (c *Conf) match(i int, value string) {
	for _, p := range words(value) {
		if _, err := path.Match(p, "x"); err != nil {
			c.problem(TestingName, i, fmt.Sprintf("match: %q is not a valid pattern", p))
			return
		}
	}
	c.Files.Match = words(value)
}

type line struct {
	n    int // position in the file, counting from zero
	text string
}

// lines gives the lines of a config text that say something, trimmed, each with its position.
func lines(text string) []line {
	var out []line
	for i, raw := range strings.Split(text, "\n") {
		if t := strings.TrimSpace(raw); t != "" && !strings.HasPrefix(t, "#") {
			out = append(out, line{i, t})
		}
	}
	return out
}

// words splits a value on commas and spaces.
func words(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
}
