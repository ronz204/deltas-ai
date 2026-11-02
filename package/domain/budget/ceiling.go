package budget

import (
	"slices"
	"strconv"
	"strings"
)

type Ceilings struct {
	Total  int
	Router int
	Rule   int
	Skill  int
	Agent  int
}

func Defaults() Ceilings {
	return Ceilings{Total: 3500, Router: 80, Rule: 50, Skill: 1024, Agent: 1024}
}

var unread = []string{"rule_scoped_lines", "skill_body_lines", "agent_body_lines", "doc_lines", "reference_lines", "reference_chars"}

func (c *Ceilings) fields() map[string]*int {
	return map[string]*int{
		"always_tokens":     &c.Total,
		"claude_lines":      &c.Router,
		"rule_always_lines": &c.Rule,
		"skill_desc_chars":  &c.Skill,
		"agent_desc_chars":  &c.Agent,
	}
}

func IsKey(key string) bool {
	_, read := new(Ceilings).fields()[key]
	return read || slices.Contains(unread, key)
}

func ParseCeilings(conf string) Ceilings {
	c := Defaults()
	fields := c.fields()
	for _, line := range strings.Split(conf, "\n") {
		key, value, ok := strings.Cut(line, "=")
		field, known := fields[strings.TrimSpace(key)]
		if n, err := strconv.Atoi(strings.TrimSpace(value)); ok && known && err == nil && n >= 0 {
			*field = n
		}
	}
	return c
}
