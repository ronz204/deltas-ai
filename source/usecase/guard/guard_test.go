package guard

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/usecase/status"
)

var ctx = context.Background()

type specs map[string]string

func (s specs) Path(slice string) string { return "deltas/" + slice + ".spec.md" }
func (s specs) List(context.Context) ([]string, error) {
	var names []string
	for _, n := range []string{"billing", "payments"} {
		if _, ok := s[n]; ok {
			names = append(names, n)
		}
	}
	return names, nil
}
func (s specs) Load(_ context.Context, slice string) ([]byte, error) {
	src, ok := s[slice]
	if !ok {
		return nil, errors.New("missing")
	}
	return []byte(src), nil
}

func newGuard(t *testing.T, s specs) (*Guard, string) {
	t.Helper()
	root := t.TempDir()
	return New(root, status.New(s), nil), root
}

func TestFileToolsCannotEditTheKnowledgeBase(t *testing.T) {
	g, root := newGuard(t, nil)
	blocked := []string{"deltas/billing.spec.md", "CLAUDE.md", ".claude/rules/x.md", ".claude/skills/quill/SKILL.md", "svc/deltas/a.spec.md"}
	for _, tool := range []string{"Write", "Edit"} {
		for _, rel := range blocked {
			d, err := g.Decide(ctx, Event{Name: "PreToolUse", Tool: tool, Path: filepath.Join(root, filepath.FromSlash(rel))})
			if err != nil || !d.Block || !strings.Contains(d.Reason, rel) || !strings.Contains(d.Reason, "delta spec write") {
				t.Errorf("%s on %s = %+v, %v; want a block that names the file and the way to write it", tool, rel, d, err)
			}
		}
	}
}

func TestOtherFilesAndToolsAreAllowed(t *testing.T) {
	g, root := newGuard(t, nil)
	other := t.TempDir()
	cases := []Event{
		{Name: "PreToolUse", Tool: "Write", Path: filepath.Join(root, "main.go")},
		{Name: "PreToolUse", Tool: "Edit", Path: filepath.Join(root, ".claude", "settings.json")},
		{Name: "PreToolUse", Tool: "Edit", Path: filepath.Join(other, "deltas", "x.spec.md")}, // outside the project
		{Name: "PreToolUse", Tool: "Read", Path: filepath.Join(root, "deltas", "billing.spec.md")},
		{Name: "PreToolUse", Tool: "Write"},
		{Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "deltas", "billing.spec.md")},
		{Name: "Stop"},
		{Name: "SomethingNew"},
	}
	for _, e := range cases {
		if d, err := g.Decide(ctx, e); err != nil || d.Block || d.Context != "" {
			t.Errorf("%+v = %+v, %v; want it allowed", e, d, err)
		}
	}
}

func TestARelativePathIsTakenFromTheRoot(t *testing.T) {
	g, _ := newGuard(t, nil)
	d, _ := g.Decide(ctx, Event{Name: "PreToolUse", Tool: "Write", Path: "deltas/a.spec.md"})
	if !d.Block {
		t.Fatal("a relative path into deltas/ was allowed")
	}
}

func TestCommitsAreCheckedBeforeTheyHappen(t *testing.T) {
	g, _ := newGuard(t, nil)
	bad, err := g.Decide(ctx, Event{Name: "PreToolUse", Tool: "Bash", Command: `git add . && git commit -m "Fixed it."`})
	if err != nil || !bad.Block || !strings.Contains(bad.Reason, "COMMIT-") || !strings.Contains(bad.Reason, "commit again") {
		t.Fatalf("a bad message = %+v, %v", bad, err)
	}
	allowed := []string{
		`git commit -m "feat(domain): add spec parser and claim rules"`,
		"git commit -m \"$(cat <<'EOF'\nfeat(domain): add spec parser and claim rules\n\nWhy.\nEOF\n)\"",
		`git commit --amend`,
		`git commit -F msg.txt`,
		`git status`,
		`go test ./...`,
	}
	for _, command := range allowed {
		if d, err := g.Decide(ctx, Event{Name: "PreToolUse", Tool: "Bash", Command: command}); err != nil || d.Block {
			t.Errorf("%q = %+v, %v; want it allowed", command, d, err)
		}
	}
}

func TestSessionStartTellsTheSlices(t *testing.T) {
	g, _ := newGuard(t, specs{
		"billing":  "---\nformat: 1\nslice: billing\nstatus: approved\n---\n- INV-01 [test: go test ./a]\n  When x, y.\n- INV-02 [none]\n  When x, z.\n",
		"payments": "not a spec",
	})
	d, err := g.Decide(ctx, Event{Name: "SessionStart"})
	if err != nil || d.Block {
		t.Fatalf("%+v, %v", d, err)
	}
	for _, want := range []string{
		"- billing: approved, 2 claims (1 test, 0 manual, 1 none)",
		"- payments: unreadable, 0 claims (0 test, 0 manual, 0 none), 1 problems in its spec",
		"delta spec write",
	} {
		if !strings.Contains(d.Context, want) {
			t.Errorf("the context lacks %q:\n%s", want, d.Context)
		}
	}
}

func TestSessionStartSaysNothingWithoutSlices(t *testing.T) {
	g, _ := newGuard(t, nil)
	if d, err := g.Decide(ctx, Event{Name: "SessionStart"}); err != nil || d.Context != "" || d.Block {
		t.Fatalf("%+v, %v", d, err)
	}
}

type health []string

func (h health) Check(context.Context) []string { return h }

func TestSessionStartAddsTheSetupProblems(t *testing.T) {
	root := t.TempDir()
	g := New(root, status.New(specs{"billing": "---\nformat: 1\nslice: billing\nstatus: approved\n---\n"}), health{"git does not check commit messages yet: run delta init"})
	d, err := g.Decide(ctx, Event{Name: "SessionStart"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"- billing: approved", "Delta setup problems:", "- git does not check commit messages yet: run delta init"} {
		if !strings.Contains(d.Context, want) {
			t.Errorf("the context lacks %q:\n%s", want, d.Context)
		}
	}
}

func TestSessionStartWithOnlySetupProblemsStillSpeaks(t *testing.T) {
	g := New(t.TempDir(), status.New(specs{}), health{"delta is not on the PATH, so the git hook cannot run it"})
	d, _ := g.Decide(ctx, Event{Name: "SessionStart"})
	if strings.Contains(d.Context, "Delta slices:") || !strings.Contains(d.Context, "delta is not on the PATH") {
		t.Fatalf("context = %q, want the problem and no slice list", d.Context)
	}
}

func TestSessionStartStaysQuietWhenAllIsWell(t *testing.T) {
	g := New(t.TempDir(), status.New(specs{}), health{})
	if d, _ := g.Decide(ctx, Event{Name: "SessionStart"}); d.Context != "" {
		t.Fatalf("context = %q, want none", d.Context)
	}
}
