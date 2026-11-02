package sentinel

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/package/domain/findings"
	"github.com/ronz204/deltas-ai/package/domain/judgment"
	"github.com/ronz204/deltas-ai/package/domain/linking"
	"github.com/ronz204/deltas-ai/package/usecase/tracking"
)

var ctx = context.Background()

type specs map[string]string

func (s specs) Path(slice string) string { return "deltas/specs/" + slice + ".spec.md" }
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
	return New(root, tracking.New(s), nil, nil, nil), root
}

func TestFileToolsCannotEditTheKnowledgeBase(t *testing.T) {
	g, root := newGuard(t, nil)
	blocked := []string{"deltas/specs/billing.spec.md", "CLAUDE.md", ".claude/rules/x.md", ".claude/skills/drafter/SKILL.md", "svc/deltas/specs/a.spec.md"}
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
		{Name: "PreToolUse", Tool: "Edit", Path: filepath.Join(other, "deltas", "specs", "x.spec.md")},
		{Name: "PreToolUse", Tool: "Read", Path: filepath.Join(root, "deltas", "specs", "billing.spec.md")},
		{Name: "PreToolUse", Tool: "Write"},
		{Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "deltas", "specs", "billing.spec.md")},
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
	d, _ := g.Decide(ctx, Event{Name: "PreToolUse", Tool: "Write", Path: "deltas/specs/a.spec.md"})
	if !d.Block {
		t.Fatal("a relative path into deltas/ was allowed")
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

func TestEveryPromptIsOneFixedLineWhateverTheSlices(t *testing.T) {
	g, _ := newGuard(t, specs{
		"billing":  "---\nformat: 1\nslice: billing\nstatus: draft\n---\n- INV-01 [none]\n  When x, y.\n",
		"payments": "---\nformat: 1\nslice: payments\nstatus: approved\n---\n- INV-01 [none]\n  When x, y.\n",
	})
	d, err := g.Decide(ctx, Event{Name: "UserPromptSubmit"})
	if err != nil || d.Block || d.Context != promptLine {
		t.Fatalf("%+v, %v; want the fixed line", d, err)
	}
	for _, want := range []string{"comper skill", "Bug fixes, refactors and tweaks need no ceremony"} {
		if !strings.Contains(d.Context, want) {
			t.Errorf("the line lacks %q:\n%s", want, d.Context)
		}
	}
}

func TestAPromptWithoutSlicesStillPointsAtHelm(t *testing.T) {
	g, _ := newGuard(t, nil)
	d, err := g.Decide(ctx, Event{Name: "UserPromptSubmit"})
	if err != nil || !strings.Contains(d.Context, "comper skill") || strings.Contains(d.Context, "\n-") {
		t.Fatalf("%+v, %v; want only the pointer to comper", d, err)
	}
}

func TestAVerifiedSliceWaitsForNothing(t *testing.T) {
	g, _ := newGuard(t, specs{"billing": "---\nformat: 1\nslice: billing\nstatus: verified\n---\n- INV-01 [none]\n  When x, y.\n"})
	if d, _ := g.Decide(ctx, Event{Name: "UserPromptSubmit"}); strings.Contains(d.Context, "billing") {
		t.Fatalf("context = %q, want no line for a verified slice", d.Context)
	}
}

type health []string

func (h health) Check(context.Context) []string { return h }

func TestSessionStartAddsTheSetupProblems(t *testing.T) {
	root := t.TempDir()
	g := New(root, tracking.New(specs{"billing": "---\nformat: 1\nslice: billing\nstatus: approved\n---\n"}), health{"git does not check commit messages yet: run delta init"}, nil, nil)
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
	g := New(t.TempDir(), tracking.New(specs{}), health{"delta is not on the PATH, so the hooks of Claude Code cannot run it"}, nil, nil)
	d, _ := g.Decide(ctx, Event{Name: "SessionStart"})
	if strings.Contains(d.Context, "Delta slices:") || !strings.Contains(d.Context, "delta is not on the PATH") {
		t.Fatalf("context = %q, want the problem and no slice list", d.Context)
	}
}

func TestSessionStartStaysQuietWhenAllIsWell(t *testing.T) {
	g := New(t.TempDir(), tracking.New(specs{}), health{}, nil, nil)
	if d, _ := g.Decide(ctx, Event{Name: "SessionStart"}); d.Context != "" {
		t.Fatalf("context = %q, want none", d.Context)
	}
}

type evidence struct {
	by      map[string]judgment.Standing
	err     error
	asked   map[string]linking.Scope
	touched []string
	unknown bool
	began   string
}

func (e *evidence) Standings(_ context.Context, slices map[string]linking.Scope) (map[string]judgment.Standing, error) {
	e.asked = slices
	return e.by, e.err
}

func (e *evidence) Begin(_ context.Context, session string) error {
	e.began = session
	return nil
}

func (e *evidence) Touched(context.Context, string) ([]string, bool, error) {
	return e.touched, !e.unknown, nil
}

const withTests = "---\nformat: 1\nslice: %s\nstatus: built\n---\n- INV-01 [test: go test ./%s]\n  When x, y.\n"

func stopGuard(t *testing.T, ev Evidence, slices ...string) *Guard {
	t.Helper()
	s := specs{}
	for _, name := range slices {
		s[name] = fmt.Sprintf(withTests, name, name)
	}
	return New(t.TempDir(), tracking.New(s), nil, ev, nil)
}

var stop = Event{Name: "Stop", Session: "s1"}

func TestStopAsksForACheckOfTheSlicesThisSessionTouched(t *testing.T) {
	ev := &evidence{
		by:      map[string]judgment.Standing{"billing": judgment.Unchecked, "payments": judgment.Unchecked},
		touched: []string{"billing/total.go"},
	}
	d, err := stopGuard(t, ev, "billing", "payments").Decide(ctx, stop)
	want := "Your changes touched code of billing: run delta check billing before you finish."
	if err != nil || !d.Block || d.Reason != want {
		t.Fatalf("Decide = %+v, %v; want a block that names only billing: %q", d, err, want)
	}
	if len(ev.asked) != 1 || !ev.asked["billing"].Covers("billing/x.go") {
		t.Fatalf("asked about %v, want only billing with the scope its command names", ev.asked)
	}
}

func TestStopNamesEverySliceTheSessionTouched(t *testing.T) {
	ev := &evidence{
		by:      map[string]judgment.Standing{"billing": judgment.Unchecked, "payments": judgment.Unchecked},
		touched: []string{"billing/total.go", "payments/pay.go"},
	}
	d, _ := stopGuard(t, ev, "billing", "payments").Decide(ctx, stop)
	if want := "touched code of billing, payments: run delta check billing payments"; !strings.Contains(d.Reason, want) {
		t.Fatalf("Reason = %q, want %q", d.Reason, want)
	}
}

func TestStopSaysWhenAClaimWasViolated(t *testing.T) {
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Failing}, touched: []string{"billing/total.go"}}
	d, _ := stopGuard(t, ev, "billing").Decide(ctx, stop)
	if !d.Block || !strings.Contains(d.Reason, "violated claims for: billing") || !strings.Contains(d.Reason, "/drafter") {
		t.Fatalf("Decide = %+v, want a block about violated claims", d)
	}
}

func TestStopLetsClaudeFinishWhenThereIsNothingToProve(t *testing.T) {
	touched := []string{"billing/total.go"}
	passing := &evidence{by: map[string]judgment.Standing{"billing": judgment.Passing}, touched: touched}
	none := "---\nformat: 1\nslice: billing\nstatus: built\n---\n- INV-01 [none]\n  When x, y.\n"
	cases := map[string]*Guard{
		"every touched slice stands": stopGuard(t, passing, "billing"),
		"no slice has a test claim":  New(t.TempDir(), tracking.New(specs{"billing": none}), nil, &evidence{touched: touched}, nil),
		"there are no slices":        stopGuard(t, &evidence{touched: touched}),
		"there is no evidence":       stopGuard(t, nil, "billing"),
	}
	for name, g := range cases {
		if d, err := g.Decide(ctx, stop); err != nil || d.Block {
			t.Errorf("%s: Decide = %+v, %v; want it allowed", name, d, err)
		}
	}
}

func TestStopIgnoresChangesOutsideEverySlice(t *testing.T) {
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked}, touched: []string{"README.md", "deltas/specs/x.spec.md", ".claude/rules/a.md"}}
	d, err := stopGuard(t, ev, "billing").Decide(ctx, stop)
	if err != nil || d.Block || ev.asked != nil {
		t.Fatalf("Decide = %+v, %v (asked %v); documentation and what delta installs are not code", d, err, ev.asked)
	}
}

func TestStopIgnoresCodeOfAnotherSlice(t *testing.T) {
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked}, touched: []string{"payments/pay.go"}}
	if d, err := stopGuard(t, ev, "billing").Decide(ctx, stop); err != nil || d.Block || ev.asked != nil {
		t.Fatalf("Decide = %+v, %v (asked %v); payments/ is not what billing's command names", d, err, ev.asked)
	}
}

func TestStopIgnoresCodeWhenNoSliceWasBuilt(t *testing.T) {
	head := func(slice, status string) string {
		return "---\nformat: 1\nslice: " + slice + "\nstatus: " + status + "\n---\n- INV-01 [test: go test ./" + slice + "]\n  When x, y.\n"
	}
	s := specs{"billing": head("billing", "approved"), "payments": head("payments", "draft")}
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked, "payments": judgment.Unchecked}, touched: []string{"billing/total.go", "payments/pay.go", "go.mod"}}
	if d, err := New(t.TempDir(), tracking.New(s), nil, ev, nil).Decide(ctx, stop); err != nil || d.Block {
		t.Fatalf("Decide = %+v, %v; approved and draft slices are backlog, never a reason to hold Claude", d, err)
	}
}

func TestStopAsksForACheckOfAStaleSliceWhoseCodeWasTouched(t *testing.T) {
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked}, touched: []string{"billing/total.go"}}
	g := New(t.TempDir(), tracking.New(verifiedBilling("stale")), nil, ev, nil)
	d, err := g.Decide(ctx, stop)
	if err != nil || !d.Block || !strings.Contains(d.Reason, "delta check billing") || strings.Contains(d.Reason, "/builder") {
		t.Fatalf("Decide = %+v, %v", d, err)
	}
}

func TestStopWithoutASessionBaselineLetsClaudeFinish(t *testing.T) {
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked}, touched: []string{"billing/total.go"}, unknown: true}
	g := stopGuard(t, ev, "billing")
	if d, err := g.Decide(ctx, stop); err != nil || d.Block || ev.asked != nil {
		t.Fatalf("Decide = %+v, %v (asked %v); with no baseline nothing is known to be this session's", d, err, ev.asked)
	}
	if d, err := g.Decide(ctx, Event{Name: "Stop"}); err != nil || d.Block {
		t.Fatalf("an event without a session: %+v, %v", d, err)
	}
}

func TestStopAsksOnlyOnce(t *testing.T) {
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked}, touched: []string{"billing/total.go"}}
	d, err := stopGuard(t, ev, "billing").Decide(ctx, Event{Name: "Stop", Session: "s1", StopActive: true})
	if err != nil || d.Block || ev.asked != nil {
		t.Fatalf("Decide = %+v, %v (asked %v); a Stop already continuing must be let through without a look", d, err, ev.asked)
	}
}

func TestStopFailureIsAnErrorNotABlock(t *testing.T) {
	ev := &evidence{err: errors.New("git is gone"), touched: []string{"billing/total.go"}}
	if d, err := stopGuard(t, ev, "billing").Decide(ctx, stop); err == nil || d.Block {
		t.Fatalf("Decide = %+v, %v; a broken delta must not block", d, err)
	}
}

func TestSessionStartKeepsTheBaselineOfTheSession(t *testing.T) {
	ev := &evidence{}
	g := stopGuard(t, ev, "billing")
	if _, err := g.Decide(ctx, Event{Name: "SessionStart", Session: "s1"}); err != nil || ev.began != "s1" {
		t.Fatalf("began = %q, %v; want the session s1", ev.began, err)
	}
}

type linter struct {
	diags []findings.Diagnostic
	err   error
	ran   bool
}

func (l *linter) IsTest(rel string) bool { return linking.IsTestFile(rel) }

func (l *linter) Lint(context.Context) ([]findings.Diagnostic, error) {
	l.ran = true
	return l.diags, l.err
}

func lintGuard(t *testing.T, l Linter) (*Guard, string) {
	t.Helper()
	root := t.TempDir()
	return New(root, tracking.New(specs{}), nil, nil, l), root
}

func TestPostToolUseTellsClaudeAboutProblemsInTheTestFileItWrote(t *testing.T) {
	l := &linter{diags: []findings.Diagnostic{
		{File: "billing/total_test.go", Line: 7, Code: "LINT-SKIP", Message: "skipped", Warn: true},
		{File: "billing/other_test.go", Line: 1, Code: "LINT-SKIP", Message: "skipped", Warn: true},
		{File: "deltas/specs/billing.spec.md", Code: "LINT-NO-CITE", Message: "no test cites INV-02"},
	}}
	g, root := lintGuard(t, l)
	for _, tool := range []string{"Write", "Edit"} {
		d, err := g.Decide(ctx, Event{Name: "PostToolUse", Tool: tool, Path: filepath.Join(root, "billing", "total_test.go")})
		if err != nil || !d.Block || !strings.Contains(d.Reason, "billing/total_test.go:7") || !strings.Contains(d.Reason, "LINT-SKIP") {
			t.Errorf("%s: Decide = %+v, %v; want the problem in that file", tool, d, err)
		}
		if strings.Contains(d.Reason, "other_test.go") || strings.Contains(d.Reason, "LINT-NO-CITE") {
			t.Errorf("%s: %q reports problems that are not in the file just written", tool, d.Reason)
		}
	}
}

func TestPostToolUseStaysQuietWhenThereIsNothingToSay(t *testing.T) {
	l := &linter{diags: []findings.Diagnostic{{File: "deltas/specs/billing.spec.md", Code: "LINT-NO-CITE", Message: "x"}}}
	g, root := lintGuard(t, l)
	cases := map[string]Event{
		"a test file with no problem": {Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "pay_test.go")},
		"a file that is not a test":   {Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "main.go")},
		"a tool that is not a write":  {Name: "PostToolUse", Tool: "Bash", Path: filepath.Join(root, "pay_test.go")},
		"a file outside the project":  {Name: "PostToolUse", Tool: "Write", Path: filepath.Join(t.TempDir(), "x_test.go")},
	}
	for name, e := range cases {
		if d, err := g.Decide(ctx, e); err != nil || d.Block {
			t.Errorf("%s: Decide = %+v, %v; want it allowed", name, d, err)
		}
	}
}

func TestPostToolUseDoesNotLintWhatIsNotATestFile(t *testing.T) {
	l := &linter{}
	g, root := lintGuard(t, l)
	_, _ = g.Decide(ctx, Event{Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "main.go")})
	if l.ran {
		t.Fatal("lint ran for a file that is not a test: it reads every test file of the project")
	}
	if d, err := New(root, tracking.New(specs{}), nil, nil, nil).Decide(ctx, Event{Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "a_test.go")}); err != nil || d.Block {
		t.Fatalf("without a linter: %+v, %v", d, err)
	}
}

func TestPostToolUseFailureIsAnErrorNotAMessage(t *testing.T) {
	g, root := lintGuard(t, &linter{err: errors.New("disk on fire")})
	if d, err := g.Decide(ctx, Event{Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "a_test.go")}); err == nil || d.Block {
		t.Fatalf("Decide = %+v, %v", d, err)
	}
}

type drifting map[string][]string

func (d drifting) Drift(_ context.Context, slice string, _ linking.Scope) ([]string, bool, error) {
	files, ok := d[slice]
	return files, ok, nil
}

func verifiedBilling(state string) specs {
	return specs{"billing": "---\nformat: 1\nslice: billing\nstatus: " + state + "\n---\n- INV-01 [test: go test ./billing]\n  When x, y.\n"}
}

func driftGuard(t *testing.T, s specs, d drifting) (*Guard, string) {
	t.Helper()
	root := t.TempDir()
	return New(root, tracking.New(s).WithDrift(d), nil, nil, nil), root
}

func TestPostToolUseNoLongerMarksASliceStale(t *testing.T) {
	s := verifiedBilling("verified")
	g, root := driftGuard(t, s, drifting{"billing": {"billing/total.go"}})
	for _, tool := range []string{"Write", "Edit"} {
		e := Event{Name: "PostToolUse", Tool: tool, Path: filepath.Join(root, "billing", "total.go")}
		if d, err := g.Decide(ctx, e); err != nil || d.Block || d.Context != "" {
			t.Fatalf("%s: Decide = %+v, %v; editing code is ordinary work", tool, d, err)
		}
	}
	if !strings.Contains(s["billing"], "status: verified") {
		t.Fatalf("the spec changed: %q", s["billing"])
	}
}

func TestSessionStartNamesTheFilesThatChangedAVerifiedSlice(t *testing.T) {
	s := verifiedBilling("verified")
	g, _ := driftGuard(t, s, drifting{"billing": {"a.go", "b.go", "c.go", "d.go"}})
	d, err := g.Decide(ctx, Event{Name: "SessionStart"})
	if err != nil || !strings.Contains(d.Context, "billing: verified") || !strings.Contains(d.Context, "a.go, b.go, c.go and 1 more") || !strings.Contains(d.Context, "next: delta check billing confirms it still holds") {
		t.Fatalf("Decide = %+v, %v", d, err)
	}
}
