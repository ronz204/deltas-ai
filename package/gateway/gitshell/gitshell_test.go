package gitshell

import (
	"context"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ronz204/deltas-ai/package/domain/diffing"
	"github.com/ronz204/deltas-ai/package/domain/linking"
)

var ctx = context.Background()

var everything = linking.NewScope("", nil, nil)

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

func TestTreeIgnoresDocumentation(t *testing.T) {
	dir := repo(t)
	put(t, dir, "billing.go", "one")
	put(t, dir, "payments.go", "one")
	put(t, dir, "README.md", "one")
	commit(t, dir)
	g := New(dir)

	put(t, dir, "README.md", "two")
	if hash, dirty, _ := g.Tree(ctx, everything); dirty || hash != "" {
		t.Fatal("a change to documentation made the tree dirty")
	}
	put(t, dir, "payments.go", "two")
	if _, dirty, _ := g.Tree(ctx, everything); !dirty {
		t.Fatal("a change to code is not seen by the whole-project scope")
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
	if _, ok := files["a.go"]; !ok || files["b.go"] == "" || len(files) != 3 {
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

func touched(t *testing.T, g *Git, b diffing.Baseline) []string {
	t.Helper()
	names, err := g.Since(ctx, b)
	if err != nil {
		t.Fatal(err)
	}
	return names
}

func TestSinceSeesOnlyWhatChangedAfterTheBaseline(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.go", "one")
	put(t, dir, "b.go", "one")
	put(t, dir, "dirty.go", "one")
	commit(t, dir)
	put(t, dir, "dirty.go", "already edited")
	put(t, dir, "wip.go", "already new")
	g := New(dir)
	b, err := g.Baseline(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := touched(t, g, b); len(got) != 0 {
		t.Fatalf("right after the baseline: %v, want nothing", got)
	}

	put(t, dir, "a.go", "two")
	put(t, dir, "fresh.go", "x")
	if got := touched(t, g, b); !slices.Equal(got, []string{"a.go", "fresh.go"}) {
		t.Fatalf("clean file edited and a new one: %v", got)
	}

	put(t, dir, "dirty.go", "edited again")
	put(t, dir, "wip.go", "changed")
	if got := touched(t, g, b); !slices.Equal(got, []string{"a.go", "dirty.go", "fresh.go", "wip.go"}) {
		t.Fatalf("files that were already dirty and changed again: %v", got)
	}

	put(t, dir, "a.go", "one")
	if got := touched(t, g, b); slices.Contains(got, "a.go") {
		t.Fatalf("an edit that was reverted still counts: %v", got)
	}
}

func TestSinceSeesWhatWasCommittedAfterTheBaseline(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.go", "one")
	commit(t, dir)
	g := New(dir)
	b, _ := g.Baseline(ctx)

	put(t, dir, "a.go", "two")
	put(t, dir, "b.go", "x")
	commit(t, dir)
	if got := touched(t, g, b); !slices.Equal(got, []string{"a.go", "b.go"}) {
		t.Fatalf("committed after the baseline: %v", got)
	}
}

func TestSinceSeesDeletedFiles(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.go", "one")
	commit(t, dir)
	g := New(dir)
	b, _ := g.Baseline(ctx)
	if err := os.Remove(filepath.Join(dir, "a.go")); err != nil {
		t.Fatal(err)
	}
	if got := touched(t, g, b); !slices.Equal(got, []string{"a.go"}) {
		t.Fatalf("a deleted file: %v", got)
	}
}

func TestSinceInARepositoryWithoutCommits(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.go", "one")
	g := New(dir)
	b, err := g.Baseline(ctx)
	if err != nil || b.Head != "" {
		t.Fatalf("Baseline = %+v, %v", b, err)
	}
	if got := touched(t, g, b); len(got) != 0 {
		t.Fatalf("right after the baseline: %v", got)
	}
	put(t, dir, "b.go", "x")
	if got := touched(t, g, b); !slices.Equal(got, []string{"b.go"}) {
		t.Fatalf("a first new file: %v", got)
	}
}

func filesOf(t *testing.T, g *Git) []string {
	t.Helper()
	files, err := g.Files(ctx, everything)
	if err != nil {
		t.Fatal(err)
	}
	return slices.Sorted(maps.Keys(files))
}

func TestFilesLeaveOutLockfilesAndBinariesButKeepManifestsAndTextAssets(t *testing.T) {
	dir := repo(t)
	put(t, dir, "go.mod", "module x")
	put(t, dir, "go.sum", "h1:abc")
	put(t, dir, "package.json", "{}")
	put(t, dir, "logo.png", "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	put(t, dir, "logo.svg", "<svg></svg>")
	put(t, dir, "main.go", "package main")
	if err := os.MkdirAll(filepath.Join(dir, "web"), 0o755); err != nil {
		t.Fatal(err)
	}
	put(t, dir, "web/package-lock.json", "{}")
	commit(t, dir)

	got := filesOf(t, New(dir))
	if want := []string{"go.mod", "logo.svg", "main.go", "package.json"}; !slices.Equal(got, want) {
		t.Fatalf("Files = %v, want %v", got, want)
	}
}

func TestChangingOnlyLockfilesAndBinariesLeavesTheTreeClean(t *testing.T) {
	dir := repo(t)
	put(t, dir, "main.go", "one")
	put(t, dir, "go.sum", "one")
	put(t, dir, "logo.png", "\x00one")
	commit(t, dir)
	g := New(dir)

	put(t, dir, "go.sum", "two")
	put(t, dir, "logo.png", "\x00two")
	put(t, dir, "new.png", "\x00new")
	if hash, dirty, err := g.Tree(ctx, everything); err != nil || dirty || hash != "" {
		t.Fatalf("Tree = %q, %v, %v; lockfiles and binaries are not code", hash, dirty, err)
	}
	put(t, dir, "go.mod", "module x")
	if _, dirty, _ := g.Tree(ctx, everything); !dirty {
		t.Fatal("a new manifest left the tree clean")
	}
}

func TestIgnoringReplacesTheDefaultPatterns(t *testing.T) {
	dir := repo(t)
	put(t, dir, "go.sum", "x")
	if err := os.MkdirAll(filepath.Join(dir, "gen"), 0o755); err != nil {
		t.Fatal(err)
	}
	put(t, dir, "gen/code.go", "x")
	put(t, dir, "main.go", "x")
	commit(t, dir)

	if got := filesOf(t, New(dir).Ignoring([]string{"gen/**"})); !slices.Equal(got, []string{"go.sum", "main.go"}) {
		t.Fatalf("Files = %v; the project's patterns replace the lockfile defaults", got)
	}
}

func TestSinceDoesNotSeeLockfilesOrBinaries(t *testing.T) {
	dir := repo(t)
	put(t, dir, "a.go", "one")
	commit(t, dir)
	put(t, dir, "go.sum", "already there")
	g := New(dir)
	b, err := g.Baseline(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, kept := b.Dirty["go.sum"]; kept {
		t.Fatalf("the baseline kept a lockfile: %v", b.Dirty)
	}

	b.Dirty["go.sum"] = "a baseline from before lockfiles were ignored"
	put(t, dir, "go.sum", "changed")
	put(t, dir, "logo.png", "\x00png")
	put(t, dir, "b.go", "x")
	if got := touched(t, g, b); !slices.Equal(got, []string{"b.go"}) {
		t.Fatalf("Since = %v, want only b.go", got)
	}
}

func TestADeletedBinaryDoesNotBreakAnything(t *testing.T) {
	dir := repo(t)
	put(t, dir, "logo.png", "\x00png")
	put(t, dir, "a.go", "one")
	commit(t, dir)
	g := New(dir)
	b, err := g.Baseline(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "logo.png")); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Since(ctx, b); err != nil {
		t.Fatalf("Since failed on a deleted binary: %v", err)
	}
	if _, _, err := g.Tree(ctx, everything); err != nil {
		t.Fatalf("Tree failed on a deleted binary: %v", err)
	}
}
