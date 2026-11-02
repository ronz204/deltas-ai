package terminal

import (
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/package/domain/budget"
)

func TestInitReportsEachStep(t *testing.T) {
	m := newMemory()
	out, _, err := m.run("", "init")
	if err != nil || out != "installed 1 file from the harness\ninstalled .claude/settings.json\n" {
		t.Fatalf("out = %q, err = %v", out, err)
	}

	out, _, err = m.run("", "init")
	if err != nil || !strings.Contains(out, "already installed") {
		t.Fatalf("a second run: out = %q, err = %v", out, err)
	}
}

func TestInitTakesNoArguments(t *testing.T) {
	if _, _, err := newMemory().run("", "init", "x"); err == nil {
		t.Fatal("init accepted an argument")
	}
}

func healthyProject() *memory {
	m := newMemory()
	m.onPath, m.self = "/usr/bin/delta", true
	return m
}

func TestStatusWarnsOnStderrAboutTheSetup(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = head + claim("INV-01", "none")
	out, errOut, err := m.run("", "status")
	if err != nil || !strings.Contains(out, "billing") {
		t.Fatalf("out = %q, err = %v; a warning must not fail the command", out, err)
	}
	for _, want := range []string{
		"warning: delta is not on the PATH, so the hooks of Claude Code cannot run it",
	} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr lacks %q:\n%s", want, errOut)
		}
	}
	if strings.Contains(out, "warning") {
		t.Errorf("a warning leaked into the table:\n%s", out)
	}
}

func TestStatusIsSilentWhenDeltaIsSetUp(t *testing.T) {
	m := healthyProject()
	m.specs["billing"] = head + claim("INV-01", "none")
	if _, errOut, err := m.run("", "status"); err != nil || errOut != "" {
		t.Fatalf("stderr = %q, err = %v; want silence", errOut, err)
	}
}

func TestStatusSaysHowManyDocsAreWrittenAndWhichIsNext(t *testing.T) {
	m := newMemory()
	m.kb = []budget.File{
		{Path: "deltas/docs/overview.md", Text: "# Overview\n\nWritten.\n"},
		{Path: "deltas/docs/structure.md", Text: "# Structure\n\n<!-- todo -->\n"},
	}
	out, _, err := m.run("", "status")
	if err != nil || !strings.Contains(out, "docs: 1 of 2 written; next: /scriber structure\n") {
		t.Fatalf("status = %q, %v", out, err)
	}
	m.kb = []budget.File{{Path: "deltas/docs/overview.md", Text: "Written."}}
	if out, _, _ := m.run("", "status"); !strings.Contains(out, "docs: all 1 written\n") {
		t.Fatalf("status = %q", out)
	}
	m.kb = nil
	if out, _, _ := m.run("", "status"); strings.Contains(out, "docs:") {
		t.Fatalf("status = %q; a project with no docs says nothing about them", out)
	}
}
