package cmd

import (
	"strings"
	"testing"
)

func TestInitReportsEachStep(t *testing.T) {
	m := newMemory()
	out, _, err := m.run("", "init")
	if err != nil || out != "installed .githooks/commit-msg\nset core.hooksPath to .githooks\n" {
		t.Fatalf("out = %q, err = %v", out, err)
	}
	if _, ok := m.project[".githooks/commit-msg"]; !ok || m.hooksDir != ".githooks" {
		t.Fatalf("hook stored: %v, hooksPath = %q", ok, m.hooksDir)
	}

	out, _, err = m.run("", "init")
	if err != nil || !strings.Contains(out, "already installed") || !strings.Contains(out, "already points") {
		t.Fatalf("a second run: out = %q, err = %v", out, err)
	}
}

func TestInitLeavesAnotherHooksPathAlone(t *testing.T) {
	m := newMemory()
	m.hooksDir = ".husky"
	out, _, err := m.run("", "init")
	if err != nil || m.hooksDir != ".husky" || !strings.Contains(out, "core.hooksPath is .husky: left as it is") {
		t.Fatalf("out = %q, hooksPath = %q, err = %v", out, m.hooksDir, err)
	}
}

func TestInitTakesNoArguments(t *testing.T) {
	if _, _, err := newMemory().run("", "init", "x"); err == nil {
		t.Fatal("init accepted an argument")
	}
}
