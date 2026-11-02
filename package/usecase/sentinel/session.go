package sentinel

import (
	"context"
	"fmt"
	"strings"

	"github.com/ronz204/deltas-ai/package/domain/states"
	"github.com/ronz204/deltas-ai/package/usecase/tracking"
)

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

func (g *Guard) prompt(ctx context.Context) (Decision, error) {
	rows, err := g.slices.Report(ctx)
	if err != nil {
		return Decision{}, err
	}
	var b strings.Builder
	b.WriteString("Delta: if this request adds, changes, removes, checks or documents what the project does, use the comper skill before touching any file; the person may speak plainly or call a skill by name, and never types delta commands.\n")
	for _, r := range rows {
		if next := nextStep(r); next != "" {
			fmt.Fprintf(&b, "- %s: %s\n", r.Slice, next)
		}
	}
	return Decision{Context: strings.TrimRight(b.String(), "\n")}, nil
}

func nextStep(r tracking.Row) string {
	switch {
	case r.Issues > 0:
		return "its spec has problems: fix them with the drafter skill"
	case r.Status == states.Draft:
		return "draft: ask the person in plain words whether to approve it, and on a yes run delta approve"
	case r.Status == states.Approved:
		return "approved, not built: the builder skill builds it"
	case r.Status == states.Stale:
		return "stale: the builder skill rebuilds it, or the drafter skill if the contract is out of date"
	case r.Status == states.Built:
		return "built, not verified: the auditor skill checks it"
	}
	return ""
}
