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
	if err == nil || !strings.Contains(err.Error(), "deltas/specifies/ghost.spec.md") {
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
	write(t, root, "deltas/specifies/a.spec.md", "x")
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
	write(t, root, "deltas/specifies/a.spec.md", "contract")
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

func TestListGivesTheSlicesSortedAndNothingWithoutASpecifiesFolder(t *testing.T) {
	root := t.TempDir()
	if names, err := New(root).List(ctx); err != nil || len(names) != 0 {
		t.Fatalf("List without specifies/ = %v, %v", names, err)
	}
	write(t, root, "deltas/specifies/payments.spec.md", "x")
	write(t, root, "deltas/specifies/billing.spec.md", "x")
	write(t, root, "billing/deltas/specifies/stray.spec.md", "x") // a deltas/ folder elsewhere is not read
	names, err := New(root).List(ctx)
	if want := []string{"billing", "payments"}; err != nil || !reflect.DeepEqual(names, want) {
		t.Fatalf("List = %v, %v; want %v", names, err, want)
	}
}

func TestANameWithAFolderIsNotASlice(t *testing.T) {
	s := New(t.TempDir())
	for _, name := range []string{"billing/invoicing", "../x", "", "A"} {
		if err := s.Create(ctx, name, []byte("a")); err == nil {
			t.Errorf("Create(%q) was accepted", name)
		}
		if _, err := s.Load(ctx, name); err == nil || !strings.Contains(err.Error(), "not a slice name") {
			t.Errorf("Load(%q) err = %v", name, err)
		}
	}
}

func TestCreateAndSaveGoToSpecifies(t *testing.T) {
	root := t.TempDir()
	s := New(root)
	if err := s.Create(ctx, "invoicing", []byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, "invoicing", []byte("b")); err == nil {
		t.Fatal("a second Create overwrote")
	}
	if err := s.Save(ctx, "invoicing", []byte("c")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, "deltas", "specifies", "invoicing.spec.md"))
	if err != nil || string(got) != "c" {
		t.Fatalf("file = %q, %v", got, err)
	}
}
