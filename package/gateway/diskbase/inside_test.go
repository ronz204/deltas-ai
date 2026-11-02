package diskbase

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteFileThenReadFile(t *testing.T) {
	root := t.TempDir()
	v := New(root)
	if err := v.WriteFile(ctx, ".claude/rules/x.md", []byte("#!/bin/sh\n")); err != nil {
		t.Fatal(err)
	}
	got, err := v.ReadFile(ctx, ".claude/rules/x.md")
	if err != nil || string(got) != "#!/bin/sh\n" {
		t.Fatalf("ReadFile = %q, %v", got, err)
	}
	if err := v.WriteFile(ctx, ".claude/rules/x.md", []byte("new")); err != nil {
		t.Fatal(err)
	}
	if got, _ := v.ReadFile(ctx, ".claude/rules/x.md"); string(got) != "new" {
		t.Fatalf("a second write left %q", got)
	}
	entries, _ := os.ReadDir(filepath.Join(root, ".claude/rules"))
	if len(entries) != 1 {
		t.Fatalf(".claude/rules holds %d entries, want only the rule (no temporary file)", len(entries))
	}
}

func TestReadFileOfAMissingFileWrapsNotExist(t *testing.T) {
	if _, err := New(t.TempDir()).ReadFile(ctx, "nope"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestProjectFilesCannotLeaveTheProject(t *testing.T) {
	root := t.TempDir()
	v := New(root)
	for _, bad := range []string{"", ".", "..", "../evil", "a/../../evil", "/abs", `a\b`, `..\evil`} {
		if err := v.WriteFile(ctx, bad, []byte("x")); err == nil {
			t.Errorf("WriteFile accepted %q", bad)
		}
		if _, err := v.ReadFile(ctx, bad); err == nil || errors.Is(err, fs.ErrNotExist) {
			t.Errorf("ReadFile(%q) = %v, want a refusal and not a missing file", bad, err)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "evil")); err == nil {
		t.Fatal("a file was written outside the project")
	}
}

func TestGlobReadsWhatMatchesAndSkipsFolders(t *testing.T) {
	root := t.TempDir()
	write(t, root, "CLAUDE.md", "router")
	write(t, root, ".claude/rules/a.md", "rule a")
	write(t, root, ".claude/rules/b.txt", "not a rule")
	write(t, root, ".claude/rules/deep/c.md", "too deep for *")
	write(t, root, ".claude/skills/drafter/SKILL.md", "skill")
	write(t, root, ".claude/skills/drafter/references/guide.md", "reference")
	if err := os.MkdirAll(filepath.Join(root, ".claude", "agents", "folder.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := New(root).Glob(ctx, []string{"CLAUDE.md", ".claude/CLAUDE.md", ".claude/rules/*.md", ".claude/skills/*/SKILL.md", ".claude/agents/*.md"})
	if err != nil {
		t.Fatal(err)
	}
	var seen []string
	for _, f := range got {
		seen = append(seen, f.Path+"="+f.Text)
	}
	want := []string{"CLAUDE.md=router", ".claude/rules/a.md=rule a", ".claude/skills/drafter/SKILL.md=skill"}
	if !reflect.DeepEqual(seen, want) {
		t.Fatalf("Glob = %v, want %v", seen, want)
	}
}

func TestGlobWithNothingThereIsEmpty(t *testing.T) {
	got, err := New(t.TempDir()).Glob(ctx, []string{"CLAUDE.md", ".claude/rules/*.md"})
	if err != nil || len(got) != 0 {
		t.Fatalf("Glob = %v, %v", got, err)
	}
}

func TestMatchingChoosesWhichFilesAreTests(t *testing.T) {
	root := t.TempDir()
	write(t, root, "spec/user_spec.rb", "# INV-01")
	write(t, root, "user_test.go", "// INV-01")
	own, custom := New(root), New(root).Matching([]string{"*_spec.rb"})
	if !own.IsTest("user_test.go") || own.IsTest("spec/user_spec.rb") {
		t.Error("delta's own rule changed")
	}
	if custom.IsTest("user_test.go") || !custom.IsTest("spec/user_spec.rb") {
		t.Error("the patterns do not replace the default rule")
	}
	files, err := custom.TestFiles(ctx)
	if err != nil || len(files) != 1 || files[0].Path != "spec/user_spec.rb" {
		t.Fatalf("TestFiles = %v, %v; want only the Ruby spec", files, err)
	}
}
