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

func TestEveryPromptCarriesTheNextStepOfEachSlice(t *testing.T) {
	g, _ := newGuard(t, specs{
		"billing":  "---\nformat: 1\nslice: billing\nstatus: draft\n---\n- INV-01 [none]\n  When x, y.\n",
		"payments": "---\nformat: 1\nslice: payments\nstatus: approved\n---\n- INV-01 [none]\n  When x, y.\n",
	})
	d, err := g.Decide(ctx, Event{Name: "UserPromptSubmit"})
	if err != nil || d.Block {
		t.Fatalf("%+v, %v", d, err)
	}
	for _, want := range []string{"comper skill", "- billing: draft: ask the person in plain words", "- payments: approved, not built: the builder skill"} {
		if !strings.Contains(d.Context, want) {
			t.Errorf("the context lacks %q:\n%s", want, d.Context)
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
	by    map[string]judgment.Standing
	err   error
	asked map[string]linking.Scope
}

func (e *evidence) Standings(_ context.Context, slices map[string]linking.Scope) (map[string]judgment.Standing, error) {
	e.asked = slices
	return e.by, e.err
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

func TestStopAsksForACheckOfEverySliceThatIsNotCheckedYet(t *testing.T) {
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked, "payments": judgment.Passing}}
	d, err := stopGuard(t, ev, "billing", "payments").Decide(ctx, Event{Name: "Stop"})
	if err != nil || !d.Block || !strings.Contains(d.Reason, "billing") || strings.Contains(d.Reason, "payments") || !strings.Contains(d.Reason, "delta check") {
		t.Fatalf("Decide = %+v, %v; want a block that names only billing", d, err)
	}
	if len(ev.asked) != 2 || !ev.asked["billing"].Covers("billing/x.go") || ev.asked["billing"].Covers("README.md") {
		t.Fatalf("asked about %v, want both slices with the scope their commands name", ev.asked)
	}
}

func TestStopSaysWhenAClaimWasViolated(t *testing.T) {
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Failing}}
	d, _ := stopGuard(t, ev, "billing").Decide(ctx, Event{Name: "Stop"})
	if !d.Block || !strings.Contains(d.Reason, "violated") || !strings.Contains(d.Reason, "billing") {
		t.Fatalf("Decide = %+v, want a block about violated claims", d)
	}
}

func TestStopLetsClaudeFinishWhenThereIsNothingToProve(t *testing.T) {
	passing := &evidence{by: map[string]judgment.Standing{"billing": judgment.Passing}}
	cases := map[string]*Guard{
		"every slice stands":           stopGuard(t, passing, "billing"),
		"no slice has a test claim":    New(t.TempDir(), tracking.New(specs{"billing": "---\nformat: 1\nslice: billing\nstatus: built\n---\n- INV-01 [none]\n  When x, y.\n"}), nil, &evidence{}, nil),
		"there are no slices":          stopGuard(t, &evidence{}),
		"there is no evidence to read": stopGuard(t, nil, "billing"),
	}
	for name, g := range cases {
		if d, err := g.Decide(ctx, Event{Name: "Stop"}); err != nil || d.Block {
			t.Errorf("%s: Decide = %+v, %v; want it allowed", name, d, err)
		}
	}
}

func TestStopPointsAtForgeForAnApprovedSliceWhoseCodeChanged(t *testing.T) {
	head := func(slice, status string) string {
		return "---\nformat: 1\nslice: " + slice + "\nstatus: " + status + "\n---\n"
	}
	approved := map[string]string{
		"billing":  head("billing", "approved") + "- INV-01 [none]\n  When x, y.\n",
		"payments": head("payments", "stale") + "- INV-01 [test: go test ./payments]\n  When x, y.\n",
	}
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked, "payments": judgment.Failing}}
	g := New(t.TempDir(), tracking.New(specs(approved)), nil, ev, nil)
	d, err := g.Decide(ctx, Event{Name: "Stop"})
	if err != nil || !d.Block || !strings.Contains(d.Reason, "/builder billing") || !strings.Contains(d.Reason, "/builder payments") || strings.Contains(d.Reason, "delta check") {
		t.Fatalf("Decide = %+v, %v; want a block that points at /builder for both and not at delta check", d, err)
	}
	if len(ev.asked) != 2 {
		t.Fatalf("asked about %v: an approved slice without test claims still has a tree to look at", ev.asked)
	}
}

func TestStopLeavesAnApprovedSliceAloneWhileNothingChanged(t *testing.T) {
	spec := "---\nformat: 1\nslice: billing\nstatus: approved\n---\n- INV-01 [none]\n  When x, y.\n"
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Passing}}
	g := New(t.TempDir(), tracking.New(specs{"billing": spec}), nil, ev, nil)
	if d, err := g.Decide(ctx, Event{Name: "Stop"}); err != nil || d.Block {
		t.Fatalf("Decide = %+v, %v; an untouched approved slice is not pending work", d, err)
	}
}

func TestStopAsksOnlyOnce(t *testing.T) {
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked}}
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

func (s specs) Save(_ context.Context, slice string, content []byte) error {
	s[slice] = string(content)
	return nil
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
	return New(root, tracking.New(s).WithDrift(d, s), nil, nil, nil), root
}

func TestPostToolUseTellsClaudeThatAVerifiedSliceWentStale(t *testing.T) {
	s := verifiedBilling("verified")
	g, root := driftGuard(t, s, drifting{"billing": {"billing/total.go"}})
	e := Event{Name: "PostToolUse", Tool: "Edit", Path: filepath.Join(root, "billing", "total.go")}
	d, err := g.Decide(ctx, e)
	for _, want := range []string{"billing", "stale", "billing/total.go", "/builder billing"} {
		if err != nil || !d.Block || !strings.Contains(d.Reason, want) {
			t.Fatalf("Decide = %+v, %v; want a block that mentions %q", d, err, want)
		}
	}
	if !strings.Contains(s["billing"], "status: stale") {
		t.Fatalf("the spec was not marked stale: %q", s["billing"])
	}
	if d, err := g.Decide(ctx, e); err != nil || d.Block {
		t.Fatalf("second Decide = %+v, %v; a slice already stale is not news again", d, err)
	}
}

func TestPostToolUseLeavesAVerifiedSliceAloneWhenNothingChanged(t *testing.T) {
	s := verifiedBilling("verified")
	g, root := driftGuard(t, s, drifting{"billing": {}})
	if d, err := g.Decide(ctx, Event{Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "billing", "total.go")}); err != nil || d.Block {
		t.Fatalf("Decide = %+v, %v", d, err)
	}
	if !strings.Contains(s["billing"], "status: verified") {
		t.Fatalf("the spec changed: %q", s["billing"])
	}
}

func TestPostToolUseDoesNotLookForDriftWhenTheKnowledgeBaseWasTheEdit(t *testing.T) {
	s := verifiedBilling("verified")
	g, root := driftGuard(t, s, drifting{"billing": {"billing/total.go"}})
	if d, err := g.Decide(ctx, Event{Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "CLAUDE.md")}); err != nil || d.Block {
		t.Fatalf("Decide = %+v, %v", d, err)
	}
}

func TestPostToolUseKeepsTheLintMessageBesideTheStaleOne(t *testing.T) {
	s := verifiedBilling("verified")
	root := t.TempDir()
	l := &linter{diags: []findings.Diagnostic{{File: "billing/total_test.go", Line: 3, Code: "LINT-SKIP", Message: "skipped", Warn: true}}}
	g := New(root, tracking.New(s).WithDrift(drifting{"billing": {"billing/total_test.go"}}, s), nil, nil, l)
	d, err := g.Decide(ctx, Event{Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "billing", "total_test.go")})
	if err != nil || !d.Block || !strings.Contains(d.Reason, "stale") || !strings.Contains(d.Reason, "LINT-SKIP") {
		t.Fatalf("Decide = %+v, %v; want both messages", d, err)
	}
}

func TestPostToolUseSaysWhenItCouldNotTellWhatWentStale(t *testing.T) {
	s := verifiedBilling("verified")
	root := t.TempDir()
	g := New(root, tracking.New(s).WithDrift(brokenDrifting{}, s), nil, nil, nil)
	if d, err := g.Decide(ctx, Event{Name: "PostToolUse", Tool: "Write", Path: filepath.Join(root, "a.go")}); err == nil || d.Block {
		t.Fatalf("Decide = %+v, %v; a broken delta must not block", d, err)
	}
}

type brokenDrifting struct{}

func (brokenDrifting) Drift(context.Context, string, linking.Scope) ([]string, bool, error) {
	return nil, true, errors.New("git is gone")
}

func TestStopSaysAStaleSliceIsStaleNotUnbuilt(t *testing.T) {
	s := verifiedBilling("stale")
	ev := &evidence{by: map[string]judgment.Standing{"billing": judgment.Unchecked}}
	g := New(t.TempDir(), tracking.New(s), nil, ev, nil)
	d, err := g.Decide(ctx, Event{Name: "Stop"})
	if err != nil || !d.Block || !strings.Contains(d.Reason, "billing is stale") || !strings.Contains(d.Reason, "/builder billing") || strings.Contains(d.Reason, "approved but not built") {
		t.Fatalf("Decide = %+v, %v", d, err)
	}
}

func TestSessionStartNamesTheFilesThatMadeASliceStale(t *testing.T) {
	s := verifiedBilling("verified")
	g, _ := driftGuard(t, s, drifting{"billing": {"a.go", "b.go", "c.go", "d.go"}})
	d, err := g.Decide(ctx, Event{Name: "SessionStart"})
	if err != nil || !strings.Contains(d.Context, "billing: stale") || !strings.Contains(d.Context, "a.go, b.go, c.go and 1 more") {
		t.Fatalf("Decide = %+v, %v", d, err)
	}
}
