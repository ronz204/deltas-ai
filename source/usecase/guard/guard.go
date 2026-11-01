// Package guard decides what the agent may do while it works on a project.
package guard

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ronz204/deltas-ai/source/domain/commits"
	"github.com/ronz204/deltas-ai/source/domain/knowledge"
	"github.com/ronz204/deltas-ai/source/usecase/status"
)

// Event is what a Claude Code hook tells us is about to happen or has just happened.
type Event struct {
	Name    string // SessionStart, PreToolUse, PostToolUse or Stop
	Tool    string
	Path    string // a file the tool is about to touch
	Command string // the command line a Bash call is about to run
}

// Decision is the answer to an event.
type Decision struct {
	Block   bool
	Reason  string // why the action is blocked; Claude reads it
	Context string // facts Claude gets at the start of a session
}

type Guard struct {
	root   string
	slices *status.Reporter
}

func New(root string, slices *status.Reporter) *Guard { return &Guard{root: root, slices: slices} }

var fileTools = map[string]bool{"Write": true, "Edit": true}

// Decide answers an event. Events it has no opinion about are allowed.
func (g *Guard) Decide(ctx context.Context, e Event) (Decision, error) {
	switch e.Name {
	case "SessionStart":
		return g.start(ctx)
	case "PreToolUse":
		return g.before(e), nil
	}
	return Decision{}, nil
}

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
	if err != nil || len(rows) == 0 {
		return Decision{}, err
	}
	var b strings.Builder
	b.WriteString("Delta slices:\n")
	for _, r := range rows {
		state := string(r.Status)
		if state == "" {
			state = "unreadable"
		}
		fmt.Fprintf(&b, "- %s: %s, %d claims (%d test, %d manual, %d none)", r.Slice, state, r.Claims, r.Test, r.Manual, r.None)
		if r.Issues > 0 {
			fmt.Fprintf(&b, ", %d problems in its spec", r.Issues)
		}
		b.WriteByte('\n')
	}
	b.WriteString("The knowledge base is not edited directly: specs are written with delta spec write and tests are linked with delta link.")
	return Decision{Context: b.String()}, nil
}
