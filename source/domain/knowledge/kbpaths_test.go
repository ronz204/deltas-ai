package knowledge

import "testing"

func TestCovers(t *testing.T) {
	yes := []string{
		"CLAUDE.md", ".claude/CLAUDE.md",
		".claude/docs/overview.md", ".claude/rules/commits.md", ".claude/agents/reviewer.md", ".claude/skills/quill/SKILL.md",
		"deltas/billing.spec.md", "deltas/billing.design.md", "billing/deltas/orders.spec.md",
	}
	no := []string{
		"", "README.md", "docs/CLAUDE.md", "main.go",
		".claude/settings.json", ".claude/settings.local.json", ".claude/budget.conf", ".claude/docs", ".claudeX/docs/a.md",
		"deltas", "a/b/deltas/c.md", "mydeltas/a.md", "internal/harness/rules/core/x.md",
	}
	for _, p := range yes {
		if !Covers(p) {
			t.Errorf("%q should belong to the knowledge base", p)
		}
	}
	for _, p := range no {
		if Covers(p) {
			t.Errorf("%q should not belong to the knowledge base", p)
		}
	}
}
