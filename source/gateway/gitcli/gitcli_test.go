package gitcli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var ctx = context.Background()

// repo makes an empty repository that ignores the global and system git configuration, so the
// tests do not depend on who runs them.
func repo(t *testing.T) string {
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
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return dir
}

func TestHooksPathIsEmptyUntilItIsSet(t *testing.T) {
	g := New(repo(t))
	if got, err := g.HooksPath(ctx); err != nil || got != "" {
		t.Fatalf("HooksPath = %q, %v; want empty and no error", got, err)
	}
	if err := g.SetHooksPath(ctx, ".githooks"); err != nil {
		t.Fatal(err)
	}
	if got, err := g.HooksPath(ctx); err != nil || got != ".githooks" {
		t.Fatalf("HooksPath = %q, %v; want .githooks", got, err)
	}
}

func TestSetHooksPathWritesTheLocalConfigOnly(t *testing.T) {
	dir := repo(t)
	if err := New(dir).SetHooksPath(ctx, ".githooks"); err != nil {
		t.Fatal(err)
	}
	cfg, err := os.ReadFile(filepath.Join(dir, ".git", "config"))
	if err != nil || !strings.Contains(string(cfg), "hooksPath = .githooks") {
		t.Fatalf("the repository config lacks the setting: %q, %v", cfg, err)
	}
}

func TestHooksPathSeesAValueFromAnotherLevel(t *testing.T) {
	dir := repo(t)
	global := filepath.Join(t.TempDir(), "global")
	if err := os.WriteFile(global, []byte("[core]\n\thooksPath = /company/hooks\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	if got, err := New(dir).HooksPath(ctx); err != nil || got != "/company/hooks" {
		t.Fatalf("HooksPath = %q, %v; want the global value", got, err)
	}
}

func TestSetHooksPathOutsideARepositoryFailsWithGitsReason(t *testing.T) {
	repo(t) // only to isolate the configuration and to skip when git is missing
	err := New(t.TempDir()).SetHooksPath(ctx, ".githooks")
	if err == nil || !strings.Contains(err.Error(), "git config") {
		t.Fatalf("err = %v, want a git config error", err)
	}
}
