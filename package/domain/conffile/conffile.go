package conffile

import (
	"fmt"
	"path"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/budget"
	"github.com/ronz204/deltas-ai/package/domain/linking"
)

const (
	BudgetName  = "deltas/budget.conf"
	TestingName = "deltas/testing.conf"
)

type Conf struct {
	Budget   budget.Ceilings
	Files    Files
	Problems []string
}

type Files struct {
	Match  []string
	Skip   []string
	Ignore []string
}

type line struct {
	n    int
	text string
}

func Parse(budgetText, testingText string) Conf {
	c := Conf{Budget: budget.ParseCeilings(budgetText)}
	c.readBudget(budgetText)
	c.readTesting(testingText)
	return c
}

func (c *Conf) problem(name string, l line, format string, args ...any) {
	c.Problems = append(c.Problems, fmt.Sprintf("%s:%d: %s", name, l.n+1, fmt.Sprintf(format, args...)))
}

func (c *Conf) readBudget(text string) {
	for _, l := range lines(text) {
		key, _, ok := strings.Cut(l.text, "=")
		switch key = strings.TrimSpace(key); {
		case !ok:
			c.problem(BudgetName, l, "expected key=value")
		case !budget.IsKey(key):
			c.problem(BudgetName, l, "unknown key %q", key)
		}
	}
}

func (c *Conf) readTesting(text string) {
	for _, l := range lines(text) {
		key, value, ok := strings.Cut(l.text, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch {
		case !ok:
			c.problem(TestingName, l, "expected key=value")
		case key == "match":
			c.readMatch(l, value)
		case key == "skip":
			c.Files.Skip = words(value)
		case key == "ignore":
			c.readIgnore(l, value)
		default:
			c.problem(TestingName, l, "unknown key %q", key)
		}
	}
}

func (c *Conf) readMatch(l line, value string) {
	patterns := words(value)
	for _, p := range patterns {
		if _, err := path.Match(p, "x"); err != nil {
			c.problem(TestingName, l, "match: %q is not a valid pattern", p)
			return
		}
	}
	c.Files.Match = patterns
}

func (c *Conf) readIgnore(l line, value string) {
	patterns := words(value)
	for _, p := range patterns {
		if !linking.ValidIgnore(p) {
			c.problem(TestingName, l, "ignore: %q is not a valid pattern", p)
			return
		}
	}
	c.Files.Ignore = patterns
}

func lines(text string) []line {
	var out []line
	for i, raw := range strings.Split(text, "\n") {
		if t := strings.TrimSpace(raw); t != "" && !strings.HasPrefix(t, "#") {
			out = append(out, line{i, t})
		}
	}
	return out
}

func words(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
}
