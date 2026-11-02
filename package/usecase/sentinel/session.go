package sentinel

import (
	"context"
	"fmt"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/states"
	"github.com/ronz204/deltas-ai/package/usecase/tracking"
)

func (g *Guard) start(ctx context.Context, e Event) (Decision, error) {
	if g.evidence != nil && e.Session != "" {
		_ = g.evidence.Begin(ctx, e.Session)
	}
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
		if next := nextStep(r); next != "" {
			fmt.Fprintf(&b, "; next: %s", next)
		}
		b.WriteByte('\n')
	}
	if len(problems) > 0 {
		b.WriteString("Delta setup problems:\n")
		for _, p := range problems {
			b.WriteString("- " + p + "\n")
		}
	}
	b.WriteString("The knowledge base is not edited directly: specs are written with delta spec write and tests are linked with delta link. A request that changes what a feature promises goes through the comper skill; anything else is ordinary work.")
	return Decision{Context: b.String()}, nil
}

const promptLine = "Delta: if this request changes what a feature promises (new behavior, or a guarantee that is different), use the comper skill first. Bug fixes, refactors and tweaks need no ceremony: work as usual, and delta asks for a check of the slices you touched before you finish."

func (g *Guard) prompt() Decision { return Decision{Context: promptLine} }

func nextStep(r tracking.Row) string {
	switch {
	case r.Issues > 0:
		return "fix its spec with the drafter skill"
	case r.Status == states.Draft:
		return "ask the person in plain words whether to approve it, and on a yes run delta approve"
	case r.Status == states.Approved:
		return "the builder skill builds it"
	case r.Status == states.Stale, r.Status == states.Verified && len(r.Changed) > 0:
		return "delta check " + r.Slice + " confirms it still holds, or the builder skill rebuilds it, or the drafter skill if the contract is out of date"
	case r.Status == states.Built:
		return "the auditor skill checks it"
	}
	return ""
}
