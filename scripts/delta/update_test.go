package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/manifest"
)

// init records what it installed, and update then keeps the harness current without touching what the project
// changed or what is the project's own.
func TestUpdateKeepsTheHarnessCurrentAndTheProjectsChanges(t *testing.T) {
	p := gitProject(t)
	p.delta("", "init")
	if !strings.Contains(p.read(manifest.Name), ".claude/skills/forge/SKILL.md") {
		t.Fatalf("init did not record the harness:\n%s", p.read(manifest.Name))
	}
	if code, out, _ := p.delta("", "update", "--check"); code != 0 || out != "" {
		t.Fatalf("a fresh install is out of date: %d %q", code, out)
	}

	// The project edits a rule and loses a skill; a skill delta used to ship and a config of its own remain.
	p.write(".claude/rules/fidelity.md", "# our own rule\n")
	if err := os.Remove(filepath.Join(p.root, ".claude", "skills", "seal", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	p.write("deltas/deltas.config", "command_timeout=3m\n")
	old := ".claude/skills/archivist/SKILL.md"
	p.write(old, "# archivist\n")
	p.write(manifest.Name, p.read(manifest.Name)+manifest.Hash([]byte("# archivist\n"))+"  "+old+"\n")

	code, out, _ := p.delta("", "update", "--check")
	if code != 1 || !strings.Contains(out, "install .claude/skills/seal/SKILL.md") || !strings.Contains(out, "remove "+old) || strings.Contains(out, "fidelity") {
		t.Fatalf("update --check = %d %q", code, out)
	}
	if _, err := os.Stat(filepath.Join(p.root, ".claude", "skills", "seal", "SKILL.md")); err == nil {
		t.Fatal("--check wrote a file")
	}

	code, out, errOut := p.delta("", "update")
	if code != 0 || errOut != "" || !strings.Contains(out, "installed 1 new file") || !strings.Contains(out, "removed 1 file") {
		t.Fatalf("update = %d %q %q", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(p.root, ".claude", "skills", "seal", "SKILL.md")); err != nil {
		t.Errorf("the missing skill was not installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p.root, ".claude", "skills", "archivist")); err == nil {
		t.Errorf("the skill the harness no longer ships was not removed with its folder")
	}
	if p.read(".claude/rules/fidelity.md") != "# our own rule\n" || p.read("deltas/deltas.config") != "command_timeout=3m\n" {
		t.Error("update replaced a file the project changed")
	}
	if !strings.Contains(out, "left as they are") || !strings.Contains(out, ".claude/rules/fidelity.md") {
		t.Errorf("update did not say what it kept: %q", out)
	}
	if strings.Contains(p.read(manifest.Name), "archivist") {
		t.Error("a removed file is still in the manifest")
	}
	if code, out, _ := p.delta("", "update", "--check"); code != 0 || out != "" {
		t.Errorf("after update, --check = %d %q", code, out)
	}
}

// status tells a project when its harness is behind this program's.
func TestStatusWarnsWhenTheHarnessIsOutOfDate(t *testing.T) {
	p := gitProject(t)
	p.delta("", "init")
	old := ".claude/rules/old.md"
	p.write(old, "# old\n")
	p.write(manifest.Name, p.read(manifest.Name)+manifest.Hash([]byte("# old\n"))+"  "+old+"\n")
	if _, _, errOut := p.delta("", "status"); !strings.Contains(errOut, "out of date") || !strings.Contains(errOut, "delta update") {
		t.Fatalf("status warnings lack the out-of-date harness: %q", errOut)
	}
}
