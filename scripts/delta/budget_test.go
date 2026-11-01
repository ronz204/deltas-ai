package main

import (
	"strings"
	"testing"
)

func budgetProject(t *testing.T) *project {
	t.Helper()
	p := newProject(t)
	p.write("CLAUDE.md", "# Acme\n"+strings.Repeat("a line\n", 9))
	p.write(".claude/rules/always.md", "# Rule\n"+strings.Repeat("a line\n", 10))
	p.write(".claude/rules/scoped.md", "---\npaths:\n  - \"src/**\"\n---\n"+strings.Repeat("a line\n", 200))
	p.write(".claude/skills/quill/SKILL.md", "---\nname: quill\ndescription: >-\n  Drafts the contract\n  of one slice.\n---\n"+strings.Repeat("a line\n", 300))
	p.write(".claude/docs/overview.md", strings.Repeat("a line\n", 500))
	return p
}

func TestBudgetWeighsOnlyWhatLoadsOnEveryTurn(t *testing.T) {
	p := budgetProject(t)
	code, out, errOut := p.delta("", "budget")
	if code != 0 || errOut != "" {
		t.Fatalf("budget = %d %q %q", code, out, errOut)
	}
	for _, want := range []string{"CLAUDE.md", ".claude/rules/always.md", ".claude/skills/quill/SKILL.md", "total: about", "of 3500"} {
		if !strings.Contains(out, want) {
			t.Errorf("the table lacks %q:\n%s", want, out)
		}
	}
	for _, not := range []string{"scoped.md", "overview.md"} {
		if strings.Contains(out, not) {
			t.Errorf("%s does not load on every turn, so it is not weighed:\n%s", not, out)
		}
	}
}

func TestBudgetFailsWhenTheConfCeilingIsPassed(t *testing.T) {
	p := budgetProject(t)
	p.write(".claude/budget.conf", "always_tokens=20\nrule_always_lines=5\n")
	code, out, errOut := p.delta("", "budget")
	if code != 1 {
		t.Fatalf("exit %d, want 1\n%s", code, out)
	}
	for _, want := range []string{".claude/rules/always.md: BUDGET-ITEM: 11 lines; the ceiling for a rule is 5", "BUDGET-TOTAL"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr lacks %q:\n%s", want, errOut)
		}
	}
	if !strings.Contains(out, "of 20") {
		t.Errorf("the table should show the ceiling from the conf:\n%s", out)
	}
}
