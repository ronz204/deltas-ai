package budget

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/findings"
)

// Diagnostic codes. Hooks and CI depend on these, so they do not change once released.
const (
	CodeItem  = "BUDGET-ITEM"
	CodeTotal = "BUDGET-TOTAL"
)

// Patterns are the files that can be loaded on every turn; Measure keeps the ones that are.
var Patterns = []string{"CLAUDE.md", ".claude/CLAUDE.md", ".claude/rules/*.md", ".claude/skills/*/SKILL.md", ".claude/agents/*.md"}

// File is a knowledge-base file: its path relative to the project root, with forward slashes, and its text.
type File struct{ Path, Text string }

// Item is one thing loaded on every turn.
type Item struct {
	Path   string
	Kind   string // router, rule, skill or agent
	Unit   string // lines or chars
	Size   int
	Limit  int // 0 when there is no ceiling
	Tokens int
}

type Report struct {
	Items []Item // the heaviest first
	Total int    // tokens
	Limit int    // 0 when there is no ceiling
}

// tokens estimates at 3.3 characters per token, the ratio the ceilings were calibrated with against
// Claude Code's own /context report. It is integer arithmetic so it matches the budget script exactly.
func tokens(chars int) int { return chars * 10 / 33 }

// Measure weighs what the files load on every turn: the router, the rules without paths:, and the
// descriptions of skills and agents. The rest of a skill or a scoped rule costs nothing until used.
func Measure(files []File, c Ceilings) Report {
	var items []Item
	for _, f := range files {
		if item, ok := weigh(f, c); ok {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Tokens != items[j].Tokens {
			return items[i].Tokens > items[j].Tokens
		}
		return items[i].Path < items[j].Path
	})
	total := 0
	for _, it := range items {
		total += it.Tokens
	}
	return Report{Items: items, Total: total, Limit: c.Total}
}

func weigh(f File, c Ceilings) (Item, bool) {
	lines := strings.Split(strings.ReplaceAll(f.Text, "\r", ""), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	m := readMatter(lines)
	body := len(lines) - m.end
	whole := tokens(len(f.Text))

	rule, _ := path.Match(".claude/rules/*.md", f.Path)
	skill, _ := path.Match(".claude/skills/*/SKILL.md", f.Path)
	agent, _ := path.Match(".claude/agents/*.md", f.Path)
	switch {
	case f.Path == "CLAUDE.md" || f.Path == ".claude/CLAUDE.md":
		return Item{f.Path, "router", "lines", body, c.Router, whole}, true
	case rule && !m.paths:
		return Item{f.Path, "rule", "lines", body, c.Rule, whole}, true
	case skill:
		return Item{f.Path, "skill", "chars", m.desc, c.Skill, tokens(m.desc)}, true
	case agent:
		return Item{f.Path, "agent", "chars", m.desc, c.Agent, tokens(m.desc)}, true
	}
	return Item{}, false
}

// Findings lists every ceiling the report is over: each item's own, then the total.
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
