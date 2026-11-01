package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/budget"
)

func kb(path string, lines int) budget.File {
	return budget.File{Path: path, Text: strings.Repeat("x\n", lines)}
}

func TestBudgetPrintsTheTableAndTheTotal(t *testing.T) {
	m := newMemory()
	m.kb = []budget.File{kb("CLAUDE.md", 10), kb(".claude/rules/a.md", 30)}
	out, errOut, err := m.run("", "budget")
	if err != nil || errOut != "" {
		t.Fatalf("err = %v, stderr = %q", err, errOut)
	}
	var rows []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		rows = append(rows, strings.Join(strings.Fields(line), " "))
	}
	want := []string{
		"KIND FILE SIZE LIMIT TOKENS",
		"rule .claude/rules/a.md 30 lines 50 18",
		"router CLAUDE.md 10 lines 80 6",
		"total: about 24 tokens of 3500",
	}
	if strings.Join(rows, "|") != strings.Join(want, "|") {
		t.Fatalf("rows:\n%s\nwant:\n%s", strings.Join(rows, "\n"), strings.Join(want, "\n"))
	}
}

func TestBudgetFailsAndSaysWhatIsOver(t *testing.T) {
	m := newMemory()
	m.kb = []budget.File{kb(".claude/rules/big.md", 51)}
	out, errOut, err := m.run("", "budget")
	if !errors.Is(err, ErrFound) || !strings.Contains(errOut, ".claude/rules/big.md: BUDGET-ITEM: 51 lines; the ceiling for a rule is 50") {
		t.Fatalf("err = %v, stderr = %q", err, errOut)
	}
	if !strings.Contains(out, "big.md") {
		t.Fatalf("the table still prints when a ceiling is passed:\n%s", out)
	}
}

func TestBudgetOfAProjectWithNothingLoadedPasses(t *testing.T) {
	out, errOut, err := newMemory().run("", "budget")
	if err != nil || errOut != "" || !strings.Contains(out, "total: about 0 tokens of 3500") {
		t.Fatalf("out = %q, stderr = %q, err = %v", out, errOut, err)
	}
}

func TestBudgetTakesNoArguments(t *testing.T) {
	if _, _, err := newMemory().run("", "budget", "x"); err == nil {
		t.Fatal("budget accepted an argument")
	}
}
