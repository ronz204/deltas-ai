package cmd

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func event(name, tool, path, command string) string {
	return `{"hook_event_name":"` + name + `","tool_name":"` + tool + `","tool_input":{"file_path":` + quote(path) + `,"command":` + quote(command) + `}}`
}

func quote(s string) string {
	s = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`).Replace(s)
	return `"` + s + `"`
}

func exitCode(t *testing.T, err error) int {
	t.Helper()
	var exit *ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("err = %v, want an ExitError", err)
	}
	return exit.Code
}

func TestHookBlocksWithExitTwoAndTheReason(t *testing.T) {
	m := newMemory()
	out, errOut, err := m.run(event("PreToolUse", "Write", filepath.Join(root, "deltas", "a.spec.md"), ""), "hook")
	if exitCode(t, err) != 2 || out != "" || !strings.Contains(errOut, "deltas/a.spec.md") {
		t.Fatalf("err = %v, stdout = %q, stderr = %q", err, out, errOut)
	}
	_, errOut, err = m.run(event("PreToolUse", "Bash", "", `git commit -m "Fixed it."`), "hook")
	if exitCode(t, err) != 2 || !strings.Contains(errOut, "COMMIT-") {
		t.Fatalf("a bad commit: err = %v, stderr = %q", err, errOut)
	}
}

func TestHookLetsOtherEventsThroughInSilence(t *testing.T) {
	m := newMemory()
	for _, input := range []string{
		event("PreToolUse", "Write", filepath.Join(root, "main.go"), ""),
		event("PreToolUse", "Bash", "", "go test ./..."),
		event("Stop", "", "", ""),
	} {
		if out, errOut, err := m.run(input, "hook"); err != nil || out != "" || errOut != "" {
			t.Errorf("%s = %q %q, %v; want it let through in silence", input, out, errOut, err)
		}
	}
}

func TestHookStartsASessionWithTheSlices(t *testing.T) {
	m := newMemory()
	m.specs["billing"] = head + claim("INV-01", "none")
	out, _, err := m.run(`{"hook_event_name":"SessionStart","source":"startup"}`, "hook")
	if err != nil || !strings.Contains(out, "- billing: approved, 1 claims") {
		t.Fatalf("out = %q, err = %v", out, err)
	}
}

func TestHookNeverBlocksBecauseDeltaFailed(t *testing.T) {
	_, errOut, err := newMemory().run("not json", "hook")
	if exitCode(t, err) != 1 || !strings.Contains(errOut, "delta hook:") {
		t.Fatalf("err = %v, stderr = %q; a failure of delta must exit 1, not 2", err, errOut)
	}
}

func TestHookTakesNoArguments(t *testing.T) {
	if _, _, err := newMemory().run("{}", "hook", "PreToolUse"); err == nil {
		t.Fatal("hook accepted an argument: the event comes in the JSON")
	}
}
