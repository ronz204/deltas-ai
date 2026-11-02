package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func hookEvent(t *testing.T, name, tool, path, command string) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"hook_event_name": name,
		"tool_name":       tool,
		"tool_input":      map[string]string{"file_path": path, "command": command},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestHookEndToEnd(t *testing.T) {
	p := newProject(t)
	p.write("deltas/specs/billing.spec.md", spec("billing", claimOf("INV-01", "none")))

	code, out, errOut := p.delta(hookEvent(t, "SessionStart", "", "", ""), "hook")
	if code != 0 || errOut != "" || !strings.Contains(out, "- billing: approved, 1 claims (0 test, 0 manual, 1 none)") {
		t.Fatalf("SessionStart = %d %q %q", code, out, errOut)
	}

	code, _, errOut = p.delta(hookEvent(t, "PreToolUse", "Write", filepath.Join(p.root, "deltas", "specs", "billing.spec.md"), ""), "hook")
	if code != 2 || !strings.Contains(errOut, "deltas/specs/billing.spec.md") || !strings.Contains(errOut, "delta spec write") {
		t.Fatalf("a write into deltas/ = %d %q, want a block that says how to write it", code, errOut)
	}
	if code, _, errOut = p.delta(hookEvent(t, "PreToolUse", "Write", filepath.Join(p.root, "main.go"), ""), "hook"); code != 0 || errOut != "" {
		t.Fatalf("a write to main.go = %d %q, want it allowed", code, errOut)
	}

}

func TestHookInputThatIsNotJSONNeverBlocks(t *testing.T) {
	if code, _, errOut := newProject(t).delta("not json", "hook"); code != 1 || !strings.HasPrefix(errOut, "delta hook: ") {
		t.Fatalf("code = %d, stderr = %q; Claude Code blocks only on 2", code, errOut)
	}
}

func isolatedGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	empty := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", empty)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

func git(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	full := append([]string{"-C", dir, "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	return string(out), err
}

func gitProject(t *testing.T) *project {
	t.Helper()
	isolatedGit(t)
	isolateCache(t)
	root := t.TempDir()
	if out, err := git(t, root, "init", "-q"); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return &project{t: t, root: root}
}
