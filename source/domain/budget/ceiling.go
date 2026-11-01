// Package budget weighs the context a project loads on every turn against its ceilings.
package budget

import (
	"strconv"
	"strings"
)

// Ceilings are the most the always-loaded context may cost. Zero means no ceiling.
type Ceilings struct {
	Total  int // tokens for everything loaded on every turn
	Router int // lines of CLAUDE.md
	Rule   int // lines of a rule that has no paths:
	Skill  int // characters of a skill description
	Agent  int // characters of an agent description
}

func Defaults() Ceilings {
	return Ceilings{Total: 3500, Router: 80, Rule: 50, Skill: 1024, Agent: 1024}
}

// ParseCeilings reads the always-loaded keys of a budget.conf over the defaults: always_tokens,
// claude_lines, rule_always_lines, skill_desc_chars and agent_desc_chars. Other keys, comments
// and values that are not numbers are ignored.
func ParseCeilings(conf string) Ceilings {
	c := Defaults()
	targets := map[string]*int{
		"always_tokens":     &c.Total,
		"claude_lines":      &c.Router,
		"rule_always_lines": &c.Rule,
		"skill_desc_chars":  &c.Skill,
		"agent_desc_chars":  &c.Agent,
	}
	for _, line := range strings.Split(conf, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		target, known := targets[strings.TrimSpace(key)]
		if n, err := strconv.Atoi(strings.TrimSpace(value)); ok && known && err == nil && n >= 0 {
			*target = n
		}
	}
	return c
}
