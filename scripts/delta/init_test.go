package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/conffile"
)

// Everything init installs in a new repository, and that delta works on it afterwards.
func TestInitInstallsTheHarnessAndDeltaWorksOnIt(t *testing.T) {
	p := gitProject(t)
	code, out, errOut := p.delta("", "init")
	if code != 0 || errOut != "" || !strings.Contains(out, "files from the harness") || !strings.Contains(out, "installed .claude/settings.json") {
		t.Fatalf("init = %d %q %q", code, out, errOut)
	}

	for _, rel := range []string{
		".githooks/commit-msg", "deltas/specs/.gitkeep", "deltas/deltas.config", "CLAUDE.md",
		".claude/settings.json", ".claude/rules/fidelity.md",
		".claude/skills/quill/SKILL.md", ".claude/skills/forge/SKILL.md", ".claude/skills/seal/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(p.root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s was not installed: %v", rel, err)
		}
	}
	if !strings.HasPrefix(p.read("CLAUDE.md"), "# "+filepath.Base(p.root)+"\n") {
		t.Errorf("CLAUDE.md does not name the project:\n%s", p.read("CLAUDE.md"))
	}
	if c := conffile.Parse(p.read("deltas/deltas.config")); len(c.Problems) != 0 {
		t.Errorf("the installed deltas.config has problems: %v", c.Problems)
	}
	var settings struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(p.read(".claude/settings.json")), &settings); err != nil || len(settings.Hooks) != 4 {
		t.Errorf("settings.json: %v, %d hook events", err, len(settings.Hooks))
	}
	for _, gone := range []string{"archivist", "surveyor"} {
		if _, err := os.Stat(filepath.Join(p.root, ".claude", "skills", gone)); err == nil {
			t.Errorf("the %s skill is still installed", gone)
		}
	}

	if code, out, errOut := p.delta("", "status"); code != 0 || strings.Contains(errOut, "warning") && !strings.Contains(errOut, "PATH") {
		t.Errorf("status = %d %q %q", code, out, errOut)
	}
	if code, out, errOut := p.delta("", "budget"); code != 0 {
		t.Errorf("budget on what init installed = %d %q %q", code, out, errOut)
	}
}

func TestInitTwiceChangesNothingTheSecondTime(t *testing.T) {
	p := gitProject(t)
	p.delta("", "init")
	before := p.read(".claude/settings.json")
	code, out, _ := p.delta("", "init")
	if code != 0 || strings.Contains(out, "installed .claude") || !strings.Contains(out, "already installed") {
		t.Fatalf("a second init = %d %q", code, out)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.HasPrefix(line, "installed ") || strings.HasPrefix(line, "set ") || strings.HasPrefix(line, "added ") {
			t.Errorf("a second init did work: %q", line)
		}
	}
	if p.read(".claude/settings.json") != before {
		t.Error("settings.json changed on the second run")
	}
}

func TestInitKeepsWhatTheProjectAlreadyHas(t *testing.T) {
	p := gitProject(t)
	p.write("CLAUDE.md", "# my own router\n")
	p.write(".claude/rules/fidelity.md", "# my fidelity rule\n")
	p.write(".claude/settings.json", `{"model": "opus", "permissions": {"deny": ["Bash(rm -rf *)"]}}`)

	code, out, _ := p.delta("", "init")
	if code != 0 || !strings.Contains(out, "CLAUDE.md") || !strings.Contains(out, ".claude/rules/fidelity.md") || !strings.Contains(out, "left as they are") {
		t.Fatalf("init = %d %q", code, out)
	}
	if p.read("CLAUDE.md") != "# my own router\n" || p.read(".claude/rules/fidelity.md") != "# my fidelity rule\n" {
		t.Error("init replaced a file the project had")
	}
	var settings map[string]any
	if err := json.Unmarshal([]byte(p.read(".claude/settings.json")), &settings); err != nil {
		t.Fatalf("settings.json: %v\n%s", err, p.read(".claude/settings.json"))
	}
	deny := settings["permissions"].(map[string]any)["deny"].([]any)
	if settings["model"] != "opus" || deny[0] != "Bash(rm -rf *)" || len(deny) < 2 || settings["hooks"] == nil {
		t.Errorf("settings.json = %s; want the project's keys kept and delta's added", p.read(".claude/settings.json"))
	}
}

// A clone that never ran init is told so when a Claude Code session starts.
func TestSessionStartTellsAFreshCloneToRunInit(t *testing.T) {
	p := gitProject(t)
	code, out, errOut := p.delta(hookEvent(t, "SessionStart", "", "", ""), "hook")
	if code != 0 || errOut != "" || !strings.Contains(out, "core.hooksPath") && !strings.Contains(out, "git does not check commit messages") {
		t.Fatalf("SessionStart = %d %q %q", code, out, errOut)
	}
	if !strings.Contains(out, "run delta init") {
		t.Errorf("the warning does not say what to do:\n%s", out)
	}
}

// base=off in deltas.config keeps the base preset out, and update retires it once installed.
func TestInitLeavesOutTheBasePresetWhenTheConfigTurnsItOff(t *testing.T) {
	p := gitProject(t)
	p.write("deltas/deltas.config", "base=off\n")
	if code, out, errOut := p.delta("", "init"); code != 0 {
		t.Fatalf("init = %d %q %q", code, out, errOut)
	}
	for rel, want := range map[string]bool{".claude/rules/fidelity.md": true, ".claude/rules/honesty.md": true, ".claude/rules/base-scope.md": false, ".claude/rules/base-reliability.md": false} {
		if _, err := os.Stat(filepath.Join(p.root, filepath.FromSlash(rel))); (err == nil) != want {
			t.Errorf("%s installed = %v, want %v", rel, err == nil, want)
		}
	}
}
