package vaults

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var ctx = context.Background()

func write(t *testing.T, root, rel, text string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSaveLoadListRoundTrip(t *testing.T) {
	s := New(t.TempDir())
	if names, err := s.List(ctx); err != nil || len(names) != 0 {
		t.Fatalf("an empty project lists %v, %v", names, err)
	}
	for _, name := range []string{"payments", "billing"} {
		if err := s.Save(ctx, name, []byte("spec of "+name)); err != nil {
			t.Fatal(err)
		}
	}
	names, err := s.List(ctx)
	if err != nil || !reflect.DeepEqual(names, []string{"billing", "payments"}) {
		t.Fatalf("List = %v, %v", names, err)
	}
	got, err := s.Load(ctx, "billing")
	if err != nil || string(got) != "spec of billing" {
		t.Fatalf("Load = %q, %v", got, err)
	}
}

func TestSaveReplacesAndLeavesNoTempFile(t *testing.T) {
	root := t.TempDir()
	s := New(root)
	if err := s.Save(ctx, "a", []byte("one")); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(ctx, "a", []byte("two")); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Load(ctx, "a"); string(got) != "two" {
		t.Fatalf("got %q after the second save", got)
	}
	entries, _ := os.ReadDir(filepath.Join(root, "deltas"))
	if len(entries) != 1 {
		t.Fatalf("deltas/ holds %d entries after saving, want only the spec", len(entries))
	}
}

func TestCreateRefusesToOverwrite(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Create(ctx, "a", []byte("one")); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, "a", []byte("two")); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("err = %v, want an already exists error", err)
	}
	if got, _ := s.Load(ctx, "a"); string(got) != "one" {
		t.Fatalf("the existing spec was changed to %q", got)
	}
}

func TestLoadMissingSliceSaysWhere(t *testing.T) {
	_, err := New(t.TempDir()).Load(ctx, "ghost")
	if err == nil || !strings.Contains(err.Error(), "deltas/ghost.spec.md") {
		t.Fatalf("err = %v, want the expected path in the message", err)
	}
}

func TestSliceNamesCannotEscapeDeltas(t *testing.T) {
	root := t.TempDir()
	s := New(root)
	for _, bad := range []string{"../evil", "a/../../evil", "/abs", `a\b`, "", ".", "A", "a/B"} {
		if err := s.Save(ctx, bad, []byte("x")); err == nil {
			t.Errorf("Save accepted %q", bad)
		}
		if _, err := s.Load(ctx, bad); err == nil {
			t.Errorf("Load accepted %q", bad)
		}
		if err := s.Create(ctx, bad, []byte("x")); err == nil {
			t.Errorf("Create accepted %q", bad)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "evil.spec.md")); err == nil {
		t.Fatal("a file was written outside deltas/")
	}
}

func TestListIgnoresWhatIsNotASpec(t *testing.T) {
	root := t.TempDir()
	write(t, root, "deltas/a.spec.md", "x")
	write(t, root, "deltas/notes.md", "x")
	write(t, root, "deltas/b.design.md", "x")
	write(t, root, "deltas/.delta-123.tmp", "x")
	if err := os.MkdirAll(filepath.Join(root, "deltas", "dir.spec.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	names, err := New(root).List(ctx)
	if err != nil || !reflect.DeepEqual(names, []string{"a"}) {
		t.Fatalf("List = %v, %v", names, err)
	}
}

func TestTestFilesFindsTestsAndSkipsVendoredCode(t *testing.T) {
	root := t.TempDir()
	write(t, root, "billing/total_test.go", "go")
	write(t, root, "web/src/cart.test.ts", "ts")
	write(t, root, "billing/total.go", "not a test")
	write(t, root, "node_modules/lib/x.test.js", "vendored")
	write(t, root, "vendor/y_test.go", "vendored")
	write(t, root, ".git/hooks/z_test.go", "git")
	write(t, root, "deltas/a.spec.md", "contract")
	got, err := New(root).TestFiles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"billing/total_test.go", "web/src/cart.test.ts"}
	var paths []string
	for _, f := range got {
		paths = append(paths, f.Path)
	}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
	if got[0].Text != "go" {
		t.Fatalf("text = %q", got[0].Text)
	}
}

func TestFindRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := FindRoot(deep); got != root {
		t.Fatalf("FindRoot(%s) = %s, want %s", deep, got, root)
	}
}

func TestFindRootFallsBackToStart(t *testing.T) {
	// a fresh temp dir sits under no repository on a normal machine
	dir := t.TempDir()
	if got := FindRoot(dir); got != dir {
		if _, err := os.Stat(filepath.Join(got, ".git")); err != nil {
			t.Fatalf("FindRoot(%s) = %s, which has no .git", dir, got)
		}
	}
}

func TestListFindsEveryDeltasFolderAndSkipsVendoredOnes(t *testing.T) {
	root := t.TempDir()
	write(t, root, "deltas/own.spec.md", "x")
	write(t, root, "billing/deltas/invoicing.spec.md", "x")
	write(t, root, "services/auth/deltas/login.spec.md", "x")
	write(t, root, "node_modules/pkg/deltas/lib.spec.md", "x")
	write(t, root, ".claude/deltas/hidden.spec.md", "x")
	names, err := New(root).List(ctx)
	want := []string{"billing/invoicing", "own", "services/auth/login"}
	if err != nil || !reflect.DeepEqual(names, want) {
		t.Fatalf("List = %v, %v; want %v", names, err, want)
	}
	if p := New(root).Path("billing/invoicing"); p != "billing/deltas/invoicing.spec.md" {
		t.Fatalf("Path = %q", p)
	}
}

func TestANameWorksBareWhenOnlyOneDeltasHasIt(t *testing.T) {
	root := t.TempDir()
	write(t, root, "billing/deltas/invoicing.spec.md", "billing")
	write(t, root, "billing/deltas/refunds.spec.md", "one")
	write(t, root, "shop/deltas/refunds.spec.md", "two")
	s := New(root)
	if id, err := s.Resolve(ctx, "invoicing"); err != nil || id != "billing/invoicing" {
		t.Fatalf("Resolve(invoicing) = %q, %v", id, err)
	}
	if src, err := s.Load(ctx, "invoicing"); err != nil || string(src) != "billing" {
		t.Fatalf("Load(invoicing) = %q, %v", src, err)
	}
	if _, err := s.Resolve(ctx, "refunds"); err == nil || !strings.Contains(err.Error(), "billing/refunds") {
		t.Fatalf("a name in two folders must be ambiguous and list them: %v", err)
	}
	if src, err := s.Load(ctx, "shop/refunds"); err != nil || string(src) != "two" {
		t.Fatalf("Load(shop/refunds) = %q, %v", src, err)
	}
}

func TestTheProjectsOwnSliceWinsOverANestedOneWithTheSameName(t *testing.T) {
	root := t.TempDir()
	write(t, root, "deltas/refunds.spec.md", "own")
	write(t, root, "shop/deltas/refunds.spec.md", "nested")
	if src, err := New(root).Load(ctx, "refunds"); err != nil || string(src) != "own" {
		t.Fatalf("Load(refunds) = %q, %v", src, err)
	}
}

func TestCreateAndSaveGoToTheNamedFolder(t *testing.T) {
	root := t.TempDir()
	s := New(root)
	if err := s.Create(ctx, "billing/invoicing", []byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, "billing/invoicing", []byte("b")); err == nil {
		t.Fatal("a second Create overwrote")
	}
	if err := s.Save(ctx, "invoicing", []byte("c")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, "billing", "deltas", "invoicing.spec.md"))
	if err != nil || string(got) != "c" {
		t.Fatalf("file = %q, %v: a bare name must save to the spec that has it", got, err)
	}
	if _, err := os.Stat(filepath.Join(root, "deltas")); err == nil {
		t.Fatal("a root deltas/ was created for a nested slice")
	}
}
