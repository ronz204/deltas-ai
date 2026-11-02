package budget

import (
	"cmp"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/findings"
)

const (
	CodeItem  = "BUDGET-ITEM"
	CodeTotal = "BUDGET-TOTAL"
)

const (
	rulesGlob  = ".claude/rules/*.md"
	skillsGlob = ".claude/skills/*/SKILL.md"
	agentsGlob = ".claude/agents/*.md"
)

var Patterns = []string{"CLAUDE.md", ".claude/CLAUDE.md", rulesGlob, skillsGlob, agentsGlob}

type File struct{ Path, Text string }

type Item struct {
	Path   string
	Kind   string
	Unit   string
	Size   int
	Limit  int
	Tokens int
}

type Report struct {
	Items []Item
	Total int
	Limit int
}

func tokens(chars int) int { return chars * 10 / 33 }

func Measure(files []File, c Ceilings) Report {
	var items []Item
	total := 0
	for _, f := range files {
		if item, ok := weigh(f, c); ok {
			items = append(items, item)
			total += item.Tokens
		}
	}
	slices.SortFunc(items, func(a, b Item) int {
		return cmp.Or(cmp.Compare(b.Tokens, a.Tokens), strings.Compare(a.Path, b.Path))
	})
	return Report{Items: items, Total: total, Limit: c.Total}
}

func matches(pattern, file string) bool {
	ok, _ := path.Match(pattern, file)
	return ok
}

func weigh(f File, c Ceilings) (Item, bool) {
	lines := strings.Split(strings.ReplaceAll(f.Text, "\r", ""), "\n")
	if last := len(lines) - 1; lines[last] == "" {
		lines = lines[:last]
	}
	m := readMatter(lines)
	item := func(kind, unit string, size, limit, cost int) (Item, bool) {
		return Item{Path: f.Path, Kind: kind, Unit: unit, Size: size, Limit: limit, Tokens: cost}, true
	}
	switch {
	case f.Path == "CLAUDE.md" || f.Path == ".claude/CLAUDE.md":
		return item("router", "lines", len(lines)-m.end, c.Router, tokens(len(f.Text)))
	case matches(rulesGlob, f.Path) && !m.paths:
		return item("rule", "lines", len(lines)-m.end, c.Rule, tokens(len(f.Text)))
	case matches(skillsGlob, f.Path):
		return item("skill", "chars", m.desc, c.Skill, tokens(m.desc))
	case matches(agentsGlob, f.Path):
		return item("agent", "chars", m.desc, c.Agent, tokens(m.desc))
	}
	return Item{}, false
}

func (r Report) Findings() []findings.Diagnostic {
	var out []findings.Diagnostic
	for _, it := range r.Items {
		if it.Limit > 0 && it.Size > it.Limit {
			out = append(out, findings.Diagnostic{File: it.Path, Code: CodeItem,
				Message: fmt.Sprintf("%d %s; the ceiling for a %s is %d", it.Size, it.Unit, it.Kind, it.Limit)})
		}
	}
	if r.Limit > 0 && r.Total > r.Limit {
		out = append(out, findings.Diagnostic{File: "always-loaded context", Code: CodeTotal,
			Message: fmt.Sprintf("about %d tokens; the ceiling is %d", r.Total, r.Limit)})
	}
	return out
}
