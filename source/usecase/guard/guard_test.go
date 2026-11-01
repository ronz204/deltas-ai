package guard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/source/domain/judgment"
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
	return New(root, status.New(s), nil, nil), root
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
	g := New(root, status.New(specs{"billing": "---\nformat: 1\nslice: billing\nstatus: approved\n---\n"}), health{"git does not check commit messages yet: run delta init"}, nil)
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
	g := New(t.TempDir(), status.New(specs{}), health{"delta is not on the PATH, so the git hook cannot run it"}, nil)
	d, _ := g.Decide(ctx, Event{Name: "SessionStart"})
	if strings.Contains(d.Context, "Delta slices:") || !strings.Contains(d.Context, "delta is not on the PATH") {
		t.Fatalf("context = %q, want the problem and no slice list", d.Context)
	}
}

func TestSessionStartStaysQuietWhenAllIsWell(t *testing.T) {
	g := New(t.TempDir(), status.New(specs{}), health{}, nil)
	if d, _ := g.Decide(ctx, Event{Name: "SessionStart"}); d.Context != "" {
		t.Fatalf("context = %q, want none", d.Context)
	}
}

type evidence struct {
	dirty bool
	by    map[string]judgment.Standing
	err   error
	asked []string
}

func (e *evidence) Standings(_ context.Context, slices []string) (bool, map[string]judgment.Standing, error) {
	e.asked = slices
	return e.dirty, e.by, e.err
}

const withTests = "---\nformat: 1\nslice: %s\nstatus: approved\n---\n- INV-01 [test: go test ./%s]\n  When x, y.\n"

func stopGuard(t *testing.T, ev Evidence, slices ...string) *Guard {
	t.Helper()
	s := specs{}
	for _, name := range slices {
		s[name] = fmt.Sprintf(withTests, name, name)
	}
	return New(t.TempDir(), status.New(s), nil, ev)
}

func TestStopAsksForACheckOfEverySliceThatIsNotCheckedYet(t *testing.T) {
	ev := &evidence{dirty: true, by: map[string]judgment.Standing{"billing": judgment.Unchecked, "payments": judgment.Passing}}
	d, err := stopGuard(t, ev, "billing", "payments").Decide(ctx, Event{Name: "Stop"})
	if err != nil || !d.Block || !strings.Contains(d.Reason, "billing") || strings.Contains(d.Reason, "payments") || !strings.Contains(d.Reason, "delta check") {
		t.Fatalf("Decide = %+v, %v; want a block that names only billing", d, err)
	}
	if !reflect.DeepEqual(ev.asked, []string{"billing", "payments"}) {
		t.Fatalf("asked about %v", ev.asked)
	}
}

func TestStopSaysWhenAClaimWasViolated(t *testing.T) {
	ev := &evidence{dirty: true, by: map[string]judgment.Standing{"billing": judgment.Failing}}
	d, _ := stopGuard(t, ev, "billing").Decide(ctx, Event{Name: "Stop"})
	if !d.Block || !strings.Contains(d.Reason, "violated") || !strings.Contains(d.Reason, "billing") {
		t.Fatalf("Decide = %+v, want a block about violated claims", d)
	}
}

func TestStopLetsClaudeFinishWhenThereIsNothingToProve(t *testing.T) {
	passing := &evidence{dirty: true, by: map[string]judgment.Standing{"billing": judgment.Passing}}
	cases := map[string]*Guard{
		"every slice passes":           stopGuard(t, passing, "billing"),
		"the tree has no changes":      stopGuard(t, &evidence{dirty: false}, "billing"),
		"no slice has a test claim":    New(t.TempDir(), status.New(specs{"billing": "---\nformat: 1\nslice: billing\nstatus: approved\n---\n- INV-01 [none]\n  When x, y.\n"}), nil, &evidence{dirty: true}),
		"there are no slices":          stopGuard(t, &evidence{dirty: true}),
		"there is no evidence to read": stopGuard(t, nil, "billing"),
	}
	for name, g := range cases {
		if d, err := g.Decide(ctx, Event{Name: "Stop"}); err != nil || d.Block {
			t.Errorf("%s: Decide = %+v, %v; want it allowed", name, d, err)
		}
	}
}

func TestStopAsksOnlyOnce(t *testing.T) {
	ev := &evidence{dirty: true, by: map[string]judgment.Standing{"billing": judgment.Unchecked}}
	d, err := stopGuard(t, ev, "billing").Decide(ctx, Event{Name: "Stop", StopActive: true})
	if err != nil || d.Block || ev.asked != nil {
		t.Fatalf("Decide = %+v, %v (asked %v); a Stop already continuing must be let through without a look", d, err, ev.asked)
	}
}

func TestStopFailureIsAnErrorNotABlock(t *testing.T) {
	ev := &evidence{err: errors.New("git is gone")}
	if d, err := stopGuard(t, ev, "billing").Decide(ctx, Event{Name: "Stop"}); err == nil || d.Block {
		t.Fatalf("Decide = %+v, %v; a broken delta must not block", d, err)
	}
}
