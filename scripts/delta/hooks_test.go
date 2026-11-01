package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	p.write("deltas/billing.spec.md", spec("billing", claimOf("INV-01", "none")))

	code, out, errOut := p.delta(hookEvent(t, "SessionStart", "", "", ""), "hook")
	if code != 0 || errOut != "" || !strings.Contains(out, "- billing: approved, 1 claims (0 test, 0 manual, 1 none)") {
		t.Fatalf("SessionStart = %d %q %q", code, out, errOut)
	}

	code, _, errOut = p.delta(hookEvent(t, "PreToolUse", "Write", filepath.Join(p.root, "deltas", "billing.spec.md"), ""), "hook")
	if code != 2 || !strings.Contains(errOut, "deltas/billing.spec.md") || !strings.Contains(errOut, "delta spec write") {
		t.Fatalf("a write into deltas/ = %d %q, want a block that says how to write it", code, errOut)
	}
	if code, _, errOut = p.delta(hookEvent(t, "PreToolUse", "Write", filepath.Join(p.root, "main.go"), ""), "hook"); code != 0 || errOut != "" {
		t.Fatalf("a write to main.go = %d %q, want it allowed", code, errOut)
	}

	code, _, errOut = p.delta(hookEvent(t, "PreToolUse", "Bash", "", `git commit -m "Fixed it."`), "hook")
	if code != 2 || !strings.Contains(errOut, "COMMIT-LENGTH") {
		t.Fatalf("a bad commit = %d %q", code, errOut)
	}
	good := "git add . && git commit -m \"$(cat <<'EOF'\nfeat(domain): add spec parser and claim rules\n\nWhy.\nEOF\n)\""
	if code, _, errOut = p.delta(hookEvent(t, "PreToolUse", "Bash", "", good), "hook"); code != 0 || errOut != "" {
		t.Fatalf("a good commit = %d %q, want it allowed", code, errOut)
	}
}

func TestHookInputThatIsNotJSONNeverBlocks(t *testing.T) {
	if code, _, errOut := newProject(t).delta("not json", "hook"); code != 1 || !strings.HasPrefix(errOut, "delta hook: ") {
		t.Fatalf("code = %d, stderr = %q; Claude Code blocks only on 2", code, errOut)
	}
}

// isolatedGit skips the test without git, and keeps the machine's own git configuration out of it.
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
	root := t.TempDir()
	if out, err := git(t, root, "init", "-q"); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return &project{t: t, root: root}
}

func TestInitInstallsTheHookAndPointsGitAtIt(t *testing.T) {
	p := gitProject(t)
	code, out, errOut := p.delta("", "init")
	if code != 0 || out != "installed .githooks/commit-msg\nset core.hooksPath to .githooks\n" {
		t.Fatalf("init = %d %q %q", code, out, errOut)
	}
	if !strings.HasPrefix(p.read(".githooks/commit-msg"), "#!/bin/sh\n") {
		t.Fatal("the hook was not installed")
	}
	if path, _ := git(t, p.root, "config", "--get", "core.hooksPath"); strings.TrimSpace(path) != ".githooks" {
		t.Fatalf("core.hooksPath = %q", path)
	}
	if runtime.GOOS != "windows" {
		if info, err := os.Stat(filepath.Join(p.root, ".githooks", "commit-msg")); err != nil || info.Mode().Perm()&0o111 == 0 {
			t.Fatalf("the hook is not executable: %v, %v", info, err)
		}
	}

	code, out, _ = p.delta("", "init")
	if code != 0 || !strings.Contains(out, "already installed") || !strings.Contains(out, "already points") {
		t.Fatalf("a second init = %d %q", code, out)
	}
}

func TestInitDoesNotOverrideAnotherHooksPath(t *testing.T) {
	p := gitProject(t)
	if out, err := git(t, p.root, "config", "--local", "core.hooksPath", ".husky"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	code, out, _ := p.delta("", "init")
	if code != 0 || !strings.Contains(out, "core.hooksPath is .husky: left as it is") {
		t.Fatalf("init = %d %q", code, out)
	}
	if path, _ := git(t, p.root, "config", "--get", "core.hooksPath"); strings.TrimSpace(path) != ".husky" {
		t.Fatalf("core.hooksPath became %q", path)
	}
}

func TestInitOutsideAGitRepositoryWritesNothing(t *testing.T) {
	isolatedGit(t)
	p := &project{t: t, root: t.TempDir()}
	code, _, errOut := p.delta("", "init")
	if code != 2 || !strings.Contains(errOut, "git config") {
		t.Fatalf("init = %d %q, want 2 and the reason from git", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(p.root, ".githooks")); err == nil {
		t.Fatal("a .githooks folder was left behind")
	}
}

func buildDelta(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "delta")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

// The whole loop for a person: delta init, then git refuses a bad message and takes a good one.
func TestTheInstalledHookStopsABadCommitMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the delta binary")
	}
	p := gitProject(t)
	t.Setenv("PATH", filepath.Dir(buildDelta(t))+string(os.PathListSeparator)+os.Getenv("PATH"))
	if code, _, errOut := p.delta("", "init"); code != 0 {
		t.Fatalf("init = %d %q", code, errOut)
	}
	p.write("a.txt", "x")
	if out, err := git(t, p.root, "add", "a.txt"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}

	out, err := git(t, p.root, "commit", "-m", "Fixed it.")
	if err == nil || !strings.Contains(out, "COMMIT-") {
		t.Fatalf("a bad message got through: %v\n%s", err, out)
	}
	if out, err := git(t, p.root, "commit", "-m", "feat(files): add the first file to the project"); err != nil {
		t.Fatalf("a good message was refused: %v\n%s", err, out)
	}
}

// Without delta on the PATH the hook warns and lets the commit through, rather than blocking the person.
func TestTheHookLetsACommitThroughWhenDeltaIsMissing(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh is not on the PATH")
	}
	hook, err := os.ReadFile(filepath.Join("..", "..", "source", "harness", "core", "githooks", "commit-msg"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	script, msg := filepath.Join(dir, "commit-msg"), filepath.Join(dir, "MSG")
	if err := os.WriteFile(script, hook, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(msg, []byte("whatever"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd := exec.Command(sh, script, msg)
	cmd.Env = []string{"PATH=" + t.TempDir()}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil || !strings.Contains(stderr.String(), "delta is not installed") {
		t.Fatalf("err = %v, stderr = %q; want exit 0 with a warning", err, stderr.String())
	}
}
