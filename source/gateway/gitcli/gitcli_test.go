package gitcli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
)

var ctx = context.Background()

// everything is the scope of a slice whose commands name no path: every code file.
var everything = linkcheck.NewScope("", nil)

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

func commit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{{"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "x"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func put(t *testing.T, dir, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTreeIsCleanRightAfterACommit(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.txt", "one")
	commit(t, dir)
	if hash, dirty, err := New(dir).Tree(ctx, everything); err != nil || dirty || hash != "" {
		t.Fatalf("Tree = %q, %v, %v; want clean", hash, dirty, err)
	}
}

func TestTreeChangesWithTheContentAndComesBack(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.txt", "one")
	commit(t, dir)
	g := New(dir)

	put(t, dir, "a.txt", "two")
	first, dirty, err := g.Tree(ctx, everything)
	if err != nil || !dirty || first == "" {
		t.Fatalf("a modified file: %q, %v, %v", first, dirty, err)
	}
	if again, _, _ := g.Tree(ctx, everything); again != first {
		t.Fatal("the same tree gave two fingerprints")
	}
	put(t, dir, "a.txt", "three")
	if other, _, _ := g.Tree(ctx, everything); other == first {
		t.Fatal("another content gave the same fingerprint")
	}
	put(t, dir, "a.txt", "one")
	if hash, dirty, _ := g.Tree(ctx, everything); dirty || hash != "" {
		t.Fatalf("reverting the edit left the tree dirty: %q", hash)
	}
}

func TestTreeSeesUntrackedStagedAndDeletedFiles(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.txt", "one")
	put(t, dir, "b.txt", "two")
	commit(t, dir)
	g := New(dir)

	put(t, dir, "new.txt", "x")
	untracked, dirty, _ := g.Tree(ctx, everything)
	if !dirty {
		t.Fatal("an untracked file left the tree clean")
	}
	put(t, dir, ".gitignore", "ignored.txt\n")
	put(t, dir, "ignored.txt", "x")
	withIgnore, _, _ := g.Tree(ctx, everything)
	put(t, dir, "ignored.txt", "changed")
	if again, _, _ := g.Tree(ctx, everything); again != withIgnore {
		t.Fatal("an ignored file changed the fingerprint")
	}
	if withIgnore == untracked {
		t.Fatal("a new .gitignore did not change the fingerprint")
	}

	if err := os.Remove(filepath.Join(dir, "b.txt")); err != nil {
		t.Fatal(err)
	}
	if hash, _, _ := g.Tree(ctx, everything); hash == withIgnore {
		t.Fatal("a deleted file did not change the fingerprint")
	}

	if out, err := exec.Command("git", "-C", dir, "add", "-A").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if _, dirty, _ := g.Tree(ctx, everything); !dirty {
		t.Fatal("staged changes left the tree clean")
	}
}

func TestTreeOfARepositoryWithoutCommits(t *testing.T) {
	dir := repo(t)
	g := New(dir)
	if hash, dirty, err := g.Tree(ctx, everything); err != nil || dirty || hash != "" {
		t.Fatalf("an empty repository: %q, %v, %v", hash, dirty, err)
	}
	put(t, dir, "a.txt", "one")
	if hash, dirty, err := g.Tree(ctx, everything); err != nil || !dirty || hash == "" {
		t.Fatalf("a first file: %q, %v, %v", hash, dirty, err)
	}
}

func TestTreeOutsideARepositoryFails(t *testing.T) {
	repo(t)
	if _, _, err := New(t.TempDir()).Tree(ctx, everything); err == nil {
		t.Fatal("no error outside a repository")
	}
}

func TestTreeLooksOnlyAtTheScopeAndIgnoresDocumentation(t *testing.T) {
	dir := repo(t)
	put(t, dir, "billing.go", "one")
	put(t, dir, "payments.go", "one")
	put(t, dir, "README.md", "one")
	commit(t, dir)
	g := New(dir)
	billing := linkcheck.NewScope("deltas/specs/billing.spec.md", []string{"billing.go"})

	put(t, dir, "README.md", "two")
	if hash, dirty, _ := g.Tree(ctx, everything); dirty || hash != "" {
		t.Fatal("a change to documentation made the tree dirty")
	}
	put(t, dir, "payments.go", "two")
	if _, dirty, _ := g.Tree(ctx, billing); dirty {
		t.Fatal("a change to payments.go is outside what billing depends on")
	}
	if _, dirty, _ := g.Tree(ctx, everything); !dirty {
		t.Fatal("a change to code is not seen by the whole-project scope")
	}
	put(t, dir, "billing.go", "two")
	if _, dirty, _ := g.Tree(ctx, billing); !dirty {
		t.Fatal("a change inside the scope was missed")
	}
}

func TestFilesFingerprintsEveryCoveredFileEvenRightAfterACommit(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.go", "one")
	put(t, dir, "notes.md", "docs")
	commit(t, dir)
	put(t, dir, "b.go", "untracked")
	put(t, dir, ".gitignore", "ignored.go\n")
	put(t, dir, "ignored.go", "x")
	files, err := New(dir).Files(ctx, everything)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := files["a.go"]; !ok || files["b.go"] == "" || len(files) != 3 { // a.go, b.go and .gitignore; not notes.md nor ignored.go
		t.Fatalf("Files = %v, want a.go, b.go and .gitignore", files)
	}
}

func TestFilesChangeWithTheContentOfAFileOnly(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.go", "one")
	put(t, dir, "b.go", "one")
	commit(t, dir)
	g := New(dir)
	before, _ := g.Files(ctx, everything)
	put(t, dir, "a.go", "two")
	after, err := g.Files(ctx, everything)
	if err != nil || before["a.go"] == after["a.go"] || before["b.go"] != after["b.go"] {
		t.Fatalf("before %v, after %v, err %v; only a.go should differ", before, after, err)
	}
	if err := os.Remove(filepath.Join(dir, "b.go")); err != nil {
		t.Fatal(err)
	}
	if gone, _ := g.Files(ctx, everything); gone["b.go"] == after["b.go"] {
		t.Fatal("a deleted file must not keep the hash it had")
	}
}

func TestFilesKeepsToTheScope(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.go", "one")
	if err := os.Mkdir(filepath.Join(dir, "billing"), 0o755); err != nil {
		t.Fatal(err)
	}
	put(t, dir, "billing/total.go", "one")
	commit(t, dir)
	files, err := New(dir).Files(ctx, linkcheck.NewScope("", []string{"billing"}))
	if err != nil || len(files) != 1 || files["billing/total.go"] == "" {
		t.Fatalf("Files = %v, %v; want only billing/total.go", files, err)
	}
}
