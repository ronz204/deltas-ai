package cmd

import (
	"strings"
	"testing"
)

func TestInitReportsEachStep(t *testing.T) {
	m := newMemory()
	out, _, err := m.run("", "init")
	if err != nil || out != "installed .githooks/commit-msg\nset core.hooksPath to .githooks\ninstalled 1 file from the harness\ninstalled .claude/settings.json\n" {
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

func healthyProject() *memory {
	m := newMemory()
	m.project[".githooks/commit-msg"] = []byte("#!/bin/sh\n")
	m.hooksDir = ".githooks"
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
		"warning: git does not check commit messages yet: run delta init",
		"warning: delta is not on the PATH, so the git hook cannot run it",
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
