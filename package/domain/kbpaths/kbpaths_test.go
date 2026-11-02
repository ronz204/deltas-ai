package kbpaths

import "testing"

func TestCovers(t *testing.T) {
	yes := []string{
		"CLAUDE.md", ".claude/CLAUDE.md",
		".claude/docs/overview.md", ".claude/rules/commits.md", ".claude/agents/reviewer.md", ".claude/skills/drafter/SKILL.md",
		"deltas/specs/billing.spec.md", "deltas/deltas.config", "deltas/deltas.manifest", "billing/deltas/specs/orders.spec.md", "services/auth/deltas/specs/login.spec.md",
	}
	no := []string{
		"", "README.md", "docs/CLAUDE.md", "main.go",
		".claude/settings.json", ".claude/settings.local.json", ".claude/budget.conf", ".claude/docs", ".claudeX/docs/a.md",
		"deltas", "mydeltas/a.md", "deltas/sub/x.md", "deltas/docs/overview.md", "specs/x.spec.md", "src/specs/x.md", "internal/harness/rules/core/x.md",
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
