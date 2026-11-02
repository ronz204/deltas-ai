// Package guard decides what the agent may do while it works on a project.
package guard

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/commits"
	"github.com/ronz204/deltas-ai/source/domain/findings"
	"github.com/ronz204/deltas-ai/source/domain/judgment"
	"github.com/ronz204/deltas-ai/source/domain/knowledge"
	"github.com/ronz204/deltas-ai/source/domain/lifecycle"
	"github.com/ronz204/deltas-ai/source/domain/linkcheck"
	"github.com/ronz204/deltas-ai/source/usecase/status"
)

// Event is what a Claude Code hook tells us is about to happen or has just happened.
type Event struct {
	Name    string // SessionStart, PreToolUse, PostToolUse or Stop
	Tool    string
	Path    string // a file the tool is about to touch
	Command string // the command line a Bash call is about to run

	// StopActive is true when Claude is already continuing because an earlier Stop hook blocked it.
	StopActive bool
}

// Decision is the answer to an event.
type Decision struct {
	Block   bool   // after a tool ran there is nothing to block: Claude just gets the reason
	Reason  string // why the action is blocked; Claude reads it
	Context string // facts Claude gets at the start of a session
}

// Health lists what is wrong with how delta is installed in the project.
type Health interface {
	Check(ctx context.Context) []string
}

// Evidence says what the last check of each slice proves about the part of the tree the slice depends on.
// A slice whose part has no changes has nothing to prove and stands.
type Evidence interface {
	Standings(ctx context.Context, slices map[string]linkcheck.Scope) (map[string]judgment.Standing, error)
}

// Linter cross-checks the specs against the test files.
type Linter interface {
	Lint(ctx context.Context) ([]findings.Diagnostic, error)
	IsTest(rel string) bool
}

type Guard struct {
	root     string
	slices   *status.Reporter
	health   Health
	evidence Evidence
	linter   Linter
}

// New builds the guard of the project at root. health, evidence and linter may be nil: then no setup problem
// is reported, Stop never asks for a check and a test file that was edited is not linted.
func New(root string, slices *status.Reporter, health Health, evidence Evidence, linter Linter) *Guard {
	return &Guard{root: root, slices: slices, health: health, evidence: evidence, linter: linter}
}

var fileTools = map[string]bool{"Write": true, "Edit": true}

// Decide answers an event. Events it has no opinion about are allowed.
func (g *Guard) Decide(ctx context.Context, e Event) (Decision, error) {
	switch e.Name {
	case "SessionStart":
		return g.start(ctx)
	case "PreToolUse":
		return g.before(e), nil
	case "PostToolUse":
		return g.after(ctx, e)
	case "Stop":
		return g.stop(ctx, e)
	}
	return Decision{}, nil
}

// after tells Claude what the file it just wrote changed: a slice that was verified and is now stale, and, for a
// test file, the lint problems in that file.
func (g *Guard) after(ctx context.Context, e Event) (Decision, error) {
	if !fileTools[e.Tool] {
		return Decision{}, nil
	}
	rel, ok := g.relative(e.Path)
	if !ok {
		return Decision{}, nil
	}
	var lines []string
	if !knowledge.Covers(rel) {
		drifted, err := g.slices.Stale(ctx)
		if err != nil {
			return Decision{}, err
		}
		for _, d := range drifted {
			if d.Marked {
				lines = append(lines, fmt.Sprintf("%s was verified and is now stale: %s changed. Check that its contract still holds: if it does, rebuild with /forge %s; if the contract is out of date, say so to the user, who decides it with /quill.", d.Slice, fileList(d.Files), d.Slice))
			}
		}
	}
	lint, err := g.lintOf(ctx, rel)
	if err != nil {
		return Decision{}, err
	}
	if lines = append(lines, lint...); len(lines) == 0 {
		return Decision{}, nil
	}
	return Decision{Block: true, Reason: strings.Join(lines, "\n")}, nil
}

// lintOf is the lint problems of one file, if it is a test file. Problems in other files are not its business here:
// a claim without a test yet is normal while the tests are being written.
func (g *Guard) lintOf(ctx context.Context, rel string) ([]string, error) {
	if g.linter == nil || !g.linter.IsTest(rel) {
		return nil, nil
	}
	ds, err := g.linter.Lint(ctx)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, d := range ds {
		if d.File == rel {
			lines = append(lines, "- "+d.String())
		}
	}
	if len(lines) == 0 {
		return nil, nil
	}
	return append([]string{"delta lint found problems in the test file you just wrote:"}, lines...), nil
}

// fileList names up to three files and counts the rest.
func fileList(files []string) string {
	const shown = 3
	if len(files) <= shown {
		return strings.Join(files, ", ")
	}
	return fmt.Sprintf("%s and %d more", strings.Join(files[:shown], ", "), len(files)-shown)
}

// stop keeps Claude from finishing while a slice with test claims has not been checked against the tree as
// it stands, or while an approved slice has had code written for it without Forge finishing it (delta built).
// It asks once: when Claude is already continuing because of a stop hook, it lets it finish.
func (g *Guard) stop(ctx context.Context, e Event) (Decision, error) {
	if e.StopActive || g.evidence == nil {
		return Decision{}, nil
	}
	rows, err := g.slices.Report(ctx)
	if err != nil {
		return Decision{}, err
	}
	var names []string
	scopes := map[string]linkcheck.Scope{}
	for _, r := range rows {
		if r.Issues == 0 && (r.Test > 0 || waitsForForge(r.Status)) {
			names = append(names, r.Slice)
			scopes[r.Slice] = r.Scope
		}
	}
	if len(names) == 0 {
		return Decision{}, nil
	}
	by, err := g.evidence.Standings(ctx, scopes)
	if err != nil {
		return Decision{}, err
	}
	state := map[string]lifecycle.Status{}
	for _, r := range rows {
		state[r.Slice] = r.Status
	}
	var unchecked, failing, unbuilt []string
	for _, n := range names {
		switch {
		case by[n] == judgment.Passing:
		case waitsForForge(state[n]):
			unbuilt = append(unbuilt, n)
		case by[n] == judgment.Unchecked:
			unchecked = append(unchecked, n)
		case by[n] == judgment.Failing:
			failing = append(failing, n)
		}
	}
	var lines []string
	for _, n := range unbuilt {
		if state[n] == lifecycle.Stale {
			lines = append(lines, fmt.Sprintf("%s is stale and its code changed again: bring it up to date with /forge %s, which ends with delta built %s, or tell the user what is left.", n, n, n))
			continue
		}
		lines = append(lines, fmt.Sprintf("%s is approved but not built, and the code changed since: finish it with /forge %s, which ends with delta built %s, or tell the user what is left.", n, n, n))
	}
	if len(unchecked) > 0 {
		lines = append(lines, "delta check has not run on the current changes for: "+strings.Join(unchecked, ", ")+". Run it before you finish.")
	}
	if len(failing) > 0 {
		lines = append(lines, "delta check found violated claims for: "+strings.Join(failing, ", ")+". Fix them, or tell the user what is left.")
	}
	if len(lines) == 0 {
		return Decision{}, nil
	}
	return Decision{Block: true, Reason: strings.Join(lines, "\n")}, nil
}

// waitsForForge says whether a slice is approved to be built (or rebuilt) and nobody has said it is built.
func waitsForForge(s lifecycle.Status) bool { return s == lifecycle.Approved || s == lifecycle.Stale }

func (g *Guard) before(e Event) Decision {
	switch {
	case fileTools[e.Tool]:
		if rel, ok := g.relative(e.Path); ok && knowledge.Covers(rel) {
			return Decision{Block: true, Reason: fmt.Sprintf(
				"%s is part of the knowledge base, which is not edited directly. Specs are written with delta spec write and tests are linked with delta link; a person can edit it by hand.", rel)}
		}
	case e.Tool == "Bash":
		if message, ok := commits.FromCommand(e.Command); ok {
			return commitDecision(message)
		}
	}
	return Decision{}
}

func commitDecision(message string) Decision {
	var lines []string
	for _, d := range commits.Validate("commit message", message) {
		lines = append(lines, "- "+d.Code+": "+d.Message)
	}
	if len(lines) == 0 {
		return Decision{}
	}
	return Decision{Block: true, Reason: "The commit message breaks the project's rules:\n" + strings.Join(lines, "\n") + "\nFix the message and commit again."}
}

// relative gives path relative to the project root, with forward slashes, and false when it is
// outside the project.
func (g *Guard) relative(path string) (string, bool) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(g.root, path)
	}
	rel, err := filepath.Rel(g.root, path)
	if err != nil {
		return "", false
	}
	rel = filepath.ToSlash(rel)
	return rel, rel != ".." && !strings.HasPrefix(rel, "../")
}

func (g *Guard) start(ctx context.Context) (Decision, error) {
	rows, err := g.slices.Report(ctx)
	if err != nil {
		return Decision{}, err
	}
	var problems []string
	if g.health != nil {
		problems = g.health.Check(ctx)
	}
	if len(rows) == 0 && len(problems) == 0 {
		return Decision{}, nil
	}
	var b strings.Builder
	if len(rows) > 0 {
		b.WriteString("Delta slices:\n")
	}
	for _, r := range rows {
		state := string(r.Status)
		if state == "" {
			state = "unreadable"
		}
		fmt.Fprintf(&b, "- %s: %s, %d claims (%d test, %d manual, %d none)", r.Slice, state, r.Claims, r.Test, r.Manual, r.None)
		if r.Issues > 0 {
			fmt.Fprintf(&b, ", %d problems in its spec", r.Issues)
		}
		if len(r.Changed) > 0 {
			fmt.Fprintf(&b, ", changed since it was verified: %s", fileList(r.Changed))
		}
		b.WriteByte('\n')
	}
	if len(problems) > 0 {
		b.WriteString("Delta setup problems:\n")
		for _, p := range problems {
			b.WriteString("- " + p + "\n")
		}
	}
	b.WriteString("The knowledge base is not edited directly: specs are written with delta spec write and tests are linked with delta link.")
	return Decision{Context: b.String()}, nil
}
